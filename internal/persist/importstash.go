package persist

import (
	"context"
	"fmt"
	"strings"
)

// GlobalStashSlot is the stash row of the shared (non-hero) stash: the
// PostgreSQL-style "no owning hero" marker the store documents for
// StashRecord.Slot.
const GlobalStashSlot = -1

// SpawnStashSlot is the stash row of the spawn (shareware) stash, the second
// half of the family pair upstream keeps in one stash record (pfile.cpp:96-104).
const SpawnStashSlot = -2

// ImportStash reads a legacy stash archive and writes it into the store row
// its family name selects: stash.sv into the global row, stash_spawn.sv into
// the spawn row (pfile.cpp:91-104).
func ImportStash(ctx context.Context, store Store, path, name string) error {
	slot, mode, err := stashFamilyFor(name)
	if err != nil {
		return fmt.Errorf("persist: import stash %s: %w", path, err)
	}

	return importStashInto(ctx, store, path, mode, slot)
}

// ExportStash writes a stored stash row back into a legacy stash archive, the
// inverse of ImportStash.
func ExportStash(ctx context.Context, store Store, path string, mode GameMode, slot int) error {
	record, err := store.LoadStash(ctx, slot)
	if err != nil {
		return fmt.Errorf("persist: export stash %s: %w", path, err)
	}

	stash, err := DecodeStashSave(record.Blob, ItemSizeFor(mode))
	if err != nil {
		return fmt.Errorf("persist: export stash %s: %w", path, err)
	}

	archive := LegacySVStore{Path: path, Mode: mode}
	if err := archive.WriteStash(stash, ItemSizeFor(mode)); err != nil {
		return fmt.Errorf("persist: export stash %s: %w", path, err)
	}

	return nil
}

// importStashInto reads a legacy stash archive and writes it into the given
// stash row. The blob kept in the stash row is the stash stream re-encoded with
// the mode's item size, so ExportStash can hand it back unchanged; PageCount is
// the number of pages the stream carries (loadsave.cpp:2712-2765).
func importStashInto(ctx context.Context, store Store, path string, mode GameMode, slot int) error {
	archive := LegacySVStore{Path: path, Mode: mode}

	stash, err := archive.ReadStash(ItemSizeFor(mode))
	if err != nil {
		return fmt.Errorf("persist: import stash %s: %w", path, err)
	}

	blob, err := stash.Encode(ItemSizeFor(mode))
	if err != nil {
		return fmt.Errorf("persist: import stash %s: %w", path, err)
	}

	record := StashRecord{Slot: slot, PageCount: len(stash.Pages), Blob: blob}
	if err := store.SaveStash(ctx, record); err != nil {
		return fmt.Errorf("persist: import stash %s: %w", path, err)
	}

	return nil
}

// stashFamilyFor resolves a stash file name to its store row and mode: the
// one place the family pair is decided (pfile.cpp:91-104).
func stashFamilyFor(name string) (int, GameMode, error) {
	switch name {
	case "stash.sv":
		return GlobalStashSlot, GameMode{}, nil
	case "stash_spawn.sv":
		return SpawnStashSlot, GameMode{Spawn: true}, nil
	default:
		return 0, GameMode{}, fmt.Errorf("persist: unknown stash family %q", name)
	}
}

// isStashSaveName reports the stash file names GetStashSavePath produces
// (pfile.cpp:91-99): "stash" and "stash_spawn", with the .sv extension.
func isStashSaveName(name string) bool {
	if !strings.HasSuffix(name, ".sv") {
		return false
	}

	base := strings.TrimSuffix(name, ".sv")

	return base == "stash" || base == "stash_spawn"
}
