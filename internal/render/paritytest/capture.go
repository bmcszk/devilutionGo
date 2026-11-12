// Package paritytest renders scripted parity frames through the real
// presentation pass and hands them to the comparator as FrameDumps (devilGo-7aul
// slice 1). Test-only: it exists for parity capture tests and the goldens
// parity-record subcommand (`go run ./tools/goldens parity-record DIR`).
//
// Dump law (devilutionX e1ad0c2b6, dx.h GlobalBackBuffer + capture.cpp): the
// dump's index bytes are the logical 640x480 index surface pre-scale —
// render.Surface.Pixels, the PalSurface equivalent — never the window-scaled
// image and never the palette-expanded shader output. The palette row is the
// active palette variants row at fade 256 (cycling off). The presentation pass
// still runs on every captured frame, so a capture fails if the real GPU
// pipeline cannot present the frame.
package paritytest

import (
	"errors"
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/bmcszk/devilutionGo/internal/render"
	"github.com/bmcszk/devilutionGo/internal/render/parity"
)

const (
	// warmupFrames is how many presentation frames run before the captured one,
	// so shader upload and allocation frames never land in a golden.
	warmupFrames = 3
	// frameGap is the presentation-frame distance between the pair dumps.
	frameGap = 2
	// Seed is the fixed LCG seed of the scripted world: SetRndSeed parks the
	// engine RNG here immediately before world assembly, so the frame is
	// reproducible.
	Seed = 0x5EED071A
)

// FrameAfterWarmup selects the first frame after the warm-up window: the
// deterministic frame index golden captures pick, not a busiest-frame
// heuristic.
var FrameAfterWarmup = warmupFrames

// FramePair carries the dumps of two presentation frames rendered from the same
// world surface: the determinism evidence.
type FramePair struct {
	A, B parity.FrameDump
}

// captureGame builds the presenter inside the live loop (image creation is
// illegal outside it), presents the surface every frame and dumps the pair.
type captureGame struct {
	presenter *render.Presenter
	state     render.PaletteState
	surface   render.Surface
	targets   [2]int
	frame     int
	dumps     chan parity.FrameDump
	done      chan error
}

// CapturePair presents surface for two target frames and returns both dumps.
// Ebitengine hosts exactly one run loop per process, so every caller shares one
// call per test binary (presenter_test's inRunLoop contract).
func CapturePair(state render.PaletteState, surface render.Surface, first int) (FramePair, error) {
	game := &captureGame{
		state:   state,
		surface: surface,
		targets: [2]int{first, first + frameGap},
		dumps:   make(chan parity.FrameDump, 2),
		done:    make(chan error, 1),
	}

	ran, loopErr := runLoop(game)
	if !ran {
		return FramePair{}, loopErr
	}

	return FramePair{A: <-game.dumps, B: <-game.dumps}, nil
}

// Capture presents surface for one target frame and returns its dump: the
// single-frame convenience for the goldens parity-record subcommand, one run
// loop per process.
func Capture(state render.PaletteState, surface render.Surface, frame int) (parity.FrameDump, error) {
	game := &captureGame{
		state:   state,
		surface: surface,
		targets: [2]int{frame, frame},
		dumps:   make(chan parity.FrameDump, 1),
		done:    make(chan error, 1),
	}

	ran, loopErr := runLoop(game)
	if !ran {
		return parity.FrameDump{}, loopErr
	}

	return <-game.dumps, nil
}

// GPUAvailable reports whether this machine can run the Ebitengine capture
// loop: the presenter's shader needs a live GL context, so headless CI skips
// (same DISPLAY contract as presenter_test's gpuAvailable).
func GPUAvailable() bool {
	if os.Getenv("DISPLAY") == "" {
		return false
	}

	if _, err := os.Stat("/tmp/.X11-unix"); err != nil {
		return false
	}

	return true
}

// DumpFrom builds the FrameDump of one presented frame: the index bytes are the
// presented surface itself (the PalSurface equivalent, pre-scale) and the
// palette row is the active variants row at fade 256.
func DumpFrom(surface render.Surface, state render.PaletteState) parity.FrameDump {
	return parity.FrameDump{
		Width:   surface.Width,
		Height:  surface.Height,
		Index:   surface.Pixels,
		Palette: PaletteFromVariants(state),
	}
}

// Update builds the presenter on the first frame (image creation needs the live
// loop), dumps on each target frame and stops after the second.
func (g *captureGame) Update() error {
	if g.presenter == nil {
		g.presenter = render.NewPresenter(g.state)
	}

	if g.frame == g.targets[0] || g.frame == g.targets[1] {
		g.dumps <- DumpFrom(g.surface, g.state)
	}

	g.frame++
	if g.frame > g.targets[1] {
		return ebiten.Termination
	}

	return nil
}

// Draw presents one frame and blits the pre-scale target to the screen.
func (g *captureGame) Draw(screen *ebiten.Image) {
	if g.presenter == nil {
		return
	}

	g.presenter.Draw(g.surface, nil, render.KeyTile)
	screen.DrawImage(g.presenter.Target(), nil)
}

// Layout sizes the window to the world logical surface.
func (*captureGame) Layout(_, _ int) (width, height int) {
	return render.WorldLogicalWidth, render.WorldLogicalHeight
}

// runLoop runs one Ebitengine loop to completion; loopErr carries the failure
// and ran holds when the loop finished without one.
func runLoop(game *captureGame) (ran bool, loopErr error) {
	go func() { game.done <- ebiten.RunGame(game) }()

	err := <-game.done
	if err != nil && !errors.Is(err, ebiten.Termination) {
		return false, fmt.Errorf("paritytest: run loop: %w", err)
	}

	return true, nil
}
