package persist

import (
	"context"
	"errors"
	"fmt"
	"os"
)

// HeroStoreSet is the per-hero save store set: one SQLite file per hero slot,
// mirroring upstream's one-archive-per-saveNum layout (GetSavePath,
// Source/pfile.cpp:80-92 @ca91d778f; OpenSaveArchive opens the archive per
// call and closes it when it leaves scope, pfile.cpp:609-612). The set holds
// only the directory and the save namespace — no open handle lives here.
type HeroStoreSet struct {
	dir string
	ns  SaveNamespace
}

// NewHeroStoreSet binds the per-hero store set to dir and the save namespace,
// with the extension forced to "db" — the live store layout the migrate CLI
// already writes (cmd/devilutionGo-migrate/saves.go heroDBPath, saves.go:188-192);
// the .sv extension stays upstream's legacy format (GetSavePath's ext token,
// pfile.cpp:74-78 @ca91d778f). No file is opened at construction; every
// operation opens, acts and closes.
func NewHeroStoreSet(dir string, ns SaveNamespace) *HeroStoreSet {
	ns.Extension = "db"

	return &HeroStoreSet{dir: dir, ns: ns}
}

// Path is GetSavePath(saveNum) with the live store extension: the slot's own
// database file (pfile.cpp:80-92 @ca91d778f; persist.SavePath carries the
// 0..MaxCharacters range check, pfile.go:59-72).
func (s *HeroStoreSet) Path(slot int) (string, error) {
	return s.ns.SavePath(s.dir, slot, "")
}

// SaveHero upserts the hero row into the slot's own file
// (pfile_ui_save_create writes the slot's archive, pfile.cpp @ca91d778f).
func (s *HeroStoreSet) SaveHero(ctx context.Context, hero HeroRecord) error {
	slot := hero.Slot

	return s.with(ctx, slot, func(store *SQLiteStore) error {
		if err := store.SaveHero(ctx, hero); err != nil {
			return fmt.Errorf("persist: save hero slot %d: %w", slot, err)
		}

		return nil
	})
}

// LoadHero reads the hero row from the slot's own file
// (pfile_read_player_from_save opens the archive by saveNum, pfile.cpp:786-798
// @ca91d778f). A missing file is the ErrHeroNotFound sentinel.
func (s *HeroStoreSet) LoadHero(ctx context.Context, slot int) (HeroRecord, error) {
	return s.loadSlotFile(ctx, slot)
}

// ListHeroes probes every slot 0..MaxCharacters-1 like
// pfile_ui_set_hero_infos does (pfile.cpp:695-726 @ca91d778f): a missing file
// is an empty slot; a file that fails to open or read is an ERROR — a broken
// hero must not look like an empty slot (deviation 1; upstream skips a failed
// ReadHero silently).
func (s *HeroStoreSet) ListHeroes(ctx context.Context) ([]HeroRecord, error) {
	heroes := make([]HeroRecord, 0, MaxCharacters)

	for slot := range MaxCharacters {
		if !s.HasSlot(slot) {
			continue // empty slot: upstream's failed open (pfile.cpp:700-704)
		}

		hero, err := s.loadSlotFile(ctx, slot)
		if err != nil {
			return nil, fmt.Errorf("persist: list heroes slot %d: %w", slot, err)
		}

		heroes = append(heroes, hero)
	}

	return heroes, nil
}

// DeleteHero removes the slot's file — the whole hero IS the file
// (pfile_delete_save calls RemoveFile(GetSavePath(saveNum)), pfile.cpp:775-783
// @ca91d778f). A missing file is the ErrHeroNotFound sentinel (deviation 4).
func (s *HeroStoreSet) DeleteHero(_ context.Context, slot int) error {
	path, err := s.Path(slot)
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("persist: delete hero slot %d: %w", slot, ErrHeroNotFound)
		}

		return fmt.Errorf("persist: delete hero slot %d: %w", slot, err)
	}

	return nil
}

// HasSlot reports whether the slot's own file exists on disk.
func (s *HeroStoreSet) HasSlot(slot int) bool {
	path, err := s.Path(slot)
	if err != nil {
		return false
	}

	_, err = os.Stat(path)

	return err == nil
}

// with opens the slot's own store, runs fn against it and closes the handle:
// upstream's open-per-operation shape (pfile.cpp:609-612 @ca91d778f) — no
// handle ever outlives the call.
func (s *HeroStoreSet) with(ctx context.Context, slot int, fn func(*SQLiteStore) error) error {
	path, err := s.Path(slot)
	if err != nil {
		return err
	}

	store, err := OpenSQLiteStore(ctx, path)
	if err != nil {
		return fmt.Errorf("persist: open hero slot %d: %w", slot, err)
	}

	defer func() {
		if cerr := store.Close(); cerr != nil {
			err = fmt.Errorf("persist: close hero slot %d: %w", slot, cerr)
		}
	}()

	return fn(store)
}

// loadSlotFile reads one existing slot file's hero row through its own open.
func (s *HeroStoreSet) loadSlotFile(ctx context.Context, slot int) (HeroRecord, error) {
	var hero HeroRecord

	load := func(store *SQLiteStore) error {
		row, err := store.LoadHero(ctx, slot)
		if err != nil {
			return err
		}

		hero = row

		return nil
	}

	err := s.with(ctx, slot, load)
	if err != nil {
		return HeroRecord{}, err
	}

	return hero, nil
}
