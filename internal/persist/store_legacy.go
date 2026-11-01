package persist

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/suprsokr/go-mpq"
)

// Legacy .sv container entry names and sizing (Source/pfile.cpp:158, :178,
// Source/loadsave.cpp:2771).
const (
	// HeroEntryName is the archive entry holding the codec-encoded hero.
	HeroEntryName = "hero"
	// GameEntryName is the archive entry holding the SaveGameData stream.
	GameEntryName = "game"
	// saveArchiveEntries is the hash-table size for a hero archive: hero plus
	// game, with room for the level entries mpq.Create rounds up anyway.
	saveArchiveEntries = 16
)

// LegacySVStore reads and writes a legacy .sv archive: an MPQ container whose
// hero entry is codec-encoded with the mode password (pfile.cpp:171-179) and
// whose other entries are plain STORE sectors
// (docs/research/dg-requirements.md, "Gaps/risks").
type LegacySVStore struct {
	// Path is the .sv file.
	Path string
	// Mode selects the password and the magic of the contained stream.
	Mode GameMode
}

// Write stores a hero record and, when game is non-nil, its SaveGameData
// stream into a fresh archive (pfile.cpp:527-560: hero first, then game).
func (s LegacySVStore) Write(hero HeroRecord, game []byte) error {
	if err := validateHero(hero); err != nil {
		return err
	}

	password := HeroPassword(SaveNamespace{Spawn: s.Mode.Spawn, Multiplayer: s.Mode.Multiplayer})

	heroEntry, err := EncodeHeroEntry(hero.Pack, password)
	if err != nil {
		return err
	}

	return s.writeEntries(map[string][]byte{
		HeroEntryName: heroEntry,
		GameEntryName: game,
	})
}

// WriteStash stores a stash stream as the mode's stash entry
// (loadsave.cpp:2712-2722).
func (s LegacySVStore) WriteStash(stash StashSave, itemSize int) error {
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o750); err != nil {
		return fmt.Errorf("persist: stash dir: %w", err)
	}

	stream, err := stash.Encode(itemSize)
	if err != nil {
		return err
	}

	return s.writeEntries(map[string][]byte{
		StashEntryName(s.Mode): stream,
	})
}

// Read returns the hero record and its game stream, decoding the hero entry
// with the mode password (pfile.cpp:628-635).
func (s LegacySVStore) Read() (HeroRecord, []byte, error) {
	entries, err := s.readEntries([]string{HeroEntryName, GameEntryName})
	if err != nil {
		return HeroRecord{}, nil, err
	}

	password := HeroPassword(SaveNamespace{Spawn: s.Mode.Spawn, Multiplayer: s.Mode.Multiplayer})

	pack, err := DecodeHeroEntry(entries[HeroEntryName], password)
	if err != nil {
		return HeroRecord{}, nil, err
	}

	if len(pack) != PlayerPackSize {
		return HeroRecord{}, nil, fmt.Errorf("persist: hero entry len %d != %d: %w", len(pack), PlayerPackSize,
			ErrShortRecord)
	}

	hero := HeroRecord{Pack: pack}
	hero.Spawn = s.Mode.Spawn
	hero.Hellfire = s.Mode.Hellfire
	hero.Multiplayer = s.Mode.Multiplayer

	return hero, entries[GameEntryName], nil
}

// ReadStash returns the stash stream of the mode's stash entry.
func (s LegacySVStore) ReadStash(itemSize int) (StashSave, error) {
	name := StashEntryName(s.Mode)

	entries, err := s.readEntries([]string{name})
	if err != nil {
		return StashSave{}, err
	}

	return DecodeStashSave(entries[name], itemSize)
}

// writeEntries packs the given entries into the archive, skipping nil values.
// mpq.AddFile reads from disk, so the payloads go through a scratch directory.
func (s LegacySVStore) writeEntries(entries map[string][]byte) error {
	dir := filepath.Dir(s.Path)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("persist: save dir: %w", err)
	}

	scratch, err := os.MkdirTemp(dir, ".persist-*")
	if err != nil {
		return fmt.Errorf("persist: scratch dir: %w", err)
	}

	defer func() { _ = os.RemoveAll(scratch) }()

	archive, err := mpq.Create(s.Path, saveArchiveEntries)
	if err != nil {
		return fmt.Errorf("persist: create %s: %w", s.Path, err)
	}

	if err := addEntries(archive, entries, scratch); err != nil {
		_ = archive.Close()

		return err
	}

	if err := archive.Close(); err != nil {
		return fmt.Errorf("persist: close %s: %w", s.Path, err)
	}

	return nil
}

// readEntries extracts the named entries into memory.
func (s LegacySVStore) readEntries(names []string) (map[string][]byte, error) {
	archive, err := mpq.Open(s.Path)
	if err != nil {
		return nil, fmt.Errorf("persist: open %s: %w", s.Path, err)
	}

	defer func() { _ = archive.Close() }()

	scratch, err := os.MkdirTemp(filepath.Dir(s.Path), ".persist-*")
	if err != nil {
		return nil, fmt.Errorf("persist: scratch dir: %w", err)
	}

	defer func() { _ = os.RemoveAll(scratch) }()

	entries := make(map[string][]byte, len(names))
	for _, name := range names {
		target := filepath.Join(scratch, name)

		if err := archive.ExtractFile(name, target); err != nil {
			return nil, fmt.Errorf("persist: extract %q: %w", name, err)
		}

		data, err := os.ReadFile(target) //nolint:gosec // G304: scratch path built from a constant name
		if err != nil {
			return nil, fmt.Errorf("persist: read %q: %w", name, err)
		}

		entries[name] = data
	}

	return entries, nil
}

// addEntries writes each non-nil entry through a scratch file.
func addEntries(archive *mpq.Archive, entries map[string][]byte, scratch string) error {
	for name, data := range entries {
		if data == nil {
			continue
		}

		path := filepath.Join(scratch, name)
		if err := os.WriteFile(path, data, 0o600); err != nil {
			return fmt.Errorf("persist: write %q: %w", name, err)
		}

		if err := archive.AddFile(path, name); err != nil {
			return fmt.Errorf("persist: add %q: %w", name, err)
		}
	}

	return nil
}
