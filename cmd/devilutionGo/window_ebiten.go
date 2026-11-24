//go:build windowed

package main

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/game"
)

// windowTitle is the window caption.
const windowTitle = "devilutionGo"

// windowGame is the ebiten.Game wrapper (v9xy): Update drives the tick
// shim, Draw presents the presenter target, ESC/q ends rc 0.
// The window layer is the only place raw input codes appear (v9xy note).
type windowGame struct {
	present *scenePresenter
	lastLog time.Time
	ticks   int

	// lane is the kernel-wired game the window drives: the bridge session the
	// window layer submits decisions through, and the world slice behind it.
	lane *deviceLane
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

	slog.Info("devilutionGo started", "mode", "windowed", "debug", g.Debug != nil)

	ebiten.SetWindowTitle(windowTitle)

	err := ebiten.RunGame(newWindowGame(lane, args, screenshotDir))
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

// newWindowGame builds the window wrapper with its presenter.
func newWindowGame(lane *deviceLane, args []string, screenshotDir string) *windowGame {
	boot := &windowGame{lane: lane}
	boot.present = newScenePresenter(args, screenshotDir)
	boot.lastLog = clockNow()

	return boot
}

// Update advances the game one frame and stops on the quit keys.
func (w *windowGame) Update() error {
	if isQuitKey(windowKey()) || windowClosed() {
		slog.Info("windowed boot exit", "frames", w.ticks)

		return ebiten.Termination
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
