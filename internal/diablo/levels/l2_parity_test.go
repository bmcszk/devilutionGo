package levels_test

import (
	"fmt"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/levels"
)

// l2Variants is the catacombs matrix the corpus was produced for: levels 5-8,
// both entries.
var l2Variants = []struct {
	level int
	entry levels.Entry
	name  string
}{
	{5, levels.EntryMain, "main"},
	{5, levels.EntryPrev, "prev"},
	{6, levels.EntryMain, "main"},
	{6, levels.EntryPrev, "prev"},
	{7, levels.EntryMain, "main"},
	{7, levels.EntryPrev, "prev"},
	{8, levels.EntryMain, "main"},
	{8, levels.EntryPrev, "prev"},
}

// TestL2CatacombsParity is the catacombs acceptance gate: every committed
// catacombs fixture must be reproduced byte for byte, tile layer,
// transparency window and view position.
//
//nolint:revive // the seed x variant matrix is the AC; splitting it would hide which golden failed
func TestL2CatacombsParity(t *testing.T) {
	for _, seed := range committedSeeds {
		goldens := loadGoldens(t, seed)
		setPieces := committedSetPieces(t)

		for _, variant := range l2Variants {
			name := fmt.Sprintf("l%d-%s", variant.level, variant.name)
			t.Run(fmt.Sprintf("%s-seed-%d", name, seed), func(t *testing.T) {
				// given: the committed fixture of this level/entry/seed variant.
				golden := goldens[name]
				if golden == nil {
					t.Fatalf("missing committed fixture for %s seed %d", name, seed)
				}

				dungeon := levels.NewDungeon(levels.Catacombs, variant.level)
				dungeon.Rnd.Seed(seed)

				// when: GenerateL2 runs with the options the probe used.
				err := levels.GenerateL2(dungeon, levels.Options{
					Entry:     variant.entry,
					Quests:    catacombsQuests(variant.level),
					SetPieces: setPieces,
				})
				if err != nil {
					t.Fatalf("GenerateL2: %v", err)
				}

				// then: the tile layer, the transparency window and the view position match.
				compareTiles(t, dungeon, golden)
				compareTransparency(t, dungeon, golden)
				compareView(t, dungeon, golden)
			})
		}
	}
}

// catacombsQuests is the quest availability the probe ran with on the
// catacombs levels: full quests, randomize-quests off, single player
// (internal/testdata/gamedata/quests/questdat.tsv + tables/questdat.cpp:64-76).
// All three are Quest::IsAvailable() results; Q_BLIND sits on level 7, Q_BLOOD
// on level 5 and Q_SCHAMB on level 6.
func catacombsQuests(level int) levels.Quests {
	return levels.Quests{
		Blind:  level == 7,
		Blood:  level == 5,
		Schamb: level == 6,
	}
}
