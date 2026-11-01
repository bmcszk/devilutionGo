// Package missiles_test pins AddNovaBall (missiles.cpp:1944-1957).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
)

// novaBallCases holds the hand-computed rows: velocity is the 16.16 normal
// vector times 16, the start frame is RandomIntBetween(1, 8), duration is the
// fixed 255, and var1/var2 anchor on the caster's tile when the missile has a
// source and on its own launch tile otherwise (missiles.cpp:1949-1956).
var novaBallCases = []novaBallCase{
	{
		name: "player source anchors the nova on the caster tile",
		script: addScript{
			dst: [2]int{10, 10}, midir: missiles.South, caster: 0, source: 0,
			tile: [2]int{7, 9}, rnd: []int{4},
		},
		wantVel: [2]int32{0, 524288}, wantFrame: 5, wantAnchor: [2]int{7, 9},
	},
	{
		name: "sourceless missile anchors on its own launch tile",
		script: addScript{
			start: [2]int{3, 4}, dst: [2]int{10, 10}, midir: missiles.South,
			caster: 0, source: -1, rnd: []int{0},
		},
		wantVel: [2]int32{80416, 522736}, wantFrame: 1, wantAnchor: [2]int{3, 4},
	},
	{
		name: "diagonal destination keeps the 16.16 normal vector",
		script: addScript{
			dst: [2]int{5, 0}, midir: missiles.South, caster: 0, source: 0,
			tile: [2]int{1, 1}, rnd: []int{7},
		},
		wantVel: [2]int32{741440, 370720}, wantFrame: 8, wantAnchor: [2]int{1, 1},
	},
}

// novaBallCase pins one AddNovaBall row.
type novaBallCase struct {
	name       string
	script     addScript
	wantVel    [2]int32
	wantFrame  int
	wantAnchor [2]int
}

// TestAddNovaBall pins velocity, start frame, duration and the anchor ternary.
func TestAddNovaBall(t *testing.T) {
	// given
	cases := novaBallCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			fx := newAddFixture(t, tc.script)

			// when
			missiles.AddNovaBall(fx.world, fx.missile, fx.param)
			fx.tape.assertConsumed()

			// then
			assertNovaBall(t, fx.missile, tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// assertNovaBall checks one post-AddNovaBall row.
func assertNovaBall(t *testing.T, m *missiles.Missile, tc novaBallCase) {
	t.Helper()

	velocity := m.Position.Velocity
	if velocity.DeltaX != tc.wantVel[0] || velocity.DeltaY != tc.wantVel[1] {
		t.Fatalf("velocity = (%d, %d), want (%d, %d)",
			velocity.DeltaX, velocity.DeltaY, tc.wantVel[0], tc.wantVel[1])
	}

	if m.AnimFrame != tc.wantFrame { // cpp:1947
		t.Fatalf("AnimFrame = %d, want %d", m.AnimFrame, tc.wantFrame)
	}

	if m.Duration != 255 { // cpp:1948
		t.Fatalf("Duration = %d, want 255", m.Duration)
	}

	if m.Var1 != tc.wantAnchor[0] || m.Var2 != tc.wantAnchor[1] { // cpp:1950-1956
		t.Fatalf("anchor = (%d, %d), want (%d, %d)",
			m.Var1, m.Var2, tc.wantAnchor[0], tc.wantAnchor[1])
	}
}
