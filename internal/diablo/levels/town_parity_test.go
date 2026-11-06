package levels_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/levels"
	"github.com/bmcszk/devilutionGo/internal/diablo/testsupport"
)

// townDir is the committed CreateTown corpus (internal/testdata/drlg/README.md,
// "Town").
const townDir = drlgDir + "/town"

// townInputs are the five `.dun` layers CreateTown loads and the fixture name
// each is expected under: townDir/inputs/<fixture>.csv in the setpieces format,
// `width,height` then width*height uint16 tile values.
var townInputs = []struct {
	resource string
	fixture  string
}{
	{levels.TownSector1, "sector1s"},
	{levels.TownSector2, "sector2s"},
	{levels.TownSector3, "sector3s"},
	{levels.TownSector4, "sector4s"},
	{levels.TownAutomap, "automap"},
}

// townGolden is one parsed town fixture: the layers CreateTown writes (dPiece,
// dSpecial), the automap tile layer it patches (`dungeon`), and the entry/seed
// the file name encodes.
type townGolden struct {
	seed    uint32
	entry   levels.Entry
	piece   [levels.MAXDUNX][levels.MAXDUNY]uint16
	special [levels.MAXDUNX][levels.MAXDUNY]int8
	tiles   [levels.DMAXX][levels.DMAXY]levels.Tile
}

// TestTownParity is the CreateTown acceptance gate: every committed town golden
// must be reproduced cell for cell, in both the world-tile grids the pass builds
// (dPiece, dSpecial) and the automap tile layer it patches.
func TestTownParity(t *testing.T) {
	// given: the level's mega-tile table, its fixed .dun inputs and the corpus.
	megaTiles := loadMegaTiles(t, "town")
	files, missing := townDunFiles(t)

	// when: every committed seed drives the town generator (levels/town.cpp:359).
	for _, golden := range townGoldens(t) {
		t.Run(fmt.Sprintf("seed-%d", golden.seed), func(t *testing.T) {
			// given: a fresh town dungeon whose RNG is pinned to the fixture seed.
			dungeon := levels.NewDungeon(levels.Town, 0)
			dungeon.Rnd.Seed(golden.seed)

			// when: the town generator runs (levels/town.cpp:359-387).
			grid := levels.CreateTown(dungeon, levels.TownOptions{
				Entry:     golden.entry,
				Files:     files,
				MegaTiles: megaTiles,
				Quests:    levels.Quests{PWATER: true},
			})

			// then: dPiece, dSpecial and the automap tile layer match.
			checkPieceLayer(t, grid.Piece, golden.piece)
			checkSpecialLayer(t, grid.Special, golden.special)
			checkTownTiles(t, dungeon.Tiles, golden.tiles)
		})
	}

	// then: no input the generator loads is missing from the committed corpus.
	if len(missing) > 0 {
		t.Fatalf("missing town .dun fixtures %v: CreateTown loads them at levels/town.cpp:202-207 "+
			"(docs/research/dg-town-inventory.md, \"Upstream files town loads\"); commit them as "+
			"setpieces-format CSVs (width,height then width*height uint16 tile values). Retail "+
			"DIABDAT.MPQ is not a legal test input (AGENTS.md fixture rule, the go-mpq defect "+
			"under triage)", missing)
	}
}

// TestTownCorpus pins the corpus itself, so a truncated or renamed golden fails
// here rather than silently shrinking TestTownParity's matrix.
func TestTownCorpus(t *testing.T) {
	// when: the committed town fixtures are globbed.
	goldens := townGoldens(t)

	// then: the corpus holds one fixture per pinned seed.
	if len(goldens) != len(committedSeeds) {
		t.Fatalf("committed town fixtures = %d, want %d", len(goldens), len(committedSeeds))
	}
}

// townGoldens parses every committed town fixture, one per pinned seed.
func townGoldens(t *testing.T) []*townGolden {
	t.Helper()

	paths, err := filepath.Glob(filepath.Join(townDir, "l0-town-*.csv"))
	if err != nil {
		t.Fatalf("glob town goldens: %v", err)
	}

	goldens := make([]*townGolden, 0, len(paths))
	for _, path := range paths {
		goldens = append(goldens, loadTownGolden(t, path))
	}

	return goldens
}

