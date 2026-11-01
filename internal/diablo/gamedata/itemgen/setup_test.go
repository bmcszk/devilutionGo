package itemgen_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
)

func TestSetupAllItemsFlagsAndQuality(t *testing.T) {
	// given
	items, uniques := setupTables()
	req := itemgen.SetupRequest{ItemIndex: 0, Seed: 1, Level: 40, Uper: 1, Pregen: true}
	rnd := &tapeRnd{t: t, values: []int32{0, 0, 0, 0, 0, 0}}
	discarded := 0

	// when
	plan, err := itemgen.SetupAllItems(items, uniques, nil, nil, req, rnd.rnd, func(n int) { discarded += n })
	// then
	if err != nil {
		t.Fatalf("SetupAllItems: %v", err)
	}
	assertSetupFlags(t, plan, req)
	if discarded != 0 {
		t.Fatalf("discarded = %d, want 0 (empty unique list skips the burn)", discarded)
	}
}

func TestSetupAllItemsUniqueBranch(t *testing.T) {
	// given
	items, uniques := setupTables()
	req := itemgen.SetupRequest{ItemIndex: 1, Seed: 7, Level: 10, Uper: 15}
	rnd := &tapeRnd{t: t, values: []int32{0, 1, 0}}
	discarded := 0

	// when
	plan, err := itemgen.SetupAllItems(items, uniques, nil, nil, req, rnd.rnd, func(n int) { discarded += n })
	// then
	if err != nil {
		t.Fatalf("SetupAllItems: %v", err)
	}
	if plan.Quality != itemgen.QualityUnique || plan.UID != 0 {
		t.Fatalf("plan = %+v, want the unique row 0 applied", plan)
	}
	assertUniqueApplied(t, plan)
	if plan.Durability != 120 || discarded != 1 {
		t.Fatalf("durability/burn = %d/%d, want 120 and 1 (uniques skip ItemRndDur)", plan.Durability, discarded)
	}
	if plan.CreateInfo&itemgen.CreateInfoUnique == 0 {
		t.Fatalf("createInfo = %#x, want CF_UNIQUE", plan.CreateInfo)
	}
}

func TestSetupAllItemsForcedNonUnique(t *testing.T) {
	// given
	items, uniques := setupTables()
	req := itemgen.SetupRequest{ItemIndex: 1, Seed: 7, Level: 10, Uper: 15, ForceNotUnique: true}
	rnd := &tapeRnd{t: t, values: []int32{0, 0, 0, 0, 0, 0}}
	discarded := 0

	// when
	plan, err := itemgen.SetupAllItems(items, uniques, nil, nil, req, rnd.rnd, func(n int) { discarded += n })
	// then
	if err != nil {
		t.Fatalf("SetupAllItems: %v", err)
	}
	if plan.Quality == itemgen.QualityUnique || plan.UID != itemgen.UniqueNone {
		t.Fatalf("plan = %+v, want the forced non-unique path", plan)
	}
	if discarded != 1 {
		t.Fatalf("discarded = %d, want 1 (CheckUnique discards once when forced)", discarded)
	}
}

func TestSetupAllItemsUniqueMiscSeed(t *testing.T) {
	// given
	items, uniques := setupTables()
	req := itemgen.SetupRequest{ItemIndex: 2, Seed: 1, Level: 10}
	rnd := &tapeRnd{t: t, values: []int32{0, 0, 0, 0}}

	// when
	plan, err := itemgen.SetupAllItems(items, uniques, nil, nil, req, rnd.rnd, func(int) {})
	// then
	if err != nil {
		t.Fatalf("SetupAllItems: %v", err)
	}
	if plan.Quality != itemgen.QualityUnique || plan.UID != 1 || plan.Name != "Ring of Engagement" {
		t.Fatalf("plan = %+v, want uid 1 from the seed", plan)
	}
}

func TestSetupAllItemsUniqueMiscMismatch(t *testing.T) {
	// given
	items, uniques := setupTables()
	req := itemgen.SetupRequest{ItemIndex: 2, Seed: 0, Level: 10}
	rnd := &tapeRnd{t: t}

	// when
	plan, err := itemgen.SetupAllItems(items, uniques, nil, nil, req, rnd.rnd, func(int) {})
	// then
	if err != nil {
		t.Fatalf("SetupAllItems: %v", err)
	}
	if !plan.Cleared || len(rnd.args) != 0 {
		t.Fatalf("plan = %+v draws = %v, want the seed/base mismatch to clear the item", plan, rnd.args)
	}
}

func setupTables() ([]itemdat.ItemData, []itemdat.UniqueItemData) {
	items := []itemdat.ItemData{
		{ID: "SWORD", Name: "Sword", ItemType: "Sword", Durability: 100, Value: 1000},
		{ID: "CLEAVER", Name: "Cleaver", ItemType: "Sword", Durability: 120, Value: 800},
		{ID: "RING", Name: "Ring", ItemType: "Misc", MiscID: "UNIQUE", Durability: 0, Value: 500},
	}
	uniques := []itemdat.UniqueItemData{
		{
			Name: "The Butcher's Cleaver", UniqueBaseItem: "CLEAVER", MinLevel: 1, Value: 2500, Power0: "TOHIT",
			Power0Value1: 26, Power0Value2: 26,
		},
		{Name: "Ring of Engagement", UniqueBaseItem: "RING", MinLevel: 1, Value: 1500},
	}

	return items, uniques
}

func assertSetupFlags(t *testing.T, plan itemgen.SetupPlan, req itemgen.SetupRequest) {
	t.Helper()

	wantInfo := req.Level | itemgen.CreateInfoPregen | itemgen.CreateInfoUper1
	if plan.CreateInfo != wantInfo {
		t.Fatalf("createInfo = %#x, want %#x", plan.CreateInfo, wantInfo)
	}
	if plan.Quality != itemgen.QualityNormal || plan.Iblvl != 40 {
		t.Fatalf("plan = %+v, want a normal item at iblvl 40", plan)
	}
	if plan.Durability != 26 {
		t.Fatalf("durability = %d, want 26 (rnd 0 + maxDur/4 + 1)", plan.Durability)
	}
}

func assertUniqueApplied(t *testing.T, plan itemgen.SetupPlan) {
	t.Helper()

	if plan.Name != "The Butcher's Cleaver" || len(plan.Powers) != 1 {
		t.Fatalf("name/powers = %q/%+v, want the unique name and one power", plan.Name, plan.Powers)
	}
}
