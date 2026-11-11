package render

// This file ports Source/levels/reencode_dun_cels.cpp verbatim: the padding
// strips (:24-106), the floor foliage split (:108-200), the size calculator
// (:202-238), the frame-table rewrite (:242-303) and the cell-block
// adjustments (:305-320). Every helper keeps the upstream name so a reader
// can diff against the C++ line by line.

const (
	// DunFrameWidth is DunFrameWidth, half the tile width (dun_tile.hpp:123).
	DunFrameWidth = 32

	// DunFrameHeight is DunFrameHeight, the tile height (dun_tile.hpp:126).
	DunFrameHeight = 32

	// numFirstHalfStripes is the lower triangle's growing-stripe count
	// (reencode_dun_cels.cpp:26).
	numFirstHalfStripes = 8

	// numUpperStripes is the upper triangle's shrinking-stripe count plus
	// the final row (reencode_dun_cels.cpp:43-54).
	numUpperStripes = 7

	// reencodedTriangleFrameSize is ReencodedTriangleFrameSize = 544-32
	// (dun_tile.hpp:131): 512 copied bytes, 32 padding bytes dropped.
	reencodedTriangleFrameSize = 544 - 32

	// reencodedTrapezoidFrameSize is ReencodedTrapezoidFrameSize = 800-16
	// (dun_tile.hpp:132): 784 copied bytes, 16 padding bytes dropped.
	reencodedTrapezoidFrameSize = 800 - 16
)

type foliagePacker struct {
	dst  []byte
	used int
	run  int // the current run: positive solid, negative transparent
}

// GetReencodedSize walks the frame list over the source frame table; the
// floor branch measures by running the foliage split into scratch, exactly
// upstream (reencode_dun_cels.cpp:202-238).
func GetReencodedSize(dungeonCels []byte, frames []DunFrame) int {
	result := (2 + len(frames)) * 4

	for _, frame := range frames {
		result += reencodedFrameSize(dungeonCels, frame)
	}

	return result
}

// ReencodeDungeonCels rewrites a dungeon CEL for faster rendering: it strips
// triangle/trapezoid padding, splits floor transparent squares into their
// foliage pair, and renumbers the frame table to the kept frames only
// (reencode_dun_cels.cpp:242-303).
func ReencodeDungeonCels(dungeonCels []byte, frames []DunFrame) []byte {
	outSize := GetReencodedSize(dungeonCels, frames)
	result := make([]byte, outSize)
	writeLE32(result, uint32(len(frames)))

	header := (2 + len(frames)) * 4
	lookup := result[4:]
	out := result[header:]
	written := 0

	for _, frame := range frames {
		writeLE32(lookup, uint32(header+written))
		lookup = lookup[4:]

		src := dungeonCels[le32Offset(dungeonCels, int(frame.Frame)):]
		produced := 0

		switch frame.Info.Type {
		case TileTransparentSquare:
			produced = reencodeTransparentSquare(out, dungeonCels, frame, src)
		case TileSquare:
			produced = DunFrameWidth * DunFrameHeight
			copy(out, src[:produced])
		case TileLeftTriangle:
			produced = reencodeDungeonCelsLeftTriangle(out, src)
		case TileRightTriangle:
			produced = reencodeDungeonCelsRightTriangle(out, src)
		case TileLeftTrapezoid:
			produced = reencodeDungeonCelsLeftTrapezoid(out, src)
		case TileRightTrapezoid:
			produced = reencodeDungeonCelsRightTrapezoid(out, src)
		default:
			// unreachable: TileType has six valid shapes, every one handled
		}

		out = out[produced:]

		written += produced
	}

	writeLE32(lookup, uint32(outSize))

	return result
}

