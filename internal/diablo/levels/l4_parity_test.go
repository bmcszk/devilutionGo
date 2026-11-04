package levels_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/levels"
)

// l4Variants is the hell matrix the corpus was produced for: levels 13-16,
// both entries.
var l4Variants = []struct {
	level int
	entry levels.Entry
	name  string
}{
	{13, levels.EntryMain, "main"},
	{13, levels.EntryPrev, "prev"},
	{14, levels.EntryMain, "main"},
	{14, levels.EntryPrev, "prev"},
	{15, levels.EntryMain, "main"},
	{15, levels.EntryPrev, "prev"},
	{16, levels.EntryMain, "main"},
	{16, levels.EntryPrev, "prev"},
}

// TestL4HellParity is the hell acceptance gate: every committed hell fixture
// must be reproduced byte for byte, tile layer, transparency window and view
// position.
//
// One dungeon carries the whole matrix because the probe generates it in a
// single process (tools/probe/dvg_probe.cpp DumpSeedVariants): the view position
// and pdungeon are globals that survive between variants. Level 16 entered from
// below never sets a view position, so its fixture keeps the one the level's
// main entry left (levels/drlg_l4.cpp:1106-1170).
//
//nolint:revive // the seed x variant matrix is the AC; splitting it would hide which golden failed
func TestL4HellParity(t *testing.T) {
	dungeon := levels.NewDungeon(levels.Hell, l4Variants[0].level)
	setPieces := l4SetPieces(t)

	for _, seed := range committedSeeds {
		goldens := loadGoldens(t, seed)

		for _, variant := range l4Variants {
			name := fmt.Sprintf("l%d-%s", variant.level, variant.name)
			t.Run(fmt.Sprintf("%s-seed-%d", name, seed), func(t *testing.T) {
				// given: the committed fixture of this level/entry/seed variant.
				golden := goldens[name]
				if golden == nil {
					t.Fatalf("missing committed fixture for %s seed %d", name, seed)
				}

				dungeon.CurrLevel = variant.level
				dungeon.LevelSeed = 0
				dungeon.Rnd.Seed(seed)

				// when: GenerateL4 runs with the options the probe used.
				err := levels.GenerateL4(dungeon, levels.Options{
					Entry:     variant.entry,
					Quests:    hellQuests(variant.level),
					SetPieces: setPieces,
				})

				// then: the tile layer, the transparency window and the view position match.
				compareParity(t, dungeon, err, golden)
			})
		}
	}
}

// l4SetPieces loads the hell .dun resources the generator asks for by path: the
// warlord room (levels/drlg_l4.cpp:175), the four Diablo quadrants and their
// pre-gate variants (levels/drlg_l4.cpp:970-985). vile1.dun is the multiplayer
// Lazarus room (levels/drlg_l4.cpp:177), which the single-player probe never
// reaches.
func l4SetPieces(t *testing.T) map[string][]byte {
	t.Helper()

	names := map[string]string{
		"levels/l4data/warlord.dun":  "warlord",
		"levels/l4data/warlord2.dun": "warlord2",
		"levels/l4data/diab1.dun":    "diab1",
		"levels/l4data/diab2a.dun":   "diab2a",
		"levels/l4data/diab2b.dun":   "diab2b",
		"levels/l4data/diab3a.dun":   "diab3a",
		"levels/l4data/diab3b.dun":   "diab3b",
		"levels/l4data/diab4a.dun":   "diab4a",
		"levels/l4data/diab4b.dun":   "diab4b",
	}

	pieces := make(map[string][]byte, len(names))
	for key, name := range names {
		pieces[key] = encodeDun(t, filepath.Join("testdata", "l4data", name+".csv"))
	}

	return pieces
}

// hellQuests is the quest availability the probe ran with on the hell levels:
// full quests, randomize-quests off, single player
// (internal/testdata/gamedata/quests/questdat.tsv + tables/questdat.cpp:64-76).
// The hell generator reads Q_WARLORD (qdlvl 13) and Q_DIABLO (qdlvl 15), so each
// is available on its own level alone.
func hellQuests(level int) levels.Quests {
	return levels.Quests{
		Warlord: level == 13, // Q_WARLORD qdlvl 13
		Diablo:  level == 15, // Q_DIABLO qdlvl 15
	}
}
