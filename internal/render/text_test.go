package render_test

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/assets"
	"github.com/bmcszk/devilutionGo/internal/render"
)

const (
	fontFixtures = "../testdata/assets/fonts"
	fontSheet    = "12-05"
	spacing      = 1
	// lineHeight is the 12-05 glyph row height: every frame on the sheet is 20
	// pixels tall (internal/testdata/assets/fonts/12-05.sha256).
	lineHeight = 20
)

// TestTextAdvanceWidth pins single-glyph advance: Font.Width on the golden
// glyph dumps d0 (8x20), de (10x20) and the zero-width 00.
func TestTextAdvanceWidth(t *testing.T) {
	// the decoded 12-05 sheet carries the frame-header widths
	font := loadTestFont(t)

	tests := []struct {
		name string
		code rune
		want int
	}{
		{"d0", 0xD0, 8},
		{"de", 0xDE, 10},
		{"00", 0x00, 0},
		{"ascii-a-is-blank", 'a', 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// when the glyph width is measured
			got := font.Width(test.code)

			// then it matches the committed frame header
			if got != test.want {
				t.Fatalf("Width(0x%02X) = %d, want %d", test.code, got, test.want)
			}
		})
	}
}

// TestFontLineWidthParity pins GetLineWidth (text_render.cpp:631-659) against
// hand-summed golden literals over the d0..ea glyph set, including the ZWSP
// skip (cpp:46 zero width, cpp:1012 breakable whitespace) and the '\n' stop
// (cpp:653-654). Strings are valid UTF-8 over the Latin-1 accented row
// U+00D0..U+00EA (the sheet's byte codes 0xd0..0xea); a lone 0xd0 byte would
// range as U+FFFD, not U+00D0.
func TestFontLineWidthParity(t *testing.T) {
	// the decoded sheet feeds hand-summed golden literals over sample strings
	font := loadTestFont(t)

	tests := []struct {
		name string
		text string
		want int
	}{
		{"empty", "", 0},
		{"single", "\u00d0", 8},
		{"two", "\u00d0\u00de", 8 + 10 + spacing},
		{"three", "\u00d0\u00d5\u00de", 8 + 3 + 10 + 2*spacing},
		{"five", "\u00d0\u00de\u00d2\u00d8\u00e9", 8 + 10 + 6 + 9 + 10 + 4*spacing},
		// d0..d9 widths 8 8 6 9 7 3 6 7 9 3, plus nine inner spacings
		{"ten", "\u00d0\u00d1\u00d2\u00d3\u00d4\u00d5\u00d6\u00d7\u00d8\u00d9", 75},
		{"narrow-repeat", "\u00d5\u00df\u00d9", 3 + 3 + 3 + 2*spacing},
		{"mixed-ascii-blank", "ab\u00d0", 0 + 0 + 8 + 2*spacing},
		{"wide-tail", "\u00d3\u00de\u00ea", 9 + 10 + 10 + 2*spacing},
		{"zwsp-solo", "\u200b", 0},
		{"zwsp-between", "\u200b\u00d0", 8},
		{"zwsp-run", "\u200b\u200b", 0},
		{"zwsp-word", "\u00d0\u200b\u00de", 8 + 10 + spacing},
		{"ascii-then-zwsp", "a\u200b", 0},
		{"newline-solo", "\n", 0},
		{"newline-first", "\n\u00d0", 0},
		{"newline-last", "\u00d0\n", 8},
		{"newline-middle", "\u00d0\n\u00de", 8},
		{"zwsp-then-newline", "\u200b\n", 0},
		{"crlf", "\u00d0\r\n", 9}, // \r is a zero-width glyph that still takes spacing
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// when the line width is measured
			got := font.LineWidth(test.text, spacing)

			// then it equals the hand-summed golden literal
			if got != test.want {
				t.Fatalf("LineWidth(%q) = %d, want %d", test.text, got, test.want)
			}
		})
	}
}

