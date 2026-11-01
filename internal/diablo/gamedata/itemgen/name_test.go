package itemgen_test

import (
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
)

func TestDrawItemPowerPrefixAndSuffixDiscards(t *testing.T) {
	tests := []struct {
		name     string
		prefix   itemgen.Affix
		wantBurn int
	}{
		{name: "default power", prefix: itemgen.Affix{Name: "Bronze", Power: "TOHIT", Chance: 1}, wantBurn: 1},
		{name: "tohit damp", prefix: itemgen.Affix{Name: "Leech", Power: "TOHIT_DAMP", Chance: 1}, wantBurn: 3},
		{name: "doppelganger", prefix: itemgen.Affix{Name: "Doppelganger", Power: "DOPPELGANGER", Chance: 1}, wantBurn: 3},
		{name: "tohit damp curse", prefix: itemgen.Affix{Name: "Drain", Power: "TOHIT_DAMP_CURSE", Chance: 1}, wantBurn: 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// when
			prefix, burned := runPrefixDraw(t, tc.prefix)

			// then
			if prefix.Name != tc.prefix.Name || burned != tc.wantBurn {
				t.Fatalf("prefix/burned = %s/%d, want %s/%d (items.cpp:2341-2351)",
					prefix.Name, burned, tc.prefix.Name, tc.wantBurn)
			}
		})
	}
}

func TestMagicItemNameFallback(t *testing.T) {
	// given
	prefix := &itemgen.Affix{Name: "Bronze"}
	suffix := &itemgen.Affix{Name: "health"}
	fits := func(name string) bool { return len(name) <= 20 }

	// when
	short := itemgen.MagicItemName("Great Sword of the Whale", "G.Sword", prefix, suffix, fits)
	long := itemgen.MagicItemName("Sword", "Sword", prefix, suffix, fits)
	unchecked := itemgen.MagicItemName("Sword", "Sword", prefix, suffix, nil)

	// then
	if !strings.Contains(short, "G.Sword") {
		t.Fatalf("name = %q, want the short-name retry", short)
	}
	if long != "Bronze Sword of health" {
		t.Fatalf("name = %q, want the long name kept", long)
	}
	if unchecked != "Bronze Sword of health" {
		t.Fatalf("name = %q, want the long name with no fit check", unchecked)
	}
}

func runPrefixDraw(t *testing.T, prefix itemgen.Affix) (itemgen.Affix, int) {
	t.Helper()
	prefix.Types = itemgen.AffixWeapon
	prefix.MinLevel = 1
	rnd := &tapeRnd{t: t, values: []int32{0, 0, 0, 0, 0, 0}}
	burned := 0
	out := itemgen.DrawItemPowerPrefixAndSuffix(
		[]itemgen.Affix{prefix}, nil, 1, 6, itemgen.AffixWeapon, false, rnd.rnd,
		func(n int) { burned += n },
	)
	if out.Prefix == nil {
		t.Fatal("DrawItemPowerPrefixAndSuffix: no prefix drawn")
	}

	return *out.Prefix, burned
}
