package levels_test

import (
	"fmt"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/levels"
)

// cryptVariants is the crypt matrix the corpus was produced for: the Hellfire
// crypt levels 21-24, both entries (internal/testdata/drlg/README.md).
var cryptVariants = []struct {
	level int
	entry levels.Entry
	name  string
}{
	{21, levels.EntryMain, "main"},
	{21, levels.EntryPrev, "prev"},
	{22, levels.EntryMain, "main"},
	{22, levels.EntryPrev, "prev"},
	{23, levels.EntryMain, "main"},
	{23, levels.EntryPrev, "prev"},
	{24, levels.EntryMain, "main"},
	{24, levels.EntryPrev, "prev"},
}

// TestCryptParity is AC 2: every committed crypt fixture must be reproduced
// byte for byte, tile layer, transparency window and view position.
//
//nolint:revive // the seed x variant matrix is the AC; splitting it would hide which golden failed
func TestCryptParity(t *testing.T) {
	for _, seed := range committedSeeds {
		goldens := loadGoldens(t, seed)
		setPieces := committedSetPieces(t)

		// ViewPosition is a process global upstream, and level 24 has no stairs
		// down, so its ENTRY_PREV run keeps the view the previous variant left
		// (levels/crypt.cpp:739-748). The probe ran variants in this order, so
		// the carry is generator behaviour, not a fixture quirk.
		view := levels.Point{}

		for _, variant := range cryptVariants {
			name := fmt.Sprintf("l%d-%s", variant.level, variant.name)
			t.Run(fmt.Sprintf("%s-seed-%d", name, seed), func(t *testing.T) {
				// given: the committed fixture of this level/entry/seed variant.
				golden := goldens[name]
				if golden == nil {
					t.Fatalf("missing committed fixture for %s seed %d", name, seed)
				}

				dungeon := levels.NewDungeon(levels.Crypt, variant.level)
				dungeon.Rnd.Seed(seed)
				dungeon.ViewPosition = view

				// when: GenerateL1 runs the crypt branch with the probe's options.
				err := levels.GenerateL1(dungeon, levels.Options{
					Entry:     variant.entry,
					Quests:    cryptQuests(variant.level),
					SetPieces: setPieces,
				})
				if err != nil {
					t.Fatalf("GenerateL1: %v", err)
				}

				view = dungeon.ViewPosition

				// then: the tile layer, the transparency window and the view position match.
				compareTiles(t, dungeon, golden)
				compareTransparency(t, dungeon, golden)
				compareView(t, dungeon, golden)
			})
		}
	}
}

// cryptQuests is the quest availability the probe ran with on the crypt
// levels: no quest has a qdlvl of 21-24, so only the cornerstone item is
// available, and only on level 21 (items.cpp:4865-4868).
func cryptQuests(level int) levels.Quests {
	return levels.Quests{CornerStone: level == 21}
}
