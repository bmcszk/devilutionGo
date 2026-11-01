package persist

import (
	"context"
	"errors"
	"fmt"
)

// Game kinds derived from the save mode (pfile.cpp:79-104).
const (
	// KindSingle is a single-player save.
	KindSingle GameKind = "single"
	// KindMulti is a multiplayer save.
	KindMulti GameKind = "multi"
	// KindSpawn is a shareware save.
	KindSpawn GameKind = "spawn"
	// KindSpawnMulti is a shareware multiplayer save.
	KindSpawnMulti GameKind = "spawn_multi"
)

// GameKind is the games-table kind of a hero save archive
// (docs/research/dg-requirements.md, games table: 'single'|'multi'|'spawn').
type GameKind string

// KindFor is the games-table kind of a save mode.
func KindFor(mode GameMode) GameKind {
	switch {
	case mode.Spawn && mode.Multiplayer:
		return KindSpawnMulti
	case mode.Spawn:
		return KindSpawn
	case mode.Multiplayer:
		return KindMulti
	default:
		return KindSingle
	}
}

// ImportSV reads a legacy .sv archive and writes its hero row plus its game
// stream into the store: the one-time .sv importer of the SQLite-only live
// format (ADR-003).
func ImportSV(ctx context.Context, store Store, path string, mode GameMode, slot int) error {
	archive := LegacySVStore{Path: path, Mode: mode}

	hero, game, err := archive.Read()
	if err != nil {
		return fmt.Errorf("persist: import %s: %w", path, err)
	}

	hero.Slot = slot

	if err := store.SaveHero(ctx, hero); err != nil {
		return fmt.Errorf("persist: import %s: %w", path, err)
	}

	if len(game) == 0 {
		return nil
	}

	if err := store.SaveGame(ctx, slot, string(KindFor(mode)), game); err != nil {
		return fmt.Errorf("persist: import %s: %w", path, err)
	}

	return nil
}

// ExportSV writes the stored hero row and game stream back into a legacy .sv
// archive, so a migrated save stays readable by upstream builds.
func ExportSV(ctx context.Context, store Store, path string, mode GameMode, slot int) error {
	hero, err := store.LoadHero(ctx, slot)
	if err != nil {
		return fmt.Errorf("persist: export %s: %w", path, err)
	}

	game, err := store.LoadGame(ctx, slot, string(KindFor(mode)))
	if err != nil && !isNotFound(err) {
		return fmt.Errorf("persist: export %s: %w", path, err)
	}

	archive := LegacySVStore{Path: path, Mode: mode}
	if err := archive.Write(hero, game); err != nil {
		return fmt.Errorf("persist: export %s: %w", path, err)
	}

	return nil
}

// isNotFound reports whether err is the store's missing-row error.
func isNotFound(err error) bool {
	return err != nil && errors.Is(err, ErrHeroNotFound)
}