// loadTownGolden parses one fixture: "112,112" and the two world-tile records,
// then "40,40" and the automap tile layer (internal/testdata/drlg/README.md).
func loadTownGolden(t *testing.T, path string) *townGolden {
	t.Helper()

	records := readCSVRecords(t, path)
	if len(records) != 5 {
		t.Fatalf("%s: %d records, want 5", path, len(records))
	}

	checkPieceHeader(t, path, records[0])

	golden := &townGolden{}
	golden.entry, golden.seed = parseTownFilename(t, path)
	fillPieceRecord(t, path, records[1], &golden.piece)
	fillSpecialRecord(t, path, records[2], &golden.special)
	checkTownTileHeader(t, path, records[3])
	fillTownTiles(t, path, records[4], &golden.tiles)

	return golden
}

// parseTownFilename reads the entry and seed out of "l0-town-<entry>-<seed>.csv".
func parseTownFilename(t *testing.T, path string) (entry levels.Entry, seed uint32) {
	t.Helper()

	base := strings.TrimSuffix(filepath.Base(path), ".csv")
	parts := strings.Split(base, "-")
	if len(parts) != 4 || parts[0] != "l0" || parts[1] != "town" {
		t.Fatalf("%s: not an l0-town-<entry>-<seed> fixture name", path)
	}

	switch parts[2] {
	case "main":
		entry = levels.EntryMain
	case "prev":
		entry = levels.EntryPrev
	default:
		t.Fatalf("%s: unknown entry %q", path, parts[2])
	}

	parsed, err := strconv.ParseUint(parts[3], 10, 32)
	if err != nil {
		t.Fatalf("%s: seed %q: %v", path, parts[3], err)
	}

	return entry, uint32(parsed)
}

// checkTownTileHeader pins the "40,40" record of the automap tile layer.
func checkTownTileHeader(t *testing.T, path string, record []string) {
	t.Helper()

	got := parseUint16Record(t, path, record)
	if len(got) != 2 || got[0] != levels.DMAXX || got[1] != levels.DMAXY {
		t.Fatalf("%s: tile layer header %v, want %d,%d", path, record, levels.DMAXX, levels.DMAXY)
	}
}

// fillTownTiles reads the automap tile layer into grid, index x*DMAXY+y.
func fillTownTiles(t *testing.T, path string, record []string, grid *[levels.DMAXX][levels.DMAXY]levels.Tile) {
	t.Helper()

	if want := levels.DMAXX * levels.DMAXY; len(record) != want {
		t.Fatalf("%s: %d tile layer values, want %d", path, len(record), want)
	}

	for index, field := range record {
		value, err := strconv.ParseUint(field, 10, 8)
		if err != nil {
			t.Fatalf("%s: dungeon value %d %q: %v", path, index, field, err)
		}
		grid[index/levels.DMAXY][index%levels.DMAXY] = levels.Tile(value)
	}
}

// townDunFiles loads the fixed `.dun` layers CreateTown reads, keyed by resource
// path. A missing fixture is returned in missing rather than fatal, so the
// corpus gap reports as its own sub-test instead of masking the layout diff.
func townDunFiles(t *testing.T) (files map[string][]byte, missing []string) {
	t.Helper()

	files = make(map[string][]byte, len(townInputs))
	for _, input := range townInputs {
		path := filepath.Join(townDir, "inputs", input.fixture+".csv")
		data, err := os.ReadFile(path)
		if err != nil {
			missing = append(missing, path)
			continue
		}

		dun, err := testsupport.DecodeDunTilesCSV(data)
		if err != nil {
			t.Fatalf("decode %s: %v", path, err)
		}

		files[input.resource] = dun
	}

	return files, missing
}

// checkTownTiles compares two automap tile layers index by index.
func checkTownTiles(t *testing.T, got, want [levels.DMAXX][levels.DMAXY]levels.Tile) {
	t.Helper()

	for x := range got {
		for y := range got[x] {
			if got[x][y] != want[x][y] {
				t.Fatalf("dungeon tile (%d,%d) = %d, want %d", x, y, got[x][y], want[x][y])
			}
		}
	}
}
