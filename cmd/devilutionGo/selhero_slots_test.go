//go:build windowed

package main

import (
	"context"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game/ui"
	"github.com/bmcszk/devilutionGo/internal/persist"
)

// TestSelHeroSecondCreateKeepsTheFirstHero pins the slot law: the second
// create lands in the FIRST FREE slot and never overwrites the first hero's
// row (dcob D3; persist.HeroSlots.FirstFreeSlot, internal/persist/pfile.go).
func TestSelHeroSecondCreateKeepsTheFirstHero(t *testing.T) {
	boot := newSelHeroTestBoot(t)

	create := func(name string) {
		t.Helper()

		boot.handleKey(ui.KeySpace)
		boot.handleKey(ui.KeyEnter)

		if err := boot.selhero.beginCreate(); err != nil {
			t.Fatalf("beginCreate: %v", err)
		}

		if err := boot.selhero.typeName(name); err != nil {
			t.Fatalf("typeName: %v", err)
		}

		boot.selhero.confirmName()
		boot.selhero.classMenu.Focus()
		boot.selhero.classMenu.HandleKey(ui.KeyEnter) // warrior
		boot.frame()

		if rows, err := boot.store.ListHeroes(context.Background()); err != nil {
			t.Fatalf("ListHeroes: %v", err)
		} else if got := len(rows); got < 1 {
			t.Fatalf("create of %q left %d rows, want at least one", name, got)
		}
	}

	create("First Hero")
	create("Second Hero")

	rows, err := boot.store.ListHeroes(context.Background())
	if err != nil {
		t.Fatalf("ListHeroes: %v", err)
	}

	if len(rows) != 2 {
		t.Fatalf("store rows after two creates = %d (%v), want 2 distinct heroes", len(rows), rows)
	}

	if rows[0].Name != "First Hero" || rows[0].Slot != 0 {
		t.Fatalf("row 0 = %+v, want First Hero on slot 0", rows[0])
	}

	if rows[1].Name != "Second Hero" || rows[1].Slot != 1 {
		t.Fatalf("row 1 = %+v, want Second Hero on the first free slot 1", rows[1])
	}
}

// TestSelHeroCreateRefusedWhenAllSlotsTaken pins the cap law: when the store
// reports all 99 slots taken, the create is refused BEFORE any kernel submit —
// no CreatedPack event — and the refusal surfaces through the error row
// (dcob D4; MAX_CHARACTERS = 99, Source/pfile.h:24).
func TestSelHeroCreateRefusedWhenAllSlotsTaken(t *testing.T) {
	boot := newSelHeroTestBoot(t)

	ctx := context.Background()
	for slot := 0; slot < persist.MaxCharacters; slot++ {
		if err := boot.store.SaveHero(ctx, persist.HeroRecord{
			Slot: slot, Name: "Full", ClassID: 0, Pack: make([]byte, persist.PlayerPackSize),
		}); err != nil {
			t.Fatalf("SaveHero slot %d: %v", slot, err)
		}
	}

	if err := boot.selhero.refresh(ctx); err != nil {
		t.Fatalf("refresh: %v", err)
	}

	boot.handleKey(ui.KeySpace)
	boot.handleKey(ui.KeyEnter)

	if err := boot.selhero.beginCreate(); err != nil {
		t.Fatalf("beginCreate: %v", err)
	}

	if err := boot.selhero.typeName("One Too Many"); err != nil {
		t.Fatalf("typeName: %v", err)
	}

	boot.selhero.confirmName()
	boot.selhero.classMenu.Focus()
	boot.selhero.classMenu.HandleKey(ui.KeyEnter) // warrior pick
	boot.frame()

	if events := boot.createdEvents(); len(events) != 0 {
		t.Fatalf("CreatedPack events = %v, want none (refusal before any kernel submit)", events)
	}

	if boot.selhero.lastCreateError == nil {
		t.Fatalf("full-store create left no lastCreateError, want the refusal surfaced")
	}

	rows, err := boot.store.ListHeroes(ctx)
	if err != nil {
		t.Fatalf("ListHeroes: %v", err)
	}

	for _, row := range rows {
		if row.Name == "One Too Many" {
			t.Fatalf("refused create still wrote a row: %+v", row)
		}
	}
}

// TestSelHeroCreateLandsInFreedFirstSlot pins the hole law: after slot 0 is
// deleted through the store, the next create lands back on slot 0 and never
// touches the occupied slot 1 (review R1 F6 — the first-free-slot loop's
// gapped-store case).
func TestSelHeroCreateLandsInFreedFirstSlot(t *testing.T) {
	boot := newSelHeroTestBoot(t)

	ctx := context.Background()
	for slot, name := range map[int]string{0: "First Hero", 1: "Second Hero"} {
		if err := boot.store.SaveHero(ctx, persist.HeroRecord{
			Slot: slot, Name: name, ClassID: 0, Pack: make([]byte, persist.PlayerPackSize),
		}); err != nil {
			t.Fatalf("SaveHero slot %d: %v", slot, err)
		}
	}

	if err := boot.selhero.refresh(ctx); err != nil {
		t.Fatalf("refresh: %v", err)
	}

	// free slot 0 through the store's own delete
	if err := boot.store.DeleteHero(ctx, 0); err != nil {
		t.Fatalf("DeleteHero slot 0: %v", err)
	}

	if err := boot.selhero.refresh(ctx); err != nil {
		t.Fatalf("refresh after delete: %v", err)
	}

	boot.handleKey(ui.KeySpace)
	boot.handleKey(ui.KeyEnter)

	if err := boot.selhero.beginCreate(); err != nil {
		t.Fatalf("beginCreate: %v", err)
	}

	if err := boot.selhero.typeName("Third Hero"); err != nil {
		t.Fatalf("typeName: %v", err)
	}

	boot.selhero.confirmName()
	boot.selhero.classMenu.Focus()
	boot.selhero.classMenu.HandleKey(ui.KeyEnter) // warrior
	boot.frame()

	rows, err := boot.store.ListHeroes(ctx)
	if err != nil {
		t.Fatalf("ListHeroes: %v", err)
	}

	if len(rows) != 2 {
		t.Fatalf("store rows after freed-slot create = %d (%v), want 2", len(rows), rows)
	}

	for _, row := range rows {
		if row.Slot == 0 && row.Name != "Third Hero" {
			t.Fatalf("slot 0 = %+v, want the new hero on the freed first slot", row)
		}

		if row.Slot == 1 && row.Name != "Second Hero" {
			t.Fatalf("slot 1 = %+v, want Second Hero untouched", row)
		}
	}
}
