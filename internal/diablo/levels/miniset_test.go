package levels_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/levels"
)

// TestCanReplaceTile pins the neighbour screen PlaceMiniSetRandom applies to
// the cathedral wall variants. The four probes are the isometric displacements
// of Direction::{NorthWest,SouthEast,SouthWest,NorthEast}
// (Source/engine/displacement.hpp:257-263), not the diagonals of the point
// grid, and the partner cell is always the NorthWest one
// (levels/drlg_l1.cpp:397-413).
//
//nolint:revive,funlen // the case table is the AC; splitting it would hide which screen case failed
func TestCanReplaceTile(t *testing.T) {
	t.Parallel()

	// The screen only runs for replace ids in [VWallEnd2, VWall8].
	tests := []struct {
		name    string
		replace levels.Tile
		walls   []levels.Point
		want    bool
	}{
		{
			name:    "replace id outside the range is never screened",
			replace: 13,
			walls:   []levels.Point{{X: 6, Y: 5}},
			want:    true,
		},
		{
			name:    "east wall blocks",
			replace: 95,
			walls:   []levels.Point{{X: 6, Y: 5}},
			want:    false,
		},
		{
			name:    "south wall blocks",
			replace: 95,
			walls:   []levels.Point{{X: 5, Y: 6}},
			want:    false,
		},
		{
			name:    "north wall blocks",
			replace: 95,
			walls:   []levels.Point{{X: 5, Y: 4}},
			want:    false,
		},
		{
			name:    "west wall is both probes and blocks",
			replace: 95,
			walls:   []levels.Point{{X: 4, Y: 5}},
			want:    false,
		},
		{
			name:    "diagonal walls are not probed",
			replace: 95,
			walls:   []levels.Point{{X: 4, Y: 4}, {X: 6, Y: 6}, {X: 4, Y: 6}, {X: 6, Y: 4}},
			want:    true,
		},
		{
			name:    "west partner above the range still blocks through the south probe",
			replace: 95,
			walls:   []levels.Point{{X: 6, Y: 5}, {X: 4, Y: 5}},
			want:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// given: a delta-free layer with only the named wall neighbours set.
			dungeon := levels.NewDungeon(levels.Cathedral, 1)
			for _, wall := range test.walls {
				dungeon.Tiles[wall.X][wall.Y] = 90
			}

			// when: the screen runs at the centre of the layer.
			got := levels.CanReplaceTile(dungeon, test.replace, levels.Point{X: 5, Y: 5})

			// then
			if got != test.want {
				t.Fatalf("CanReplaceTile = %v, want %v", got, test.want)
			}
		})
	}
}
