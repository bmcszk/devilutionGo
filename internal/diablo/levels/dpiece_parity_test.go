package levels_test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/levels"
)

// dpieceVariants is the matrix the dPiece corpus was produced for: levels 1-16
// plus the Hellfire crypt levels 21-24, both entries, each with the family
// generator that serves it and the mega-tile table that generator was run
// against (gendung.cpp:326-344).
var dpieceVariants = []struct {
	level     int
	entry     levels.Entry
	name      string
	levelType levels.LevelType
	family    string
}{
	{1, levels.EntryMain, "main", levels.Cathedral, "l1"},
	{1, levels.EntryPrev, "prev", levels.Cathedral, "l1"},
	{2, levels.EntryMain, "main", levels.Cathedral, "l1"},
	{2, levels.EntryPrev, "prev", levels.Cathedral, "l1"},
	{3, levels.EntryMain, "main", levels.Cathedral, "l1"},
	{3, levels.EntryPrev, "prev", levels.Cathedral, "l1"},
	{4, levels.EntryMain, "main", levels.Cathedral, "l1"},
	{4, levels.EntryPrev, "prev", levels.Cathedral, "l1"},
	{5, levels.EntryMain, "main", levels.Catacombs, "l2"},
	{5, levels.EntryPrev, "prev", levels.Catacombs, "l2"},
	{6, levels.EntryMain, "main", levels.Catacombs, "l2"},
	{6, levels.EntryPrev, "prev", levels.Catacombs, "l2"},
	{7, levels.EntryMain, "main", levels.Catacombs, "l2"},
	{7, levels.EntryPrev, "prev", levels.Catacombs, "l2"},
	{8, levels.EntryMain, "main", levels.Catacombs, "l2"},
	{8, levels.EntryPrev, "prev", levels.Catacombs, "l2"},
	{9, levels.EntryMain, "main", levels.Caves, "l3"},
	{9, levels.EntryPrev, "prev", levels.Caves, "l3"},
	{10, levels.EntryMain, "main", levels.Caves, "l3"},
	{10, levels.EntryPrev, "prev", levels.Caves, "l3"},
	{11, levels.EntryMain, "main", levels.Caves, "l3"},
	{11, levels.EntryPrev, "prev", levels.Caves, "l3"},
	{12, levels.EntryMain, "main", levels.Caves, "l3"},
	{12, levels.EntryPrev, "prev", levels.Caves, "l3"},
	{13, levels.EntryMain, "main", levels.Hell, "l4"},
	{13, levels.EntryPrev, "prev", levels.Hell, "l4"},
	{14, levels.EntryMain, "main", levels.Hell, "l4"},
	{14, levels.EntryPrev, "prev", levels.Hell, "l4"},
	{15, levels.EntryMain, "main", levels.Hell, "l4"},
	{15, levels.EntryPrev, "prev", levels.Hell, "l4"},
	{16, levels.EntryMain, "main", levels.Hell, "l4"},
	{16, levels.EntryPrev, "prev", levels.Hell, "l4"},
	{21, levels.EntryMain, "main", levels.Crypt, "l5"},
	{21, levels.EntryPrev, "prev", levels.Crypt, "l5"},
	{22, levels.EntryMain, "main", levels.Crypt, "l5"},
	{22, levels.EntryPrev, "prev", levels.Crypt, "l5"},
	{23, levels.EntryMain, "main", levels.Crypt, "l5"},
	{23, levels.EntryPrev, "prev", levels.Crypt, "l5"},
	{24, levels.EntryMain, "main", levels.Crypt, "l5"},
	{24, levels.EntryPrev, "prev", levels.Crypt, "l5"},
}

// TestDPieceParity is the Pass3 acceptance gate: every committed dPiece fixture
// must be reproduced world tile for world tile, and its dSpecial record too
// where the corpus has one.
//
//nolint:revive // the seed x variant matrix is the AC; splitting it would hide which golden failed
func TestDPieceParity(t *testing.T) {
	megaTiles := pieceMegaTiles(t)

	for _, seed := range committedSeeds {
		goldens := loadPieceGoldens(t, seed)

		for _, variant := range dpieceVariants {
			name := fmt.Sprintf("l%d-%s", variant.level, variant.name)
			t.Run(fmt.Sprintf("%s-seed-%d", name, seed), func(t *testing.T) {
				// given: the committed fixture of this level/entry/seed variant.
				golden := goldens[name]
				if golden == nil {
					t.Fatalf("missing committed fixture for %s seed %d", name, seed)
				}

				dungeon := levels.NewDungeon(variant.levelType, variant.level)
				dungeon.Rnd.Seed(seed)

				// when: the family generator runs and Pass3 expands the layer.
				setPieces := familySetPieces(t, variant.levelType)
				if err := generateFamily(dungeon, variant.levelType, variant.entry, setPieces); err != nil {
					t.Fatalf("generate %s: %v", name, err)
				}
				grid := levels.ExpandPieces(dungeon, megaTiles[variant.family])

				// then: dPiece, and dSpecial where committed, match the fixture.
				comparePieces(t, grid, golden)
			})
		}
	}
}

