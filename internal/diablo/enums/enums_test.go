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
	// when
	got := make([]uint8, 0, len(want))
	for _, c := range want {
		got = append(got, uint8(c))
	}
	// then
	for i, value := range got {
		if value != uint8(i) {
			t.Fatalf("cursor %d = %d, want %d", i, value, i)
		}
	}
}

func TestSelectionRegionBits(t *testing.T) {
	// when
	none, bottom, middle, top := enums.SelectionNone, enums.SelectionBottom, enums.SelectionMiddle, enums.SelectionTop
	// then
	if none != 0 {
		t.Fatal("None must be 0")
	}

	if bottom != 1 || middle != 2 || top != 4 {
		t.Fatalf("region bits = %d/%d/%d", bottom, middle, top)
	}
}

func TestGoldLimits(t *testing.T) {
	// when
	small, medium, maxGold := enums.GoldSmallLimit, enums.GoldMediumLimit, enums.GoldMaxLimit
	auric, dur := enums.GoldMaxLimitAuric, enums.DurIndestructible
	// then
	if small != 1000 || medium != 2500 || maxGold != 5000 {
		t.Fatalf("gold limits = %d/%d/%d", small, medium, maxGold)
	}

	if auric != 2*maxGold {
		t.Fatalf("auric gold cap = %d, want %d", auric, 2*maxGold)
	}

	if dur != 255 {
		t.Fatalf("DurIndestructible = %d, want 255", dur)
	}
}

func TestSpawnGateValuesAreDistinct(t *testing.T) {
	// given
	gates := []enums.SpawnGate{
		enums.SpawnGateNone,
		enums.SpawnGateQuests,
		enums.SpawnGateMonsters,
	}
	// when
	seen := make(map[enums.SpawnGate]bool, len(gates))
	duplicates := make([]enums.SpawnGate, 0)
	for _, g := range gates {
		if seen[g] {
			duplicates = append(duplicates, g)
		}

		seen[g] = true
	}
	// then
	for _, g := range duplicates {
		t.Fatalf("duplicate spawn gate value %d", g)
	}
}
