package itemgen_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
	"github.com/bmcszk/devilutionGo/internal/diablo/txtdata"
)

// TestSetupAllItemsIdentifiedValueMagic pins the price calc split: _ivalue keeps
// the base item value while _iIvalue takes the CalcItemValue result once an
// affix landed (items.cpp:1226-1227, 606-609).
func TestSetupAllItemsIdentifiedValueMagic(t *testing.T) {
	// given
	items, uniques := setupTables()
	prefixes, _ := lockAffixes()
	req := itemgen.SetupRequest{ItemIndex: 0, Seed: 3, Level: 1, Uper: 1}
	rnd := &tapeRnd{t: t, values: []int32{0, 0, 0, 0, 0, 0, 0, 0}}

	// when
	plan, err := itemgen.SetupAllItems(items, uniques, prefixes, nil, req, rnd.rnd, func(int) {})
	// then
	if err != nil {
		t.Fatalf("SetupAllItems: %v", err)
	}
	if plan.Quality != itemgen.QualityMagic {
		t.Fatalf("plan = %+v, want the Bronze prefix to make the item magic", plan)
	}
	if want := itemgen.CalcItemValue(plan.Values, items[0].Value); plan.Ivalue != want {
		t.Fatalf("Ivalue = %d, want %d (CalcItemValue)", plan.Ivalue, want)
	}
	if plan.Value != items[0].Value {
		t.Fatalf("Value = %d, want %d (_ivalue keeps the base value)", plan.Value, items[0].Value)
	}
}

// TestSetupAllItemsIdentifiedValueUnique pins the unique override: GetUniqueItem
// writes the unique table value into _iIvalue only (items.cpp:1465).
func TestSetupAllItemsIdentifiedValueUnique(t *testing.T) {
	// given
	items, uniques := setupTables()
	req := itemgen.SetupRequest{ItemIndex: 1, Seed: 1, Level: 10, Uper: 15}
	rnd := &tapeRnd{t: t, values: []int32{0, 0, 0, 0, 0, 0}}

	// when
	plan, err := itemgen.SetupAllItems(items, uniques, nil, nil, req, rnd.rnd, func(int) {})
	// then
	if err != nil {
		t.Fatalf("SetupAllItems: %v", err)
	}
	if plan.Quality != itemgen.QualityUnique {
		t.Fatalf("plan = %+v, want the unique path", plan)
	}
	if plan.Ivalue != uniques[0].Value {
		t.Fatalf("Ivalue = %d, want %d (UIValue)", plan.Ivalue, uniques[0].Value)
	}
	if plan.Value != items[1].Value {
		t.Fatalf("Value = %d, want %d (_ivalue keeps the base value)", plan.Value, items[1].Value)
	}
}

// TestSetupAllItemsIdentifiedValueStaffShare pins the staff spell tail: the
// share lands on both slots (items.cpp:1271-1275) and GetStaffPower then
// re-prices _iIvalue unconditionally (:1153), so an affix-free staff ends at 1.
func TestSetupAllItemsIdentifiedValueStaffShare(t *testing.T) {
	// given
	items := ivalueStaffItems()
	req := itemgen.SetupRequest{
		ItemIndex: 0, Seed: 3, Level: 10, Uper: 1,
		Spells: ivalueStaffSpells(), Env: itemgen.SpellEnv{Hellfire: true},
	}
	rnd := &tapeRnd{t: t, values: []int32{0, 0, 0, 20, 1, 0}}

	// when
	plan, err := itemgen.SetupAllItems(items, nil, nil, nil, req, rnd.rnd, func(int) {})
	// then
	if err != nil {
		t.Fatalf("SetupAllItems: %v", err)
	}
	if want := items[0].Value + 60*5*10/5; plan.Value != want {
		t.Fatalf("Value = %d, want %d (base + staff share)", plan.Value, want)
	}
	if plan.Ivalue != 1 {
		t.Fatalf("Ivalue = %d, want 1 (CalcItemValue over empty accumulators)", plan.Ivalue)
	}
}

func ivalueStaffItems() []itemdat.ItemData {
	return []itemdat.ItemData{{
		ID: "STAFF", Name: "Staff", ItemType: "Staff", Durability: 20, Value: 1000,
	}}
}

func ivalueStaffSpells() []txtdata.SpellData {
	return []txtdata.SpellData{
		{Name: "Null"},
		{Name: "Firebolt", StaffLevel: 1, MinInt: 15, StaffMin: 40, StaffMax: 80, StaffCost10: 5},
	}
}
