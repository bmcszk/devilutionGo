package itemgen_test

import (
	"slices"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
)

// TestV16WorkedExamplePool pins the duplicate-weight expansion of the doc's
// example: two Chance-2 prefixes plus two Chance-1 suffixes give a 6-slot
// vector, and index 0 is Tin.
func TestV16WorkedExamplePool(t *testing.T) {
	// given
	prefixes, suffixes := v16Prefixes(), v16Suffixes()
	pool := slices.Concat(prefixes, suffixes)
	rnd := &tapeRnd{t: t, values: []int32{0}}

	// when
	pick, ok := itemgen.SelectAffix(pool, itemgen.AffixWeapon, 1, 6, false, itemgen.AlignAny, false, rnd.rnd)

	// then
	if !ok || pick.Name != "Tin" {
		t.Fatalf("pick = %q ok = %v, want Tin", pick.Name, ok)
	}
	if rnd.args[0] != 6 {
		t.Fatalf("pool size = %d, want 6 (Tin x2, Brass x2, sharpness, mending)", rnd.args[0])
	}
	if pick.Power != "TOHIT_CURSE" || pick.Value1 != 6 || pick.Value2 != 10 || pick.MultVal != -3 {
		t.Fatalf("pick = %+v, want TOHIT_CURSE 6/10 mult -3", pick)
	}
	if pick.Alignment != itemgen.AlignEvil {
		t.Fatalf("alignment = %v, want Evil", pick.Alignment)
	}
}

// TestV16WorkedExampleAfflock pins the prefix->goe ordering: once Tin (Evil)
// is drawn, the suffix pool loses every Good suffix and the out-of-window
// mayhem, leaving sharpness alone.
func TestV16WorkedExampleAfflock(t *testing.T) {
	// given
	suffixes := v16Suffixes()
	rnd := &tapeRnd{t: t, values: []int32{0}}

	// when
	pick, ok := itemgen.SelectAffix(suffixes, itemgen.AffixWeapon, 1, 6, false, itemgen.AlignEvil, false, rnd.rnd)

	// then
	if !ok || pick.Name != "sharpness" {
		t.Fatalf("pick = %q ok = %v, want sharpness only", pick.Name, ok)
	}
	if rnd.args[0] != 1 {
		t.Fatalf("pool size = %d, want 1 (mending Good and mayhem lvl 7 excluded)", rnd.args[0])
	}
}

// TestV16WorkedExampleValue follows the Tin draw into SaveItemAffix: zero
// minVal/maxVal keeps Add1 at 0 and puts multVal -3 in the first multiplier.
func TestV16WorkedExampleValue(t *testing.T) {
	// given
	tin := v16Prefixes()[0]
	rnd := &tapeRnd{t: t, values: []int32{0}}
	values := itemgen.AffixValues{}

	// when
	bonus, err := itemgen.SaveItemAffix(&values, tin, rnd.rnd)
	// then
	if err != nil {
		t.Fatalf("SaveItemAffix: %v", err)
	}
	if bonus.Value != 6 || values.Add1 != 0 || values.Mult1 != -3 {
		t.Fatalf("bonus = %+v values = %+v, want RndPL(6,10)=6 with Add1 0 / Mult1 -3", bonus, values)
	}
	// v = 1000 / -3 = -333, then max(v, 1) clamps the cursed item to 1
	// (items.cpp:606-609).
	if got := itemgen.CalcItemValue(values, 1000); got != 1 {
		t.Fatalf("value = %d, want 1", got)
	}
}

// v16Prefixes is the requirements doc's worked-example pool
// (docs/research/dg-requirements-append.md V1.6): Tin x2 and Brass x2. Both
// rows are Evil-aligned, which is why the worked example's goe lock (the
// prefix's alignment carries into the suffix pick) leaves the pool empty.
func v16Prefixes() []itemgen.Affix {
	return []itemgen.Affix{
		{
			Name: "Tin", Power: "TOHIT_CURSE", Value1: 6, Value2: 10, MinLevel: 3,
			Types: itemgen.AffixWeapon, Alignment: itemgen.AlignEvil, Chance: 2, MultVal: -3,
		},
		{
			Name: "Brass", Power: "TOHIT_CURSE", Value1: 1, Value2: 5, MinLevel: 1,
			Types: itemgen.AffixWeapon, Alignment: itemgen.AlignEvil, Chance: 2, MultVal: -2,
		},
	}
}

// v16Suffixes holds the doc's suffix set: sharpness (Any), mending (Good) and
// mayhem (level 7, so the minlvl>maxlvl filter drops it).
func v16Suffixes() []itemgen.Affix {
	return []itemgen.Affix{
		{Name: "sharpness", Power: "TOHIT", MinLevel: 1, Types: itemgen.AffixWeapon, Chance: 1},
		{
			Name: "mending", Power: "DUR", MinLevel: 1, Types: itemgen.AffixWeapon,
			Alignment: itemgen.AlignGood, Chance: 1,
		},
		{
			Name: "mayhem", Power: "DAMMOD", Value1: 9, Value2: 13, MinLevel: 7,
			Types: itemgen.AffixWeapon, Chance: 1,
		},
	}
}
