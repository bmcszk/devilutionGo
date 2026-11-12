package persist_test

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	gametestsupport "github.com/bmcszk/devilutionGo/internal/game/testsupport"
	"github.com/bmcszk/devilutionGo/internal/persist"
)

// sessionNamespace is the save family the session test runs under: a spawned
// single-player hero (pfile.cpp:78-97).
var sessionNamespace = persist.SaveNamespace{Spawn: true}

// sessionState adapts the persisted spawn-session blocks to the shared
// canonical StateHash contract (internal/game/testsupport StateHash, FNV-1a
// over named field bytes): the game header, the hero pack, the town level
// block and the dropped items.
type sessionState struct {
	header persist.GameHeader
	hero   persist.PlayerPack
	level  sessionLevel
	items  persist.DroppedItems
}

// sessionLevel is the level slice of the canonical session state.
type sessionLevel struct {
	header persist.LevelHeader
	grids  persist.TileGrids
}

// persistedSession carries the raw blob form of one saved session.
type persistedSession struct {
	game  []byte
	pack  []byte
	level []byte
	items []byte
}

// CanonicalFields returns the session state in save order: game header, hero
// pack, level header, tile grids, dropped items (loadsave.cpp:2763-2800,
// pack.cpp:152, :1954-2008, :1838-1852).
func (s sessionState) CanonicalFields() []gametestsupport.HashField {
	return []gametestsupport.HashField{
		{Name: "gameheader", Bytes: encodeHeader(s.header)},
		{Name: "heropack", Bytes: s.hero.Encode()},
		{Name: "levelheader", Bytes: s.level.header.Encode()},
		{Name: "tilegrids", Bytes: s.level.grids.Encode()},
		{Name: "droppeditems", Bytes: encodeItems(s.items)},
	}
}

// TestCoreSessionSaveLoadRoundTrip pins the c6wt core-session AC on the
// persist surface that exists today: a minimal spawned state (warrior, town,
// a walked town path, one item picked up) survives save -> load with an
// identical StateHash, and a second save after load is byte-identical to the
// first (roundtrip idempotence, loadsave.cpp SaveGameData/SaveLevel/
// SaveDroppedItems + pack.cpp PackPlayer/UnpackPlayer).
func TestCoreSessionSaveLoadRoundTrip(t *testing.T) {
	// given a spawned warrior who walked to the west and picked up an item
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "hero.db")

	sent := spawnSessionState()
	heroEntry := sealedHeroEntry(t, sent)

	store := saveSession(ctx, t, path, heroEntry, sent)

	// when the store is reopened and every block is loaded back
	store = reopenSessionStore(ctx, t, store, path)

	loaded := loadSessionState(ctx, t, store, heroEntry)

	// then the reloaded state hashes identically
	if gotHash, wantHash := gametestsupport.StateHash(loaded), gametestsupport.StateHash(sent); gotHash != wantHash {
		t.Fatalf("StateHash after round-trip = %x, want %x", gotHash, wantHash)
	}

	// and the second save is byte-identical to the first (idempotence)
	expectResaveMatches(t, heroEntry, sent, resaveSession(ctx, t, store, heroEntry, loaded))
}

// spawnSessionState builds the minimal spawned session: a warrior on town,
// a walked path toward the west edge (the town walkthrough path), and one
// picked-up item in the belt.
func spawnSessionState() sessionState {
	state := sessionState{}
	state.header = persist.GameHeader{
		Hellfire: false, Spawn: true,
		CurrLevel: 0, LevelType: 0,
		ViewX: 75, ViewY: 68,
		ActiveMonsters: 0, ActiveItems: 1, ActiveMissiles: 0, ActiveObjects: 0,
	}

	state.hero = sessionHero()

	// the town level: quiet grids, one dLight row marked walked
	state.level.header = persist.LevelHeader{ActiveMonsters: 0, ActiveItems: 1, ActiveObjects: 0}
	for i := range state.level.grids.Lights {
		state.level.grids.Lights[i] = 15 // unlit town (dLight, light_tables.cpp:19-70)
	}

	for i := range state.level.grids.Automap {
		state.level.grids.Automap[i] = uint8(i % 7) // automap learned along the path
	}

	state.items = sessionDroppedItem()

	return state
}

// sessionHero builds the walked warrior: level 1, tile (79, 68), one belt
// item, one picked-up inventory item.
func sessionHero() persist.PlayerPack {
	player := persist.PlayerPack{Level: 1}
	player.Class = 0 // warrior (pack.cpp HeroClass table order)
	copy(player.Name[:], "Conan")
	player.Px, player.Py = 79, 68
	player.BaseStr, player.BaseMag, player.BaseDex, player.BaseVit = 30, 10, 20, 25
	player.HPBase, player.MaxHP = 96, 128
	player.ManaBase, player.MaxMana = 8, 8
	player.Exp, player.Gold = 100, 500

	player.SpdList[0] = persist.ItemPack{Idx: 5} // a real itemdat row, not the empty slot
	player.NumInv = 1
	player.InvList[0] = persist.ItemPack{Idx: 42, Seed: 1234}
	player.InvGrid[0] = 1

	return player
}

