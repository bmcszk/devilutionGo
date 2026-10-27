// Package assets implements decoders for Diablo asset formats (CLX, CL2,
// CEL, PCX, TRN, palette) into engine-ready indexed pixel data.
//
// All decoders are pure Go and produce 256-color indexed []byte frames;
// palette expansion happens in the render layer, not here.
package assets

import (
	"fmt"
	"io"
)

// dungeon_type ordinals (levels/gendung_defs.hpp:15-26); this package only
// needs the ones with palette special cases.
const (
	levelTypeTown  = 0
	levelTypeCaves = 3
	levelTypeNest  = 5
	levelTypeCrypt = 6
)

// Palette represents a 256-color palette (RGB triples).
type Palette [256][3]uint8

// Blending is the palette range GenerateBlendedLookupTable leaves unblended:
// caves and crypt skip 1..31, the nest skips 1..15, everything else blends the
// full palette (palette.cpp:188-195).
type Blending struct {
	SkipFrom int
	SkipTo   int
	Full     bool
}

// AssetSource is the asset lookup the palette and TRN loaders need;
// *engine/assets.Stack satisfies it.
type AssetSource interface {
	ReadFile(name string) ([]byte, bool)
}

// LevelPaletteName derives the .pal path for a level type (palette.cpp:196-229).
// Town and crypt are fixed; the nest adds !alternateNestArt to rv; every other
// type embeds its dungeon_type ordinal and rv.
//
//nolint:revive // the alternateNestArt flag is upstream's signature (options Graphics.alternateNestArt).
func LevelPaletteName(levelType, rv int, alternateNestArt bool) string {
	switch levelType {
	case levelTypeTown:
		return `levels\towndata\town.pal`
	case levelTypeCrypt:
		return `nlevels\l5data\l5base.pal`
	case levelTypeNest:
		if !alternateNestArt {
			rv++
		}

		return fmt.Sprintf(`nlevels\l6data\l6base%d.pal`, rv)
	}

	return fmt.Sprintf(`levels\l%ddata\l%d_%d.pal`, levelType, levelType, rv)
}

// BlendingFor returns the blend ranges for a level type (palette.cpp:188-195).
func BlendingFor(levelType int) Blending {
	switch levelType {
	case levelTypeCaves, levelTypeCrypt:
		return Blending{SkipFrom: 1, SkipTo: 31}
	case levelTypeNest:
		return Blending{SkipFrom: 1, SkipTo: 15}
	}

	return Blending{Full: true}
}

// LoadRndLvlPal picks the level palette path and blend ranges (palette.cpp:196-229).
// nextInt is the injected RNG (engine.RandomIntBetween at the call site) and is
// only drawn for level types whose palette is random (1..4).
func LoadRndLvlPal(
	_ AssetSource,
	levelType int,
	alternateNestArt bool,
	nextInt func(lo, hi int32) int32,
) (path string, blending Blending, err error) {
	if levelType < levelTypeTown || levelType > levelTypeCrypt {
		return "", Blending{}, fmt.Errorf("palette: unknown level type %d", levelType)
	}

	rv := 0
	if levelPaletteIsRandom(levelType) {
		rv = int(nextInt(1, 4))
	}

	return LevelPaletteName(levelType, rv, alternateNestArt), BlendingFor(levelType), nil
}

// LoadPaletteFromStack loads a palette from the asset stack; a missing asset is
// an error, never a nil palette.
func LoadPaletteFromStack(src AssetSource, path string) (*Palette, error) {
	data, ok := src.ReadFile(path)
	if !ok {
		return nil, fmt.Errorf("palette: asset %s not found", path)
	}

	p, err := LoadPaletteFromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("palette %s: %w", path, err)
	}

	return p, nil
}

// LoadPalette loads a palette from a .pal file (768 bytes: 256*3 RGB).
// In devilutionX: palette.cpp LoadPalette reads from MPQ.
func LoadPalette(r io.Reader) (*Palette, error) {
	var p Palette

	data := make([]byte, 768)

	n, err := io.ReadFull(r, data)
	if err != nil {
		return nil, fmt.Errorf("palette: failed to read 768 bytes: %w", err)
	}

	if n != 768 {
		return nil, fmt.Errorf("palette: expected 768 bytes, got %d", n)
	}

	for i := range 256 {
		p[i][0] = data[i*3]
		p[i][1] = data[i*3+1]
		p[i][2] = data[i*3+2]
	}

	return &p, nil
}

// LoadPaletteFromBytes loads a palette from 768-byte slice.
func LoadPaletteFromBytes(data []byte) (*Palette, error) {
	if len(data) < 768 {
		return nil, fmt.Errorf("palette: data too short (need 768, have %d)", len(data))
	}

	var p Palette
	for i := range 256 {
		p[i][0] = data[i*3]
		p[i][1] = data[i*3+1]
		p[i][2] = data[i*3+2]
	}

	return &p, nil
}

// ExpandPalette applies a palette to indexed pixel data, producing RGBA bytes.
// This is used by the render layer (not the decoder).
func (p *Palette) ExpandPalette(indices []byte) []byte {
	rgba := make([]byte, len(indices)*4)
	for i, idx := range indices {
		r := p[idx][0]
		g := p[idx][1]
		b := p[idx][2]
		rgba[i*4] = r
		rgba[i*4+1] = g
		rgba[i*4+2] = b
		rgba[i*4+3] = 255 // opaque
	}

	return rgba
}

// CycleColors rotates the [from, to] range one step forward, leaving the rest
// of the palette untouched (palette.cpp:52-69). Only the logical palette is in
// scope here; the system palette and transparency LUTs belong to the render seam.
func CycleColors(p *Palette, from, to int) {
	if !validCycleRange(from, to) {
		return
	}

	first := p[from]
	copy(p[from:to], p[from+1:to+1])
	p[to] = first
}

// CycleColorsReverse rotates the [from, to] range one step backward
// (palette.cpp:71-88). Same scope note as CycleColors.
func CycleColorsReverse(p *Palette, from, to int) {
	if !validCycleRange(from, to) {
		return
	}

	last := p[to]
	copy(p[from+1:to+1], p[from:to])
	p[from] = last
}

// PaletteUpdateForCaves cycles the lava range every tick (palette.cpp:363-367);
// frame is unused because the caves animation has no delay counter.
func PaletteUpdateForCaves(p *Palette, _ int) {
	CycleColors(p, 1, 31)
}

// PaletteUpdateForCrypt cycles the lava every other frame and the glow every
// frame (palette.cpp:372-386). The inner delayLava=false assignment is a no-op;
// the toggle after the glow cycle is what alternates the lava.
func PaletteUpdateForCrypt(p *Palette, frame int) {
	if frame%2 == 0 {
		CycleColorsReverse(p, 1, 15)
	}

	CycleColorsReverse(p, 16, 31)
}

// PaletteUpdateForHive cycles the pond and bubble ranges every third frame
// (palette.cpp:389-405).
func PaletteUpdateForHive(p *Palette, frame int) {
	if frame%3 != 2 {
		return
	}

	CycleColorsReverse(p, 1, 8)
	CycleColorsReverse(p, 9, 15)
}

// levelPaletteIsRandom reports whether LoadRndLvlPal draws rv: town and crypt
// return before RandomIntBetween (palette.cpp:198-206).
func levelPaletteIsRandom(levelType int) bool {
	return levelType != levelTypeTown && levelType != levelTypeCrypt
}

func validCycleRange(from, to int) bool {
	return from >= 0 && to < 256 && from < to
}
