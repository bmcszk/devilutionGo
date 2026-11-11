package render

import (
	"github.com/bmcszk/devilutionGo/internal/assets"
)

// zwsp is the zero-width space U+200B: a breakable whitespace rune with zero
// advance (text_render.cpp:46, IsBreakableWhitespace cpp:1012).
const zwsp = '\u200b'

// gameFontLineHeight is the glyph row height every line break advances by:
// every 12-05 frame is 20 pixels tall (internal/testdata/assets/fonts
// 12-05.sha256 row heights).
// TODO: upstream GetLineHeight is per-font (LineHeights[],
// text_render.cpp:385-397); make this per-font when the second sheet lands —
// no speculative per-font API before that.
const gameFontLineHeight = 20

// Alignment modes (TextRenderFlags, text_render.cpp:690-692).
const (
	AlignLeft Align = iota
	AlignCenter
	AlignRight
)

// Align is the horizontal line alignment (GetLineStartX, text_render.cpp:690-692).
type Align int

// Rect is the layout rectangle the string draws into (text_render.cpp:783).
type Rect struct {
	X, Y, W, H int
}

// Options are the DrawString layout knobs (spacing 1 default, text_render.cpp:1000).
type Options struct {
	Spacing int
	Align   Align
	Outline bool
	// Color is the font TRN slot; a null slot (upstream 4, 5, 7,
	// text_render.cpp:87-111) draws with the base palette.
	Color int
	// NoTable says no translation table is supplied: the draw keeps the sheet
	// indices even in a non-null colour slot. Explicit because the zero TRN is
	// itself a legitimate table (every index remaps to 0).
	NoTable bool
	// TRN is the translation table used when Color names a non-null slot and
	// NoTable is unset.
	TRN assets.TRN
}

// Result reports what DrawString drew: the drawn line count, the last drawn
// line's measured width (GetLineWidth semantics), and the cursor after the
// final glyph of the last drawn line.
type Result struct {
	LineCount     int
	LastLineWidth int
	LastCursorX   int
	LastCursorY   int
}

// DrawString wraps text into rect and draws every line, reporting the layout
// result. Semantics are the port of text_render.cpp: WordWrap (cpp:783-820),
// DrawString (cpp:268-300), GetLineStartX (cpp:690-692), DrawFont
// (cpp:269-280). A word alone longer than the line keeps drawing from the line
// start — the upstream hard-split at the overflowing character (cpp:803-807,
// upstream 99bf0e4e3 changed keep-whole to hard-break).
func DrawString(surface Surface, font *assets.Font, text string, rect Rect, options Options) Result {
	spacing := options.Spacing
	if spacing == 0 {
		spacing = 1
	}

	var result Result
	lineOffset := 0
	for _, line := range wrapLines(font, text, rect.W, spacing) {
		total := font.LineWidth(line, spacing)
		if total == 0 {
			lineOffset += gameFontLineHeight
			continue
		}

		x := lineStartX(options.Align, rect, total)
		for _, code := range line {
			drawGlyph(surface, font, code, x, rect.Y+lineOffset, rect, options)
			x += font.Width(code) + spacing
		}
		result.LineCount++
		result.LastLineWidth = total
		result.LastCursorX = x
		result.LastCursorY = lineOffset
		lineOffset += gameFontLineHeight
	}

	return result
}

// wrapLines splits text into drawn sub-lines (WordWrap, text_render.cpp:783-820):
// '\n' breaks (cpp:788), the breakable whitespace {' ', U+200B} contributes no
// glyph (cpp:1012), and a glyph that no longer fits starts a new line unless it
// already sits at the line start — that residue is the upstream hard-split for
// a word alone longer than the width (cpp:803-807).
func wrapLines(font *assets.Font, text string, width, spacing int) []string {
	var (
		lines     []string
		line      []rune
		lineWidth int
	)
	flush := func() {
		lines = append(lines, string(line))
		line = nil
		lineWidth = 0
	}

	for _, code := range text {
		switch code {
		case '\n':
			flush()
		case ' ', zwsp:
			// The breakable space never triggers the break itself: upstream
			// checks IsBreakableWhitespace before the fit check
			// (text_render.cpp:787-790 before cpp:794), sheet-independently.
			line = append(line, code)
			lineWidth += font.Width(code) + spacing
		default:
			glyphWidth := font.Width(code)
			if lineWidth > 0 && lineWidth+glyphWidth > width {
				flush()
			}
			line = append(line, code)
			lineWidth += glyphWidth + spacing
		}
	}
	return append(lines, string(line))
}

