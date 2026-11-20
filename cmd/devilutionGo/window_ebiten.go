//go:build windowed

package main

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/bmcszk/devilutionGo/internal/diablo/inputbridge"
	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/game"
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// windowTitle is the window caption.
const windowTitle = "devilutionGo"

// windowGame is the ebiten.Game wrapper (v9xy): Update drives the tick
// shim, Draw presents the presenter target, ESC/q ends rc 0.
// The window layer is the only place raw input codes appear (v9xy note).
type windowGame struct {
	clock   *frameClock
	present *scenePresenter
	lastLog time.Time
	ticks   int

	// input is the bridge session that submits movement decisions at the tick boundary.
	input *inputbridge.Session
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

	// Wire the real kernel with session gate and movement handlers
	state := world.NewState()
	state.CreateFromPack(world.StartingPack{Class: 0}, world.Tile{X: 50, Y: 50})
	streams := kernel.NewStreams(0xdeadbeef, "movement")
	gate := world.NewSessionGate(state, state, 0)
	k := kernel.New(gate, streams)
	world.RegisterMovement(k, state)
	g.Kernel = k

	slog.Info("devilutionGo started", "mode", "windowed", "debug", g.Debug != nil)

	ebiten.SetWindowTitle(windowTitle)

	err := ebiten.RunGame(newWindowGame(src, g, args, screenshotDir))
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

// newWindowGame builds the window wrapper with its tick shim and presenter.
func newWindowGame(src *engine.MockTickSource, g *game.Game, args []string, screenshotDir string) *windowGame {
	boot := &windowGame{clock: newFrameClock(src, g)}
	boot.present = newScenePresenter(args, screenshotDir)
	boot.lastLog = clockNow()
	boot.input = inputbridge.NewSession(g.Kernel)

	return boot
}

// Update advances the game one frame and stops on the quit keys.
func (w *windowGame) Update() error {
	w.clock.step(frameDuration)
	w.ticks++

	if isQuitKey(windowKey()) || windowClosed() {
		slog.Info("windowed boot exit", "frames", w.ticks)

		return ebiten.Termination
	}

	// Read keyboard arrows and submit stick decisions through the inputbridge.
	// The window layer is the only place raw input codes appear (v9xy input note).
	windowMovement(w.input)

	// Close the tick boundary: buffered decisions submit at this tick in
	// decision order, then the kernel runs its single ordered pass (ADR-018).
	w.input.Tick(uint32(w.ticks - 1))

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
