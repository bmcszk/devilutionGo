package assets_test

import (
	"encoding/binary"
	"github.com/bmcszk/devilutionGo/internal/assets"
	"testing"
)

func TestDecodeCL2(t *testing.T) {
	t.Run("literal and fill runs", func(t *testing.T) {
		// 4x2: row 0 = [9,9,9,9] fill, row 1 = [1,2] literals + 2 transparent.
		stream := append(cl2F(9, 4), append(cl2L(1, 2), cl2T(2)...)...)
		data := buildCL2([][]byte{stream})
		frames, err := assets.DecodeCL2(data, 4, 2)
		if err != nil {
			t.Fatalf("assets.DecodeCL2: %v", err)
		}
		want := []byte{9, 9, 9, 9, 1, 2, 0, 0}
		if string(frames[0].Pixels) != string(want) {
			t.Fatalf("pixels = %v, want %v", frames[0].Pixels, want)
		}
		if frames[0].Width != 4 || frames[0].Height != 2 {
			t.Fatalf("dims = %dx%d, want 4x2", frames[0].Width, frames[0].Height)
		}
	})

	t.Run("transparent run crosses line boundary", func(t *testing.T) {
		// 2x2: 3 transparent overruns row 0, then 1 literal on row 1.
		stream := append(cl2T(3), cl2L(4)...)
		data := buildCL2([][]byte{stream})
		frames, err := assets.DecodeCL2(data, 2, 2)
		if err != nil {
			t.Fatalf("assets.DecodeCL2: %v", err)
		}
		want := []byte{0, 0, 0, 4}
		if string(frames[0].Pixels) != string(want) {
			t.Fatalf("pixels = %v, want %v", frames[0].Pixels, want)
		}
	})

	t.Run("skip-line", func(t *testing.T) {
		// 3x2: row 0 literals, row 1 skipped with a full-width transparent run.
		stream := append(cl2L(5, 6, 7), cl2T(3)...)
		data := buildCL2([][]byte{stream})
		frames, err := assets.DecodeCL2(data, 3, 2)
		if err != nil {
			t.Fatalf("assets.DecodeCL2: %v", err)
		}
		want := []byte{5, 6, 7, 0, 0, 0}
		if string(frames[0].Pixels) != string(want) {
			t.Fatalf("pixels = %v, want %v", frames[0].Pixels, want)
		}
	})

	t.Run("multiple frames with distinct offsets", func(t *testing.T) {
		data := buildCL2([][]byte{cl2L(1, 1), cl2L(2, 2)})
		frames, err := assets.DecodeCL2(data, 2, 1)
		if err != nil {
			t.Fatalf("assets.DecodeCL2: %v", err)
		}
		if len(frames) != 2 || frames[0].Pixels[0] != 1 || frames[1].Pixels[0] != 2 {
			t.Fatalf("frames = %+v", frames)
		}
	})

	t.Run("incomplete pixels error", func(t *testing.T) {
		data := buildCL2([][]byte{cl2L(1, 1)}) // only 2 of 8 pixels for 4x2
		if _, err := assets.DecodeCL2(data, 4, 2); err == nil {
			t.Fatal("expected error for incomplete pixel stream, got nil")
		}
	})
}

func BenchmarkCL2Decode(b *testing.B) {
	// Synthetic 64x64 frame: 64 rows of 32 literals + 32 transparent.
	var stream []byte
	for row := 0; row < 64; row++ {
		lit := make([]byte, 32)
		for i := range lit {
			lit[i] = byte(i)
		}
		stream = append(stream, cl2L(lit...)...)
		stream = append(stream, cl2T(32)...)
	}
	data := buildCL2([][]byte{stream})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := assets.DecodeCL2(data, 64, 64); err != nil {
			b.Fatal(err)
		}
	}
}

// buildCL2 assembles a minimal single-list CL2 file: frame count, numFrames+1
// offsets (last == file size), then per-frame 2-byte headers + pixel streams.
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
