package levels_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/levels"
)

// scriptedMega is one hand-built pMegaTiles entry: the mega-tile layer cell it
// is named by, its index in the table, and the one micro slot it fills.
// specialCell is one expected dSpecial marker at a world-tile position.
type specialCell struct {
	name     string
	x, y     int
	wantMark int8
}

type scriptedMega struct {
	tile  [2]int
	index int
	slot  uint8 // 1-4 == micro1-micro4 (levels/gendung.cpp:600-603)
	micro uint16
}

// TestInitDungeonPiecesArchLastWriteWins pins the iteration order of the
// catacombs arch pass (levels/drlg_l2.cpp:1678-1688): the loops run y-outer, so
// when a 131 arch and a 134 arch claim one cell the 134 marker wins because its
// own cell sits on the later row.
func TestInitDungeonPiecesArchLastWriteWins(t *testing.T) {
	t.Parallel()

	// when: InitDungeonPieces runs over a grid holding a 131 at (20,20) and a
	// 134 at (19,21), both of which claim dSpecial[20][21].
	grid := expandScripted(t, levels.Catacombs, levels.MegaTile{}, []scriptedMega{
		{tile: [2]int{2, 2}, index: 0, slot: 1, micro: 131},
		{tile: [2]int{1, 2}, index: 1, slot: 4, micro: 134},
	})

	// then: upstream's y-outer order leaves the 134 marker in place.
	assertSpecialCells(t, grid, []specialCell{
		{name: "134 claims the contested cell", x: 20, y: 21, wantMark: 3},
		{name: "131 reserves its second cell", x: 20, y: 22, wantMark: 1},
		{name: "134 reserves its second cell", x: 21, y: 21, wantMark: 4},
	})
}

// assertSpecialCells checks each expected dSpecial marker.
func assertSpecialCells(t *testing.T, grid levels.PieceGrid, cells []specialCell) {
	t.Helper()

	for _, c := range cells {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if got := grid.Special[c.x][c.y]; got != c.wantMark {
				t.Errorf("dSpecial[%d][%d] = %d, want %d", c.x, c.y, got, c.wantMark)
			}
		})
	}
}

// expandScripted runs ExpandPieces (levels/gendung.cpp:591-624) over a
// hand-built mega-tile layer: dirt holds the dirt entry at the level family's
// own index, and every other entry is zeros except the scripted ones.
func expandScripted(
	t *testing.T, levelType levels.LevelType, dirt levels.MegaTile, scripted []scriptedMega,
) levels.PieceGrid {
	t.Helper()

	megaTiles := make([]levels.MegaTile, levels.DirtMegaTileIndexForTest(levelType)+1)
	megaTiles[len(megaTiles)-1] = dirt
	d := &levels.Dungeon{LevelType: levelType}
	for _, s := range scripted {
		d.Tiles[s.tile[0]][s.tile[1]] = levels.Tile(s.index + 1)
		assignMicro(&megaTiles[s.index], s.slot, s.micro)
	}

	return levels.ExpandPieces(d, megaTiles)
}

// assignMicro writes micro into the MegaTile slot named by slot.
func assignMicro(mega *levels.MegaTile, slot uint8, micro uint16) {
	switch slot {
	case 1:
		mega.Micro1 = micro
	case 2:
		mega.Micro2 = micro
	case 3:
		mega.Micro3 = micro
	case 4:
		mega.Micro4 = micro
	default:
		panic("assignMicro: slot must be 1-4")
	}
}
