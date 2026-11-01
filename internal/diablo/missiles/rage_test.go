// Package missiles_test pins AddRage (missiles.cpp:2879-2897) and ProcessRage
// (cpp:3625-3660).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// rageCases pin the fizzle guards, the damage and duration formula
// (cpp:2886-2887) and the two-phase process: the active flag turns into the
// cooldown on the first expiry and the missile dies on the second
// (cpp:3632-3641).
var rageCases = []rageCase{
	{
		name: "active flag fizzles the cast", level: 4, spellFlags: missiles.SpellFlagRageActive,
		wantDel: true, wantFizzled: true,
	},
	{
		name: "cooldown flag fizzles the cast", level: 4, spellFlags: missiles.SpellFlagRageCooldown,
		wantDel: true, wantFizzled: true,
	},
	{
		name: "low life fizzles the cast", level: 4, hitPoints: 4 << 6,
		wantDel: true, wantFizzled: true,
	},
	{
		name: "cast arms rage", level: 4, hitPoints: 40 << 6,
		wantDam: 24, wantDuration: 253, wantFlags: missiles.SpellFlagRageActive,
	},
}

// rageCase pins one AddRage call.
type rageCase struct {
	name         string
	level        int
	spellFlags   int
	hitPoints    int
	wantDel      bool
	wantFizzled  bool
	wantDam      int
	wantDuration int
	wantFlags    int
}

// rageTickCase pins one ProcessRage expiry tick.
type rageTickCase struct {
	name       string
	spellFlags int
	wantFlags  int
	wantDel    bool
	wantDamage int
	wantDurSet int
}

// rageSpies records the player writes of one rage call.
type rageSpies struct {
	flags     int
	hitPoints int
	damage    int
	speeches  int
	recalcs   int
	redraws   int
}

// TestAddRage pins the guards, the damage/duration formula and the flag write.
func TestAddRage(t *testing.T) {
	// given
	cases := rageCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			spies := &rageSpies{flags: tc.spellFlags}
			w, _ := rageWorld(spies, tc)
			m := missiles.Missile{Type: missiles.Rage, Source: 0}
			p := &missiles.AddParameter{}

			// when
			missiles.AddRage(w, &m, p)

			// then
			assertRageAdd(t, m, p, spies, tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// TestProcessRage pins the active-to-cooldown handover, the cooldown expiry
// delete and the life floor (cpp:3644-3658).
func TestProcessRage(t *testing.T) {
	// given
	cases := []rageTickCase{
		{
			name: "active turns into cooldown", spellFlags: missiles.SpellFlagRageActive,
			wantFlags: missiles.SpellFlagRageCooldown, wantDurSet: 253,
		},
		{
			name: "cooldown ends the rage", spellFlags: missiles.SpellFlagRageCooldown,
			wantFlags: 0, wantDel: true, wantDamage: 24 << 6,
		}, // cpp:3658 passes Fixed26_6
	}

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			w, spies := rageWorld(&rageSpies{flags: tc.spellFlags}, rageCase{})
			m := missiles.Missile{Type: missiles.Rage, Source: 0, Dam: 24, Var1: 253, Duration: 1}

			// when
			missiles.ProcessRage(w, &m)

			// then
			assertRageTick(t, m, spies, tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// rageWorld wires the flag cell, the life pool, the recalcs and the speech.
func rageWorld(spies *rageSpies, tc rageCase) (*missiles.HitWorld, *rageSpies) {
	cb := missiles.NoopProcessCallbacks()
	cb.PlayerSpellFlags = func(int) int { return spies.flags }
	cb.SetPlayerSpellFlags = func(_, flags int) { spies.flags = flags }
	cb.SetPlayerHitPoints = func(_, hitPoints int) { spies.hitPoints = hitPoints }
	cb.DamagePlayer = func(_, _, damage, _ int) { spies.damage += damage }
	cb.Say = func(int, int) { spies.speeches++ }
	cb.RecalcPlayerItems = func(int) { spies.recalcs++ }
	cb.RedrawEverything = func() { spies.redraws++ }

	return &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: cb,
		AddSource: missiles.AddSource{
			PlayerStats: func(int) missiles.AddPlayerStats {
				return missiles.AddPlayerStats{Level: tc.level, HitPoints: tc.hitPoints}
			},
		},
	}, spies
}

// assertRageAdd checks the add body outcome.
func assertRageAdd(t *testing.T, m missiles.Missile, p *missiles.AddParameter, spies *rageSpies, tc rageCase) {
	t.Helper()

	if m.DelFlag != tc.wantDel || p.SpellFizzled != tc.wantFizzled { // cpp:2882-2884
		t.Fatalf("del/fizzled = %v/%v, want %v/%v", m.DelFlag, p.SpellFizzled, tc.wantDel, tc.wantFizzled)
	}

	if tc.wantDel {
		return
	}

	if spies.flags != tc.wantFlags { // cpp:2892
		t.Fatalf("spell flags = %d, want %d", spies.flags, tc.wantFlags)
	}

	if spies.recalcs != 1 || spies.speeches != 1 { // cpp:2893-2894
		t.Fatalf("recalcs/speeches = %d/%d, want 1/1", spies.recalcs, spies.speeches)
	}

	assertRageArmed(t, m, tc)
}

// assertRageArmed checks the damage and the timer the cast arms (cpp:2886-2888).
func assertRageArmed(t *testing.T, m missiles.Missile, tc rageCase) {
	t.Helper()

	if m.Dam != tc.wantDam || m.Duration != tc.wantDuration || m.Var1 != tc.wantDuration {
		t.Fatalf("dam/dur/var1 = %d/%d/%d, want %d/%d/%d",
			m.Dam, m.Duration, m.Var1, tc.wantDam, tc.wantDuration, tc.wantDuration)
	}
}

// assertRageTick checks one ProcessRage expiry.
func assertRageTick(t *testing.T, m missiles.Missile, spies *rageSpies, tc rageTickCase) {
	t.Helper()

	if spies.flags != tc.wantFlags || m.DelFlag != tc.wantDel { // cpp:3632-3641
		t.Fatalf("flags/del = %d/%v, want %d/%v", spies.flags, m.DelFlag, tc.wantFlags, tc.wantDel)
	}

	if tc.wantDurSet != 0 && m.Duration != tc.wantDurSet { // cpp:3637
		t.Fatalf("duration = %d, want %d", m.Duration, tc.wantDurSet)
	}

	if spies.damage != tc.wantDamage { // cpp:3658
		t.Fatalf("damage = %d, want %d", spies.damage, tc.wantDamage)
	}

	if spies.redraws != 1 || spies.speeches != 1 { // cpp:3653-3654
		t.Fatalf("redraws/speeches = %d/%d, want 1/1", spies.redraws, spies.speeches)
	}
}
