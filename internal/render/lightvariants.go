package render

import (
	"github.com/bmcszk/devilutionGo/internal/assets"
	"github.com/bmcszk/devilutionGo/internal/light"
)

// PaletteState is the engine's live palette: the palette colours, the 16 light
// variants derived from them, and the blend look-up table. Cycling a fountain or
// lava range rotates all three together (palette.cpp:52-90).
type PaletteState struct {
	Palette  assets.Palette
	Variants Variants
	LUT      BlendLUT
	tables   light.LightTables
}

// NewPaletteState derives the light variants and the blend LUT for a loaded
// palette, the Go equivalent of LoadPaletteAndInitBlending (palette.cpp:188-200).
func NewPaletteState(pal assets.Palette, tables light.LightTables, blending assets.Blending) PaletteState {
	return PaletteState{
		Palette:  pal,
		Variants: buildVariants(pal, tables.Tables),
		LUT:      blendLUTFor(pal, blending),
		tables:   tables,
	}
}

// CyclePalette rotates the colour range from..to forward by one, the fountain
// animation cycle (CycleColors, palette.cpp:52-64).
//
// The rotation covers the palette, every row and every column of the blend LUT (a
// rotated index means a rotated LUT axis), then re-derives the light variants from
// the rotated palette.
func (s *PaletteState) CyclePalette(from, to int) {
	s.rotateColors(from, to, rotateForward)
}

// CyclePaletteReverse rotates the colour range from..to backward by one, the
// reverse animation cycle (CycleColorsReverse, palette.cpp:72-84).
func (s *PaletteState) CyclePaletteReverse(from, to int) {
	s.rotateColors(from, to, rotateBackward)
}

// WithBrightness returns a copy tone-mapped by brightness percent, upstream's
// ApplyGlobalBrightness (palette.cpp:96-110):
// a = -(brightness/100)*2; f(x) = a*x^2 + (1-a)*x over x = c/255;
// out = uint8(clamp(f,0,1)*255 + 0.5). Brightness 0 is byte-identity, so the
// parity oracle stays green at the default (UpdateSystemPalette applies the
// map to the whole palette before the light variants derive from it).
func (s PaletteState) WithBrightness(percent int) PaletteState {
	s.Palette = ApplyBrightness(s.Palette, percent)
	s.Variants = buildVariants(s.Palette, s.tables.Tables)

	return s
}

// ApplyBrightness tone-maps one palette by brightness percent with the
// upstream curve (ApplyGlobalBrightness, palette.cpp:96-110); 0 is identity.
func ApplyBrightness(pal assets.Palette, percent int) assets.Palette {
	a := -float64(percent) / 100 * 2.0

	var out assets.Palette
	for i, color := range pal {
		out[i] = [3]uint8{
			brightnessComponent(color[0], a),
			brightnessComponent(color[1], a),
			brightnessComponent(color[2], a),
		}
	}

	return out
}

// brightnessComponent maps one channel through the brightness curve
// (ApplyGlobalBrightness, palette.cpp:104-108).
func brightnessComponent(c uint8, a float64) uint8 {
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

// blendLUTFor builds the LUT for a level's blend ranges, the branch
// LoadPaletteAndInitBlending takes per dungeon type (palette.cpp:188-200).
func blendLUTFor(pal assets.Palette, blending assets.Blending) BlendLUT {
	if blending.Full {
		return BuildBlendingLUT(pal, noSkip, noSkip)
	}

	return BuildBlendingLUT(pal, blending.SkipFrom, blending.SkipTo)
}

// buildVariants maps every palette index through every light table:
// variant[shade][index] = palette[tables[shade][index]]
// (light_render.hpp:30-33 adjustColor, applied per entry).
func buildVariants(pal assets.Palette, tables [numShades][paletteSize]uint8) Variants {
	var variants Variants
	for shade := range numShades {
		for index := range paletteSize {
			variants[shade][index] = RGB(pal[tables[shade][index]])
		}
	}

	return variants
}

// rotateColors applies one rotation to the palette colours, both LUT axes and the
// light variants.
func (s *PaletteState) rotateColors(from, to int, dir rotation) {
	if from >= to || to >= paletteSize {
		return
	}

	rotateRange(s.Palette[:], from, to, dir)
	rotateRange(s.LUT[:], from, to, dir)

	for row := range paletteSize {
		rotateRange(s.LUT[row][:], from, to, dir)
	}

	s.Variants = buildVariants(s.Palette, s.tables.Tables)
}

// rotateRange rotates elements from..to by one position in dir, the equivalent of
// std::rotate over a one-element-shifted range (palette.cpp:54, :74).
func rotateRange[T any](elems []T, from, to int, dir rotation) {
	switch dir {
	case rotateForward:
		first := elems[from]
		copy(elems[from:to], elems[from+1:to+1])
		elems[to] = first
	case rotateBackward:
		last := elems[to]
		copy(elems[from+1:to+1], elems[from:to])
		elems[from] = last
	default:
		panic("render: unknown rotation direction")
	}
}
