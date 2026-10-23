// Package world_test ports test/tile_properties_test.cpp 1:1 (the upstream
// gtest cases, same setups and expectations).
package world_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game/world"
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
	tests := []struct {
		name       string
		setup      func(*world.Grid)
		x, y       int
		wantSolid  bool
		wantNotSol bool
	}{
		{
			name: "solid in-bounds tile",
			setup: func(g *world.Grid) {
				g.DPiece[5][5] = 0
				g.SOLData[0] = world.TileSolid
			},
			x: 5, y: 5,
			wantSolid:  true,
			wantNotSol: false,
		},
		{
			name: "non-solid in-bounds tile",
			setup: func(g *world.Grid) {
				g.DPiece[6][6] = 1
				g.SOLData[1] = world.TileNone
			},
			x: 6, y: 6,
			wantSolid:  false,
			wantNotSol: true,
		},
		{
			// Upstream quirk: out-of-bounds makes both predicates false.
			name:      "out of bounds tile",
			setup:     func(*world.Grid) {},
			x:         -1,
			y:         1,
			wantSolid: false, wantNotSol: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			g := resetGrid()
			tc.setup(g)
			// when
			gotSolid := g.IsTileSolid(tc.x, tc.y)
			gotNotSolid := g.IsTileNotSolid(tc.x, tc.y)
			// then
			if gotSolid != tc.wantSolid {
				t.Errorf("IsTileSolid(%d, %d) = %v, want %v", tc.x, tc.y, gotSolid, tc.wantSolid)
			}
			if gotNotSolid != tc.wantNotSol {
				t.Errorf("IsTileNotSolid(%d, %d) = %v, want %v", tc.x, tc.y, gotNotSolid, tc.wantNotSol)
			}
		})
	}
}

// TilePropertiesTest.Walkable
func TestWalkable(t *testing.T) {
	tests := []struct {
		name            string
		setup           func(*world.Grid)
		wantWalkable    bool
		wantIgnoreDoors bool
	}{
		{
			name: "solid tile",
			setup: func(g *world.Grid) {
				g.DPiece[5][5] = 0
				g.SOLData[0] = world.TileSolid
			},
			wantWalkable:    false,
			wantIgnoreDoors: false,
		},
		{
			name: "non-solid tile",
			setup: func(g *world.Grid) {
				g.DPiece[5][5] = 0
				g.SOLData[0] = world.TileNone
			},
			wantWalkable:    true,
			wantIgnoreDoors: true,
		},
		{
			name: "tile occupied by a solid non-door object",
			setup: func(g *world.Grid) {
				g.DPiece[5][5] = 0
				g.SOLData[0] = world.TileNone
				g.DObject[5][5] = 1
				g.Objects = append(g.Objects, world.WorldObject{SolidFlag: true})
			},
			wantWalkable:    false,
			wantIgnoreDoors: false,
		},
		{
			name: "tile occupied by a solid door (OBJ_L1LDOOR)",
			setup: func(g *world.Grid) {
				g.DPiece[5][5] = 0
				g.SOLData[0] = world.TileNone
				g.DObject[5][5] = 1
				g.Objects = append(g.Objects, world.WorldObject{SolidFlag: true, IsDoor: true})
			},
			wantWalkable:    false,
			wantIgnoreDoors: true,
		},
		{
			name: "tile occupied by an open door",
			setup: func(g *world.Grid) {
				g.DPiece[5][5] = 0
				g.SOLData[0] = world.TileNone
				g.DObject[5][5] = 1
				g.Objects = append(g.Objects, world.WorldObject{IsDoor: true})
			},
			wantWalkable:    true,
			wantIgnoreDoors: true,
		},
		{
			name: "solid tile occupied by an open door",
			setup: func(g *world.Grid) {
				g.DPiece[5][5] = 0
				g.SOLData[0] = world.TileSolid
				g.DObject[5][5] = 1
				g.Objects = append(g.Objects, world.WorldObject{IsDoor: true})
			},
			wantWalkable:    false,
			wantIgnoreDoors: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			g := resetGrid()
			tc.setup(g)
			// when
			gotWalkable := g.IsTileWalkable(5, 5)
			gotIgnoreDoors := g.IsTileWalkableIgnoreDoors(5, 5)
			// then
			if gotWalkable != tc.wantWalkable {
				t.Errorf("%s: IsTileWalkable = %v, want %v", tc.name, gotWalkable, tc.wantWalkable)
			}
			if gotIgnoreDoors != tc.wantIgnoreDoors {
				t.Errorf("%s: IsTileWalkableIgnoreDoors = %v, want %v", tc.name, gotIgnoreDoors, tc.wantIgnoreDoors)
			}
		})
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
	steps := [][4]int{{0, 0, 1, 1}, {1, 1, 0, 0}, {1, 0, 0, 1}, {0, 1, 1, 0}}
	// given
	g := canStepGrid()
	// when
	got := make([]bool, len(steps))
	for i, s := range steps {
		got[i] = g.CanStep(s[0], s[1], s[2], s[3])
	}
	// then
	for i, s := range steps {
		if !got[i] {
			t.Errorf("step %v: a step in open space is free of solid pieces", s)
		}
	}
}

// TestCanStepCardinalIgnoresSolids: cardinal steps ignore solid pieces;
// pathing to/from a solid destination is allowed.
func TestCanStepCardinalIgnoresSolids(t *testing.T) {
	steps := [][4]int{{0, 1, 1, 0}, {1, 0, 0, 1}, {0, 1, 1, 1}, {1, 0, 1, 1}, {0, 0, 1, 0}, {1, 1, 1, 0}}
	// given
	g := canStepGrid()
	g.SOLData[1] = world.TileSolid
	g.DPiece[1][0] = 1
	// when
	got := make([]bool, len(steps))
	for i, s := range steps {
		got[i] = g.CanStep(s[0], s[1], s[2], s[3])
	}
	// then
	for i, s := range steps {
		if !got[i] {
			t.Errorf("step %v: cardinal steps ignore solid pieces", s)
		}
	}
}

// TestCanStepCornerCuts: diagonal steps refuse to cut solid corners and
// refuse to pass through the boundary between two solid corners.
func TestCanStepCornerCuts(t *testing.T) {
	cases := []struct {
		name  string
		solid [][2]int
		steps [][4]int
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
		t.Run(tc.name, func(t *testing.T) {
			// given
			g := canStepGrid()
			g.SOLData[1] = world.TileSolid
			for _, p := range tc.solid {
				g.DPiece[p[0]][p[1]] = 1
			}
			// when
			got := make([]bool, len(tc.steps))
			for i, s := range tc.steps {
				got[i] = g.CanStep(s[0], s[1], s[2], s[3])
			}
			// then
			for i, s := range tc.steps {
				if got[i] {
					t.Errorf("%s: step %v must not cut the corner", tc.name, s)
				}
			}
		})
	}
}
