package render

// This file ports the low-level dungeon tile rasteriser,
// Source/engine/render/dun_render.cpp (pinned ca91d778f): the mask prefix
// arithmetic, the per-type frame renderers and the black tile helper. Every
// helper keeps the upstream name so a reader can diff against the C++.
//
// WIP scope: the unclipped paths plus the surface-bounds guard (CalculateClip
// short-circuit). The upstream partially-visible variants (vertical and
// left/right clip) are deferred until a viewport draw can reach them; the
// light dispatch ports the non-per-pixel path (FullyLit / PartiallyLit /
// FullyDark); the PerPixel option branch is the later lighting bean's lane.

const (
	// LowerHeight is the lower diamond half's row count (dun_render.hpp:43).
	LowerHeight = DunFrameHeight / 2
	// TriangleUpperHeight is the upper diamond rows (dun_render.hpp:46).
	TriangleUpperHeight = (DunFrameHeight / 2) - 1
	// TrapezoidUpperHeight is the trapezoid rect rows (dun_render.hpp:49).
	TrapezoidUpperHeight = DunFrameHeight / 2
	// TriangleHeight is DunFrameTriangleHeight, the 31 diamond rows.
	TriangleHeight = DunFrameHeight - 1
	// XStep is the per-row diamond x advance (dun_render.hpp:54).
	XStep = 2
	// triangleUpperByteCount is the upper half's total payload (30+28+..+2),
	// the slice length RenderTriangleUpper walks (dun_render.cpp:475-493).
	triangleUpperByteCount = 240

	// NumShades exposes the palette core's light-table count to tile consumers.
	NumShades = numShades
	// PaletteSize exposes the palette width to tile consumers.
	PaletteSize = paletteSize
)

// MaskSolid copies the frame's opaque pixels only (MaskType::Solid).
const MaskSolid MaskType = iota

// MaskTransparent blends every opaque pixel (MaskType::Transparent).
const MaskTransparent MaskType = iota + 1

// MaskRight blends the upper-right part of a rect-shaped frame (MaskType::Right).
const MaskRight MaskType = iota + 2

// MaskLeft blends the upper-left part of a rect-shaped frame (MaskType::Left).
const MaskLeft MaskType = iota + 3

// MaskType selects how a tile frame blends into the destination
// (dun_render.hpp:28). Left/Right apply only to TransparentSquare and the
// trapezoid of the matching side.
type MaskType uint8

// tileClip is the partial-draw bounds of one frame (dun_render.cpp:249).
type tileClip struct {
	top, bottom int
	left, right int
	width       int
	height      int
}

// tileBlits holds the two index writers one light flavour selects
// (dun_render.cpp:117-160): opaque is the RenderLineOpaque flavour, blended
// the RenderLineTransparent one.
type tileBlits struct {
	opaque  func(dst []uint8, src []byte)
	blended func(dst []uint8, src []byte)
}

// TileLighting carries the light context of one tile draw: the 16 light
// tables, the tile's light level (0 = fully lit, 15 = fully dark) and the
// blend LUT for the transparent flavours.
type TileLighting struct {
	// Tables holds the 16 light tables the L1-L8 lighting chain produced.
	Tables *[numShades][paletteSize]uint8
	// Level is the tile's light level into Tables.
	Level uint8
	// Blend drives the transparent flavours' 50% blend lookups.
	Blend BlendLUT
}

// RenderTileFrame blits one re-encoded tile frame at pos; pos's row is the
// frame's bottom row (dun_render.cpp:1040-1076), light owns the flavour
// dispatch.
func RenderTileFrame(
	out Surface,
	pos PixelPos,
	tile TileType,
	src []byte,
	height int,
	mask MaskType,
	light TileLighting,
) {
	clip := calculateTileClip(pos.X, pos.Y, DunFrameWidth, height, out)
	if clip.width <= 0 || clip.height <= 0 {
		return
	}
	if clip.left != 0 || clip.right != 0 || clip.top != 0 || clip.bottom != 0 {
		// partially-visible draws are the deferred clip seam
		return
	}

	b := newTileBlits(light.Tables, light.Level, light.Blend)

	switch mask {
	case MaskSolid, MaskTransparent:
		renderUniformMask(b, out, pos, tile, src, height, mask)
	case MaskLeft, MaskRight:
		renderSideMask(b, out, pos, tile, src, height, mask)
	default:
	}
}

