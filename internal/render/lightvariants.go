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
