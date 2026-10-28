package assets_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/assets"
)

// clxExpectation is one parsed <flat>.sha256 row.
type clxExpectation struct {
	digest               string
	bytes                int
	width, height, count int
}

func TestDecodeCLX(t *testing.T) {
	// when a hand-built stream is decoded through decodeSingleCLX, or a multi-frame
	// header through DecodeCLX
	t.Run("single frame literal run", func(t *testing.T) {
		// 4x2 frame: row 0 = [1,2,3,4], row 1 = [5,5,5,5]
		f := decodeSingleCLX(t, append(cl2L(1, 2, 3, 4), cl2F(5, 4)...), 4, 2)
		clxExpectFrame(t, f, []byte{1, 2, 3, 4, 5, 5, 5, 5}, 4, 2)
	})

	t.Run("transparent run", func(t *testing.T) {
		// 4x1: 2 transparent tiles followed by 2 literals; index 0 is transparent
		f := decodeSingleCLX(t, append(cl2T(2), cl2L(7, 8)...), 4, 1)
		clxExpectFrame(t, f, []byte{0, 0, 7, 8}, 0, 0)
	})

	t.Run("transparent run crosses line boundary", func(t *testing.T) {
		// 2x2: one command of 3 transparent pixels overruns row 0 into row 1;
		// the decoder writes a flat buffer so the overrun naturally lands on row 1.
		f := decodeSingleCLX(t, append(cl2T(3), cl2L(9)...), 2, 2)
		clxExpectFrame(t, f, []byte{0, 0, 0, 9}, 0, 0)
	})

	t.Run("fill run wraps control byte past 0x80", func(t *testing.T) {
		// control 0xBF - 1 = 0xBE encodes run length 1; the wrap math is checked
		// against clx_decode.hpp GetClxOpaqueFillWidth.
		f := decodeSingleCLX(t, cl2F(0xAA, 1), 1, 1)
		if f.Pixels[0] != 0xAA {
			t.Fatalf("pixel = %d, want 0xAA", f.Pixels[0])
		}
	})

	t.Run("skip-line via full-width transparent command", func(t *testing.T) {
		// 3x3: row 0 literals, row 1 skipped with T(3), row 2 literals.
		stream := append(append(cl2L(1, 2, 3), cl2T(3)...), cl2L(7, 8, 9)...)
		f := decodeSingleCLX(t, stream, 3, 3)
		clxExpectFrame(t, f, []byte{1, 2, 3, 0, 0, 0, 7, 8, 9}, 0, 0)
	})

	t.Run("multiple frames decode independently", func(t *testing.T) {
		clxExpectIndependentFrames(t)
	})

	// then the frame pixels and dims match the encoder input, and malformed input is
	// rejected instead of decoded
	t.Run("sheet of lists", func(t *testing.T) {
		clxExpectSheetDecodes(t)
	})

	t.Run("truncated literal run errors", func(t *testing.T) {
		clxExpectTruncationError(t)
	})
}

func TestCLXDecodeMatchesCommittedHashes(t *testing.T) {
	// given
	// The committed small-CLX corpus: <flat>.clx.hex holds the file bytes and
	// <flat>.sha256 the independently computed decode digest, byte count, dims
	// and frame count (internal/testdata/gfx/README.md).
	names := clxFixtureNames(t)
	if len(names) < 20 {
		t.Fatalf("committed CLX corpus = %d files, want at least 20", len(names))
	}
	// when every file is decoded from its committed hex
	decoded := make([]*assets.ClxSpriteList, len(names))
	for i, name := range names {
		decoded[i] = decodeCLXFixture(t, name)
	}

	// then
	for i, name := range names {
		t.Run(name, func(t *testing.T) {
			expectCLXFixture(t, readCLXExpectation(t, name), decoded[i])
		})
	}
}

func TestCLXGoldenFrame0Bytes(t *testing.T) {
	// given
	// Three frames dumped in full; the rest of the corpus is digest-only.
	names := []string{"data_health", "data_talkbutton", "data_boxrightend"}
	// when each file is decoded from its committed hex
	frames := make([]assets.ClxFrame, len(names))
	for i, name := range names {
		frames[i] = decodeCLXFixture(t, name).Frames[0]
	}

	// then frame 0 pixels equal the committed dump
	for i, name := range names {
		t.Run(name, func(t *testing.T) {
			expectFrameZero(t, name, frames[i])
		})
	}
}

