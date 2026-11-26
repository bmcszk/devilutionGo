//go:build windowed

package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/bmcszk/devilutionGo/internal/diablo/levels"
	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/game"
	"github.com/bmcszk/devilutionGo/internal/game/ui"
	"github.com/bmcszk/devilutionGo/internal/game/world"
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

	// setGameMenuSession re-installs the in-game menu with the hero slot and
	// the level seed of the session that just entered the game: wired at
	// boot construction to installGameMenu, called from the selhero entry
	// path (review F3; the same closure-arg pattern as onCloseGame).
	setGameMenuSession func(slot int, seed uint32)

	// lane is the kernel-wired game the window drives: the bridge session the
	// window layer submits decisions through, and the world slice behind it.
	lane *deviceLane

	// stores is the per-hero store set the selhero screen reads and writes:
	// dir-bound, handle-free (ADR-003 per-hero single_N.db model).
	stores *persist.HeroStoreSet

	// selhero is the screen buildWindowGame installed on the flow: the test
	// boot's read of the production wiring (review F1's black-box probe).
	selhero *selHeroScreen

	// menu is the in-game Esc menu the flow's FrontGame arm delegates to
	// (devilGo-1a4r S4); nil until installGameMenu runs.
	menu *gameMenuScreen

	// menuSeed is the level seed of the session the walk entered, threaded
	// from the selhero load/create outcome into installGameMenu (review F3).
	menuSeed uint32

	// menuOpen reports whether the in-game menu is open: the Game.MenuOpen
	// closure the pause rule reads (X:diablo.cpp:229-232). Nil = never open.
	menuOpen func() bool

	// simOpts is the in-memory SimOptions the boot owns; the SetTickRate
	// onRate closure updates it. NOTHING persists here — the window boot
	// opens no engine.Ini; the INI write lands with the runtime-ini bean
	// (recorded deviation, X:gmenu.cpp:246-248 SaveOptions).
	simOpts engine.SimOptions
}

// buildWindowGame is the windowed boot's construction surface: the lane, the
// store and the window game over them. Ebiten-free, so the tests drive the
// exact wiring runWindowed boots (review F1).
func buildWindowGame(storeDir string) (*windowGame, error) {
	src := &engine.MockTickSource{}
	g := game.NewGame(src)
	lane := bootDeviceLane(src, g)

	stores := persist.NewHeroStoreSet(storeDir, persist.SaveNamespace{})

	// The save/load verbs register over the set: the archive opens the slot's
	// own file per operation, and the level source is the town-only one (see
	// the bootDeviceLane disclosure — a dungeon record restores nothing until
	// a production dungeon corpus loader exists).
	lane.attachSaveLoad(persist.NewSessionArchive(stores, lanePackSource), townLevelSource())

	boot, err := newWindowGame(lane, stores, "")
	if err != nil {
		return nil, err
	}

	return boot, nil
}

// townLevelSource is the boot's LevelSource: town only. NewTownLayoutSource
// over nil corpus has NO level image to serve: calling LevelImage at (0,0)
// would panic on the nil corpus (GetDunSize reads a nil slice). The production
// path never reaches that read: a Loaded record always has Level >= 1
// (session_archive.go derives Loaded = CurrLevel != 0) and the (level,
// levelType) guard refuses before the corpus is touched — so a dungeon record
// restores nothing (silent, ADR-018 law 2) until a production dungeon corpus
// loader lands.
func townLevelSource() world.LevelSource { return levels.NewTownLayoutSource(nil, nil, nil, 0) }

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

	// The real selhero screen: the per-hero store set binds beside the working
	// directory and the screen installs on the flow (review F1). Nothing stays
	// open — the set opens each slot's file per operation (pfile.cpp:609-612).
	stores := persist.NewHeroStoreSet("", persist.SaveNamespace{})

	// The save/load verbs register over the set (see buildWindowGame).
	lane.attachSaveLoad(persist.NewSessionArchive(stores, lanePackSource), townLevelSource())

	slog.Info("devilutionGo started", "mode", "windowed", "debug", g.Debug != nil)

	ebiten.SetWindowTitle(windowTitle)

	run, err := newWindowGame(lane, stores, screenshotDir)
	if err != nil {
		return fmt.Errorf("new window game: %w", err)
	}

	err = ebiten.RunGame(run)
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
func newWindowGame(lane *deviceLane, stores *persist.HeroStoreSet, screenshotDir string) (*windowGame, error) {
	boot := &windowGame{lane: lane, flow: scene.NewFrontFlow(), stores: stores, simOpts: engine.DefaultSimOptions()}

	screen := newSelHeroScreenWithGameSession(stores, lane, boot.flow.RequestBackWalk, boot.flow.EnterGame,
		func(slot int, seed uint32) { boot.setGameMenuSession(slot, seed) })
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

	boot.installGameMenu(-1, 0)

	// The session identity threads at game entry: the close path (a settled
	// load or new game) knows the hero slot, and the load path peeked the
	// level seed (selhero.go:421-428) — the same closure-arg pattern as
	// onCloseGame (review F3).
	boot.setGameMenuSession = func(slot int, seed uint32) {
		boot.installGameMenu(slot, seed)
	}

	return boot, nil
}

// installGameMenu builds the in-game Esc menu over the boot-owned surfaces
// and wires the three pause/rate seams: the flow's FrontGame delegate, the
// Game.MenuOpen closure (the ProcessInput rule reads it) and the
// SetTickRate onRate closure over the boot's in-memory SimOptions. slot is
// the hero slot the save verb freezes; -1 keeps the load-time slot probe.
// seed is the session's level seed threaded from the entry walk (review F3).
func (w *windowGame) installGameMenu(slot int, seed uint32) {
	w.menu = newGameMenuScreen(w.lane, w.flow, w.stores, max(slot, 0), seed)
	w.menuSeed = seed
	w.flow.SetGameScreen(w.menu)
	w.flow.SetGameMenuContent(w.menu)
	w.menuOpen = w.menu.isOpen
	game.RegisterTickRate(w.lane.game.Kernel, w.lane.game, func(rate int) {
		w.simOpts.TickRate = rate
	})
	w.lane.game.MenuOpen = func() bool { return w.menuOpen != nil && w.menuOpen() }
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

	// While the front flow is up (title/menu/selhero) the town stays the
	// terminal render behind it but takes no input: the flow owns the keys
	// (devilGo-uuyh), so the movement lane runs only once the flow entered
	// the game (FrontGame). The front screens still run the frame so a
	// pending create or load settles: the CreatedPack/LoadedSlot outcome
	// lands and the hero persists (dcob D10 / RF1). The frame also advances
	// the game clock and TickCount on every front screen — the clock runs
	// from the first frame, even behind the title/menu (review R1 F10; each
	// ebiten Update is one ~16 ms cadence step).
	if w.flow.Screen() == scene.FrontTitle || w.flow.Screen() == scene.FrontMainMenu ||
		w.flow.Screen() == scene.FrontSelHero {
		w.lane.frame(func() {})
		w.ticks++
		w.logProgress()

		return nil
	}

	// The in-game menu owns the keys while open: the movement lane is
	// suppressed and the WORLD tick stays paused (ProcessInput,
	// X:diablo.cpp:229-232), but the kernel pass runs — review F1 fix shape
	// (b1): RunKernelPassWhilePaused drains the command queue on a tick that
	// is allowed to run, so a Speed (SetTickRate) decision applies while the
	// player watches the slider. Save/Load close the menu on activation
	// (upstream gamemenu_off-first) and settle on the resumed ticks.
	if w.menu != nil && w.menu.isOpen() {
		w.lane.framePaused(func() {})
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
