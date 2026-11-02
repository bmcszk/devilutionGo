// Package missiles_test pins the Hellfire Warp add body (missiles.cpp:
// 1507-1580).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
)

// warpSpy scripts the trigger table and the walkability answers.
type warpSpy struct {
	messages []int
	tiles    []missiles.TilePosition
	level    int
	setLevel bool
	setNum   int
	okTiles  map[missiles.TilePosition]bool
}

// callbacks wires the spy into a HitWorld.
func (s *warpSpy) callbacks() missiles.ProcessCallbacks {
	cb := missiles.NoopProcessCallbacks()
	cb.LevelType = s.level
	cb.IsSetLevel = func() bool { return s.setLevel }
	cb.SetLevelNum = func() int { return s.setNum }
	cb.TriggerCount = func() int { return len(s.messages) }
	cb.TriggerMessage = func(i int) int { return s.messages[i] }
	cb.TriggerPosition = func(i int) missiles.TilePosition { return s.tiles[i] }
	cb.PosOkPlayer = func(_, x, y int) bool { return s.okTiles[missiles.TilePosition{X: x, Y: y}] }

	return cb
}

func TestAddWarp_TeleportsInFrontOfTheClosestStair(t *testing.T) {
	// given
	spy := &warpSpy{
		messages: []int{0, 1}, // WM_DIABNEXTLVL, WM_DIABPREVLVL
		tiles:    []missiles.TilePosition{{X: 50, Y: 50}, {X: 31, Y: 40}},
		level:    3, // DTYPE_CAVES: next → +1 x, prev → +1 y
		okTiles:  map[missiles.TilePosition]bool{{X: 31, Y: 41}: true},
	}
	fx := newAddFixture(t, addScript{
		start: [2]int{40, 40}, dst: [2]int{40, 40}, source: 0, tile: [2]int{40, 40},
	})
	fx.world.ProcessCallbacks = spy.callbacks()

	// when
	missiles.AddWarp(fx.world, fx.missile, fx.param)

	// then
	if fx.missile.Position.Tile != (missiles.TilePosition{X: 31, Y: 41}) {
		t.Fatalf("tile = %v, want the closest stair tile 31/41", fx.missile.Position.Tile)
	}

	if fx.missile.Duration != 2 || fx.missile.DelFlag || fx.param.SpellFizzled {
		t.Fatalf("duration/delFlag/fizzled = %d/%v/%v, want 2/false/false",
			fx.missile.Duration, fx.missile.DelFlag, fx.param.SpellFizzled)
	}
}

func TestAddWarp_FizzlesWithoutAFreeDestination(t *testing.T) {
	// given
	spy := &warpSpy{
		messages: []int{0},
		tiles:    []missiles.TilePosition{{X: 50, Y: 50}},
		level:    3,
		okTiles:  map[missiles.TilePosition]bool{},
	}
	fx := newAddFixture(t, addScript{
		start: [2]int{40, 40}, dst: [2]int{40, 40}, source: 0, tile: [2]int{40, 40},
	})
	fx.world.ProcessCallbacks = spy.callbacks()

	// when
	missiles.AddWarp(fx.world, fx.missile, fx.param)

	// then
	if !fx.missile.DelFlag || !fx.param.SpellFizzled {
		t.Fatalf("delFlag/fizzled = %v/%v, want true/true", fx.missile.DelFlag, fx.param.SpellFizzled)
	}
}

func TestAddWarp_SkipsStairMessagesAndTriggerTiles(t *testing.T) {
	// given
	spy := &warpSpy{
		// WM_DIABSETLVL is not a stair message, and the only walkable tile is
		// the trigger tile itself, which the placement predicate rejects.
		messages: []int{3, 0},
		tiles:    []missiles.TilePosition{{X: 31, Y: 41}, {X: 50, Y: 50}},
		level:    3,
		okTiles:  map[missiles.TilePosition]bool{{X: 50, Y: 51}: false, {X: 51, Y: 50}: false},
	}
	fx := newAddFixture(t, addScript{
		start: [2]int{40, 40}, dst: [2]int{40, 40}, source: 0, tile: [2]int{40, 40},
	})
	fx.world.ProcessCallbacks = spy.callbacks()

	// when
	missiles.AddWarp(fx.world, fx.missile, fx.param)

	// then
	if !fx.missile.DelFlag {
		t.Fatal("warp row not deleted when no destination exists")
	}
}

func TestAddWarp_UsesTheDungeonTypeStairOffset(t *testing.T) {
	// given
	cases := []struct {
		name     string
		level    int
		message  int
		wantTile missiles.TilePosition
	}{
		{"cathedral next-level", 1, 0, missiles.TilePosition{X: 50, Y: 51}},
		{"cathedral previous-level", 1, 1, missiles.TilePosition{X: 51, Y: 52}},
		{"catacombs next-level", 2, 0, missiles.TilePosition{X: 50, Y: 51}},
		{"caves previous-level", 3, 1, missiles.TilePosition{X: 50, Y: 51}},
		{"hell ignores the message", 4, 2, missiles.TilePosition{X: 51, Y: 50}},
		{"nest return-level", 5, 2, missiles.TilePosition{X: 50, Y: 51}},
		{"crypt next-level", 6, 0, missiles.TilePosition{X: 50, Y: 51}},
		{"crypt previous-level", 6, 1, missiles.TilePosition{X: 51, Y: 51}},
	}

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()
			runWarpOffset(t, tc.level, tc.message, tc.wantTile)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// runWarpOffset walks one stair message through the offset table.
func runWarpOffset(t *testing.T, level, message int, wantTile missiles.TilePosition) {
	t.Helper()

	// given
	spy := &warpSpy{
		messages: []int{message}, tiles: []missiles.TilePosition{{X: 50, Y: 50}},
		level: level, okTiles: map[missiles.TilePosition]bool{wantTile: true},
	}
	fx := newAddFixture(t, addScript{
		start: [2]int{40, 40}, dst: [2]int{40, 40}, source: 0, tile: [2]int{40, 40},
	})
	fx.world.ProcessCallbacks = spy.callbacks()

	// when
	missiles.AddWarp(fx.world, fx.missile, fx.param)

	// then
	if fx.missile.Position.Tile != wantTile {
		t.Fatalf("tile = %v, want %v", fx.missile.Position.Tile, wantTile)
	}
}