// TestDPieceCorpus pins the corpus itself, so a truncated or renamed golden
// fails here rather than silently shrinking the matrix above. Nest fixtures
// share the directory and are pinned by TestHiveCorpus.
func TestDPieceCorpus(t *testing.T) {
	// given: the committed fixtures, indexed by base name.
	paths := pieceGoldenPaths(t)
	present := make(map[string]bool, len(paths))
	for _, path := range paths {
		present[strings.TrimSuffix(filepath.Base(path), ".csv")] = true
	}

	// when: every variant of this matrix is looked up.
	for _, variant := range dpieceVariants {
		for _, seed := range committedSeeds {
			key := fmt.Sprintf("l%d-dpiece-%s-%d", variant.level, variant.name, seed)

			// then: its fixture exists.
			if !present[key] {
				t.Errorf("missing committed dPiece fixture %s.csv", key)
			}
		}
	}
}

// familySetPieces loads the .dun resources the generator of one level family
// asks for by path: the caves anvil room and the hell rooms live outside the
// shared l1data/l2data set (levels/drlg_l3.cpp:1745, drlg_l4.cpp:175-177).
func familySetPieces(t *testing.T, levelType levels.LevelType) map[string][]byte {
	t.Helper()

	switch levelType {
	case levels.Caves:
		return l3SetPieces(t)
	case levels.Hell:
		return l4SetPieces(t)
	default:
		return committedSetPieces(t)
	}
}

// generateFamily dispatches one variant to the generator of its level family,
// with the options the committed tile-layer goldens were produced with.
func generateFamily(d *levels.Dungeon, levelType levels.LevelType, entry levels.Entry,
	setPieces map[string][]byte,
) error {
	switch levelType {
	case levels.Cathedral:
		return generateCathedral(d, entry, setPieces)
	case levels.Catacombs:
		return generateCatacombs(d, entry, setPieces)
	case levels.Caves:
		return generateCaves(d, entry, setPieces)
	case levels.Hell:
		return generateHell(d, entry, setPieces)
	case levels.Crypt:
		return generateCrypt(d, entry, setPieces)
	default:
		return fmt.Errorf("no generator for level type %d", levelType)
	}
}

// generateCathedral runs CreateL5Dungeon's cathedral branch.
func generateCathedral(d *levels.Dungeon, entry levels.Entry, setPieces map[string][]byte) error {
	err := levels.GenerateL1(d, levels.Options{
		Entry:             entry,
		OriginalCathedral: true,
		Quests:            cathedralQuests(d.CurrLevel),
		SetPieces:         setPieces,
	})
	if err != nil {
		return fmt.Errorf("generate cathedral level %d: %w", d.CurrLevel, err)
	}

	return nil
}

// generateCatacombs runs CreateL2Dungeon.
func generateCatacombs(d *levels.Dungeon, entry levels.Entry, setPieces map[string][]byte) error {
	err := levels.GenerateL2(d, levels.Options{
		Entry:     entry,
		Quests:    catacombsQuests(d.CurrLevel),
		SetPieces: setPieces,
	})
	if err != nil {
		return fmt.Errorf("generate catacombs level %d: %w", d.CurrLevel, err)
	}

	return nil
}

// generateCaves runs CreateL3Dungeon.
func generateCaves(d *levels.Dungeon, entry levels.Entry, setPieces map[string][]byte) error {
	err := levels.GenerateL3(d, levels.Options{
		Entry:     entry,
		Quests:    cavesQuests(d.CurrLevel),
		SetPieces: setPieces,
	})
	if err != nil {
		return fmt.Errorf("generate caves level %d: %w", d.CurrLevel, err)
	}

	return nil
}

// generateCrypt runs CreateL5Dungeon's crypt branch (`drlg_l1.cpp:1298-1313`).
func generateCrypt(d *levels.Dungeon, entry levels.Entry, setPieces map[string][]byte) error {
	err := levels.GenerateL1(d, levels.Options{
		Entry:     entry,
		Quests:    cryptQuests(d.CurrLevel),
		SetPieces: setPieces,
	})
	if err != nil {
		return fmt.Errorf("generate crypt level %d: %w", d.CurrLevel, err)
	}

	return nil
}

// generateHell runs CreateL4Dungeon.
func generateHell(d *levels.Dungeon, entry levels.Entry, setPieces map[string][]byte) error {
	err := levels.GenerateL4(d, levels.Options{
		Entry:     entry,
		Quests:    hellQuests(d.CurrLevel),
		SetPieces: setPieces,
	})
	if err != nil {
		return fmt.Errorf("generate hell level %d: %w", d.CurrLevel, err)
	}

	return nil
}
