package persist_test

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

func TestLegacySVWriteReadRoundTrip(t *testing.T) {
	// given a hero record and a game stream
	store := legacyStore(t)
	pack := goldenPlayerPack().Encode()
	game, err := persist.GameHeader{CurrLevel: 2, LevelType: 1}.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	// when the archive is written and read back
	if err := store.Write(persist.HeroRecord{Slot: 0, Name: "Conan", Pack: pack}, game); err != nil {
		t.Fatalf("Write: %v", err)
	}

	hero, gameBack, err := store.Read()
	// then the codec-decoded hero pack and the game stream round-trip
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if !bytes.Equal(hero.Pack, pack) {
		t.Fatal("hero pack changed")
	}

	if !bytes.Equal(gameBack, game) {
		t.Fatal("game stream changed")
	}
}

func TestLegacySVHeroEntryIsEncoded(t *testing.T) {
	// given an archive written with the single-player password
	store := legacyStore(t)
	pack := goldenPlayerPack().Encode()

	if err := store.Write(persist.HeroRecord{Slot: 0, Pack: pack}, nil); err != nil {
		t.Fatalf("Write: %v", err)
	}
	// when it is read with a store configured for multiplayer
	multi := persist.LegacySVStore{Path: store.Path, Mode: persist.GameMode{Multiplayer: true}}
	// then the wrong password fails to decode the hero entry
	if _, _, err := multi.Read(); err == nil {
		t.Fatal("read with the wrong password succeeded")
	}
}

func TestLegacySVStashRoundTrip(t *testing.T) {
	// given a stash stream
	store := legacyStore(t)
	save := stashFixture()
	// when it is written and read back through the archive
	if err := store.WriteStash(save, persist.ItemSaveSizeDiablo); err != nil {
		t.Fatalf("WriteStash: %v", err)
	}

	back, err := store.ReadStash(persist.ItemSaveSizeDiablo)
	// then the stream round-trips
	if err != nil {
		t.Fatalf("ReadStash: %v", err)
	}

	assertStashEqual(t, back, save)
}

func TestLegacySVWriteRejectsBadHero(t *testing.T) {
	// when a hero with a short pack is written
	store := legacyStore(t)
	err := store.Write(persist.HeroRecord{Slot: 0, Pack: make([]byte, 5)}, nil)
	// then the write is refused before the archive is created
	if err == nil {
		t.Fatal("write of a short pack succeeded")
	}
}

func TestLegacySVReadMissingArchive(t *testing.T) {
	// given a store pointing at a file that does not exist
	store := legacyStore(t)
	// when it is read
	_, _, err := store.Read()
	// then the missing archive is reported
	if err == nil {
		t.Fatal("read of a missing archive succeeded")
	}
}

func TestLegacySVWriteSaveExtraEntries(t *testing.T) {
	// given a save with levelseeds and hotkeys entries alongside hero+game
	store := legacyStore(t)
	pack := goldenPlayerPack().Encode()
	seeds := persist.LevelSeeds{{Present: true, Seed: 99}}.Encode(persist.GameMode{})
	hotkeys := persist.HeroHotkeys{Hotkeys: []persist.SpellHotkey{{SpellID: 4, SpellType: 1}}}.Encode()
	game, err := persist.GameHeader{CurrLevel: 1}.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	// when the archive is written and the entries read back
	err = store.WriteSave(persist.HeroSave{
		Hero:  persist.HeroRecord{Slot: 0, Pack: pack},
		Game:  game,
		Extra: map[string][]byte{persist.LevelSeedsEntryName: seeds, persist.HotkeysEntryName: hotkeys},
	})
	if err != nil {
		t.Fatalf("WriteSave: %v", err)
	}

	// then every entry round-trips byte-identically
	assertSaveEntries(t, store, pack, game, map[string][]byte{
		persist.LevelSeedsEntryName: seeds,
		persist.HotkeysEntryName:    hotkeys,
	})
}

func TestImportExportSV(t *testing.T) {
	// given a legacy .sv archive written by an upstream-compatible build
	ctx := context.Background()
	archive := legacyStore(t)
	pack := goldenPlayerPack().Encode()
	game, err := persist.GameHeader{CurrLevel: 6, LevelType: 5}.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	if err := archive.Write(persist.HeroRecord{Slot: 0, Pack: pack}, game); err != nil {
		t.Fatalf("Write: %v", err)
	}
	// when it is imported into the SQLite store
	store := openStore(ctx, t)

	if err := persist.ImportSV(ctx, store, archive.Path, persist.GameMode{}, 2); err != nil {
		t.Fatalf("ImportSV: %v", err)
	}

	hero, err := store.LoadHero(ctx, 2)
	// then the hero lands in the requested slot with its game stream
	if err != nil {
		t.Fatalf("LoadHero: %v", err)
	}

	stored, err := store.LoadGame(ctx, 2, string(persist.KindFor(persist.GameMode{})))
	if err != nil {
		t.Fatalf("LoadGame: %v", err)
	}

	if !bytes.Equal(hero.Pack, pack) || !bytes.Equal(stored, game) {
		t.Fatal("imported payload changed")
	}
}

func TestExportSVRoundTrip(t *testing.T) {
	// given a hero in the SQLite store
	ctx := context.Background()
	store := openStore(ctx, t)
	pack := goldenPlayerPack().Encode()

	if err := store.SaveHero(ctx, persist.HeroRecord{Slot: 1, Name: "Aidan", Pack: pack}); err != nil {
		t.Fatalf("SaveHero: %v", err)
	}
	// when it is exported into a .sv archive
	out := persist.LegacySVStore{Path: filepath.Join(t.TempDir(), "single_1.sv")}
	if err := persist.ExportSV(ctx, store, out.Path, persist.GameMode{}, 1); err != nil {
		t.Fatalf("ExportSV: %v", err)
	}

	hero, _, err := out.Read()
	// then the archive decodes back to the stored pack
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if !bytes.Equal(hero.Pack, pack) {
		t.Fatal("exported pack changed")
	}
}

func legacyStore(t *testing.T) persist.LegacySVStore {
	t.Helper()

	return persist.LegacySVStore{Path: filepath.Join(t.TempDir(), "single_0.sv")}
}

func assertSaveEntries(t *testing.T, store persist.LegacySVStore, pack, game []byte, extra map[string][]byte) {
	t.Helper()

	for name, want := range extra {
		assertEntryEqual(t, store, name, want)
	}

	hero, gotGame, err := store.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if !bytes.Equal(gotGame, game) || !bytes.Equal(hero.Pack, pack) {
		t.Fatal("hero or game entry changed")
	}
}

func assertEntryEqual(t *testing.T, store persist.LegacySVStore, name string, want []byte) {
	t.Helper()

	got, err := store.ReadEntry(name)
	if err != nil {
		t.Fatalf("ReadEntry(%s): %v", name, err)
	}

	if !bytes.Equal(got, want) {
		t.Fatalf("entry %s changed", name)
	}
}
