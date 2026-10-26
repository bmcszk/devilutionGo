package assets_test

import (
	"encoding/binary"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/assets"
)

func TestDecodeCEL(t *testing.T) {
	t.Run("single frame literal run", func(t *testing.T) {
		// 4x2 frame: row 0 = [1,2,3,4], row 1 = [5,5,5,5]
		// CEL: literal runs only (no fill). So we need 8 literals.
		celExpectFrame(t,
			decodeSingleCEL(t, celL(1, 2, 3, 4, 5, 5, 5, 5), 4, 2),
			[]byte{1, 2, 3, 4, 5, 5, 5, 5}, 4, 2)
	})

	t.Run("transparent run", func(t *testing.T) {
		// 4x1: 2 transparent then 2 literals -> index 0 is transparent
		celExpectFrame(t, decodeSingleCEL(t, append(celT(2), celL(7, 8)...), 4, 1), []byte{0, 0, 7, 8}, 4, 1)
	})

	t.Run("transparent run crosses line boundary", func(t *testing.T) {
		// 2x2: one command of 3 transparent pixels overruns row 0 into row 1;
		// decoder writes a flat buffer so the overrun naturally lands on row 1.
		celExpectFrame(t, decodeSingleCEL(t, append(celT(3), celL(9)...), 2, 2), []byte{0, 0, 0, 9}, 2, 2)
	})

	t.Run("multiple frames decode independently", func(t *testing.T) {
		// The test helper builds a two-frame CEL: first frame = 1, second frame = 2 (each 1x1).
		// Ensure both frames are returned in the correct order.
		celExpectIndependentFrames(t)
	})

	t.Run("long literal run", func(t *testing.T) {
		// Test a long literal run (max 63)
		vals := make([]byte, 63)
		for i := range vals {
			vals[i] = byte(i)
		}
		f := decodeSingleCEL(t, celL(vals...), 63, 1)
		if len(f.Pixels) != 63 {
			t.Fatalf("pixels len = %d, want 63", len(f.Pixels))
		}
		for i := range vals {
			if f.Pixels[i] != vals[i] {
				t.Fatalf("pixel[%d] = %d, want %d", i, f.Pixels[i], vals[i])
			}
		}
	})
}

func TestDecodePCX(t *testing.T) {
	t.Run("valid PCX with palette", func(t *testing.T) {
		// Build a minimal 2x2 PCX with palette
		data := buildPCX(t, 2, 2, []byte{1, 2, 3, 4}, true)
		frame, err := assets.DecodePCX(data)
		if err != nil {
			t.Fatalf("DecodePCX: %v", err)
		}
		if frame.Width != 2 || frame.Height != 2 {
			t.Fatalf("size = %dx%d, want 2x2", frame.Width, frame.Height)
		}
		if len(frame.Pixels) != 4 {
			t.Fatalf("pixels len = %d, want 4", len(frame.Pixels))
		}
		if !frame.HasPalette {
			t.Fatal("expected palette present")
		}
		if len(frame.Palette) != 256 {
			t.Fatalf("palette len = %d, want 256", len(frame.Palette))
		}
	})

	t.Run("valid PCX without palette", func(t *testing.T) {
		data := buildPCX(t, 2, 2, []byte{5, 6, 7, 8}, false)
		frame, err := assets.DecodePCX(data)
		if err != nil {
			t.Fatalf("DecodePCX: %v", err)
		}
		if frame.HasPalette {
			t.Fatal("expected no palette")
		}
	})

	t.Run("invalid manufacturer", func(t *testing.T) {
		data := buildPCX(t, 2, 2, []byte{1, 2, 3, 4}, true)
		data[0] = 0x00 // wrong manufacturer
		_, err := assets.DecodePCX(data)
		if err == nil {
			t.Fatal("expected error for invalid manufacturer")
		}
	})

	t.Run("unsupported encoding", func(t *testing.T) {
		data := buildPCX(t, 2, 2, []byte{1, 2, 3, 4}, true)
		data[2] = 0 // no RLE
		_, err := assets.DecodePCX(data)
		if err == nil {
			t.Fatal("expected error for unsupported encoding")
		}
	})
}

