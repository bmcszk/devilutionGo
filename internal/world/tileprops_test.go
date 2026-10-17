// Package world_test ports test/tile_properties_test.cpp 1:1 (the upstream
// gtest cases, same setups and expectations).
package world_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/world"
)

func resetGrid() *world.Grid {
	g := world.NewGrid()
	g.DPiece[5][5] = 0
	g.DPiece[6][6] = 0
	g.SOLData[0] = world.TileNone
	g.SOLData[1] = world.TileNone
	return g
}

// TilePropertiesTest.Solid
func TestSolid(t *testing.T) {
	g := resetGrid()
	g.DPiece[5][5] = 0
	g.SOLData[0] = world.TileSolid
	if !g.IsTileSolid(5, 5) {
		t.Error("Solid in-bounds tiles are solid")
	}
	if g.IsTileNotSolid(5, 5) {
		t.Error("IsTileNotSolid returns the inverse of IsTileSolid for in-bounds tiles")
	}
	g.DPiece[6][6] = 1
	g.SOLData[1] = world.TileNone
	if g.IsTileSolid(6, 6) {
		t.Error("Non-solid in-bounds tiles are not solid")
	}
	if !g.IsTileNotSolid(6, 6) {
		t.Error("IsTileNotSolid returns the inverse of IsTileSolid for in-bounds tiles")
	}
	if g.IsTileSolid(-1, 1) {
		t.Error("Out of bounds tiles are not solid")
	}
	if g.IsTileNotSolid(-1, 1) {
		t.Error("Out of bounds tiles are also not not solid")
	}
}

// TilePropertiesTest.Walkable
func TestWalkable(t *testing.T) {
	g := resetGrid()
	g.DPiece[5][5] = 0
	g.SOLData[0] = world.TileSolid
	if g.IsTileWalkable(5, 5) {
		t.Error("Tile which is marked as solid should be considered blocked")
	}
	if g.IsTileWalkableIgnoreDoors(5, 5) {
		t.Error("Solid non-door tiles remain unwalkable when ignoring doors")
	}
	g.SOLData[0] = world.TileNone
	if !g.IsTileWalkable(5, 5) {
		t.Error("Non-solid tiles are walkable")
	}
	if !g.IsTileWalkableIgnoreDoors(5, 5) {
		t.Error("Non-solid tiles remain walkable when ignoring doors")
	}
	g.DObject[5][5] = 1
	g.Objects = append(g.Objects, world.WorldObject{})
	g.Objects[0].SolidFlag = true
	if g.IsTileWalkable(5, 5) {
		t.Error("Tile occupied by a solid object is unwalkable")
	}
	if g.IsTileWalkableIgnoreDoors(5, 5) {
		t.Error("Tile occupied by a solid non-door object are unwalkable when ignoring doors")
	}
	g.Objects[0].IsDoor = true // OBJ_L1LDOOR
	if g.IsTileWalkable(5, 5) {
		t.Error("Tile occupied by a door which is marked as solid should be considered blocked")
	}
	if !g.IsTileWalkableIgnoreDoors(5, 5) {
		t.Error("Tile occupied by a door is considered walkable when ignoring doors")
	}
	g.Objects[0].SolidFlag = false
	if !g.IsTileWalkable(5, 5) {
		t.Error("Tile occupied by an open door is walkable")
	}
	if !g.IsTileWalkableIgnoreDoors(5, 5) {
		t.Error("Tile occupied by a door is considered walkable when ignoring doors")
	}
	g.SOLData[0] = world.TileSolid
	if g.IsTileWalkable(5, 5) {
		t.Error("Solid tiles occupied by an open door remain unwalkable")
	}
	if !g.IsTileWalkableIgnoreDoors(5, 5) {
		t.Error("Solid tiles occupied by an open door become walkable when ignoring doors")
	}
}

// canStepGrid builds the 2x2 open-space fixture.
func canStepGrid() *world.Grid {
	g := world.NewGrid()
	g.DPiece[0][0] = 0
	g.DPiece[0][1] = 0
	g.DPiece[1][0] = 0
	g.DPiece[1][1] = 0
	g.SOLData[0] = world.TileNone
	return g
}

// TestCanStepOpenSpace: every step in open space is free of solid pieces.
func TestCanStepOpenSpace(t *testing.T) {
	g := canStepGrid()
	for _, s := range [][4]int{{0, 0, 1, 1}, {1, 1, 0, 0}, {1, 0, 0, 1}, {0, 1, 1, 0}} {
		if !g.CanStep(s[0], s[1], s[2], s[3]) {
			t.Errorf("step %v: a step in open space is free of solid pieces", s)
		}
	}
}

// TestCanStepCardinalIgnoresSolids: cardinal steps ignore solid pieces;
// pathing to/from a solid destination is allowed.
func TestCanStepCardinalIgnoresSolids(t *testing.T) {
	g := canStepGrid()
	g.SOLData[1] = world.TileSolid
	g.DPiece[1][0] = 1
	for _, s := range [][4]int{{0, 1, 1, 0}, {1, 0, 0, 1}, {0, 1, 1, 1}, {1, 0, 1, 1}, {0, 0, 1, 0}, {1, 1, 1, 0}} {
		if !g.CanStep(s[0], s[1], s[2], s[3]) {
			t.Errorf("step %v: cardinal steps ignore solid pieces", s)
		}
	}
}

// TestCanStepCornerCuts: diagonal steps refuse to cut solid corners and
// refuse to pass through the boundary between two solid corners.
func TestCanStepCornerCuts(t *testing.T) {
	cases := []struct {
		name   string
		solid  [][2]int
		steps  [][4]int
		wantOk bool
	}{
		{
			name:  "one solid corner blocks the diagonal",
			solid: [][2]int{{1, 0}},
			steps: [][4]int{{0, 0, 1, 1}, {1, 1, 0, 0}},
		},
		{
			name:  "boundary between two corners blocks",
			solid: [][2]int{{1, 0}, {0, 1}},
			steps: [][4]int{{0, 0, 1, 1}, {1, 1, 0, 0}},
		},
		{
			name:  "still blocked with only the first corner solid",
			solid: [][2]int{{1, 0}},
			steps: [][4]int{{0, 0, 1, 1}, {1, 1, 0, 0}},
		},
		{
			name:  "solid start corner blocks the mirrored diagonal",
			solid: [][2]int{{0, 0}},
			steps: [][4]int{{1, 0, 0, 1}, {0, 1, 1, 0}},
		},
		{
			name:  "solid destination corner blocks with boundary",
			solid: [][2]int{{1, 1}},
			steps: [][4]int{{1, 0, 0, 1}, {0, 1, 1, 0}},
		},
		{
			name:  "solid destination corner alone blocks",
			solid: [][2]int{{1, 1}},
			steps: [][4]int{{1, 0, 0, 1}, {0, 1, 1, 0}},
		},
	}
	for _, tc := range cases {
		g := canStepGrid()
		g.SOLData[1] = world.TileSolid
		for _, p := range tc.solid {
			g.DPiece[p[0]][p[1]] = 1
		}
		for _, s := range tc.steps {
			if g.CanStep(s[0], s[1], s[2], s[3]) {
				t.Errorf("%s: step %v must not cut the corner", tc.name, s)
			}
		}
	}
}
