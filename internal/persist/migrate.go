package persist

import (
	"context"
	"fmt"
)

// migration is one step of the PRAGMA user_version chain: the DDL to apply and
// the version it stamps once applied.
type migration struct {
	version int
	ddl     string
}

// migrations is the ordered chain SQLiteStore.migrate applies. Each step is a
// committed schema file; sqlc is generated against the newest one.
func migrations() []migration {
	return []migration{
		{version: 1, ddl: schemaFile},
		{version: 2, ddl: schemaV2File},
	}
}

// migrate brings an open database to SchemaVersion.
func (s *SQLiteStore) migrate() error {
	current, err := s.SchemaVersion(context.Background())
	if err != nil {
		return err
	}

	// A database stamped by a newer build is left untouched: this build does
	// not know what its extra columns mean, and overwriting the stamp would
	// hide the mismatch.
	if current > SchemaVersion {
		return fmt.Errorf("persist: database user_version %d > %d: %w", current, SchemaVersion, ErrSchemaTooNew)
	}

	for _, step := range migrations() {
		if step.version <= current {
			continue
		}

		if err := s.applyMigration(step); err != nil {
			return err
		}
	}

	return nil
}

// applyMigration runs one step's DDL and stamps its version.
// The two statements are not one transaction: every schema file uses CREATE
// TABLE IF NOT EXISTS, so a crash in between replays the DDL harmlessly.
func (s *SQLiteStore) applyMigration(step migration) error {
	if _, err := s.db.ExecContext(context.Background(), step.ddl); err != nil {
		return fmt.Errorf("persist: apply migration %d: %w", step.version, err)
	}

	// PRAGMA takes no bound parameters, so the version is formatted in; see the
	// exception note on userVersionSQL.
	stmt := fmt.Sprintf("PRAGMA user_version = %d", step.version)
	if _, err := s.db.ExecContext(context.Background(), stmt); err != nil {
		return fmt.Errorf("persist: set user_version %d: %w", step.version, err)
	}

	return nil
}