// reencodedFrameSize returns one frame's re-encoded byte count
// (reencode_dun_cels.cpp:206-234).
func reencodedFrameSize(dungeonCels []byte, frame DunFrame) int {
	switch frame.Info.Type {
	case TileSquare:
		return DunFrameWidth * DunFrameHeight
	case TileTransparentSquare:
		if !frame.Info.IsFloor() {
			return passthroughFrameSize(dungeonCels, frame.Frame)
		}

		scratch := make([]byte, DunFrameWidth*DunFrameHeight)

		return reencodeFloorWithFoliage(
			scratch,
			dungeonCels[le32Offset(dungeonCels, int(frame.Frame)):],
			floorFoliageType(frame.Info))
	case TileLeftTrapezoid, TileRightTrapezoid:
		return reencodedTrapezoidFrameSize
	default:
		return reencodedTriangleFrameSize
	}
}

// reencodeTransparentSquare encodes a MIN transparent square: floors split
// into the foliage pair, the others pass through unchanged
// (reencode_dun_cels.cpp:264-275).
func reencodeTransparentSquare(out, dungeonCels []byte, frame DunFrame, src []byte) int {
	if !frame.Info.IsFloor() {
		size := passthroughFrameSize(dungeonCels, frame.Frame)
		copy(out, src[:size])

		return size
	}

	return reencodeFloorWithFoliage(out, src, floorFoliageType(frame.Info))
}

// floorFoliageType picks the foliage triangle half of a floor pair
// (reencode_dun_cels.cpp:215, :266).
func floorFoliageType(info DunFrameInfo) TileType {
	if info.IsFloorLeft() {
		return TileLeftTriangle
	}

	return TileRightTriangle
}

// passthroughFrameSize is the source frame's own length via the next
// offset (reencode_dun_cels.cpp:219-221).
func passthroughFrameSize(dungeonCels []byte, frame uint16) int {
	return le32Offset(dungeonCels, int(frame+1)) - le32Offset(dungeonCels, int(frame))
}

// le32Offset reads one frame-table entry; the retail tables are little-
// endian on the master X86 path (Swap32LE is a no-op on LE hosts, the
// dg-graphics.md G2 open question decided this: LE read + BE parity job).
func le32Offset(dungeonCels []byte, frame int) int {
	data := dungeonCels[4*frame:]

	return int(uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24)
}

// writeLE32 stores v little-endian (WriteLE32, utils/endian_write.hpp).
func writeLE32(dst []byte, v uint32) {
	dst[0] = byte(v)
	dst[1] = byte(v >> 8)
	dst[2] = byte(v >> 16)
	dst[3] = byte(v >> 24)
}

// reencodeDungeonCelsLeftTriangleLower strips the padding of the triangle's
// growing lower half (reencode_dun_cels.cpp:23-37).
func reencodeDungeonCelsLeftTriangleLower(dst, src []byte) (dpos, spos int) {
	width := 0

	for range numFirstHalfStripes {
		spos += 2 // skips the two zero bytes (aka bloat).
		width += 2

		copy(dst[dpos:dpos+width], src[spos:spos+width])
		spos += width
		dpos += width

		width += 2

		copy(dst[dpos:dpos+width], src[spos:spos+width])
		spos += width
		dpos += width
	}

	return dpos, spos
}

// reencodeDungeonCelsLeftTriangle copies a left triangle without its bottom
// half's padding (reencode_dun_cels.cpp:39-58).
func reencodeDungeonCelsLeftTriangle(dst, src []byte) int {
	dpos, spos := reencodeDungeonCelsLeftTriangleLower(dst, src)

	width := DunFrameWidth

	for range numUpperStripes {
		spos += 2 // skips the two zero bytes (aka bloat).
		width -= 2

		copy(dst[dpos:dpos+width], src[spos:spos+width])
		spos += width
		dpos += width

		width -= 2

		copy(dst[dpos:dpos+width], src[spos:spos+width])
		spos += width
		dpos += width
	}

	spos += 2 // skips the two zero bytes (aka bloat).
	width -= 2

	copy(dst[dpos:dpos+width], src[spos:spos+width])

	return reencodedTriangleFrameSize
}

