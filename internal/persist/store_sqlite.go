package persist

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	// Registers the pure-Go "sqlite" database/sql driver (ADR-009).
	_ "modernc.org/sqlite"

	"github.com/bmcszk/devilutionGo/internal/persist/sqlc"
)

// SchemaVersion is the current PRAGMA user_version of the live save store
// (docs/research/dg-requirements.md, SQLite schema proposal).
const SchemaVersion = 1

// Store is the live save store the game writes through; SQLiteStore is the
// native implementation and LegacySVStore the .sv import/export one
// (dg-requirements.md, "Go/Ebiten mapping decision").
type Store interface {
	SaveHero(ctx context.Context, hero HeroRecord) error
	LoadHero(ctx context.Context, slot int) (HeroRecord, error)
	ListHeroes(ctx context.Context) ([]HeroRecord, error)
	DeleteHero(ctx context.Context, slot int) error
	SaveGame(ctx context.Context, slot int, kind string, data []byte) error
	LoadGame(ctx context.Context, slot int, kind string) ([]byte, error)
	SaveStash(ctx context.Context, stash StashRecord) error
	LoadStash(ctx context.Context, slot int) (StashRecord, error)
	Close() error
}

// HeroRecord is one heroes-table row: the codec-ready PlayerPack bytes plus
// the mode flags that live outside the encrypted blob in a legacy .sv
// (dg-requirements.md, "Determinism constraints").
type HeroRecord struct {
	// Slot is the pfile slot, 0..MaxCharacters-1.
	Slot int
	// Name is the hero name (PlayerNameLength bytes in the pack).
	Name string
	// ClassID is the player class index.
	ClassID int
	// Pack is the PlayerPack image of PlayerPackSize bytes.
	Pack []byte
	// Spawn, Hellfire and Multiplayer are the save-file mode flags.
	Spawn       bool
	Hellfire    bool
	Multiplayer bool
	// CreatedAt and UpdatedAt are unix seconds; zero means "now".
	CreatedAt int64
	UpdatedAt int64
}

// StashRecord is one stash-table row: the shared stash pages.
type StashRecord struct {
	// Slot is the owning hero slot, or -1 for the global stash.
	Slot int
	// PageCount is the number of stash pages.
	PageCount int
	// Blob is the stash payload (item records in ItemPack layout).
	Blob []byte
}

// SQLiteStore is the modernc.org/sqlite-backed live save store: one database
// file per hero plus one global database for the shared stash, mirroring
// single_N.sv plus the shared-stash file (ADR-003).
type SQLiteStore struct {
	db      *sql.DB
	queries *sqlc.Queries
}

// OpenSQLiteStore opens (creating if needed) a save database and migrates it
// to SchemaVersion.
func OpenSQLiteStore(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("persist: open %s: %w", path, err)
	}

	store := &SQLiteStore{db: db, queries: sqlc.New(db)}

	if err := store.migrate(); err != nil {
		_ = db.Close()

		return nil, err
	}

	return store, nil
}

// SchemaVersion reports PRAGMA user_version of the open database.
func (s *SQLiteStore) SchemaVersion(ctx context.Context) (int, error) {
	var version int
	// PRAGMA cannot be parameterized and sqlc does not generate it, so this is
	// the one statement the store issues by hand; every data query is
	// generated from queries/store.sql.
	if err := s.db.QueryRowContext(ctx, userVersionSQL).Scan(&version); err != nil {
		return 0, fmt.Errorf("persist: read user_version: %w", err)
	}

	return version, nil
}

// SaveHero upserts a hero row by slot (heroes table, slot UNIQUE).
func (s *SQLiteStore) SaveHero(ctx context.Context, hero HeroRecord) error {
	if err := validateHero(hero); err != nil {
		return err
	}

	now := time.Now().Unix()
	if hero.CreatedAt == 0 {
		hero.CreatedAt = now
	}

	hero.UpdatedAt = now

	if err := s.queries.UpsertHero(ctx, heroToParams(hero)); err != nil {
		return fmt.Errorf("persist: save hero slot %d: %w", hero.Slot, err)
	}

	return nil
}

// LoadHero reads one hero row by slot.
func (s *SQLiteStore) LoadHero(ctx context.Context, slot int) (HeroRecord, error) {
	row, err := s.queries.GetHero(ctx, int64(slot))
	if err != nil {
		return HeroRecord{}, fmt.Errorf("persist: load hero slot %d: %w", slot, wrapNoRows(err))
	}

	return heroFromModel(row.Hero), nil
}

