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
	// draws: prefix coin, suffix coin, onlygood coin, prefix pick, suffix pick,
	// followed by the prefix RndPL and PLVal inputs.
	rnd := &tapeRnd{t: t, values: []int32{0, 1, 0, 1, 0, 0, 3, 0}}

	// when
	out, err := itemgen.GetItemPower(prefixes, suffixes, req, rnd.rnd)
	// then
	if err != nil {
		t.Fatalf("GetItemPower: %v", err)
	}
	assertPowerIdentity(t, out)
	assertPrefixSlot(t, out)
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
