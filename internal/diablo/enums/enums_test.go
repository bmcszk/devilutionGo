package enums_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/enums"
)

func TestCursorIDOrderMatchesUpstream(t *testing.T) {
	// given
	// Positional values are the 1:1 contract.
	want := []enums.CursorID{
		enums.CursorNone, enums.CursorHand, enums.CursorIdentify,
		enums.CursorRepair, enums.CursorRecharge, enums.CursorDisarm,
		enums.CursorOil, enums.CursorTelekinesis, enums.CursorResurrect,
		enums.CursorTeleport, enums.CursorHealOther, enums.CursorHourglass,
		enums.CursorFirstItem,
	}
	for i, c := range want {
		if uint8(c) != uint8(i) {
			t.Fatalf("cursor %d = %d, want %d", i, c, i)
		}
	}
}

func TestSelectionRegionBits(t *testing.T) {
	// given
	if enums.SelectionNone != 0 {
		t.Fatal("None must be 0")
	}
	if enums.SelectionBottom != 1 || enums.SelectionMiddle != 2 || enums.SelectionTop != 4 {
		t.Fatalf("region bits = %d/%d/%d", enums.SelectionBottom, enums.SelectionMiddle, enums.SelectionTop)
	}
}

func TestGoldLimits(t *testing.T) {
	// given
	if enums.GoldSmallLimit != 1000 || enums.GoldMediumLimit != 2500 || enums.GoldMaxLimit != 5000 {
		t.Fatalf("gold limits = %d/%d/%d", enums.GoldSmallLimit, enums.GoldMediumLimit, enums.GoldMaxLimit)
	}
	if enums.GoldMaxLimitAuric != 2*enums.GoldMaxLimit {
		t.Fatalf("auric gold cap = %d, want %d", enums.GoldMaxLimitAuric, 2*enums.GoldMaxLimit)
	}
	if enums.DurIndestructible != 255 {
		t.Fatalf("DurIndestructible = %d, want 255", enums.DurIndestructible)
	}
}

func TestSpawnGateValuesAreDistinct(t *testing.T) {
	// given
	if enums.SpawnGateNone != 0 {
		t.Fatal("SpawnGateNone must be 0")
	}
	gates := []enums.SpawnGate{
		enums.SpawnGateNone,
		enums.SpawnGateQuests,
		enums.SpawnGateMonsters,
	}
	seen := make(map[enums.SpawnGate]bool, len(gates))
	for _, g := range gates {
		if seen[g] {
			t.Fatalf("duplicate spawn gate value %d", g)
		}
		seen[g] = true
	}
}
