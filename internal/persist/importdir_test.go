package persist_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

// saveNameCase is one file-name classification.
type saveNameCase struct {
	name   string
	file   string
	prefix string
	slot   int
	ok     bool
}

// saveNameResult is what ParseHeroSaveName returned.
type saveNameResult struct {
	prefix string
	slot   int
	ok     bool
}

// TestParseHeroSaveName pins the file names GetSavePath produces
// (pfile.cpp:81-88) and the names it does not.
func TestParseHeroSaveName(t *testing.T) {
	t.Parallel()

	for _, c := range heroSaveNameCases() {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			// when the file name is classified
			prefix, slot, ok := persist.ParseHeroSaveName(c.file)

			// then it is accepted or rejected as upstream would
			requireSaveName(t, c, saveNameResult{prefix: prefix, slot: slot, ok: ok})
		})
	}
}

// TestImportDirectoryImportsEveryHero checks a whole pref path is picked up:
// four prefixes, their modes preserved, non-save files skipped.
func TestImportDirectoryImportsEveryHero(t *testing.T) {
	dir := t.TempDir()
	store := openTempSQLiteStore(t)

	for slot, file := range []string{"single_0.sv", "multi_1.sv", "spawn_2.sv", "share_3.sv"} {
		writeHeroArchive(t, filepath.Join(dir, file), heroSaveName(slot))
	}

	// given stray files an import must ignore
	writeTextFile(t, filepath.Join(dir, "notes.txt"))
	writeTextFile(t, filepath.Join(dir, "single_100.sv"))

	// when the directory is imported
	report, err := persist.ImportDirectory(context.Background(), store, dir)
	if err != nil {
		t.Fatalf("ImportDirectory: %v", err)
	}

	// then the four hero saves are the report and the strays are not
	if len(report.Imported) != 4 || report.Errors() {
		t.Fatalf("report = %+v, want 4 imported and no failures", report)
	}

	// and each landed in its own slot, with its name and game stream
	requireImportedSaves(t, store)
}

// TestImportDirectoryReportsBadSave checks a corrupt archive is reported per
// file instead of aborting the batch.
func TestImportDirectoryReportsBadSave(t *testing.T) {
	dir := t.TempDir()
	store := openTempSQLiteStore(t)

	writeHeroArchive(t, filepath.Join(dir, "single_0.sv"), "Good")

	// given one archive that is not a save
	if err := os.WriteFile(filepath.Join(dir, "single_1.sv"), []byte("not a save"), 0o600); err != nil {
		t.Fatalf("write bad save: %v", err)
	}

	// when the directory is imported
	report, err := persist.ImportDirectory(context.Background(), store, dir)
	if err != nil {
		t.Fatalf("ImportDirectory: %v", err)
	}

	// then the good save landed and the bad one is reported
	if len(report.Imported) != 1 || len(report.Failed) != 1 || !report.Errors() {
		t.Fatalf("report = %+v, want 1 imported and 1 failure", report)
	}

	if report.Summary() == "" {
		t.Fatal("Summary() is empty for a failed import")
	}
}

// TestImportDirectoryImportsStash checks the shared stash lands in the global
// row and comes back out unchanged.
func TestImportDirectoryImportsStash(t *testing.T) {
	dir := t.TempDir()
	store := openTempSQLiteStore(t)

	writeStashArchive(t, filepath.Join(dir, "stash.sv"), persist.GameMode{})

	// when the directory is imported
	report, err := persist.ImportDirectory(context.Background(), store, dir)
	if err != nil {
		t.Fatalf("ImportDirectory: %v", err)
	}

	// then the stash file is reported as imported
	if len(report.Imported) != 1 || report.Errors() {
		t.Fatalf("report = %+v, want one imported stash", report)
	}

	// and the global stash row round-trips back into an archive
	out := filepath.Join(t.TempDir(), "stash.sv")
	err = persist.ExportStash(context.Background(), store, out, persist.GameMode{}, persist.GlobalStashSlot)
	if err != nil {
		t.Fatalf("ExportStash: %v", err)
	}

	if _, err := os.Stat(out); err != nil {
		t.Fatalf("exported stash: %v", err)
	}
}

// requireImportedSaves checks the four slots landed with their mode's kind.
func requireImportedSaves(t *testing.T, store *persist.SQLiteStore) {
	t.Helper()

	kinds := map[int]persist.GameKind{
		0: persist.KindSingle,
		1: persist.KindMulti,
		2: persist.KindSpawn,
		3: persist.KindSpawnMulti,
	}

	for slot, kind := range kinds {
		requireImportedSave(t, store, slot, kind)
	}
}

