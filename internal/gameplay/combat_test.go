// Package gameplay_test ports the PlrHitMonst roll pipeline
// (Source/player.cpp:526-660) as a roll-tape parity harness: every
// GenerateRnd/RandomIntBetween call is fed from a scripted tape, so the
// exact upstream roll ORDER is pinned. 3 classes x 3 monster classes x
// scripted tapes per devilGo-7be9's AC, plus clamp and petrified rules.
package gameplay_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/gameplay"
)

type fixture struct {
	lvl         int
	adjacent    bool
	petrified   bool
	toHitBonus  int
	armorPierce int
	minD, maxD  int
	weapon      gameplay.WeaponType
	mClass      gameplay.MonsterClass
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
	case gameplay.Undead:
		switch f.weapon {
		case gameplay.Sword:
			return dam - dam/2
		case gameplay.Mace:
			return dam + dam/2
		default:
			return dam
		}
	case gameplay.Animal:
		switch f.weapon {
		case gameplay.Mace:
			return dam - dam/2
		case gameplay.Sword:
			return dam + dam/2
		default:
			return dam
		}
	default:
		if f.mClass == gameplay.Demon && f.tripleDmg {
			return dam * 3
		}
		return dam
	}
}

// expect traces the C++ body (player.cpp:526-613) for fixture+tape.
func expect(vals []int, f fixture) (bool, int) {
	i := 0
	rnd := func(int) int { v := vals[i%len(vals)]; i++; return v }
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
	dam = applyClassModifiers(dam, f)
	if f.adjacent {
		dam >>= 2
	}
	return true, dam
}

func runCase(t *testing.T, name string, tv []int, f fixture) {
	t.Helper()
	ref := gameplay.Ref{
		Level: f.lvl, Adjacent: f.adjacent, Petrified: f.petrified,
		ToHitBonus: f.toHitBonus, ArmorPierce: f.armorPierce,
		MinDam: f.minD, MaxDam: f.maxD, Weapon: f.weapon,
		MClass: f.mClass, TripleDemon: f.tripleDmg,
	}
	hit, dam64 := gameplay.PlrHitMonstCore(ref, tap(tv))
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
		{50},       // everything mid-range, repeated
		{0, 7},     // hit always lands, low damage roll
		{94, 6, 3}, // boundary hit (94 < 95-clamped hper)
		{95, 2},    // hit misses at the clamp boundary
		{5, 8},     // minimal hit that lands
	}
	classes := []struct {
		name string
		flag gameplay.ClassFlag
	}{
		{"warrior", gameplay.Warrior},
		{"rogue", gameplay.Rogue},
		{"sorcerer", gameplay.Sorcerer},
	}
	mclasses := []struct {
		name string
		m    gameplay.MonsterClass
	}{
		{"undead", gameplay.Undead},
		{"animal", gameplay.Animal},
		{"demon", gameplay.Demon},
	}
	base := fixture{
		lvl: 12, toHitBonus: 70, armorPierce: 10,
		minD: 2, maxD: 8, weapon: gameplay.Sword,
	}
	for _, cls := range classes {
		for _, mc := range mclasses {
			for ti, tv := range tapes {
				f := base
				f.mClass = mc.m
				runCase(t, cls.name+"/"+mc.name+"/tape"+string(rune('0'+ti)), tv, f)
			}
		}
	}
}

// TestAdjacentPenalty: ranged-hitting while adjacent (-toHit, dam>>2).
func TestAdjacentPenalty(t *testing.T) {
	f := fixture{
		lvl: 10, adjacent: true, toHitBonus: 80, armorPierce: 0,
		minD: 4, maxD: 12, weapon: gameplay.Mace, mClass: gameplay.Animal,
	}
	runCase(t, "adjacent-animal-mace", []int{30, 8}, f)
	f2 := f
	f2.lvl = 25 // >20 branch: flat -30
	runCase(t, "adjacent-lvl25", []int{30, 8}, f2)
}

// TestPetrifiedAlwaysHits: petrified forces hit=0 which always lands.
func TestPetrifiedAlwaysHits(t *testing.T) {
	f := fixture{
		lvl: 5, petrified: true, toHitBonus: 10, armorPierce: 0,
		minD: 1, maxD: 3, weapon: gameplay.Sword, mClass: gameplay.Undead,
	}
	runCase(t, "petrified", []int{99, 2}, f)
}

// TestClampBounds: hper clamps 5..95 on both ends.
func TestClampBounds(t *testing.T) {
	f := fixture{
		lvl: 1, toHitBonus: 300, armorPierce: 0, // would be 300 -> clamp 95
		minD: 1, maxD: 2, weapon: gameplay.Sword, mClass: gameplay.Demon,
	}
	runCase(t, "clamp-high", []int{94, 1}, f)
	runCase(t, "clamp-high-miss", []int{95, 1}, f)
	f2 := f
	f2.toHitBonus = -500 // clamp to 5
	runCase(t, "clamp-low", []int{5, 1}, f2)
	runCase(t, "clamp-low-miss", []int{6, 1}, f2)
}
