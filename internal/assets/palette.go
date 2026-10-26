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

// Predefined palettes for dungeon levels
var (
	// Level 1 (Cathedral) palette - placeholder, loaded from l1_2.pal
	L1Palette = Palette{}
	// Level 2 (Catacombs) palette
	L2Palette = Palette{}
	// Level 3 (Caves) palette
	L3Palette = Palette{}
	// Level 4 (Hell) palette
	L4Palette = Palette{}
	// Town palette
	TownPalette = Palette{}
)

// Palette represents a 256-color palette (RGB triples).
type Palette [256][3]uint8

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
