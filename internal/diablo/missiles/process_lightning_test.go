// Package missiles_test pins ProcessLightning (missiles.cpp:3373-3386).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// lightningProcessCases holds the hand-computed ticks: the duration decrements
// first (cpp:3375), the collision only runs once the bolt left its start tile
// (cpp:3377), a hit restores the tick the collision zeroed (cpp:3379-3380),
// and 0 deletes the bolt and drops its light (cpp:3381-3384).
var lightningProcessCases = []lightningProcessCase{
	{
		name:     "bolt still on its start tile skips the collision",
		duration: 3, start: [2]int{4, 4}, tile: [2]int{4, 4},
		wantDuration: 2,
	},
	{
		name:     "hit on the current tile restores the tick",
		duration: 3, start: [2]int{4, 4}, tile: [2]int{6, 4}, hit: true,
		wantDuration: 2,
	},
	{
		name:     "clean flight just burns the tick",
		duration: 3, start: [2]int{4, 4}, tile: [2]int{6, 4},
		wantDuration: 2,
	},
	{
		name:     "reaching zero deletes the bolt and unlights it",
		duration: 1, start: [2]int{4, 4}, tile: [2]int{4, 4},
		wantDuration: 0, wantDel: true, wantUnlight: true,
	},
}

// lightningProcessCase pins one ProcessLightning tick.
type lightningProcessCase struct {
	name         string
	duration     int
	start        [2]int
	tile         [2]int
	hit          bool
	wantDuration int
	wantDel      bool
	wantUnlight  bool
}

// TestProcessLightning pins the decrement, the start-tile collision gate, the
// hit refund and the delete/unlight tail.
func TestProcessLightning(t *testing.T) {
	// given
	cases := lightningProcessCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			w, unlit := lightningWorld(tc)
			m := lightningMissile(tc)

			// when
			missiles.ProcessLightning(w, &m)

			// then
			assertLightningTick(t, m, tc)
			if unlit() != tc.wantUnlight {
				t.Fatalf("unlit = %v, want %v", unlit(), tc.wantUnlight)
			}
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// lightningWorld wires the tick's sinks: an empty grid plus an AddUnLight
// recorder that reports whether the bolt's light was dropped.
func lightningWorld(tc lightningProcessCase) (*missiles.HitWorld, func() bool) {
	w := &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: missiles.NoopProcessCallbacks(),
	}
	if tc.hit {
		w.DMonster[tc.tile[1]][tc.tile[0]] = 1
		w.HitCallbacks.MonsterTargeted = func(int16, *missiles.Direction) bool { return true }
		w.HitCallbacks.MonsterMHit = func(int, int, int, missiles.MissileID,
			missiles.DamageType, bool, missiles.KeepOnCollision,
		) bool {
			return true
		}
	}

	unlit := false
	w.ProcessCallbacks.AddUnLight = func(int) { unlit = true }

	return w, func() bool { return unlit }
}

// lightningMissile seeds one flying bolt with the case's tiles and duration.
func lightningMissile(tc lightningProcessCase) missiles.Missile {
	m := missiles.Missile{Type: missiles.Lightning, Duration: tc.duration, Dam: 12, Lid: 42}
	m.SetTraveledFromTile(tc.tile[0], tc.tile[1])
	m.Position.Start = missiles.TilePosition{X: tc.start[0], Y: tc.start[1]}

	return m
}

// assertLightningTick checks one post-tick state.
func assertLightningTick(t *testing.T, m missiles.Missile, tc lightningProcessCase) {
	t.Helper()

	if m.Duration != tc.wantDuration {
		t.Fatalf("Duration = %d, want %d", m.Duration, tc.wantDuration)
	}

	if m.DelFlag != tc.wantDel {
		t.Fatalf("DelFlag = %v, want %v", m.DelFlag, tc.wantDel)
	}
}
