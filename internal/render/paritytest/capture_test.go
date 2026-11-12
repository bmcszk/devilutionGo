package paritytest_test

import (
	"bytes"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/render"
	"github.com/bmcszk/devilutionGo/internal/render/parity"
	"github.com/bmcszk/devilutionGo/internal/render/paritytest"
)

// TestCapturePair pins the harness contract through the real GPU pass, in one
// loop (a process hosts exactly one run loop): the pair of presentations is
// byte-identical (determinism), the dumped index bytes are the CPU rasteriser's
// surface (the PalSurface equivalent, pre-scale) and the palette row is the
// active variants row at fade 256.
func TestCapturePair(t *testing.T) {
	// given
	if !paritytest.GPUAvailable() {
		t.Skip("frame capture needs a GPU-capable run loop; headless CI has none")
	}

	world := paritytest.World()
	surface := paritytest.SoftwareSurface(world)
	state := paritytest.PaletteState()

	// when
	pair, err := paritytest.CapturePair(state, surface, paritytest.FrameAfterWarmup)
	if err != nil {
		t.Fatalf("capture: %v", err)
	}

	// then
	expectDumpShape(t, pair.A, render.WorldLogicalWidth, render.WorldLogicalHeight)
	expectFramePairEqual(t, pair)

	if !bytes.Equal(pair.B.Index, surface.Pixels) {
		t.Fatal("presented dump index bytes diverge from the CPU rasteriser surface")
	}

	if pair.B.Palette != paritytest.PaletteFromVariants(state) {
		t.Fatal("dump palette is not the active presenter variants row at fade 256")
	}
}

// TestSoftwareSurfaceIsDeterministic pins the CPU side of the harness without a
// display: the rasterised surface is identical across calls, so comparator
// tests keep a display-free oracle.
func TestSoftwareSurfaceIsDeterministic(t *testing.T) {
	// given
	world := paritytest.World()

	// when
	first := paritytest.SoftwareSurface(world)
	second := paritytest.SoftwareSurface(world)

	// then
	if !bytes.Equal(first.Pixels, second.Pixels) {
		t.Fatal("CPU rasterisation of the scripted world is not deterministic")
	}
}

// TestSeedFreeze pins the RNG freeze point: assembling the scripted world
// leaves the LCG parked at the fixture seed, so nothing on the assembly path
// draws random numbers.
func TestSeedFreeze(t *testing.T) {
	// given
	world := paritytest.World()

	// when
	state := paritytest.RandomStateAfter(world)

	// then
	if state != paritytest.Seed {
		t.Fatalf("LCG state %#x after world assembly, want the frozen seed %#x", state, paritytest.Seed)
	}
}

// expectDumpShape asserts the dump geometry matches the surface.
func expectDumpShape(t *testing.T, dump parity.FrameDump, width, height int) {
	t.Helper()

	if dump.Width != width || dump.Height != height {
		t.Fatalf("dump %dx%d, want %dx%d", dump.Width, dump.Height, width, height)
	}

	if len(dump.Index) != width*height {
		t.Fatalf("index buffer %d bytes, want %d", len(dump.Index), width*height)
	}
}

// expectFramePairEqual asserts two dumps carry identical index bytes and
// palettes: the determinism law.
func expectFramePairEqual(t *testing.T, pair paritytest.FramePair) {
	t.Helper()

	if !bytes.Equal(pair.A.Index, pair.B.Index) {
		t.Fatal("two presentations of the same scripted world diverged in index bytes")
	}

	if pair.A.Palette != pair.B.Palette {
		t.Fatal("two presentations of the same scripted world diverged in palette rows")
	}
}
