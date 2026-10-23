// Package testsupport: golden-test foundations shared by parity tests
// (devilGo-xcdz).
package testsupport

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// dunDir resolves the committed fixture corpus relative to this package
// (go test runs with the package dir as cwd).
var dunDir = filepath.Join("..", "..", "..", "internal", "testdata", "dun")

// DunMeta is the 5-byte metadata of one dun fixture pair: two native-order
// u16 ViewPosition values followed by one dungeon_type level byte.
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
	// standard dungeons); the parser trusts the header because .dun files may
	// exceed the 40x40 logical play area.
	Width, Height int
	// TileLayer[y][x]: the dump is x-major, so the parser mirrors the
	// writer's layout and exposes [y][x].
	TileLayer [][]uint8
	// TransparencyLayer[y][x]: the doubled dTransVal window (rows y∈[16,96),
	// columns x∈[16,96)); the dump is x-major, mirrored to [y][x].
	TransparencyLayer [][]uint8
	Meta              DunMeta
}

// DunFixture is all dun pairs of one seed, keyed by the fixture base name
// ("l1-main", "l17-prev", ...).
type DunFixture struct {
	Seed uint32
	// Pairs: base name -> parsed layers.
	Pairs map[string]*DunPair
}

// LoadDunFixture parses every committed dun pair of one seed. Parse failures
// fail the test; there is no error-returning variant.
func LoadDunFixture(t *testing.T, seed uint32) *DunFixture {
	t.Helper()

	paths, err := filepath.Glob(filepath.Join(dunDir, "*.csv"))
	if err != nil {
		t.Fatalf("glob dun fixtures: %v", err)
	}
	if len(paths) == 0 {
		t.Fatalf("no committed dun fixtures under %s", dunDir)
	}

	fixture := &DunFixture{Seed: seed, Pairs: map[string]*DunPair{}}
	for _, path := range paths {
		parsedSeed, ok := seedFromFilename(path)
		if !ok || parsedSeed != seed {
			continue
		}
		named := strings.TrimSuffix(filepath.Base(path), ".csv")
		base := named[:strings.LastIndex(named, "-")]
		pair, err := loadDunPair(named)
		if err != nil {
			t.Fatalf("%s: %v", base, err)
		}
		fixture.Pairs[base] = pair
	}
	return fixture
}

// seedFromFilename extracts the trailing -<seed> token; ok is false when the
// name carries no numeric tail.
func seedFromFilename(path string) (uint32, bool) {
	base := strings.TrimSuffix(filepath.Base(path), ".csv")
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

// loadDunPair parses one committed CSV fixture: header, tile layer,
// doubled transparency layer, then the meta triple (see EncodeDunCSV).
func loadDunPair(named string) (*DunPair, error) {
	file, err := os.Open(filepath.Join(dunDir, named+".csv"))
	if err != nil {
		return nil, fmt.Errorf("read dun csv: %w", err)
	}
	defer func() { _ = file.Close() }()
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse dun csv: %w", err)
	}
	if len(records) != 4 || len(records[0]) != 2 || len(records[3]) != 3 {
		return nil, errors.New("CSV records have invalid shape")
	}
	width, err := strconv.Atoi(records[0][0])
	if err != nil {
		return nil, fmt.Errorf("width: %w", err)
	}
	height, err := strconv.Atoi(records[0][1])
	if err != nil {
		return nil, fmt.Errorf("height: %w", err)
	}
	tile, err := csvLayer(records[1], width, height)
	if err != nil {
		return nil, fmt.Errorf("tile layer: %w", err)
	}
	transparency, err := csvLayer(records[2], width*2, height*2)
	if err != nil {
		return nil, fmt.Errorf("transparency layer: %w", err)
	}
	meta := [3]int{}
	for i, value := range records[3] {
		meta[i], err = strconv.Atoi(value)
		if err != nil {
			return nil, fmt.Errorf("meta %d: %w", i, err)
		}
	}
	return &DunPair{Width: width, Height: height, TileLayer: tile,
		TransparencyLayer: transparency, Meta: DunMeta{ViewX: meta[0], ViewY: meta[1], LevelType: meta[2]}}, nil
}

func csvLayer(values []string, width, height int) ([][]uint8, error) {
	if len(values) != width*height {
		return nil, fmt.Errorf("values %d, want %d", len(values), width*height)
	}
	layer := make([][]uint8, height)
	for y := range layer {
		layer[y] = make([]uint8, width)
	}
	for major := range width {
		for minor := range height {
			index := major*height + minor
			value, err := strconv.Atoi(values[index])
			if err != nil {
				return nil, fmt.Errorf("cell %d: %w", index, err)
			}
			if value < 0 || value > 255 {
				return nil, fmt.Errorf("cell %d = %d outside byte range", index, value)
			}
			layer[minor][major] = uint8(value)
		}
	}
	return layer, nil
}

// EncodeDunCSV re-encodes a parsed pair into the committed CSV fixture
// format (round-trip golden gate: encode(parse(f)) == f).
func EncodeDunCSV(pair *DunPair) []byte {
	var out bytes.Buffer
	writer := csv.NewWriter(&out)
	_ = writer.Write([]string{strconv.Itoa(pair.Width), strconv.Itoa(pair.Height)})
	_ = writer.Write(flattenLayer(pair.TileLayer))
	_ = writer.Write(flattenLayer(pair.TransparencyLayer))
	_ = writer.Write([]string{
		strconv.Itoa(pair.Meta.ViewX), strconv.Itoa(pair.Meta.ViewY), strconv.Itoa(pair.Meta.LevelType),
	})
	writer.Flush()
	return out.Bytes()
}

func flattenLayer(layer [][]uint8) []string {
	width, height := len(layer[0]), len(layer)
	values := make([]string, 0, width*height)
	for major := range width {
		for minor := range height {
			values = append(values, strconv.Itoa(int(layer[minor][major])))
		}
	}
	return values
}

// EncodeDun re-encodes a parsed pair into the exact probe dump layout
// (round-trip golden gate: encode(parse(f)) == f).
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
