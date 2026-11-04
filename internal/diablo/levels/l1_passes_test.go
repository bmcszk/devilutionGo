package levels_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/levels"
)

// TestFixDirtTiles pins the six dirt-face rules: a dirt tile whose neighbour no
// longer continues its face is rewritten to the closing variant, and one whose
// neighbour does continue it is left alone (levels/drlg_l1.cpp:1111-1136).
func TestFixDirtTiles(t *testing.T) {
	t.Parallel()

	// HDirtCorner is probed twice upstream: against the V wall on its right and
	// against the H wall below it, so both have to continue for the tile to stay.
	tests := []struct {
		name  string
		tile  levels.Tile
		right levels.Tile
		below levels.Tile
		want  levels.Tile
	}{
		{
			name: "H dirt corner closed when the V wall is gone", tile: levels.TestHDirtCorner,
			right: levels.TestFloor, below: levels.TestDirtHWall, want: levels.TestDirtCorner2,
		},
		{
			name: "H dirt corner closed when the H wall below is gone", tile: levels.TestHDirtCorner,
			right: levels.TestDirtVWall, below: levels.TestFloor, want: levels.TestDirtCorner2,
		},
		{
			name: "H dirt corner kept while both walls continue", tile: levels.TestHDirtCorner,
			right: levels.TestDirtVWall, below: levels.TestDirtHWall, want: levels.TestHDirtCorner,
		},
		{
			name: "H dirt wall closed when the wall below ends", tile: levels.TestDirtHWall,
			below: levels.TestFloor, want: levels.TestDirtHWall2,
		},
		{
			name: "H dirt wall kept while the wall below continues", tile: levels.TestDirtHWall,
			below: levels.TestDirtHWall, want: levels.TestDirtHWall,
		},
		{
			name: "H dirt wall end closed when the wall below ends", tile: levels.TestDirtHWallEnd,
			below: levels.TestFloor, want: levels.TestDirtHWallEnd2,
		},
		{
			name: "V dirt wall end closed when the wall right ends", tile: levels.TestDirtVWallEnd,
			right: levels.TestFloor, want: levels.TestDirtVWallEnd2,
		},
		{
			name: "V dirt wall kept while the wall right continues", tile: levels.TestDirtVWall,
			right: levels.TestDirtVWall, want: levels.TestDirtVWall,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// given: one dirt tile at (2,2) and the neighbours the rules probe.
			dungeon := passDungeon()
			dungeon.Tiles[2][2] = test.tile
			dungeon.Tiles[3][2] = test.right
			dungeon.Tiles[2][3] = test.below

			// when: the pass runs over the layer.
			dungeon.FixDirtTilesForTest()

			// then: only the rule's target tile changed.
			if got := dungeon.Tiles[2][2]; got != test.want {
				t.Errorf("tile = %d, want %d", got, test.want)
			}
		})
	}
}

// TestFixCornerTiles pins the three corner rules and their guards: the HCorner
// case skips protected tiles and needs a floor on the left and a V wall above,
// the two DirtCorner2 cases need a floor right/below with the matching wall on
// the other side (levels/drlg_l1.cpp:1137-1154).
//
//nolint:revive // the rule x guard matrix is the AC; splitting the table would hide which case failed
func TestFixCornerTiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		tile      levels.Tile
		left      levels.Tile
		above     levels.Tile
		right     levels.Tile
		below     levels.Tile
		protected bool
		want      levels.Tile
	}{
		{
			name: "H corner becomes V corner", tile: levels.TestHCorner,
			left: levels.TestFloor, above: levels.TestVWall, want: levels.TestVCorner,
		},
		{
			name: "protected H corner stays", tile: levels.TestHCorner,
			left: levels.TestFloor, above: levels.TestVWall, protected: true, want: levels.TestHCorner,
		},
		{
			name: "H corner needs floor on the left", tile: levels.TestHCorner,
			left: levels.TestVWall, above: levels.TestVWall, want: levels.TestHCorner,
		},
		{
			name: "H corner needs a V wall above", tile: levels.TestHCorner,
			left: levels.TestFloor, above: levels.TestFloor, want: levels.TestHCorner,
		},
		{
			name: "floor right and V wall below give the H arch end", tile: levels.TestDirtCorner2,
			right: levels.TestFloor, below: levels.TestVWall, want: levels.TestHArchEnd,
		},
		{
			name: "floor below and H wall right give the V arch end", tile: levels.TestDirtCorner2,
			right: levels.TestHWall, below: levels.TestFloor, want: levels.TestVArchEnd,
		},
		{
			name: "dirt corner with no wall stays", tile: levels.TestDirtCorner2,
			right: levels.TestFloor, below: levels.TestFloor, want: levels.TestDirtCorner2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// given: one corner tile at (2,2) with the neighbours the rules probe.
			dungeon := passNeighbourhood(test.tile, test.left, test.above, test.right, test.below)

			if test.protected {
				dungeon.Protected.Set(2, 2, true)
			}

			// when: the pass runs over the layer.
			dungeon.FixCornerTilesForTest()

			// then: the corner took (or kept) the shape the rule selects.
			if got := dungeon.Tiles[2][2]; got != test.want {
				t.Errorf("tile = %d, want %d", got, test.want)
			}
		})
	}
}

