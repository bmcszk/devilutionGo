package levels_test

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/levels"
)

// drlgDir is the committed DRLG golden corpus (internal/testdata/drlg).
const drlgDir = "../../testdata/drlg"

// committedSeeds are the eight pinned seeds of tools/probe/seeds.txt.
var committedSeeds = []uint32{2588, 743271966, 1383137027, 844660068, 19770182, 1522546307, 125121312, 1511478689}

// dunGolden is one parsed corpus fixture: the tile layer, the doubled
// transparency window and the meta record. Both grids keep the generator's
// indexing — first index x, second index y — because the corpus stores them
// flattened in that order.
type dunGolden struct {
	level, entry string
	tiles        [40][40]uint8
	trans        [80][80]uint8
	view         levels.Point
	levelType    int
}

// loadGoldens parses every <level>-<entry>-<seed>.csv fixture of one seed.
func loadGoldens(t *testing.T, seed uint32) map[string]*dunGolden {
	t.Helper()

	paths, err := filepath.Glob(filepath.Join(drlgDir, fmt.Sprintf("l*-%d.csv", seed)))
	if err != nil {
		t.Fatalf("glob goldens: %v", err)
	}

	goldens := make(map[string]*dunGolden, len(paths))
	for _, path := range paths {
		base := strings.TrimSuffix(filepath.Base(path), ".csv")
		if !strings.HasSuffix(base, fmt.Sprintf("-%d", seed)) {
			continue
		}

		key := strings.TrimSuffix(base, fmt.Sprintf("-%d", seed))
		goldens[key] = loadGolden(t, path)
	}

	return goldens
}

func loadGolden(t *testing.T, path string) *dunGolden {
	t.Helper()

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open golden %s: %v", path, err)
	}
	defer file.Close() //nolint:errcheck // read-only fixture

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("parse golden %s: %v", path, err)
	}

	if len(records) != 4 {
		t.Fatalf("%s: %d records, want 4", path, len(records))
	}

	if len(records[1]) != 40*40 || len(records[2]) != 80*80 {
		t.Fatalf("%s: layer sizes %d/%d, want 1600/6400", path, len(records[1]), len(records[2]))
	}

	golden := &dunGolden{}

	for i, cell := range records[1] {
		golden.tiles[i/40][i%40] = parseCell(t, path, cell)
	}

	for i, cell := range records[2] {
		golden.trans[i/80][i%80] = parseCell(t, path, cell)
	}

	golden.view = levels.Point{X: parseNum(t, path, records[3][0]), Y: parseNum(t, path, records[3][1])}
	golden.levelType = parseNum(t, path, records[3][2])

	base := strings.TrimSuffix(filepath.Base(path), ".csv")
	parts := strings.Split(base, "-")
	golden.level, golden.entry = parts[0], parts[1]

	return golden
}

func parseCell(t *testing.T, path, cell string) uint8 {
	t.Helper()

	value := parseNum(t, path, cell)
	if value < 0 || value > 255 {
		t.Fatalf("%s: cell %d out of range", path, value)
	}

	return uint8(value)
}

func parseNum(t *testing.T, path, cell string) int {
	t.Helper()

	value, err := strconv.Atoi(cell)
	if err != nil {
		t.Fatalf("%s: bad number %q: %v", path, cell, err)
	}

	return value
}