// ListHeroes returns every hero row ordered by slot, the order pfile walks
// hero_names in (pfile.cpp:700-710).
func (s *SQLiteStore) ListHeroes(ctx context.Context) ([]HeroRecord, error) {
	rows, err := s.queries.ListHeroes(ctx)
	if err != nil {
		return nil, fmt.Errorf("persist: list heroes: %w", err)
	}

	heroes := make([]HeroRecord, 0, len(rows))
	for _, row := range rows {
		heroes = append(heroes, heroFromModel(row.Hero))
	}

	return heroes, nil
}

// DeleteHero removes a hero row and everything hanging off it.
func (s *SQLiteStore) DeleteHero(ctx context.Context, slot int) error {
	if err := s.queries.DeleteStash(ctx, int64(slot)); err != nil {
		return fmt.Errorf("persist: delete hero slot %d: %w", slot, err)
	}

	if err := s.queries.DeleteGame(ctx, int64(slot)); err != nil {
		return fmt.Errorf("persist: delete hero slot %d: %w", slot, err)
	}

	if err := s.queries.DeleteHero(ctx, int64(slot)); err != nil {
		return fmt.Errorf("persist: delete hero slot %d: %w", slot, err)
	}

	return nil
}

// SaveGame upserts the SaveGameData stream of one hero/game-kind pair (games
// table). The blob is the un-codeced stream from GameHeader onward.
func (s *SQLiteStore) SaveGame(ctx context.Context, slot int, kind string, data []byte) error {
	if len(data) < GameHeaderSize {
		return fmt.Errorf("persist: save game %q: %w", kind, ErrShortRecord)
	}

	err := s.queries.UpsertGame(ctx, sqlc.UpsertGameParams{
		HeroSlot:  int64(slot),
		Kind:      kind,
		Data:      data,
		UpdatedAt: time.Now().Unix(),
	})
	if err != nil {
		return fmt.Errorf("persist: save game %q: %w", kind, err)
	}

	return nil
}

// LoadGame reads the SaveGameData stream of one hero/game-kind pair.
func (s *SQLiteStore) LoadGame(ctx context.Context, slot int, kind string) ([]byte, error) {
	data, err := s.queries.GetGame(ctx, sqlc.GetGameParams{HeroSlot: int64(slot), Kind: kind})
	if err != nil {
		return nil, fmt.Errorf("persist: load game %q: %w", kind, wrapNoRows(err))
	}

	return data, nil
}

// SaveStash upserts a stash row.
func (s *SQLiteStore) SaveStash(ctx context.Context, stash StashRecord) error {
	err := s.queries.UpsertStash(ctx, sqlc.UpsertStashParams{
		HeroSlot:  int64(stash.Slot),
		PageCount: int64(stash.PageCount),
		Blob:      stash.Blob,
	})
	if err != nil {
		return fmt.Errorf("persist: save stash slot %d: %w", stash.Slot, err)
	}

	return nil
}

// LoadStash reads a stash row.
func (s *SQLiteStore) LoadStash(ctx context.Context, slot int) (StashRecord, error) {
	row, err := s.queries.GetStash(ctx, int64(slot))
	if err != nil {
		return StashRecord{}, fmt.Errorf("persist: load stash slot %d: %w", slot, wrapNoRows(err))
	}

	return StashRecord{Slot: slot, PageCount: int(row.PageCount), Blob: row.Blob}, nil
}

// Close closes the underlying database handle.
func (s *SQLiteStore) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("persist: close store: %w", err)
	}

	return nil
}

// migrate runs the user_version migration chain (dg-requirements.md,
// "Versioning: user_version + a Go migrate() chain").
func (s *SQLiteStore) migrate() error {
	if _, err := s.db.ExecContext(context.Background(), schemaFile); err != nil {
		return fmt.Errorf("persist: create schema: %w", err)
	}

	// PRAGMA takes no bound parameters, so the version is formatted in; see the
	// exception note on userVersionSQL.
	stmt := fmt.Sprintf("PRAGMA user_version = %d", SchemaVersion)
	if _, err := s.db.ExecContext(context.Background(), stmt); err != nil {
		return fmt.Errorf("persist: set user_version: %w", err)
	}

	return nil
}
