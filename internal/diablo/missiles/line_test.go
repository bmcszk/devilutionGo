// Package missiles_test pins the LineClear walk (monster.cpp:1599-1670) and
// LineClearMissile's tile predicate (monster.cpp:4393-4396, 3281-3284).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// lineCases pin the two Bresenham arms by hand: the walk reports true only
// when it reaches the end tile without a blocked tile in the way
// (monster.cpp:1670).
var lineCases = []lineCase{
	{name: "clear horizontal", end: missiles.TilePosition{X: 4, Y: 0}, want: true},
	{
		name: "blocked horizontal", end: missiles.TilePosition{X: 4, Y: 0},
		blocked: missiles.TilePosition{X: 2, Y: 0}, want: false,
	},
	{name: "clear vertical", end: missiles.TilePosition{X: 0, Y: 4}, want: true},
	{
		name: "blocked vertical", end: missiles.TilePosition{X: 0, Y: 4},
		blocked: missiles.TilePosition{X: 0, Y: 3}, want: false,
	},
	{name: "clear diagonal", end: missiles.TilePosition{X: 4, Y: 4}, want: true},
	{
		name: "blocked diagonal", end: missiles.TilePosition{X: 4, Y: 4},
		blocked: missiles.TilePosition{X: 2, Y: 2}, want: false,
	},
	{
		name: "reversed horizontal", end: missiles.TilePosition{X: -4, Y: 0},
		blocked: missiles.TilePosition{X: -2, Y: 0}, want: false,
	},
	{
		name: "blocked end tile still reaches", end: missiles.TilePosition{X: 3, Y: 0},
		blocked: missiles.TilePosition{X: 3, Y: 0}, want: true,
	},
	{
		name: "start tile is never tested", end: missiles.TilePosition{X: 3, Y: 0},
		blocked: missiles.TilePosition{X: 0, Y: 0}, want: true,
	},
}

// lineCase pins one walk from the origin.
type lineCase struct {
	name    string
	end     missiles.TilePosition
	blocked missiles.TilePosition
	want    bool
}

// TestLineClear pins the walk over a single blocked tile.
func TestLineClear(t *testing.T) {
	// given
	start := missiles.TilePosition{X: 0, Y: 0}

	// when
	subtests := make([]subtest, 0, len(lineCases))
	for _, tc := range lineCases {
		subtests = append(subtests, subtest{tc.name, runLineCase(start, tc)})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// TestLineClearMissile pins the tile predicate: only a BlockMissile tile
// stops the walk (monster.cpp:3281-3284, 4393-4396), a Solid-only tile does
// not.
func TestLineClearMissile(t *testing.T) {
	// given
	w := newSilentWorld()
	w.Grid.DPiece[0][2] = 1378
	w.Grid.SOLData[1378] = world.TileSolid
	w.Grid.DPiece[2][2] = 1377
	w.Grid.SOLData[1377] = world.TileBlockMissile | world.TileSolid
	start := missiles.TilePosition{X: 0, Y: 0}

	// when
	overSolid := missiles.LineClearMissile(w, start, missiles.TilePosition{X: 4, Y: 0})
	overBlocking := missiles.LineClearMissile(w, start, missiles.TilePosition{X: 4, Y: 4})

	// then
	if !overSolid {
		t.Fatal("a Solid-only tile stopped the missile walk, want it ignored")
	}

	if overBlocking {
		t.Fatal("a BlockMissile tile did not stop the missile walk")
	}
}

// runLineCase drives one hand-pinned walk over a single blocked tile.
func runLineCase(start missiles.TilePosition, tc lineCase) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		// given
		isClear := func(x, y int) bool { return (missiles.TilePosition{X: x, Y: y}) != tc.blocked }

		// when
		got := missiles.LineClear(isClear, start, tc.end)

		// then
		if got != tc.want {
			t.Fatalf("LineClear to %v = %v, want %v", tc.end, got, tc.want)
		}
	}
}