// sessionDroppedItem builds the dropped world item the session picked up.
func sessionDroppedItem() persist.DroppedItems {
	itemRecord := make([]byte, persist.ItemSaveSizeDiablo)
	itemRecord[3] = 0x2A // seed bytes of the dropped item

	return persist.DroppedItems{
		ActiveCount: 1,
		Records:     [][]byte{itemRecord},
	}
}

// sealedHeroEntry codec-seals the hero pack (the pass every hero file takes,
// pfile.cpp:177) and returns the store record.
func sealedHeroEntry(t *testing.T, sent sessionState) persist.HeroRecord {
	t.Helper()

	pack := sent.hero.Encode()
	if len(pack) != persist.PlayerPackSize {
		t.Fatalf("pack is %d bytes, want %d", len(pack), persist.PlayerPackSize)
	}

	// the codec pass every hero file takes (pfile.cpp:177 codec_encode)
	password := persist.HeroPassword(sessionNamespace)
	encoded := encodeHeroEntry(t, pack, password)

	decoded, err := persist.DecodeHeroEntry(encoded, password)
	if err != nil {
		t.Fatalf("DecodeHeroEntry: %v", err)
	}

	if !bytes.Equal(decoded, pack) {
		t.Fatal("codec round trip changed the hero pack")
	}

	return persist.HeroRecord{Slot: 0, Name: "Conan", ClassID: 0, Pack: pack, Spawn: true}
}

// saveSession writes every session block into a fresh store.
func saveSession(
	ctx context.Context, t *testing.T, path string, heroEntry persist.HeroRecord, sent sessionState,
) *persist.SQLiteStore {
	t.Helper()

	store := openSessionStore(ctx, t, path)

	if err := store.SaveHero(ctx, heroEntry); err != nil {
		t.Fatalf("SaveHero: %v", err)
	}

	if err := store.SaveGame(ctx, heroEntry.Slot, "single", encodeHeader(sent.header)); err != nil {
		t.Fatalf("SaveGame: %v", err)
	}

	if err := store.SaveLevel(ctx, heroEntry.Slot, 0, encodeLevel(sent.level)); err != nil {
		t.Fatalf("SaveLevel: %v", err)
	}

	if err := store.SaveGame(ctx, heroEntry.Slot, "perm_0", encodeItems(sent.items)); err != nil {
		t.Fatalf("SaveGame perm level: %v", err)
	}

	closeStore(ctx, t, store)

	return store
}

// reopenSessionStore closes the open store and reopens the same file.
func reopenSessionStore(
	ctx context.Context,
	t *testing.T,
	store *persist.SQLiteStore,
	path string,
) *persist.SQLiteStore {
	t.Helper()

	closeStore(ctx, t, store)

	return openSessionStore(ctx, t, path)
}

// openSessionStore opens a save database.
func openSessionStore(ctx context.Context, t *testing.T, path string) *persist.SQLiteStore {
	t.Helper()

	store, err := persist.OpenSQLiteStore(ctx, path)
	if err != nil {
		t.Fatalf("OpenSQLiteStore: %v", err)
	}

	return store
}

