package persist

import (
	"database/sql"
	_ "embed"
	"errors"
	"fmt"

	"github.com/bmcszk/devilutionGo/internal/persist/sqlc"
)

// userVersionSQL reads PRAGMA user_version, the migration head. It is the one
// statement issued by hand: PRAGMA takes no bound parameters and sqlc emits
// nothing for it, so the equivalent of a generated query does not exist.
const userVersionSQL = "PRAGMA user_version"

// schemaFile is the versioned DDL sqlc is generated against; embedding it
// keeps the applied schema and the generated types on one source of truth.
//
//go:embed schema.sql
var schemaFile string

// schemaV2File is the version-2 migration: per-level save rows.
//
//go:embed schema_v2.sql
var schemaV2File string

// schemaV3File is the version-3 migration: the tick-log column family.
//
//go:embed schema_v3.sql
var schemaV3File string

// heroFromModel maps one heroes row onto the store record.
func heroFromModel(model sqlc.Hero) HeroRecord {
	return HeroRecord{
		Slot:        int(model.Slot),
		Name:        model.Name,
		ClassID:     int(model.ClassID),
		Pack:        model.Pack,
		Spawn:       model.Spawn != 0,
		Hellfire:    model.Hellfire != 0,
		Multiplayer: model.Multiplayer != 0,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
}

// heroToParams maps a store record onto the upsert parameters.
func heroToParams(hero HeroRecord) sqlc.UpsertHeroParams {
	spawn, hellfire, multiplayer := int64(0), int64(0), int64(0)
	if hero.Spawn {
		spawn = 1
	}

	if hero.Hellfire {
		hellfire = 1
	}

	if hero.Multiplayer {
		multiplayer = 1
	}

	return sqlc.UpsertHeroParams{
		Slot:        int64(hero.Slot),
		Name:        hero.Name,
		ClassID:     int64(hero.ClassID),
		CreatedAt:   hero.CreatedAt,
		UpdatedAt:   hero.UpdatedAt,
		Pack:        hero.Pack,
		Spawn:       spawn,
		Hellfire:    hellfire,
		Multiplayer: multiplayer,
	}
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