func TestDecodeTRN(t *testing.T) {
	t.Run("valid TRN", func(t *testing.T) {
		data := make([]byte, 256)
		for i := 0; i < 256; i++ {
			data[i] = byte(i)
		}
		list, err := assets.DecodeTRN(data)
		if err != nil {
			t.Fatalf("DecodeTRN: %v", err)
		}
		if len(list.Frames) != 1 {
			t.Fatalf("frames = %d, want 1", len(list.Frames))
		}
		if len(list.Frames[0].Data) != 256 {
			t.Fatalf("data len = %d, want 256", len(list.Frames[0].Data))
		}
	})

	t.Run("too short", func(t *testing.T) {
		_, err := assets.DecodeTRN([]byte{1, 2, 3})
		if err == nil {
			t.Fatal("expected error for too short TRN")
		}
	})
}

func TestPalette(t *testing.T) {
	t.Run("LoadPaletteFromBytes", func(t *testing.T) {
		data := make([]byte, 768)
		for i := 0; i < 256; i++ {
			data[i*3] = byte(i)
			data[i*3+1] = byte(i + 50)
			data[i*3+2] = byte(i + 100)
		}
		p, err := assets.LoadPaletteFromBytes(data)
		if err != nil {
			t.Fatalf("LoadPaletteFromBytes: %v", err)
		}
		if p[0][0] != 0 || p[0][1] != 50 || p[0][2] != 100 {
			t.Fatalf("palette[0] = %v, want [0,50,100]", p[0])
		}
	})

	t.Run("ExpandPalette", func(t *testing.T) {
		var p assets.Palette
		p[0] = [3]uint8{255, 0, 0}
		p[1] = [3]uint8{0, 255, 0}
		p[2] = [3]uint8{0, 0, 255}

		indices := []byte{0, 1, 2, 0}
		rgba := p.ExpandPalette(indices)
		if len(rgba) != 16 {
			t.Fatalf("rgba len = %d, want 16", len(rgba))
		}
		// pixel 0: red
		if rgba[0] != 255 || rgba[1] != 0 || rgba[2] != 0 || rgba[3] != 255 {
			t.Fatalf("pixel 0 = %v, want [255,0,0,255]", rgba[:4])
		}
		// pixel 1: green
		if rgba[4] != 0 || rgba[5] != 255 || rgba[6] != 0 || rgba[7] != 255 {
			t.Fatalf("pixel 1 = %v, want [0,255,0,255]", rgba[4:8])
		}
	})

	t.Run("LoadPaletteFromBytes too short", func(t *testing.T) {
		_, err := assets.LoadPaletteFromBytes([]byte{1, 2, 3})
		if err == nil {
			t.Fatal("expected error for short palette data")
		}
	})
}

// --- Test helpers for CEL ---

func decodeSingleCEL(t *testing.T, stream []byte, w, h int) *assets.CelFrame {
	t.Helper()
	data := buildCEL([][]byte{stream}, w, h)
	frames, err := assets.DecodeCEL(data)
	if err != nil {
		t.Fatalf("DecodeCEL: %v", err)
	}
	if len(frames.Frames) != 1 {
		t.Fatalf("frames = %d, want 1", len(frames.Frames))
	}
	return &frames.Frames[0]
}

func celExpectFrame(t *testing.T, f *assets.CelFrame, want []byte, wantW, wantH int) {
	t.Helper()
	if f.Width != wantW || f.Height != wantH {
		t.Fatalf("size = %dx%d, want %dx%d", f.Width, f.Height, wantW, wantH)
	}
	if len(f.Pixels) != len(want) {
		t.Fatalf("pixels len = %d, want %d", len(f.Pixels), len(want))
	}
	for i := range want {
		if f.Pixels[i] != want[i] {
			t.Fatalf("pixel[%d] = %d, want %d", i, f.Pixels[i], want[i])
		}
	}
}

func celExpectIndependentFrames(t *testing.T) {
	t.Helper()
	data := buildCEL([][]byte{
		celL(1), // 1x1 frame with pixel value 1
		celL(2), // 1x1 frame with pixel value 2
	}, 1, 1)
	frames, err := assets.DecodeCEL(data)
	if err != nil {
		t.Fatalf("DecodeCEL: %v", err)
	}
	if len(frames.Frames) != 2 || frames.Frames[0].Pixels[0] != 1 || frames.Frames[1].Pixels[0] != 2 {
		t.Fatalf("frames = %+v", frames)
	}
}

