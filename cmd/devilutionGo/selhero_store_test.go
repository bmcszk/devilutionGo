//go:build windowed

package main

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game/ui"
	"github.com/bmcszk/devilutionGo/internal/persist"
)

// TestSelHeroListReflectsFilesOnDisk drives the bean's integration law: per-hero
// files written DIRECTLY to the boot dir (not through the screen) show up on
// the list after a refresh — pfile_ui_set_hero_infos probes the directory's
// own archives (pfile.cpp:695-726 @ca91d778f).
func TestSelHeroListReflectsFilesOnDisk(t *testing.T) {
	boot := newSelHeroTestBoot(t)

	// given a per-hero file written directly at the boot dir (slot 2)
	writeSlotFile(t, boot.storeDir, 2, "Disk Hero")

	// when the screen refreshes
	if err := boot.selhero.refresh(context.Background()); err != nil {
		t.Fatalf("refresh: %v", err)
	}

	// then the list carries the disk hero on slot 2
	names := boot.selhero.heroNames()
	if len(names) != 1 || names[0] != "Disk Hero" {
		t.Fatalf("heroNames = %v, want [Disk Hero]", names)
	}

	if len(boot.selhero.heroes) != 1 || boot.selhero.heroes[0].Slot != 2 {
		t.Fatalf("heroes = %+v, want one row on slot 2", boot.selhero.heroes)
	}
}

// TestSelHeroCreateWritesItsOwnFile pins the create law: each create lands in
// its slot's OWN file — single_0.db and single_1.db after two creates, one
// hero row each, never a shared multi-row file (pfile.cpp:80-92 GetSavePath
// @ca91d778f).
func TestSelHeroCreateWritesItsOwnFile(t *testing.T) {
	boot := newSelHeroTestBoot(t)

	// given two creates through the real flow
	createSelHero(t, boot, "First Hero")
	createSelHero(t, boot, "Second Hero")

	// then each slot owns its file with exactly one hero row
	for slot, name := range map[int]string{0: "First Hero", 1: "Second Hero"} {
		path := filepath.Join(boot.storeDir, "single_"+strconv.Itoa(slot)+".db")

		if _, err := os.Stat(path); err != nil {
			t.Fatalf("%s missing after create: %v", path, err)
		}

		set := persist.NewHeroStoreSet(boot.storeDir, persist.SaveNamespace{})

		rows, err := set.ListHeroes(context.Background())
		if err != nil {
			t.Fatalf("ListHeroes: %v", err)
		}

		if len(rows) != 2 {
			t.Fatalf("ListHeroes = %d rows, want the two created heroes", len(rows))
		}

		for _, row := range rows {
			if row.Slot == slot && row.Name != name {
				t.Fatalf("slot %d holds %q, want %q", slot, row.Name, name)
			}
		}
	}
}

// createSelHero drives one create through the real flow: the keyboard open,
// the name, the warrior pick and the settling frame (the slots tests' shape).
func createSelHero(t *testing.T, boot *selHeroTestBoot, name string) {
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
}

// writeSlotFile writes one hero file directly at dir through
// persist.OpenSQLiteStore — the same layout the migrate CLI produces.
func writeSlotFile(t *testing.T, dir string, slot int, name string) {
	t.Helper()

	ctx := context.Background()

	path := filepath.Join(dir, "single_"+strconv.Itoa(slot)+".db")

	store, err := persist.OpenSQLiteStore(ctx, path)
	if err != nil {
		t.Fatalf("OpenSQLiteStore %s: %v", path, err)
	}

	defer func() {
		if err := store.Close(); err != nil {
			t.Errorf("store Close: %v", err)
		}
	}()

	hero := persist.HeroRecord{
		Slot: slot, Name: name, ClassID: 0, Pack: make([]byte, persist.PlayerPackSize),
	}

	if err := store.SaveHero(ctx, hero); err != nil {
		t.Fatalf("SaveHero slot %d: %v", slot, err)
	}
}
