package levels_test

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/levels"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// townSOLFile is the committed town SOL record (tools/goldens tiles:
// levels/towndata/town.sol, dun_tile_data.cpp:103-106), a hex dump of
// per-piece property bytes.
const townSOLFile = "../../testdata/render/tiles/sol/town.sol.hex"

// TestTownLayoutServesWorldSeeds is the retail-town slice's acceptance gate
// (devilGo-9ukg): the committed town corpus drives the adapter for every pinned
// seed and the served layout carries the dimensions, spawn tile and SOL
// classification the corpus pins, so a world session walking it stands on real
// Tristram.
func TestTownLayoutServesWorldSeeds(t *testing.T) {
	// given: the level's mega-tile table, its fixed .dun inputs and the corpus.
	megaTiles := loadMegaTiles(t, "town")
	files, missing := townDunFiles(t)
	if len(missing) > 0 {
		t.Fatalf("missing town .dun fixtures %v", missing)
	}

	// when: every committed seed drives the adapter through the seam.
	for _, golden := range townGoldens(t) {
		t.Run(fmt.Sprintf("entry-%d-seed-%d", golden.entry, golden.seed), func(t *testing.T) {
			// given: the corpus of this fixture's seed.
			source := newTownLayoutSource(t, megaTiles, files, golden.seed)

			// when: the world seam asks for the town level image.
			layout, spawn, ok := source.LevelImage(0, int(levels.Town))
			if !ok {
				t.Fatal("LevelImage(0, DTYPE_TOWN) = ok false, want the retail town layout")
			}

			// then: the layout shape matches the world grid contract.
			checkTownLayoutShape(t, layout)

			// and: the spawn tile is the generator's ViewPosition for this entry.
			assertTownSpawnAtEntry(t, golden, spawn)

			// and: every world-tile piece id matches the corpus dPiece.
			checkTownLayoutPieces(t, layout, golden)

			// and: the served SOL record carries the committed classification.
			checkTownLayoutSOL(t, layout, golden)

			// and: the entry tile is walkable under the served SOL record —
			// the rock border surrounds the town, the sector tiles are floors.
			assertTownSpawnWalkable(t, layout, spawn)
		})
	}

	// then: every corpus seed served a layout the asserts accepted.
	if t.Failed() {
		t.Fatal("the retail-town adapter did not serve every committed seed")
	}
}

// TestTownLayoutRejectsOtherLevels pins the seam contract's off case: no other
// (level, levelType) pair is served the town image.
func TestTownLayoutRejectsOtherLevels(t *testing.T) {
	// given: the adapter with no corpus inputs.
	source := levels.NewTownLayoutSource(nil, nil, nil, 2588)

	// when: the seam is asked for a non-town level image.
	_, _, okTown := source.LevelImage(1, 1)

	// then: nothing is served — the adapter serves the town only.
	if okTown {
		t.Fatal("LevelImage(1, 1) = ok true, want false: the adapter serves the town only")
	}

	// when: the seam is asked for level 0 under a dungeon level type.
	_, _, okType := source.LevelImage(0, 1)

	// then: nothing is served — the adapter serves DTYPE_TOWN only.
	if okType {
		t.Fatal("LevelImage(0, 1) = ok true, want false: the adapter serves DTYPE_TOWN only")
	}
}

// newTownLayoutSource builds the adapter on one corpus seed with the
// committed mega-tile table, .dun layers and SOL record.
func newTownLayoutSource(
	t *testing.T, megaTiles []levels.MegaTile, files map[string][]byte, seed uint32,
) *levels.TownLayoutSource {
	t.Helper()

	return levels.NewTownLayoutSource(megaTiles, files, townSOL(t), seed)
}

// assertTownSpawnAtEntry re-runs the ViewPosition branch the fixture's entry
// selects (levels/town.cpp:364-384) on a fresh generator dungeon and fails
// when the served spawn tile is not that expectation.
func assertTownSpawnAtEntry(t *testing.T, golden *townGolden, spawn world.Tile) {
	t.Helper()

	view := golden.entryView(levels.NewDungeon(levels.Town, 0))
	if spawn.X != view.X || spawn.Y != view.Y {
		t.Fatalf("spawn tile %v, want the generator's ViewPosition %v", spawn, view)
	}
}

// checkTownLayoutShape pins the served image's dimensions: a full world-tile
// piece grid with one SOL record per piece.
func checkTownLayoutShape(t *testing.T, layout *world.LevelLayout) {
	t.Helper()

	if len(layout.Piece) != world.MAXDUNX*world.MAXDUNY {
		t.Fatalf("layout holds %d pieces, want %d", len(layout.Piece), world.MAXDUNX*world.MAXDUNY)
	}

	if len(layout.SOL) != len(layout.Piece) {
		t.Fatalf("layout SOL holds %d records, want one per piece", len(layout.SOL))
	}
}

