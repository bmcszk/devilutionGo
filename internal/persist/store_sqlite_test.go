package persist_test

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

func TestSQLiteStoreSchemaVersion(t *testing.T) {
	// given a fresh store
	ctx := context.Background()
	store := openStore(ctx, t)
	// when the schema version is read
	version, err := store.SchemaVersion(ctx)
	// then it matches SchemaVersion
	if err != nil {
		t.Fatalf("SchemaVersion: %v", err)
	}

	if version != persist.SchemaVersion {
		t.Fatalf("user_version = %d, want %d", version, persist.SchemaVersion)
	}
}

func TestSQLiteStoreHeroRoundTrip(t *testing.T) {
	// given a hero record carrying the PlayerPack image
	ctx := context.Background()
	store := openStore(ctx, t)
	hero := persist.HeroRecord{
		Slot:        3,
		Name:        "Conan",
		ClassID:     1,
		Pack:        goldenPlayerPack().Encode(),
		Hellfire:    true,
		Multiplayer: false,
	}
	// when it is saved and loaded back
	if err := store.SaveHero(ctx, hero); err != nil {
		t.Fatalf("SaveHero: %v", err)
	}

	loaded, err := store.LoadHero(ctx, 3)
	// then the pack bytes and mode flags come back unchanged
	if err != nil {
		t.Fatalf("LoadHero: %v", err)
	}

	if !bytes.Equal(loaded.Pack, hero.Pack) {
		t.Fatal("hero pack bytes changed")
	}

	if loaded.Name != hero.Name || loaded.ClassID != hero.ClassID || !loaded.Hellfire ||
		loaded.Multiplayer || loaded.CreatedAt == 0 || loaded.UpdatedAt == 0 {
		t.Fatalf("loaded = %+v", loaded)
	}
}

func TestSQLiteStoreSaveHeroUpdates(t *testing.T) {
	// given a saved hero
	ctx := context.Background()
	store := openStore(ctx, t)
	pack := goldenPlayerPack().Encode()

	if err := store.SaveHero(ctx, persist.HeroRecord{Slot: 0, Name: "Old", Pack: pack}); err != nil {
		t.Fatalf("SaveHero: %v", err)
	}
	// when the same slot is saved again
	if err := store.SaveHero(ctx, persist.HeroRecord{Slot: 0, Name: "New", ClassID: 2, Pack: pack}); err != nil {
		t.Fatalf("SaveHero: %v", err)
	}

	heroes, err := store.ListHeroes(ctx)
	// then the upsert replaced the row instead of adding one
	if err != nil {
		t.Fatalf("ListHeroes: %v", err)
	}

	if len(heroes) != 1 || heroes[0].Name != "New" || heroes[0].ClassID != 2 {
		t.Fatalf("heroes = %+v", heroes)
	}
}

func TestSQLiteStoreListHeroesOrder(t *testing.T) {
	// given three heroes saved out of slot order
	ctx := context.Background()
	store := openStore(ctx, t)
	pack := goldenPlayerPack().Encode()

	for _, slot := range []int{7, 1, 4} {
		if err := store.SaveHero(ctx, persist.HeroRecord{Slot: slot, Name: "Hero", Pack: pack}); err != nil {
			t.Fatalf("SaveHero(%d): %v", slot, err)
		}
	}
	// when the list is read
	heroes, err := store.ListHeroes(ctx)
	// then it is ordered by slot, like pfile's hero_names walk
	if err != nil {
		t.Fatalf("ListHeroes: %v", err)
	}

	for i, want := range []int{1, 4, 7} {
		if heroes[i].Slot != want {
			t.Fatalf("heroes[%d].Slot = %d, want %d", i, heroes[i].Slot, want)
		}
	}
}

func TestSQLiteStoreDeleteHero(t *testing.T) {
	// given a hero with a game and a stash row
	ctx := context.Background()
	store := openStore(ctx, t)
	pack := goldenPlayerPack().Encode()

	if err := store.SaveHero(ctx, persist.HeroRecord{Slot: 2, Name: "Aidan", Pack: pack}); err != nil {
		t.Fatalf("SaveHero: %v", err)
	}

	game := make([]byte, persist.GameHeaderSize)
	if err := store.SaveGame(ctx, 2, "single", game); err != nil {
		t.Fatalf("SaveGame: %v", err)
	}

	if err := store.SaveStash(ctx, persist.StashRecord{Slot: 2, PageCount: 1}); err != nil {
		t.Fatalf("SaveStash: %v", err)
	}
	// when the hero is deleted
	if err := store.DeleteHero(ctx, 2); err != nil {
		t.Fatalf("DeleteHero: %v", err)
	}
	// then the hero, its game and its stash are all gone
	if _, err := store.LoadHero(ctx, 2); !errors.Is(err, persist.ErrHeroNotFound) {
		t.Fatalf("LoadHero after delete = %v, want ErrHeroNotFound", err)
	}

	if _, err := store.LoadGame(ctx, 2, "single"); !errors.Is(err, persist.ErrHeroNotFound) {
		t.Fatalf("LoadGame after delete = %v, want ErrHeroNotFound", err)
	}

	if _, err := store.LoadStash(ctx, 2); !errors.Is(err, persist.ErrHeroNotFound) {
		t.Fatalf("LoadStash after delete = %v, want ErrHeroNotFound", err)
	}
}

