package levels_test

// Temporary diagnostic: dumps the Go caves stage sequence so it can be diffed
// against the C++ stage dumps of tools/probe. Set L3STAGE to enable, L3LEVEL
// and L3SEED to pick the case.

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/levels"
)

// TestL3StageDump writes one tile-layer file per generator stage.
func TestL3StageDump(t *testing.T) {
	dir := os.Getenv("L3STAGE")
	if dir == "" {
		t.Skip("L3STAGE unset")
	}

	// given
	level, seed := stageCase(t)

	d := levels.NewDungeon(levels.Caves, level)
	d.Rnd.Seed(uint32(seed))
	d.DumpStage = func(tag string) { dumpCaveStage(t, dir, tag, d) }

	// when
	err := levels.GenerateL3(d, levels.Options{
		Entry:     levels.EntryMain,
		Quests:    levels.Quests{Anvil: os.Getenv("L3ANVIL") == "1"},
		SetPieces: map[string][]byte{"levels/l3data/anvil.dun": encodeDun(t, "testdata/l3data/anvil.csv")},
	})
	if err != nil {
		t.Fatal(err)
	}

	// then
	t.Logf("level=%d levelseed=%d", level, d.LevelSeed)
}

// stageCase reads the level and seed the dump runs with.
func stageCase(t *testing.T) (level, seed int) {
	t.Helper()

	level = envInt(t, "L3LEVEL", 9)
	seed = envInt(t, "L3SEED", 2588)

	return level, seed
}

// envInt parses an integer environment variable, falling back to fallback.
func envInt(t *testing.T, name string, fallback int) int {
	t.Helper()

	raw := os.Getenv(name)
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		t.Fatalf("%s=%q: %v", name, raw, err)
	}

	return value
}

// dumpCaveStage writes one stage's tile layer as a 40x40 byte grid.
func dumpCaveStage(t *testing.T, dir, tag string, d *levels.Dungeon) {
	t.Helper()

	tiles := make([]byte, 0, levels.DMAXX*levels.DMAXY)
	for i := range levels.DMAXX {
		for j := range levels.DMAXY {
			tiles = append(tiles, byte(d.Tiles[i][j]))
		}
	}

	path := fmt.Sprintf("%s/%s-%d-%d.bin", dir, tag, d.CurrLevel, d.LevelSeed)
	if err := os.WriteFile(path, tiles, 0o600); err != nil {
		t.Fatal(err)
	}
}