// TestWrapLayout pins WordWrap (text_render.cpp:783-820): breakable whitespace
// is {' ', U+200B} (cpp:1012), '\n' forces a break (cpp:788), and a word alone
// longer than the line hard-splits at the overflowing character (cpp:803-807;
// upstream 99bf0e4e3 changed keep-whole to hard-break).
func TestWrapLayout(t *testing.T) {
	// the decoded sheet drives the layout table
	font := loadTestFont(t)

	tests := []struct {
		name          string
		text          string
		rectWidth     int
		wantLineCount int
		wantLastWidth int
		wantLastX     int
		wantLastY     int
	}{
		{"fits-exactly", "\u00d0\u00de", 19, 1, 19, 20, 0},
		{"one-under", "\u00d0\u00de", 20, 1, 19, 20, 0},
		{"two-words-fit", "\u00d0 \u00de", 200, 1, 20, 21, 0},
		{"two-words-split", "\u00d0 \u00de", 10, 2, 10, 11, lineHeight},
		{"line-one-ends-at-space", "\u00d0\u00d3 \u00de", 20, 2, 10, 11, lineHeight},
		{"second-word-wider", "\u00d5 \u00de", 12, 2, 10, 11, lineHeight},
		{"third-word-forces-split", "\u00d0 \u00d5 \u00de", 14, 2, 10, 11, lineHeight},
		{"newline-two-lines", "\u00d0\n\u00de", 200, 2, 10, 11, lineHeight},
		{"newline-then-split", "\u00d0\n\u00d0 \u00de", 10, 3, 10, 11, 2 * lineHeight},
		// 8+1+8+1+6+1+9: the d2 glyph overflows a 10-wide line, so the word
		// hard-splits there and again before d3 (cpp:803-807).
		{"hard-split", "\u00d0\u00d1\u00d2\u00d3", 10, 4, 9, 10, 3 * lineHeight},
		{"hard-split-exact", "\u00d0\u00d1\u00d2", 24, 1, 24, 25, 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// when the string draws into a rect of that width
			surface := render.NewSurface(400, 200)
			result := render.DrawString(surface, font, test.text,
				render.Rect{X: 0, Y: 0, W: test.rectWidth, H: 40}, render.Options{})

			// then the line count, last-line width and final cursor match WordWrap
			assertLayout(t, result, test.wantLineCount, test.wantLastWidth, test.wantLastX, test.wantLastY)
		})
	}
}

// TestDrawStringPlainGlyphPixels pins the draw loop against the committed
// golden d0 glyph: its 52 opaque pixels land inside the glyph box, the frame's
// transparent rows draw nothing (DrawFont, text_render.cpp:269-280).
func TestDrawStringPlainGlyphPixels(t *testing.T) {
	// given a blank surface and the decoded sheet
	font := loadTestFont(t)
	surface := render.NewSurface(32, 24)

	// when the golden d0 glyph draws at (2, 2)
	result := render.DrawString(surface, font, "\u00d0",
		render.Rect{X: 2, Y: 2, W: 8, H: 20}, render.Options{})

	// then the 52 opaque sheet pixels appear inside the glyph box: d0's opaque
	// rows are 5..13 of the 20-row frame, the rest is transparent
	if got := countOpaque(surface); got != 52 {
		t.Fatalf("opaque pixels = %d, want 52", got)
	}
	assertOpaqueWithin(t, surface, render.Rect{X: 2, Y: 2, W: 8, H: 20})
	if result.LastCursorX != 11 || result.LastCursorY != 0 {
		t.Fatalf("cursor = (%d, %d), want (11, 0)", result.LastCursorX, result.LastCursorY)
	}
}

// TestDrawStringTRNRemapsOpaquePixels pins the colour remap: every drawn pixel
// goes through the slot's TRN (cpp:275-278).
func TestDrawStringTRNRemapsOpaquePixels(t *testing.T) {
	// given a surface and a synthetic identity TRN that maps the d0 palette
	// (indices 195..207) down to 1
	font := loadTestFont(t)
	surface := render.NewSurface(32, 24)
	trn := identityTRN()
	for i := 195; i <= 207; i++ {
		trn[i] = 1
	}

	// when the glyph draws through non-null colour slot 8 (fonts\black.trn)
	render.DrawString(surface, font, "\u00d0",
		render.Rect{X: 0, Y: 0, W: 8, H: 20}, render.Options{Color: 8, TRN: trn})

	// then every drawn pixel carries the remapped index 1
	assertOnlyIndices(t, surface, 1, 52)
}