// reencodeDungeonCelsRightTriangleLower, same as the left-half strip but the
// padding follows each even row (reencode_dun_cels.cpp:60-73).
func reencodeDungeonCelsRightTriangleLower(dst, src []byte) (dpos, spos int) {
	width := 0

	for range numFirstHalfStripes {
		width += 2

		copy(dst[dpos:dpos+width], src[spos:spos+width])
		spos += width + 2 // skips the two zero bytes (aka bloat).
		dpos += width

		width += 2

		copy(dst[dpos:dpos+width], src[spos:spos+width])
		spos += width
		dpos += width
	}

	return dpos, spos
}

// reencodeDungeonCelsRightTriangle copies a right triangle without its
// bottom half's padding (reencode_dun_cels.cpp:75-92).
func reencodeDungeonCelsRightTriangle(dst, src []byte) int {
	dpos, spos := reencodeDungeonCelsRightTriangleLower(dst, src)

	width := DunFrameWidth

	for range numUpperStripes {
		width -= 2

		copy(dst[dpos:dpos+width], src[spos:spos+width])
		spos += width + 2 // skips the two zero bytes (aka bloat).
		dpos += width

		width -= 2

		copy(dst[dpos:dpos+width], src[spos:spos+width])
		spos += width
		dpos += width
	}

	width -= 2

	copy(dst[dpos:dpos+width], src[spos:spos+width])

	return reencodedTriangleFrameSize
}

// reencodeDungeonCelsLeftTrapezoid keeps the triangle lower and the flat
// rectangle above (reencode_dun_cels.cpp:94-99).
func reencodeDungeonCelsLeftTrapezoid(dst, src []byte) int {
	dpos, spos := reencodeDungeonCelsLeftTriangleLower(dst, src)
	copy(dst[dpos:dpos+DunFrameWidth*16], src[spos:spos+DunFrameWidth*16])

	return reencodedTrapezoidFrameSize
}

// reencodeDungeonCelsRightTrapezoid, mirrored right variant
// (reencode_dun_cels.cpp:101-106).
func reencodeDungeonCelsRightTrapezoid(dst, src []byte) int {
	dpos, spos := reencodeDungeonCelsRightTriangleLower(dst, src)
	copy(dst[dpos:dpos+DunFrameWidth*16], src[spos:spos+DunFrameWidth*16])

	return reencodedTrapezoidFrameSize
}

// reencodeTransparentSquareFlat decodes an RLE transparent square into the
// flat surface buffer bottom-up (reencode_dun_cels.cpp:108-124). It is the
// decode side of RenderTransparentSquare, used by the foliage split.
func reencodeTransparentSquareFlat(dst, src []byte, line *int) {
	// Upstream rewinds `dst -= 2 * DunFrameWidth` per row so the decoded rows
	// land bottom-row first (reencode_dun_cels.cpp:110); the flat walk keeps
	// that order by flipping the row index.
	for row := range DunFrameHeight {
		at := dst[(DunFrameHeight-1-row)*DunFrameWidth:]
		drawWidth := DunFrameWidth
		for drawWidth > 0 {
			v := int(int8(src[*line]))
			*line++
			if v > 0 {
				copy(at[:v], src[*line:(*line)+v])
				*line += v
			} else {
				v = -v
			}
			at = at[v:]
			drawWidth -= v
		}
	}
}

// extractFoliageLeftTriangle copies the left-triangle pixels into dst and
// zeroes them in the surface (reencode_dun_cels.cpp:126-138). The 31 rows
// match the triangle: 16 growing bottom rows, then 15 shrinking top rows.
func extractFoliageLeftTriangle(dst, surface []byte) int {
	used := 0

	for stripe := range DunFrameHeight / 2 {
		w := stripe*2 + 2
		y := DunFrameHeight - 1 - stripe
		row := surface[DunFrameWidth*y : DunFrameWidth*(y+1)]
		copy(dst[used:used+w], row[DunFrameWidth-w:])
		clear(row[DunFrameWidth-w:])
		used += w
	}

	for stripe := range DunFrameHeight/2 - 1 {
		w := DunFrameWidth - 2 - stripe*2
		y := DunFrameHeight/2 - 1 - stripe
		row := surface[DunFrameWidth*y : DunFrameWidth*(y+1)]
		copy(dst[used:used+w], row[DunFrameWidth-w:])
		clear(row[DunFrameWidth-w:])
		used += w
	}

	return used
}

