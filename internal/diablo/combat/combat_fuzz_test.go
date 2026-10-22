package combat_test

import (
	"math/rand"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/combat"
)

// TestRandomizedSeedHarness catches roll-order drift the scripted tapes miss.
func TestRandomizedSeedHarness(t *testing.T) {
	rng := rand.New(rand.NewSource(20261018))
	classes := []combat.Class{combat.Warrior, combat.Rogue, combat.Sorcerer}
	mclasses := []combat.MonsterClass{combat.Undead, combat.Animal, combat.Demon}
	weapons := []combat.WeaponType{combat.WeaponNone, combat.Sword, combat.Mace}
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
				ref := refOf(f)
				// both sides consume the same tape, so a mismatch is logic
				// drift, not RNG drift.
				vals := []int{rng.Intn(100), rng.Intn(100), rng.Intn(100)}
				// given
				rnd := tap(vals)
				hit, dam64 := combat.PlrHitMonstCore(ref, rnd)
				wantHit, wantDam := expect(vals, f)
				// then
				if hit != wantHit {
					t.Fatalf("seed %d cls %v mc %v: hit=%v want %v", seed, cls, mc, hit, wantHit)
				}
				if hit && dam64 != int64(wantDam) {
					t.Fatalf("seed %d cls %v mc %v: dam64=%d want %d", seed, cls, mc, dam64, wantDam)
				}
			}
		}
	}
}
