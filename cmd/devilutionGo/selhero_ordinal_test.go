//go:build windowed

package main

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/txtdata"
	"github.com/bmcszk/devilutionGo/internal/game/ui"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// scriptedTestPacks is a PackSource that accepts every class ordinal (a
// minimal warrior pack), so the ordinal assertion is independent of which
// classes the production pack source serves.
type scriptedTestPacks struct{}

func (scriptedTestPacks) StartingPack(class int) (world.StartingPack, bool) {
	return world.StartingPack{Class: class}, true
}

// TestSelHeroCreateSubmitsClassOrdinalNotMenuIndex pins the ordinal law: the
// pick submits the picked class's position in the classdat row order, never
// the menu row the player activated. With Rogue (classdat row 1) absent from
// the scripted list, picking menu index 1 must still submit class ordinal 2
// (Sorcerer) — not 1 (dcob D2).
func TestSelHeroCreateSubmitsClassOrdinalNotMenuIndex(t *testing.T) {
	boot := newSelHeroTestBoot(t)
	// a pack source accepting every ordinal: the assertion targets the
	// submitted class word, not which classes the demo source serves
	world.RegisterCreatePack(boot.lane.game.Kernel, boot.lane.state, scriptedTestPacks{})
	boot.handleKey(ui.KeySpace)
	boot.handleKey(ui.KeyEnter)

	if err := boot.selhero.beginCreate(); err != nil {
		t.Fatalf("beginCreate: %v", err)
	}

	if err := boot.selhero.typeName("Sorcerer Hero"); err != nil {
		t.Fatalf("typeName: %v", err)
	}

	boot.selhero.confirmName()

	// script a class list with the middle row removed: Rogue absent — the
	// classdat row order (Warrior=0, Sorcerer=2) diverges from the menu order
	boot.selhero.classes = []txtdata.ClassData{
		{ClassName: "Warrior", FolderName: "warrior"},
		{ClassName: "Sorcerer", FolderName: "sorcerer"},
	}

	boot.selhero.classMenu = ui.NewMenuList(classNames(boot.selhero.classes))
	boot.selhero.classMenu.OnSelect(boot.selhero.pickClass)
	boot.selhero.classMenu.Focus()
	boot.selhero.classMenu.HandleKey(ui.KeyArrowDown) // menu index 1 = Sorcerer row

	boot.selhero.classMenu.HandleKey(ui.KeyEnter) // pick it

	boot.frame()

	events := boot.createdEvents()
	if len(events) != 1 {
		t.Fatalf("CreatedPack events = %d (%v), want exactly one", len(events), events)
	}

	if events[0].A != 2 {
		t.Fatalf("CreatedPack class = %d, want 2 (the true classdat ordinal of Sorcerer, not the menu index 1)", events[0].A)
	}
}
