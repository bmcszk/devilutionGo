package assets_test

import (
	"encoding/binary"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/assets"
)

func TestDecodeCEL(t *testing.T) {
	// when
	subtests := []struct {
		name string
		run  func(*testing.T)
	}{
		{name: "single frame literal run", run: celExpectLiteralRun},
		{name: "transparent run", run: celExpectTransparentRun},
		{name: "transparent run crosses line boundary", run: celExpectTransparentRunCrossesLine},
		{name: "multiple frames decode independently", run: celExpectIndependentFrames},
		{name: "long literal run", run: celExpectLongLiteralRun},
	}
	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

func TestDecodePCX(t *testing.T) {
	// when
	subtests := []struct {
		name string
		run  func(*testing.T)
	}{
		{name: "valid PCX with palette", run: pcxExpectFrameWithPalette},
		{name: "valid PCX without palette", run: pcxExpectFrameWithoutPalette},
		{name: "invalid manufacturer", run: pcxExpectInvalidManufacturerError},
		{name: "unsupported encoding", run: pcxExpectUnsupportedEncodingError},
	}
	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

func TestDecodeTRN(t *testing.T) {
	// when
	subtests := []struct {
		name string
		run  func(*testing.T)
	}{
		{name: "valid TRN", run: trnExpectSingleFrame},
		{name: "too short", run: trnExpectTooShortError},
	}
	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

func TestPalette(t *testing.T) {
	// when
	subtests := []struct {
		name string
		run  func(*testing.T)
	}{
		{name: "LoadPaletteFromBytes", run: paletteExpectLoaded},
		{name: "ExpandPalette", run: paletteExpectExpanded},
		{name: "LoadPaletteFromBytes too short", run: paletteExpectTooShortError},
	}
	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

func celExpectLiteralRun(t *testing.T) {
	t.Helper()
	// 4x2 frame: row 0 = [1,2,3,4], row 1 = [5,5,5,5]
	// CEL: literal runs only (no fill). So we need 8 literals.
	celExpectFrame(t,
		decodeSingleCEL(t, celL(1, 2, 3, 4, 5, 5, 5, 5), 4, 2),
		[]byte{1, 2, 3, 4, 5, 5, 5, 5}, 4, 2)
}

func celExpectTransparentRun(t *testing.T) {
	t.Helper()
	// 4x1: 2 transparent then 2 literals -> index 0 is transparent
	celExpectFrame(t, decodeSingleCEL(t, append(celT(2), celL(7, 8)...), 4, 1), []byte{0, 0, 7, 8}, 4, 1)
}

func celExpectTransparentRunCrossesLine(t *testing.T) {
	t.Helper()
	// 2x2: one command of 3 transparent pixels overruns row 0 into row 1;
	// decoder writes a flat buffer so the overrun naturally lands on row 1.
	celExpectFrame(t, decodeSingleCEL(t, append(celT(3), celL(9)...), 2, 2), []byte{0, 0, 0, 9}, 2, 2)
}

func celExpectLongLiteralRun(t *testing.T) {
	t.Helper()
	// Test a long literal run (max 63)
	vals := make([]byte, 63)
	for i := range vals {
		vals[i] = byte(i)
	}

	f := decodeSingleCEL(t, celL(vals...), 63, 1)
	celExpectPixels(t, f.Pixels, vals)
}

func pcxExpectFrameWithPalette(t *testing.T) {
	t.Helper()
	// Build a minimal 2x2 PCX with palette
	data := buildPCXWithPalette(t, 2, 2, []byte{1, 2, 3, 4})

	frame := mustDecodePCX(t, data)
	pcxExpectSize(t, frame, 2, 2, 4)

	if !frame.HasPalette {
		t.Fatal("expected palette present")
	}

	if len(frame.Palette) != 256 {
		t.Fatalf("palette len = %d, want 256", len(frame.Palette))
	}
}

func pcxExpectFrameWithoutPalette(t *testing.T) {
	t.Helper()
	data := buildPCXWithoutPalette(t, 2, 2, []byte{5, 6, 7, 8})

	frame := mustDecodePCX(t, data)
	if frame.HasPalette {
		t.Fatal("expected no palette")
	}
}

func pcxExpectInvalidManufacturerError(t *testing.T) {
	t.Helper()
	data := buildPCXWithPalette(t, 2, 2, []byte{1, 2, 3, 4})
	data[0] = 0x00 // wrong manufacturer

	pcxExpectDecodeError(t, data, "invalid manufacturer")
}

func pcxExpectUnsupportedEncodingError(t *testing.T) {
	t.Helper()
	data := buildPCXWithPalette(t, 2, 2, []byte{1, 2, 3, 4})
	data[2] = 0 // no RLE

	pcxExpectDecodeError(t, data, "unsupported encoding")
}

func mustDecodePCX(t *testing.T, data []byte) *assets.PcxFrame {
	t.Helper()

	frame, err := assets.DecodePCX(data)
	if err != nil {
		t.Fatalf("DecodePCX: %v", err)
	}

	return frame
}

func pcxExpectDecodeError(t *testing.T, data []byte, reason string) {
	t.Helper()

	if _, err := assets.DecodePCX(data); err == nil {
		t.Fatalf("expected error for %s", reason)
	}
}

func pcxExpectSize(t *testing.T, frame *assets.PcxFrame, wantW, wantH, wantPixels int) {
	t.Helper()

	if frame.Width != wantW || frame.Height != wantH {
		t.Fatalf("size = %dx%d, want %dx%d", frame.Width, frame.Height, wantW, wantH)
	}

	if len(frame.Pixels) != wantPixels {
		t.Fatalf("pixels len = %d, want %d", len(frame.Pixels), wantPixels)
	}
}

func trnExpectSingleFrame(t *testing.T) {
	t.Helper()

	data := make([]byte, 256)
	for i := range 256 {
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
}

func trnExpectTooShortError(t *testing.T) {
	t.Helper()

	if _, err := assets.DecodeTRN([]byte{1, 2, 3}); err == nil {
		t.Fatal("expected error for too short TRN")
	}
}

func paletteExpectLoaded(t *testing.T) {
	t.Helper()

	data := make([]byte, 768)
	for i := range 256 {
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
}

func paletteExpectExpanded(t *testing.T) {
	t.Helper()

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
}

func paletteExpectTooShortError(t *testing.T) {
	t.Helper()

	if _, err := assets.LoadPaletteFromBytes([]byte{1, 2, 3}); err == nil {
		t.Fatal("expected error for short palette data")
	}
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

	celExpectPixels(t, f.Pixels, want)
}

func celExpectPixels(t *testing.T, got, want []byte) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("pixels len = %d, want %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("pixel[%d] = %d, want %d", i, got[i], want[i])
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

func buildPCXWithPalette(t *testing.T, w, h int, pixels []byte) []byte {
	t.Helper()

	return appendPCXPalette(buildPCXWithoutPalette(t, w, h, pixels))
}

func buildPCXWithoutPalette(t *testing.T, w, h int, pixels []byte) []byte {
	t.Helper()
	encoded := encodePCXRLE(pixels)

	// PCX header (128 bytes)
	data := make([]byte, 128, 128+len(encoded))
	data[0] = 0x0A                                          // manufacturer
	data[1] = 5                                             // version
	data[2] = 1                                             // encoding RLE
	data[3] = 8                                             // bits per pixel
	binary.LittleEndian.PutUint16(data[4:6], 0)             // Xmin
	binary.LittleEndian.PutUint16(data[6:8], 0)             // Ymin
	binary.LittleEndian.PutUint16(data[8:10], uint16(w-1))  // Xmax
	binary.LittleEndian.PutUint16(data[10:12], uint16(h-1)) // Ymax
	binary.LittleEndian.PutUint16(data[66:68], uint16(w))   // bytes per line
	data[65] = 1                                            // NPlanes

	return append(data, encoded...)
}

func appendPCXPalette(data []byte) []byte {
	// Add palette marker + 256*3 RGB
	data = append(data, 0x0C)

	palette := make([]byte, 768)
	for i := range 256 {
		palette[i*3] = byte(i)
		palette[i*3+1] = byte((i + 85) % 256)
		palette[i*3+2] = byte((i + 170) % 256)
	}

	return append(data, palette...)
}

func encodePCXRLE(pixels []byte) []byte {
	var out []byte

	for i := 0; i < len(pixels); {
		run := pcxRunLength(pixels, i)
		if run >= 4 || pixels[i] >= 0xC0 {
			// Emit run
			out = append(out, byte(0xC0|run), pixels[i])
			i += run

			continue
		}

		// Emit literals
		literalRun := pcxLiteralRunLength(pixels, i)
		out = append(out, pixels[i:i+literalRun]...)
		i += literalRun
	}

	return out
}

func pcxRunLength(pixels []byte, i int) int {
	run := 1
	for run < 63 && i+run < len(pixels) && pixels[i+run] == pixels[i] {
		run++
	}

	return run
}

func pcxLiteralRunLength(pixels []byte, i int) int {
	literalRun := 1
	for literalRun < 63 && i+literalRun < len(pixels) && pixels[i+literalRun] < 0xC0 {
		literalRun++
	}

	return literalRun
}