// TestDrawStringNullColourSlotIgnoresTRN pins the null colour slots (upstream
// 4, 5, 7, text_render.cpp:87-111): they draw with the base palette even when
// a table is supplied.
func TestDrawStringNullColourSlotIgnoresTRN(t *testing.T) {
	// given two surfaces and the synthetic TRN mapping the d0 palette to 1
	font := loadTestFont(t)
	trn := identityTRN()
	for i := 195; i <= 207; i++ {
		trn[i] = 1
	}
	plain := render.NewSurface(32, 24)
	mapped := render.NewSurface(32, 24)

	// when slot 4 (null upstream) draws with and without the TRN
	render.DrawString(plain, font, "\u00d0", render.Rect{X: 0, Y: 0, W: 8, H: 20}, render.Options{Color: 4})
	render.DrawString(mapped, font, "\u00d0", render.Rect{X: 0, Y: 0, W: 8, H: 20}, render.Options{Color: 4, TRN: trn})

	// then both surfaces keep the raw sheet indices and are identical
	for i := range plain.Pixels {
		if plain.Pixels[i] != mapped.Pixels[i] {
			t.Fatalf("pixel %d differs: %d vs %d", i, plain.Pixels[i], mapped.Pixels[i])
		}
		if index := plain.Pixels[i]; index > 1 && index < 195 {
			t.Fatalf("pixel %d = %d, want a sheet index or 0", i, index)
		}
	}
}

// TestDrawStringZeroWidthGlyphAdvancesCursor pins the draw-loop advance: a
// zero-width glyph draws nothing but still advances the cursor by
// width+spacing (cpp:46).
func TestDrawStringZeroWidthGlyphAdvancesCursor(t *testing.T) {
	// given a blank surface and the string d0, blank 0x00 glyph, de
	font := loadTestFont(t)
	surface := render.NewSurface(48, 24)

	// when the string draws left-aligned
	render.DrawString(surface, font, "\u00d0\u0000\u00de",
		render.Rect{X: 0, Y: 0, W: 48, H: 20}, render.Options{})

	// then the blank glyph draws nothing but advances the cursor: the de
	// pixels start at 8 + 1 + 0 + 1 = 10 and the sheet pixel count is 52+51
	if got := countOpaque(surface); got != 103 {
		t.Fatalf("opaque pixels = %d, want 103", got)
	}
	assertOpaqueWithin(t, surface, render.Rect{X: 0, Y: 0, W: 20, H: 24})
	// d0's first opaque pixel (0,5) and de's (0,6) sit at their cursor slots
	if surface.At(0, 5) == 0 || surface.At(10, 6) == 0 {
		t.Fatalf("glyph pixels missing at d0(0,5)=%d de(10,6)=%d", surface.At(0, 5), surface.At(10, 6))
	}
}

// TestDrawStringZWSPDrawsNothing pins that U+200B draws nothing (cpp:46).
func TestDrawStringZWSPDrawsNothing(t *testing.T) {
	// given a blank surface and a ZWSP-only string
	font := loadTestFont(t)
	surface := render.NewSurface(8, 8)

	// when it draws
	render.DrawString(surface, font, "\u200b", render.Rect{X: 0, Y: 0, W: 8, H: 8}, render.Options{})

	// then no pixel is touched
	if got := countOpaque(surface); got != 0 {
		t.Fatalf("opaque pixels = %d, want 0", got)
	}
}

// TestDrawStringAlignment pins GetLineStartX (text_render.cpp:690-692) on a
// 64-wide rect with the 8-wide d0 glyph.
func TestDrawStringAlignment(t *testing.T) {
	// the decoded sheet supplies the 8-wide d0 glyph
	font := loadTestFont(t)

	tests := []struct {
		name  string
		align render.Align
		wantX int
	}{
		{"left", render.AlignLeft, 0},
		{"center", render.AlignCenter, 28}, // (64-8)/2
		{"right", render.AlignRight, 56},   // 64-8
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// given a blank surface
			surface := render.NewSurface(64, 20)

			// when the glyph draws aligned in the rect
			render.DrawString(surface, font, "\u00d0",
				render.Rect{X: 0, Y: 0, W: 64, H: 20}, render.Options{Align: test.align})

			// then the first opaque pixel sits at the aligned line start
			if got := firstOpaqueX(surface); got != test.wantX {
				t.Fatalf("first opaque x = %d, want %d", got, test.wantX)
			}
		})
	}
}

