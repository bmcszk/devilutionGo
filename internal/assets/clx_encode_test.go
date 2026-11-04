package assets_test

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/assets"
)

// surfaceFixture is one SurfaceToClx case: a row-major pixel surface, the
// frame geometry, an optional transparent palette index, and the C++ golden.
type surfaceFixture struct {
	name        string
	pixels      []byte
	width       int
	numFrames   int
	transparent []byte
	golden      string
}

func TestEncodeCLXSpriteListMatchesCXXGolden(t *testing.T) {
	fixtures := surfaceFixtures(t)
	if len(fixtures) == 0 {
		t.Fatal("no surface fixtures under internal/testdata/gfx/surface")
	}

	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			// given a stacked pixel surface and the devilutionX SurfaceToClx output
			want := readHexFixture(t, fixture.golden)

			// when
			got, err := assets.EncodeCLXSpriteList(fixture.pixels, fixture.width,
				fixture.numFrames, fixture.transparent...)
			// then
			if err != nil {
				t.Fatalf("EncodeCLXSpriteList: %v", err)
			}
			assertBytesEqual(t, fixture.name, got, want)
		})
	}
}

func TestEncodeCLXSpriteListRejectsBadGeometry(t *testing.T) {
	tests := []struct {
		name      string
		pixels    []byte
		width     int
		numFrames int
	}{
		{name: "zero width", pixels: []byte{0}, width: 0, numFrames: 1},
		{name: "zero frames", pixels: []byte{0, 0}, width: 2, numFrames: 0},
		{name: "ragged rows", pixels: []byte{0, 0, 0}, width: 2, numFrames: 1},
		{name: "height not divisible", pixels: []byte{0, 0, 0, 0, 0, 0}, width: 2, numFrames: 4},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// when
			_, err := assets.EncodeCLXSpriteList(test.pixels, test.width, test.numFrames)
			// then
			if err == nil {
				t.Fatal("expected an error for invalid geometry")
			}
		})
	}
}

// surfaceFixtures loads every committed surface case. The producer is
// internal/testdata/gfx/producer/make_fixtures.py; a missing golden fails.
func surfaceFixtures(tb testing.TB) []surfaceFixture {
	tb.Helper()

	inputs, err := filepath.Glob(filepath.Join(gfxFixtureDir, "surface", "*.surface.hex"))
	if err != nil {
		tb.Fatalf("glob surface fixtures: %v", err)
	}

	fixtures := make([]surfaceFixture, 0, len(inputs))
	for _, input := range inputs {
		base := strings.TrimSuffix(input, ".surface.hex")
		name := filepath.Base(base)
		width, numFrames, transparent := readSurfaceArgs(tb, base+".surface.args")

		fixtures = append(fixtures, surfaceFixture{
			name:        name,
			pixels:      readHexFixture(tb, input),
			width:       width,
			numFrames:   numFrames,
			transparent: transparent,
			golden:      base + ".surface.clx.hex",
		})
	}

	return fixtures
}

// readSurfaceArgs parses "width,numFrames,transparentColor|-".
func readSurfaceArgs(tb testing.TB, path string) (width, numFrames int, transparent []byte) {
	tb.Helper()

	dump, err := os.ReadFile(path)
	if err != nil {
		tb.Fatalf("read fixture %s: %v", path, err)
	}
	fields := strings.Split(strings.TrimSpace(string(dump)), ",")
	if len(fields) != 3 {
		tb.Fatalf("%s: want width,numFrames,transparent, got %q", path, string(dump))
	}

	width, err = strconv.Atoi(fields[0])
	if err != nil {
		tb.Fatalf("parse width %s: %v", path, err)
	}
	numFrames, err = strconv.Atoi(fields[1])
	if err != nil {
		tb.Fatalf("parse numFrames %s: %v", path, err)
	}
	if fields[2] == "-" {
		return width, numFrames, nil
	}
	color, convErr := strconv.Atoi(fields[2])
	if convErr != nil {
		tb.Fatalf("parse transparent %s: %v", path, convErr)
	}

	return width, numFrames, []byte{byte(color)}
}
