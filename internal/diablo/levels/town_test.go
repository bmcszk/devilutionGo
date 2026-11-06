package levels_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/levels"
)

// TestCreateTownHellfireDoors covers the hive and grave door patches: the
// committed town goldens run with gbIsHellfire off, so nothing else exercises
// the four door tables (levels/town.cpp:228-238, 281-360).
func TestCreateTownHellfireDoors(t *testing.T) {
	files, missing := townDunFiles(t)
	if len(missing) > 0 {
		t.Fatalf("missing town .dun fixtures %v", missing)
	}

	megaTiles := loadMegaTiles(t, "town")

	cases := []struct {
		name  string
		warps map[levels.LevelType]bool
		want  map[[2]int]uint16
		tiles map[[2]int]levels.Tile
	}{
		{
			name:  "closed",
			warps: nil,
			want:  map[[2]int]uint16{{78, 60}: 1161, {79, 60}: 1258, {36, 21}: 1322},
			tiles: map[[2]int]levels.Tile{{35, 27}: 18, {36, 27}: 63},
		},
		{
			name:  "open",
			warps: map[levels.LevelType]bool{levels.Nest: true, levels.Crypt: true},
			want:  map[[2]int]uint16{{78, 60}: 1161, {79, 60}: 1162, {36, 21}: 1330},
			tiles: map[[2]int]levels.Tile{{36, 27}: 47, {14, 8}: 47},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			// given: the committed town inputs and a pinned seed.
			seed := uint32(2588)
			warps := test.warps

			// when: the town runs with the Hellfire doors in this state.
			dungeon, grid := runTownWithDoors(t, files, megaTiles, warps, seed)

			// then: the door cells and the automap tiles match upstream.
			checkDoorCells(t, grid, dungeon, test.want, test.tiles)
		})
	}
}

// runTownWithDoors runs the town generator once with the Hellfire doors in the
// given state.
func runTownWithDoors(t *testing.T, files map[string][]byte, megaTiles []levels.MegaTile,
	warps map[levels.LevelType]bool, seed uint32,
) (*levels.Dungeon, levels.PieceGrid) {
	t.Helper()

	dungeon := levels.NewDungeon(levels.Town, 0)
	dungeon.Rnd.Seed(seed)

	grid := levels.CreateTown(dungeon, levels.TownOptions{
		Files:     files,
		MegaTiles: megaTiles,
		Warps:     warps,
		Hellfire:  true,
	})

	return dungeon, grid
}

// checkDoorCells compares the door dPiece cells and automap tiles.
func checkDoorCells(t *testing.T, grid levels.PieceGrid, dungeon *levels.Dungeon,
	pieces map[[2]int]uint16, tiles map[[2]int]levels.Tile,
) {
	t.Helper()

	for cell, want := range pieces {
		if got := grid.Piece[cell[0]][cell[1]]; got != want {
			t.Fatalf("dPiece (%d,%d) = %d, want %d", cell[0], cell[1], got, want)
		}
	}

	for cell, want := range tiles {
		if got := dungeon.Tiles[cell[0]][cell[1]]; got != want {
			t.Fatalf("dungeon tile (%d,%d) = %d, want %d", cell[0], cell[1], got, want)
		}
	}
}