// CEL encoding: control < 0x80 = literal (control+1 bytes), control >= 0x80 = transparent run (-int8(control))
func celL(vals ...byte) []byte {
	if len(vals) > 63 {
		panic("celL: max literal run 63")
	}
	return append([]byte{byte(len(vals) - 1)}, vals...)
}

func celF(val byte, count int) []byte {
	if count < 1 || count > 65 {
		panic("celF: max fill run 65")
	}
	// Fill run: control = 0xBF - count (0xBE..0x80)
	control := byte(0xBF - count)
	return []byte{control, val}
}

func celT(count int) []byte {
	if count < 1 || count > 128 {
		panic("celT: transparent run count 1..128")
	}
	// Transparent run: control = 0x80..0xFF, width = -int8(control)
	// For count=1, control=0xFF; count=128, control=0x80
	control := byte(256 - count)
	return []byte{control}
}

func buildCEL(streams [][]byte, w, h int) []byte {
	numFrames := len(streams)
	totalSize := 4 + 4*(numFrames+1)
	for _, s := range streams {
		totalSize += 6 + len(s) // 6 byte header + pixel data
	}

	data := make([]byte, totalSize)
	binary.LittleEndian.PutUint32(data[0:4], uint32(numFrames))

	offset := uint32(4 + 4*(numFrames+1))
	for i, s := range streams {
		binary.LittleEndian.PutUint32(data[4+4*i:8+4*i], offset)
		// Frame header: 2 bytes header size (6), 2 bytes width, 2 bytes height
		data[offset] = 6
		data[offset+1] = 0
		binary.LittleEndian.PutUint16(data[offset+2:offset+4], uint16(w))
		binary.LittleEndian.PutUint16(data[offset+4:offset+6], uint16(h))
		copy(data[offset+6:offset+6+uint32(len(s))], s)
		offset += 6 + uint32(len(s))
	}
	binary.LittleEndian.PutUint32(data[4+4*numFrames:8+4*numFrames], offset)
	return data
}

// --- Test helpers for PCX ---

func buildPCX(t *testing.T, w, h int, pixels []byte, withPalette bool) []byte {
	t.Helper()
	// PCX header (128 bytes)
	data := make([]byte, 128)
	data[0] = 0x0A // manufacturer
	data[1] = 5    // version
	data[2] = 1    // encoding RLE
	data[3] = 8    // bits per pixel
	binary.LittleEndian.PutUint16(data[4:6], 0)   // Xmin
	binary.LittleEndian.PutUint16(data[6:8], 0)   // Ymin
	binary.LittleEndian.PutUint16(data[8:10], uint16(w-1)) // Xmax
	binary.LittleEndian.PutUint16(data[10:12], uint16(h-1)) // Ymax
	binary.LittleEndian.PutUint16(data[66:68], uint16(w)) // bytes per line
	data[65] = 1 // NPlanes

	pixelData := encodePCXRLE(pixels)
	data = append(data, pixelData...)

	if withPalette {
		// Add palette marker + 256*3 RGB
		data = append(data, 0x0C)
		palette := make([]byte, 768)
		for i := 0; i < 256; i++ {
			palette[i*3] = byte(i)
			palette[i*3+1] = byte((i + 85) % 256)
			palette[i*3+2] = byte((i + 170) % 256)
		}
		data = append(data, palette...)
	}

	return data
}

func encodePCXRLE(pixels []byte) []byte {
	var out []byte
	for i := 0; i < len(pixels); {
		run := 1
		for run < 63 && i+run < len(pixels) && pixels[i+run] == pixels[i] {
			run++
		}
		if run >= 4 || pixels[i] >= 0xC0 {
			// Emit run
			out = append(out, byte(0xC0|run), pixels[i])
			i += run
		} else {
			// Emit literals
			literalRun := 1
			for literalRun < 63 && i+literalRun < len(pixels) && pixels[i+literalRun] < 0xC0 {
				literalRun++
			}
			for j := 0; j < literalRun; j++ {
				out = append(out, pixels[i+j])
			}
			i += literalRun
		}
	}
	return out
}