package testsupport_test

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/testsupport"
)

// TestDunTilesCSVRoundTrip pins the tile-layer fixture codec: decoding a
// committed fixture and re-encoding it must reproduce the same bytes, so a
// corpus refresh through EncodeDunTilesCSV cannot drift from what the loaders
// read.
func TestDunTilesCSVRoundTrip(t *testing.T) {
	for _, path := range townInputFixtures(t) {
		t.Run(filepath.Base(path), func(t *testing.T) {
			// given: the committed fixture of this fixed `.dun` file.
			committed := readDunTilesFixture(t, path)

			// when: the fixture is decoded to a `.dun` prefix and re-encoded.
			got := encodeRoundTrip(t, path, committed)

			// then: the re-encoded bytes match the committed fixture exactly.
			if got != string(committed) {
				t.Fatalf("%s: re-encoded %d bytes, want the committed %d", path, len(got), len(committed))
			}
		})
	}
}

// TestDunTilesCSVRejectsMalformed pins the decoder's shape checks.
func TestDunTilesCSVRejectsMalformed(t *testing.T) {
	cases := map[string]string{
		"one-record":  "2,2\n",
		"bad-header":  "2\n1,2,3,4\n",
		"short-layer": "2,2\n1,2,3\n",
		"bad-tile":    "1,1\nnotanumber\n",
	}

	for name, fixture := range cases {
		t.Run(name, func(t *testing.T) {
			// when: the malformed fixture is decoded.
			_, err := testsupport.DecodeDunTilesCSV([]byte(fixture))

			// then: decoding reports the defect instead of returning short bytes.
			if err == nil {
				t.Fatalf("%q decoded without error", fixture)
			}
		})
	}
}

// townInputFixtures lists the committed town `.dun` fixtures, failing the test
// when the corpus is empty.
func townInputFixtures(t *testing.T) []string {
	t.Helper()

	paths, err := filepath.Glob(filepath.Join("..", "..", "..", "internal", "testdata", "drlg", "town", "inputs", "*.csv"))
	if err != nil {
		t.Fatalf("glob town inputs: %v", err)
	}
	if len(paths) == 0 {
		t.Fatal("no committed town input fixtures")
	}

	return paths
}

// readDunTilesFixture reads one committed tile-layer fixture.
func readDunTilesFixture(t *testing.T, path string) []byte {
	t.Helper()

	committed, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	return committed
}

// encodeRoundTrip decodes one fixture into `.dun` bytes and re-encodes it.
func encodeRoundTrip(t *testing.T, path string, committed []byte) string {
	t.Helper()

	dun, err := testsupport.DecodeDunTilesCSV(committed)
	if err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}

	width := int(binary.LittleEndian.Uint16(dun[0:]))
	height := int(binary.LittleEndian.Uint16(dun[2:]))

	tiles := make([]uint16, width*height)
	for i := range tiles {
		tiles[i] = binary.LittleEndian.Uint16(dun[4+2*i:])
	}

	return string(testsupport.EncodeDunTilesCSV(width, height, tiles))
}
