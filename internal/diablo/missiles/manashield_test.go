// Package missiles_test pins AddManaShield (missiles.cpp:2770-2782), the
// add-only row whose corpus carries no process body (misdat.tsv:13).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// manaShieldCase pins one AddManaShield call.
type manaShieldCase struct {
	name        string
	hasShield   bool
	wantWrites  int
	wantFizzled bool
}

// TestAddManaShield pins the three outcomes: the shield is raised and the
// missile dropped, an already-shielded caster fizzles the spell, and nothing is
// written on the fizzle path.
func TestAddManaShield(t *testing.T) {
	// given
	cases := []manaShieldCase{
		{name: "raises the shield", wantWrites: 1},
		{name: "already shielded fizzles", hasShield: true, wantFizzled: true},
	}

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			raised := 0
			w := manaShieldWorld(tc.hasShield, &raised)
			m := missiles.Missile{Type: missiles.ManaShield, Source: 0}
			p := &missiles.AddParameter{}

			// when
			missiles.AddManaShield(w, &m, p)

			// then
			assertManaShieldAdd(t, m, p, raised, tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// assertManaShieldAdd checks the shield write, the fizzle flag and the dropped
// missile.
func assertManaShieldAdd(t *testing.T, m missiles.Missile, p *missiles.AddParameter,
	raised int, tc manaShieldCase,
) {
	t.Helper()

	if raised != tc.wantWrites { // cpp:2773-2779
		t.Fatalf("shield writes = %d, want %d", raised, tc.wantWrites)
	}

	if p.SpellFizzled != tc.wantFizzled { // cpp:2774
		t.Fatalf("SpellFizzled = %v, want %v", p.SpellFizzled, tc.wantFizzled)
	}

	if !m.DelFlag { // cpp:2772
		t.Fatal("DelFlag = false, want the missile dropped")
	}
}

// manaShieldWorld wires the shield read and the shield write.
func manaShieldWorld(hasShield bool, raised *int) *missiles.HitWorld {
	cb := missiles.NoopProcessCallbacks()
	cb.PlayerManaShield = func(int) bool { return hasShield }
	cb.SetManaShield = func(int) { *raised++ }

	return &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: cb,
	}
}