// InitPrefix is the row 0 prefix (dun_render.hpp:99-101): the transparent
// prefix starts fully beyond the row's far side and grows towards the blend
// boundary.
func InitPrefix(mask MaskType) int {
	if mask == MaskLeft {
		return -DunFrameWidth
	}

	return DunFrameWidth * 2
}

// PrefixIncrement grows the prefix per line, bottom-up (dun_render.hpp:94-99).
func PrefixIncrement(mask MaskType) int {
	switch mask {
	case MaskLeft:
		return 2
	case MaskRight:
		return -2
	default:
		return 0
	}
}

// WorldDrawBlackTile fills the black diamond, the two fully-dark triangle
// halves at the tile's floor positions (dun_render.cpp:1080-1116).
func WorldDrawBlackTile(out Surface, sx, sy int) {
	clipLeft := calculateTileClip(sx, sy, DunFrameWidth, TriangleHeight, out)
	if clipLeft.height <= 0 {
		return
	}

	// the fill blits never read their source; upstream hands over a nullptr
	fill := func(dst []uint8, _ []byte) { clear(dst) }

	renderLeftTriangleFull(out, PixelPos{X: sx, Y: sy}, []byte{}, fill)
	renderRightTriangleFull(out, PixelPos{X: sx + DunFrameWidth, Y: sy}, []byte{}, fill)
}

// renderUniformMask is the Mask Solid/Transparent arm of the RenderTileFrame
// dispatch (dun_render.cpp:1009-1028; Solid/Transparent reach every tile).
func renderUniformMask(b tileBlits, out Surface, pos PixelPos, tile TileType, src []byte, height int, mask MaskType) {
	writer := b.opaque
	if mask == MaskTransparent {
		writer = b.blended
	}

	switch tile {
	case TileSquare:
		renderSquareFull(b, out, pos, src, height, mask)
	case TileTransparentSquare:
		renderTransparentSquareFull(b, out, pos, src, height, mask)
	case TileLeftTriangle:
		renderLeftTriangleFull(out, pos, src, writer)
	case TileRightTriangle:
		renderRightTriangleFull(out, pos, src, writer)
	case TileLeftTrapezoid:
		renderLeftTrapezoidFull(b, out, pos, src, mask)
	case TileRightTrapezoid:
		renderRightTrapezoidFull(b, out, pos, src, mask)
	default:
	}
}

// renderSideMask is the Mask Left/Right arm: only a transparent square and the
// matching-side trapezoid accept it (dun_render.cpp:1030-1046); everything
// else aborts upstream.
func renderSideMask(b tileBlits, out Surface, pos PixelPos, tile TileType, src []byte, height int, mask MaskType) {
	switch tile {
	case TileTransparentSquare:
		renderTransparentSquareFull(b, out, pos, src, height, mask)
	case TileLeftTrapezoid:
		renderLeftTrapezoidFull(b, out, pos, src, mask)
	case TileRightTrapezoid:
		renderRightTrapezoidFull(b, out, pos, src, mask)
	default:
		// upstream aborts: "Given mask can only be applied to
		// TransparentSquare or [matching] trapezoid tiles"
		panic("given mask can only be applied to TransparentSquare or matching trapezoid tiles")
	}
}

// triangleWriter picks the triangle base's writer for a trapezoid draw: the
// lower 16 rows stay opaque for the side masks (dun_render.cpp:807, 850:
// Transparent == Mask == MaskType::Transparent).
func triangleWriter(b tileBlits, mask MaskType) func(dst []uint8, src []byte) {
	if mask == MaskTransparent {
		return b.blended
	}

	return b.opaque
}

