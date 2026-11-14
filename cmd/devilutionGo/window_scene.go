//go:build windowed

package main

import (
	"fmt"
	"image"
	"image/png"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/bmcszk/devilutionGo/internal/present/scene"
	"github.com/bmcszk/devilutionGo/internal/render"
)

// windowBootConfig is the slice-level wiring of the window boot: the frame
// step, the screenshot cadence (devilGo-1awt) and the window title.

// clockNow is the wall-clock read of the window layer; a variable so tests
// can script it.
var clockNow = time.Now

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

// displayEnv is the DISPLAY environment read; a variable for the tests.
var displayEnv = os.Getenv

// x11SocketDirStat stats /tmp/.X11-unix; a variable for the tests.
var x11SocketDirStat = func() (os.FileInfo, error) { return os.Stat("/tmp/.X11-unix") }

// newScenePresenter builds the presentation half of the boot: the composed
// first-scene surface (devilGo-6asa) presented through the palette pass.
// Image creation is illegal outside the run loop, so the presenter builds on
// the first Draw (paritytest captureGame's contract).
func newScenePresenter(_ []string, screenshotDir string) *scenePresenter {
	present := &scenePresenter{debug: screenshotDir}

	return present
}

// scenePresenter wires the composed town frame through the Presenter
// pipeline; the slice renders the static first scene (devilGo-6asa output).
type scenePresenter struct {
	frame int

	presenter  *render.Presenter
	surface    render.Surface
	shades     []uint8
	composeErr error
	debug      string
}

// Update builds the presenter on the first frame (image creation needs the
// live loop) and captures the periodic screenshot (devilGo-1awt).
func (p *scenePresenter) ensureBuilt() {
	if p.presenter != nil || p.composeErr != nil {
		return
	}

	composed, err := composeFirstScene()
	if err != nil {
		slog.Error("windowed boot: compose first scene", "err", err)
		p.composeErr = err

		return
	}

	p.presenter = render.NewPresenter(composed.State, render.WithFadeLevel(256))
	p.surface = composed.Surface
	p.shades = render.UniformShades(composed.Surface.Width, composed.Surface.Height, 0)
}

// draw presents one frame: compose lazily, run the palette pass, scale to
// the window; capture periodic screenshots (devilGo-1awt).
func (p *scenePresenter) draw(screen *ebiten.Image) {
	p.ensureBuilt()
	if p.presenter == nil {
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
	slog.Debug("capture", "screenBounds", screen.Bounds())
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
	f.Close()
}

// composeFirstScene builds the static first scene for the windowed boot
// (devilGo-6asa + 6asa adapter).
func composeFirstScene() (scene.TownScene, error) {
	// Reuse the present scene's ComposeTownFrame with the same archive
	// search paths (DG_ASSET_ARCHIVES).
	dirs := filepath.SplitList(os.Getenv("DG_ASSET_ARCHIVES"))
	if len(dirs) == 0 {
		return scene.TownScene{}, fmt.Errorf("DG_ASSET_ARCHIVES not set")
	}

	return scene.ComposeTownFrame(dirs)
}
