package assets_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/assets"
)

func TestDecodeCL2(t *testing.T) {
	// when
	subtests := []struct {
		name string
		run  func(*testing.T)
	}{
		{name: "literal and fill runs", run: cl2ExpectLiteralAndFillRuns},
		{name: "transparent run crosses line boundary", run: cl2ExpectTransparentRunCrossesLine},
		{name: "skip-line", run: cl2ExpectSkipLine},
		{name: "multiple frames with distinct offsets", run: cl2ExpectMultipleFrameOffsets},
		{name: "incomplete pixels error", run: cl2ExpectIncompletePixelsError},
		{name: "literal run at the 0xBF boundary", run: cl2ExpectLiteralRunAtBF},
		{name: "fill run at the 0xBE boundary", run: cl2ExpectFillRunAtBE},
		{name: "run width boundaries", run: cl2ExpectRunWidthBoundaries},
		{name: "sheet of lists", run: cl2ExpectSheetDecodes},
	}
	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

func BenchmarkCL2Decode(b *testing.B) {
	// Synthetic 64x64 frame: 64 rows of 32 literals + 32 transparent.
	// Each row costs one literal-run header + 32 literals + one transparent run.
	stream := make([]byte, 0, 64*(1+32+1))

	for range 64 {
		lit := make([]byte, 32)
		for i := range lit {
			lit[i] = byte(i)
		}

		stream = append(stream, cl2L(lit...)...)
		stream = append(stream, cl2T(32)...)
	}

	data := buildCL2([][]byte{stream})

	b.ResetTimer()

	for range b.N {
		if _, err := assets.DecodeCL2(data, 64, 64); err != nil {
			b.Fatal(err)
		}
	}
}

func cl2ExpectLiteralAndFillRuns(t *testing.T) {
	t.Helper()
	// 4x2: row 0 = [9,9,9,9] fill, row 1 = [1,2] literals + 2 transparent.
	stream := append(cl2F(9, 4), append(cl2L(1, 2), cl2T(2)...)...)
	cl2ExpectDecode(t, [][]byte{stream}, 4, 2, []byte{9, 9, 9, 9, 1, 2, 0, 0}, 4, 2)
}

func cl2ExpectTransparentRunCrossesLine(t *testing.T) {
	t.Helper()
	// 2x2: 3 transparent overruns row 0, then 1 literal on row 1.
	stream := append(cl2T(3), cl2L(4)...)
	cl2ExpectDecode(t, [][]byte{stream}, 2, 2, []byte{0, 0, 0, 4}, 2, 2)
}

func cl2ExpectSkipLine(t *testing.T) {
	t.Helper()
	// 3x2: row 0 literals, row 1 skipped with a full-width transparent run.
	stream := append(cl2L(5, 6, 7), cl2T(3)...)
	cl2ExpectDecode(t, [][]byte{stream}, 3, 2, []byte{5, 6, 7, 0, 0, 0}, 3, 2)
}

func cl2ExpectMultipleFrameOffsets(t *testing.T) {
	t.Helper()
	data := buildCL2([][]byte{cl2L(1, 1), cl2L(2, 2)})

	list, err := assets.DecodeCL2(data, 2, 1)
	if err != nil {
		t.Fatalf("assets.DecodeCL2: %v", err)
	}

	if len(list.Frames) != 2 || list.Frames[0].Pixels[0] != 1 || list.Frames[1].Pixels[0] != 2 {
		t.Fatalf("frames = %+v", list.Frames)
	}
}

func cl2ExpectLiteralRunAtBF(t *testing.T) {
	t.Helper()
	// clx_decode.hpp:28-30: 0xBF is a literal run of 256-0xBF = 65 pixels
	// (clx_encode.hpp:44-49 emits exactly this for long opaque runs).
	stream := append([]byte{0xBF}, ascending65()...)
	cl2ExpectDecode(t, [][]byte{stream}, 65, 1, ascending65(), 65, 1)
}

func cl2ExpectFillRunAtBE(t *testing.T) {
	t.Helper()
	// clx_decode.hpp:31-35: ClxFillMax is 0xBE, width 0xBF-0xBE = 1.
	stream := []byte{0xBE, 7}
	cl2ExpectDecode(t, [][]byte{stream}, 1, 1, []byte{7}, 1, 1)
}

func cl2ExpectRunWidthBoundaries(t *testing.T) {
	t.Helper()
	tests := []struct {
		name   string
		stream []byte
		width  int
		want   []byte
	}{
		{name: "literal width 1", stream: append([]byte{0xFF}, 5), width: 1, want: []byte{5}},
		{name: "literal width 65", stream: append([]byte{0xBF}, ascending65()...), width: 65, want: ascending65()},
		{name: "fill width 1", stream: []byte{0xBE, 5}, width: 1, want: []byte{5}},
		{name: "fill width 63", stream: []byte{0x80, 5}, width: 63, want: filled(63, 5)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl2ExpectDecode(t, [][]byte{tt.stream}, tt.width, 1, tt.want, tt.width, 1)
		})
	}
}

