package assets

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

// FontGlyphCount is the number of glyphs in a CLX font sheet: devilutionX
// indexes the sheet by the byte value of the character, so every sheet carries
// exactly 256 frames (text_render.cpp:653 `currentFont.glyph(frame)`).
const FontGlyphCount = 256

// fontFrameHeaderSize is the CLX frame header printed by clx_sprite.hpp:15-27:
// pixel-data offset, width, then height, all uint16.
const fontFrameHeaderSize = 6

// Glyph is one decoded font glyph as 256-color indexed pixels. A glyph whose
// Width is zero has no pixel stream at all — 229 of the 256 glyphs in the
// shipped 12-05 sheet are blank (internal/testdata/assets/fonts/README.md).
type Glyph struct {
	Width  int
	Height int
	// Pixels is row-major indexed pixel data of size Width*Height; nil for a
	// zero-width glyph.
	Pixels []byte
}

// Font is a decoded CLX font sheet addressed by character code. The UI lane
// asks it for glyph widths and line widths; palette expansion stays in the
// render layer.
type Font struct {
	glyphs []Glyph
	digest string
}

// DecodeFontSheet parses a CLX font sheet and decodes all FontGlyphCount glyphs.
//
// Unlike DecodeCLX, a zero-width frame is legal here: blank glyphs carry a
// 6-byte header and no pixel stream, which the sprite decoder rejects because a
// sprite is never empty ((clx_sprite.hpp:15-27, text_render.cpp:653).
func DecodeFontSheet(data []byte) (*Font, error) {
	glyphs, err := decodeFontGlyphs(data)
	if err != nil {
		return nil, err
	}

	return &Font{glyphs: glyphs, digest: glyphDigest(glyphs)}, nil
}

// Glyphs returns the decoded glyph table, one entry per character code.
func (f *Font) Glyphs() []Glyph {
	if f == nil {
		return nil
	}
	return f.glyphs
}

// Glyph returns the glyph of one character code, or a zero-width glyph when the
// code is outside the sheet.
func (f *Font) Glyph(code rune) Glyph {
	if f == nil || code < 0 || int(code) >= len(f.glyphs) {
		return Glyph{}
	}
	return f.glyphs[code]
}

// Width returns the pixel width of one character code.
func (f *Font) Width(code rune) int {
	return f.Glyph(code).Width
}

// Digest is the sha256 of every glyph's pixels concatenated in sheet order, the
// value the committed `<sheet>.sha256` golden carries.
func (f *Font) Digest() string {
	if f == nil {
		return ""
	}
	return f.digest
}

// LineWidth is the ported text_render.cpp:631-659 GetLineWidth: the sum of each
// glyph width plus spacing, minus the trailing spacing. An empty line is zero
// wide. Characters outside the sheet contribute nothing because upstream
// addresses the glyph ROW by code>>8 and the frame by code&0xFF; this Go
// version indexes by the full rune and returns zero width for codes >= 0x100
// until multi-row fonts land (text_render.cpp:640-652).
func (f *Font) LineWidth(text string, spacing int) int {
	if f == nil {
		return 0
	}

	width := 0
	for _, code := range text {
		width += f.Width(code) + spacing
	}
	if width == 0 {
		return 0
	}
	return width - spacing
}

// decodeFontGlyphs walks the frame list and decodes each glyph.
func decodeFontGlyphs(data []byte) ([]Glyph, error) {
	numFrames, offsets, err := parseFrameOffsets(data, "CLX font")
	if err != nil {
		return nil, err
	}
	if numFrames != FontGlyphCount {
		return nil, fmt.Errorf("assets: CLX font holds %d frames, want %d", numFrames, FontGlyphCount)
	}

	glyphs := make([]Glyph, 0, numFrames)
	for i := range numFrames {
		glyph, err := decodeFontGlyph(data[offsets[i]:offsets[i+1]])
		if err != nil {
			return nil, fmt.Errorf("assets: CLX font glyph %d (0x%02x): %w", i, i, err)
		}
		glyphs = append(glyphs, *glyph)
	}
	return glyphs, nil
}

// decodeFontGlyph decodes one glyph frame, tolerating a zero width.
func decodeFontGlyph(frame []byte) (*Glyph, error) {
	if len(frame) < fontFrameHeaderSize {
		return nil, fmt.Errorf("frame too short for CLX header: %d bytes", len(frame))
	}

	headerSize := int(binary.LittleEndian.Uint16(frame[0:2]))
	width := int(binary.LittleEndian.Uint16(frame[2:4]))
	height := int(binary.LittleEndian.Uint16(frame[4:6]))

	if headerSize < fontFrameHeaderSize || headerSize > len(frame) {
		return nil, fmt.Errorf("invalid CLX frame header size %d (frame %d bytes)", headerSize, len(frame))
	}
	if width < 0 || height < 1 {
		return nil, fmt.Errorf("invalid CLX font glyph dimensions %dx%d", width, height)
	}
	if width == 0 {
		return &Glyph{Height: height}, nil
	}

	pixels, err := decodeCl2Pixels(frame[headerSize:], width, height)
	if err != nil {
		return nil, err
	}
	return &Glyph{Width: width, Height: height, Pixels: pixels}, nil
}

// glyphDigest hashes the concatenated glyph pixels in sheet order.
func glyphDigest(glyphs []Glyph) string {
	hash := sha256.New()
	for _, glyph := range glyphs {
		hash.Write(glyph.Pixels)
	}
	return hex.EncodeToString(hash.Sum(nil))
}
