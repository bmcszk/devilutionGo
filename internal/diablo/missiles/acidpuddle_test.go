// Package missiles_test pins AddAcidPuddle (missiles.cpp:2355-2361) and
// ProcessAcidPuddle (cpp:2295-2320).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// acidPuddleAddCases pin the lifetime: GenerateRnd(15) plus 40 per point of
// the source monster's intelligence + 1 (cpp:2358).
var acidPuddleCases = []acidPuddleCase{
	{
		name: "countdown probes and restores", duration: 5, group: 0,
		wantDuration: 4, wantProbes: 1,
	},
	{
		name: "zero past the idle frames deletes", duration: 1, group: 1,
		wantDel: true, wantProbes: 1,
	},
	{
		name: "zero on the idle frames starts the end frames", duration: 1, group: 0,
		animLen: 6, wantDuration: 6, wantGroup: 1, wantProbes: 1,
	},
}

var acidPuddleAddCases = []acidPuddleAddCase{
	{name: "intelligence 0", intelligence: 0, rnd: []int{7}, wantDuration: 47},
	{name: "intelligence 2", intelligence: 2, rnd: []int{0}, wantDuration: 120},
}

// acidPuddleAddCase pins one AddAcidPuddle row.
type acidPuddleAddCase struct {
	name         string
	intelligence int
	rnd          []int
	wantDuration int
}

// acidPuddleCases pin the ticks: the puddle counts down, probes its own tile
// with the duration restored afterwards (cpp:2300-2306), and on the zero tick
// either deletes (past the idle frames) or swaps to the End frames and lives
// another AnimLen (cpp:2307-2314).
// acidPuddleCase pins one ProcessAcidPuddle tick.
type acidPuddleCase struct {
	name         string
	duration     int
	group        int
	animLen      int
	wantDuration int
	wantGroup    int
	wantProbes   int
	wantDel      bool
}

// acidPuddleSpies records the collision probes of one tick.
type acidPuddleSpies struct {
	probes int
}

// TestAddAcidPuddle pins the random lifetime, the render flags and the
// intelligence scaling.
func TestAddAcidPuddle(t *testing.T) {
	// given
	cases := acidPuddleAddCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			tape := &rndTape{t: t, tape: tc.rnd}
			w := acidPuddleWorld(&acidPuddleSpies{}, tape, tc.intelligence)
			m := missiles.Missile{Type: missiles.AcidPuddle, Source: 2}

			// when
			missiles.AddAcidPuddle(w, &m, &missiles.AddParameter{})
			tape.assertConsumed()

			// then
			assertAcidPuddleAdd(t, m, tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// TestProcessAcidPuddle pins the countdown, the own-tile probe and both zero
// branches.
func TestProcessAcidPuddle(t *testing.T) {
	// given
	cases := acidPuddleCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			spies := &acidPuddleSpies{}
			w := acidPuddleWorld(spies, &rndTape{t: t}, 0)

			// when
			m := acidPuddleMissile(tc)
			missiles.ProcessAcidPuddle(w, &m)

			// then
			if spies.probes != tc.wantProbes {
				t.Fatalf("collision probes = %d, want %d", spies.probes, tc.wantProbes)
			}

			assertAcidPuddleState(t, m, tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// assertAcidPuddleAdd checks the random lifetime and the two render flags of
// one add (cpp:2357-2360).
func assertAcidPuddleAdd(t *testing.T, m missiles.Missile, tc acidPuddleAddCase) {
	t.Helper()

	if m.Duration != tc.wantDuration { // cpp:2358
		t.Fatalf("Duration = %d, want %d", m.Duration, tc.wantDuration)
	}

	if !m.PreFlag || !m.LightFlag { // cpp:2357, 2360
		t.Fatalf("pre/light flag = %v/%v, want true/true", m.PreFlag, m.LightFlag)
	}
}

// acidPuddleWorld wires the monster read, the draw tape and a probe counter.
func acidPuddleWorld(spies *acidPuddleSpies, tape *rndTape, intelligence int) *missiles.HitWorld {
	cb := missiles.NoopProcessCallbacks()
	cb.GenerateRnd = tape.draw

	w := &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: cb,
		AddSource: missiles.AddSource{
			MonsterStats: func(int) missiles.AddMonsterStats {
				return missiles.AddMonsterStats{Intelligence: intelligence}
			},
		},
	}
	w.DMonster[4][4] = 1
	w.HitCallbacks.MonsterTargeted = func(int16, *missiles.Direction) bool {
		spies.probes++

		return false
	}

	return w
}

// acidPuddleMissile seeds the puddle at tile (4, 4).
func acidPuddleMissile(tc acidPuddleCase) missiles.Missile {
	m := missiles.Missile{
		Type: missiles.AcidPuddle, Duration: tc.duration, AnimLen: tc.animLen,
		MFNum: tc.group, Source: 2, Dam: 3,
	}
	m.SetTraveledFromTile(4, 4)

	return m
}

// assertAcidPuddleState checks the duration, frame group and delete flag.
func assertAcidPuddleState(t *testing.T, m missiles.Missile, tc acidPuddleCase) {
	t.Helper()

	if m.Duration != tc.wantDuration || m.DelFlag != tc.wantDel {
		t.Fatalf("duration/del = %d/%v, want %d/%v",
			m.Duration, m.DelFlag, tc.wantDuration, tc.wantDel)
	}

	if tc.wantGroup != 0 && m.MFNum != tc.wantGroup { // missiles.h:69-72
		t.Fatalf("frame group = %d, want %d", m.MFNum, tc.wantGroup)
	}
}
