//go:build windowed

package main

import (
	"errors"
	"fmt"
	"image"
	"image/png"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/bmcszk/devilutionGo/internal/diablo/controls"
	"github.com/bmcszk/devilutionGo/internal/present/scene"
	"github.com/bmcszk/devilutionGo/internal/render"
)

// displayEnv is the DISPLAY environment read; a variable for the tests.
var displayEnv = os.Getenv

// clockNow is the wall-clock read of the window layer; a variable so tests
// can script it.
var clockNow = time.Now

// x11SocketDirStat stats /tmp/.X11-unix; a variable for the tests.
var x11SocketDirStat = func() (os.FileInfo, error) { return os.Stat("/tmp/.X11-unix") }

// scenePresenter wires the composed town frame through the Presenter
// pipeline. The composed town scene is presentation state: the hero tile it
// composes at is derived each frame from the kernel state through heroTile —
// the same state the device lane reads, one source of truth, no shadow
// position and no view-side prediction (devilGo-njzu; ADR-018 law 2).
type scenePresenter struct {
	frame int

	heroTile func() render.TilePos // the kernel state's hero tile read
	compose  func(render.TilePos) (scene.TownScene, error)

	presenter  *render.Presenter
	surface    render.Surface
	shades     []uint8
	composed   render.TilePos
	hasCompose bool
	composeErr error
	debug      string
}

// newScenePresenter builds the presentation half of the boot: it reads the
// hero tile from heroTile and composes the town scene through compose on the
// first Draw and on every tile change. Image creation is illegal outside the
// run loop, so the presenter builds on the first Draw (paritytest
// captureGame's contract).
func newScenePresenter(
	screenshotDir string,
	heroTile func() render.TilePos,
	compose func(render.TilePos) (scene.TownScene, error),
) *scenePresenter {
	return &scenePresenter{debug: screenshotDir, heroTile: heroTile, compose: compose}
}

// syncScene (re)composes the town scene when the kernel state's hero tile
// moved: the view follows the hero, nothing else (devilGo-njzu). The
// ebiten-free half of the presenter; deterministic and testable headless.
func (p *scenePresenter) syncScene() {
	if p.composeErr != nil {
		return
	}

	hero := p.heroTile()
	if p.hasCompose && p.composed == hero {
		return
	}

	composed, err := p.compose(hero)
	if err != nil {
		slog.Error("windowed boot: compose town scene", "err", err)
		p.composeErr = err

		return
	}

	p.presenter = render.NewPresenter(composed.State, render.WithFadeLevel(256))
	p.surface = composed.Surface
	p.shades = render.UniformShades(composed.Surface.Width, composed.Surface.Height, 0)
	p.composed = hero
	p.hasCompose = true
}

// composedHeroTile reports the tile the current scene was composed at: the
// black-box probe the live-view test reads.
func (p *scenePresenter) composedHeroTile() render.TilePos {
	return p.composed
}

// windowKey maps ebiten's pressed keys onto the quit surface; the window
// layer is the only place raw input codes appear (v9xy input note).
func windowKey() keyEvent {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return keyEscape
	}

	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		return keyQ
	}

	return keySpace
}

// windowClosed reports the user closing the window.
func windowClosed() bool {
	return ebiten.IsWindowBeingClosed()
}

// windowMovement reads arrow keys and hands the stick decisions to the lane's
// submit site, which queues them for the next settled kernel pass. The window
// layer is the only place raw key codes appear (v9xy input note); it runs no
// kernel pass (review s6e F2).
func windowMovement(lane *deviceLane) {
	const push = controls.PushFirm

	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		lane.PushStick(push, controls.AxisDirection{X: controls.AxisXRight, Y: controls.AxisYNone})
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		lane.PushStick(push, controls.AxisDirection{X: controls.AxisXLeft, Y: controls.AxisYNone})
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		lane.PushStick(push, controls.AxisDirection{X: controls.AxisXNone, Y: controls.AxisYUp})
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		lane.PushStick(push, controls.AxisDirection{X: controls.AxisXNone, Y: controls.AxisYDown})
	}
}

// composeSceneAt composes the town scene at the given hero tile (devilGo-njzu):
// the live view's compose site, reading the same retail archives the first
// scene slice did (DG_ASSET_ARCHIVES).
func composeSceneAt(hero render.TilePos) (scene.TownScene, error) {
	dirs := filepath.SplitList(os.Getenv("DG_ASSET_ARCHIVES"))
	if len(dirs) == 0 {
		return scene.TownScene{}, errors.New("DG_ASSET_ARCHIVES not set")
	}

	composed, err := scene.ComposeTownFrameAtHero(dirs, hero)
	if err != nil {
		return scene.TownScene{}, fmt.Errorf("compose town scene: %w", err)
	}

	return composed, nil
}

// draw presents one frame: compose lazily at the kernel state's hero tile
// (devilGo-njzu), run the palette pass, scale to the window.
func (p *scenePresenter) draw(screen *ebiten.Image) {
	p.syncScene()
	if !p.hasCompose || p.presenter == nil {
		return
	}

	p.presenter.Draw(p.surface, p.shades, render.KeyTile)
	p.presenter.ScaleToWindow(screen, screen.Bounds().Dx(), screen.Bounds().Dy())

	// screenshot cadence (devilGo-1awt): every N frames
	if p.debug != "" {
		p.frame++
		if p.frame%60 == 0 { // every 60 frames ~ 3 seconds at 20 TPS
			p.capture(screen)
		}
	}
}

func (p *scenePresenter) capture(screen *ebiten.Image) {
	screenImg := image.NewRGBA(screen.Bounds())
	for y := range screenImg.Bounds().Dy() {
		for x := range screenImg.Bounds().Dx() {
			screenImg.Set(x, y, screen.At(x, y))
		}
	}
	path := filepath.Join(p.debug, fmt.Sprintf("frame%04d.png", p.frame))
	if err := os.MkdirAll(p.debug, 0o700); err != nil {
		slog.Error("windowed boot: mkdir screenshots", "err", err)
		return
	}
	f, err := os.Create(path)
	if err != nil {
		slog.Error("windowed boot: create screenshot", "err", err)
		return
	}
	if err := png.Encode(f, screenImg); err != nil {
		slog.Error("windowed boot: encode screenshot", "err", err)
	}

	if err := f.Close(); err != nil {
		slog.Error("windowed boot: close screenshot", "err", err)
	}
}
