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

const (
	outcomeMiss hitOutcome = false
	outcomeHit  hitOutcome = true
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

// hitOutcome names the expected hit result of a runCase row — a domain
// answer, not a control flag (player.cpp:527-630 hit/miss outcomes).
type hitOutcome bool

func TestRollTapeParity(t *testing.T) {
	// given
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
	// then — base hper = 0+70-10 = 60 (player.cpp:547-548, no adjacent
	// penalty); hit roll is tape[0] (player.cpp:542); damage roll
	// rnd(7)+2 (player.cpp:567); crit roll tape[2] only for the warrior
	// (player.cpp:574-578); undead sword halves, animal sword adds half
	// (player.cpp:589-602); dam<<=6 (player.cpp:618).
	for _, cls := range classes {
		for _, mc := range mclasses {
			for ti, tv := range tapes {
				f := base
				f.cls = cls.flag
				f.mClass = mc.m
				// wantDam per tape: t0 dam=52 → undead 26/animal 78/demon 52;
				// t1 dam=9 → 5/13/9; t2,t3 miss; t4 dam=10 → 5/15/10.
				// The warrior's crit roll never lands (<12 fails), so the
				// literals are identical for all three classes.
				wantHit, wantDam := rollTapeWant(ti, mc.m)
				runCase(t, cls.name+"/"+mc.name+"/tape"+string(rune('0'+ti)), tv, f, wantHit, wantDam)
			}
		}
	}
}

func TestAdjacentPenalty(t *testing.T) {
	// given
	f := fixture{
		cls: combat.Sorcerer, lvl: 10, adjacent: true, toHitBonus: 80, armorPierce: 0,
		minD: 4, maxD: 12, weapon: combat.Mace, mClass: combat.Animal,
	}
	runCase(t, "adjacent-animal-mace", []int{30, 8}, f, outcomeMiss, 0) // hit roll 30 ≥ hper 30 (player.cpp:553-558)
	f2 := f
	f2.lvl = 25 // >20 branch: flat -30 (player.cpp:527-533)
	// hper = -30+80 = 50; dam = 8+4 = 12, animal mace halves → 6, <<6 =
	// 384, adjacent >>=2 → 96 (player.cpp:567,589-602,618,626-627).
	runCase(t, "adjacent-lvl25", []int{30, 8}, f2, outcomeHit, 96)
}

func TestPetrifiedAlwaysHits(t *testing.T) {
	// given
	f := fixture{
		cls: combat.Sorcerer, lvl: 5, petrified: true, toHitBonus: 10, armorPierce: 0,
		minD: 1, maxD: 3, weapon: combat.Sword, mClass: combat.Undead,
	}
	// hper = 10, petrified forces hit=0 → always lands (player.cpp:543-545);
	// dam = 2+1 = 3, undead sword halves → 2, <<6 → 128 (player.cpp:567,589-590,618).
	runCase(t, "petrified", []int{99, 2}, f, outcomeHit, 128)
}

func TestClampBounds(t *testing.T) {
	// given
	f := fixture{
		cls: combat.Sorcerer, lvl: 1, toHitBonus: 300, armorPierce: 0, // would be 300 -> clamp 95
		minD: 1, maxD: 2, weapon: combat.Sword, mClass: combat.Demon,
	}
	// hper clamps to 95 (player.cpp:551-552): hit 94 lands; dam = 1+1 =
	// 2 → 128 (player.cpp:567,618).
	runCase(t, "clamp-high", []int{94, 1}, f, outcomeHit, 128)
	runCase(t, "clamp-high-miss", []int{95, 1}, f, outcomeMiss, 0) // 95 ≥ 95 (player.cpp:553-558)
	f2 := f
	f2.toHitBonus = -500                                     // clamp to 5 (player.cpp:549-550)
	runCase(t, "clamp-low", []int{5, 1}, f2, outcomeMiss, 0) // 5 ≥ 5
	runCase(t, "clamp-low-miss", []int{6, 1}, f2, outcomeMiss, 0)
}

// TestCriticalStrikeDamageWord: the warrior's class flag adds a roll
// after the damage modifiers and before the weapon ones
// (player.cpp:574-578) — the tape's third value is that roll, and it
// doubles the damage when it lands below the character level.
func TestCriticalStrikeDamageWord(t *testing.T) {
	// given
	base := fixture{
		cls: combat.Warrior, lvl: 12, toHitBonus: 70, armorPierce: 10,
		minD: 2, maxD: 8, weapon: combat.Sword, mClass: combat.Undead,
	}
	// when
	// tape {hit 50, damage 6 (+min 2 = 8), crit 3}: the crit lands
	// (3 < level 12) → 8*2 = 16, sword vs undead halves it to 8 → 512.
	hit, dam64 := combat.PlrHitMonstCore(refOf(base), tap([]int{50, 6, 3}))
	// then
	if !hit {
		t.Fatal("warrior-crit: no hit")
	}

	if dam64 != 512 {
		t.Fatalf("warrior-crit: dam64 = %d, want 512", dam64)
	}

	// when
	// tape {hit 50, damage 6, crit 99}: the crit misses (99 > level 12)
	// → 8, sword vs undead halves it to 4 → 256.
	hit, dam64 = combat.PlrHitMonstCore(refOf(base), tap([]int{50, 6, 99}))
	// then
	if !hit {
		t.Fatal("warrior-no-crit: no hit")
	}

	if dam64 != 256 {
		t.Fatalf("warrior-no-crit: dam64 = %d, want 256", dam64)
	}

	// when
	// The rogue has no CriticalStrike flag (playerdat attributes.tsv),
	// so it draws no third value and keeps the halved damage → 256.
	rogue := base
	rogue.cls = combat.Rogue
	hit, dam64 = combat.PlrHitMonstCore(refOf(rogue), tap([]int{50, 6, 3}))
	// then
	if !hit {
		t.Fatal("rogue: no hit")
	}

	if dam64 != 256 {
		t.Fatalf("rogue: dam64 = %d, want 256", dam64)
	}
}

// rollTapeWant returns the expected outcome and 1/64-unit damage for tape
// index ti against monster class mc (player.cpp:553-558 hit; :567/:589-602/
// :618 damage); tapes 2,3 miss at the clamp boundary.
func rollTapeWant(ti int, mc combat.MonsterClass) (hitOutcome, int64) {
	damByTape := map[int]map[combat.MonsterClass]int64{
		0: {combat.Undead: 1664, combat.Animal: 4992, combat.Demon: 3328},
		1: {combat.Undead: 320, combat.Animal: 832, combat.Demon: 576},
		4: {combat.Undead: 320, combat.Animal: 960, combat.Demon: 640},
	}
	if dm, ok := damByTape[ti]; ok {
		return outcomeHit, dm[mc]
	}

	return outcomeMiss, 0 // tapes 2,3 miss; damage not asserted
}

func tap(vals []int) func(int) int {
	i := 0

	return func(int) int {
		v := vals[i%len(vals)]
		i++

		return v
	}
}

// weaponMod applies the undead/animal weapon modifiers
// (player.cpp:589-602); other combos are unchanged.
func weaponMod(dam int, mClass combat.MonsterClass, weapon combat.WeaponType) int {
	switch mClass {
	case combat.Undead:
		switch weapon {
		case combat.Sword:
			return dam - dam/2
		case combat.Mace:
			return dam + dam/2
		default:
			return dam
		}
	case combat.Animal:
		switch weapon {
		case combat.Mace:
			return dam - dam/2
		case combat.Sword:
			return dam + dam/2
		default:
			return dam
		}
	default:
		return dam
	}
}

// expectHper derives the effective to-hit chance (player.cpp:527-548):
// adjacent penalty, bonus/pierce, clamped to [5,95].
func expectHper(f fixture) int {
	hper := 0

	if f.adjacent { // player.cpp:527-533
		if f.lvl > 20 {
			hper -= 30
		} else {
			hper -= (35 - f.lvl) * 2
		}
	}

	hper += f.toHitBonus - f.armorPierce // player.cpp:547-548
	if hper < 5 {
		hper = 5
	}

	if hper > 95 {
		hper = 95
	}

	return hper
}

// expect is the reference oracle consumed ONLY by
// TestRandomizedSeedHarness (combat_fuzz_test.go, outside this file's edit
// scope). Every table case in this file pins literal want values instead;
// expect re-derives the upstream arithmetic (player.cpp:527-630) so the
// randomized ladder can differential-check the port against it.
// expectHitRoll rolls the hit die and applies the petrified override;
// returns whether the blow connects (player.cpp:542-558).
func expectHitRoll(rnd func(int) int, f fixture, hper int) bool {
	hit := rnd(100)  // player.cpp:542
	if f.petrified { // player.cpp:543-545
		hit = 0
	}

	return hit < hper // player.cpp:553-558
}

func expect(vals []int, f fixture) (bool, int64) {
	rnd := tap(vals)

	hper := expectHper(f)
	if !expectHitRoll(rnd, f, hper) { // player.cpp:542-558
		return false, 0
	}
	// RandomIntBetween(minD, maxD) = GenerateRnd(max-min+1)+min (player.cpp:567)
	dam := rnd(f.maxD-f.minD+1) + f.minD
	// critical strike: flagged class only (player.cpp:574-578)
	if f.cls == combat.Warrior {
		if rnd(100) < f.lvl {
			dam *= 2
		}
	}
	// weapon modifiers (player.cpp:589-602)
	dam = weaponMod(dam, f.mClass, f.weapon)
	if f.mClass == combat.Demon && f.tripleDmg { // player.cpp:603-606
		dam *= 3
	}

	dam <<= 6       // player.cpp:618
	if f.adjacent { // player.cpp:626-627
		dam >>= 2
	}

	return true, int64(dam)
}

// runCase drives the port with the scripted tape and pins the outcome
// to literal want values derived by hand from player.cpp:527-630 —
// the test never re-derives the arithmetic.
func runCase(t *testing.T, name string, tv []int, f fixture, expectHit hitOutcome, wantDam int64) {
	t.Helper()
	// when
	hit, dam64 := combat.PlrHitMonstCore(refOf(f), tap(tv))
	// then
	if hit != bool(expectHit) {
		t.Errorf("%s: hit=%v want %v", name, hit, bool(expectHit))
	}

	if hit && dam64 != wantDam {
		t.Errorf("%s: dam64=%d want %d (1/64 units)", name, dam64, wantDam)
	}
}

func refOf(f fixture) combat.Ref {
	return combat.Ref{
		Class: f.cls, Level: f.lvl, Adjacent: f.adjacent, Petrified: f.petrified,
		ToHitBonus: f.toHitBonus, ArmorPierce: f.armorPierce,
		MinDam: f.minD, MaxDam: f.maxD, Weapon: f.weapon,
		MClass: f.mClass, TripleDemon: f.tripleDmg,
	}
}
