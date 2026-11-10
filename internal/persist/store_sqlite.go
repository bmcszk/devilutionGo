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
const SchemaVersion = 3

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
	// Slot is the owning hero slot; -1 is the global stash row
	// (GlobalStashSlot), -2 the spawn stash row (SpawnStashSlot).
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
func OpenSQLiteStore(ctx context.Context, path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("persist: open %s: %w", path, err)
	}

	store := &SQLiteStore{db: db, queries: sqlc.New(db)}

	if err := store.migrate(ctx); err != nil {
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

	nowSec := time.Now().Unix()
	if hero.CreatedAt == 0 {
		hero.CreatedAt = nowSec
	}

	hero.UpdatedAt = nowSec

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

// DeleteHero removes a hero row and everything hanging off it in one
// transaction, so a mid-way failure cannot orphan dependent rows.
func (s *SQLiteStore) DeleteHero(ctx context.Context, slot int) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("persist: begin delete hero slot %d: %w", slot, err)
	}

	if err := deleteHeroRows(ctx, s.queries.WithTx(tx), slot); err != nil {
		_ = tx.Rollback()

		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("persist: commit delete hero slot %d: %w", slot, err)
	}

	return nil
}

// deleteHeroRows removes the dependent rows, then the hero itself.
func deleteHeroRows(ctx context.Context, queries *sqlc.Queries, slot int) error {
	deletes := []struct {
		table string
		run   func(context.Context, int64) error
	}{
		{"stash", queries.DeleteStash},
		{"levels", queries.DeleteLevels},
		{"games", queries.DeleteGame},
		{"heroes", queries.DeleteHero},
	}

	for _, del := range deletes {
		if err := del.run(ctx, int64(slot)); err != nil {
			return fmt.Errorf("persist: delete hero slot %d: %s: %w", slot, del.table, err)
		}
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

// SaveLevel stores one dungeon level's save data for a hero slot. The data is
// the archive entry body upstream writes as "perml<NN>"
// (loadsave.cpp:2771-2900); SQLite keys it by slot and level id instead.
func (s *SQLiteStore) SaveLevel(ctx context.Context, slot, level int, data []byte) error {
	err := s.queries.UpsertLevel(ctx, sqlc.UpsertLevelParams{
		HeroSlot:  int64(slot),
		LevelID:   int64(level),
		Data:      data,
		UpdatedAt: time.Now().Unix(),
	})
	if err != nil {
		return fmt.Errorf("persist: save level %d of slot %d: %w", level, slot, err)
	}

	return nil
}

// LoadLevel reads one dungeon level's save data.
func (s *SQLiteStore) LoadLevel(ctx context.Context, slot, level int) ([]byte, error) {
	data, err := s.queries.GetLevel(ctx, sqlc.GetLevelParams{HeroSlot: int64(slot), LevelID: int64(level)})
	if err != nil {
		return nil, fmt.Errorf("persist: load level %d of slot %d: %w", level, slot, wrapNoRows(err))
	}

	return data, nil
}
