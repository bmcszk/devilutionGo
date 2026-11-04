package assets_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/assets"
)

func TestConvertPCXToCLXPreservesPixelsAndGeometry(t *testing.T) {
	// given
	// A committed 4x3 PCX whose file rows are constant 0x10, 0x20, 0x30
	// (internal/testdata/gfx/pcx/four_by_three.pcx.hex).
	data := readHexFixture(t, filepath.Join(gfxFixtureDir, "pcx/four_by_three.pcx.hex"))
	// when
	converted, err := assets.ConvertPCXToCLX(data, 1)
	// then
	if err != nil {
		t.Fatalf("ConvertPCXToCLX: %v", err)
	}

	list, err := assets.DecodeCLX(converted)
	if err != nil {
		t.Fatalf("DecodeCLX: %v", err)
	}

	want := []byte{0x30, 0x30, 0x30, 0x30, 0x20, 0x20, 0x20, 0x20, 0x10, 0x10, 0x10, 0x10}
	if len(list.Frames) != 1 {
		t.Fatalf("converted list has %d frames, want 1", len(list.Frames))
	}

	frame := list.Frames[0]
	if frame.Width != 4 || frame.Height != 3 {
		t.Fatalf("frame size = %dx%d, want 4x3", frame.Width, frame.Height)
	}
	if string(frame.Pixels) != string(want) {
		t.Fatalf("pixels = %v, want %v (bottom-up)", frame.Pixels, want)
	}
}

func TestConvertPCXToCLXFramesByCountAndByHeight(t *testing.T) {
	// Two 4x2 frames: file rows 0x10, 0x20 (frame 0) then 0x30, 0x40 (frame 1).
	rows := [][]byte{
		{0x10, 0x10, 0x10, 0x10},
		{0x20, 0x20, 0x20, 0x20},
		{0x30, 0x30, 0x30, 0x30},
		{0x40, 0x40, 0x40, 0x40},
	}
	data := buildPaddedPCX(t, 4, 4, 4, rows)

	tests := []struct {
		name    string
		frames  int
		wantLen int
	}{
		{name: "positive value is the frame count", frames: 2, wantLen: 2},
		{name: "negative value is the frame height", frames: -2, wantLen: 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// when
			converted, err := assets.ConvertPCXToCLX(data, test.frames)
			// then
			if err != nil {
				t.Fatalf("ConvertPCXToCLX: %v", err)
			}

			assertPCXFrameList(t, converted, test.wantLen)
		})
	}
}

func TestConvertPCXToCLXTransparentColourBecomesTransparentRun(t *testing.T) {
	// The transparent palette index is dropped from the stream and becomes a
	// CLX transparent run, which decodes as index 0 (pcx_to_clx.cpp:169-190).
	rows := [][]byte{{0x0F, 1, 0x0F, 0}, {2, 2, 0x0F, 0}}
	data := buildPaddedPCX(t, 3, 2, 4, rows)
	// when
	converted, err := assets.ConvertPCXToCLX(data, 1, 0x0F)
	// then
	if err != nil {
		t.Fatalf("ConvertPCXToCLX: %v", err)
	}

	list, err := assets.DecodeCLX(converted)
	if err != nil {
		t.Fatalf("DecodeCLX: %v", err)
	}

	want, err := assets.DecodePCX(data, 0x0F)
	if err != nil {
		t.Fatalf("DecodePCX: %v", err)
	}
	if string(list.Frames[0].Pixels) != string(want.Pixels) {
		t.Fatalf("pixels = %v, want %v", list.Frames[0].Pixels, want.Pixels)
	}
}

func TestConvertPCXToCLXRejectsBadInput(t *testing.T) {
	// A valid 4x2 PCX whose header will be cut short in one case.
	data := buildPaddedPCX(t, 4, 2, 4, [][]byte{{1, 1, 1, 1}, {2, 2, 2, 2}})

	tests := []struct {
		name    string
		input   []byte
		frames  int
		wantErr string
	}{
		{name: "header truncated", input: data[:64], frames: 1, wantErr: "header"},
		{name: "zero frame geometry", input: data, frames: 0, wantErr: "numFramesOrFrameHeight"},
		{name: "frame count does not divide height", input: data, frames: 3, wantErr: "not a multiple"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// when
			_, err := assets.ConvertPCXToCLX(test.input, test.frames)
			// then
			assertConvertError(t, err, test.wantErr)
		})
	}
}

// assertPCXFrameList checks a converted two-frame 4x2 list and its bottom-up
// row order.
func assertPCXFrameList(t *testing.T, converted []byte, wantLen int) {
	t.Helper()

	list, err := assets.DecodeCLX(converted)
	if err != nil {
		t.Fatalf("DecodeCLX: %v", err)
	}
	if len(list.Frames) != wantLen {
		t.Fatalf("frames = %d, want %d", len(list.Frames), wantLen)
	}
	for i, frame := range list.Frames {
		if frame.Width != 4 || frame.Height != 2 {
			t.Fatalf("frame %d size = %dx%d, want 4x2", i, frame.Width, frame.Height)
		}
	}

	want := []byte{0x20, 0x20, 0x20, 0x20, 0x10, 0x10, 0x10, 0x10}
	if got := list.Frames[0].Pixels; string(got) != string(want) {
		t.Fatalf("frame 0 pixels = %v, want %v", got, want)
	}
}

// assertConvertError asserts that conversion failed with the expected reason.
func assertConvertError(t *testing.T, err error, wantErr string) {
	t.Helper()

	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), wantErr) {
		t.Fatalf("error %q does not contain %q", err, wantErr)
	}
}
