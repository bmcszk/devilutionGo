package persist_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

// TestDeleteHeroFailureKeepsDependentRows pins the atomicity of the four
// sequential deletes (store_sqlite.go DeleteHero, PR14 finding): a delete that
// fails midway must roll the earlier ones back, never leave partial state.
func TestDeleteHeroFailureKeepsDependentRows(t *testing.T) {
	// given a hero with a stash row, a level row and a game stream
	ctx := context.Background()
	path := seedCascadingHero(t)

	// and a dependent delete that fails after the stash and levels deletes ran
	injectGamesDeleteFailure(t, path)

	store := reopenStore(ctx, t, path)

	// when the hero is deleted
	err := store.DeleteHero(ctx, 5)
	// then the delete reports the injected failure...
	if err == nil {
		t.Fatal("DeleteHero = nil, want the injected games delete failure")
	}
	// ...and every row is still there: nothing was partially removed
	requireCascadeIntact(ctx, t, store)
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

// injectGamesDeleteFailure makes the third delete of deleteHeroRows fail, so a
// non-transactional DeleteHero would already have dropped the stash and level
// rows by the time it errors.
func injectGamesDeleteFailure(t *testing.T, path string) {
	t.Helper()

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	defer func() { _ = db.Close() }()

	const trigger = `CREATE TRIGGER fail_games_delete BEFORE DELETE ON games
BEGIN
  SELECT RAISE(ABORT, 'injected games delete failure');
END`

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
