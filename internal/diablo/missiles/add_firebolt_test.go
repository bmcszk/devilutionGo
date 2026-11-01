// Package missiles_test pins AddFirebolt (missiles.cpp:1860-1894): the
// spell-level velocity cap, the 16-step frame group, the light and the
// player/monster/trap damage roll.
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
)

// addFireboltCases holds the hand-computed AddFirebolt rows. The velocity is
// 16.16 pixels/tick for the (tile - dst) normal vector (missiles.cpp:173-185).
var addFireboltCases = []addFireboltCase{
	{
		name: "player spell level 0", script: addScript{
			dst: [2]int{10, 10}, caster: 0, source: 0, spllvl: 0, dam: 0, magic: 8,
			light: 42, rnd: []int{7},
		},
		wantVel: [2]int32{0, 524288}, wantDam: 9, wantLid: 42,
	},
	{
		name: "spell level capped at 47", script: addScript{
			dst: [2]int{10, 10}, caster: 0, source: 0, spllvl: 30, magic: 0,
			light: 42, rnd: []int{0},
		},
		wantVel: [2]int32{0, 2064384}, wantDam: 31, wantLid: 42,
	},
	{
		name: "trap uses base velocity and trap dice", script: addScript{
			dst: [2]int{10, 10}, caster: 0, source: -1, spllvl: 30, currlevel: 5,
			light: 42, rnd: []int{3},
		},
		wantVel: [2]int32{0, 524288}, wantDam: 8, wantLid: 42,
	},
	{
		name: "monster rolls its damage range", script: addScript{
			dst: [2]int{10, 10}, caster: 1, source: 3, monsterMin: 2, monsterMax: 5,
			light: 42, rnd: []int{2},
		},
		wantVel: [2]int32{0, 851968}, wantDam: 4, wantLid: 42,
	},
	{
		name: "pre-rolled damage skips the roll", script: addScript{
			dst: [2]int{10, 10}, caster: 0, source: 0, dam: 100, light: 42,
		},
		wantVel: [2]int32{0, 524288}, wantDam: 100, wantLid: 42,
	},
	{
		name: "same tile shot shifts by midir", script: addScript{
			start: [2]int{5, 5}, dst: [2]int{5, 5}, midir: missiles.South, caster: 0, source: 0,
			dam: 1, light: 42,
		},
		wantVel: [2]int32{0, 524288}, wantDam: 1, wantLid: 42,
	},
	{
		name: "diagonal destination sets the 16 frame group", script: addScript{
			dst: [2]int{5, 0}, caster: 0, source: 0, dam: 1, light: 42,
		},
		wantVel: [2]int32{741440, 370720}, wantDam: 1, wantLid: 42, wantMFNum: 14,
	},
}

// addFireboltCase is one AddFirebolt row.
type addFireboltCase struct {
	name      string
	script    addScript
	wantVel   [2]int32
	wantDam   int
	wantLid   int
	wantMFNum int
}

func TestAddFirebolt(t *testing.T) {
	for _, tt := range addFireboltCases {
		t.Run(tt.name, func(t *testing.T) {
			// given
			fx := newAddFixture(t, tt.script)
			w, m, p := fx.world, fx.missile, fx.param

			// when
			missiles.AddFirebolt(w, m, p)
			// then
			fx.tape.assertConsumed()

			assertFireboltState(t, m, tt)
		})
	}
}

// assertFireboltState checks the post-AddFirebolt motion, frame, light and
// damage fields (missiles.cpp:1871-1892).
func assertFireboltState(t *testing.T, m *missiles.Missile, tt addFireboltCase) {
	t.Helper()

	assertFireboltMotion(t, m, tt)
	assertFireboltFields(t, m, tt)
}

// assertFireboltMotion checks the velocity, frame group, frame and duration.
func assertFireboltMotion(t *testing.T, m *missiles.Missile, tt addFireboltCase) {
	t.Helper()

	velocity := m.Position.Velocity
	if velocity.DeltaX != tt.wantVel[0] || velocity.DeltaY != tt.wantVel[1] {
		t.Fatalf("velocity = (%d, %d), want (%d, %d)", velocity.DeltaX, velocity.DeltaY,
			tt.wantVel[0], tt.wantVel[1])
	}

	if m.MFNum != tt.wantMFNum {
		t.Fatalf("frame group = %d, want %d", m.MFNum, tt.wantMFNum)
	}

	if m.AnimFrame != 1 { // setAnimation resets the frame (missiles.cpp:96)
		t.Fatalf("animFrame = %d, want 1", m.AnimFrame)
	}

	if m.Duration != 256 { // cpp:1877
		t.Fatalf("duration = %d, want 256", m.Duration)
	}
}

// assertFireboltFields checks the scratch registers, light and damage.
func assertFireboltFields(t *testing.T, m *missiles.Missile, tt addFireboltCase) {
	t.Helper()

	if m.Var1 != m.Position.Start.X || m.Var2 != m.Position.Start.Y { // cpp:1878-1879
		t.Fatalf("var1/var2 = (%d, %d), want start tile (%d, %d)", m.Var1, m.Var2,
			m.Position.Start.X, m.Position.Start.Y)
	}

	if m.Lid != tt.wantLid { // cpp:1880
		t.Fatalf("lid = %d, want %d", m.Lid, tt.wantLid)
	}

	if m.Dam != tt.wantDam { // cpp:1881-1892
		t.Fatalf("dam = %d, want %d", m.Dam, tt.wantDam)
	}
}
