package persist_test

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

// TestDeleteHeroFailureKeepsDependentRows pins DeleteHero's rollback (PR14
// finding 1): a delete failing at any step must leave the hero and every
// dependent row intact, never a partial state.
func TestDeleteHeroFailureKeepsDependentRows(t *testing.T) {
	for _, table := range []string{"stash", "levels", "games", "heroes"} {
		t.Run("abort on "+table+" delete", func(t *testing.T) {
			// given a hero with a stash row, a level row and a game stream
			ctx := context.Background()
			path := seedCascadingHero(t)
			abortDeletesOf(t, path, table)

			store := reopenStore(ctx, t, path)

			// when the hero is deleted
			err := store.DeleteHero(ctx, 5)

			// then the injected failure surfaces and nothing was partially removed
			requireInjectedFailure(t, err, table)
			requireCascadeIntact(ctx, t, store)
		})
	}
}

// requireInjectedFailure asserts DeleteHero reported the abort trigger of table.
func requireInjectedFailure(t *testing.T, err error, table string) {
	t.Helper()

	if err == nil {
		t.Fatalf("DeleteHero = nil, want the injected %s delete failure", table)
	}

	marker := fmt.Sprintf("injected %s delete failure", table)
	if !strings.Contains(err.Error(), marker) {
		t.Fatalf("DeleteHero = %v, want it to wrap %q", err, marker)
	}
}

// seedCascadingHero writes one hero with all three dependent row groups and
// returns the closed database path.
func seedCascadingHero(t *testing.T) string {
	t.Helper()

	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "hero.db")

	store, err := persist.OpenSQLiteStore(ctx, path)
	if err != nil {
		t.Fatalf("OpenSQLiteStore: %v", err)
	}

	hero := persist.HeroRecord{Slot: 5, Name: "Cascade", Pack: make([]byte, persist.PlayerPackSize)}
	if err := store.SaveHero(ctx, hero); err != nil {
		t.Fatalf("SaveHero: %v", err)
	}

	if err := store.SaveStash(ctx, persist.StashRecord{Slot: 5, PageCount: 1}); err != nil {
		t.Fatalf("SaveStash: %v", err)
	}

	if err := store.SaveLevel(ctx, 5, 1, []byte("cathedral")); err != nil {
		t.Fatalf("SaveLevel: %v", err)
	}

	if err := store.SaveGame(ctx, 5, "single", make([]byte, persist.GameHeaderSize)); err != nil {
		t.Fatalf("SaveGame: %v", err)
	}

	if err := store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	return path
}

// abortDeletesOf makes every DELETE on table abort, so DeleteHero fails at that
// step: a non-transactional delete would already have dropped the rows before it.
func abortDeletesOf(t *testing.T, path, table string) {
	t.Helper()

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	defer func() { _ = db.Close() }()

	// table is one of the four literals named by the caller, never user input
	trigger := fmt.Sprintf(`CREATE TRIGGER fail_%[1]s_delete BEFORE DELETE ON %[1]s
BEGIN
  SELECT RAISE(ABORT, 'injected %[1]s delete failure');
END`, table)

	if _, err := db.ExecContext(context.Background(), trigger); err != nil {
		t.Fatalf("create trigger: %v", err)
	}
}

// reopenStore opens the database the test mutated and closes it on cleanup.
func reopenStore(ctx context.Context, t *testing.T, path string) *persist.SQLiteStore {
	t.Helper()

	store, err := persist.OpenSQLiteStore(ctx, path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}

	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Fatalf("Close: %v", err)
		}
	})

	return store
}

// requireCascadeIntact asserts the hero and all dependent rows still load.
func requireCascadeIntact(ctx context.Context, t *testing.T, store *persist.SQLiteStore) {
	t.Helper()

	if _, err := store.LoadHero(ctx, 5); err != nil {
		t.Fatalf("LoadHero after failed delete: %v", err)
	}

	if _, err := store.LoadStash(ctx, 5); err != nil {
		t.Fatalf("LoadStash after failed delete: %v", err)
	}

	if _, err := store.LoadLevel(ctx, 5, 1); err != nil {
		t.Fatalf("LoadLevel after failed delete: %v", err)
	}

	if _, err := store.LoadGame(ctx, 5, "single"); err != nil {
		t.Fatalf("LoadGame after failed delete: %v", err)
	}
}