// requireImportedSave checks one slot's hero name and game stream.
func requireImportedSave(t *testing.T, store *persist.SQLiteStore, slot int, kind persist.GameKind) {
	t.Helper()

	hero, err := store.LoadHero(context.Background(), slot)
	if err != nil {
		t.Fatalf("LoadHero %d: %v", slot, err)
	}

	if hero.Name != heroSaveName(slot) {
		t.Fatalf("slot %d name = %q, want %q", slot, hero.Name, heroSaveName(slot))
	}

	if _, err := store.LoadGame(context.Background(), slot, string(kind)); err != nil {
		t.Fatalf("LoadGame %d %s: %v", slot, kind, err)
	}
}

// heroSaveNameCases covers the four prefixes and the near-misses.
func heroSaveNameCases() []saveNameCase {
	return []saveNameCase{
		{name: "single", file: "single_0.sv", prefix: "single_", slot: 0, ok: true},
		{name: "multi high slot", file: "multi_98.sv", prefix: "multi_", slot: 98, ok: true},
		{name: "spawn", file: "spawn_7.sv", prefix: "spawn_", slot: 7, ok: true},
		{name: "share", file: "share_2.sv", prefix: "share_", slot: 2, ok: true},
		{name: "stash is not a hero", file: "stash.sv"},
		{name: "stash spawn is not a hero", file: "stash_spawn.sv"},
		{name: "wrong extension", file: "single_0.sav"},
		{name: "slot out of range", file: "single_99.sv"},
		{name: "not a number", file: "single_x.sv"},
		{name: "unknown prefix", file: "hero_0.sv"},
	}
}

// requireSaveName compares a classification with its expectation.
func requireSaveName(t *testing.T, want saveNameCase, got saveNameResult) {
	t.Helper()

	if got.ok != want.ok || got.prefix != want.prefix || got.slot != want.slot {
		t.Fatalf("parse(%q) = %+v, want (%q, %d, %v)", want.file, got, want.prefix, want.slot, want.ok)
	}
}

// writeStashArchive writes a valid legacy stash archive.
func writeStashArchive(t *testing.T, path string, mode persist.GameMode) {
	t.Helper()

	stash := persist.StashSave{CurrentPage: 1}
	itemSize := persist.ItemSizeFor(mode)

	archive := persist.LegacySVStore{Path: path, Mode: mode}
	if err := archive.WriteStash(stash, itemSize); err != nil {
		t.Fatalf("write stash %s: %v", path, err)
	}
}

// openTempSQLiteStore opens a store on a throwaway file.
func openTempSQLiteStore(t *testing.T) *persist.SQLiteStore {
	t.Helper()

	store, err := persist.OpenSQLiteStore(filepath.Join(t.TempDir(), "hero.db"))
	if err != nil {
		t.Fatalf("OpenSQLiteStore: %v", err)
	}

	t.Cleanup(func() { _ = store.Close() })

	return store
}

// writeHeroArchive writes a legacy save named after its slot, with the mode its
// prefix implies.
func writeHeroArchive(t *testing.T, path, name string) {
	t.Helper()

	prefix, slot, ok := persist.ParseHeroSaveName(filepath.Base(path))
	if !ok {
		t.Fatalf("not a hero save name: %s", path)
	}

	game, err := persist.GameHeader{CurrLevel: 1}.Encode()
	if err != nil {
		t.Fatalf("GameHeader.Encode: %v", err)
	}

	pack := goldenPlayerPack()
	copy(pack.Name[:], name)

	archive := persist.LegacySVStore{Path: path, Mode: saveMode(prefix)}
	hero := persist.HeroRecord{Slot: slot, Pack: pack.Encode()}

	if err := archive.Write(hero, game); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// saveMode maps a save prefix to its mode.
func saveMode(prefix string) persist.GameMode {
	switch prefix {
	case "share_":
		return persist.GameMode{Spawn: true, Multiplayer: true}
	case "spawn_":
		return persist.GameMode{Spawn: true}
	case "multi_":
		return persist.GameMode{Multiplayer: true}
	default:
		return persist.GameMode{}
	}
}

// writeTextFile writes a file the importer must skip.
func writeTextFile(t *testing.T, path string) {
	t.Helper()

	if err := os.WriteFile(path, []byte("not a save"), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// heroSaveName is the hero name the fixture writes for a slot.
func heroSaveName(slot int) string {
	return "Hero" + string(rune('A'+slot))
}
