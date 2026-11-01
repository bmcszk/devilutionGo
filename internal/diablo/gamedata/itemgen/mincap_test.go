package itemgen_test

import (
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
)

// TestAffixWindowMinCap pins GetItemBonus' min(minlvl, 25) cap
// (items.cpp:1311): at iblvl 60 the window starts at 25, not 30, so a real
// table affix with minLevel 25 still applies.
func TestAffixWindowMinCap(t *testing.T) {
	// given
	items := []itemdat.ItemData{{ID: "SWORD", Name: "Sword", ItemType: "Sword", Durability: 100, Value: 1000}}
	suffixes := onlySuffix(t, "gore") // real row: minLevel 25, Weapon
	req := itemgen.SetupRequest{ItemIndex: 0, Seed: 5, Level: 60, Uper: 1, OnlyGood: true}
	rnd := &tapeRnd{t: t, values: []int32{1, 1, 1, 1, 0, 0, 0, 0, 0, 0}}

	// when
	plan, err := itemgen.SetupAllItems(items, nil, nil, suffixes, req, rnd.rnd, func(int) {})

	// then
	assertSuffixApplied(t, plan, err)
}

// onlySuffix keeps the named suffix from the committed table.
func onlySuffix(t *testing.T, name string) []itemgen.Affix {
	t.Helper()

	kept := make([]itemgen.Affix, 0, 1)
	for _, affix := range realSuffixes(t) {
		if affix.Name == name {
			kept = append(kept, affix)
		}
	}
	if len(kept) != 1 {
		t.Fatalf("suffix %q rows = %d, want 1", name, len(kept))
	}

	return kept
}

func assertSuffixApplied(t *testing.T, plan itemgen.SetupPlan, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("SetupAllItems: %v", err)
	}
	if plan.Quality != itemgen.QualityMagic || !strings.Contains(plan.Name, "of gore") {
		t.Fatalf("plan = %+v, want the minLevel-25 suffix inside a capped window", plan)
	}
}
