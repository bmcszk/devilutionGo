package persist_test

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strconv"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

// TestMigrationFreshDatabase checks a new file is created at the current
// version.
func TestMigrationFreshDatabase(t *testing.T) {
	// when a store is opened on a file that does not exist yet
	store, err := persist.OpenSQLiteStore(filepath.Join(t.TempDir(), "fresh.db"))
	if err != nil {
		t.Fatalf("OpenSQLiteStore: %v", err)
	}

	defer func() { _ = store.Close() }()

	// then it is stamped with the current schema version
	requireVersion(t, store)
}

// TestMigrationKeepsExistingRows checks reopening a current database does not
// re-run the chain over live data.
func TestMigrationKeepsExistingRows(t *testing.T) {
	path := filepath.Join(t.TempDir(), "hero.db")

	hero := migrationHero(1, "Migration")

	// given a store that already holds a hero
	store, err := persist.OpenSQLiteStore(path)
	if err != nil {
		t.Fatalf("OpenSQLiteStore: %v", err)
	}

	if err := store.SaveHero(context.Background(), hero); err != nil {
		t.Fatalf("SaveHero: %v", err)
	}

	if err := store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// when the same file is opened again
	reopened, err := persist.OpenSQLiteStore(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}

	defer func() { _ = reopened.Close() }()

	// then the row survived and the version is unchanged
	loaded, err := reopened.LoadHero(context.Background(), hero.Slot)
	if err != nil {
		t.Fatalf("LoadHero: %v", err)
	}

	if loaded.Name != hero.Name {
		t.Fatalf("name = %q, want %q", loaded.Name, hero.Name)
	}

	requireVersion(t, reopened)
}

// TestMigrationReplaysUnstampedV1 checks the documented crash window of
// applyMigration: the version-1 DDL and its stamp are not one transaction, so a
// crash in between leaves a v1 schema stamped 0. Replaying the chain must be
// harmless (every schema file is CREATE TABLE IF NOT EXISTS) and must neither
// lose rows nor skip the version-2 step.
func TestMigrationReplaysUnstampedV1(t *testing.T) {
	// given a v1-era save whose user_version stamp never landed
	path := seedUnstampedV1(t)

	// when the chain replays from version 0
	reopened, err := persist.OpenSQLiteStore(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}

	defer func() { _ = reopened.Close() }()

	// then the file is at the current version with its row intact
	requireVersion(t, reopened)
	requireHeroName(t, reopened, 2, "Replay")
	// ...and the version-2 table is usable
	requireLevelsWritable(t, reopened, 2)
}

// TestMigrationRefusesNewerDatabase checks a save file written by a newer build
// is refused instead of being stamped back or migrated blindly.
func TestMigrationRefusesNewerDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "future.db")

	stampVersion(t, path, persist.SchemaVersion+1)

	// when this build opens a database from the future
	_, err := persist.OpenSQLiteStore(path)

	// then it refuses with ErrSchemaTooNew
	if !errors.Is(err, persist.ErrSchemaTooNew) {
		t.Fatalf("OpenSQLiteStore error = %v, want ErrSchemaTooNew", err)
	}
}

// seedUnstampedV1 writes a hero, drops the version stamp back to 0 and returns
// the file path: the state a crash between the v1 DDL and its stamp leaves.
func seedUnstampedV1(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "unstamped.db")

	store, err := persist.OpenSQLiteStore(path)
	if err != nil {
		t.Fatalf("OpenSQLiteStore: %v", err)
	}

	if err := store.SaveHero(context.Background(), migrationHero(2, "Replay")); err != nil {
		t.Fatalf("SaveHero: %v", err)
	}

	if err := store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	stampVersion(t, path, 0)

	return path
}

// requireHeroName asserts the stored hero name of one slot.
func requireHeroName(t *testing.T, store *persist.SQLiteStore, slot int, want string) {
	t.Helper()

	loaded, err := store.LoadHero(context.Background(), slot)
	if err != nil {
		t.Fatalf("LoadHero: %v", err)
	}

	if loaded.Name != want {
		t.Fatalf("name = %q, want %q", loaded.Name, want)
	}
}

// requireLevelsWritable asserts the version-2 level table accepts a round-trip.
func requireLevelsWritable(t *testing.T, store *persist.SQLiteStore, slot int) {
	t.Helper()

	ctx := context.Background()

	if err := store.SaveLevel(ctx, slot, 1, []byte("cathedral")); err != nil {
		t.Fatalf("SaveLevel: %v", err)
	}

	if _, err := store.LoadLevel(ctx, slot, 1); err != nil {
		t.Fatalf("LoadLevel: %v", err)
	}
}

// requireVersion asserts the store's user_version is the current schema.
func requireVersion(t *testing.T, store *persist.SQLiteStore) {
	t.Helper()

	got, err := store.SchemaVersion(context.Background())
	if err != nil {
		t.Fatalf("SchemaVersion: %v", err)
	}

	if got != persist.SchemaVersion {
		t.Fatalf("user_version = %d, want %d", got, persist.SchemaVersion)
	}
}

// stampVersion writes a user_version into a database file, simulating a save
// written by another build.
func stampVersion(t *testing.T, path string, version int) {
	t.Helper()

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	defer func() { _ = db.Close() }()

	if _, err := db.ExecContext(context.Background(), "CREATE TABLE IF NOT EXISTS probe (id INTEGER)"); err != nil {
		t.Fatalf("create probe: %v", err)
	}

	// the version is interpolated because PRAGMA takes no bound parameters
	stmt := "PRAGMA user_version = " + strconv.Itoa(version)
	if _, err := db.ExecContext(context.Background(), stmt); err != nil {
		t.Fatalf("stamp version: %v", err)
	}
}

// migrationHero builds a valid store record: a zero pack of the exact size
// validateHero requires.
func migrationHero(slot int, name string) persist.HeroRecord {
	return persist.HeroRecord{Slot: slot, Name: name, Pack: make([]byte, persist.PlayerPackSize)}
}
