package levels_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/levels"
)

// l3Variants is the caves matrix the corpus was produced for: levels 9-12,
// both entries.
var l3Variants = []struct {
	level int
	entry levels.Entry
	name  string
}{
	{9, levels.EntryMain, "main"},
	{9, levels.EntryPrev, "prev"},
	{10, levels.EntryMain, "main"},
	{10, levels.EntryPrev, "prev"},
	{11, levels.EntryMain, "main"},
	{11, levels.EntryPrev, "prev"},
	{12, levels.EntryMain, "main"},
	{12, levels.EntryPrev, "prev"},
}

// TestL3CavesParity is the caves acceptance gate: every committed caves
// fixture must be reproduced byte for byte, tile layer, transparency window
// and view position.
//
//nolint:revive // the seed x variant matrix is the AC; splitting it would hide which golden failed
func TestL3CavesParity(t *testing.T) {
	for _, seed := range committedSeeds {
		goldens := loadGoldens(t, seed)
		setPieces := l3SetPieces(t)

		for _, variant := range l3Variants {
			name := fmt.Sprintf("l%d-%s", variant.level, variant.name)
			t.Run(fmt.Sprintf("%s-seed-%d", name, seed), func(t *testing.T) {
				// given: the committed fixture of this level/entry/seed variant.
				golden := goldens[name]
				if golden == nil {
					t.Fatalf("missing committed fixture for %s seed %d", name, seed)
				}

				dungeon := levels.NewDungeon(levels.Caves, variant.level)
				dungeon.Rnd.Seed(seed)

				// when: GenerateL3 runs with the options the probe used.
				err := levels.GenerateL3(dungeon, levels.Options{
					Entry:     variant.entry,
					Quests:    cavesQuests(variant.level),
					SetPieces: setPieces,
				})

				// then: the tile layer, the transparency window and the view position match.
				compareParity(t, dungeon, err, golden)
			})
		}
	}
}

// l3SetPieces loads the caves .dun resources the generator asks for by path.
// The anvil room is the only one: PlaceAnvil loads levels/l3data/anvil.dun
// (levels/drlg_l3.cpp:1745).
func l3SetPieces(t *testing.T) map[string][]byte {
	t.Helper()

	return map[string][]byte{
		"levels/l3data/anvil.dun": encodeDun(t, filepath.Join("testdata", "l3data", "anvil.csv")),
	}
}

// cavesQuests is the quest availability the probe ran with on the caves
// levels: full quests, randomize-quests off, single player
// (internal/testdata/gamedata/quests/questdat.tsv + tables/questdat.cpp:64-76).
// Q_ANVIL is the only quest the caves generator reads; its qdlvl is 10
// (questdat.tsv row 10), so it is available on level 10 alone.
func cavesQuests(level int) levels.Quests {
	return levels.Quests{Anvil: level == 10}
}
