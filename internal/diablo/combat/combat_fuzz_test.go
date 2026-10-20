// Package combat_test: randomized 1000-seed harness for the AC
// "3 classes x 5 monsters x 1000 seeds" — core vs independent trace
// with random inputs, catching any roll-ORDER drift the scripted tapes
// miss.
package combat_test

import (
	"math/rand"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/combat"
)

// TestRandomizedSeedHarness: 1000 seeds; each seed runs one random
// fixture per (class, monster-class) pair; core and trace must agree on
// verdict AND damage word.
func TestRandomizedSeedHarness(t *testing.T) {
	rng := rand.New(rand.NewSource(20261018))
	classes := []combat.Class{combat.Warrior, combat.Rogue, combat.Sorcerer}
	mclasses := []combat.MonsterClass{combat.Undead, combat.Animal, combat.Demon}
	weapons := []combat.WeaponType{combat.WeaponNone, combat.Sword, combat.Mace}
	for seed := 0; seed < 1000; seed++ {
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
				ref := refOf(f)
				// shared scripted tape: both sides see the SAME rolls; the
				// trace below consumes them in the same upstream order.
				vals := []int{rng.Intn(100), rng.Intn(100), rng.Intn(100)}
				rnd := tap(vals)
				hit, dam64 := combat.PlrHitMonstCore(ref, rnd)
				wantHit, wantDam := expect(vals, f)
				if hit != wantHit {
					t.Fatalf("seed %d cls %v mc %v: hit=%v want %v", seed, cls, mc, hit, wantHit)
				}
				if hit && dam64 != int64(wantDam)<<6 {
					t.Fatalf("seed %d cls %v mc %v: dam64=%d want %d", seed, cls, mc, dam64, int64(wantDam)<<6)
				}
			}
		}
	}
}
