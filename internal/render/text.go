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
const gameFontLineHeight = 12

// gameFontBaseLineOffset is the vertical offset from rect.Y to the first
// glyph baseline (text_render.cpp:385-397, GetBaseLineOffset for
// GameFont12 returns -3). Every frame on the 12-05 sheet is 20px tall;
// the ascent is 17px from the top of the frame, so baseline sits 3px
// above the rect top.
const gameFontBaseLineOffset = -3

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
	// KerningFit is upstream UiFlags::KerningFitSpacing (text_render.cpp:524):
	// per line the spacing shrinks by AdjustSpacingToFitHorizontally
	// (cpp:415-424) so the line fits rect.W exactly.
	KerningFit bool
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

// ColorRun is one per-argument run of DrawStringWithColors (DrawStringFormatArg,
// text_render.cpp:220-265): the run's text draws with its own colour TRN.
type ColorRun struct {
	Text  string
	Color int
	TRN   assets.TRN
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

// highlightState carries the invariant draw context of one
// DrawStringWithHighlight call across its wrapped lines; charIndex advances by
// one per line (the dropped line break, see the caller).
type highlightState struct {
	surface        Surface
	font           *assets.Font
	rect           Rect
	options        Options
	spacing        int
	charIndex      int
	highlightStart int
	highlightEnd   int
}

// lineState carries the per-line mutable draw state drawOneLine updates: the
// (possibly shrunk) inter-glyph spacing, the measured line width, and the
// cumulative Result the caller aggregates across lines.
type lineState struct {
	spacing int
	total   int
	result  *Result
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

		state := lineState{spacing: spacing, total: total, result: &result}
		drawOneLine(surface, font, line, rect, lineOffset, options, &state)
		lineOffset += gameFontLineHeight
	}

	return result
}

// DrawStringWithColors draws a format string's per-argument runs, each with
// its own colour TRN (DrawStringFormat / DrawStringFormatArg,
// text_render.cpp:220-265, 873-903): the runs join into one wrapped paragraph
// and every glyph of a run draws through that run's table.
func DrawStringWithColors(
	surface Surface,
	font *assets.Font,
	runs []ColorRun,
	rect Rect,
	options Options,
) Result {
	var joined string
	for _, run := range runs {
		joined += run.Text
	}

	spacing := kerningFitSpacing(font, joined, rect, options)
	wrapWidth := rect.W

	var result Result
	lineOffset := 0
	runeOffset := 0
	for _, line := range wrapLines(font, joined, wrapWidth, spacing) {
		total := font.LineWidth(line, spacing)
		if total == 0 {
			lineOffset += gameFontLineHeight
			continue
		}

		x := lineStartX(options.Align, rect, total)
		for _, code := range line {
			run := runForRune(runs, runeOffset)
			runOpts := options
			runOpts.Color = run.Color
			runOpts.TRN = run.TRN
			drawGlyph(
				surface,
				font,
				code,
				x,
				rect.Y+lineOffset+gameFontBaseLineOffset,
				rect,
				runOpts,
			)
			x += font.Width(code) + spacing
			runeOffset++
		}
		result.LineCount++
		result.LastLineWidth = total
		result.LastCursorX = x
		result.LastCursorY = lineOffset
		lineOffset += gameFontLineHeight
		runeOffset++
	}

	return result
}

// DrawStringWithCursor draws text with a cursor glyph at the given character
// position (text_render.cpp: DrawStringWithCursor overlay). The cursor is the
// base font's '|' glyph (row 0, code 124) drawn with the same colour TRN.
func DrawStringWithCursor(
	surface Surface,
	font *assets.Font,
	text string,
	rect Rect,
	options Options,
	cursorPos int,
) Result {
	result := DrawString(surface, font, text, rect, options)

	spacing := options.Spacing
	if spacing == 0 {
		spacing = 1
	}

	lineOffset := 0
	for _, line := range wrapLines(font, text, rect.W, spacing) {
		drawn := drawCursorLine(surface, font, line, rect, lineOffset, options, spacing, &cursorPos)
		if drawn {
			break
		}
		lineOffset += gameFontLineHeight
	}

	return result
}

