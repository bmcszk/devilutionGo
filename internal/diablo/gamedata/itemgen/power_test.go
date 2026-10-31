package itemgen_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
)

func TestGetItemPowerFullDraw(t *testing.T) {
	// given
	prefixes, suffixes := lockAffixes()
	req := itemgen.ItemPowerRequest{
		BaseName:  "Sword",
		BaseValue: 1000,
		MinLevel:  1,
		MaxLevel:  6,
		Types:     itemgen.AffixWeapon,
	}
	// draws: prefix coin (4), suffix coin (3), onlygood coin (3), prefix pick
	// (rnd(2)), the prefix SaveItemAffix rolls (5, 3), then the suffix pick:
	// each affix is rolled inside its callback, so the tapes interleave
	// (items.cpp:1214-1226). The pre-fix order came out as [4 3 3 2 3 5 1].
	rnd := &tapeRnd{t: t, values: []int32{0, 1, 0, 1, 0, 0, 3, 0}}

	// when
	out, err := itemgen.GetItemPower(prefixes, suffixes, req, rnd.rnd)
	// then
	if err != nil {
		t.Fatalf("GetItemPower: %v", err)
	}
	assertPowerIdentity(t, out)
	assertPrefixSlot(t, out)
	assertInterleavedArgs(t, rnd.args)
	if out.Value != itemgen.CalcItemValue(out.Values, 1000) {
		t.Fatalf("value = %d, want the CalcItemValue result", out.Value)
	}
}

func assertPowerIdentity(t *testing.T, out itemgen.ItemPowerResult) {
	t.Helper()

	if !out.Magic || out.Name != "Bronze Sword of health" {
		t.Fatalf("name = %q magic = %v, want Bronze Sword of health", out.Name, out.Magic)
	}
	if out.Prefix == nil || out.Suffix == nil {
		t.Fatalf("pair = %+v, want both affixes reported", out)
	}
}

// Bronze carries a zero minVal/maxVal band, so PLVal stores 0 and only the
// multiplier lands in slot 1.
func assertPrefixSlot(t *testing.T, out itemgen.ItemPowerResult) {
	t.Helper()

	if out.Values.Add1 != 0 || out.Values.Mult1 != 2 {
		t.Fatalf("values = %+v, want Add1 0 / Mult1 2", out.Values)
	}
}

// assertInterleavedArgs pins the interleaved roll order: the prefix pick is
// followed by that prefix's own rolls before the suffix pick runs.
func assertInterleavedArgs(t *testing.T, args []int32) {
	t.Helper()

	want := []int32{4, 3, 3, 2, 5, 3, 1}
	if len(args) != len(want) {
		t.Fatalf("draw args = %v, want %v", args, want)
	}
	for i := range want {
		if args[i] != want[i] {
			t.Fatalf("draw args = %v, want %v (picks must not run first)", args, want)
		}
	}
}
