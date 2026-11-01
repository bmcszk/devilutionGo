// Package missiles_test pins AddLightning (missiles.cpp:2001-2018).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
)

// lightningCases holds the hand-computed rows: the launch tile snaps to dst,
// the frame group starts at RandomIntBetween(1, 8), a trap or a demonic
// familiar lives 8 ticks, any other player-cast bolt 10, and a monster bolt
// (_mispllvl / 2) + 6 (missiles.cpp:2003-2016). The light rides the pre-snap
// tile (missiles.cpp:2017).
var lightningCases = []lightningCase{
	{
		name: "trap bolt snaps the start tile and lives 8 ticks",
		script: addScript{
			start: [2]int{3, 4}, dst: [2]int{9, 6}, caster: int(missiles.TargetPlayers),
			source: -1, rnd: []int{0}, light: 77,
		},
		wantStart: [2]int{9, 6}, wantFrame: 1, wantDuration: 8, wantLight: [2]int{3, 4},
	},
	{
		name: "familiar-sourced player bolt also lives 8 ticks",
		script: addScript{
			start: [2]int{5, 5}, dst: [2]int{6, 5}, caster: int(missiles.TargetPlayers),
			source: 3, monsterType: missiles.MonsterTypeFamiliar, rnd: []int{7},
		},
		wantStart: [2]int{6, 5}, wantFrame: 8, wantDuration: 8, wantLight: [2]int{5, 5},
	},
	{
		name: "player-sourced bolt from a normal monster lives 10 ticks",
		script: addScript{
			start: [2]int{1, 2}, dst: [2]int{4, 2}, caster: int(missiles.TargetPlayers),
			source: 3, monsterType: 7, rnd: []int{2},
		},
		wantStart: [2]int{4, 2}, wantFrame: 3, wantDuration: 10, wantLight: [2]int{1, 2},
	},
	{
		name: "monster bolt duration is half the spell level plus six",
		script: addScript{
			start: [2]int{8, 8}, dst: [2]int{8, 9}, caster: int(missiles.TargetMonsters),
			source: 3, spllvl: 7, rnd: []int{3},
		},
		wantStart: [2]int{8, 9}, wantFrame: 4, wantDuration: 9, wantLight: [2]int{8, 8},
	},
}

// lightningCase pins one AddLightning row.
type lightningCase struct {
	name         string
	script       addScript
	wantStart    [2]int
	wantFrame    int
	wantDuration int
	wantLight    [2]int
}

// TestAddLightning pins the start-tile snap, tile-fixed light, start frame and
// the three-way duration branch.
func TestAddLightning(t *testing.T) {
	// given
	cases := lightningCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			fx := newAddFixture(t, tc.script)

			// when
			missiles.AddLightning(fx.world, fx.missile, fx.param)
			fx.tape.assertConsumed()

			// then
			assertLightning(t, fx, tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// TestAddLightningInheritsParentPosition pins SyncPositionWithParent
// (missiles.cpp:792-800): the spawned bolt copies the parent's pixel offset
// and traveled vector, not its tiles.
func TestAddLightningInheritsParentPosition(t *testing.T) {
	// given
	fx := newAddFixture(t, addScript{
		start: [2]int{2, 2}, dst: [2]int{4, 4}, caster: int(missiles.TargetPlayers),
		source: -1, rnd: []int{1},
	})
	parent := &missiles.Missile{}
	parent.Position.Offset = missiles.PositionDisplacement{DeltaX: 7, DeltaY: -3}
	parent.Position.Traveled = missiles.PositionDisplacement{DeltaX: 131072, DeltaY: 262144}
	fx.param.Parent = parent

	// when
	missiles.AddLightning(fx.world, fx.missile, fx.param)

	// then
	offset := fx.missile.Position.Offset
	if offset != parent.Position.Offset { // cpp:798
		t.Fatalf("offset = %+v, want %+v", offset, parent.Position.Offset)
	}

	traveled := fx.missile.Position.Traveled
	if traveled != parent.Position.Traveled { // cpp:799
		t.Fatalf("traveled = %+v, want %+v", traveled, parent.Position.Traveled)
	}
}

// assertLightning checks one post-AddLightning row.
func assertLightning(t *testing.T, fx *addFixture, tc lightningCase) {
	t.Helper()

	start := fx.missile.Position.Start
	if start.X != tc.wantStart[0] || start.Y != tc.wantStart[1] { // cpp:2003
		t.Fatalf("start = (%d, %d), want (%d, %d)", start.X, start.Y, tc.wantStart[0], tc.wantStart[1])
	}

	if fx.missile.AnimFrame != tc.wantFrame { // cpp:2007
		t.Fatalf("AnimFrame = %d, want %d", fx.missile.AnimFrame, tc.wantFrame)
	}

	if fx.missile.Duration != tc.wantDuration { // cpp:2009-2016
		t.Fatalf("Duration = %d, want %d", fx.missile.Duration, tc.wantDuration)
	}

	want := lightCall{x: tc.wantLight[0], y: tc.wantLight[1], radius: 4}
	if *fx.light != want { // cpp:2017
		t.Fatalf("light = %+v, want %+v", *fx.light, want)
	}
}