// calculateTileClip is CalculateClip (dun_render.cpp:259-272): the frame's
// bottom row sits on pos's row, its columns at pos's x.
func calculateTileClip(x, y, w, h int, surface Surface) tileClip {
	var clip tileClip
	clip.top = max(h-(y+1), 0)
	clip.bottom = max((y+1)-surface.Height, 0)
	clip.left = max(-x, 0)
	clip.right = max(x+w-surface.Width, 0)
	clip.width = w - clip.left - clip.right
	clip.height = h - clip.top - clip.bottom
	return clip
}

// newTileBlits maps the light dispatch to index blits
// (dun_render.cpp:962-1007, dun_render_benchmark.cpp:64-67): the fully-lit
// and fully-dark tables are the first and last of the 16, and their pointer
// membership is how upstream picks a flavour (light_render.hpp:49-50).
func newTileBlits(tables *[numShades][paletteSize]uint8, level uint8, lut BlendLUT) tileBlits {
	switch level {
	case fullyLitShade:
		return fullyLitBlits(lut)
	case numShades - 1:
		return fullyDarkBlits(lut)
	default:
		return partiallyLitBlits(tables, level, lut)
	}
}

// fullyLitBlits is the LightType::FullyLit pair (dun_render.cpp:122-144):
// opaque pixels copy through, transparent pixels blend with the destination.
func fullyLitBlits(lut BlendLUT) tileBlits {
	return tileBlits{
		opaque: func(dst []uint8, src []byte) {
			for i, index := range src {
				dst[i] = uint8(index)
			}
		},
		blended: func(dst []uint8, src []byte) {
			for i, index := range src {
				dst[i] = lut[dst[i]][index]
			}
		},
	}
}

// fullyDarkBlits is the LightType::FullyDark pair (dun_render.cpp:118-121,
// 146-154): opaque pixels fill black, transparent pixels blend black in.
func fullyDarkBlits(lut BlendLUT) tileBlits {
	return tileBlits{
		opaque: func(dst []uint8, _ []byte) {
			clear(dst)
		},
		blended: func(dst []uint8, _ []byte) {
			for i := range dst {
				dst[i] = lut[dst[i]][0]
			}
		},
	}
}

// partiallyLitBlits is the LightType::PartiallyLit pair (dun_render.cpp:156-160,
// 215-222): both flavours pass through the level's light table first.
func partiallyLitBlits(tables *[numShades][paletteSize]uint8, level uint8, lut BlendLUT) tileBlits {
	table := tables[level]

	return tileBlits{
		opaque: func(dst []uint8, src []byte) {
			for i, index := range src {
				dst[i] = table[index]
			}
		},
		blended: func(dst []uint8, src []byte) {
			for i, index := range src {
				dst[i] = lut[dst[i]][table[index]]
			}
		},
	}
}

// renderLine writes one run of a transparent square, splitting it at the mask
// boundary exactly like RenderLine (dun_render.cpp:237-247): for Left/Right
// masks the boundary crossings split into a blended and an opaque part.
func (b tileBlits) renderLine(dst []uint8, src []byte, mask MaskType, prefix int) {
	n := len(src)
	switch {
	case mask == MaskSolid:
		b.opaque(dst, src)
	case mask == MaskTransparent:
		b.blended(dst, src)
	case prefix >= n:
		if mask == MaskRight {
			b.opaque(dst, src)
		} else {
			b.blended(dst, src)
		}
	case prefix <= 0:
		if mask == MaskLeft {
			b.opaque(dst, src)
		} else {
			b.blended(dst, src)
		}
	case mask == MaskRight:
		b.opaque(dst[:prefix], src[:prefix])
		b.blended(dst[prefix:], src[prefix:])
	default:
		b.blended(dst[:prefix], src[:prefix])
		b.opaque(dst[prefix:], src[prefix:])
	}
}

