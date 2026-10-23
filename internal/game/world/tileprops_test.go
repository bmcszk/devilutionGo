// Package world_test ports test/tile_properties_test.cpp 1:1 (the upstream
// gtest cases, same setups and expectations).
package world_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game/world"
)

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

// step is one diagonal hop: from-tile to an adjacent to-tile.
type step struct {
	fromX, fromY int
	toX, toY     int
}

// TestCanStepOpenSpace: every step in open space is free of solid pieces.
func TestCanStepOpenSpace(t *testing.T) {
	steps := []step{
		{fromX: 0, fromY: 0, toX: 1, toY: 1},
		{fromX: 1, fromY: 1, toX: 0, toY: 0},
		{fromX: 1, fromY: 0, toX: 0, toY: 1},
		{fromX: 0, fromY: 1, toX: 1, toY: 0},
	}
	// given
	g := canStepGrid()
	// when / then
	for _, s := range steps {
		if !g.CanStep(s.fromX, s.fromY, s.toX, s.toY) {
			t.Errorf("step (%d, %d) -> (%d, %d): a step in open space is free of solid pieces",
				s.fromX, s.fromY, s.toX, s.toY)
		}
	}
}

// TestCanStepCardinalIgnoresSolids: cardinal steps ignore solid pieces;
// pathing to/from a solid destination is allowed.
func TestCanStepCardinalIgnoresSolids(t *testing.T) {
	steps := []step{
		{fromX: 0, fromY: 1, toX: 1, toY: 0},
		{fromX: 1, fromY: 0, toX: 0, toY: 1},
		{fromX: 0, fromY: 1, toX: 1, toY: 1},
		{fromX: 1, fromY: 0, toX: 1, toY: 1},
		{fromX: 0, fromY: 0, toX: 1, toY: 0},
		{fromX: 1, fromY: 1, toX: 1, toY: 0},
	}
	// given
	g := canStepGrid()
	g.SOLData[1] = world.TileSolid
	g.DPiece[1][0] = 1
	// when / then
	for _, s := range steps {
		if !g.CanStep(s.fromX, s.fromY, s.toX, s.toY) {
			t.Errorf("step (%d, %d) -> (%d, %d): cardinal steps ignore solid pieces",
				s.fromX, s.fromY, s.toX, s.toY)
		}
	}
}

// TestCanStepCornerCuts: diagonal steps refuse to cut solid corners and
// refuse to pass through the boundary between two solid corners.
// solidTile marks one grid cell solid.
type solidTile struct{ x, y int }

// TestCanStepCornerCuts: diagonal steps refuse to cut solid corners and
// refuse to pass through the boundary between two solid corners.
func TestCanStepCornerCuts(t *testing.T) {
	cases := []struct {
		name  string
		solid []solidTile
		steps []step
	}{
		{
			name:  "one solid corner blocks the diagonal",
			solid: []solidTile{{1, 0}},
			steps: []step{{fromX: 0, fromY: 0, toX: 1, toY: 1}, {fromX: 1, fromY: 1, toX: 0, toY: 0}},
		},
		{
			name:  "boundary between two corners blocks",
			solid: []solidTile{{1, 0}, {0, 1}},
			steps: []step{{fromX: 0, fromY: 0, toX: 1, toY: 1}, {fromX: 1, fromY: 1, toX: 0, toY: 0}},
		},
		{
			name:  "still blocked with only the first corner solid",
			solid: []solidTile{{1, 0}},
			steps: []step{{fromX: 0, fromY: 0, toX: 1, toY: 1}, {fromX: 1, fromY: 1, toX: 0, toY: 0}},
		},
		{
			name:  "solid start corner blocks the mirrored diagonal",
			solid: []solidTile{{0, 0}},
			steps: []step{{fromX: 1, fromY: 0, toX: 0, toY: 1}, {fromX: 0, fromY: 1, toX: 1, toY: 0}},
		},
		{
			name:  "solid destination corner blocks with boundary",
			solid: []solidTile{{1, 1}},
			steps: []step{{fromX: 1, fromY: 0, toX: 0, toY: 1}, {fromX: 0, fromY: 1, toX: 1, toY: 0}},
		},
		{
			name:  "solid destination corner alone blocks",
			solid: []solidTile{{1, 1}},
			steps: []step{{fromX: 1, fromY: 0, toX: 0, toY: 1}, {fromX: 0, fromY: 1, toX: 1, toY: 0}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			g := canStepGrid()
			g.SOLData[1] = world.TileSolid
			for _, p := range tc.solid {
				g.DPiece[p.x][p.y] = 1
			}
			// when / then
			for _, s := range tc.steps {
				if g.CanStep(s.fromX, s.fromY, s.toX, s.toY) {
					t.Errorf("step (%d, %d) -> (%d, %d) must not cut the corner",
						s.fromX, s.fromY, s.toX, s.toY)
				}
			}
		})
	}
}

func resetGrid() *world.Grid {
	g := world.NewGrid()
	g.DPiece[5][5] = 0
	g.DPiece[6][6] = 0
	g.SOLData[0] = world.TileNone
	g.SOLData[1] = world.TileNone
	return g
}
