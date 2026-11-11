package render_test

import (
	"bytes"
	"encoding/binary"
	"path/filepath"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/render"
)

// frame48GoldenLeads are the first re-encoded bytes of retail frame 48: the
// vendor LeftTriangle rows (2/4/6 wide) after the two-byte bloat strips,
// od-documented in docs/research/dg-graphics.md G2.
var frame48GoldenLeads = []byte{0x29, 0x29, 0x2a, 0x79, 0x29, 0x29, 0x28, 0x77, 0x77, 0x29, 0x77, 0x77}

// foliageFrame is a 32x32 pixel frame of palette indexes.
type foliageFrame [32][32]byte

// TestReencodeFrameSizesAndTable pins the re-encoded sizes and LE frame
// table of one fixture CEL with a frame per upstream type
// (reencode_dun_cels.cpp:206-234, :249-293).
func TestReencodeFrameSizesAndTable(t *testing.T) {
	// given
	squares := make([]byte, 1024)
	solidPass := make([]byte, 0, 64)

	for i := range squares {
		squares[i] = 9
	}

	for range 32 {
		solidPass = append(solidPass, byte(32))
		solidPass = append(solidPass, squares[0]) // one full solid row
	}

	triangle := readFrameFixture(t, 48)
	rightTrap := rightTrapezoidFixture()
	floorFrame := encodeTransparentSquare(upstreamFoliage())

	frames := []render.DunFrame{
		{Frame: 1, Info: render.DunFrameInfo{Type: render.TileSquare}},
		{Frame: 2, Info: render.DunFrameInfo{
			MicroTileIndex: 5, Type: render.TileTransparentSquare, Properties: render.TilePropertiesSolid,
		}},
		{Frame: 3, Info: render.DunFrameInfo{Type: render.TileLeftTriangle}},
		{Frame: 4, Info: render.DunFrameInfo{Type: render.TileRightTrapezoid}},
		{Frame: 5, Info: render.DunFrameInfo{MicroTileIndex: 0, Type: render.TileTransparentSquare}},
	}

	rawSource := testCel([][]byte{squares, solidPass, triangle, rightTrap, floorFrame})

	// when
	cels := render.ReencodeDungeonCels(rawSource, frames)

	// then
	if total := frameOffsetFixture(cels, 6); total != len(cels) {
		t.Fatalf("table end %d, file %d", total, len(cels))
	}

	expectFrameContent(t, cels, 1, len(squares), squares)
	expectFrameContent(t, cels, 2, len(solidPass), solidPass)
	expectFrameContent(t, cels, 3, 512, nil)
	expectFrameContent(t, cels, 4, 784, nil)

	// the floor pair keeps frame 5's partition: triangle, then foliage
	// (both the gtest port and the exact vendor strips cover the halves).
	start5 := frameOffsetFixture(cels, 5)
	if got := binary.LittleEndian.Uint32(cels[4*5:]); got == 0 {
		t.Fatal("frame 5 offset missing")
	}

	// the floor pair is the 512-byte triangle plus the foliage RLE stream:
	// the upstream gtest frame's foliage encodes thirteen one-byte
	// all-transparent rows, the nine-byte row 7 and the four-byte rows 0 and
	// 10 - 30 bytes total.
	if size := frameOffsetFixture(cels, 6) - start5; size != 512+30 {
		t.Fatalf("floor pair size %d, want 512 triangle + 30 foliage bytes", size)
	}
}

// TestFloorFoliageKeepsItsPosition ports the upstream gtest of the same
// name: a floor transparent square's solid pixels of the top half land in
// the extracted foliage stream at the same (x, y).
func TestFloorFoliageKeepsItsPosition(t *testing.T) {
	// given
	frames := []render.DunFrame{
		{Frame: 1, Info: render.DunFrameInfo{MicroTileIndex: 0, Type: render.TileTransparentSquare}},
	}

	rawSource := testCel([][]byte{encodeTransparentSquare(upstreamFoliage())})

	// when
	cels := render.ReencodeDungeonCels(rawSource, frames)

	// then: the foliage stream starts at frame 1 + the triangle
	foliageStart := frameOffsetFixture(cels, 1) + 512
	tail := decodeTransparentSquare(cels[foliageStart:], 16)

	frame := upstreamFoliage()
	for y := range 16 {
		for x := range 32 - 2*y {
			if tail[y][x] != frame[y][x] {
				t.Errorf("foliage[%d][%d] = %d, want %d", y, x, tail[y][x], frame[y][x])
			}
		}
	}
}

// TestReencodeLeftTriangleStripsVendorPadding re-encodes retail frame 48 and
// pins the vendor row bytes after the bloat strips are gone
// (docs/research/dg-graphics.md G2 worked example).
func TestReencodeLeftTriangleStripsVendorPadding(t *testing.T) {
	// given
	triangle := readFrameFixture(t, 48)

	frames := []render.DunFrame{
		{Frame: 1, Info: render.DunFrameInfo{Type: render.TileLeftTriangle}},
	}

	rawSource := testCel([][]byte{triangle})

	// when
	cels := render.ReencodeDungeonCels(rawSource, frames)

	// then
	if got := len(cels) - (2+len(frames))*4; got != 512 {
		t.Fatalf("reencoded frame = %d bytes, want 512", got)
	}

	start := frameOffsetFixture(cels, 1)
	for i, want := range frame48GoldenLeads {
		if cels[start+i] != want {
			t.Fatalf("byte %d = %#x, want %#x", i, cels[start+i], want)
		}
	}
}

func equalBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// testCel assembles an in-memory source CEL with LE offsets
// (the layout ReencodeDungeonCels walks, reencode_dun_cels.cpp:242-257).
func testCel(sources [][]byte) []byte {
	out := make([]byte, 4)

	out = writeLE32Fixture(out, 0, uint32(len(sources)))

	written := (2 + len(sources)) * 4

	for i, source := range sources {
		out = writeLE32Fixture(out, i+1, uint32(written))
		written += len(source)
	}

	out = writeLE32Fixture(out, len(sources)+1, uint32(written))

	for _, source := range sources {
		out = append(out, source...)
	}

	return out
}

// writeLE32Fixture stores v little-endian at dst[4*i:], growing dst.
func writeLE32Fixture(dst []byte, i int, v uint32) []byte {
	if 4*(i+1) > len(dst) {
		grown := make([]byte, 4*(i+1))
		copy(grown, dst)
		dst = grown
	}

	binary.LittleEndian.PutUint32(dst[4*i:], v)

	return dst
}

// frameOffsetFixture reads one frame's start offset from a re-encoded table.
func frameOffsetFixture(cels []byte, frame int) int {
	return int(binary.LittleEndian.Uint32(cels[4*frame : 4*frame+4]))
}

// encodeTransparentSquare encodes one frame bottom-up as row runs
// (upstream gtest helper, test/reencode_dun_cels_test.cpp:23-42).
func encodeTransparentSquare(frame foliageFrame) []byte {
	out := []byte{}

	for row := range render.DunFrameHeight {
		out = encodeRowRuns(frame[render.DunFrameHeight-1-row], out)
	}

	return out
}

// encodeRowRuns runs one top-index row, in runs that do not cross rows.
func encodeRowRuns(row [32]byte, out []byte) []byte {
	for x := 0; x < render.DunFrameWidth; {
		solid := row[x] != 0
		run := solidRunLength(row, x, solid)

		if solid {
			out = append(out, byte(run))
			out = append(out, row[x:x+run]...)
		} else {
			out = append(out, byte(-run))
		}

		x += run
	}

	return out
}

// solidRunLength walk the run of one pixel's solidity.
func solidRunLength(row [32]byte, from int, solid bool) int {
	run := 0

	for from+run < render.DunFrameWidth && (row[from+run] != 0) == solid {
		run++
	}

	return run
}

// decodeTransparentSquare reads back the top height rows of an RLE stream
// (upstream gtest helper, test/reencode_dun_cels_test.cpp:44-61).
func decodeTransparentSquare(src []byte, height int) [32][32]byte {
	var out [32][32]byte

	for y := height - 1; y >= 0; y-- {
		for x := 0; x < 32; {
			run := int(int8(src[0]))
			src = src[1:]

			if run > 0 {
				copy(out[y][x:x+run], src[:run])
				src = src[run:]
				x += run
			} else {
				x -= run
			}
		}
	}

	return out
}

// rightTrapezoidFixture fabricates a right-layout trapezoid source: the
// growing rows of widths 2..32 with the 2-byte pad AFTER even rows, then the
// flat 32x16 rect.
func rightTrapezoidFixture() []byte {
	src := []byte{}

	for width, row := 2, 0; row < 16; width, row = width+2, row+1 {
		src = append(src, bytes.Repeat([]byte{byte(row%32 + 7)}, width)...)

		if row%2 == 0 {
			src = append(src, 0, 0)
		}
	}

	return append(src, bytes.Repeat([]byte{0x11}, 512)...)
}

// upstreamFoliage is the exact fixture frame of the gtest.
func upstreamFoliage() foliageFrame {
	var frame foliageFrame

	for y := 16; y < 32; y++ { // the floor
		for x := range frame[y] {
			frame[y][x] = 100
		}
	}

	// foliage outside the left floor triangle (x >= 32 - 2y in rows 1-15),
	// in runs that end before the end of the row and at it
	frame[7][2], frame[7][3], frame[7][4] = 10, 11, 12
	frame[7][6] = 13
	frame[10][0], frame[10][1] = 14, 15
	frame[0][30], frame[0][31] = 16, 17 // row 0 runs may reach the row end

	return frame
}

// expectFrameContent compares one re-encoded frame with its source bytes.
func expectFrameContent(t *testing.T, cels []byte, frame, size int, want []byte) {
	t.Helper()

	start := frameOffsetFixture(cels, frame)
	end := frameOffsetFixture(cels, frame+1)

	if end-start != size {
		t.Fatalf("frame %d = %d bytes, want %d", frame, end-start, size)
	}

	if want != nil && !equalBytes(cels[start:end], want) {
		t.Errorf("frame %d bytes mismatch the source", frame)
	}
}

// readFrameFixture extracts one frame of the committed l1-head fixture
// (tests never touch MPQs or C++ trees).
func readFrameFixture(t *testing.T, frame int) []byte {
	t.Helper()

	data := readFlatSize(t, filepath.Join(fixtureDir, "tiles", "cel", "l1-head.hex"))
	start := int(binary.LittleEndian.Uint32(data[4*frame : 4*frame+4]))
	end := int(binary.LittleEndian.Uint32(data[4*(frame+1) : 4*(frame+2)]))

	return data[start:end]
}
