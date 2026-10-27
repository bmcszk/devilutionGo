package combat_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/combat"
)

// seedCase is one point of the randomized-seed coverage space, materialized
// into a named table row so a failure names its exact deterministic inputs.
type seedCase struct {
	name string
	f    fixture
	vals []int
}

// seedResult carries one randomized row's scripted-tape output alongside
// its oracle expectation, so a failure still names its exact inputs.
type seedResult struct {
	name    string
	want    hitOutcome
	wantDam int64
	got     hitOutcome
	dam64   int64
}

func TestRandomizedSeedHarness(t *testing.T) {
	// given
	cases := buildSeedCases()
	// when
	results := make([]seedResult, 0, len(cases))

	for _, tc := range cases {
		wantHit, wantDam := expect(tc.vals, tc.f)
		hit, dam64 := combat.PlrHitMonstCore(refOf(tc.f), tap(tc.vals))
		results = append(results, seedResult{
			name:    tc.name,
			want:    hitOutcome(wantHit),
			wantDam: wantDam,
			got:     hitOutcome(hit),
			dam64:   dam64,
		})
	}
	// then
	for _, r := range results {
		t.Run(r.name, func(t *testing.T) {
			assertSeedOutcome(t, r.name, r.want, r.wantDam, r.got, r.dam64)
		})
	}
}

func buildSeedCases() []seedCase {
	rng := rand.New(rand.NewSource(20261018))
	classes := []combat.Class{combat.Warrior, combat.Rogue, combat.Sorcerer}
	mclasses := []combat.MonsterClass{combat.Undead, combat.Animal, combat.Demon}
	weapons := []combat.WeaponType{combat.WeaponNone, combat.Sword, combat.Mace}

	cases := make([]seedCase, 0, len(classes)*len(mclasses)*1000)

	for seed := range 1000 {
		for _, cls := range classes {
			for _, mc := range mclasses {
				cases = append(cases, oneSeedCase(rng, seed, cls, mc, weapons))
			}
		}
	}

	return cases
}

// oneSeedCase rolls one randomized fixture row; both sides consume the same
// tape, so a mismatch is logic drift, not RNG drift.
func oneSeedCase(rng *rand.Rand, seed int, cls combat.Class, mc combat.MonsterClass,
	weapons []combat.WeaponType,
) seedCase {
	f := fixture{
		cls:         cls,
		lvl:         1 + rng.Intn(30),
		adjacent:    rng.Intn(4) == 0,
		petrified:   rng.Intn(8) == 0,
		toHitBonus:  rng.Intn(200) - 40,
		armorPierce: rng.Intn(30),
		minD:        1 + rng.Intn(4),
		maxD:        5 + rng.Intn(20),
		weapon:      weapons[rng.Intn(len(weapons))],
		mClass:      mc,
		tripleDmg:   mc == combat.Demon && rng.Intn(2) == 0,
	}
	if f.maxD < f.minD {
		f.maxD = f.minD
	}

	vals := []int{rng.Intn(100), rng.Intn(100), rng.Intn(100)}

	return seedCase{
		name: fmt.Sprintf("seed%d/%v/%v", seed, cls, mc),
		f:    f,
		vals: vals,
	}
}

// assertSeedOutcome pins the randomized-harness hit flag and damage.
func assertSeedOutcome(t *testing.T, name string, want hitOutcome, wantDam int64, got hitOutcome, dam64 int64,
) {
	t.Helper()

	if got != want {
		t.Fatalf("%s: hit=%v want %v", name, bool(got), bool(want))
	}

	if want == outcomeHit && dam64 != wantDam {
		t.Fatalf("%s: dam64=%d want %d", name, dam64, wantDam)
	}
}
