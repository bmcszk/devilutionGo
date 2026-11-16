package assets_test

import (
	"encoding/binary"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/assets"
)

// cl2ExpectGroupHeaderRelativeRanges pins the sheet range rule: a group's
// frame offsets are relative to its OWN header offset and the group's last
// pointer is that group's frame-data end (cl2_to_clx.cpp GetGroupRanges).
// plrgfx sheets lay all group headers first, so group i's data extends past
// group i+1's header — slicing group i to group i+1's header truncates it.
func cl2ExpectGroupHeaderRelativeRanges(t *testing.T) {
	t.Helper()

	// given a two-group sheet laid out headers-first: directory, header0,
	// header1, then group 0's frame data, then group 1's
	const (
		dirSize = 8
		header0 = dirSize      // one frame: u32 count + 2 offsets
		header1 = header0 + 16 // two frames: u32 count + 3 offsets
		g0Data  = header1 + 16 // group 0 data starts past BOTH headers
	)
	g0DataSz := 2 + 3 // u16 header + one literal run (control byte + 2 pixels)
	g1DataSz := 4 + 3 + 3
	g1Data := g0Data + g0DataSz

	data := make([]byte, g1Data+g1DataSz)
	binary.LittleEndian.PutUint32(data[0:4], header0)
	binary.LittleEndian.PutUint32(data[4:8], header1)

	// header0: 1 frame, offsets relative to header0
	binary.LittleEndian.PutUint32(data[header0:header0+4], 1)
	binary.LittleEndian.PutUint32(data[header0+4:header0+8], uint32(g0Data-header0))
	binary.LittleEndian.PutUint32(data[header0+8:header0+12], uint32(g0Data+g0DataSz-header0))

	// header1: 2 frames, offsets relative to header1
	binary.LittleEndian.PutUint32(data[header1:header1+4], 2)
	binary.LittleEndian.PutUint32(data[header1+4:header1+8], uint32(g1Data-header1))
	binary.LittleEndian.PutUint32(data[header1+8:header1+12], uint32(g1Data+2+3-header1))
	binary.LittleEndian.PutUint32(data[header1+12:header1+16], uint32(g1Data+g1DataSz-header1))

	binary.LittleEndian.PutUint16(data[g0Data:g0Data+2], 2)
	copy(data[g0Data+2:], cl2L(1, 1))
	binary.LittleEndian.PutUint16(data[g1Data:g1Data+2], 2)
	copy(data[g1Data+2:], cl2L(2, 2))
	binary.LittleEndian.PutUint16(data[g1Data+2+3:g1Data+2+5], 2)
	copy(data[g1Data+2+5:], cl2L(3, 3))

	// when the sheet decodes
	sheet, err := assets.DecodeCL2(data, 2, 1)
	if err != nil {
		t.Fatalf("assets.DecodeCL2: %v", err)
	}

	// then both groups carry their own frames, decoded from the right bytes
	if len(sheet.GroupSizes) != 2 || sheet.GroupSizes[0] != 1 || sheet.GroupSizes[1] != 2 {
		t.Fatalf("group sizes = %v, want [1 2]", sheet.GroupSizes)
	}

	want := [][]byte{{1, 1}, {2, 2}, {3, 3}}
	for i, frame := range want {
		if got := sheet.Frames[i].Pixels; string(got) != string(frame) {
			t.Fatalf("frame %d pixels = %v, want %v", i, got, frame)
		}
	}
}
