package persist_test

import (
	"bytes"
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

func legacyStore(t *testing.T) persist.LegacySVStore {
	t.Helper()

	return persist.LegacySVStore{Path: filepath.Join(t.TempDir(), "single_0.sv")}
}
