package persist_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

func TestStashEntryName(t *testing.T) {
	// when the stash entry name is picked per mode (loadsave.cpp:2714-2720)
	single := persist.StashEntryName(persist.GameMode{})
	multi := persist.StashEntryName(persist.GameMode{Multiplayer: true})
	// then single player uses spstashitems and multiplayer mpstashitems
	if single != "spstashitems" || multi != "mpstashitems" {
		t.Fatalf("single = %q multi = %q", single, multi)
	}
}

func TestItemSizeFor(t *testing.T) {
	// when the SaveItem record length is picked per game (loadsave.cpp:2114-2115)
	diablo := persist.ItemSizeFor(persist.GameMode{})
	hellfire := persist.ItemSizeFor(persist.GameMode{Hellfire: true})
	// then Diablo is 368 bytes and Hellfire 372
	if diablo != 368 || hellfire != 372 {
		t.Fatalf("diablo = %d hellfire = %d", diablo, hellfire)
	}
}

func TestStashEncodeGoldenHead(t *testing.T) {
	// given a stash with one page, one item and gold
	save := stashFixture()
	save.Items = save.Items[:1]
	// when
	encoded, err := save.Encode(persist.ItemSaveSizeDiablo)
	// then the head is version 0, gold LE, page count LE, page index LE,
	// 100 LE cells, item count LE (loadsave.cpp:2733-2765)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	wantHead := []byte{
		0x00,                   // StashVersion = 0
		0x39, 0x30, 0x00, 0x00, // gold 12345 LE
		0x01, 0x00, 0x00, 0x00, // one page
		0x04, 0x00, 0x00, 0x00, // page index 4
		0x01, 0x00, // cell 0 = 1
	}
	if !bytes.Equal(encoded[:len(wantHead)], wantHead) {
		t.Fatalf("head = %x, want %x", encoded[:len(wantHead)], wantHead)
	}

	cell57 := 13 + 57*2
	if got := encoded[cell57 : cell57+2]; !bytes.Equal(got, []byte{0x09, 0x00}) {
		t.Fatalf("cell 57 = %x, want 0900", got)
	}
}

func TestStashRoundTrip(t *testing.T) {
	// given a stash with two pages of items
	save := stashFixture()
	// when it is encoded and decoded back
	decoded := roundTripStash(t, save, persist.ItemSaveSizeDiablo)
	// then gold, page grid, items and current page round-trip
	assertStashEqual(t, decoded, save)
}

func TestStashEncodeRejectsWrongItemSize(t *testing.T) {
	// given an item record of the wrong stride
	save := stashFixture()
	save.Items = [][]byte{make([]byte, 10)}
	// when it is encoded
	_, err := save.Encode(persist.ItemSaveSizeDiablo)
	// then the mismatch is reported
	if !errors.Is(err, persist.ErrShortRecord) {
		t.Fatalf("err = %v, want ErrShortRecord", err)
	}
}

func TestDecodeStashSaveRejectsUnknownVersion(t *testing.T) {
	// given a stream with a future version byte
	buf := make([]byte, 9)
	buf[0] = 7
	// when it is decoded
	_, err := persist.DecodeStashSave(buf, persist.ItemSaveSizeDiablo)
	// then the version is rejected
	if !errors.Is(err, persist.ErrUnsupportedVersion) {
		t.Fatalf("err = %v, want ErrUnsupportedVersion", err)
	}
}

func TestDecodeStashSaveShortStream(t *testing.T) {
	// given a stream truncated after the gold field
	buf := []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00}
	// when it is decoded
	_, err := persist.DecodeStashSave(buf, persist.ItemSaveSizeDiablo)
	// then the truncation is reported
	if !errors.Is(err, persist.ErrShortRecord) {
		t.Fatalf("err = %v, want ErrShortRecord", err)
	}
}

func stashFixture() persist.StashSave {
	page := persist.StashPage{Index: 4}
	page.Cells[0] = 1
	page.Cells[57] = 9

	return persist.StashSave{
		Gold:        12345,
		Pages:       []persist.StashPage{page},
		Items:       [][]byte{make([]byte, persist.ItemSaveSizeDiablo), stashItemFixture()},
		CurrentPage: 4,
	}
}

func stashItemFixture() []byte {
	item := make([]byte, persist.ItemSaveSizeDiablo)
	item[0], item[1] = 7, 8

	return item
}

func roundTripStash(t *testing.T, save persist.StashSave, itemSize int) persist.StashSave {
	t.Helper()

	encoded, err := save.Encode(itemSize)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	decoded, err := persist.DecodeStashSave(encoded, itemSize)
	if err != nil {
		t.Fatalf("DecodeStashSave: %v", err)
	}

	return decoded
}

func assertStashEqual(t *testing.T, decoded, save persist.StashSave) {
	t.Helper()

	if decoded.Gold != save.Gold || decoded.CurrentPage != save.CurrentPage {
		t.Fatalf("decoded = %+v", decoded)
	}

	if len(decoded.Pages) != 1 || decoded.Pages[0] != save.Pages[0] {
		t.Fatalf("pages = %+v", decoded.Pages)
	}

	assertStashItems(t, decoded.Items, save.Items)
}

func assertStashItems(t *testing.T, got, want [][]byte) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("item count = %d, want %d", len(got), len(want))
	}

	for i := range got {
		if !bytes.Equal(got[i], want[i]) {
			t.Fatalf("item %d changed", i)
		}
	}
}
