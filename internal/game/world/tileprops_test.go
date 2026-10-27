// Package world_test ports test/tile_properties_test.cpp 1:1 (the upstream
// gtest cases, same setups and expectations).
package world_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// solidExpectation carries the two predicate expectations for one tile.
// walkableCases is the world.IsWalkable table: solid piece, non-solid,
// solid non-door object, solid door, open door, and a door over a solid tile.
var walkableCases = []walkableCase{
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

type solidExpectation struct {
	x, y       int
	wantSolid  bool
	wantNotSol bool
}

// walkExpectation carries the two predicate expectations for one tile.
type walkExpectation struct {
	wantWalkable    bool
	wantIgnoreDoors bool
}

// step is one diagonal hop: from-tile to an adjacent to-tile.
type step struct {
	fromX, fromY int
	toX, toY     int
}

// TestCanStepCornerCuts: diagonal steps refuse to cut solid corners and
// refuse to pass through the boundary between two solid corners.
// solidTile marks one grid cell solid.
type solidTile struct{ x, y int }

// TilePropertiesTest.Solid
// walkableCase is one world.IsWalkable expectation.
type walkableCase struct {
	name  string
	setup func(*world.Grid)
	walkExpectation
}

func TestSolid(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*world.Grid)
		solidExpectation
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
	// when every case builds its grid
	grids := make([]*world.Grid, len(tests))
	for i, tc := range tests {
		grids[i] = resetGrid()
		tc.setup(grids[i])
	}

	// then each grid reports the expected solidity predicates
	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assertSolidPredicates(t, tc.solidExpectation, grids[i])
		})
	}
}

// TilePropertiesTest.Walkable
func TestWalkable(t *testing.T) {
	// when every case builds its grid
	grids := make([]*world.Grid, len(walkableCases))
	for i, tc := range walkableCases {
		grids[i] = resetGrid()
		tc.setup(grids[i])
	}

	// then each grid reports the expected walkability predicates
	for i, tc := range walkableCases {
		t.Run(tc.name, func(t *testing.T) {
			assertWalkablePredicates(t, tc.name, grids[i], walkExpectation{tc.wantWalkable, tc.wantIgnoreDoors})
		})
	}
}

// TestCanStepOpenSpace: every step in open space is free of solid pieces.
func TestCanStepOpenSpace(t *testing.T) {
	steps := []step{
		{fromX: 0, fromY: 0, toX: 1, toY: 1},
		{fromX: 1, fromY: 1, toX: 0, toY: 0},
		{fromX: 1, fromY: 0, toX: 0, toY: 1},
		{fromX: 0, fromY: 1, toX: 1, toY: 0},
	}
	// given every direction in open space
	g := canStepGrid()

	// when every step is tested
	allowed := make([]bool, len(steps))
	for i, s := range steps {
		allowed[i] = g.CanStep(s.fromX, s.fromY, s.toX, s.toY)
	}

	// then each step is free of solid pieces
	for i, s := range steps {
		if !allowed[i] {
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
	// given a 2x2 grid whose solid data and solid piece sit on the path
	g := canStepGrid()
	g.SOLData[1] = world.TileSolid
	g.DPiece[1][0] = 1

	// when every cardinal step is tested
	allowed := make([]bool, len(steps))
	for i, s := range steps {
		allowed[i] = g.CanStep(s.fromX, s.fromY, s.toX, s.toY)
	}

	// then every step is allowed despite the solid corners
	for i, s := range steps {
		if !allowed[i] {
			t.Errorf("step (%d, %d) -> (%d, %d): cardinal steps ignore solid pieces",
				s.fromX, s.fromY, s.toX, s.toY)
		}
	}
}

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
	// when every case builds its solid layout and tests its steps
	allowed := make([][]bool, len(cases))
	for i, tc := range cases {
		g := canStepGrid()

		g.SOLData[1] = world.TileSolid
		for _, p := range tc.solid {
			g.DPiece[p.x][p.y] = 1
		}

		allowed[i] = make([]bool, len(tc.steps))
		for j, s := range tc.steps {
			allowed[i][j] = g.CanStep(s.fromX, s.fromY, s.toX, s.toY)
		}
	}

	// then no step cuts a solid corner
	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertBlockedSteps(t, tc.steps, allowed[i])
		})
	}
}

// assertSolidPredicates pins IsTileSolid and IsTileNotSolid for one case.
func assertSolidPredicates(t *testing.T, tc solidExpectation, g *world.Grid) {
	t.Helper()

	gotSolid := g.IsTileSolid(tc.x, tc.y)

	gotNotSolid := g.IsTileNotSolid(tc.x, tc.y)
	if gotSolid != tc.wantSolid {
		t.Errorf("IsTileSolid(%d, %d) = %v, want %v", tc.x, tc.y, gotSolid, tc.wantSolid)
	}

	if gotNotSolid != tc.wantNotSol {
		t.Errorf("IsTileNotSolid(%d, %d) = %v, want %v", tc.x, tc.y, gotNotSolid, tc.wantNotSol)
	}
}

// assertWalkablePredicates pins IsTileWalkable and IsTileWalkableIgnoreDoors
// at (5,5).
func assertWalkablePredicates(t *testing.T, name string, g *world.Grid, want walkExpectation) {
	t.Helper()

	gotWalkable := g.IsTileWalkable(5, 5)
	gotIgnoreDoors := g.IsTileWalkableIgnoreDoors(5, 5)

	if gotWalkable != want.wantWalkable {
		t.Errorf("%s: IsTileWalkable = %v, want %v", name, gotWalkable, want.wantWalkable)
	}

	if gotIgnoreDoors != want.wantIgnoreDoors {
		t.Errorf("%s: IsTileWalkableIgnoreDoors = %v, want %v", name, gotIgnoreDoors, want.wantIgnoreDoors)
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

// assertBlockedSteps checks every step is rejected by CanStep.
func assertBlockedSteps(t *testing.T, steps []step, allowed []bool) {
	t.Helper()

	for i, s := range steps {
		if allowed[i] {
			t.Errorf("step (%d, %d) -> (%d, %d) must not cut the corner",
				s.fromX, s.fromY, s.toX, s.toY)
		}
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