func TestSQLiteStoreGameRoundTrip(t *testing.T) {
	// given a SaveGameData stream for a hero
	ctx := context.Background()
	store := openStore(ctx, t)
	pack := goldenPlayerPack().Encode()

	if err := store.SaveHero(ctx, persist.HeroRecord{Slot: 1, Name: "Rogue", Pack: pack}); err != nil {
		t.Fatalf("SaveHero: %v", err)
	}

	header := persist.GameHeader{CurrLevel: 4, LevelType: 5, ActiveMonsters: 3}

	data, err := header.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	// when the stream is saved and loaded back
	if err := store.SaveGame(ctx, 1, "single", data); err != nil {
		t.Fatalf("SaveGame: %v", err)
	}

	loaded, err := store.LoadGame(ctx, 1, "single")
	// then it round-trips byte-identically
	if err != nil {
		t.Fatalf("LoadGame: %v", err)
	}

	if !bytes.Equal(loaded, data) {
		t.Fatal("game stream changed")
	}
}

func TestSQLiteStoreSaveGameRejectsShortStream(t *testing.T) {
	// when a stream shorter than the game header is stored
	ctx := context.Background()
	store := openStore(ctx, t)
	// then it is rejected before reaching SQLite
	err := store.SaveGame(ctx, 1, "single", make([]byte, persist.GameHeaderSize-1))
	if !errors.Is(err, persist.ErrShortRecord) {
		t.Fatalf("err = %v, want ErrShortRecord", err)
	}
}

func TestSQLiteStoreStashRoundTrip(t *testing.T) {
	// given a stash with two pages of item records
	ctx := context.Background()
	store := openStore(ctx, t)
	stash := persist.StashRecord{Slot: -1, PageCount: 2, Blob: []byte{1, 2, 3, 4}}
	// when it is saved and loaded back
	if err := store.SaveStash(ctx, stash); err != nil {
		t.Fatalf("SaveStash: %v", err)
	}

	loaded, err := store.LoadStash(ctx, -1)
	// then the row round-trips
	if err != nil {
		t.Fatalf("LoadStash: %v", err)
	}

	if loaded.PageCount != 2 || !bytes.Equal(loaded.Blob, stash.Blob) {
		t.Fatalf("loaded = %+v", loaded)
	}
}

func TestSQLiteStoreSaveHeroValidation(t *testing.T) {
	// given invalid hero records
	ctx := context.Background()
	store := openStore(ctx, t)

	cases := []struct {
		name string
		hero persist.HeroRecord
		want error
	}{
		{
			"slot too high",
			persist.HeroRecord{Slot: persist.MaxCharacters, Pack: make([]byte, persist.PlayerPackSize)},
			persist.ErrSlotOutOfRange,
		},
		{"short pack", persist.HeroRecord{Slot: 0, Pack: make([]byte, 4)}, persist.ErrShortRecord},
	}
	// when they are saved
	for _, tc := range cases {
		err := store.SaveHero(ctx, tc.hero)
		// then the store refuses each of them
		if !errors.Is(err, tc.want) {
			t.Fatalf("%s: err = %v, want %v", tc.name, err, tc.want)
		}
	}
}

func TestSQLiteStoreReopenKeepsSchema(t *testing.T) {
	// given a store written once in a database file
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "hero.db")
	pack := goldenPlayerPack().Encode()

	store, err := persist.OpenSQLiteStore(ctx, path)
	if err != nil {
		t.Fatalf("OpenSQLiteStore: %v", err)
	}

	if err := store.SaveHero(ctx, persist.HeroRecord{Slot: 4, Name: "Rogue", Pack: pack}); err != nil {
		t.Fatalf("SaveHero: %v", err)
	}

	// when the file is reopened, which reruns the migration chain
	reopened, err := persist.OpenSQLiteStore(ctx, path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	// then the version stays SchemaVersion and the row survives the migration
	assertReopenedStore(ctx, t, reopened, pack)
}

func openStore(ctx context.Context, t *testing.T) *persist.SQLiteStore {
	t.Helper()

	store, err := persist.OpenSQLiteStore(ctx, filepath.Join(t.TempDir(), "hero.db"))
	if err != nil {
		t.Fatalf("OpenSQLiteStore: %v", err)
	}

	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Fatalf("Close: %v", err)
		}
	})

	return store
}

func assertReopenedStore(ctx context.Context, t *testing.T, store *persist.SQLiteStore, pack []byte) {
	t.Helper()

	defer func() { _ = store.Close() }()

	version, err := store.SchemaVersion(ctx)
	if err != nil {
		t.Fatalf("SchemaVersion: %v", err)
	}

	hero, err := store.LoadHero(ctx, 4)
	if err != nil {
		t.Fatalf("LoadHero after reopen: %v", err)
	}

	if version != persist.SchemaVersion {
		t.Fatalf("user_version = %d, want %d", version, persist.SchemaVersion)
	}

	if hero.Name != "Rogue" || !bytes.Equal(hero.Pack, pack) {
		t.Fatalf("hero after reopen = %+v", hero)
	}
}
