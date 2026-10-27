package assets_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/assets"
)

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
	t.Run("sheet rejected", func(t *testing.T) {
		clxExpectSheetRejected(t)
	})

	t.Run("truncated literal run errors", func(t *testing.T) {
		clxExpectTruncationError(t)
	})
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

// Sheet-like data: group offset in place of frame count.
func clxExpectSheetRejected(t *testing.T) {
	t.Helper()

	data := buildCLX([][]byte{cl2L(1, 1)}, 2, 1)
	binary.LittleEndian.PutUint32(data[0:4], 8)

	if _, err := assets.DecodeCLX(data); err == nil {
		t.Fatal("expected error for sheet-like data, got nil")
	}
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