// renderLineTransparentOrOpaque is the Solid/Transparent pair
// (dun_render.cpp:204-213).
func (b tileBlits) renderLineTransparentOrOpaque(dst []uint8, src []byte, mask MaskType) {
	if mask == MaskSolid {
		b.opaque(dst, src)
	} else {
		b.blended(dst, src)
	}
}

// renderSquareFull is RenderSquareFull (dun_render.cpp:279-293): 32 raw rows,
// bottom row first.
func renderSquareFull(b tileBlits, out Surface, pos PixelPos, src []byte, height int, mask MaskType) {
	dst := pos.Y*out.Width + pos.X

	for i := range height {
		b.renderLineTransparentOrOpaque(out.Pixels[dst:dst+DunFrameWidth], src[i*DunFrameWidth:(i+1)*DunFrameWidth], mask)
		dst -= out.Width
	}
}

// renderTransparentSquareFull is RenderTransparentSquareFull
// (dun_render.cpp:295-310): RLE rows bottom-up, the mask prefix tracked per
// line and re-decoded relative to each run's start column; the row rewind
// carries +Width like the upstream pointer walk.
func renderTransparentSquareFull(b tileBlits, out Surface, pos PixelPos, src []byte, height int, mask MaskType) {
	dst := pos.Y*out.Width + pos.X
	prefix := InitPrefix(mask)
	line := 0

	for range height {
		drawWidth := DunFrameWidth
		for drawWidth > 0 {
			run := int(int8(src[line]))
			line++
			if run > 0 {
				b.renderLine(
					out.Pixels[dst:dst+run],
					src[line:line+run],
					mask,
					prefix-(DunFrameWidth-drawWidth),
				)
				line += run
			} else {
				run = -run
			}
			dst += run
			drawWidth -= run
		}
		dst -= out.Width + DunFrameWidth
		prefix += PrefixIncrement(mask)
	}
}

// renderTriangleLower is RenderTriangleLower (dun_render.cpp:452-473): 16
// growing rows; startX is line 0's x offset and xShift the per-row drift
// (left rows anchor at the right edge and drift -XStep; right rows anchor at
// x and stay put, the unrolled dst strides of the C++ two loops).
func renderTriangleLower(
	out Surface,
	pos PixelPos,
	src []byte,
	startX, xShift int,
	writer func(dst []uint8, src []byte),
) {
	line := 0

	for k := range LowerHeight {
		width := 2 * (k + 1)
		row := (pos.Y-k)*out.Width + pos.X + startX + xShift*k
		writer(out.Pixels[row:row+width], srcRow(src, line, width))
		line += width
	}
}

// renderTriangleUpper is RenderTriangleUpper (dun_render.cpp:475-493): 15
// shrinking rows; the xShift per row and the start row come from the caller's
// full flow.
func renderTriangleUpper(
	out Surface,
	pos PixelPos,
	src []byte,
	rowStart, xStart, xShift int,
	writer func(dst []uint8, src []byte),
) {
	line := 0

	for k := range TriangleUpperHeight {
		width := DunFrameWidth - 2 - 2*k
		row := (pos.Y-rowStart-k)*out.Width + pos.X + xStart + xShift*k
		writer(out.Pixels[row:row+width], srcRow(src, line, width))
		line += width
	}
}

// srcRow returns the src bytes for one row when the draw reads them; the fill
// blits (the black tile) never do and pass a short source, the way upstream
// hands over a nullptr (dun_render.cpp:1102-1108).
func srcRow(src []byte, line, width int) []byte {
	if line+width <= len(src) {
		return src[line : line+width]
	}

	return nil
}

// renderLeftTriangleFull is RenderLeftTriangleFull (dun_render.cpp:560-567).
func renderLeftTriangleFull(out Surface, pos PixelPos, src []byte, writer func(dst []uint8, src []byte)) {
	renderTriangleLower(out, pos, src, XStep*(LowerHeight-1), -XStep, writer)
	srcUpper := srcRow(src, LowerHeight*(LowerHeight+1), triangleUpperByteCount)
	renderTriangleUpper(out, PixelPos{X: pos.X, Y: pos.Y}, srcUpper, LowerHeight, XStep, XStep, writer)
}

