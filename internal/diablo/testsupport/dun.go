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

// dunArtifactMetaLen is the fixed size of one `.meta` sidecar: two native-order
// u16 ViewPosition values plus the dungeon_type byte (dvg_probe.cpp
// DumpSeededDungeon).
const dunArtifactMetaLen = 5

// DunDir resolves the committed fixture corpus relative to this package
// (go test runs with the package dir as cwd).
var DunDir = filepath.Join("..", "..", "..", "internal", "testdata", "dun")

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
	Seed  uint32
	Pairs map[string]*DunPair
}

// LoadDunFixture parses every committed dun pair of one seed. Parse failures
// fail the test; there is no error-returning variant.
func LoadDunFixture(t *testing.T, seed uint32) *DunFixture {
	t.Helper()

	paths := globDunFixtures(t)
	fixture := &DunFixture{Seed: seed, Pairs: map[string]*DunPair{}}

	for _, path := range paths {
		parsedSeed, ok := SeedFromFilename(path)
		if !ok || parsedSeed != seed {
			continue
		}

		named := strings.TrimSuffix(filepath.Base(path), ".csv")
		base := named[:strings.LastIndex(named, "-")]
		pair := loadDunPairOrFatal(t, named) // pass full named (with seed)
		fixture.Pairs[base] = pair
	}

	return fixture
}