// DrawStringWithHighlight draws text with a highlight range [start, end) drawn
// in PAL8_BLUE before the glyphs (text_render.cpp highlight fill).
func DrawStringWithHighlight(
	surface Surface,
	font *assets.Font,
	text string,
	rect Rect,
	options Options,
	highlightStart, highlightEnd int,
) Result {
	spacing := options.Spacing
	if spacing == 0 {
		spacing = 1
	}

	var result Result
	st := highlightState{
		surface:        surface,
		font:           font,
		rect:           rect,
		options:        options,
		spacing:        spacing,
		highlightStart: highlightStart,
		highlightEnd:   highlightEnd,
	}
	lineOffset := 0
	for _, line := range wrapLines(font, text, rect.W, spacing) {
		total := font.LineWidth(line, spacing)
		if total == 0 {
			lineOffset += gameFontLineHeight
			continue
		}
		drawHighlightedLine(&st, line, lineOffset, total, &result)
		lineOffset += gameFontLineHeight
		// The highlight range is a byte index over the whole string; each line
		// break consumes one extra byte ('\n') the wrapped lines drop.
		st.charIndex += len([]rune(line)) + 1
	}

	return result
}

// drawOneLine draws one wrapped line: KerningFit shrinks the spacing to fit
// (AdjustSpacingToFitHorizontally, text_render.cpp:415-424) before the glyph
// walk, which advances by width+spacing (cpp:508-510).
func drawOneLine(
	surface Surface,
	font *assets.Font,
	line string,
	rect Rect,
	lineOffset int,
	options Options,
	state *lineState,
) {
	if options.KerningFit {
		runes := []rune(line)
		state.spacing = adjustSpacingToFitHorizontally(state.total, state.spacing, len(runes), rect.W)
		state.total = font.LineWidth(line, state.spacing)
	}

	x := lineStartX(options.Align, rect, state.total)
	for _, code := range line {
		drawGlyph(
			surface,
			font,
			code,
			x,
			rect.Y+lineOffset+gameFontBaseLineOffset,
			rect,
			options,
		)
		x += font.Width(code) + state.spacing
	}
	state.result.LineCount++
	state.result.LastLineWidth = state.total
	state.result.LastCursorX = x
	state.result.LastCursorY = lineOffset
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
	minX := max(x, rect.X)
	maxX := min(x+glyph.Width, rect.X+rect.W, surface.Width)
	minY := max(y, 0)
	maxY := min(y+glyph.Height, rect.Y+rect.H, surface.Height)
	if minX >= maxX || minY >= maxY {
		return
	}
	for py := minY - y; py < maxY-y; py++ {
		for px := minX - x; px < maxX-x; px++ {
			setGlyphPixel(surface, glyph, x, y, px, py, remap)
		}
	}
}

