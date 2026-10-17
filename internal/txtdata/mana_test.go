package txtdata_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/txtdata"
)

func TestNullSpellPrepended(t *testing.T) {
	tbl, err := txtdata.Parse([]byte(spellRow))
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
	tbl, err := txtdata.Parse([]byte(spellRow))
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

func TestGetManaAmountBaseSpell(t *testing.T) {
	// plain spell: ma = manaCost - sl*adj, clamped, <<6, class mult
	got := txtdata.GetManaAmount(
		txtdata.ManaInput{SpellLevel: 3, ClassManaCost: 64},
		txtdata.SpellCosts{ManaCost: 10, ManaAdj: 2},
	)
	// sl=2, adj=4, ma=10-4=6, <<6=384, *64>>6=384
	if got != 384 {
		t.Fatalf("GetManaAmount = %d, want 384", got)
	}
}

func TestGetManaAmountClampsNegative(t *testing.T) {
	got := txtdata.GetManaAmount(
		txtdata.ManaInput{SpellLevel: 10, ClassManaCost: 64},
		txtdata.SpellCosts{ManaCost: 5, ManaAdj: 2},
	)
	if got != 0 {
		t.Fatalf("clamped mana = %d, want 0", got)
	}
}

func TestGetManaAmountHealing(t *testing.T) {
	// Healing: ma = manaCost + 2*level - adj
	got := txtdata.GetManaAmount(
		txtdata.ManaInput{SpellLevel: 4, ClassManaCost: 64},
		txtdata.SpellCosts{ManaCost: 5, ManaAdj: 1, IsHealing: true},
	)
	// ma = 5 + 8 - 3 = 10, <<6 = 640
	if got != 640 {
		t.Fatalf("healing mana = %d, want 640", got)
	}
}

func TestGetManaAmountFireboltHalvesAdj(t *testing.T) {
	got := txtdata.GetManaAmount(
		txtdata.ManaInput{SpellLevel: 3, ClassManaCost: 64},
		txtdata.SpellCosts{ManaCost: 10, ManaAdj: 2, IsFirebolt: true},
	)
	// sl=2, adj=2*1=2, ma=10-2=8, <<6=512
	if got != 512 {
		t.Fatalf("firebolt mana = %d, want 512", got)
	}
}

func TestGetManaAmountResurrectDividesCost(t *testing.T) {
	got := txtdata.GetManaAmount(
		txtdata.ManaInput{SpellLevel: 3, ClassManaCost: 64},
		txtdata.SpellCosts{ManaCost: 40, ManaAdj: 2, IsResurrect: true},
	)
	// adj = 2*(40/8)=10, ma=40-10=30, <<6=1920
	if got != 1920 {
		t.Fatalf("resurrect mana = %d, want 1920", got)
	}
}

func TestGetManaAmountCost255UsesMaxMana(t *testing.T) {
	got := txtdata.GetManaAmount(
		txtdata.ManaInput{SpellLevel: 2, ClassManaCost: 64, MaxManaBase: 640},
		txtdata.SpellCosts{ManaCost: 255, ManaAdj: 1},
	)
	// adj=1, ma = 640>>6 - 1 = 9, <<6 = 576
	if got != 576 {
		t.Fatalf("manaCost-255 mana = %d, want 576", got)
	}
}