func clxExpectIndependentFrames(t *testing.T) {
	t.Helper()

	data := buildCLX([][]byte{cl2L(1, 1), cl2L(2, 2)}, 2, 1)

	frames, err := assets.DecodeCLX(data)
	if err != nil {
		t.Fatalf("assets.DecodeCLX: %v", err)
	}

	if len(frames.Frames) != 2 {
		t.Fatalf("frames = %d, want 2", len(frames.Frames))
	}

	if frames.Frames[0].Pixels[0] != 1 || frames.Frames[1].Pixels[0] != 2 {
		t.Fatalf("frame pixels wrong: %v %v", frames.Frames[0].Pixels, frames.Frames[1].Pixels)
	}
}

// clxExpectSheetDecodes pins multi-group CLX support: a buffer whose offset
// table ends before the buffer end is a sheet of lists
// (clx_sprite.hpp:447-457, cl2_to_clx.cpp:23-28).
func clxExpectSheetDecodes(t *testing.T) {
	t.Helper()

	data := buildCLXSheet(t, [][][]byte{
		{cl2L(1, 1)},
		{cl2L(2, 2), cl2L(3, 3)},
	}, 2, 1)

	sheet, err := assets.DecodeCLX(data)
	if err != nil {
		t.Fatalf("assets.DecodeCLX: %v", err)
	}

	expectSheetShape(t, sheet, []int{1, 2}, [][]byte{{1, 1}, {2, 2}, {3, 3}})

	second := sheet.Group(1)
	if len(second) != 2 || !bytes.Equal(second[0].Pixels, []byte{2, 2}) {
		t.Fatalf("group 1 = %+v, want the two frames of the second list", second)
	}
}

// expectSheetShape checks group sizes and per-frame pixels of a decoded sheet.
func expectSheetShape(t *testing.T, sheet *assets.ClxSpriteList, sizes []int, want [][]byte) {
	t.Helper()

	if fmt.Sprint(sheet.GroupSizes) != fmt.Sprint(sizes) {
		t.Fatalf("group sizes = %v, want %v", sheet.GroupSizes, sizes)
	}

	if len(sheet.Frames) != len(want) {
		t.Fatalf("frames = %d, want %d", len(sheet.Frames), len(want))
	}

	for i := range want {
		if !bytes.Equal(sheet.Frames[i].Pixels, want[i]) {
			t.Fatalf("frame %d pixels = %v, want %v", i, sheet.Frames[i].Pixels, want[i])
		}
	}
}

// buildCLXSheet lays out a group-offset table followed by one CLX list per group.
func buildCLXSheet(t *testing.T, groups [][][]byte, width, height int) []byte {
	t.Helper()

	offsets := make([]uint32, len(groups))
	body := make([]byte, 0, 64)

	for i, frames := range groups {
		offsets[i] = uint32(4*len(groups) + len(body))
		body = append(body, buildCLX(frames, width, height)...)
	}

	out := make([]byte, 0, 4*len(groups)+len(body))
	for _, off := range offsets {
		out = binary.LittleEndian.AppendUint32(out, off)
	}

	return append(out, body...)
}

func clxExpectTruncationError(t *testing.T) {
	t.Helper()

	data := buildCLX([][]byte{{0xFE, 1}}, 2, 1) // literal run of 2, only 1 byte present
	if _, err := assets.DecodeCLX(data); err == nil {
		t.Fatal("expected truncation error, got nil")
	}
}

func clxExpectFrame(t *testing.T, f assets.ClxFrame, want []byte, wantW, wantH int) {
	t.Helper()

	if wantW != 0 && (f.Width != wantW || f.Height != wantH) {
		t.Fatalf("dims = %dx%d, want %dx%d", f.Width, f.Height, wantW, wantH)
	}

	if !bytes.Equal(f.Pixels, want) {
		t.Fatalf("pixels = %v, want %v", f.Pixels, want)
	}
}

func buildCLX(frames [][]byte, width, height int) []byte {
	numFrames := len(frames)

	size := 4 + 4*(numFrames+1)
	for _, f := range frames {
		size += 6 + len(f)
	}

	buf := make([]byte, 0, size)
	off := uint32(4 + 4*(numFrames+1))
	buf = binary.LittleEndian.AppendUint32(buf, uint32(numFrames))

	buf = binary.LittleEndian.AppendUint32(buf, off)
	for _, f := range frames {
		off += 6 + uint32(len(f))
		buf = binary.LittleEndian.AppendUint32(buf, off)
	}

	for _, f := range frames {
		buf = binary.LittleEndian.AppendUint16(buf, 6) // header size
		buf = binary.LittleEndian.AppendUint16(buf, uint16(width))
		buf = binary.LittleEndian.AppendUint16(buf, uint16(height))
		buf = append(buf, f...)
	}

	return buf
}

