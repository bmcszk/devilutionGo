package persist

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// ImportFailure is one save that could not be imported.
type ImportFailure struct {
	// Path is the file that failed.
	Path string
	// Err is why it failed.
	Err error
}

// ImportReport is the outcome of importing a whole save directory.
type ImportReport struct {
	// Imported lists the save files that were written into the store, in
	// path order.
	Imported []string
	// Failed lists the saves that were rejected, with the reason.
	Failed []ImportFailure
}

// Errors reports whether any save was rejected.
func (r ImportReport) Errors() bool {
	return len(r.Failed) > 0
}

// Summary summarises the rejected saves so a CLI can print one line and exit
// non-zero.
func (r ImportReport) Summary() string {
	if len(r.Failed) == 0 {
		return ""
	}

	return fmt.Sprintf("%d save(s) failed to import, first: %s: %v", len(r.Failed), r.Failed[0].Path,
		r.Failed[0].Err)
}

// ImportDirectory imports every legacy .sv hero save in dir into the store: the
// bulk form of ImportSV that the one-time importer CLI drives.
//
// Stash files (stash.sv / stash_spawn.sv, pfile.cpp:91-99) are not hero saves;
// they are imported separately into the global stash row (GlobalStashSlot).
func ImportDirectory(ctx context.Context, store Store, dir string) (ImportReport, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ImportReport{}, fmt.Errorf("persist: read save directory %s: %w", dir, err)
	}

	report := ImportReport{}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		report.add(importSaveFile(ctx, store, dir, entry.Name()))
	}

	slices.Sort(report.Imported)

	return report, nil
}

// ParseHeroSaveName splits "<prefix><slot>.sv" into its prefix and slot, the
// inverse of GetSavePath's naming (pfile.cpp:81-88). It reports false for
// anything that is not a hero save, including the stash files.
func ParseHeroSaveName(name string) (string, int, bool) {
	for _, prefix := range heroSavePrefixes() {
		if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, ".sv") {
			continue
		}

		digits := strings.TrimSuffix(strings.TrimPrefix(name, prefix), ".sv")
		slot, err := strconv.Atoi(digits)
		if err != nil || slot < 0 || slot >= MaxCharacters {
			return "", 0, false
		}

		return prefix, slot, true
	}

	return "", 0, false
}

// add records the outcome of one entry.
func (r *ImportReport) add(path string, failure *ImportFailure) {
	if failure != nil {
		r.Failed = append(r.Failed, *failure)

		return
	}

	if path != "" {
		r.Imported = append(r.Imported, path)
	}
}

// importSaveFile imports one directory entry: a hero save or a stash, returning
// its path when it landed and a failure when it was a save that did not.
func importSaveFile(ctx context.Context, store Store, dir, name string) (string, *ImportFailure) {
	if isStashSaveName(name) {
		return importStashFile(ctx, store, dir, name)
	}

	return importHeroSave(ctx, store, dir, name)
}

// importStashFile imports the shared stash into the global stash row.
func importStashFile(ctx context.Context, store Store, dir, name string) (string, *ImportFailure) {
	path := filepath.Join(dir, name)
	mode := GameMode{Spawn: name == "stash_spawn.sv"}

	if err := ImportStash(ctx, store, path, mode, GlobalStashSlot); err != nil {
		return "", &ImportFailure{Path: path, Err: err}
	}

	return path, nil
}

// importHeroSave imports one directory entry, returning its path when it was a
// hero save that landed and a failure when it was a hero save that did not.
func importHeroSave(ctx context.Context, store Store, dir, name string) (string, *ImportFailure) {
	prefix, slot, ok := ParseHeroSaveName(name)
	if !ok {
		return "", nil
	}

	path := filepath.Join(dir, name)

	if err := ImportSV(ctx, store, path, modeForPrefix(prefix), slot); err != nil {
		return "", &ImportFailure{Path: path, Err: err}
	}

	return path, nil
}

// heroSavePrefixes are the four GetSavePath prefixes (pfile.cpp:81-88), the
// files a one-time import has to pick up.
func heroSavePrefixes() []string {
	return []string{"single_", "multi_", "spawn_", "share_"}
}

// modeForPrefix maps a GetSavePath prefix back to its save mode
// (pfile.cpp:81-88).
func modeForPrefix(prefix string) GameMode {
	switch prefix {
	case "share_":
		return GameMode{Spawn: true, Multiplayer: true}
	case "spawn_":
		return GameMode{Spawn: true}
	case "multi_":
		return GameMode{Multiplayer: true}
	default:
		return GameMode{}
	}
}