// drawGlyph draws one glyph through the outline underlay and the colour TRN
// (DrawFont, text_render.cpp:269-280): the outline is an index-0 underlay of
// the 4-neighbours of every drawn pixel, skipping pixels that are themselves
// transparent (clx_render.cpp:406-425, GetOutline<SkipColorIndexZero>). The
// draw clips to the rect (ClipSurface, cpp:400-408, applied at cpp:855).
func drawGlyph(surface Surface, font *assets.Font, code rune, x, y int, rect Rect, options Options) {
	glyph := font.Glyph(code)
	if glyph.Width == 0 {
		return
	}

	if options.Outline {
		drawOutline(surface, glyph, x, y, rect)
	}

	if trn, ok := translationFor(options); ok {
		drawPixels(surface, glyph, x, y, rect, func(sheet uint8) uint8 { return trn[sheet] })
		return
	}
	drawPixels(surface, glyph, x, y, rect, func(sheet uint8) uint8 { return sheet })
}

// translationFor resolves the colour slot's TRN; upstream null slots draw with
// the base palette (text_render.cpp:87-111, 269-280), and an explicitly
// unsupplied table (NoTable, review F2) draws plain.
func translationFor(options Options) (assets.TRN, bool) {
	if options.NoTable {
		return assets.TRN{}, false
	}
	if _, ok := assets.ColorTranslationPath(options.Color); !ok {
		return assets.TRN{}, false
	}
	return options.TRN, true
}

// drawOutline underlays index 0 on the 4-neighbours of every opaque glyph
// pixel; a transparent pixel generates no outline (clx_render.cpp:406-425).
func drawOutline(surface Surface, glyph assets.Glyph, x, y int, rect Rect) {
	for py := range glyph.Height {
		for px := range glyph.Width {
			if glyph.Pixels[py*glyph.Width+px] == 0 {
				continue
			}
			outlineAround(surface, x+px, y+py, rect)
		}
	}
}

// outlineAround underlays the 4-neighbours of one opaque pixel
// (clx_render.cpp:418-421).
//
// Unlike upstream, neighbours are written per opaque pixel with no dedup:
// upstream collects the outline pixels into a per-sprite set first
// (clx_render.cpp:408-417) so each neighbour is written once. The final
// indices are identical; the overdraw is perf-only, revisit on evidence — no
// perf rewrite this wave.
func outlineAround(surface Surface, x, y int, rect Rect) {
	for _, n := range [4][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
		surface.setOutline(x+n[0], y+n[1], rect)
	}
}

// setOutline writes the outline colour into a surface pixel, skipping the draw
// when the neighbour falls off the surface (clx_render.cpp:418-421) or outside
// the rect clip region (ClipSurface, text_render.cpp:400-408).
func (s Surface) setOutline(x, y int, rect Rect) {
	if x < 0 || y < 0 || x >= s.Width || y >= s.Height {
		return
	}
	if x >= rect.X+rect.W || y >= rect.Y+rect.H {
		return
	}
	s.Set(x, y, 0)
}

// drawPixels writes a glyph's opaque pixels, remapped, into the surface
// (RenderClxSprite through ClxDraw, clx_render.cpp:589), clipped to the rect
// like ClipSurface (text_render.cpp:400-408, applied at cpp:855) so an
// overlong word or a line past rect.H never paints outside the rect.
func drawPixels(surface Surface, glyph assets.Glyph, x, y int, rect Rect, remap func(uint8) uint8) {
	maxX := min(rect.X+rect.W, surface.Width)
	maxY := min(rect.Y+rect.H, surface.Height)
	for py := max(0, -y); py < min(glyph.Height, maxY-y); py++ {
		for px := max(0, -x); px < min(glyph.Width, maxX-x); px++ {
			if index := glyph.Pixels[py*glyph.Width+px]; index != 0 {
				surface.Set(x+px, y+py, remap(index))
			}
		}
	}
}

// lineStartX is GetLineStartX (text_render.cpp:690-692): left-aligned text
// starts at rect.X, centred text at rect.X+(rect.W-total)/2, right-aligned at
// rect.X+rect.W-total, all in C++ int math.
func lineStartX(align Align, rect Rect, total int) int {
	switch align {
	case AlignCenter:
		return rect.X + (rect.W-total)/2
	case AlignRight:
		return rect.X + rect.W - total
	default:
		return rect.X
	}
}
