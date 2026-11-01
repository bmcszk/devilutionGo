package itemgen_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
)

// drawCount is the sample size of the tolerance-based distribution checks.
const drawCount = 12000

// TestSelectAffixWeightFrequenciesNondeterministic checks that PLChance
// weights shape the pick distribution. Tolerance-based: no exact sequence or
// seed-to-output mapping is asserted, so any correct RNG passes.
func TestSelectAffixWeightFrequenciesNondeterministic(t *testing.T) {
	// given
	affixes := []itemgen.Affix{
		{Name: "rarest", Types: itemgen.AffixWeapon, MinLevel: 1, Chance: 1},
		{Name: "common", Types: itemgen.AffixWeapon, MinLevel: 1, Chance: 3},
		{Name: "ubiquitous", Types: itemgen.AffixWeapon, MinLevel: 1, Chance: 6},
	}
	rnd := seededRnd(4711)

	// when
	counts := pickCounts(t, affixes, rnd)

	// then
	assertShare(t, counts, "rarest", 0.10, 0.025)
	assertShare(t, counts, "common", 0.30, 0.025)
	assertShare(t, counts, "ubiquitous", 0.60, 0.025)
}

// TestRndItemForMonsterLevelDropSharesNondeterministic checks the drop-tree
// odds: >40 means no drop (59%), >25 means gold (41%*74%), else an item
// (41%*25%). Tolerance-based, same rules as above.
func TestRndItemForMonsterLevelDropSharesNondeterministic(t *testing.T) {
	// given
	// Row 0 is the gold sentinel with a zero drop rate, so the pool can never
	// return it and the three branches stay distinguishable.
	items := testItems(3)
	items[0].DropRate = 0
	env := itemgen.DropEnv{Hellfire: true}
	rnd := seededRnd(20177)

	// when
	none, gold, item := dropShares(t, items, env, rnd)

	// then
	assertRate(t, "no drop", none, 0.59, 0.03)
	assertRate(t, "gold", gold, 0.3034, 0.03)
	assertRate(t, "item", item, 0.1025, 0.03)
}

func pickCounts(t *testing.T, affixes []itemgen.Affix, rnd func(int32) int32) map[string]int {
	t.Helper()

	counts := make(map[string]int, len(affixes))
	for range drawCount {
		pick, ok := itemgen.SelectAffix(affixes, itemgen.AffixWeapon, 1, 6, false, itemgen.AlignAny, false, rnd)
		if !ok {
			t.Fatal("SelectAffix: empty pool inside the distribution check")
		}
		counts[pick.Name]++
	}

	return counts
}

func assertShare(t *testing.T, counts map[string]int, name string, want, tolerance float64) {
	t.Helper()

	share := float64(counts[name]) / float64(drawCount)
	if share < want-tolerance || share > want+tolerance {
		t.Fatalf("%s share = %.3f, want %.3f ±%.3f (counts %v)", name, share, want, tolerance, counts)
	}
}

func dropShares(
	t *testing.T,
	items []itemdat.ItemData,
	env itemgen.DropEnv,
	rnd func(int32) int32,
) (none, gold, item int) {
	t.Helper()

	for range drawCount {
		idx, err := itemgen.RndItemForMonsterLevel(items, 5, env, rnd)
		if err != nil {
			t.Fatalf("RndItemForMonsterLevel: %v", err)
		}
		switch idx {
		case itemgen.ItemNone:
			none++
		case 0:
			gold++
		default:
			item++
		}
	}

	return none, gold, item
}

func assertRate(t *testing.T, name string, count int, want, tolerance float64) {
	t.Helper()

	if rate := float64(count) / float64(drawCount); rate < want-tolerance || rate > want+tolerance {
		t.Fatalf("%s rate = %.4f, want %.4f ±%.3f", name, rate, want, tolerance)
	}
}
