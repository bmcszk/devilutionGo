package itemgen_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
)

// TestGetItemPowerPrefixAndSuffixForceOneCoin covers the both-coins-false
// branch: FlipCoin(4) and !FlipCoin(3) both fail, then FlipCoin(2) decides
// (items.cpp:1181-1187).
func TestGetItemPowerPrefixAndSuffixForceOneCoin(t *testing.T) {
	tests := []struct {
		name       string
		coin       int32
		wantPrefix bool
	}{
		{name: "coin hits", coin: 0, wantPrefix: true},
		{name: "coin misses", coin: 1, wantPrefix: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			prefixes, suffixes := lockAffixes()
			// prefix coin misses (1), suffix coin hits (0), fallback coin.
			rnd := &tapeRnd{t: t, values: []int32{1, 0, tc.coin, 0, 0, 0, 0}}

			// when
			out := itemgen.GetItemPowerPrefixAndSuffix(
				prefixes, suffixes, 1, 6, itemgen.AffixWeapon, false, rnd.rnd)

			// then
			assertForcedSlot(t, out, tc.wantPrefix)
			if rnd.args[2] != 2 {
				t.Fatalf("fallback coin = %d, want 2 (FlipCoin())", rnd.args[2])
			}
		})
	}
}

//nolint:revive // wantPrefix selects which slot the fallback filled.
func assertForcedSlot(t *testing.T, out itemgen.AffixLoadout, wantPrefix bool) {
	t.Helper()

	if wantPrefix && out.Prefix == nil {
		t.Fatal("Prefix = nil, want the forced prefix")
	}
	if !wantPrefix && out.Suffix == nil {
		t.Fatal("Suffix = nil, want the forced suffix")
	}
}