// SeedFromFilename extracts the trailing -<seed> token; ok is false when the
// name carries no numeric tail.
func SeedFromFilename(path string) (uint32, bool) {
	base := strings.TrimSuffix(filepath.Base(path), ".csv")

	_, after, ok := strings.CutLast(base, "-")
	if !ok {
		return 0, false
	}

	var seed uint32
	if _, err := fmt.Sscanf(after, "%d", &seed); err != nil {
		return 0, false
	}

	return seed, true
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

// ParseDunArtifact rebuilds a DunPair from one dvg_probe `.dun` + `.meta`
// artifact pair, so corpus refreshes regenerate goldens through the same
// encoder the tests read them back with.
func ParseDunArtifact(dun, meta []byte) (*DunPair, error) {
	width, height, err := parseDunArtifactDims(dun)
	if err != nil {
		return nil, err
	}

	tiles, err := parseDunArtifactGrid(dun, 4, width, height)
	if err != nil {
		return nil, fmt.Errorf("parse tile layer: %w", err)
	}

	trans, err := parseDunArtifactGrid(dun, 4+width*height, width*2, height*2)
	if err != nil {
		return nil, fmt.Errorf("parse transparency layer: %w", err)
	}

	pairMeta, err := parseDunArtifactMeta(meta)
	if err != nil {
		return nil, err
	}

	return &DunPair{Width: width, Height: height, TileLayer: tiles, TransparencyLayer: trans, Meta: pairMeta}, nil
}

// parseDunArtifactDims reads the u16 LE header of a `.dun` artifact.
func parseDunArtifactDims(dun []byte) (width, height int, err error) {
	if len(dun) < 4 {
		return 0, 0, fmt.Errorf("dun artifact: %d bytes, want at least a 4-byte header", len(dun))
	}

	width = int(binary.LittleEndian.Uint16(dun[0:2]))
	height = int(binary.LittleEndian.Uint16(dun[2:4]))
	if width <= 0 || height <= 0 {
		return 0, 0, fmt.Errorf("dun artifact: %dx%d header", width, height)
	}

	return width, height, nil
}

// parseDunArtifactGrid extracts one grid from a `.dun` artifact. Both layers
// are dumped x-major over a `[x][y]` C array (`dungeon[MAXDUNX][MAXDUNY]`,
// `dTransVal[MAXDUNX][MAXDUNY]` in Source/levels/dun_tile_data.cpp:52-53), so
// cell (x, y) sits at x*height+y and the parser exposes it as [y][x], the same
// indexing the CSV loader uses.
func parseDunArtifactGrid(dun []byte, offset, width, height int) ([][]uint8, error) {
	if len(dun) < offset+width*height {
		return nil, fmt.Errorf("%d bytes, want %d", len(dun), offset+width*height)
	}

	grid := make([][]uint8, height)
	for row := range height {
		grid[row] = make([]uint8, width)
		for col := range width {
			grid[row][col] = dun[offset+col*height+row]
		}
	}

	return grid, nil
}

// parseDunArtifactMeta reads the fixed-size `.meta` sidecar.
func parseDunArtifactMeta(meta []byte) (DunMeta, error) {
	if len(meta) != dunArtifactMetaLen {
		return DunMeta{}, fmt.Errorf("dun meta: %d bytes, want %d", len(meta), dunArtifactMetaLen)
	}

	return DunMeta{
		ViewX:     int(binary.LittleEndian.Uint16(meta[0:2])),
		ViewY:     int(binary.LittleEndian.Uint16(meta[2:4])),
		LevelType: int(meta[4]),
	}, nil
}

// globDunFixtures finds all committed dun CSV fixtures.
func globDunFixtures(t *testing.T) []string {
	t.Helper()

	paths, err := filepath.Glob(filepath.Join(DunDir, "*.csv"))
	if err != nil {
		t.Fatalf("glob dun fixtures: %v", err)
	}

	if len(paths) == 0 {
		t.Fatalf("no committed dun fixtures under %s", DunDir)
	}

	return paths
}

// loadDunPairOrFatal loads a dun pair or fatals with context.
// Expects full named string (with -seed suffix).
func loadDunPairOrFatal(t *testing.T, named string) *DunPair {
	t.Helper()

	pair, err := loadDunPair(named)
	if err != nil {
		t.Fatalf("%s: %v", named, err)
	}

	return pair
}

// loadDunPair parses one committed CSV fixture: header, tile layer,
// doubled transparency layer, then the meta triple (see encodeDunCSV).
func loadDunPair(named string) (*DunPair, error) {
	file, err := os.Open(filepath.Join(DunDir, named+".csv"))
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

	if err := validateDunCSVShape(records); err != nil {
		return nil, err
	}

	width, height, err := parseDunDimensions(records[0])
	if err != nil {
		return nil, err
	}

	tile, err := csvLayer(records[1], width, height)
	if err != nil {
		return nil, fmt.Errorf("tile layer: %w", err)
	}

	transparency, err := csvLayer(records[2], width*2, height*2)
	if err != nil {
		return nil, fmt.Errorf("transparency layer: %w", err)
	}

	meta, err := parseDunMeta(records[3])
	if err != nil {
		return nil, err
	}

	return &DunPair{
		Width:             width,
		Height:            height,
		TileLayer:         tile,
		TransparencyLayer: transparency,
		Meta:              DunMeta{ViewX: meta[0], ViewY: meta[1], LevelType: meta[2]},
	}, nil
}

// validateDunCSVShape checks the CSV has the expected 4-record shape.
// Mirrors the shape check in loadDunPair (dun.cpp).
func validateDunCSVShape(records [][]string) error {
	if len(records) != 4 || len(records[0]) != 2 || len(records[3]) != 3 {
		return errors.New("CSV records have invalid shape")
	}

	return nil
}

// parseDunDimensions parses the width/height from record 0.
// Mirrors the dimension parsing in loadDunPair (dun.cpp).
func parseDunDimensions(record []string) (width, height int, err error) {
	width, err = strconv.Atoi(record[0])
	if err != nil {
		return 0, 0, fmt.Errorf("width: %w", err)
	}

	height, err = strconv.Atoi(record[1])
	if err != nil {
		return 0, 0, fmt.Errorf("height: %w", err)
	}

	return width, height, nil
}

// parseDunMeta parses the 3-int meta array from record 3.
// Mirrors the meta parsing in loadDunPair (dun.cpp).
func parseDunMeta(record []string) ([3]int, error) {
	var meta [3]int

	for i, value := range record {
		val, err := strconv.Atoi(value)
		if err != nil {
			return meta, fmt.Errorf("meta %d: %w", i, err)
		}

		meta[i] = val
	}

	return meta, nil
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
		column, err := parseDunColumn(values, major, height)
		if err != nil {
			return nil, err
		}

		for minor := range height {
			layer[minor][major] = column[minor]
		}
	}

	return layer, nil
}

// parseDunColumn parses one x-major column of the CSV layer.
// Mirrors the inner loop of csvLayer (dun.cpp).
func parseDunColumn(values []string, major, height int) ([]uint8, error) {
	column := make([]uint8, height)
	for minor := range height {
		value, err := parseDunCell(values, major*height+minor)
		if err != nil {
			return nil, err
		}

		column[minor] = value
	}

	return column, nil
}

// parseDunCell parses one CSV layer cell into a byte.
// Mirrors the cell validation in csvLayer (dun.cpp).
func parseDunCell(values []string, index int) (uint8, error) {
	value, err := strconv.Atoi(values[index])
	if err != nil {
		return 0, fmt.Errorf("cell %d: %w", index, err)
	}

	if value < 0 || value > 255 {
		return 0, fmt.Errorf("cell %d = %d outside byte range", index, value)
	}

	return uint8(value), nil
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
