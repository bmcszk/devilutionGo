package persist

import (
	"database/sql"
	"errors"
	"fmt"
)

// SQL statements of the live save store; the schema follows the proposal in
// docs/research/dg-requirements.md ("Data structures (SQLite schema
// proposal)").
const (
	schemaSQL = `
CREATE TABLE IF NOT EXISTS meta (key TEXT PRIMARY KEY, value BLOB);
CREATE TABLE IF NOT EXISTS heroes (
  id INTEGER PRIMARY KEY,
  slot INTEGER UNIQUE,
  name TEXT,
  class_id INTEGER,
  created_at INTEGER,
  updated_at INTEGER,
  pack BLOB NOT NULL,
  spawn INTEGER,
  hellfire INTEGER,
  multiplayer INTEGER
);
CREATE TABLE IF NOT EXISTS games (
  hero_slot INTEGER,
  kind TEXT,
  data BLOB NOT NULL,
  updated_at INTEGER,
  PRIMARY KEY (hero_slot, kind)
);
CREATE TABLE IF NOT EXISTS stash (
  hero_slot INTEGER PRIMARY KEY,
  page_count INTEGER,
  blob BLOB
);`

	upsertHeroSQL = `
INSERT INTO heroes (slot, name, class_id, created_at, updated_at, pack, spawn, hellfire, multiplayer)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(slot) DO UPDATE SET
  name = excluded.name, class_id = excluded.class_id, updated_at = excluded.updated_at,
  pack = excluded.pack, spawn = excluded.spawn, hellfire = excluded.hellfire,
  multiplayer = excluded.multiplayer`

	selectHeroSQL = `
SELECT slot, name, class_id, created_at, updated_at, pack, spawn, hellfire, multiplayer
FROM heroes WHERE slot = ?`

	listHeroesSQL = `
SELECT slot, name, class_id, created_at, updated_at, pack, spawn, hellfire, multiplayer
FROM heroes ORDER BY slot`

	deleteHeroSQL  = `DELETE FROM heroes WHERE slot = ?`
	deleteGameSQL  = `DELETE FROM games WHERE hero_slot = ?`
	deleteStashSQL = `DELETE FROM stash WHERE hero_slot = ?`

	upsertGameSQL = `
INSERT INTO games (hero_slot, kind, data, updated_at) VALUES (?, ?, ?, ?)
ON CONFLICT(hero_slot, kind) DO UPDATE SET data = excluded.data, updated_at = excluded.updated_at`

	selectGameSQL = `SELECT data FROM games WHERE hero_slot = ? AND kind = ?`

	upsertStashSQL = `
INSERT INTO stash (hero_slot, page_count, blob) VALUES (?, ?, ?)
ON CONFLICT(hero_slot) DO UPDATE SET page_count = excluded.page_count, blob = excluded.blob`

	selectStashSQL = `SELECT page_count, blob FROM stash WHERE hero_slot = ?`
)

// scanHero reads one heroes-table row in the column order the SELECTs use.
func scanHero(row interface{ Scan(...any) error }) (HeroRecord, error) {
	var (
		hero        HeroRecord
		spawn       int
		hellfire    int
		multiplayer int
	)

	err := row.Scan(&hero.Slot, &hero.Name, &hero.ClassID, &hero.CreatedAt, &hero.UpdatedAt,
		&hero.Pack, &spawn, &hellfire, &multiplayer)
	if err != nil {
		return HeroRecord{}, wrapNoRows(err)
	}

	hero.Spawn = spawn != 0
	hero.Hellfire = hellfire != 0
	hero.Multiplayer = multiplayer != 0

	return hero, nil
}

// validateHero rejects records the store cannot represent: slots outside
// MAX_CHARACTERS and packs of the wrong size.
func validateHero(hero HeroRecord) error {
	if hero.Slot < 0 || hero.Slot >= MaxCharacters {
		return fmt.Errorf("persist: hero slot %d: %w", hero.Slot, ErrSlotOutOfRange)
	}

	if len(hero.Pack) != PlayerPackSize {
		return fmt.Errorf("persist: hero %q pack len %d != %d: %w", hero.Name, len(hero.Pack), PlayerPackSize,
			ErrShortRecord)
	}

	return nil
}

// wrapNoRows turns sql.ErrNoRows into ErrHeroNotFound so callers can match on
// one package error.
func wrapNoRows(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrHeroNotFound
	}

	return err
}
