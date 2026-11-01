// Package missiles_test pins AddTrapDisarm (missiles.cpp:2636-2650).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/enums"
	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// trapDisarmCases pin the four hands: the disarm cursor is always armed for the
// local player, and under a non-mouse control scheme an object under the cursor
// is disarmed while an empty tile falls back to the hand cursor
// (cpp:2642-2647).
var trapDisarmCases = []trapDisarmCase{
	{
		name: "mouse control stops at the disarm cursor", mouse: true,
		wantCursor: enums.CursorDisarm,
	},
	{
		name: "object under cursor is disarmed", objectUnderCursor: true,
		wantCursor: enums.CursorDisarm, wantDisarms: 1,
	},
	{
		name:       "empty tile falls back to the hand cursor",
		wantCursor: enums.CursorHand,
	},
}

// trapDisarmCase pins one AddTrapDisarm call.
type trapDisarmCase struct {
	name              string
	mouse             bool
	objectUnderCursor bool
	wantCursor        enums.CursorID
	wantDisarms       int
}

// trapDisarmSpies records the cursor and the disarm command.
type trapDisarmSpies struct {
	cursor  enums.CursorID
	disarms int
}

// TestAddTrapDisarm pins the cursor handoff and the disarm command.
func TestAddTrapDisarm(t *testing.T) {
	// given
	cases := trapDisarmCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			spies := &trapDisarmSpies{}
			w := trapDisarmWorld(spies, tc)
			m := missiles.Missile{Source: 0}

			// when
			missiles.AddTrapDisarm(w, &m, &missiles.AddParameter{})

			// then
			assertTrapDisarm(t, m, spies, tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// assertTrapDisarm checks the cursor handoff, the disarm command and the
// dropped missile.
func assertTrapDisarm(t *testing.T, m missiles.Missile, spies *trapDisarmSpies, tc trapDisarmCase) {
	t.Helper()

	if spies.cursor != tc.wantCursor { // cpp:2642-2647
		t.Fatalf("cursor = %v, want %v", spies.cursor, tc.wantCursor)
	}

	if spies.disarms != tc.wantDisarms { // cpp:2645
		t.Fatalf("disarm commands = %d, want %d", spies.disarms, tc.wantDisarms)
	}

	if !m.DelFlag { // cpp:2640
		t.Fatal("DelFlag = false, want the missile dropped")
	}
}

// trapDisarmWorld wires the control mode, the cursor object and the command.
func trapDisarmWorld(spies *trapDisarmSpies, tc trapDisarmCase) *missiles.HitWorld {
	cb := missiles.NoopProcessCallbacks()
	cb.IsMyPlayer = func(int) bool { return true }
	cb.ControlIsMouse = func() bool { return tc.mouse }
	cb.ObjectUnderCursor = func() bool { return tc.objectUnderCursor }
	cb.SetCursor = func(cursor enums.CursorID) { spies.cursor = cursor }
	cb.SendDisarmAtCursor = func() { spies.disarms++ }

	return &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: cb,
	}
}
