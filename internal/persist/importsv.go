package persist

import (
	"bytes"
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

	// The store keeps the hero name and class as their own columns so the
	// character list can be read without decoding every pack; take them from
	// the decoded pack (pfile_read_hero fills hero_names from pName,
	// pfile.cpp:613-628; Game2UiPlayer copies _pClass into heroclass,
	// pfile.cpp:215).
	if err := fillHeroColumns(&hero); err != nil {
		return fmt.Errorf("persist: import %s: %w", path, err)
	}

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
//
// Manual verification of "export readable by devilutionX" (bean devilGo-6xu2
// AC4): CI cannot run the C++ binary, so this check is manual and is repeated
// only when the archive layout changes, against pinned upstream c20b2f2456.
//
// Upstream has no standalone save-reader binary; the two probes are gtests in
// the upstream test/ tree (run them from an upstream build directory):
//
//  1. Build the upstream gtest suite and run the Writehero case
//     (test/writehero_test.cpp:365). It writes multi_0.sv through
//     pfile_write_hero and hashes the whole file with picosha2
//     (test/writehero_test.cpp:414-421). That hash is container-level, not
//     payload-level: the comment above it (test/writehero_test.cpp:415-419)
//     records that the archive layout is not byte-stable, so a hash mismatch
//     alone does NOT prove incompatibility.
//  2. The payload-level check is test/timedemo_test.cpp:94, which calls
//     pfile_compare_hero_demo(demoNumber, true) and asserts
//     HeroCompareResult::Same. That path compares *decoded* archive entries
//     (pfile.cpp:476-523) and ignores container order, so it is the check to
//     trust for our exports.
//
// To verify one of our exports by hand: put it in the upstream pref path as
// single_<slot>.sv (or multi_/spawn_/share_ per SaveNamespace.PrimaryPrefix),
// let upstream load the hero and save again, then compare the two archives
// with CompareSaves in demo_compare_saves.go: identical decoded entries mean
// upstream understood every byte. Step 2 needs no hash of our file.
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

// fillHeroColumns copies the hero's list columns out of the decoded pack into
// the record.
func fillHeroColumns(hero *HeroRecord) error {
	pack, err := DecodePlayerPack(hero.Pack)
	if err != nil {
		return fmt.Errorf("persist: decode pack columns: %w", err)
	}

	name := pack.Name[:]
	if end := bytes.IndexByte(name, 0); end >= 0 {
		name = name[:end]
	}

	hero.Name = string(name)
	hero.ClassID = int(pack.Class)

	return nil
}
