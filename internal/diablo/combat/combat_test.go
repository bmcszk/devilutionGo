// Package combat_test ports the PlrHitMonst roll pipeline
// (Source/player.cpp:526-660) as a roll-tape parity harness: every
// GenerateRnd/RandomIntBetween call is fed from a scripted tape, so the
// exact upstream roll ORDER is pinned. 3 classes x 3 monster classes x
// scripted tapes per devilGo-7be9's AC, plus clamp and petrified rules.
package combat_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/combat"
)

type fixture struct {
	cls         combat.Class
	lvl         int
	adjacent    bool
	petrified   bool
	toHitBonus  int
	armorPierce int
	minD, maxD  int
	weapon      combat.WeaponType
	mClass      combat.MonsterClass
	tripleDmg   bool
}

// tap replays a scripted tape exactly like GenerateRnd consumes the RNG.
func tap(vals []int) func(int) int {
	i := 0
	return func(int) int {
		v := vals[i%len(vals)]
		i++
		return v
	}
}

// applyClassModifiers mirrors the monsterClass switch in the trace.
func applyClassModifiers(dam int, f fixture) int {
	switch f.mClass {
	case combat.Undead:
		switch f.weapon {
		case combat.Sword:
			return dam - dam/2
		case combat.Mace:
			return dam + dam/2
		default:
			return dam
		}
	case combat.Animal:
		switch f.weapon {
		case combat.Mace:
			return dam - dam/2
		case combat.Sword:
			return dam + dam/2
		default:
			return dam
		}
	default:
		if f.mClass == combat.Demon && f.tripleDmg {
			return dam * 3
		}
		return dam
	}
}

// expect traces the C++ body (player.cpp:527-630) for fixture+tape.
func expect(vals []int, f fixture) (bool, int) {
	rnd := tap(vals)
	hper := 0
	if f.adjacent {
		if f.lvl > 20 {
			hper -= 30
		} else {
			hper -= (35 - f.lvl) * 2
		}
	}
	hit := rnd(100)
	if f.petrified {
		hit = 0
	}
	hper += f.toHitBonus - f.armorPierce
	if hper < 5 {
		hper = 5
	}
	if hper > 95 {
		hper = 95
	}
	if hit >= hper {
		return false, 0
	}
	// damage roll: RandomIntBetween(minD, maxD) = GenerateRnd(max-min+1)+min
	dam := rnd(f.maxD-f.minD+1) + f.minD
	// critical strike (player.cpp:574-578): only the flagged class draws
	if f.cls == combat.Warrior {
		if rnd(100) < f.lvl {
			dam *= 2
		}
	}
	dam = applyClassModifiers(dam, f)
	if f.adjacent {
		dam >>= 2
	}
	return true, dam
}

// refOf builds the port's input struct from the fixture.
func refOf(f fixture) combat.Ref {
	return combat.Ref{
		Class: f.cls, Level: f.lvl, Adjacent: f.adjacent, Petrified: f.petrified,
		ToHitBonus: f.toHitBonus, ArmorPierce: f.armorPierce,
		MinDam: f.minD, MaxDam: f.maxD, Weapon: f.weapon,
		MClass: f.mClass, TripleDemon: f.tripleDmg,
	}
}

func runCase(t *testing.T, name string, tv []int, f fixture) {
	t.Helper()
	ref := refOf(f)
	hit, dam64 := combat.PlrHitMonstCore(ref, tap(tv))
	wantHit, wantDam := expect(tv, f)
	if hit != wantHit {
		t.Errorf("%s: hit=%v want %v", name, hit, wantHit)
	}
	if hit && dam64 != int64(wantDam)<<6 {
		t.Errorf("%s: dam64=%d want %d (1/64 units)", name, dam64, int64(wantDam)<<6)
	}
	if hit && dam64%64 != 0 {
		t.Errorf("%s: dam64=%d is not a whole 1/64 word", name, dam64)
	}
}

