package persist_test

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

// TestSaveLevelRoundTrip checks per-level data survives the store, which is
// what replaces the archive's per-level entries (loadsave.cpp:2771-2900).
func TestSaveLevelRoundTrip(t *testing.T) {
	ctx := context.Background()
	store := openTempSQLiteStore(ctx, t)

	town := []byte("town level")
	dungeon := []byte("cathedral 1")

	// when two levels of one hero are written
	for level, data := range map[int][]byte{0: town, 1: dungeon} {
		if err := store.SaveLevel(ctx, 0, level, data); err != nil {
			t.Fatalf("SaveLevel %d: %v", level, err)
		}
	}

	// then each reads back unchanged
	requireLevel(t, store, 0, town)
	requireLevel(t, store, 1, dungeon)

	// and an unwritten level is a missing row, not an empty one
	requireLevelMissing(t, store, 15)
}

// TestDeleteHeroRemovesLevels checks the cascade: a deleted hero leaves no
// level rows behind, so a later hero in the same slot starts clean.
func TestDeleteHeroRemovesLevels(t *testing.T) {
	ctx := context.Background()
	store := openTempSQLiteStore(ctx, t)

	hero := migrationHero(3, "Cascade")
	if err := store.SaveHero(ctx, hero); err != nil {
		t.Fatalf("SaveHero: %v", err)
	}

	if err := store.SaveLevel(ctx, 3, 1, []byte("level")); err != nil {
		t.Fatalf("SaveLevel: %v", err)
	}

	// when the hero is deleted
	if err := store.DeleteHero(ctx, 3); err != nil {
		t.Fatalf("DeleteHero: %v", err)
	}

	// then its level rows are gone too
	if _, err := store.LoadLevel(ctx, 3, 1); !errors.Is(err, persist.ErrHeroNotFound) {
		t.Fatalf("LoadLevel after delete = %v, want ErrHeroNotFound", err)
	}
}

// TestMigrateV1ToV2 checks the chain actually steps: a database left at
// version 1 gets the version-2 table on the next open, and its rows survive.
func TestMigrateV1ToV2(t *testing.T) {
	path := filepath.Join(t.TempDir(), "v1.db")
	ctx := context.Background()

	// given a database that holds a hero at version 1
	store, err := persist.OpenSQLiteStore(ctx, path)
	if err != nil {
		t.Fatalf("OpenSQLiteStore: %v", err)
	}

	hero := migrationHero(1, "Old")
	if err := store.SaveHero(ctx, hero); err != nil {
		t.Fatalf("SaveHero: %v", err)
	}

	if err := store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	downgradeToV1(t, path)

	// when it is opened by this build
	upgraded, err := persist.OpenSQLiteStore(ctx, path)
	if err != nil {
		t.Fatalf("upgrade open: %v", err)
	}

	defer func() { _ = upgraded.Close() }()

	// then it is at the current version, the old row is intact and levels work
	requireVersion(t, upgraded)

	loaded, err := upgraded.LoadHero(ctx, hero.Slot)
	if err != nil || loaded.Name != hero.Name {
		t.Fatalf("LoadHero after migration = (%q, %v), want %q", loaded.Name, err, hero.Name)
	}

	if err := upgraded.SaveLevel(ctx, hero.Slot, 1, []byte("post-migration")); err != nil {
		t.Fatalf("SaveLevel after migration: %v", err)
	}
}

// requireLevel asserts one stored level's bytes.
func requireLevel(t *testing.T, store *persist.SQLiteStore, level int, want []byte) {
	t.Helper()

	got, err := store.LoadLevel(context.Background(), 0, level)
	if err != nil {
		t.Fatalf("LoadLevel %d: %v", level, err)
	}

	if string(got) != string(want) {
		t.Fatalf("level %d = %q, want %q", level, got, want)
	}
}

// requireLevelMissing asserts the level row does not exist.
func requireLevelMissing(t *testing.T, store *persist.SQLiteStore, level int) {
	t.Helper()

	if _, err := store.LoadLevel(context.Background(), 0, level); !errors.Is(err, persist.ErrHeroNotFound) {
		t.Fatalf("LoadLevel %d error = %v, want ErrHeroNotFound", level, err)
	}
}

// downgradeToV1 strips the version-2 table and restamps the file, simulating a
// save written by the previous build.
func downgradeToV1(t *testing.T, path string) {
	t.Helper()

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	defer func() { _ = db.Close() }()

	for _, stmt := range []string{"DROP TABLE IF EXISTS levels", "PRAGMA user_version = 1"} {
		if _, err := db.ExecContext(context.Background(), stmt); err != nil {
			t.Fatalf("%s: %v", stmt, err)
		}
	}
}
