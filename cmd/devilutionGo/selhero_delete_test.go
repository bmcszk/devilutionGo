//go:build windowed

package main

import (
	"context"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game/ui"
	"github.com/bmcszk/devilutionGo/internal/persist"
)

// TestSelHeroDeleteAsksBeforeDeleting pins the confirm gate: Delete on a
// focused hero opens a Yes/No step and NOTHING is deleted until Yes; No or
// Escape returns to the list with the hero intact (dcob D5; upstream gates
// the removal on UiSelHeroYesNoDialog, selhero.cpp:613-624 @e1ad0c2b6).
func TestSelHeroDeleteAsksBeforeDeleting(t *testing.T) {
	boot := newSelHeroTestBoot(t)

	ctx := context.Background()
	if err := boot.store.SaveHero(ctx, persist.HeroRecord{
		Slot: 0, Name: "Doomed Hero", ClassID: 0, Pack: make([]byte, persist.PlayerPackSize),
	}); err != nil {
		t.Fatalf("SaveHero: %v", err)
	}

	boot.handleKey(ui.KeySpace)
	boot.handleKey(ui.KeyEnter)
	boot.handleKey(ui.KeyArrowDown) // hero row -> Create
	boot.handleKey(ui.KeyArrowDown) // Create -> Delete action
	boot.handleKey(ui.KeyEnter)     // Delete: the confirm step opens

	// then the confirm step is armed and the hero is still alive
	if boot.selhero.deleteConfirmFor == "" {
		t.Fatalf("Delete did not open the confirm step")
	}

	rows, err := boot.store.ListHeroes(ctx)
	if err != nil {
		t.Fatalf("ListHeroes: %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("rows before confirm = %d, want the hero still present", len(rows))
	}

	// when the player answers No (Escape also answers No)
	boot.handleKey(ui.KeyEscape)

	rows, err = boot.store.ListHeroes(ctx)
	if err != nil {
		t.Fatalf("ListHeroes after No: %v", err)
	}

	if len(rows) != 1 || rows[0].Name != "Doomed Hero" {
		t.Fatalf("rows after No = %+v, want the hero intact", rows)
	}

	// and Delete asks again (the list focus stayed on Delete), Yes finally
	// removes the hero
	boot.handleKey(ui.KeyEnter)
	boot.handleKey(ui.KeyEnter) // Yes

	rows, err = boot.store.ListHeroes(ctx)
	if err != nil {
		t.Fatalf("ListHeroes after Yes: %v", err)
	}

	if len(rows) != 0 {
		t.Fatalf("rows after Yes = %+v, want the hero deleted", rows)
	}
}

// TestSelHeroDeleteWithoutFocusedHeroIsNoOp pins D6: Delete with no hero row
// ever focused stays a no-op — the confirm step never opens and nothing is
// removed (upstream disables the Delete button until a hero holds focus,
// selhero.cpp:557 ElementDisabled).
func TestSelHeroDeleteWithoutFocusedHeroIsNoOp(t *testing.T) {
	boot := newSelHeroTestBoot(t)

	// no heroes in the store: no row ever focused
	boot.handleKey(ui.KeySpace)
	boot.handleKey(ui.KeyEnter)

	if err := boot.selhero.refresh(context.Background()); err != nil {
		t.Fatalf("refresh: %v", err)
	}

	// focus the Quit action directly (no hero row exists to focus)
	boot.handleKey(ui.KeyArrowUp) // wrap to Quit
	boot.handleKey(ui.KeyArrowUp) // Quit -> Delete
	boot.handleKey(ui.KeyEnter)   // Delete with no focused hero

	if boot.selhero.deleteConfirmFor != "" {
		t.Fatalf("Delete with no focused hero opened the confirm step for %q", boot.selhero.deleteConfirmFor)
	}
}