// closeStore closes a save database.
func closeStore(_ context.Context, t *testing.T, store *persist.SQLiteStore) {
	t.Helper()

	if err := store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

// loadSessionState reloads every session block from the store.
func loadSessionState(
	ctx context.Context,
	t *testing.T,
	store *persist.SQLiteStore,
	entry persist.HeroRecord,
) sessionState {
	t.Helper()

	var loaded sessionState

	hero, err := store.LoadHero(ctx, entry.Slot)
	if err != nil {
		t.Fatalf("LoadHero: %v", err)
	}

	loaded.hero = decodePlayerPack(t, hero.Pack)
	loaded.header = decodeSessionHeader(t, loadGameBlob(ctx, t, store, entry.Slot, "single"))
	loaded.level = decodeSessionLevel(t, loadLevelBlob(ctx, t, store, entry.Slot, 0))
	loaded.items = decodeSessionItems(t, loadGameBlob(ctx, t, store, entry.Slot, "perm_0"))

	return loaded
}

// loadGameBlob loads one named game blob.
func loadGameBlob(
	ctx context.Context,
	t *testing.T,
	store *persist.SQLiteStore,
	slot int,
	kind string,
) []byte {
	t.Helper()

	blob, err := store.LoadGame(ctx, slot, kind)
	if err != nil {
		t.Fatalf("LoadGame %s: %v", kind, err)
	}

	return blob
}

// loadLevelBlob loads one level block.
func loadLevelBlob(
	ctx context.Context,
	t *testing.T,
	store *persist.SQLiteStore,
	slot, level int,
) []byte {
	t.Helper()

	blob, err := store.LoadLevel(ctx, slot, level)
	if err != nil {
		t.Fatalf("LoadLevel: %v", err)
	}

	return blob
}

// decodePlayerPack decodes a hero-file record image.
func decodePlayerPack(t *testing.T, pack []byte) persist.PlayerPack {
	t.Helper()

	packed, err := persist.DecodePlayerPack(pack)
	if err != nil {
		t.Fatalf("DecodePlayerPack: %v", err)
	}

	return packed
}

// decodeSessionHeader decodes a game-header blob.
func decodeSessionHeader(t *testing.T, blob []byte) persist.GameHeader {
	t.Helper()

	header, _, err := persist.DecodeGameHeader(blob)
	if err != nil {
		t.Fatalf("DecodeGameHeader: %v", err)
	}

	return header
}

// decodeSessionLevel splits a level block into header and grids.
func decodeSessionLevel(t *testing.T, blob []byte) sessionLevel {
	t.Helper()

	header, err := persist.DecodeLevelHeader(blob)
	if err != nil {
		t.Fatalf("DecodeLevelHeader: %v", err)
	}

	grids, err := persist.DecodeTileGrids(blob[persist.LevelHeaderSize:])
	if err != nil {
		t.Fatalf("DecodeTileGrids: %v", err)
	}

	return sessionLevel{header: header, grids: grids}
}

// decodeSessionItems decodes a dropped-items block.
func decodeSessionItems(t *testing.T, blob []byte) persist.DroppedItems {
	t.Helper()

	items, err := persist.DecodeDroppedItems(blob, persist.ItemSaveSizeDiablo)
	if err != nil {
		t.Fatalf("DecodeDroppedItems: %v", err)
	}

	return items
}

// resaveSession re-encodes every loaded block: the byte-identity half of the
// roundtrip (a load/save cycle must not drift).
func resaveSession(
	ctx context.Context,
	t *testing.T,
	store *persist.SQLiteStore,
	entry persist.HeroRecord,
	loaded sessionState,
) persistedSession {
	t.Helper()

	var resaved persistedSession

	resaved.pack = loaded.hero.Encode()

	resaveHero(ctx, t, store, entry, loaded.hero, resaved.pack)
	resaved.game = encodeHeader(loaded.header)
	resaved.level = encodeLevel(loaded.level)
	resaved.items = encodeItems(loaded.items)

	return resaved
}

// resaveHero writes the re-encoded hero record back.
func resaveHero(
	ctx context.Context, t *testing.T, store *persist.SQLiteStore,
	entry persist.HeroRecord, hero persist.PlayerPack, pack []byte,
) {
	t.Helper()

	resaved := persist.HeroRecord{
		Slot: entry.Slot, Name: heroName(hero), ClassID: int(hero.Class),
		Pack: pack, Spawn: entry.Spawn,
	}

	if err := store.SaveHero(ctx, resaved); err != nil {
		t.Fatalf("SaveHero (resave): %v", err)
	}
}

// expectResaveMatches asserts the second save is byte-identical to the first.
func expectResaveMatches(
	t *testing.T,
	heroEntry persist.HeroRecord,
	sent sessionState,
	resaved persistedSession,
) {
	t.Helper()

	if !bytes.Equal(resaved.game, encodeHeader(sent.header)) {
		t.Fatal("re-saved game header diverges")
	}

	if !bytes.Equal(resaved.pack, heroEntry.Pack) {
		t.Fatal("re-saved hero pack diverges")
	}

	if !bytes.Equal(resaved.level, encodeLevel(sent.level)) {
		t.Fatal("re-saved level block diverges")
	}

	if !bytes.Equal(resaved.items, encodeItems(sent.items)) {
		t.Fatal("re-saved dropped items diverge")
	}
}

// encodeHeader encodes a game header or fails the test.
func encodeHeader(header persist.GameHeader) []byte {
	body, err := header.Encode()
	if err != nil {
		panic(err)
	}

	return body
}

// encodeItems encodes the dropped-items block or fails the test.
func encodeItems(items persist.DroppedItems) []byte {
	body, err := items.Encode(persist.ItemSaveSizeDiablo)
	if err != nil {
		panic(err)
	}

	return body
}

// encodeLevel encodes a level block or fails the test.
func encodeLevel(level sessionLevel) []byte {
	return append(level.header.Encode(), level.grids.Encode()...)
}

// heroName reads the packed name field.
func heroName(pack persist.PlayerPack) string {
	return string(bytes.TrimRight(pack.Name[:], "\x00"))
}
