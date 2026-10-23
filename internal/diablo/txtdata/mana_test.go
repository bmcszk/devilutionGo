package txtdata_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/txtdata"
	enginetxtdata "github.com/bmcszk/devilutionGo/internal/engine/txtdata"
)

func TestNullSpellPrepended(t *testing.T) {
	// given
	// when
	tbl, err := enginetxtdata.Parse([]byte(spellRow))
	// then
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	spells, err := txtdata.LoadSpellDataWithNull(tbl)
	if err != nil {
		t.Fatalf("LoadSpellDataWithNull: %v", err)
	}
	if len(spells) != 3 { // Null + Firebolt + Healing
		t.Fatalf("spells = %d, want 3", len(spells))
	}
	n := spells[0]
	if n.Name != "Null" || n.ManaCost != 0 || len(n.Missiles) != 2 || n.Missiles[0] != "Null" {
		t.Fatalf("null spell = %+v", n)
	}
}

func TestSpellIDByName(t *testing.T) {
	// given
	// when
	tbl, err := enginetxtdata.Parse([]byte(spellRow))
	// then
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	i, err := txtdata.SpellIDByName(tbl, "Healing")
	if err != nil || i != 1 {
		t.Fatalf("SpellIDByName(Healing) = %d, %v", i, err)
	}
	if _, err := txtdata.SpellIDByName(tbl, "Nope"); err == nil {
		t.Fatal("unknown id = nil error, want error")
	}
}

// TestGetManaAmountFireboltHalvesAdjAfterMultiply: Firebolt halves the
// adjustment after the multiply (spells.cpp:116-118), so an odd product
// loses its low bit.
func TestGetManaAmountFireboltHalvesAdjAfterMultiply(t *testing.T) {
	// given
	// adj = sl(3) * ManaAdj(3) = 9, halved to 4; ma = 6 - 4 = 2 → 2<<6 = 128.
	// when
	got := txtdata.GetManaAmount(
		txtdata.ManaInput{SpellLevel: 4, ClassManaCost: 64},
		txtdata.SpellCosts{ManaCost: 6, ManaAdj: 3, IsFirebolt: true},
	)
	// then
	if got != 128 {
		t.Fatalf("firebolt mana = %d, want 128", got)
	}
}

// TestGetManaAmountClassMultiplierAppliesAfterShift: the class multiplier
// is a 1/64 fraction applied to the shifted cost (spells.cpp:132-134).
func TestGetManaAmountClassMultiplierAppliesAfterShift(t *testing.T) {
	// given
	// ma = 16<<6 = 1024; 1024 * 48 >> 6 = 768 (0.75 class mana cost).
	// when
	got := txtdata.GetManaAmount(
		txtdata.ManaInput{SpellLevel: 1, ClassManaCost: 48},
		txtdata.SpellCosts{ManaCost: 16, ManaAdj: 1},
	)
	// then
	if got != 768 {
		t.Fatalf("class-scaled mana = %d, want 768", got)
	}
}

// TestGetManaAmountHealingUsesHealingRowCost: the Healing branch reads
// SpellData[Healing].sManaCost, not the cast spell's own cost
// (spells.cpp:124).
func TestGetManaAmountHealingUsesHealingRowCost(t *testing.T) {
	// given
	// adj = sl(2) * 3 = 6; ma = 5 (Healing row) + 2*10 (character level) - 6 = 19 → 19<<6 = 1216.
	// when
	got := txtdata.GetManaAmount(
		txtdata.ManaInput{SpellLevel: 3, CharacterLevel: 10, ClassManaCost: 64},
		txtdata.SpellCosts{ManaCost: 99, ManaAdj: 3, HealingManaCost: 5, IsHealing: true},
	)
	// then
	if got != 1216 {
		t.Fatalf("healing mana = %d, want 1216", got)
	}
}

