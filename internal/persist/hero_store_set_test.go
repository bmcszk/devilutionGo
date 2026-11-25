package persist_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

// heroSetNS is the single-player namespace the windowed boot binds: the db
// extension the live store uses.
var heroSetNS = persist.SaveNamespace{}

// TestHeroStoreSetListHeroesProbesFiles pins the per-hero file law: the list
// probes every slot file on disk (pfile_ui_set_hero_infos walks 0..98,
// pfile.cpp:695-726 @ca91d778f) — a missing file is an empty slot, a present
// file contributes its row, and nothing but the touched slots' files exist.
func TestHeroStoreSetListHeroesProbesFiles(t *testing.T) {
	dir := t.TempDir()
	set := persist.NewHeroStoreSet(dir, heroSetNS)

	ctx := context.Background()

	// given two heroes saved through the set, files single_0.db + single_3.db
	if err := set.SaveHero(ctx, heroRow(0, "Zero")); err != nil {
		t.Fatalf("SaveHero slot 0: %v", err)
	}

	if err := set.SaveHero(ctx, heroRow(3, "Three")); err != nil {
		t.Fatalf("SaveHero slot 3: %v", err)
	}

	// when the list probes the directory
	rows, err := set.ListHeroes(ctx)
	if err != nil {
		t.Fatalf("ListHeroes: %v", err)
	}

	// then the rows and the files reflect exactly the touched slots
	assertRowSlots(t, rows)
	assertFilesOnDisk(t, dir)
}

// TestHeroStoreSetDeleteRemovesTheFile pins delete-as-file-removal
// (pfile_delete_save removes GetSavePath's file, pfile.cpp:775-783
// @ca91d778f): the slot's file is gone and a second delete is the
// ErrHeroNotFound sentinel, not a crash.
func TestHeroStoreSetDeleteRemovesTheFile(t *testing.T) {
	dir := t.TempDir()
	set := persist.NewHeroStoreSet(dir, heroSetNS)

	ctx := context.Background()

	// given a hero saved on slot 2
	if err := set.SaveHero(ctx, heroRow(2, "Gone")); err != nil {
		t.Fatalf("SaveHero slot 2: %v", err)
	}

	// when the hero is deleted
	if err := set.DeleteHero(ctx, 2); err != nil {
		t.Fatalf("DeleteHero slot 2: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "single_2.db")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("single_2.db stat after delete = %v, want not-exist", err)
	}

	// then the missing-file delete is the row-store's sentinel
	err := set.DeleteHero(ctx, 2)
	if !errors.Is(err, persist.ErrHeroNotFound) {
		t.Fatalf("DeleteHero on missing file = %v, want ErrHeroNotFound", err)
	}
}

// TestHeroStoreSetPathIsGetSavePath pins the path law: the set names files
// through SaveNamespace.SavePath — dir/single_N.db — never a hand-assembled
// string (pfile.cpp:80-92 GetSavePath @ca91d778f), and out-of-range slots
// carry the SavePath range error.
func TestHeroStoreSetPathIsGetSavePath(t *testing.T) {
	dir := t.TempDir()
	set := persist.NewHeroStoreSet(dir, heroSetNS)

	// when the slot's path is asked
	path, err := set.Path(7)
	if err != nil {
		t.Fatalf("Path(7): %v", err)
	}

	want := filepath.Join(dir, "single_7.db")
	if path != want {
		t.Fatalf("Path(7) = %q, want %q", path, want)
	}

	// then the out-of-range slot carries the SavePath range error
	if _, err := set.Path(persist.MaxCharacters); !errors.Is(err, persist.ErrSlotOutOfRange) {
		t.Fatalf("Path(%d) error = %v, want ErrSlotOutOfRange", persist.MaxCharacters, err)
	}
}

// heroRow builds the named hero record for slot.
func heroRow(slot int, name string) persist.HeroRecord {
	return persist.HeroRecord{
		Slot: slot, Name: name, ClassID: 0, Pack: make([]byte, persist.PlayerPackSize),
	}
}

// assertRowSlots checks the probe result: two rows, slots [0 3].
func assertRowSlots(t *testing.T, rows []persist.HeroRecord) {
	t.Helper()

	if len(rows) != 2 {
		t.Fatalf("ListHeroes = %d rows (%v), want 2", len(rows), rows)
	}

	slots := make([]int, 0, len(rows))
	for _, row := range rows {
		slots = append(slots, row.Slot)
	}

	if !slices.Equal(slots, []int{0, 3}) {
		t.Fatalf("row slots = %v, want [0 3]", slots)
	}
}

// assertFilesOnDisk checks each hero owns exactly one file.
func assertFilesOnDisk(t *testing.T, dir string) {
	t.Helper()

	for _, name := range []string{"single_0.db", "single_3.db"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("%s missing: %v", name, err)
		}
	}

	assertAbsent(t, dir, "single_1.db")
}

// assertAbsent fails when the named file exists.
func assertAbsent(t *testing.T, dir, name string) {
	t.Helper()

	if _, err := os.Stat(filepath.Join(dir, name)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("%s stat = %v, want not-exist", name, err)
	}
}
