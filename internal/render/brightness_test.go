package render_test

import (
	"fmt"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/assets"
	"github.com/bmcszk/devilutionGo/internal/light"
	"github.com/bmcszk/devilutionGo/internal/render"
)

// rampPalette is a linear grey ramp: palette[i] = (i, i, i), so the tone-map's

// TestApplyBrightnessIdentity pins brightness 0 = byte-identity: the parity
// oracle at the default brightness must not move one byte.
func TestApplyBrightnessIdentity(t *testing.T) {
	t.Parallel()

	// given the grey ramp palette
	pal := rampPalette()

	// when the tone-map applies at brightness 0
	out := render.ApplyBrightness(pal, 0)

	// then every entry survives unchanged
	for i := range 256 {
		if out[i] != pal[i] {
			t.Fatalf("index %d = %v, want identity %v", i, out[i], pal[i])
		}
	}
}

// applyBrightnessReference is the literal upstream formula, for differential

// TestApplyBrightnessMatchesUpstreamFormula sweeps the ramp at several
// brightness settings against the reference arithmetic.
func TestApplyBrightnessMatchesUpstreamFormula(t *testing.T) {
	t.Parallel()

	// given the grey ramp palette
	pal := rampPalette()

	// when sweeping brightness levels over the ramp
	for _, brightness := range []int{0, 30, 60, 100} {
		t.Run(fmt.Sprintf("brightness=%d", brightness), func(t *testing.T) {
			t.Parallel()

			// when the tone-map applies
			out := render.ApplyBrightness(pal, brightness)

			// then every entry matches the upstream curve
			expectToneMappedRamp(t, out, brightness)
		})
	}

	// then the subtests cover all brightness levels
	if len([]int{0, 30, 60, 100}) != 4 {
		t.Fatal("expected 4 brightness levels")
	}
}

// TestWithBrightnessUploadsToneMappedVariants pins that a Presenter built with
// WithBrightness uploads tone-mapped variant rows: the texel buffer for one
// known index differs from the un-tonemapped upload, and brightness 0's buffer
// is byte-identical to the plain upload.
func TestWithBrightnessUploadsToneMappedVariants(t *testing.T) {
	t.Parallel()

	// given a palette state over the town-style ramp
	state := render.NewPaletteState(rampPalette(), light.MakeLightTables(light.DTYPE_TOWN), assets.BlendingFor(0))

	// when the variant textures are built at brightness 0 vs 60
	plain := render.PaletteVariantsTexture(&state.Variants, 256)
	brightened := state.WithBrightness(60)
	identityState := state.WithBrightness(0)
	bright := render.PaletteVariantsTexture(&brightened.Variants, 256)
	identity := render.PaletteVariantsTexture(&identityState.Variants, 256)

	// then brightness 0 is byte-identity against the plain upload
	for i := range plain {
		if identity[i] != plain[i] {
			t.Fatalf("brightness 0 texel %d = %d, want %d", i, identity[i], plain[i])
		}
	}

	// and brightness 60 changed at least the mid-ramp texel
	offset := 128 * 4 // shade 0, index 128
	if bright[offset] == plain[offset] {
		t.Fatalf("brightness 60 left mid-ramp texel %d unchanged", plain[offset])
	}
}

// curve is visible per entry.
func rampPalette() assets.Palette {
	var pal assets.Palette
	for i := range 256 {
		pal[i] = render.RGB{uint8(i), uint8(i), uint8(i)}
	}

	return pal
}

// expectToneMappedRamp asserts one brightness sweep over the whole ramp.
func expectToneMappedRamp(t *testing.T, out assets.Palette, brightness int) {
	t.Helper()

	for c := range 256 {
		want := applyBrightnessReference(uint8(c), brightness)
		if got := out[uint8(c)][0]; got != want {
			t.Fatalf("brightness %d index %d = %d, want %d", brightness, c, got, want)
		}
	}
}

// checks over the whole ramp.
func applyBrightnessReference(c uint8, brightness int) uint8 {
	a := -float64(brightness) / 100 * 2.0
	x := float64(c) / 255
	f := a*x*x + (1-a)*x
	if f < 0 {
		f = 0
	}
	if f > 1 {
		f = 1
	}

	return uint8(f*255 + 0.5)
}
