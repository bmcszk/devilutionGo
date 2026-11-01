// Package missiles_test pins AddFireball (missiles.cpp:1970-1996).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
)

// fireballCases holds the hand-computed rows: velocity is the 16.16 normal
// vector times sp, damage is ScaleSpellEffect(2*(level+2d10)+4, spllvl)
// (missiles.cpp:119-126, 1980-1983), and the frame group is the 16-step
// facing of the launch-to-destination vector.
var fireballCases = []fireballCase{
	{
		name: "level 1 player at base speed",
		script: addScript{
			dst: [2]int{10, 10}, midir: missiles.South, caster: 0, source: 0,
			level: 1, light: 42, rnd: []int{5, 5},
		},
		// sp = 16, 2*(1 + 10) + 4 = 26, spllvl 0 keeps 26
		wantVel: [2]int32{0, 524288}, wantDam: 26, wantMF: 0,
	},
	{
		name: "spell level 5 raises speed and damage",
		script: addScript{
			dst: [2]int{10, 10}, midir: missiles.South, caster: 0, source: 0,
			spllvl: 5, level: 1, light: 42, rnd: []int{5, 5},
		},
		// sp = 16+10 = 26; 26 -> 29 -> 32 -> 36 -> 40 -> 45 over 5 levels
		wantVel: [2]int32{0, 851968}, wantDam: 45, wantMF: 0,
	},
	{
		name: "spell level 30 clamps the speed bonus",
		script: addScript{
			dst: [2]int{10, 10}, midir: missiles.South, caster: 0, source: 0,
			spllvl: 30, level: 0, light: 42, rnd: []int{0, 0},
		},
		// sp = 16+34 = 50; a base of 4 never grows (4/8 == 0)
		wantVel: [2]int32{0, 1638400}, wantDam: 4, wantMF: 0,
	},
	{
		name: "monster caster keeps the flat speed and skips the damage roll",
		script: addScript{
			dst: [2]int{10, 10}, midir: missiles.South,
			caster: int(missiles.TargetPlayers), source: 3, level: 1, light: 42,
			dam: 7,
		},
		wantVel: [2]int32{0, 524288}, wantDam: 7, wantMF: 0,
	},
	{
		name: "same tile shot shifts by midir and keeps the 16-step facing",
		script: addScript{
			start: [2]int{5, 5}, dst: [2]int{5, 5}, midir: missiles.South,
			caster: 0, source: 0, level: 0, light: 7, rnd: []int{1, 2},
		},
		// dst becomes (5,6): due south, sp 16, damage 2*(0+3)+4 = 10
		wantVel: [2]int32{0, 524288}, wantDam: 10, wantMF: 0,
	},
}

// fireballCase pins one AddFireball row.
type fireballCase struct {
	name    string
	script  addScript
	wantVel [2]int32
	wantDam int
	wantMF  int
}

// TestAddFireball pins the body against the hand-computed values above and
// asserts the tape is consumed exactly (two GenerateRnd(10) draws per damage
// roll: missiles.cpp:1981).
func TestAddFireball(t *testing.T) {
	// given
	cases := fireballCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			fx := newAddFixture(t, tc.script)

			// when
			missiles.AddFireball(fx.world, fx.missile, fx.param)
			fx.tape.assertConsumed()

			// then
			assertFireball(t, fx, tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// assertFireball checks one post-AddFireball row.
func assertFireball(t *testing.T, fx *addFixture, tc fireballCase) {
	t.Helper()

	m := fx.missile

	velocity := m.Position.Velocity
	if velocity.DeltaX != tc.wantVel[0] || velocity.DeltaY != tc.wantVel[1] {
		t.Fatalf("velocity = (%d, %d), want (%d, %d)",
			velocity.DeltaX, velocity.DeltaY, tc.wantVel[0], tc.wantVel[1])
	}

	if m.Dam != tc.wantDam {
		t.Fatalf("Dam = %d, want %d", m.Dam, tc.wantDam)
	}

	if m.MFNum != tc.wantMF {
		t.Fatalf("frame group = %d, want %d", m.MFNum, tc.wantMF)
	}

	if m.Duration != 256 { // cpp:1987
		t.Fatalf("Duration = %d, want 256", m.Duration)
	}

	assertFireballAnchor(t, fx)
}

// assertFireballAnchor checks the start-tile anchor of the add: var1/var2
// hold the launch tile (cpp:1988-1989) and the light sits there at radius 8
// (cpp:1990).
func assertFireballAnchor(t *testing.T, fx *addFixture) {
	t.Helper()

	m := fx.missile

	if m.Var1 != m.Position.Start.X || m.Var2 != m.Position.Start.Y {
		t.Fatalf("var1/var2 = %d/%d, want the start tile (%d, %d)",
			m.Var1, m.Var2, m.Position.Start.X, m.Position.Start.Y)
	}

	if fx.light.radius != 8 {
		t.Fatalf("AddLight radius = %d, want 8", fx.light.radius)
	}

	if fx.light.x != m.Position.Start.X || fx.light.y != m.Position.Start.Y {
		t.Fatalf("AddLight at (%d, %d), want the start tile", fx.light.x, fx.light.y)
	}
}
