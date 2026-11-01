package itemgen_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
	"github.com/bmcszk/devilutionGo/internal/diablo/testsupport"
)

// expansionTuples is the AC's sampled-tuple count.
const expansionTuples = 1000

// poolCap mirrors maxEligibleAffixes, the StaticVector capacity
// (items.cpp:1085-1087); the package keeps it unexported.
const poolCap = 256

// expTuple is one sampled (type, minlvl, maxlvl, onlygood, goe) combination.
type expTuple struct {
	typ      itemgen.AffixItemType
	minlvl   int
	maxlvl   int
	onlygood bool
	goe      itemgen.GoodOrEvil
	staff    bool
}

// TestEligibleAffixesExpansionInvariant samples expansionTuples combinations
// over the committed prefix and suffix tables and checks the
// duplicate-expansion invariant against the ported pool builder: the pool
// length is sum(PLChance) clamped at 256.
//
// The AC also asks for a cross-check against a C++ probe histogram, which
// needs an offline-produced committed fixture; none exists under
// internal/testdata, so that half stays open and this invariant is what the
// repo can check today.
func TestEligibleAffixesExpansionInvariant(t *testing.T) {
	// given
	prefixes, suffixes := realPrefixes(t), realSuffixes(t)
	tables := []struct {
		name    string
		affixes []itemgen.Affix
	}{{"prefixes", prefixes}, {"suffixes", suffixes}}

	// when
	failures := 0
	for i := range expansionTuples {
		tc := expTupleFor(i)
		for _, table := range tables {
			pool := itemgen.EligibleAffixes(table.affixes, tc.typ, tc.minlvl, tc.maxlvl, tc.onlygood, tc.goe, tc.staff)
			if want := expansionCeiling(table.affixes, tc); len(pool) != want {
				t.Errorf("%s tuple %d %+v: pool = %d, want %d", table.name, i, tc, len(pool), want)
				failures++
			}
		}
	}

	// then
	if failures != 0 {
		t.Fatalf("%d expansion mismatches over %d tuples", failures, expansionTuples)
	}
}

// expTupleFor derives one tuple arithmetically so the sample is exact and
// reproducible.
func expTupleFor(i int) expTuple {
	types := [...]itemgen.AffixItemType{
		itemgen.AffixWeapon, itemgen.AffixBow, itemgen.AffixShield,
		itemgen.AffixArmor, itemgen.AffixStaff, itemgen.AffixMisc,
	}

	return expTuple{
		typ:      types[i%len(types)],
		minlvl:   i % 26,
		maxlvl:   i % 61,
		onlygood: i%2 == 0,
		goe:      itemgen.GoodOrEvil(i % 3),
		staff:    i%4 == 0,
	}
}

// expansionCeiling is sum(PLChance) over the rows that pass the filters, with
// the 256-slot cap applied.
func expansionCeiling(affixes []itemgen.Affix, tc expTuple) int {
	total := 0
	for _, affix := range affixes {
		if expansionEligible(affix, tc) {
			total += min(affix.Chance, poolCap)
		}
	}

	return min(total, poolCap)
}

// expansionEligible restates the filter set of items.cpp:1074-1082.
func expansionEligible(affix itemgen.Affix, tc expTuple) bool {
	if affix.Types&tc.typ == 0 || affix.MinLevel > tc.maxlvl || affix.MinLevel < tc.minlvl {
		return false
	}
	if tc.onlygood && !affix.Useful {
		return false
	}
	if tc.staff && affix.Power == "CHARGES" {
		return false
	}

	return !alignmentClash(affix.Alignment, tc.goe)
}

// alignmentClash is ChoosePrefixSuffix' goe exclusion (:1080).
func alignmentClash(alignment, goe itemgen.GoodOrEvil) bool {
	return (goe == itemgen.AlignGood && alignment == itemgen.AlignEvil) ||
		(goe == itemgen.AlignEvil && alignment == itemgen.AlignGood)
}

func realSuffixes(t *testing.T) []itemgen.Affix {
	t.Helper()

	rows, err := itemdat.LoadItemSuffixes(testsupport.Corpus(t, "items", "assets/txtdata/items/item_suffixes.tsv"))
	if err != nil {
		t.Fatalf("LoadItemSuffixes: %v", err)
	}
	affixes, err := itemgen.SuffixesToAffixes(rows)
	if err != nil {
		t.Fatalf("SuffixesToAffixes: %v", err)
	}

	return affixes
}
