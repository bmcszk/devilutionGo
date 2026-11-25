//go:build windowed

package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/game"
	"github.com/bmcszk/devilutionGo/internal/game/ui"
	"github.com/bmcszk/devilutionGo/internal/persist"
	"github.com/bmcszk/devilutionGo/internal/present/scene"
	"github.com/bmcszk/devilutionGo/internal/render"
)

// windowTitle is the window caption.
const windowTitle = "devilutionGo"

// windowGame is the ebiten.Game wrapper (v9xy): Update drives the tick
// shim, Draw presents the presenter target, ESC/q ends rc 0.
// The window layer is the only place raw input codes appear (v9xy note).
type windowGame struct {
	flow    *scene.FrontFlow
	present *scenePresenter
	lastLog time.Time
	ticks   int

	// lane is the kernel-wired game the window drives: the bridge session the
	// window layer submits decisions through, and the world slice behind it.
	lane *deviceLane

	// store is the persist store the selhero screen reads and writes: the
	// per-hero save database the boot owns for the session (review F1).
	store *persist.SQLiteStore

	// selhero is the screen buildWindowGame installed on the flow: the test
	// boot's read of the production wiring (review F1's black-box probe).
	selhero *selHeroScreen
}

// selheroStorePath is the save database the windowed boot opens: the demo's
// single hero file, beside the executable's working directory (ADR-014's
// per-hero .db model, first slot).
func selheroStorePath() string {
	return "single_0.db"
}

// openHeroStore opens the boot's persist store under dir ("" = the working
// directory), creating the per-hero database if it does not exist yet.
func openHeroStore(dir string) (*persist.SQLiteStore, error) {
	store, err := persist.OpenSQLiteStore(
		context.Background(), filepath.Join(dir, selheroStorePath()))
	if err != nil {
		return nil, fmt.Errorf("windowed boot: open hero store: %w", err)
	}

	return store, nil
}

// buildWindowGame is the windowed boot's construction surface: the lane, the
// store and the window game over them. Ebiten-free, so the tests drive the
// exact wiring runWindowed boots (review F1).
func buildWindowGame(storeDir string) (*windowGame, error) {
	src := &engine.MockTickSource{}
	lane := bootDeviceLane(src, game.NewGame(src))

	store, err := openHeroStore(storeDir)
	if err != nil {
		return nil, err
	}

	return newWindowGame(lane, store, ""), nil
}

// runWindowed is the windowed boot (devilGo-v9xy): opens the game window,
// drives game.NewGame on real frames and exits rc 0 on ESC/q or window close.
// A windowless environment fails with a named error, never a panic
// (paritytest.GPUAvailable's DISPLAY contract).
func runWindowed(args []string, screenshotDir string) error {
	if !displayAvailable() {
		return errors.New("windowed boot: no display (DISPLAY unset); headless mode is -headless")
	}

	src := &engine.MockTickSource{}
	g := game.NewGame(src)
	g.Debug = game.NewDebugRouterFromFlag(args, game.DebugEnv{})

	// Wire the real kernel (world slice, session gate, seeded streams, the
	// movement verbs) and the bridge session the window submits through. The
	// boot's disclosures live on bootDeviceLane.
	lane := bootDeviceLane(src, g)

	// The real selhero screen: the persist store opens beside the working
	// directory and the screen installs on the flow (review F1).
	store, err := openHeroStore("")
	if err != nil {
		return err
	}

	slog.Info("devilutionGo started", "mode", "windowed", "debug", g.Debug != nil)

	ebiten.SetWindowTitle(windowTitle)

	run := newWindowGame(lane, store, screenshotDir)

	err = ebiten.RunGame(run)
	if cerr := store.Close(); cerr != nil {
		slog.Error("windowed boot: close hero store", "err", cerr)
	}

	if err != nil && !errors.Is(err, ebiten.Termination) {
		return fmt.Errorf("run windowed: %w", err)
	}

	return nil
}

// displayAvailable reports whether a real X display exists
// (paritytest.GPUAvailable's contract).
func displayAvailable() bool {
	if displayEnv("DISPLAY") == "" {
		return false
	}

	_, err := x11SocketDirStat()

	return err == nil
}