// TestFixTransparency pins the wall cases that copy a room's transparency sector
// over the second world tile of a widened wall (levels/drlg_l1.cpp:1076-1110).
func TestFixTransparency(t *testing.T) {
	t.Parallel()

	const (
		roomSector = 7
		xx         = 16 + 2*2
		yy         = 16 + 2*2
	)

	tests := []struct {
		name    string
		tile    levels.Tile
		above   levels.Tile
		right   levels.Tile
		wantSet [][2]int
	}{
		{
			name: "H dirt wall copies its sector right", tile: levels.TestDirtHWall,
			wantSet: [][2]int{{xx + 1, yy}, {xx + 1, yy + 1}},
		},
		{
			name: "V dirt wall copies its sector down", tile: levels.TestDirtVWall,
			wantSet: [][2]int{{xx, yy + 1}, {xx + 1, yy + 1}},
		},
		{
			name: "V dirt corner copies three tiles", tile: levels.TestVDirtCorner,
			wantSet: [][2]int{{xx + 1, yy}, {xx, yy + 1}, {xx + 1, yy + 1}},
		},
		{
			name: "H dirt wall end copies when the wall continues above", tile: levels.TestDirtHWallEnd,
			above: levels.TestDirtHWall, wantSet: [][2]int{{xx + 1, yy}, {xx + 1, yy + 1}},
		},
		{
			name: "V dirt wall end copies when the wall continues right", tile: levels.TestDirtVWallEnd,
			right: levels.TestDirtVWall, wantSet: [][2]int{{xx, yy + 1}, {xx + 1, yy + 1}},
		},
		{name: "plain floor leaves the window alone", tile: levels.TestFloor},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// given: one wall tile at (2,2) whose own world tile holds a
			// sector and whose three neighbours hold none.
			dungeon := passDungeon()
			dungeon.Tiles[2][2] = test.tile
			dungeon.Tiles[2][1] = test.above
			dungeon.Tiles[3][2] = test.right
			dungeon.Trans[xx][yy] = roomSector

			// when: the pass runs over the layer.
			dungeon.FixTransparencyForTest()

			// then: exactly the tiles the case names took the sector.
			assertSectorCopied(t, dungeon, test.wantSet, xx, yy, roomSector)
		})
	}
}

// TestFillFloor pins the floor variation: an unprotected plain floor tile takes
// the variant its draw selects and a protected one is never touched
// (levels/drlg_l1.cpp:414-429).
//
//nolint:revive // the draw x tile matrix is the AC; splitting the table would hide which case failed
func TestFillFloor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		draw      int32
		protected bool
		nonFloor  bool
		want      levels.Tile
	}{
		{name: "draw 0 keeps plain floor", draw: 0, want: levels.TestFloor},
		{name: "draw 1 takes the first variant", draw: 1, want: levels.TestFloor22},
		{name: "draw 2 takes the second variant", draw: 2, want: levels.TestFloor23},
		{name: "protected floor is never varied", draw: 2, protected: true, want: levels.TestFloor},
		{name: "non-floor tiles are never varied", draw: 2, nonFloor: true, want: levels.TestVWall},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// given: one tile and a one-draw RNG tape.
			dungeon := passDungeon()
			dungeon.Tiles[2][2] = levels.TestFloor
			if test.nonFloor {
				dungeon.Tiles[2][2] = levels.TestVWall
			}

			if test.protected {
				dungeon.Protected.Set(2, 2, true)
			}

			dungeon.Rnd = scriptedFloorRnd(test.draw)

			// when: the pass runs over the layer.
			dungeon.FillFloorForTest()

			// then: the tile holds the drawn variant.
			if got := dungeon.Tiles[2][2]; got != test.want {
				t.Errorf("tile = %d, want %d", got, test.want)
			}
		})
	}
}

// TestSubstitutionStaysInTheFamily pins the invariant the decoration pass
// guarantees: a substituted tile is always a member of its own original family
// (levels/drlg_l1.cpp:996-1035).
func TestSubstitutionStaysInTheFamily(t *testing.T) {
	t.Parallel()

	// given: a decorated tile and a coin that always substitutes.
	dungeon := passDungeon()
	dungeon.Tiles[5][5] = levels.TestVWall2
	dungeon.Rnd = scriptedSubstitutionRnd(3)

	family := levels.DecorationFamily(levels.TestVWall2)
	if family == 0 {
		t.Fatal("TestVWall2 has no decoration family")
	}

	// when: the pass runs over the layer.
	dungeon.SubstitutionForTest()

	// then: the tile is still a member of that family.
	if got := levels.DecorationFamily(dungeon.Tiles[5][5]); got != family {
		t.Errorf("tile %d has family %d, want %d", dungeon.Tiles[5][5], got, family)
	}
}

