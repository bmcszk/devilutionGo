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

	frames, err := assets.DecodeCL2(data, 2, 1)
	if err != nil {
		t.Fatalf("assets.DecodeCL2: %v", err)
	}

	if len(frames) != 2 || frames[0].Pixels[0] != 1 || frames[1].Pixels[0] != 2 {
		t.Fatalf("frames = %+v", frames)
	}
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

	frames, err := assets.DecodeCL2(data, w, h)
	if err != nil {
		t.Fatalf("assets.DecodeCL2: %v", err)
	}

	if !bytes.Equal(frames[0].Pixels, want) {
		t.Fatalf("pixels = %v, want %v", frames[0].Pixels, want)
	}

	if frames[0].Width != wantW || frames[0].Height != wantH {
		t.Fatalf("dims = %dx%d, want %dx%d", frames[0].Width, frames[0].Height, wantW, wantH)
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
