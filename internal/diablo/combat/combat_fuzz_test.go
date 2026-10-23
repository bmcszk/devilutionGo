// TODO remove nolint:testpackage // internal-state test: pokes unexported identifiers
package combat

import (
	"fmt"
	"math/rand"
	"testing"
)

// seedCase is one point of the randomized-seed coverage space, materialized
// into a named table row so a failure names its exact deterministic inputs.
type seedCase struct {
	name string
	f    fixture
	vals []int
}

// TestRandomizedSeedHarness catches roll-order drift the scripted tapes miss.
// Chose a table of pre-materialized deterministic cases over F.Fuzz: the
// coverage space is a fixed 1000-seed ladder shared with the reference
// implementation, not an open integer domain, and named rows make each
// failure reproducible without seed archaeology.
func TestRandomizedSeedHarness(t *testing.T) {
	for _, tc := range buildSeedCases() {
		t.Run(tc.name, func(t *testing.T) {
			// given
			ref := refOf(tc.f)
			wantHit, wantDam := expect(tc.vals, tc.f)
			// when
			hit, dam64 := plrHitMonstCore(ref, tap(tc.vals))
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

// buildSeedCases walks the same deterministic rand ladder as before (one
// shared rand.New source, fixed seed), collecting every (seed, class, mclass)
// combination with its fixture and roll tape into table rows.
func buildSeedCases() []seedCase {
	rng := rand.New(rand.NewSource(20261018))
	classes := []Class{Warrior, Rogue, Sorcerer}
	mclasses := []monsterClass{Undead, Animal, Demon}
	weapons := []weaponType{WeaponNone, Sword, Mace}
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
					tripleDmg:   mc == Demon && rng.Intn(2) == 0,
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