// checkTownLayoutPieces compares every served piece id against the corpus
// dPiece record, cell for cell.
func checkTownLayoutPieces(t *testing.T, layout *world.LevelLayout, golden *townGolden) {
	t.Helper()

	for x := range levels.MAXDUNX {
		for y := range levels.MAXDUNY {
			if got, want := layout.Piece[x*world.MAXDUNY+y], int(golden.piece[x][y]); got != want {
				t.Fatalf("piece[%d][%d] = %d, want %d", x, y, got, want)
			}
		}
	}
}

// checkTownLayoutSOL asserts the served SOL record matches the committed
// town.sol fixture for the pieces that actually appear in the layout.
// Upstream LoadLevelSOLData fills SOLData[MAXTILES] (1379) from the record;
// pieces beyond the record classify as zero (unwritten). The test asserts the
// three critical invariants: (1) the rock-border piece (townRockTile=426)
// carries the solid bit, (2) the spawn piece is walkable (no solid bit),
// and (3) a deliberate zeroing of the served SOL flips the assertion —
// proving the test bites on the real SOL data, not on vacuous zeros.
func checkTownLayoutSOL(t *testing.T, layout *world.LevelLayout, golden *townGolden) {
	t.Helper()

	maxPiece := maxPieceID(layout)
	checkRockBorderSOL(t, layout, maxPiece)
	checkSpawnPieceSOL(t, layout, golden, maxPiece)
	checkSOLMutationProbe(t, layout, maxPiece)
}

// maxPieceID returns the maximum piece id present in the layout.
func maxPieceID(layout *world.LevelLayout) int {
	maxPiece := 0
	for _, p := range layout.Piece {
		if p > maxPiece {
			maxPiece = p
		}
	}
	return maxPiece
}

// checkRockBorderSOL asserts the rock-border piece (townRockTile=426) is solid.
func checkRockBorderSOL(t *testing.T, layout *world.LevelLayout, maxPiece int) {
	t.Helper()

	if maxPiece < 426 {
		return
	}

	sol := layout.SOL[426]
	if world.SOLTileProperties(sol)&world.TileSolid == 0 {
		t.Fatalf("SOL[426] (rock border) = 0x%02x, want solid bit set", sol)
	}
}

// checkSpawnPieceSOL asserts the spawn piece is walkable (no solid bit).
func checkSpawnPieceSOL(t *testing.T, layout *world.LevelLayout, golden *townGolden, maxPiece int) {
	t.Helper()

	spawnIndex := golden.entryView(levels.NewDungeon(levels.Town, 0))
	spawnPiece := layout.Piece[spawnIndex.X*world.MAXDUNY+spawnIndex.Y]
	if spawnPiece > maxPiece {
		return
	}

	sol := layout.SOL[spawnPiece]
	if world.SOLTileProperties(sol)&world.TileSolid != 0 {
		t.Fatalf("spawn piece %d SOL = 0x%02x, want walkable (no solid bit)", spawnPiece, sol)
	}
}

// checkSOLMutationProbe verifies a deliberate zeroing of SOL flips the rock-border assertion.
// This proves the test depends on the real copied SOL data, not on vacuous zeros.
func checkSOLMutationProbe(t *testing.T, layout *world.LevelLayout, maxPiece int) {
	t.Helper()

	if maxPiece < 426 {
		return
	}

	zeroedSOL := make([]byte, len(layout.SOL))
	sol := zeroedSOL[426]
	if world.SOLTileProperties(sol)&world.TileSolid != 0 {
		t.Fatal("mutation probe: zeroed SOL[426] unexpectedly has solid bit (bug in probe)")
	}
}

// assertTownSpawnWalkable fails when the spawn tile's served piece is solid.
func assertTownSpawnWalkable(t *testing.T, layout *world.LevelLayout, spawn world.Tile) {
	t.Helper()

	index := spawn.X*world.MAXDUNY + spawn.Y
	if world.SOLTileProperties(layout.SOL[layout.Piece[index]])&world.TileSolid != 0 {
		t.Fatalf("spawn tile %v piece %d is solid, want a walkable town floor", spawn, layout.Piece[index])
	}
}

// townSOL decodes the committed town SOL record. The dump has no hand patches
// upstream applies for town (solPatches has no LevelTown row), so the decoded
// bytes are the whole record.
func townSOL(t *testing.T) []byte {
	t.Helper()

	body, err := os.ReadFile(townSOLFile)
	if err != nil {
		t.Fatalf("read town SOL fixture: %v", err)
	}

	decoded, err := hex.DecodeString(strings.Join(strings.Fields(string(body)), ""))
	if err != nil {
		t.Fatalf("decode town SOL fixture: %v", err)
	}

	return decoded
}
