// TODO remove nolint:testpackage // internal-state test: pokes unexported identifiers
package enums

import (
	"testing"
)

func TestCursorIDOrderMatchesUpstream(t *testing.T) {
	// Positional values are the 1:1 contract.
	// given
	want := []cursorID{
		CursorNone, CursorHand, CursorIdentify,
		CursorRepair, CursorRecharge, CursorDisarm,
		CursorOil, CursorTelekinesis, CursorResurrect,
		CursorTeleport, CursorHealOther, CursorHourglass,
		CursorFirstItem,
	}
	// when
	for i, c := range want {
		// then
		if uint8(c) != uint8(i) {
			t.Fatalf("cursor %d = %d, want %d", i, c, i)
		}
	}
}

func TestSelectionRegionBits(t *testing.T) {
	if SelectionNone != 0 {
		t.Fatal("None must be 0")
	}
	if SelectionBottom != 1 || SelectionMiddle != 2 || SelectionTop != 4 {
		t.Fatalf("region bits = %d/%d/%d", SelectionBottom, SelectionMiddle, SelectionTop)
	}
}

func TestGoldLimits(t *testing.T) {
	if GoldSmallLimit != 1000 || GoldMediumLimit != 2500 || GoldMaxLimit != 5000 {
		t.Fatalf("gold limits = %d/%d/%d", GoldSmallLimit, GoldMediumLimit, GoldMaxLimit)
	}
	if GoldMaxLimitAuric != 2*GoldMaxLimit {
		t.Fatalf("auric gold cap = %d, want %d", GoldMaxLimitAuric, 2*GoldMaxLimit)
	}
	if DurIndestructible != 255 {
		t.Fatalf("DurIndestructible = %d, want 255", DurIndestructible)
	}
}

func TestSpawnGateValuesAreDistinct(t *testing.T) {
	// given
	gates := []spawnGate{
		SpawnGateNone,
		SpawnGateQuests,
		SpawnGateMonsters,
	}
	seen := make(map[spawnGate]bool, len(gates))
	// when
	for _, g := range gates {
		// then
		if seen[g] {
			t.Fatalf("duplicate spawn gate value %d", g)
		}
		seen[g] = true
	}
}
