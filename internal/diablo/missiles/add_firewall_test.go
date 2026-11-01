// Package missiles_test pins AddFireWall (missiles.cpp:1954-1969).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
)

// fireWallCases holds the hand-computed rows: GenerateRndSum(10, 2) + 2 plus
// the source player's character level (or currlevel for a trap), shifted left
// by 3 (cpp:1956-1959); speed 16 (cpp:1960); duration 10 scaled by
// (spell level + 1), plus currlevel for a player-cast or trap row, times 16
// (cpp:1962-1967); and var1 = duration - the row's AnimLen (cpp:1968).
var fireWallCases = []fireWallCase{
	{
		name: "monster-cast wall from a player source",
		script: addScript{
			start: [2]int{0, 0}, dst: [2]int{10, 10}, caster: int(missiles.TargetMonsters),
			source: 0, level: 5, currlevel: 3, animLen: 20, rnd: []int{4, 6},
		},
		wantDam: 136, wantVel: [2]int32{0, 524288}, wantDuration: 160, wantVar1: 140,
	},
	{
		name: "trap wall ignores the player level and adds currlevel",
		script: addScript{
			start: [2]int{3, 4}, dst: [2]int{10, 10}, caster: int(missiles.TargetMonsters),
			source: -1, spllvl: 2, currlevel: 3, rnd: []int{1, 2},
		},
		wantDam: 64, wantVel: [2]int32{80416, 522736}, wantDuration: 528, wantVar1: 528,
	},
	{
		name: "player-cast wall scales the duration by the spell level",
		script: addScript{
			start: [2]int{0, 0}, dst: [2]int{5, 0}, caster: int(missiles.TargetPlayers),
			source: 0, level: 1, spllvl: 1, currlevel: 2, animLen: 7, rnd: []int{9, 0},
		},
		wantDam: 96, wantVel: [2]int32{741440, 370720}, wantDuration: 352, wantVar1: 345,
	},
}

// fireWallCase pins one AddFireWall row.
type fireWallCase struct {
	name         string
	script       addScript
	wantDam      int
	wantVel      [2]int32
	wantDuration int
	wantVar1     int
}

// TestAddFireWall pins the damage shift, velocity, the duration chain and the
// var1 burn-out anchor with a strict two-draw tape.
func TestAddFireWall(t *testing.T) {
	// given
	cases := fireWallCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			fx := newAddFixture(t, tc.script)

			// when
			missiles.AddFireWall(fx.world, fx.missile, fx.param)
			fx.tape.assertConsumed()

			// then
			assertFireWall(t, fx.missile, tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// assertFireWall checks one post-AddFireWall state.
func assertFireWall(t *testing.T, m *missiles.Missile, tc fireWallCase) {
	t.Helper()

	if m.Dam != tc.wantDam { // cpp:1956-1959
		t.Fatalf("Dam = %d, want %d", m.Dam, tc.wantDam)
	}

	velocity := m.Position.Velocity
	if velocity.DeltaX != tc.wantVel[0] || velocity.DeltaY != tc.wantVel[1] { // cpp:1960
		t.Fatalf("velocity = (%d, %d), want (%d, %d)",
			velocity.DeltaX, velocity.DeltaY, tc.wantVel[0], tc.wantVel[1])
	}

	if m.Duration != tc.wantDuration { // cpp:1962-1967
		t.Fatalf("Duration = %d, want %d", m.Duration, tc.wantDuration)
	}

	if m.Var1 != tc.wantVar1 { // cpp:1968
		t.Fatalf("Var1 = %d, want %d", m.Var1, tc.wantVar1)
	}
}
