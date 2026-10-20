// Package testsupport: golden-test foundations shared by parity tests
// (devilGo-xcdz).
package testsupport

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// dunDir resolves the committed fixture corpus relative to this package
// (go test runs with the package dir as cwd).
var dunDir = filepath.Join("..", "..", "test", "fixtures", "generated", "dun")

// DunMeta is the 5-byte .meta sidecar of one dun fixture pair (writer:
// tools/probe/dvg_probe.cpp DumpSeededDungeon — two native-order u16
// ViewPosition values followed by one dungeon_type level byte).
type DunMeta struct {
	// ViewX/ViewY: ViewPosition.x/.y at dump time (u16 LE on x86).
	ViewX, ViewY int
	// LevelType: the dungeon_type ordinal at dump time.
	LevelType int
}

// DunPair is one parsed dun fixture: the logical tile layer, the
// dPiece-scale transparency window and the meta sidecar.
type DunPair struct {
	// Width/Height are the header dims of the tile layer (40x40 for the
	// standard dungeons; upstream .dun files may exceed the 40x40 logical
	// play area — gendung_defs.hpp DMAXX/DMAXY vs MAXDUNX/MAXDUNY — so the
	// parser trusts the header instead of assuming 40).
	Width, Height int
	// TileLayer[y][x]: the producer dumps the raw memory of the upstream
	// `uint8_t dungeon[DMAXX][DMAXY]` (dun_tile_data.hpp:66), so the dump
	// is x-major (offset k ↔ dungeon[k/Height][k%Height]) — transposed
	// from the game-asset .dun convention (gendung.cpp PlaceDunTiles reads
	// y-major rows). The parser mirrors the writer, exposing [y][x].
	TileLayer [][]uint8
	// TransparencyLayer[y][x]: the dTransVal window rows y∈[16,96),
	// columns x∈[16,96) (dun_tile_data.hpp:103 declares [MAXDUNX][MAXDUNY];
	// the writer slices &dTransVal[x][16] for x∈[16,96), so the dump is
	// x-major rows of the y window — again mirrored to [y][x]).
	TransparencyLayer [][]uint8
	// Meta is the parsed .meta sidecar (same base name, .meta suffix).
	Meta DunMeta
}

// DunFixture is all dun pairs of one seed, keyed by the fixture base name
// ("l1-main", "l17-prev", ...).
type DunFixture struct {
	// Seed as parsed from the filenames.
	Seed uint32
	// Pairs: base name -> parsed layers.
	Pairs map[string]*DunPair
}

// LoadDunFixture parses every committed dun pair of one seed. The testing
// API variant: any parse failure fails the test (golden tests never tolerate
// a broken fixture). One API only — no error-returning twin; callers needing
// error handling parse raw bytes via parseDunPair's inputs instead.
func LoadDunFixture(t *testing.T, seed uint32) *DunFixture {
	t.Helper()

	paths, err := filepath.Glob(filepath.Join(dunDir, "*.dun"))
	if err != nil {
		t.Fatalf("glob dun fixtures: %v", err)
	}
	if len(paths) == 0 {
		t.Skipf("no committed dun fixtures under %s", dunDir)
	}

	fixture := &DunFixture{Seed: seed, Pairs: map[string]*DunPair{}}
	for _, path := range paths {
		parsedSeed, ok := seedFromFilename(path)
		if !ok || parsedSeed != seed {
			continue
		}
		named := strings.TrimSuffix(filepath.Base(path), ".dun") // with seed tail
		base := named[:strings.LastIndex(named, "-")]            // map key
		pair, err := loadDunPair(named)
		if err != nil {
			t.Fatalf("%s: %v", base, err)
		}
		fixture.Pairs[base] = pair
	}
	return fixture
}

// seedFromFilename extracts the trailing -<seed> token; ok=false when the
// name carries no numeric tail.
func seedFromFilename(path string) (uint32, bool) {
	base := strings.TrimSuffix(filepath.Base(path), ".dun")
	dash := strings.LastIndex(base, "-")
	if dash == -1 {
		return 0, false
	}
	var seed uint32
	if _, err := fmt.Sscanf(base[dash+1:], "%d", &seed); err != nil {
		return 0, false
	}
	return seed, true
}

