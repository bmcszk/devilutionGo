package persist_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

// columnSpec is one column of the migration schema as SQLite reports it.
type columnSpec struct {
	name    string
	notNull bool
	pk      int
}

// TestSchemaColumns pins the v1 migration schema applied by the store: the
// embedded schema.sql is what sqlc is generated against, so a column that
// changes there changes the generated types too, and this test is what catches
// the two drifting apart silently.
func TestSchemaColumns(t *testing.T) {
	path := filepath.Join(t.TempDir(), "schema.db")

	store, err := persist.OpenSQLiteStore(path)
	if err != nil {
		t.Fatalf("OpenSQLiteStore: %v", err)
	}

	if err := store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	db := openSchemaDB(t, path)
	defer func() { _ = db.Close() }()

	for table, want := range schemaColumns() {
		t.Run(table, func(t *testing.T) {
			// when the applied schema is inspected
			got := tableColumns(t, db, table)

			// then it has exactly the committed columns
			requireColumns(t, table, got, want)
		})
	}
}

// openSchemaDB opens the store file directly, bypassing the store so the test
// reads what SQLite actually created.
func openSchemaDB(t *testing.T, path string) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open schema db: %v", err)
	}

	return db
}

// tableColumns reads PRAGMA table_info for one table.
func tableColumns(t *testing.T, db *sql.DB, table string) []columnSpec {
	t.Helper()

	rows, err := db.QueryContext(context.Background(),
		"SELECT name, \"notnull\", pk FROM pragma_table_info(?)", table)
	if err != nil {
		t.Fatalf("table_info %s: %v", table, err)
	}

	defer func() { _ = rows.Close() }()

	var columns []columnSpec

	for rows.Next() {
		var column columnSpec
		if err := rows.Scan(&column.name, &column.notNull, &column.pk); err != nil {
			t.Fatalf("scan table_info %s: %v", table, err)
		}

		columns = append(columns, column)
	}

	if err := rows.Err(); err != nil {
		t.Fatalf("table_info %s: %v", table, err)
	}

	return columns
}

// requireColumns compares the applied columns with the committed ones.
func requireColumns(t *testing.T, table string, got, want []columnSpec) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("%s has %d columns, want %d: got %v", table, len(got), len(want), got)
	}

	for i := range want {
		if got[i].name != want[i].name || got[i].notNull != want[i].notNull || got[i].pk != want[i].pk {
			t.Fatalf("%s column %d = %+v, want %+v", table, i, got[i], want[i])
		}
	}
}

// schemaColumns is the v1 schema (internal/persist/schema.sql) as SQLite
// reports it: column order, NOT NULL and primary-key position.
func schemaColumns() map[string][]columnSpec {
	return map[string][]columnSpec{
		"meta": {
			{name: "key", notNull: true, pk: 1},
			{name: "value", notNull: true},
		},
		"heroes": {
			{name: "id", pk: 1},
			{name: "slot", notNull: true},
			{name: "name", notNull: true},
			{name: "class_id", notNull: true},
			{name: "created_at", notNull: true},
			{name: "updated_at", notNull: true},
			{name: "pack", notNull: true},
			{name: "spawn", notNull: true},
			{name: "hellfire", notNull: true},
			{name: "multiplayer", notNull: true},
		},
		"games": {
			{name: "hero_slot", notNull: true, pk: 1},
			{name: "kind", notNull: true, pk: 2},
			{name: "data", notNull: true},
			{name: "updated_at", notNull: true},
		},
		"stash": {
			{name: "hero_slot", notNull: true, pk: 1},
			{name: "page_count", notNull: true},
			{name: "blob"},
		},
	}
}