// newWindowGame builds the window wrapper with its presenter and installs
// the real selhero screen over the boot's persist store (review F1).
func newWindowGame(lane *deviceLane, store *persist.SQLiteStore, screenshotDir string) *windowGame {
	boot := &windowGame{lane: lane, flow: scene.NewFrontFlow(), store: store}

	screen := newSelHeroScreen(store, lane, boot.flow.RequestBackWalk)
	boot.selhero = screen
	boot.flow.SetSelHero(screen, screen, func() {
		if err := screen.refresh(context.Background()); err != nil {
			slog.Error("windowed boot: selhero refresh", "err", err)
		}
	})

	boot.present = newScenePresenter(screenshotDir, boot.flow,
		func() render.TilePos { tile := lane.state.Hero(); return render.TilePos{X: tile.X, Y: tile.Y} },
		composeSceneAt)
	boot.lastLog = clockNow()

	return boot
}

// Update advances the game one frame and stops on the quit keys or the front
// flow's quit request.
func (w *windowGame) Update() error {
	w.flow.HandleKey(windowFlowKey())
	// ESC and q are gone from the quit surface (review F2): the front flow
	// owns Escape end to end (walk-back on the menu/placeholder, quit on the
	// title), and a level-triggered ESC arm here fired on the SAME frame the
	// flow walked back — one Escape on the menu quit instead of walking back.
	// Only the flow's own quit request and the window close end the boot.
	if w.flow.Quit() || windowClosed() {
		slog.Info("windowed boot exit", "frames", w.ticks)

		return ebiten.Termination
	}

	// While the front flow is up (title/menu/placeholder) the town stays the
	// terminal render behind it but takes no input: the flow owns the keys
	// (devilGo-uuyh), so the movement lane runs only once the flow is done
	// with the front screens.
	if w.flow.Screen() == scene.FrontTitle || w.flow.Screen() == scene.FrontMainMenu ||
		w.flow.Screen() == scene.FrontSelHero {
		w.ticks++
		w.logProgress()

		return nil
	}

	// The frame reads the arrows first and settles its ticks second, so the
	// decisions queue before the pass that submits them with the kernel's own
	// tick number (ADR-018 law 1). The window layer is the only place raw input
	// codes appear (v9xy input note) and it never runs a kernel pass of its own:
	// the Game is the single pass driver (game.go:217-223; review s6e F2).
	w.lane.frame(func() { windowMovement(w.lane) })
	w.ticks++

	w.logProgress()

	return nil
}

// Draw presents the presenter target to the screen with integer scaling.
func (w *windowGame) Draw(screen *ebiten.Image) {
	w.present.draw(screen)
	w.present.drawFront(w.flow)
}

// Layout is the world logical surface scaled to the window
// (render.ScaleToWindow's model).
func (*windowGame) Layout(width, height int) (widthOut, heightOut int) {
	return width, height
}

// logProgress records the tick counter every second (v9xy AC: ticks advance).
func (w *windowGame) logProgress() {
	now := clockNow()
	if now.Sub(w.lastLog) < time.Second {
		return
	}

	slog.Info("windowed boot ticking", "frames", w.ticks)
	w.lastLog = now
}

// windowFlowKey maps ebiten's just-pressed keys onto the front flow's key
// surface (the window layer is the only place raw key codes appear, v9xy
// input note; menu keys are single presses, ebiten input skill).
func windowFlowKey() ui.Key {
	just := func(key ebiten.Key) bool { return inpututil.IsKeyJustPressed(key) }

	switch {
	case just(ebiten.KeyEscape):
		return ui.KeyEscape
	case just(ebiten.KeyEnter) || just(ebiten.KeyKPEnter):
		return ui.KeyEnter
	case just(ebiten.KeySpace):
		return ui.KeySpace
	case just(ebiten.KeyUp):
		return ui.KeyArrowUp
	case just(ebiten.KeyDown):
		return ui.KeyArrowDown
	case just(ebiten.KeyLeft):
		return ui.KeyArrowLeft
	case just(ebiten.KeyRight):
		return ui.KeyArrowRight
	case just(ebiten.KeyTab):
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			return ui.KeyShiftTab
		}

		return ui.KeyTab
	default:
		// Only genuinely pressed unmapped keys carry the any-key sentinel:
		// an idle frame must stay KeyUnknown so the title's real-key gate
		// holds (review F1).
		if len(inpututil.AppendPressedKeys(nil)) > 0 {
			return flowKeyAny // any unmapped key: the title advances
		}

		return ui.KeyUnknown
	}
}

// flowKeyAny is a key press the flow's mapping does not name: the title
// screen treats it as its press-any-key advance (title.cpp).
const flowKeyAny = ui.KeyShiftTab + 1
