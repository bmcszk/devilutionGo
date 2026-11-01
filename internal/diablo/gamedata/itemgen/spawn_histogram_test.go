package itemgen_test

import (
	"math"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
)

// spawnHistogramDraws is the AC's 10000-seed walk: seed i drives one
// SetupAllItems run, so the bucket counts are deterministic.
const spawnHistogramDraws = 10_000

// spawnHistogramTolerance is the absolute share band per bucket, 3σ of the
// binomial sample (sqrt(p(1-p)/n)) rounded up plus the measured GenerateRnd
// modulo bias (0.00043 on a 0.41 bucket, see histogram_test.go). The unique
// bucket is small enough that 3σ is 0.0015, so it carries its own tighter band.
var spawnHistogramTolerance = [3]float64{0.012, 0.012, 0.003}

// TestSpawnQualityHistogramSeeded pins the three quality buckets SetupAllItems
// can emit, against goldens hand-derived from the upstream draw order
// (items.cpp:3286-3296, 1345-1493) with a level-1 cleaver and uper 1:
//
//	GetItemBLevel   rnd(100)<=10 -> iblvl=lvl; else rnd(100)<=lvl (0.89*0.98 miss)
//	CheckUnique     rnd(100)<=1  -> unique, valid at iblvl 1 (0.02 of the hits)
//	affix coins     a prefix is always allocated, Bronze is the only eligible
//	                prefix at maxlvl 1, so every iblvl hit turns magical
func TestSpawnQualityHistogramSeeded(t *testing.T) {
	// given
	items, uniques := setupTables()
	prefixes, suffixes := lockAffixes()

	// when
	counts := spawnQualityCounts(items, uniques, prefixes, suffixes, 1)

	// then
	assertQualityShares(t, counts, [3]float64{0.8722, 0.125244, 0.002556})
}

// TestSpawnQualityHistogramUper15Seeded pins the uper15 level bump: iblvl is
// forced to lvl+4, so the early-exit miss cannot happen and the unique roll
// sees the full GenerateRnd(100) <= 15 window (items.cpp:1487-1489, 3289).
// The pools are crafted to be eligible at every level, so the affix coins
// always produce an affix and the only normal outcome would be a missed unique.
func TestSpawnQualityHistogramUper15Seeded(t *testing.T) {
	// given
	items, uniques := setupTables()
	prefixes, suffixes := alwaysEligibleAffixes()

	// when
	counts := spawnQualityCounts(items, uniques, prefixes, suffixes, 15)

	// then
	assertQualityShares(t, counts, [3]float64{0, 0.84, 0.16})
}

// TestSpawnQualityHistogramThreeClasses is the AC's three-class walk: the same
// 10000-seed SetupAllItems stream under the Warrior, Rogue and Sorcerer base
// pools. The only hero-dependent powers are IPL_MANATOLIFE / IPL_LIFETOMANA
// (items.cpp:1019-1028), which move player mana/life rather than change the
// item, so the quality buckets must match while the player deltas differ.
func TestSpawnQualityHistogramThreeClasses(t *testing.T) {
	// given
	classes := []struct {
		name string
		hero itemgen.HeroStats
	}{
		{name: "Warrior", hero: itemgen.HeroStats{MaxManaBase: 64, MaxHPBase: 230}},
		{name: "Rogue", hero: itemgen.HeroStats{MaxManaBase: 80, MaxHPBase: 190}},
		{name: "Sorcerer", hero: itemgen.HeroStats{MaxManaBase: 140, MaxHPBase: 140}},
	}

	// when
	buckets := make([][3]int, 0, len(classes))
	for _, class := range classes {
		buckets = append(buckets, spawnQualityCountsForHero(t, class.hero))
	}

	// then
	for i := range classes[1:] {
		if buckets[i+1] != buckets[0] {
			t.Fatalf("%s buckets = %v, want %v (itemgen never changes quality by class)",
				classes[i+1].name, buckets[i+1], buckets[0])
		}
	}
	if buckets[0] == [3]int{} {
		t.Fatalf("buckets = %v, want the seeded walk to produce items", buckets[0])
	}
}

func spawnQualityCounts(
	items []itemdat.ItemData,
	uniques []itemdat.UniqueItemData,
	prefixes, suffixes []itemgen.Affix,
	uper int,
) [3]int {
	counts := [3]int{}
	for i := range spawnHistogramDraws {
		req := itemgen.SetupRequest{ItemIndex: 1, Seed: uint32(i), Level: 1, Uper: uper}
		plan, err := itemgen.SetupAllItems(items, uniques, prefixes, suffixes, req, seededRnd(uint32(i)), func(int) {})
		if err != nil {
			panic(err)
		}
		counts[plan.Quality]++
	}

	return counts
}

func assertQualityShares(t *testing.T, counts [3]int, want [3]float64) {
	t.Helper()

	names := [3]string{"normal", "magic", "unique"}
	total := counts[0] + counts[1] + counts[2]
	if total != spawnHistogramDraws {
		t.Fatalf("total = %d, want %d", total, spawnHistogramDraws)
	}
	for i, count := range counts {
		share := float64(count) / spawnHistogramDraws
		delta := math.Abs(share - want[i])
		if delta > spawnHistogramTolerance[i] {
			t.Fatalf("%s share = %.5f, want %.5f +- %.4f (count %d)",
				names[i], share, want[i], spawnHistogramTolerance[i], count)
		}
	}
}

// alwaysEligibleAffixes is a one-prefix/one-suffix pool that survives every
// SelectAffix filter, so an allocated slot can never fall through to a plain
// item. MinLevel 2 puts both rows inside the uper15 window (minlvl
// min(iblvl/2,25) = 2 for the level-1 cleaver at iblvl 5, maxlvl 5), AlignAny
// skips the alignment lock and Useful skips the onlygood filter.
func alwaysEligibleAffixes() (prefixes, suffixes []itemgen.Affix) {
	prefixes = []itemgen.Affix{{
		Name: "Bronze", Power: "TOHIT", Value1: 1, Value2: 5, MinLevel: 2,
		Types: itemgen.AffixWeapon, Alignment: itemgen.AlignAny, Chance: 1, Useful: true,
	}}
	suffixes = []itemgen.Affix{{
		Name: "strength", Power: "STR", Value1: 1, Value2: 5, MinLevel: 2,
		Types: itemgen.AffixWeapon, Alignment: itemgen.AlignAny, Chance: 1, Useful: true,
	}}

	return prefixes, suffixes
}

func spawnQualityCountsForHero(t *testing.T, hero itemgen.HeroStats) [3]int {
	t.Helper()

	items, uniques := setupTables()
	prefixes, suffixes := lockAffixes()
	counts := [3]int{}
	for i := range spawnHistogramDraws {
		req := itemgen.SetupRequest{
			ItemIndex: 1, Seed: uint32(i), Level: 1, Uper: 1, Hero: hero,
		}
		plan, err := itemgen.SetupAllItems(items, uniques, prefixes, suffixes, req, seededRnd(uint32(i)), func(int) {})
		if err != nil {
			t.Fatalf("seed %d: %v", i, err)
		}
		counts[plan.Quality]++
	}

	return counts
}
