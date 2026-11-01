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
	requireVersion(t, store, persist.SchemaVersion)
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

	requireVersion(t, reopened, persist.SchemaVersion)
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

// requireVersion asserts the store's user_version.
func requireVersion(t *testing.T, store *persist.SQLiteStore, want int) {
	t.Helper()

	got, err := store.SchemaVersion(context.Background())
	if err != nil {
		t.Fatalf("SchemaVersion: %v", err)
	}

	if got != want {
		t.Fatalf("user_version = %d, want %d", got, want)
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