func cl2T(n byte) []byte { return []byte{n} }

func cl2F(v, n byte) []byte {
	return []byte{0xBF - n, v}
}

func cl2L(px ...byte) []byte { return append([]byte{byte(-int8(len(px)))}, px...) }

func decodeSingleCLX(t *testing.T, stream []byte, w, h int) assets.ClxFrame {
	t.Helper()

	list, err := assets.DecodeCLX(buildCLX([][]byte{stream}, w, h))
	if err != nil {
		t.Fatalf("assets.DecodeCLX: %v", err)
	}

	if len(list.Frames) != 1 {
		t.Fatalf("frames = %d, want 1", len(list.Frames))
	}

	return list.Frames[0]
}

func clxFixtureNames(tb testing.TB) []string {
	tb.Helper()

	entries, err := os.ReadDir(filepath.Join(gfxFixtureDir, "clx"))
	if err != nil {
		tb.Fatalf("read committed CLX corpus: %v", err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if name, ok := strings.CutSuffix(entry.Name(), ".sha256"); ok {
			names = append(names, name)
		}
	}

	return names
}

func clxHexPath(name string) string {
	return filepath.Join(gfxFixtureDir, "clx", name+".clx.hex")
}

func readCLXExpectation(t *testing.T, name string) clxExpectation {
	t.Helper()

	line, err := os.ReadFile(filepath.Join(gfxFixtureDir, "clx", name+".sha256"))
	if err != nil {
		t.Fatalf("read %s.sha256: %v", name, err)
	}

	fields := strings.Fields(string(line))
	if len(fields) != 4 {
		t.Fatalf("%s.sha256 has %d fields, want 4", name, len(fields))
	}

	var exp clxExpectation
	exp.digest = fields[0]
	exp.bytes, err = strconv.Atoi(fields[1])
	if err != nil {
		t.Fatalf("%s.sha256 bytes: %v", name, err)
	}

	if _, err := fmt.Sscanf(fields[2], "%dx%d", &exp.width, &exp.height); err != nil {
		t.Fatalf("%s.sha256 dims: %v", name, err)
	}

	exp.count, err = strconv.Atoi(fields[3])
	if err != nil {
		t.Fatalf("%s.sha256 frames: %v", name, err)
	}

	return exp
}

func decodeCLXFixture(t *testing.T, name string) *assets.ClxSpriteList {
	t.Helper()

	list, err := assets.DecodeCLX(readHexFixture(t, clxHexPath(name)))
	if err != nil {
		t.Fatalf("%s: DecodeCLX: %v", name, err)
	}

	return list
}

// expectFrameZero compares a decoded frame 0 with its committed dump.
func expectFrameZero(t *testing.T, name string, frame assets.ClxFrame) {
	t.Helper()

	want := readHexFixture(t, filepath.Join(gfxFixtureDir, "clx", name+".frame0.hex"))
	if !bytes.Equal(frame.Pixels, want) {
		t.Fatalf("frame 0 pixels differ from %s.frame0.hex", name)
	}
}

// clxDigest hashes the row-major indices of every frame.
func clxDigest(t *testing.T, list *assets.ClxSpriteList) []byte {
	t.Helper()

	sum := sha256.New()
	for _, frame := range list.Frames {
		sum.Write(frame.Pixels)
	}

	return sum.Sum(nil)
}

// expectCLXFixture checks the invariants the producer asserted: header parses,
// frame count, uniform dims, sum(w*h) == len(indices) and the digest.
func expectCLXFixture(t *testing.T, want clxExpectation, list *assets.ClxSpriteList) {
	t.Helper()

	if len(list.Frames) != want.count {
		t.Fatalf("frames = %d, want %d", len(list.Frames), want.count)
	}

	for i, frame := range list.Frames {
		if frame.Width != want.width || frame.Height != want.height {
			t.Fatalf("frame %d dims = %dx%d, want %dx%d", i, frame.Width, frame.Height, want.width, want.height)
		}
	}

	if got := want.width * want.height * want.count; got != want.bytes {
		t.Fatalf("sum(w*h) = %d, want %d", got, want.bytes)
	}

	if gotHex := hex.EncodeToString(clxDigest(t, list)); gotHex != want.digest {
		t.Fatalf("digest = %s, want %s", gotHex, want.digest)
	}
}