// TestCheckQuestsButcher pins the one quest branch that writes generator state:
// the Butcher's room widens by an 8x8 transparency sector
// (levels/drlg_quests.cpp:19-23).
//
//nolint:revive // the quest x sector matrix is the AC; splitting the table would hide which case failed
func TestCheckQuestsButcher(t *testing.T) {
	t.Parallel()

	const (
		setPieceMega = 5
		sector       = int8(7)
	)

	tests := []struct {
		name    string
		quests  levels.Quests
		wantSet bool
	}{
		{name: "Butcher available", quests: levels.Quests{Butcher: true}, wantSet: true},
		{name: "Butcher not available", quests: levels.Quests{}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// given: a placed set piece and an idle transparency counter.
			dungeon := passDungeon()
			dungeon.SetPiece.Position = levels.Point{X: setPieceMega, Y: setPieceMega}
			dungeon.TransVal = sector

			// when: the quest hook runs.
			dungeon.CheckQuestsForTest(test.quests)

			// then: the Butcher's sector is 8x8 at the set piece + (3,3), and the
			// counter only advances when a sector was written.
			originX, originY := 2*setPieceMega+16+3, 2*setPieceMega+16+3
			wantValue := int8(0)
			wantTransVal := sector
			if test.wantSet {
				wantValue = sector
				wantTransVal = sector + 1
			}

			if got := dungeon.Trans[originX][originY]; got != wantValue {
				t.Errorf("trans[%d][%d] = %d, want %d", originX, originY, got, wantValue)
			}

			if got := dungeon.Trans[originX+7][originY+7]; got != wantValue {
				t.Errorf("trans[%d][%d] = %d, want %d", originX+7, originY+7, got, wantValue)
			}

			if got := dungeon.Trans[originX+8][originY+8]; got != 0 {
				t.Errorf("trans[%d][%d] = %d, want the sector to stop at 8x8", originX+8, originY+8, got)
			}

			if dungeon.TransVal != wantTransVal {
				t.Errorf("TransVal = %d, want %d", dungeon.TransVal, wantTransVal)
			}
		})
	}
}

// scriptedFloorRnd is a constant RNG tape for FillFloor.
func scriptedFloorRnd(draw int32) levels.Rnd {
	return levels.Rnd{
		RandomIntLessThan: func(int32) int32 { return draw },
	}
}

// scriptedSubstitutionRnd always substitutes and always draws the same
// decoration index.
func scriptedSubstitutionRnd(draw int32) levels.Rnd {
	return levels.Rnd{
		FlipCoin:    func(uint) bool { return true },
		GenerateRnd: func(int32) int32 { return draw },
	}
}

// passDungeon is a cathedral layer the pass tests drive one pass at a time: it
// starts as solid plain floor so a case only has to state the tiles it cares
// about.
func passDungeon() *levels.Dungeon {
	dungeon := levels.NewDungeon(levels.Cathedral, 1)

	for x := range levels.DMAXX {
		for y := range levels.DMAXY {
			dungeon.Tiles[x][y] = levels.TestFloor
		}
	}

	return dungeon
}

// passNeighbourhood is passDungeon with the tile at (2,2) and its four
// orthogonal neighbours set to the given tiles.
func passNeighbourhood(tile, left, above, right, below levels.Tile) *levels.Dungeon {
	dungeon := passDungeon()
	dungeon.Tiles[2][2] = tile
	dungeon.Tiles[1][2] = left
	dungeon.Tiles[2][1] = above
	dungeon.Tiles[3][2] = right
	dungeon.Tiles[2][3] = below

	return dungeon
}

// assertSectorCopied checks the wall's own world tile, its three neighbours and
// two tiles further out: only the cells the case names carry the sector.
func assertSectorCopied(t *testing.T, dungeon *levels.Dungeon, wantSet [][2]int, xx, yy int, sector int8) {
	t.Helper()

	want := map[[2]int]bool{}
	for _, cell := range wantSet {
		want[cell] = true
	}

	probes := [][2]int{{xx + 1, yy}, {xx, yy + 1}, {xx + 1, yy + 1}, {xx - 2, yy}, {xx, yy - 2}}
	for _, cell := range probes {
		wantValue := int8(0)
		if want[cell] {
			wantValue = sector
		}

		if got := dungeon.Trans[cell[0]][cell[1]]; got != wantValue {
			t.Errorf("trans[%d][%d] = %d, want %d", cell[0], cell[1], got, wantValue)
		}
	}
}
