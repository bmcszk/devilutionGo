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

// ImportStash reads a legacy stash archive and writes it into the store.
//
// The blob kept in the stash row is the stash stream re-encoded with the mode's
// item size, so ExportStash can hand it back unchanged; PageCount is the number
// of pages the stream carries (loadsave.cpp:2712-2765).
func ImportStash(ctx context.Context, store Store, path string, mode GameMode, slot int) error {
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

// isStashSaveName reports the stash file names GetStashSavePath produces
// (pfile.cpp:91-99): "stash" and "stash_spawn", with the .sv extension.
func isStashSaveName(name string) bool {
	if !strings.HasSuffix(name, ".sv") {
		return false
	}

	base := strings.TrimSuffix(name, ".sv")

	return base == "stash" || base == "stash_spawn"
}
