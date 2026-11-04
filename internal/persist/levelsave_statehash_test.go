package persist_test

import (
	"context"
	"path/filepath"
	"testing"

	gametestsupport "github.com/bmcszk/devilutionGo/internal/game/testsupport"
	"github.com/bmcszk/devilutionGo/internal/persist"
)

// levelState adapts a level block to the shared canonical StateHash contract
// (internal/game/testsupport StateHash, FNV-1a over named field bytes).
type levelState struct {
	header persist.LevelHeader
	grids  persist.TileGrids
}

// CanonicalFields returns the level block in save order: the counter header
// then the tile grids (loadsave.cpp:1954-2008).
func (s levelState) CanonicalFields() []gametestsupport.HashField {
	return []gametestsupport.HashField{
		{Name: "header", Bytes: s.header.Encode()},
		{Name: "grids", Bytes: s.grids.Encode()},
	}
}

// TestLoadLevelStateHashRoundTrip pins the oqho AC that was only covered on
// the pack/savegame side: a level saved to SQLite and loaded back must hash to
// the identical StateHash.
func TestLoadLevelStateHashRoundTrip(t *testing.T) {
	// given a scripted level state written to a hero slot
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "hero.db")
	want := scriptedLevelState()

	store, err := persist.OpenSQLiteStore(ctx, path)
	if err != nil {
		t.Fatalf("OpenSQLiteStore: %v", err)
	}

	block := append(want.header.Encode(), want.grids.Encode()...)
	if err := store.SaveLevel(ctx, 4, 7, block); err != nil {
		t.Fatalf("SaveLevel: %v", err)
	}

	if err := store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// when the store is reopened and the level is loaded back
	store, err = persist.OpenSQLiteStore(ctx, path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}

	defer func() { _ = store.Close() }()

	loaded, err := store.LoadLevel(ctx, 4, 7)
	if err != nil {
		t.Fatalf("LoadLevel: %v", err)
	}

	got := decodeLevelState(t, loaded)

	// then the reloaded state hashes identically
	if gotHash, wantHash := gametestsupport.StateHash(got), gametestsupport.StateHash(want); gotHash != wantHash {
		t.Fatalf("StateHash after round-trip = %x, want %x", gotHash, wantHash)
	}
}

// scriptedLevelState builds a non-trivial level block: distinct counts, signed
// monster indices and every grid filled with position-dependent values.
func scriptedLevelState() levelState {
	var state levelState

	state.header = persist.LevelHeader{ActiveMonsters: 41, ActiveItems: 17, ActiveObjects: 293}

	for i := range state.grids.Monsters {
		state.grids.Monsters[i] = int32(i%7 - 3)
		state.grids.Objects[i] = int8(i % 11)
		state.grids.Lights[i] = uint8(i)
		state.grids.PreLights[i] = uint8(255 - i%255)
	}

	for i := range state.grids.Automap {
		state.grids.Automap[i] = uint8(i % 3)
	}

	return state
}

// decodeLevelState parses a loaded level block into header plus grids.
func decodeLevelState(t *testing.T, block []byte) levelState {
	t.Helper()

	header, err := persist.DecodeLevelHeader(block)
	if err != nil {
		t.Fatalf("DecodeLevelHeader: %v", err)
	}

	grids, err := persist.DecodeTileGrids(block[persist.LevelHeaderSize:])
	if err != nil {
		t.Fatalf("DecodeTileGrids: %v", err)
	}

	return levelState{header: header, grids: grids}
}
