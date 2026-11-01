// Package missiles_test pins AddInfravision (missiles.cpp:2554-2557) and
// ProcessInfravision (missiles.cpp:3839-3849).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// infravisionCases pin the ticks: the flag is raised every tick and the last
// tick deletes the missile and recomputes the player's item values.
var infravisionCases = []infravisionCase{
	{name: "first tick raises the flag", duration: 3, wantDuration: 2, wantFlag: 1},
	{name: "last tick deletes and recomputes", duration: 1, wantFlag: 1, wantDel: true, wantRecalc: 1},
}

// infravisionAddCases pin the duration: ScaleSpellEffect(1584, spllvl)
// (cpp:2556). The scale compounds — 1584, then 1584 + 198 = 1782, then
// 1782 + 222 = 2004 (ScaleSpellEffect iterates, missiles.cpp:119-125).
// infravisionAddCases pin the duration: ScaleSpellEffect(1584, spllvl)
// (cpp:2556). The scale compounds — 1584, then 1584 + 198 = 1782, then
// 1782 + 222 = 2004 (ScaleSpellEffect iterates, missiles.cpp:119-125).
var infravisionAddCases = []infravisionAddCase{
	{name: "spell level 0 keeps the base", spellLevel: 0, wantDuration: 1584},
	{name: "spell level 2 adds a compounding eighth", spellLevel: 2, wantDuration: 2004},
}

// infravisionAddCase pins one AddInfravision row.
type infravisionAddCase struct {
	name         string
	spellLevel   int
	wantDuration int
}

// infravisionCase pins one ProcessInfravision tick.
type infravisionCase struct {
	name         string
	duration     int
	wantDuration int
	wantFlag     int
	wantRecalc   int
	wantDel      bool
}

// playerWrites records the player-system writes of one tick.
type playerWrites struct {
	flags   int
	recalcs int
}

// TestAddInfravision pins the spell-level scaled duration.
func TestAddInfravision(t *testing.T) {
	// given
	cases := infravisionAddCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			w := infravisionWorld(&playerWrites{})
			m := missiles.Missile{Type: missiles.Infravision, SpellLevel: tc.spellLevel}

			// when
			missiles.AddInfravision(w, &m, &missiles.AddParameter{})

			// then
			if m.Duration != tc.wantDuration { // cpp:2556
				t.Fatalf("Duration = %d, want %d", m.Duration, tc.wantDuration)
			}
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// TestProcessInfravision pins the flag write, the countdown and the
// delete + recompute tail.
func TestProcessInfravision(t *testing.T) {
	// given
	cases := infravisionCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			writes := &playerWrites{}
			w := infravisionWorld(writes)
			m := missiles.Missile{Type: missiles.Infravision, Duration: tc.duration, Source: 2}

			// when
			missiles.ProcessInfravision(w, &m)

			// then
			assertInfravisionTick(t, m, writes, tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// assertInfravisionTick checks the countdown, the delete flag and the
// player-system writes of one tick.
func assertInfravisionTick(t *testing.T, m missiles.Missile, writes *playerWrites, tc infravisionCase) {
	t.Helper()

	if m.Duration != tc.wantDuration || m.DelFlag != tc.wantDel {
		t.Fatalf("duration/del = %d/%v, want %d/%v",
			m.Duration, m.DelFlag, tc.wantDuration, tc.wantDel)
	}

	if writes.flags != tc.wantFlag || writes.recalcs != tc.wantRecalc {
		t.Fatalf("flag/recalc writes = %d/%d, want %d/%d",
			writes.flags, writes.recalcs, tc.wantFlag, tc.wantRecalc)
	}
}

// infravisionWorld wires the player-system sinks the row writes through.
func infravisionWorld(writes *playerWrites) *missiles.HitWorld {
	w := &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: missiles.NoopProcessCallbacks(),
	}
	w.ProcessCallbacks.SetInfravision = func(int) { writes.flags++ }
	w.ProcessCallbacks.RecalcPlayerItems = func(int) { writes.recalcs++ }

	return w
}
