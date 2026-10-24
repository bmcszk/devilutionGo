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

func TestRandomizedSeedHarness(t *testing.T) {
	for _, tc := range buildSeedCases() {
		t.Run(tc.name, func(t *testing.T) {
			// given
			ref := refOf(tc.f)
			wantHit, wantDam := expect(tc.vals, tc.f)
			// when
			hit, dam64 := combat.PlrHitMonstCore(ref, tap(tc.vals))
			// then
			if hit != wantHit {
				t.Fatalf("hit=%v want %v", hit, wantHit)
			}
			if hit && dam64 != int64(wantDam) {
				t.Fatalf("dam64=%d want %d", dam64, wantDam)
			}
		})
	}
}

func buildSeedCases() []seedCase {
	rng := rand.New(rand.NewSource(20261018))
	classes := []combat.Class{combat.Warrior, combat.Rogue, combat.Sorcerer}
	mclasses := []combat.MonsterClass{combat.Undead, combat.Animal, combat.Demon}
	weapons := []combat.WeaponType{combat.WeaponNone, combat.Sword, combat.Mace}
	var cases []seedCase
	for seed := range 1000 {
		for _, cls := range classes {
			for _, mc := range mclasses {
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
				// both sides consume the same tape, so a mismatch is logic
				// drift, not RNG drift.
				vals := []int{rng.Intn(100), rng.Intn(100), rng.Intn(100)}
				cases = append(cases, seedCase{
					name: fmt.Sprintf("seed%d/%v/%v", seed, cls, mc),
					f:    f,
					vals: vals,
				})
			}
		}
	}
	return cases
}
