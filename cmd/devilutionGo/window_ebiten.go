//go:build windowed

package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/game"
)

// runWindowed is the windowed boot (devilGo-v9xy): opens the game window,
// drives game.NewGame on real frames and exits rc 0 on ESC/q or window close.
// A windowless environment fails with a named error, never a panic
// (paritytest.GPUAvailable's DISPLAY contract).
func runWindowed(args []string) error {
	if !displayAvailable() {
		return errors.New("windowed boot: no display (DISPLAY unset); headless mode is -headless")
	}

	src := &engine.MockTickSource{}
	g := game.NewGame(src)
	g.Debug = game.NewDebugRouterFromFlag(args, game.DebugEnv{})

	slog.Info("devilutionGo started", "mode", "windowed", "debug", g.Debug != nil)

	if err := ebiten.RunGame(newWindowGame(src, g, args)); err != nil && !errors.Is(err, ebiten.Termination) {
		return fmt.Errorf("run windowed: %w", err)
	}

	return nil
}

// displayAvailable reports whether a real X display exists
// (paritytest.GPUAvailable's contract).
func displayAvailable() bool {
	if os.Getenv("DISPLAY") == "" {
		return false
	}

	_, err := os.Stat("/tmp/.X11-unix")

	return err == nil
}

// windowGame is the ebiten.Game wrapper (v9xy): Update drives the tick
// shim, Draw presents the presenter target, ESC/q ends rc 0.
type windowGame struct {
	clock   *frameClock
	present *scenePresenter
	lastLog time.Time
	ticks   int
}

// newWindowGame builds the window wrapper with its tick shim and presenter.
func newWindowGame(src *engine.MockTickSource, g *game.Game, args []string) *windowGame {
	boot := &windowGame{clock: newFrameClock(src, g)}
	boot.present = newScenePresenter(args)
	boot.lastLog = clockNow()

	return boot
}

// Update advances the game one frame and stops on the quit keys.
func (w *windowGame) Update() error {
	w.ticks += w.clock.step(frameDuration)

	if isQuitKey(windowKey()) || windowClosed() {
		slog.Info("windowed boot exit", "frames", w.ticks)

		return ebiten.Termination
	}

	w.logProgress()

	return nil
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

// Draw presents the presenter target to the screen with integer scaling.
func (w *windowGame) Draw(screen *ebiten.Image) {
	w.present.draw(screen)
}

// Layout is the world logical surface scaled to the window
// (render.ScaleToWindow's model).
func (*windowGame) Layout(width, height int) (int, int) {
	return width, height
}