// renderRightTriangleFull is RenderRightTriangleFull (dun_render.cpp:686-692).
func renderRightTriangleFull(out Surface, pos PixelPos, src []byte, writer func(dst []uint8, src []byte)) {
	renderTriangleLower(out, pos, src, 0, 0, writer)
	srcUpper := srcRow(src, LowerHeight*(LowerHeight+1), triangleUpperByteCount)
	renderTriangleUpper(out, PixelPos{X: pos.X, Y: pos.Y}, srcUpper, LowerHeight, 0, 0, writer)
}

// renderTrapezoidUpperHalf is RenderTrapezoidUpperHalf (dun_render.cpp:741-779).
func renderTrapezoidUpperHalf(b tileBlits, out Surface, pos PixelPos, src []byte, mask MaskType) {
	dst := pos.Y*out.Width + pos.X
	line := 0

	if mask == MaskLeft || mask == MaskRight {
		// the first line is always fully opaque (dun_render.cpp:745-750) to
		// avoid the width=0 blit call
		b.opaque(out.Pixels[dst:dst+DunFrameWidth], src[line:line+DunFrameWidth])
		line += DunFrameWidth
		dst -= out.Width

		prefixWidth := DunFrameWidth + PrefixIncrement(mask)
		if PrefixIncrement(mask) > 0 {
			prefixWidth = PrefixIncrement(mask)
		}
		for range TrapezoidUpperHeight - 1 {
			renderTransparentAndOpaque(b, out.Pixels[dst:dst+DunFrameWidth], src[line:line+DunFrameWidth], prefixWidth, mask)
			prefixWidth += PrefixIncrement(mask)
			line += DunFrameWidth
			dst -= out.Width
		}

		return
	}

	for range TrapezoidUpperHeight {
		b.renderLineTransparentOrOpaque(out.Pixels[dst:dst+DunFrameWidth], src[line:line+DunFrameWidth], mask)
		line += DunFrameWidth
		dst -= out.Width
	}
}

// renderTransparentAndOpaque is RenderLineTransparentAndOpaque
// (dun_render.cpp:218-235).
func renderTransparentAndOpaque(b tileBlits, dst []uint8, src []byte, prefixWidth int, mask MaskType) {
	if mask == MaskRight {
		b.opaque(dst[:prefixWidth], src[:prefixWidth])
		b.blended(dst[prefixWidth:], src[prefixWidth:])
	} else {
		b.blended(dst[:prefixWidth], src[:prefixWidth])
		b.opaque(dst[prefixWidth:], src[prefixWidth:])
	}
}

// renderLeftTrapezoidFull is RenderLeftTrapezoidFull (dun_render.cpp:803-810).
func renderLeftTrapezoidFull(b tileBlits, out Surface, pos PixelPos, src []byte, mask MaskType) {
	renderTriangleLower(out, pos, src, XStep*(LowerHeight-1), -XStep, triangleWriter(b, mask))
	srcRect := srcRow(src, LowerHeight*(LowerHeight+1), TrapezoidUpperHeight*DunFrameWidth)
	renderTrapezoidUpperHalf(b, out, PixelPos{X: pos.X, Y: pos.Y - LowerHeight}, srcRect, mask)
}

// renderRightTrapezoidFull is RenderRightTrapezoidFull (dun_render.cpp:846-851).
func renderRightTrapezoidFull(b tileBlits, out Surface, pos PixelPos, src []byte, mask MaskType) {
	renderTriangleLower(out, pos, src, 0, 0, triangleWriter(b, mask))
	srcRect := srcRow(src, LowerHeight*(LowerHeight+1), TrapezoidUpperHeight*DunFrameWidth)
	renderTrapezoidUpperHalf(b, out, PixelPos{X: pos.X, Y: pos.Y - LowerHeight}, srcRect, mask)
}
