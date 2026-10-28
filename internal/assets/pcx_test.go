package assets_test

import (
	"encoding/binary"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/assets"
)

func TestPCXRowsAreBottomUp(t *testing.T) {
	// given
	// pcx_to_clx.cpp:123,169: "the lines are reversed in CEL" — frame row j
	// comes from file row frameHeight-(j+1).
	rows := [][]byte{{0x11, 0x11}, {0x22, 0x22}, {0x33, 0x33}}
	// when
	frame := mustDecodePCX(t, buildPaddedPCX(t, 2, 3, 2, rows))
	// then
	want := []byte{0x33, 0x33, 0x22, 0x22, 0x11, 0x11}
	if string(frame.Pixels) != string(want) {
		t.Fatalf("pixels = %v, want %v (bottom-up)", frame.Pixels, want)
	}
}

func TestPCXHonoursBytesPerLine(t *testing.T) {
	// given
	// pcx_to_clx.cpp:91-93,126,141-163: scanlines are bytesPerLine wide and
	// the padding past width is dropped; PCX pads odd widths to an even count.
	rows := [][]byte{{1, 2, 3, 0xEE}, {4, 5, 6, 0xEE}}
	// when
	frame := mustDecodePCX(t, buildPaddedPCX(t, 3, 2, 4, rows))
	// then
	want := []byte{4, 5, 6, 1, 2, 3}
	if string(frame.Pixels) != string(want) {
		t.Fatalf("pixels = %v, want %v", frame.Pixels, want)
	}
}

func TestPCXRejectsBytesPerLineBelowWidth(t *testing.T) {
	// given
	data := buildPaddedPCX(t, 4, 1, 2, [][]byte{{1, 1}})
	// when
	_, err := assets.DecodePCX(data)
	// then
	if err == nil {
		t.Fatal("expected an error when BytesPerLine is below the image width")
	}

	if !strings.Contains(err.Error(), "BytesPerLine") {
		t.Fatalf("error %q does not name the BytesPerLine header field", err)
	}
}

func TestPCXTransparentColourBecomesTransparentIndex(t *testing.T) {
	// given
	// pcx_to_clx.cpp:94,172-190: the optional transparent colour is emitted as
	// transparent runs, which decode as index 0.
	data := buildPaddedPCX(t, 3, 2, 4, [][]byte{{0x0F, 1, 0x0F, 0}, {2, 2, 0x0F, 0}})
	// when
	frame, err := assets.DecodePCX(data, 0x0F)
	// then
	if err != nil {
		t.Fatalf("DecodePCX: %v", err)
	}

	want := []byte{2, 2, 0, 0, 1, 0}
	if string(frame.Pixels) != string(want) {
		t.Fatalf("pixels = %v, want %v", frame.Pixels, want)
	}
}

func TestPCXCommittedFixture(t *testing.T) {
	// given
	// Committed 4x3 PCX whose file rows are constant 0x10, 0x20, 0x30
	// (internal/testdata/gfx/pcx/README in the gfx README).
	data := readHexFixture(t, filepath.Join(gfxFixtureDir, "pcx/four_by_three.pcx.hex"))
	// when
	frame := mustDecodePCX(t, data)
	// then
	want := []byte{0x30, 0x30, 0x30, 0x30, 0x20, 0x20, 0x20, 0x20, 0x10, 0x10, 0x10, 0x10}
	if frame.Width != 4 || frame.Height != 3 {
		t.Fatalf("size = %dx%d, want 4x3", frame.Width, frame.Height)
	}

	if string(frame.Pixels) != string(want) {
		t.Fatalf("pixels = %v, want %v", frame.Pixels, want)
	}
}

// buildPaddedPCX encodes rows whose scanlines are bytesPerLine bytes wide.
func buildPaddedPCX(t *testing.T, width, height, bytesPerLine int, rows [][]byte) []byte {
	t.Helper()

	pixels := make([]byte, 0, bytesPerLine*height)
	for _, row := range rows {
		if len(row) != bytesPerLine {
			t.Fatalf("row width = %d, want bytesPerLine %d", len(row), bytesPerLine)
		}

		pixels = append(pixels, row...)
	}

	data := make([]byte, 128, 128+len(pixels))
	data[0] = 0x0A
	data[1] = 5
	data[2] = 1
	data[3] = 8
	binary.LittleEndian.PutUint16(data[8:10], uint16(width-1))
	binary.LittleEndian.PutUint16(data[10:12], uint16(height-1))
	data[65] = 1
	binary.LittleEndian.PutUint16(data[66:68], uint16(bytesPerLine))

	return append(data, encodePCXRLE(pixels)...)
}
