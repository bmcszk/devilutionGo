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
	"github.com/bmcszk/devilutionGo/internal/engine"
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

	flow *scene.FrontFlow

	heroTile func() render.TilePos // the kernel state's hero tile read
	// brightness reads the live Graphics/"Brightness Correction" level: the
	// Gamma slider writes it and a change must recompose (X:palette.cpp:253-261).
	brightness func() int
	compose    func(render.TilePos, int) (scene.TownScene, error)

	presenter  *render.Presenter
	surface    render.Surface
	shades     []uint8
	composed   render.TilePos
	hasCompose bool
	// composedBrightness is the gamma level of the last compose.
	composedBrightness int
	composeErr         error
	debug              string

	// lastState is the palette state of the last compose: the capture seam's
	// CPU-expansion read (the AC-4 demo gate re-derives the frame through
	// shade 0 without touching the GPU pipeline).
	lastState render.PaletteState
}

// newScenePresenter builds the presentation half of the boot: it reads the
// hero tile from heroTile and composes the town scene through compose on the
// first Draw and on every tile change. Image creation is illegal outside the
// run loop, so the presenter builds on the first Draw (paritytest
// captureGame's contract).
func newScenePresenter(
	screenshotDir string,
	flow *scene.FrontFlow,
	heroTile func() render.TilePos,
	brightness func() int,
	compose func(render.TilePos, int) (scene.TownScene, error),
) *scenePresenter {
	return &scenePresenter{
		debug: screenshotDir, flow: flow, heroTile: heroTile,
		brightness: brightness, compose: compose,
	}
}

// drawFront draws the active front screen into the town surface before the
// palette pass: the flow sits IN FRONT of the town render, which stays the
// terminal screen (devilGo-uuyh). Same-frame pass, no one-frame label lag.
func (p *scenePresenter) drawFront(flow *scene.FrontFlow) {
	if p.surface.Pixels == nil {
		return
	}

	if err := scene.ComposeFrontFrame(&p.surface, flow); err != nil {
		slog.Error("windowed boot: compose front frame", "err", err)
	}
}

// syncScene (re)composes the town scene when the kernel state's hero tile moved
// or the gamma level changed: the view follows the hero (devilGo-njzu) and a
// brightness edit invalidates the composed frame (Options slice S4b — the
// presenter otherwise recomposes only on a tile change, so a Gamma slider move
// with the hero standing still would never render). The ebiten-free half of the
// presenter; deterministic and testable headless.
func (p *scenePresenter) syncScene() {
	if p.composeErr != nil {
		return
	}

	hero := p.heroTile()
	brightness := p.brightnessLevel()
	if p.hasCompose && p.composed == hero && p.composedBrightness == brightness {
		return
	}

	composed, err := p.compose(hero, brightness)
	if err != nil {
		slog.Error("windowed boot: compose town scene", "err", err)
		p.composeErr = err

		return
	}

	p.presenter = render.NewPresenter(composed.State, render.WithFadeLevel(256))
	p.surface = composed.Surface
	p.lastState = composed.State
	p.shades = render.UniformShades(composed.Surface.Width, composed.Surface.Height, 0)
	p.composed = hero
	p.composedBrightness = brightness
	p.hasCompose = true
}

// brightnessLevel reads the live gamma level, defaulting to the upstream 0 when
// no reader is wired (Source/options.cpp:796).
func (p *scenePresenter) brightnessLevel() int {
	if p.brightness == nil {
		return engine.BrightnessMin
	}

	return p.brightness()
}

// composedBrightnessLevel reports the gamma level of the current frame: the
// black-box probe the re-apply test reads.
func (p *scenePresenter) composedBrightnessLevel() int {
	return p.composedBrightness
}

// composedHeroTile reports the tile the current scene was composed at: the
// black-box probe the live-view test reads.
func (p *scenePresenter) composedHeroTile() render.TilePos {
	return p.composed
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
func composeSceneAt(hero render.TilePos, brightness int) (scene.TownScene, error) {
	dirs := filepath.SplitList(os.Getenv("DG_ASSET_ARCHIVES"))
	if len(dirs) == 0 {
		return scene.TownScene{}, errors.New("DG_ASSET_ARCHIVES not set")
	}

	composed, err := scene.ComposeTownFrameAtBrightness(dirs, hero, brightness)
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

	p.drawFront(p.flow)
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
