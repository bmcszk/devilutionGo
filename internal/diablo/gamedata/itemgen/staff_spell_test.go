package itemgen_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
	"github.com/bmcszk/devilutionGo/internal/diablo/txtdata"
)

func TestSetupAllItemsStaffSpellBranch(t *testing.T) {
	// given
	items := []itemdat.ItemData{{
		ID: "STAFF", Name: "Staff", ItemType: "Staff", Durability: 20, Value: 1000, MinMagic: 0,
	}}
	spells := []txtdata.SpellData{
		{Name: "Null"},
		{Name: "Firebolt", StaffLevel: 1, MinInt: 15, StaffMin: 40, StaffMax: 80, StaffCost10: 5},
	}
	req := itemgen.SetupRequest{
		ItemIndex: 0, Seed: 3, Level: 10, Uper: 1,
		Spells: spells, Env: itemgen.SpellEnv{Hellfire: true},
	}
	// draws: GetItemBLevel, CheckUnique, spell pick (0 -> Firebolt),
	// charges (20 of 41), prefix coin (1 -> no prefix)
	rnd := &tapeRnd{t: t, values: []int32{0, 0, 0, 20, 1, 0}}

	// when
	plan, err := itemgen.SetupAllItems(items, nil, nil, nil, req, rnd.rnd, func(int) {})

	// then
	assertStaffSpell(t, plan, err)
}

func TestSetupAllItemsStaffSpellPrefix(t *testing.T) {
	// given
	items := []itemdat.ItemData{{
		ID: "STAFF", Name: "Staff", ItemType: "Staff", Durability: 20, Value: 1000,
	}}
	prefixes := []itemgen.Affix{
		{Name: "of Skill", Power: "TOHIT", Types: itemgen.AffixStaff, MinLevel: 1, Chance: 1},
	}
	spells := []txtdata.SpellData{
		{Name: "Null"},
		{Name: "Firebolt", StaffLevel: 1, MinInt: 15, StaffMin: 40, StaffMax: 40, StaffCost10: 5},
	}
	req := itemgen.SetupRequest{
		ItemIndex: 0, Seed: 3, Level: 10, Uper: 1,
		Spells: spells, Env: itemgen.SpellEnv{Hellfire: true},
	}
	// prefix coin passes (0), then the prefix value roll and the durability roll
	rnd := &tapeRnd{t: t, values: []int32{0, 0, 0, 0, 0, 0, 0, 0}}

	// when
	plan, err := itemgen.SetupAllItems(items, nil, prefixes, nil, req, rnd.rnd, func(int) {})
	// then
	if err != nil {
		t.Fatalf("SetupAllItems: %v", err)
	}
	if plan.Quality != itemgen.QualityMagic {
		t.Fatalf("plan = %+v, want the magic quality GetStaffPower sets on a prefix", plan)
	}
	if plan.Spell != 1 || plan.Charges != 40 {
		t.Fatalf("plan = %+v, want the spell to survive the prefix roll", plan)
	}
}

func TestSetupAllItemsStaffWithoutSpellTable(t *testing.T) {
	// given
	items := []itemdat.ItemData{{
		ID: "STAFF", Name: "Staff", ItemType: "Staff", Durability: 20, Value: 1000,
	}}
	req := itemgen.SetupRequest{ItemIndex: 0, Seed: 3, Level: 10, Uper: 1}
	rnd := &tapeRnd{t: t, values: []int32{0, 0, 0, 0, 0, 0}}

	// when
	plan, err := itemgen.SetupAllItems(items, nil, nil, nil, req, rnd.rnd, func(int) {})
	// then
	if err != nil {
		t.Fatalf("SetupAllItems: %v", err)
	}
	if plan.Spell != 0 || plan.Charges != 0 {
		t.Fatalf("plan = %+v, want the affix path when no spell table is supplied", plan)
	}
}

func assertStaffSpell(t *testing.T, plan itemgen.SetupPlan, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("SetupAllItems: %v", err)
	}
	if plan.Spell != 1 {
		t.Fatalf("spell = %d, want Firebolt", plan.Spell)
	}
	if plan.Charges != 60 || plan.MinMagic != 15 {
		t.Fatalf("plan = %+v, want 60 charges and minInt 15", plan)
	}
	if want := 1000 + 60*5*10/5; plan.Value != want {
		t.Fatalf("value = %d, want %d (base + charges*staffCost()/5)", plan.Value, want)
	}
}
