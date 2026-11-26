package persist

import (
	"context"
	"database/sql"
)

// SeedRawSQL exists only to make the canary's no-raw-sql step red.
func SeedRawSQL(ctx context.Context, db *sql.DB) *sql.Row {
	return db.QueryRowContext(ctx, "SELECT 1")
}
