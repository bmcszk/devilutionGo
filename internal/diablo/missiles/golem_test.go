// Package missiles_test pins AddGolem (missiles.cpp:2620-2660), the add-only
// row that re-casts the spell as a golem kill and otherwise asks for a spawn
// tile, an owner check and a SpawnGolem (or the multiplayer request packet).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// golemCases pin the four exits of AddGolem: the kill of a live golem
// (cpp:2626-2630), the tile search failing, the owner-only spawn
// (cpp:2640-2658) and the non-owner's request packet (cpp:2650-2655).
var golemCases = []golemCase{
	{name: "a live golem is killed instead", hasGolem: true, wantKills: 1},
	{name: "no free tile leaves nothing behind", wantDel: true},
	{name: "the owner spawns the golem", mine: true, owns: true, wantSpawns: 1},
	{name: "the non-owner asks the owner", mine: true, wantRequests: 1},
}

// golemCase pins one AddGolem cast.
type golemCase struct {
	name         string
	hasGolem     bool
	mine         bool
	owns         bool
	wantKills    int
	wantSpawns   int
	wantRequests int
	wantDel      bool
}

// golemSpy records what the add body asked the rest of the game to do.
type golemSpy struct {
	kills    int
	spawns   []missiles.TilePosition
	spellLvl int
	requests []missiles.TilePosition
}

// TestAddGolem pins the kill / spawn / request exits.
func TestAddGolem(t *testing.T) {
	// when
	subtests := make([]subtest, 0, len(golemCases))
	for _, tc := range golemCases {
		subtests = append(subtests, subtest{tc.name, runGolemCast(tc)})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// golemWorld wires the search predicate inputs, the player gate and the golem
// sinks.
func golemWorld(tc golemCase, spy *golemSpy) *missiles.HitWorld {
	cb := missiles.NoopProcessCallbacks()
	cb.IsMyPlayer = func(int) bool { return tc.mine }
	cb.FindGolemForPlayer = func(int) bool { return tc.hasGolem }
	cb.KillGolem = func(int) { spy.kills++ }
	cb.IsLevelOwnedByLocalClient = func(int) bool { return tc.owns }
	cb.SpawnGolem = func(_ int, position missiles.TilePosition, spellLevel int) {
		spy.spawns = append(spy.spawns, position)
		spy.spellLvl = spellLevel
	}
	cb.RequestSpawnGolem = func(position missiles.TilePosition) { spy.requests = append(spy.requests, position) }
	cb.IsTileOccupied = func(int, int) bool { return false }
	cb.CheckBlock = func(int, int, int, int) bool { return false }

	return &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: cb,
	}
}

// runGolemCast drives one AddGolem.
func runGolemCast(tc golemCase) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		// given
		spy := &golemSpy{}
		w := golemWorld(tc, spy)
		if !tc.hasGolem && !tc.mine {
			// keep the tile search honest: nothing free anywhere
			w.ProcessCallbacks.IsTileOccupied = func(int, int) bool { return true }
		}

		m := missiles.Missile{Type: missiles.Golem, Source: 0, SpellLevel: 3}
		m.SetTraveledFromTile(10, 10)
		m.Position.Start = missiles.TilePosition{X: 10, Y: 10}

		// when
		missiles.AddGolem(w, &m, &missiles.AddParameter{Dst: missiles.TilePosition{X: 10, Y: 12}})

		// then
		assertGolemCast(t, m, spy, tc)
	}
}

// assertGolemCast checks the exits of one cast.
func assertGolemCast(t *testing.T, m missiles.Missile, spy *golemSpy, tc golemCase) {
	t.Helper()

	if !m.DelFlag || spy.kills != tc.wantKills {
		t.Fatalf("del/kills = %v/%d, want true/%d", m.DelFlag, spy.kills, tc.wantKills)
	}

	if len(spy.spawns) != tc.wantSpawns || len(spy.requests) != tc.wantRequests {
		t.Fatalf("spawns/requests = %v/%v, want %d/%d",
			spy.spawns, spy.requests, tc.wantSpawns, tc.wantRequests)
	}

	if tc.wantSpawns == 1 && spy.spellLvl != 3 {
		t.Fatalf("spell level = %d, want 3", spy.spellLvl)
	}
}