// TestDrawStringOutlinePlusRing pins DrawFont's outlined branch
// (text_render.cpp:269-280) and RenderClxOutline (clx_render.cpp:406-425):
// colour index 0 underlays the 4-neighbours of every drawn pixel.
func TestDrawStringOutlinePlusRing(t *testing.T) {
	// given the synthetic outline sheet and a background-filled surface
	font := outlineSheet(t)
	const background = 0xEE
	surface := render.NewSurface(16, 16)
	fillSurface(surface, background)

	// when the 3x3 plus glyph draws outlined at (4, 4)
	render.DrawString(surface, font, "\u00d0",
		render.Rect{X: 4, Y: 4, W: 12, H: 8}, render.Options{Outline: true})

	// then the glyph carries index 7 and the 8-position outline ring index 0
	for _, p := range [][2]int{{5, 4}, {4, 5}, {5, 5}, {6, 5}, {5, 6}} {
		if got := surface.At(p[0], p[1]); got != 7 {
			t.Fatalf("glyph pixel (%d, %d) = %d, want 7", p[0], p[1], got)
		}
	}
	for _, p := range [][2]int{{4, 4}, {6, 4}, {5, 3}, {3, 5}, {4, 6}, {7, 5}, {6, 6}, {5, 7}} {
		if got := surface.At(p[0], p[1]); got != 0 {
			t.Fatalf("outline pixel (%d, %d) = %d, want 0", p[0], p[1], got)
		}
	}
	if surface.At(2, 2) != background {
		t.Fatalf("untouched pixel = %d, want %d", surface.At(2, 2), background)
	}
}

// TestDrawStringOutlineTransparentPixelSkipped pins the transparent-skip rule
// (GetOutline<SkipColorIndexZero>, clx_render.cpp:406-425): a transparent
// glyph pixel generates no outline.
func TestDrawStringOutlineTransparentPixelSkipped(t *testing.T) {
	// given the synthetic outline sheet and a background-filled surface
	font := outlineSheet(t)
	const background = 0xEE
	surface := render.NewSurface(16, 16)
	fillSurface(surface, background)

	// when the d1 glyph draws outlined: opaque index 9 at (0, 0), transparent
	// pixel at (1, 0)
	render.DrawString(surface, font, "\u00d1",
		render.Rect{X: 4, Y: 4, W: 12, H: 8}, render.Options{Outline: true})

	// then the opaque pixel and its 4-neighbours are drawn/outlined ...
	if got := surface.At(4, 4); got != 9 {
		t.Fatalf("glyph pixel = %d, want 9", got)
	}
	for _, p := range [][2]int{{3, 4}, {5, 4}, {4, 3}, {4, 5}} {
		if got := surface.At(p[0], p[1]); got != 0 {
			t.Fatalf("outline pixel (%d, %d) = %d, want 0", p[0], p[1], got)
		}
	}
	// ... while the transparent pixel's would-be outline stays background
	for _, p := range [][2]int{{6, 4}, {5, 3}, {5, 5}} {
		if got := surface.At(p[0], p[1]); got != background {
			t.Fatalf("skip-rule pixel (%d, %d) = %d, want %d", p[0], p[1], got, background)
		}
	}
}

// loadTestFont decodes the committed 12-05 sheet.
func loadTestFont(t *testing.T) *assets.Font {
	t.Helper()
	data := readFontHex(t, filepath.Join(fontFixtures, fontSheet+".clx.hex"))
	font, err := assets.DecodeFontSheet(data)
	if err != nil {
		t.Fatalf("DecodeFontSheet: %v", err)
	}
	return font
}

// readFontHex reads a committed hex dump file into bytes, failing when the
// fixture is missing: a missing fixture is a corpus defect, never a skip.
func readFontHex(t *testing.T, path string) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}
	replacer := strings.NewReplacer("\n", "", " ", "")
	data, err := hex.DecodeString(replacer.Replace(string(raw)))
	if err != nil {
		t.Fatalf("decode hex %s: %v", path, err)
	}
	return data
}

