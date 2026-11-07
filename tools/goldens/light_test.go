package main

import (
	"encoding/binary"
	"testing"
)

func TestParseLightArtifactReadsShades(t *testing.T) {
	// given: a 2x1 dump with one bright and one dark cell.
	data := lightArtifactBytes(2, 1, []byte{0, 15})

	// when
	grid, err := parseLightArtifact(data)
	// then
	if err != nil {
		t.Fatalf("parseLightArtifact: %v", err)
	}
	if grid.Width != 2 || grid.Height != 1 {
		t.Fatalf("dimensions %dx%d, want 2x1", grid.Width, grid.Height)
	}
	expectLightShades(t, grid.DLight, []uint8{0, 15})
}

func TestParseLightArtifactRejectsMalformedDumps(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"truncated header", []byte{2, 0}},
		{"cell count below dimensions", lightArtifactBytes(2, 2, []byte{1, 2})},
		{"zero dimension", lightArtifactBytes(0, 5, nil)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// when
			_, err := parseLightArtifact(tc.data)

			// then
			if err == nil {
				t.Fatalf("parseLightArtifact(%v) = nil error, want one", tc.data)
			}
		})
	}
}

// lightArtifactBytes builds one probe dump: two LE uint16 dimensions then the
// dLight shade cells (tools/probe/dvg_probe.cpp, DumpLightGrids).
func lightArtifactBytes(width, height uint16, cells []byte) []byte {
	data := make([]byte, 4, 4+len(cells))
	binary.LittleEndian.PutUint16(data[0:2], width)
	binary.LittleEndian.PutUint16(data[2:4], height)

	return append(data, cells...)
}

// expectLightShades compares the parsed shades cell by cell.
func expectLightShades(t *testing.T, got, want []uint8) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("%d cells, want %d", len(got), len(want))
	}
	for i, shade := range want {
		if got[i] != shade {
			t.Fatalf("cell %d = %d, want %d", i, got[i], shade)
		}
	}
}