// TestRollTapeParity: the AC matrix — 3 classes x 3 monster classes x
// scripted tapes.
func TestRollTapeParity(t *testing.T) {
	tapes := [][]int{
		{50, 50, 50}, // everything mid-range, repeated
		{0, 7, 99},   // hit always lands, low damage roll, no crit
		{94, 6, 3},   // boundary hit (94 < 95-clamped hper), crit lands
		{95, 2, 2},   // hit misses at the clamp boundary
		{5, 8, 12},   // minimal hit that lands, crit lands at level 12
	}
	classes := []struct {
		name string
		flag combat.Class
	}{
		{"warrior", combat.Warrior},
		{"rogue", combat.Rogue},
		{"sorcerer", combat.Sorcerer},
	}
	mclasses := []struct {
		name string
		m    combat.MonsterClass
	}{
		{"undead", combat.Undead},
		{"animal", combat.Animal},
		{"demon", combat.Demon},
	}
	base := fixture{
		lvl: 12, toHitBonus: 70, armorPierce: 10,
		minD: 2, maxD: 8, weapon: combat.Sword,
	}
	for _, cls := range classes {
		for _, mc := range mclasses {
			for ti, tv := range tapes {
				f := base
				f.cls = cls.flag
				f.mClass = mc.m
				runCase(t, cls.name+"/"+mc.name+"/tape"+string(rune('0'+ti)), tv, f)
			}
		}
	}
}

// TestAdjacentPenalty: ranged-hitting while adjacent (-toHit, dam>>2).
func TestAdjacentPenalty(t *testing.T) {
	f := fixture{
		cls: combat.Sorcerer, lvl: 10, adjacent: true, toHitBonus: 80, armorPierce: 0,
		minD: 4, maxD: 12, weapon: combat.Mace, mClass: combat.Animal,
	}
	runCase(t, "adjacent-animal-mace", []int{30, 8}, f)
	f2 := f
	f2.lvl = 25 // >20 branch: flat -30
	runCase(t, "adjacent-lvl25", []int{30, 8}, f2)
}

// TestPetrifiedAlwaysHits: petrified forces hit=0 which always lands.
func TestPetrifiedAlwaysHits(t *testing.T) {
	f := fixture{
		cls: combat.Sorcerer, lvl: 5, petrified: true, toHitBonus: 10, armorPierce: 0,
		minD: 1, maxD: 3, weapon: combat.Sword, mClass: combat.Undead,
	}
	runCase(t, "petrified", []int{99, 2}, f)
}

// TestClampBounds: hper clamps 5..95 on both ends.
func TestClampBounds(t *testing.T) {
	f := fixture{
		cls: combat.Sorcerer, lvl: 1, toHitBonus: 300, armorPierce: 0, // would be 300 -> clamp 95
		minD: 1, maxD: 2, weapon: combat.Sword, mClass: combat.Demon,
	}
	runCase(t, "clamp-high", []int{94, 1}, f)
	runCase(t, "clamp-high-miss", []int{95, 1}, f)
	f2 := f
	f2.toHitBonus = -500 // clamp to 5
	runCase(t, "clamp-low", []int{5, 1}, f2)
	runCase(t, "clamp-low-miss", []int{6, 1}, f2)
}

// TestCriticalStrikeDamageWord: the warrior's class flag adds a roll
// after the damage modifiers and before the weapon ones
// (player.cpp:574-578) — the tape's third value is that roll, and it
// doubles the damage when it lands below the character level.
func TestCriticalStrikeDamageWord(t *testing.T) {
	base := fixture{
		cls: combat.Warrior, lvl: 12, toHitBonus: 70, armorPierce: 10,
		minD: 2, maxD: 8, weapon: combat.Sword, mClass: combat.Undead,
	}
	// tape {hit 50, damage 6 (+min 2 = 8), crit 3}: the crit lands
	// (3 < level 12) → 8*2 = 16, sword vs undead halves it to 8 → 512.
	hit, dam64 := combat.PlrHitMonstCore(refOf(base), tap([]int{50, 6, 3}))
	if !hit {
		t.Fatal("warrior-crit: no hit")
	}
	if dam64 != 512 {
		t.Fatalf("warrior-crit: dam64 = %d, want 512", dam64)
	}

	// tape {hit 50, damage 6, crit 99}: the crit misses (99 > level 12)
	// → 8, sword vs undead halves it to 4 → 256.
	hit, dam64 = combat.PlrHitMonstCore(refOf(base), tap([]int{50, 6, 99}))
	if !hit {
		t.Fatal("warrior-no-crit: no hit")
	}
	if dam64 != 256 {
		t.Fatalf("warrior-no-crit: dam64 = %d, want 256", dam64)
	}

	// The rogue has no CriticalStrike flag (playerdat attributes.tsv),
	// so it draws no third value and keeps the halved damage → 256.
	rogue := base
	rogue.cls = combat.Rogue
	hit, dam64 = combat.PlrHitMonstCore(refOf(rogue), tap([]int{50, 6, 3}))
	if !hit {
		t.Fatal("rogue: no hit")
	}
	if dam64 != 256 {
		t.Fatalf("rogue: dam64 = %d, want 256", dam64)
	}
}
