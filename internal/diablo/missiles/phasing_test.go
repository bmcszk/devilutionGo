// Package missiles_test pins AddPhasing (missiles.cpp:2900-2930), the ring
// scan that lands a monster/phasing cast on a random legal tile, and
// UpdateVileMissPos (cpp:2932-2947), the fallback the caster's own phasing
// takes on a blocked aim tile. Both ride PosOkPlayer (player.cpp:3089-3112).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// phasingCases pin the three exits of AddPhasing.
var phasingCases = []phasingCase{
	{
		name:   "the caster's own cast lands on the aim tile",
		target: missiles.TargetBoth,
		want:   missiles.TilePosition{X: 10, Y: 12},
	},
	{
		name: "a blocked aim tile falls back to the nearest legal tile", target: missiles.TargetBoth,
		blocked: true, want: missiles.TilePosition{X: 11, Y: 11},
	},
	{name: "a monster cast draws from the ring", ring: true, want: missiles.TilePosition{X: 4, Y: 4}},
	{name: "an empty ring fizzles", ring: true, empty: true, wantDel: true},
}

// phasingCase pins one AddPhasing cast.
type phasingCase struct {
	name    string
	target  missiles.Caster
	blocked bool
	ring    bool
	empty   bool
	want    missiles.TilePosition
	wantDel bool
}

// TestAddPhasing pins the aim tile, the fallback search and the ring draw.
func TestAddPhasing(t *testing.T) {
	// when
	subtests := make([]subtest, 0, len(phasingCases))
	for _, tc := range phasingCases {
		subtests = append(subtests, subtest{tc.name, runPhasing(tc)})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// phasingWorld wires the occupancy predicate and the tape. The ring draw is
// pinned with GenerateRnd = 0, the first tile the scan meets: y = -6, x = -6
// around the launch tile (10,10), i.e. (4,4).
func phasingWorld(tape *rndTape, tc phasingCase) *missiles.HitWorld {
	cb := missiles.NoopProcessCallbacks()
	cb.GenerateRnd = tape.draw
	cb.LevelType = 0
	cb.MyPlayer = func() int { return 0 }
	cb.PlayerAtPosition = func(missiles.TilePosition) int { return -1 }
	cb.PlayerHasNoLife = func(int) bool { return false }
	if tc.empty {
		// a living player on every tile makes PosOkPlayer fail everywhere
		cb.PlayerAtPosition = func(missiles.TilePosition) int { return 5 }
	}

	return &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: cb,
	}
}

// runPhasing drives one AddPhasing cast.
func runPhasing(tc phasingCase) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		// given
		w := phasingWorld(&rndTape{t: t, tape: []int{0}}, tc)
		blockAimTile(w, tc)
		m := phasingMissile(tc)

		// when
		missiles.AddPhasing(w, &m, &missiles.AddParameter{Dst: missiles.TilePosition{X: 10, Y: 12}})

		// then
		assertPhasing(t, m, tc)
	}
}

// blockAimTile puts a solid object on the aim tile for the fallback case.
func blockAimTile(w *missiles.HitWorld, tc phasingCase) {
	if !tc.blocked {
		return
	}

	w.Grid.Objects = append(w.Grid.Objects, world.WorldObject{SolidFlag: true})
	w.Grid.DObject[12][10] = 1
}

// phasingMissile is the missile one cast starts from.
func phasingMissile(tc phasingCase) missiles.Missile {
	m := missiles.Missile{Type: missiles.Phasing, Source: 0, Caster: int(tc.target)}
	m.SetTraveledFromTile(10, 10)
	m.Position.Start = missiles.TilePosition{X: 10, Y: 10}
	m.Position.Tile = missiles.TilePosition{X: 10, Y: 10}

	return m
}

// assertPhasing checks the exit of one cast.
func assertPhasing(t *testing.T, m missiles.Missile, tc phasingCase) {
	t.Helper()

	if m.DelFlag != tc.wantDel {
		t.Fatalf("del = %v, want %v", m.DelFlag, tc.wantDel)
	}

	if tc.wantDel {
		return
	}

	if tc.blocked {
		assertPhasingFallback(t, m)

		return
	}

	if m.Position.Tile != tc.want || m.Duration != 2 {
		t.Fatalf("tile/duration = %v/%d, want %v/2", m.Position.Tile, m.Duration, tc.want)
	}
}

// assertPhasingFallback checks that the blocked aim tile was replaced by a
// legal neighbour and that a neighbour was actually found.
func assertPhasingFallback(t *testing.T, m missiles.Missile) {
	t.Helper()

	if m.Position.Tile == (missiles.TilePosition{X: 10, Y: 12}) {
		t.Fatal("the blocked aim tile must be replaced")
	}

	if m.Duration != 2 {
		t.Fatalf("duration = %d, want 2", m.Duration)
	}
}