// assertLayout asserts one wrap row's line count, last-line width and cursor.
func assertLayout(t *testing.T, got render.Result, wantCount, wantWidth, wantX, wantY int) {
	t.Helper()
	if got.LineCount != wantCount {
		t.Fatalf("LineCount = %d, want %d", got.LineCount, wantCount)
	}
	if got.LastLineWidth != wantWidth {
		t.Fatalf("LastLineWidth = %d, want %d", got.LastLineWidth, wantWidth)
	}
	if got.LastCursorX != wantX || got.LastCursorY != wantY {
		t.Fatalf("last cursor = (%d, %d), want (%d, %d)", got.LastCursorX, got.LastCursorY, wantX, wantY)
	}
}

// assertOnlyIndices asserts every opaque pixel carries want and there are
// exactly count of them.
func assertOnlyIndices(t *testing.T, surface render.Surface, want uint8, count int) {
	t.Helper()
	got := 0
	for i, index := range surface.Pixels {
		if index == 0 {
			continue
		}
		if index != want {
			t.Fatalf("pixel (%d, %d) = %d, want %d", i%surface.Width, i/surface.Width, index, want)
		}
		got++
	}
	if got != count {
		t.Fatalf("opaque pixels = %d, want %d", got, count)
	}
}

// assertOpaqueWithin asserts every opaque pixel sits inside the given bounds.
func assertOpaqueWithin(t *testing.T, surface render.Surface, bounds render.Rect) {
	t.Helper()
	for i, index := range surface.Pixels {
		if index == 0 {
			continue
		}
		x, y := i%surface.Width, i/surface.Width
		if x < bounds.X || x >= bounds.X+bounds.W || y < bounds.Y || y >= bounds.Y+bounds.H {
			t.Fatalf("opaque pixel (%d, %d) outside %v", x, y, bounds)
		}
	}
}

// identityTRN returns an identity translation table.
func identityTRN() assets.TRN {
	var trn assets.TRN
	for i := range trn {
		trn[i] = uint8(i)
	}
	return trn
}

// outlineSheet builds a synthetic 256-frame font sheet through the
// golden-tested CLX encoder: frame 0xd0 is a 3x3 plus of palette index 7,
// frame 0xd1 is an opaque index-9 pixel with a transparent pixel to its right,
// every other frame is a fully transparent 4x4 block. EncodeCLXSpriteList
// stores every frame's rows bottom-up (SurfaceToClx), so pixels are written
// y-flipped to decode in the intended orientation.
func outlineSheet(t *testing.T) *assets.Font {
	t.Helper()

	const frameWidth, frameHeight = 4, 4
	pixels := make([]byte, frameWidth*frameHeight*assets.FontGlyphCount)
	setPixel := func(frame, x, y int, index byte) {
		pixels[frame*frameWidth*frameHeight+(frameHeight-1-y)*frameWidth+x] = index
	}
	for _, p := range [][2]int{{1, 0}, {0, 1}, {1, 1}, {2, 1}, {1, 2}} {
		setPixel(0xD0, p[0], p[1], 7)
	}
	setPixel(0xD1, 0, 0, 9)

	data, err := assets.EncodeCLXSpriteList(pixels, frameWidth, assets.FontGlyphCount, 0)
	if err != nil {
		t.Fatalf("EncodeCLXSpriteList: %v", err)
	}
	font, err := assets.DecodeFontSheet(data)
	if err != nil {
		t.Fatalf("DecodeFontSheet: %v", err)
	}
	return font
}

// fillSurface writes one palette index over every pixel.
func fillSurface(surface render.Surface, index uint8) {
	for i := range surface.Pixels {
		surface.Pixels[i] = index
	}
}

// countOpaque counts non-zero palette indices on the surface.
func countOpaque(surface render.Surface) int {
	count := 0
	for _, index := range surface.Pixels {
		if index != 0 {
			count++
		}
	}
	return count
}

// firstOpaqueX returns the x of the first non-zero pixel, -1 when none.
func firstOpaqueX(surface render.Surface) int {
	for y := range surface.Height {
		for x := range surface.Width {
			if surface.At(x, y) != 0 {
				return x
			}
		}
	}
	return -1
}
