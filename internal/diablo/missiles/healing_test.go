// Package missiles_test pins AddHealing (missiles.cpp:2860-2878), an add-only
// row whose corpus process column is empty (misdat.tsv:37).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// healingCases pin the rolled heal: GenerateRnd(10) + 1, GenerateRndSum(4,
// level) + level, GenerateRndSum(6, spell level) + spell level (cpp:2864-2866),
// all scaled by the class restore factor (cpp:2869-2870) and clamped to the
// maxima (cpp:2871-2872).
var healingCases = []healingCase{
	{
		name: "hits clamp to the maximum", level: 2, spellLevel: 1,
		rnd:       []int{3, 1, 2, 4},
		hitPoints: 50 << 6, maxHitPoints: 60 << 6, hpBase: 40 << 6, maxHPBase: 60 << 6,
		restore: 64, wantHitPoints: 60 << 6, wantHPBase: 54 << 6,
	},
	{
		name: "half restore heals half", level: 0, spellLevel: 0,
		rnd:       []int{0, 0},
		hitPoints: 10 << 6, maxHitPoints: 40 << 6, hpBase: 20 << 6, maxHPBase: 40 << 6,
		restore: 32, wantHitPoints: (10 << 6) + 32, wantHPBase: (20 << 6) + 32,
	},
}

// healingCase pins one AddHealing call; the life fields are Fixed26_6 raw
// values (player.h:236-237).
type healingCase struct {
	name          string
	level         int
	spellLevel    int
	rnd           []int
	hitPoints     int
	maxHitPoints  int
	hpBase        int
	maxHPBase     int
	restore       int
	wantHitPoints int
	wantHPBase    int
}

// healingSpies records the life writes of one heal.
type healingSpies struct {
	hitPoints int
	hpBase    int
	writes    int
	redraws   int
}

// TestAddHealing pins the heal roll, the class scaling, the clamps and the
// dropped missile.
func TestAddHealing(t *testing.T) {
	// given
	cases := healingCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			spies := &healingSpies{}
			w := healingWorld(spies, &rndTape{t: t, tape: tc.rnd}, tc)
			m := missiles.Missile{Type: missiles.Healing, SpellLevel: tc.spellLevel, Source: 0}

			// when
			missiles.AddHealing(w, &m, &missiles.AddParameter{})

			// then
			assertHealing(t, m, spies, tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// healingWorld wires the tape, the life read/write, the class factor and the
// panel redraw.
func healingWorld(spies *healingSpies, tape *rndTape, tc healingCase) *missiles.HitWorld {
	cb := missiles.NoopProcessCallbacks()
	cb.GenerateRnd = tape.draw
	cb.ClassLifeRestore = func(int) int { return tc.restore }
	cb.PlayerLife = func(int) missiles.PlayerLife {
		return missiles.PlayerLife{
			HitPoints: tc.hitPoints, MaxHitPoints: tc.maxHitPoints,
			HPBase: tc.hpBase, MaxHPBase: tc.maxHPBase,
		}
	}
	cb.SetPlayerLife = func(_, hitPoints, hpBase int) {
		spies.writes++
		spies.hitPoints = hitPoints
		spies.hpBase = hpBase
	}
	cb.RedrawHealth = func() { spies.redraws++ }

	return &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: cb,
		AddSource: missiles.AddSource{
			PlayerStats: func(int) missiles.AddPlayerStats {
				return missiles.AddPlayerStats{Level: tc.level}
			},
		},
	}
}

// assertHealing checks the clamped life write, the redraw and the dropped
// missile.
func assertHealing(t *testing.T, m missiles.Missile, spies *healingSpies, tc healingCase) {
	t.Helper()

	if spies.writes != 1 || spies.hitPoints != tc.wantHitPoints || spies.hpBase != tc.wantHPBase {
		t.Fatalf("life write = %d (%d/%d), want 1 (%d/%d)",
			spies.writes, spies.hitPoints, spies.hpBase, tc.wantHitPoints, tc.wantHPBase)
	}

	if spies.redraws != 1 { // cpp:2877
		t.Fatalf("panel redraws = %d, want 1", spies.redraws)
	}

	if !m.DelFlag { // cpp:2875
		t.Fatal("DelFlag = false, want the missile dropped")
	}
}