// TestGetManaAmountHealingUsesCharacterLevel: the Healing branch scales
// with the caster's character level, not with the spell level
// (spells.cpp:124).
func TestGetManaAmountHealingUsesCharacterLevel(t *testing.T) {
	// given
	// ma = 5 + 2*25 (character level) - 0 = 55 → 55<<6 = 3520.
	// when
	got := txtdata.GetManaAmount(
		txtdata.ManaInput{SpellLevel: 1, CharacterLevel: 25, ClassManaCost: 64},
		txtdata.SpellCosts{ManaCost: 5, ManaAdj: 3, HealingManaCost: 5, IsHealing: true},
	)
	// then
	if got != 3520 {
		t.Fatalf("healing mana = %d, want 3520", got)
	}
}

// TestGetManaAmountResurrectDividesCostByEight: Resurrect adjusts with
// its own cost divided by eight (spells.cpp:119-121).
func TestGetManaAmountResurrectDividesCostByEight(t *testing.T) {
	// given
	// adj = sl(8) * (20/8) = 16; ma = 20 - 16 = 4 → 4<<6 = 256.
	// when
	got := txtdata.GetManaAmount(
		txtdata.ManaInput{SpellLevel: 9, ClassManaCost: 64},
		txtdata.SpellCosts{ManaCost: 20, ManaAdj: 0, IsResurrect: true},
	)
	// then
	if got != 256 {
		t.Fatalf("resurrect mana = %d, want 256", got)
	}
}

// TestGetManaAmountMinManaFloor: a spell that the adjustment drains to
// zero still costs its sMinMana (spells.cpp:137-139).
func TestGetManaAmountMinManaFloor(t *testing.T) {
	// given
	// adj = sl(10) * 1 = 10; ma = 10 - 10 = 0, floored to 6<<6 = 384.
	// when
	got := txtdata.GetManaAmount(
		txtdata.ManaInput{SpellLevel: 11, ClassManaCost: 64},
		txtdata.SpellCosts{ManaCost: 10, ManaAdj: 1, MinMana: 6},
	)
	// then
	if got != 384 {
		t.Fatalf("floored mana = %d, want 384", got)
	}
}

// TestGetManaAmountCost255UsesMaxManaBase: the cost-255 branch prices
// the cast off the base mana (spells.cpp:125-126).
func TestGetManaAmountCost255UsesMaxManaBase(t *testing.T) {
	// given
	// adj = 1; ma = 640>>6 - 1 = 9 → 9<<6 = 576.
	// when
	got := txtdata.GetManaAmount(
		txtdata.ManaInput{SpellLevel: 2, ClassManaCost: 64, MaxManaBase: 640},
		txtdata.SpellCosts{ManaCost: 255, ManaAdj: 1},
	)
	// then
	if got != 576 {
		t.Fatalf("manaCost-255 mana = %d, want 576", got)
	}
}

// TestGetManaAmountClampsNegative: an adjustment above the cost clamps
// the price at zero (spells.cpp:129-131).
func TestGetManaAmountClampsNegative(t *testing.T) {
	// given
	// adj = sl(3) * 3 = 9; ma = 1 - 9 = -8 clamps to 0.
	// when
	got := txtdata.GetManaAmount(
		txtdata.ManaInput{SpellLevel: 4, ClassManaCost: 64},
		txtdata.SpellCosts{ManaCost: 1, ManaAdj: 3},
	)
	if got != 0 {
		t.Fatalf("clamped mana = %d, want 0", got)
	}
}

// TestGetManaAmountSpellLevelOnePaysNoAdjustment: the adjustment only
// starts above spell level one (spells.cpp:110-113).
func TestGetManaAmountSpellLevelOnePaysNoAdjustment(t *testing.T) {
	// given
	// sl = 0, adj = 0; ma = 6 → 6<<6 = 384.
	// when
	got := txtdata.GetManaAmount(
		txtdata.ManaInput{SpellLevel: 1, ClassManaCost: 64},
		txtdata.SpellCosts{ManaCost: 6, ManaAdj: 3},
	)
	if got != 384 {
		t.Fatalf("mana = %d, want 384", got)
	}
}
