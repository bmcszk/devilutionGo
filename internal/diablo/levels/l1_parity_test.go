package levels_test

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/levels"
)

// l1Variants is the cathedral matrix the corpus was produced for: levels 1-4,
// both entries.
var l1Variants = []struct {
	level int
	entry levels.Entry
	name  string
}{
	{1, levels.EntryMain, "main"},
	{1, levels.EntryPrev, "prev"},
	{2, levels.EntryMain, "main"},
	{2, levels.EntryPrev, "prev"},
	{3, levels.EntryMain, "main"},
	{3, levels.EntryPrev, "prev"},
	{4, levels.EntryMain, "main"},
	{4, levels.EntryPrev, "prev"},
}

// TestL1CathedralParity is AC 1: every committed cathedral fixture must be
// reproduced byte for byte, tile layer, transparency window and view position.
//
//nolint:revive // the seed x variant matrix is the AC; splitting it would hide which golden failed
func TestL1CathedralParity(t *testing.T) {
	for _, seed := range committedSeeds {
		goldens := loadGoldens(t, seed)
		setPieces := committedSetPieces(t)

		for _, variant := range l1Variants {
			name := fmt.Sprintf("l%d-%s", variant.level, variant.name)
			t.Run(fmt.Sprintf("%s-seed-%d", name, seed), func(t *testing.T) {
				// given: the committed fixture of this level/entry/seed variant.
				golden := goldens[name]
				if golden == nil {
					t.Fatalf("missing committed fixture for %s seed %d", name, seed)
				}

				dungeon := levels.NewDungeon(levels.Cathedral, variant.level)
				dungeon.Rnd.Seed(seed)

				// when: GenerateL1 runs with the options the probe used.
				err := levels.GenerateL1(dungeon, levels.Options{
					Entry:             variant.entry,
					OriginalCathedral: true,
					Quests:            cathedralQuests(variant.level),
					SetPieces:         setPieces,
				})
				if err != nil {
					t.Fatalf("GenerateL1: %v", err)
				}

				// then: the tile layer, the transparency window and the view position match.
				compareTiles(t, dungeon, golden)
				compareTransparency(t, dungeon, golden)
				compareView(t, dungeon, golden)
			})
		}
	}
}

// TestCorpusShape pins the committed corpus size so a silently truncated
// fixture set fails instead of passing with fewer cases.
func TestCorpusShape(t *testing.T) {
	// when: the committed corpus is globbed and its seed suffixes collected.
	paths, err := filepath.Glob(filepath.Join(drlgDir, "l*-*.csv"))
	if err != nil {
		t.Fatalf("glob corpus: %v", err)
	}

	seeds := map[string]struct{}{}
	for _, path := range paths {
		base := strings.TrimSuffix(filepath.Base(path), ".csv")
		seeds[base[strings.LastIndex(base, "-")+1:]] = struct{}{}
	}

	// then: the 24 variants x 2 entries x 8 seeds matrix is on disk (levels
	// 1-4 cathedral, 5-8 catacombs, 21-24 Hellfire crypt).
	if len(paths) != 192 {
		t.Fatalf("corpus files = %d, want 192", len(paths))
	}

	if len(seeds) != 8 {
		t.Fatalf("corpus seeds = %d, want 8", len(seeds))
	}
}

// cathedralQuests is the quest availability the probe ran with: full quests,
// randomize-quests off, single player, on the cathedral levels
// (internal/testdata/gamedata/quests/questdat.tsv + tables/questdat.cpp:64-76).
// All five are Quest::IsAvailable() results, not quest completion state.
func cathedralQuests(level int) levels.Quests {
	return levels.Quests{
		Butcher:  level == 2, // Q_BUTCHER qdlvl 2
		PWATER:   level == 2, // Q_PWATER qdlvl 2
		Skelking: level == 3, // Q_SKELKING qdlvl 3
		LTBanner: level == 4, // Q_LTBANNER qdlvl 4
	}
}

// committedSetPieces loads the .dun tile layers the generator can ask for.
func committedSetPieces(t *testing.T) map[string][]byte {
	t.Helper()

	names := map[string]string{
		"levels/l1data/rnd6.dun":         "rnd6",
		"levels/l1data/skngdo.dun":       "skngdo",
		"levels/l1data/banner2.dun":      "banner2",
		"levels/l2data/blind1.dun":       "blind1",
		"levels/l2data/blood1.dun":       "blood1",
		"levels/l2data/blood2.dun":       "blood2",
		"levels/l2data/bonestr1.dun":     "bonestr1",
		"levels/l2data/bonestr2.dun":     "bonestr2",
		"nlevels/l5data/cornerstone.dun": "cornerstone",
		"nlevels/l5data/uberroom.dun":    "uberroom",
	}

	pieces := make(map[string][]byte, len(names))
	for key, name := range names {
		pieces[key] = encodeDun(t, filepath.Join(drlgDir, "setpieces", name+".csv"))
	}

	return pieces
}

// encodeDun rebuilds the .dun byte prefix the loader reads: two little-endian
// dimensions followed by the tile layer as little-endian uint16 values.
func encodeDun(t *testing.T, path string) []byte {
	t.Helper()

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open set piece %s: %v", path, err)
	}

	defer file.Close() //nolint:errcheck // read-only fixture

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("parse set piece %s: %v", path, err)
	}

	if len(records) != 2 {
		t.Fatalf("%s: %d records, want 2", path, len(records))
	}

	width, height := parseNum(t, path, records[0][0]), parseNum(t, path, records[0][1])
	if len(records[1]) != width*height {
		t.Fatalf("%s: %d tiles, want %d", path, len(records[1]), width*height)
	}

	data := make([]byte, 4+2*width*height)
	data[0], data[1] = byte(width), byte(width>>8)
	data[2], data[3] = byte(height), byte(height>>8)

	for i, cell := range records[1] {
		value := parseNum(t, path, cell)
		data[4+2*i], data[5+2*i] = byte(value), byte(value>>8)
	}

	return data
}

func compareTiles(t *testing.T, dungeon *levels.Dungeon, golden *dunGolden) {
	t.Helper()

	for x := range levels.DMAXX {
		for y := range levels.DMAXY {
			if got := uint8(dungeon.Tiles[x][y]); got != golden.tiles[x][y] {
				t.Fatalf("dungeon tile (%d,%d) = %d, want %d", x, y, got, golden.tiles[x][y])
			}
		}
	}
}

// compareTransparency diffs the doubled dTransVal window. The corpus stores it
// flattened the way the generator indexes it — first index x, second index y —
// exactly like the tile layer (golden_test.go:83-87).
func compareTransparency(t *testing.T, dungeon *levels.Dungeon, golden *dunGolden) {
	t.Helper()

	for x := range 80 {
		for y := range 80 {
			got := uint8(dungeon.Trans[x+16][y+16])
			if got != golden.trans[x][y] {
				t.Fatalf("dTransVal (%d,%d) = %d, want %d", x+16, y+16, got, golden.trans[x][y])
			}
		}
	}
}

func compareView(t *testing.T, dungeon *levels.Dungeon, golden *dunGolden) {
	t.Helper()

	if dungeon.ViewPosition != golden.view {
		t.Fatalf("ViewPosition = %+v, want %+v", dungeon.ViewPosition, golden.view)
	}
}