// loadDunPair parses one (named.dun, named.meta) pair off the committed tree.
func loadDunPair(named string) (*DunPair, error) {
	rawDun, err := os.ReadFile(filepath.Join(dunDir, named+".dun"))
	if err != nil {
		return nil, fmt.Errorf("read dun: %w", err)
	}
	layers, err := parseDun(rawDun)
	if err != nil {
		return nil, err
	}

	meta := DunMeta{}
	rawMeta, err := os.ReadFile(filepath.Join(dunDir, named+".meta"))
	if err != nil {
		return nil, fmt.Errorf("read meta: %w", err)
	}
	if len(rawMeta) != 5 {
		return nil, fmt.Errorf("meta size %d, want 5", len(rawMeta))
	}
	meta.ViewX = int(binary.LittleEndian.Uint16(rawMeta[0:]))
	meta.ViewY = int(binary.LittleEndian.Uint16(rawMeta[2:]))
	meta.LevelType = int(rawMeta[4])

	return &DunPair{
		Width:             layers.width,
		Height:            layers.height,
		TileLayer:         layers.tile,
		TransparencyLayer: layers.transparency,
		Meta:              meta,
	}, nil
}

// parsedDun carries the parser's raw output before the pair fields are set.
type parsedDun struct {
	width, height int
	tile          [][]uint8
	transparency  [][]uint8
}

// parseDun parses the probe dump layout (writer: tools/probe/dvg_probe.cpp
// DumpSeededDungeon): LE u16 width, LE u16 height, then width*height tile
// bytes (x-major), then height*2 rows of width*2 transparency bytes
// (x-major rows).
func parseDun(data []byte) (parsedDun, error) {
	var out parsedDun
	if len(data) < 4 {
		return out, fmt.Errorf("size %d, want >= 4 header", len(data))
	}
	out.width = int(binary.LittleEndian.Uint16(data[0:]))
	out.height = int(binary.LittleEndian.Uint16(data[2:]))
	tileEnd := 4 + out.width*out.height
	if len(data) < tileEnd+out.height*2*out.width*2 {
		return out, fmt.Errorf("size %d, want >= %d for %dx%d layers",
			len(data), tileEnd+out.height*2*out.width*2, out.width, out.height)
	}

	out.tile = mirrorMajor(data[4:tileEnd], out.width, out.height)
	out.transparency = mirrorMajor(data[tileEnd:], out.width*2, out.height*2)
	return out, nil
}

// mirrorMajor lifts an x-major byte dump ([major][minor] with minor
// fastest) into a [y][x] slice (minor axis = x).
func mirrorMajor(dump []byte, width, height int) [][]uint8 {
	layer := make([][]uint8, height)
	for y := range layer {
		layer[y] = make([]uint8, width)
	}
	for major := 0; major < width; major++ {
		for minor := 0; minor < height; minor++ {
			layer[minor][major] = dump[major*height+minor]
		}
	}
	return layer
}

// EncodeDun re-encodes a parsed pair into the exact probe dump layout —
// the inverse of parseDun (round-trip golden gate: encode(parse(f)) == f).
func EncodeDun(pair *DunPair) []byte {
	out := make([]byte, 4, 4+pair.Width*pair.Height+pair.Width*2*pair.Height*2)
	binary.LittleEndian.PutUint16(out[0:], uint16(pair.Width))
	binary.LittleEndian.PutUint16(out[2:], uint16(pair.Height))
	for major := 0; major < pair.Width; major++ {
		for minor := 0; minor < pair.Height; minor++ {
			out = append(out, pair.TileLayer[minor][major])
		}
	}
	for major := 0; major < pair.Width*2; major++ {
		for minor := 0; minor < pair.Height*2; minor++ {
			out = append(out, pair.TransparencyLayer[minor][major])
		}
	}
	return out
}
