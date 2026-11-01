// Package missiles_test pins AddRedPortal (missiles.cpp:2695-2701) and
// ProcessRedPortal (cpp:4100-4125).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// redPortalCases pin the ticks: the duration counts down to 1 (cpp:4104-4105),
// the Start->Idle frame flip happens when the duration reaches var1
// (cpp:4106-4107), the light ramps only outside town (cpp:4109-4114) and the
// last tick deletes the portal and removes its light (cpp:4115-4118).
var redPortalCases = []redPortalCase{
	{
		// The portal is still in its Start frames, so the light ramps
		// (cpp:4109 checks the group, not the duration alone).
		name: "start frames ramp the light", duration: 90, var1: 88, levelType: 1,
		wantDuration: 89, wantLight: 1, wantChange: 1,
	},
	{
		name: "reaching var1 flips to the idle group", duration: 89, var1: 88, levelType: 1,
		wantDuration: 88, wantGroup: 1,
	},
	{
		name: "a later start-frame tick keeps ramping", duration: 80, var1: 88, var2: 1,
		levelType: 1, wantDuration: 79, wantLight: 0, wantChange: 1, wantChangeRadius: 2,
	},
	{
		name: "third ramp step moves to radius 4", duration: 80, var1: 88, var2: 3,
		levelType: 1, wantDuration: 79, wantLight: 0, wantChange: 1, wantChangeRadius: 4,
	},
	{
		name: "town portals never ramp", duration: 80, var1: 88, levelType: 0,
		wantDuration: 79,
	},
	{
		// Upstream only decrements while the duration is above 1, so a portal
		// sitting at 1 stays there; the delete branch belongs to a portal the
		// game logic zeroed (cpp:4104, 4115-4118).
		name: "a zeroed portal deletes and unlights", var1: 88, var2: 2,
		levelType: 1, wantDel: true, wantUnlight: true,
	},
}

// redPortalCase pins one ProcessRedPortal tick.
type redPortalCase struct {
	name             string
	duration         int
	var1             int
	var2             int
	levelType        int
	wantDuration     int
	wantGroup        int
	wantLight        int
	wantChange       int
	wantChangeRadius int
	wantUnlight      bool
	wantDel          bool
}

// redPortalSpies records the light calls of one tick.
type redPortalSpies struct {
	added      int
	addedRad   []uint8
	changedRad []uint8
	unlighted  int
}

// TestAddRedPortal pins the 100-tick lifetime and the var1 flip threshold.
func TestAddRedPortal(t *testing.T) {
	// given
	w := redPortalWorld(&redPortalSpies{}, 0)
	m := missiles.Missile{Type: missiles.RedPortal, AnimLen: 5}

	// when
	missiles.AddRedPortal(w, &m, &missiles.AddParameter{})

	// then
	if m.Duration != 100 || m.Var1 != 95 { // cpp:2697-2698
		t.Fatalf("duration/var1 = %d/%d, want 100/95", m.Duration, m.Var1)
	}
}

// TestProcessRedPortal pins the countdown, the frame flip, the light ramp and
// the delete tail.
func TestProcessRedPortal(t *testing.T) {
	// given
	cases := redPortalCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			spies := &redPortalSpies{}
			w := redPortalWorld(spies, tc.levelType)

			// when
			m := redPortalMissile(tc)
			missiles.ProcessRedPortal(w, &m)

			// then
			assertRedPortalState(t, m, tc)
			assertRedPortalLights(t, spies, tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// redPortalWorld wires the light sink and the dungeon-type read.
func redPortalWorld(spies *redPortalSpies, levelType int) *missiles.HitWorld {
	cb := missiles.NoopProcessCallbacks()
	cb.LevelType = levelType
	cb.AddLight = func(_, _ int, radius uint8) int {
		spies.added++
		spies.addedRad = append(spies.addedRad, radius)

		return 7
	}
	cb.ChangeLight = func(_, _, _ int, radius uint8) {
		spies.changedRad = append(spies.changedRad, radius)
	}
	cb.AddUnLight = func(int) { spies.unlighted++ }

	return &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: cb,
	}
}

// redPortalMissile seeds the portal at tile (4, 4) with a light id.
func redPortalMissile(tc redPortalCase) missiles.Missile {
	m := missiles.Missile{
		Type: missiles.RedPortal, Duration: tc.duration, Var1: tc.var1, Var2: tc.var2,
		Lid: 7,
	}
	m.SetTraveledFromTile(4, 4)

	return m
}

// assertRedPortalState checks the duration, frame group and delete flag.
func assertRedPortalState(t *testing.T, m missiles.Missile, tc redPortalCase) {
	t.Helper()

	if m.Duration != tc.wantDuration || m.DelFlag != tc.wantDel {
		t.Fatalf("duration/del = %d/%v, want %d/%v",
			m.Duration, m.DelFlag, tc.wantDuration, tc.wantDel)
	}

	if tc.wantGroup != 0 && m.MFNum != tc.wantGroup { // missiles.h:84-87
		t.Fatalf("frame group = %d, want the idle group %d", m.MFNum, tc.wantGroup)
	}
}

// assertRedPortalLights checks the ramp: AddLight once when var2 is 0, then a
// ChangeLight on every ramping tick, and the delete tick's AddUnLight.
func assertRedPortalLights(t *testing.T, spies *redPortalSpies, tc redPortalCase) {
	t.Helper()

	if spies.added != tc.wantLight {
		t.Fatalf("AddLight calls = %d, want %d", spies.added, tc.wantLight)
	}

	if len(spies.changedRad) != tc.wantChange {
		t.Fatalf("ChangeLight calls = %d, want %d", len(spies.changedRad), tc.wantChange)
	}

	if tc.wantChangeRadius != 0 && len(spies.changedRad) > 0 &&
		int(spies.changedRad[0]) != tc.wantChangeRadius {
		t.Fatalf("ChangeLight radius = %d, want %d", spies.changedRad[0], tc.wantChangeRadius)
	}

	if (spies.unlighted != 0) != tc.wantUnlight {
		t.Fatalf("AddUnLight calls = %d, want unlight %v", spies.unlighted, tc.wantUnlight)
	}
}