// ascending65 returns the 65 ascending bytes a 0xBF literal run carries.
func ascending65() []byte {
	out := make([]byte, 65)
	for i := range out {
		out[i] = byte(i + 1)
	}

	return out
}

// filled returns n copies of v.
func filled(n int, v byte) []byte {
	out := make([]byte, n)
	for i := range out {
		out[i] = v
	}

	return out
}

func cl2ExpectIncompletePixelsError(t *testing.T) {
	t.Helper()
	data := buildCL2([][]byte{cl2L(1, 1)}) // only 2 of 8 pixels for 4x2
	if _, err := assets.DecodeCL2(data, 4, 2); err == nil {
		t.Fatal("expected error for incomplete pixel stream, got nil")
	}
}

func cl2ExpectDecode(t *testing.T, streams [][]byte, w, h int, want []byte, wantW, wantH int) {
	t.Helper()

	data := buildCL2(streams)

	list, err := assets.DecodeCL2(data, w, h)
	if err != nil {
		t.Fatalf("assets.DecodeCL2: %v", err)
	}

	frame := list.Frames[0]
	if !bytes.Equal(frame.Pixels, want) {
		t.Fatalf("pixels = %v, want %v", frame.Pixels, want)
	}

	if frame.Width != wantW || frame.Height != wantH {
		t.Fatalf("dims = %dx%d, want %dx%d", frame.Width, frame.Height, wantW, wantH)
	}
}

func buildCL2(frames [][]byte) []byte {
	numFrames := len(frames)

	size := 4 + 4*(numFrames+1)
	for _, f := range frames {
		size += 2 + len(f)
	}

	buf := make([]byte, 0, size)
	off := uint32(4 + 4*(numFrames+1))
	buf = binary.LittleEndian.AppendUint32(buf, uint32(numFrames))

	buf = binary.LittleEndian.AppendUint32(buf, off)
	for _, f := range frames {
		off += 2 + uint32(len(f))
		buf = binary.LittleEndian.AppendUint32(buf, off)
	}

	for _, f := range frames {
		buf = binary.LittleEndian.AppendUint16(buf, 2) // header size
		buf = append(buf, f...)
	}

	return buf
}

// cl2ExpectSheetDecodes pins the multi-group branch of DecodeCL2. Before the
// sheet support, cl2.go's parseFrameOffsets rejected this buffer with
// "CL2: last frame offset %d != data size %d (sheets not supported)" because a
// group directory does not end at the buffer size (clx_sprite.hpp:447-457,
// cl2_to_clx.cpp:23-28).
func cl2ExpectSheetDecodes(t *testing.T) {
	t.Helper()

	// given a two-group CL2 buffer: one frame, then two frames
	data := buildCL2Sheet([][][]byte{
		{cl2L(1, 1)},
		{cl2L(2, 2), cl2L(3, 3)},
	})

	// when the sheet is decoded
	sheet, err := assets.DecodeCL2(data, 2, 1)
	if err != nil {
		t.Fatalf("assets.DecodeCL2: %v", err)
	}

	// then both groups decode with their own frame counts and pixels
	if len(sheet.GroupSizes) != 2 || sheet.GroupSizes[0] != 1 || sheet.GroupSizes[1] != 2 {
		t.Fatalf("group sizes = %v, want [1 2]", sheet.GroupSizes)
	}

	want := [][]byte{{1, 1}, {2, 2}, {3, 3}}
	if len(sheet.Frames) != len(want) {
		t.Fatalf("frames = %d, want %d", len(sheet.Frames), len(want))
	}

	for i := range want {
		if !bytes.Equal(sheet.Frames[i].Pixels, want[i]) {
			t.Fatalf("frame %d pixels = %v, want %v", i, sheet.Frames[i].Pixels, want[i])
		}
	}
}

// buildCL2Sheet lays out a group-offset table followed by one CL2 list per group.
func buildCL2Sheet(groups [][][]byte) []byte {
	offsets := make([]uint32, len(groups))
	body := make([]byte, 0, 64)

	for i, frames := range groups {
		offsets[i] = uint32(4*len(groups) + len(body))
		body = append(body, buildCL2(frames)...)
	}

	out := make([]byte, 0, 4*len(groups)+len(body))
	for _, off := range offsets {
		out = binary.LittleEndian.AppendUint32(out, off)
	}

	return append(out, body...)
}
