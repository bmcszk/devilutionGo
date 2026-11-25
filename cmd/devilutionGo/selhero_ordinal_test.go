//go:build windowed

package main

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/txtdata"
	"github.com/bmcszk/devilutionGo/internal/game/ui"
)

// TestSelHeroCreateSubmitsClassOrdinalNotMenuIndex pins the ordinal law: the
// pick submits the picked class's position in the classdat row order, never
// the menu row the player activated. With Rogue (classdat row 1) absent from
// the scripted list, picking menu index 1 must still submit class ordinal 2
// (Sorcerer) — not 1 (dcob D2).
func TestSelHeroCreateSubmitsClassOrdinalNotMenuIndex(t *testing.T) {
	boot := newSelHeroTestBoot(t)
	// the boot's data-built source serves every committed class ordinal, so
	// the assertion targets the submitted class word directly
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

// TestClassDirNameRejectsOutOfRange pins the bounds guard: a class ordinal
// outside the committed classdat row order maps to no fixture directory —
// an empty name, never a panic (review R1 F4).
func TestClassDirNameRejectsOutOfRange(t *testing.T) {
	for _, bad := range []int{-1, len(classDirs), 99} {
		if got := classDirName(bad); got != "" {
			t.Fatalf("classDirName(%d) = %q, want \"\" for an out-of-range ordinal", bad, got)
		}
	}
}