// extractFoliageRightTriangle, mirrored right variant
// (reencode_dun_cels.cpp:140-152).
func extractFoliageRightTriangle(dst, surface []byte) int {
	used := 0

	for stripe := range DunFrameHeight / 2 {
		w := stripe*2 + 2
		y := DunFrameHeight - 1 - stripe
		row := surface[DunFrameWidth*y : DunFrameWidth*(y+1)]
		copy(dst[used:used+w], row)
		clear(row[:w])
		used += w
	}

	for stripe := range DunFrameHeight/2 - 1 {
		w := DunFrameWidth - 2 - stripe*2
		y := DunFrameHeight/2 - 1 - stripe
		row := surface[DunFrameWidth*y : DunFrameWidth*(y+1)]
		copy(dst[used:used+w], row)
		clear(row[:w])
		used += w
	}

	return used
}

// extractFoliageTransparentSquare re-encodes the surface's upper half into
// the RLE stream of the foliage frame (reencode_dun_cels.cpp:154-187).
func extractFoliageTransparentSquare(dst, surface []byte) int {
	used := 0

	for row := range DunFrameHeight / 2 {
		rowStart := DunFrameWidth * (DunFrameHeight/2 - 1 - row)

		used += encodeFoliageRow(dst[used:], surface[rowStart:rowStart+DunFrameWidth])
	}

	return used
}

// encodeFoliageRow re-encodes one leaf row into its RLE runs
// (reencode_dun_cels.cpp:160-186). Runs never cross rows: a solid run flushes
// the moment a transparent pixel arrives, from its row prefix.
func encodeFoliageRow(dst, row []byte) int {
	packer := foliagePacker{dst: dst}

	for read, index := range row {
		packer.push(index, row, read)
	}

	return packer.flushEnd(row)
}

// foliagePacker accumulates one row's runs and writes them in row order.
// push accounts one pixel, flushing the opposite run at the boundary; the
// flushed counter restarts at 0 like the upstream locals (transparentRun /
// solidRun, reencode_dun_cels.cpp:163-175).
func (p *foliagePacker) push(index byte, row []byte, read int) {
	if index != 0 {
		if p.run < 0 {
			p.used += p.writeTransparent(-p.run)
			p.run = 0
		}
		p.run++

		return
	}

	if p.run > 0 {
		p.used += p.writeSolid(row[read-p.run:read], p.run)
		p.run = 0
	}

	p.run--
}

// flushEnd writes whichever run is active at the row's end.
func (p *foliagePacker) flushEnd(row []byte) int {
	if p.run > 0 {
		return p.used + p.writeSolid(row[len(row)-p.run:], p.run)
	}

	if p.run < 0 {
		return p.used + p.writeTransparent(-p.run)
	}

	return p.used
}

// writeTransparent stores one negative-run marker.
func (p *foliagePacker) writeTransparent(run int) int {
	p.dst[p.used] = byte(-run)

	return 1
}

// writeSolid stores a positive-run marker plus its pixel bytes.
func (p *foliagePacker) writeSolid(pixels []byte, run int) int {
	p.dst[p.used] = byte(run)
	copy(p.dst[p.used+1:p.used+1+run], pixels)

	return 1 + run
}

// reencodeFloorWithFoliage splits a floor transparent square into the
// triangle-then-foliage pair and returns the bytes it wrote into dst; the
// foliage square starts right after the extracted triangle bytes
// (reencode_dun_cels.cpp:189-200).
func reencodeFloorWithFoliage(dst, src []byte, tileType TileType) int {
	surface := make([]byte, DunFrameWidth*DunFrameHeight)

	cursor := 0
	reencodeTransparentSquareFlat(surface, src, &cursor)
	used := 0
	if tileType == TileLeftTriangle {
		used = extractFoliageLeftTriangle(dst, surface)
	} else {
		used = extractFoliageRightTriangle(dst, surface)
	}

	return used + extractFoliageTransparentSquare(dst[used:], surface)
}