// setGlyphPixel writes one glyph pixel through the remap when the sheet index
// is opaque (non-zero); transparent pixels draw nothing.
func setGlyphPixel(surface Surface, glyph assets.Glyph, x, y, px, py int, remap func(uint8) uint8) {
	index := glyph.Pixels[py*glyph.Width+px]
	if index == 0 {
		return
	}
	surface.Set(x+px, y+py, remap(index))
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

// adjustSpacingToFitHorizontally (text_render.cpp:415-424): shrink the spacing
// so lineWidth + (chars-1)*spacing fits availableWidth; a single character or
// a fitting line keeps maxSpacing.
func adjustSpacingToFitHorizontally(lineWidth, maxSpacing, charactersInLine, availableWidth int) int {
	if lineWidth <= availableWidth || charactersInLine < 2 {
		return maxSpacing
	}
	overhang := lineWidth - availableWidth
	spacingRedux := (overhang + charactersInLine - 2) / (charactersInLine - 1)
	return maxSpacing - spacingRedux
}

// kerningFitSpacing applies upstream AdjustSpacingToFitHorizontally before the
// draw loop when KerningFit is set (text_render.cpp:415-424); plain spacing
// defaults to 1 otherwise (text_render.cpp:1000).
func kerningFitSpacing(font *assets.Font, text string, rect Rect, options Options) int {
	spacing := options.Spacing
	if spacing == 0 {
		spacing = 1
	}
	if options.KerningFit {
		spacing = adjustSpacingToFitHorizontally(
			font.LineWidth(text, spacing),
			spacing,
			len([]rune(text)),
			rect.W,
		)
	}
	return spacing
}

// runForRune resolves the colour run holding the rune at joined-string index
// i (the format-string placeholder order is the run order, cpp:222-265).
func runForRune(runs []ColorRun, i int) ColorRun {
	for _, run := range runs {
		runes := len([]rune(run.Text))
		if i < runes {
			return run
		}
		i -= runes
	}
	if len(runs) == 0 {
		return ColorRun{}
	}
	return runs[len(runs)-1]
}

// drawCursorLine draws the cursor glyph when it falls on this line and
// advances the whole-string cursor position past the line (the line break
// consumes one extra byte the wrapped lines drop). Returns true once drawn.
func drawCursorLine(
	surface Surface,
	font *assets.Font,
	line string,
	rect Rect,
	lineOffset int,
	options Options,
	spacing int,
	cursorPos *int,
) bool {
	total := font.LineWidth(line, spacing)
	if total == 0 {
		return false
	}

	x := lineStartX(options.Align, rect, total)
	lineRunes := []rune(line)
	drawn := false
	for i, code := range lineRunes {
		if i == *cursorPos {
			drawCursorGlyph(surface, font, x, rect.Y+lineOffset+gameFontBaseLineOffset, rect, options)
			drawn = true
		}
		x += font.Width(code) + spacing
	}
	if *cursorPos == len(lineRunes) && !drawn {
		drawCursorGlyph(surface, font, x, rect.Y+lineOffset+gameFontBaseLineOffset, rect, options)
		drawn = true
	}
	*cursorPos = advanceCursorPos(*cursorPos, len(lineRunes))
	return drawn
}

// drawCursorGlyph draws the '|' cursor glyph at the given pixel position.
func drawCursorGlyph(surface Surface, font *assets.Font, x, y int, rect Rect, options Options) {
	drawGlyph(surface, font, '|', x, y, rect, options)
}

// advanceCursorPos moves the whole-string cursor position past a wrapped
// line's runes plus the dropped line break.
func advanceCursorPos(cursorPos, lineRunes int) int {
	cursorPos -= lineRunes + 1
	if cursorPos < 0 {
		cursorPos = 0
	}
	return cursorPos
}

// drawHighlightedLine pre-draws this line's highlight rectangles, then the
// glyphs themselves through drawOneLine.
func drawHighlightedLine(st *highlightState, line string, lineOffset, total int, result *Result) {
	x := lineStartX(st.options.Align, st.rect, total)
	lineRunes := []rune(line)
	for i, code := range lineRunes {
		fillHighlightRect(st, code, x, lineOffset, st.charIndex+i)
		x += st.font.Width(code) + st.spacing
	}
	state := lineState{spacing: st.spacing, total: total, result: result}
	drawOneLine(st.surface, st.font, line, st.rect, lineOffset, st.options, &state)
}

// fillHighlightRect paints the PAL8_BLUE highlight under one in-range glyph.
func fillHighlightRect(st *highlightState, code rune, x, lineOffset, globalIdx int) {
	if globalIdx < st.highlightStart || globalIdx >= st.highlightEnd {
		return
	}
	glyph := st.font.Glyph(code)
	if glyph.Width == 0 {
		return
	}
	clipFill(st.surface, st.rect, highlightRect(st, glyph, x, lineOffset, globalIdx))
}

// highlightRect measures one highlight rect; every in-range char but the
// range's last widens by the spacing (text_render.cpp:501-503).
func highlightRect(st *highlightState, glyph assets.Glyph, x, lineOffset, globalIdx int) Rect {
	h := Rect{
		X: x,
		Y: st.rect.Y + lineOffset + gameFontBaseLineOffset,
		W: glyph.Width,
		H: glyph.Height,
	}
	if globalIdx+1 != st.highlightEnd {
		h.W += st.spacing
	}
	return h
}

// clipFill clips r to the rect clip region and surface bounds
// (ClipSurface, text_render.cpp:400-408) and fills it with PAL8_BLUE (128).
func clipFill(surface Surface, clip, r Rect) {
	maxX := min(clip.X+clip.W, surface.Width)
	maxY := min(clip.Y+clip.H, surface.Height)
	if r.X < clip.X {
		r.W -= clip.X - r.X
		r.X = clip.X
	}
	if r.Y < 0 {
		r.H -= -r.Y
		r.Y = 0
	}
	if r.X+r.W > maxX {
		r.W = maxX - r.X
	}
	if r.Y+r.H > maxY {
		r.H = maxY - r.Y
	}
	for py := range r.H {
		for px := range r.W {
			surface.Set(r.X+px, r.Y+py, 128)
		}
	}
}
