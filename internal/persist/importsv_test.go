package persist_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

// TestImportSVFillsClassID checks the imported hero row carries the pack's
// class, the second list column upstream fills alongside the name from the
// decoded pack (pfile.cpp:215).
func TestImportSVFillsClassID(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()
	store := openTempSQLiteStore(ctx, t)

	// given a class-2 save archive
	path := filepath.Join(dir, "single_0.sv")
	writeClassSave(t, path, 2)

	// when the archive is imported
	if err := persist.ImportSV(ctx, store, path, persist.GameMode{}, 0); err != nil {
		t.Fatalf("ImportSV: %v", err)
	}

	// then the row class matches the pack class
	hero, err := store.LoadHero(ctx, 0)
	if err != nil {
		t.Fatalf("LoadHero: %v", err)
	}

	if hero.ClassID != 2 {
		t.Fatalf("row ClassID = %d, want 2", hero.ClassID)
	}
}

// writeClassSave writes a valid save archive whose pack carries the class.
func writeClassSave(t *testing.T, path string, class uint8) {
	t.Helper()

	game, err := persist.GameHeader{CurrLevel: 1}.Encode()
	if err != nil {
		t.Fatalf("encode game header: %v", err)
	}

	pack := goldenPlayerPack()
	pack.Class = class

	archive := persist.LegacySVStore{Path: path, Mode: persist.GameMode{}}
	hero := persist.HeroRecord{Slot: 0, Name: "HeroA", Pack: pack.Encode()}

	if err := archive.Write(hero, game); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
