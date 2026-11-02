package persist_test

import (
	"bytes"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

func TestHeroItemsSize(t *testing.T) {
	// given the 55 records of the entry
	items := heroItemsFixture(persist.ItemSaveSizeDiablo)
	// when the size is computed (loadsave.cpp:2701-2702)
	got := items.Size(persist.ItemSaveSizeDiablo)
	// then it is one flag byte plus 55 records of the mode stride
	if got != 1+55*368 {
		t.Fatalf("Size = %d, want %d", got, 1+55*368)
	}

	hellfire := heroItemsFixture(persist.ItemSaveSizeHellfire)

	encoded, err := hellfire.Encode(persist.ItemSaveSizeHellfire)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	if len(encoded) != 1+55*372 {
		t.Fatalf("len = %d, want %d", len(encoded), 1+55*372)
	}
}

func TestHeroItemsEncodeHead(t *testing.T) {
	// when the stream is encoded (loadsave.cpp:2704-2710)
	items := heroItemsFixture(persist.ItemSaveSizeDiablo)
	encoded, err := items.Encode(persist.ItemSaveSizeDiablo)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	// then the Hellfire flag is first, followed by InvBody, InvList and SpdList
	if encoded[0] != 1 {
		t.Fatalf("flag = %d, want 1", encoded[0])
	}

	body := encoded[1 : 1+persist.ItemSaveSizeDiablo]
	if body[0] != 1 {
		t.Fatalf("first body record marker = %d", body[0])
	}

	inv := encoded[1+persist.NumInvLoc*persist.ItemSaveSizeDiablo:]
	if inv[0] != 20 {
		t.Fatalf("first inventory record marker = %d, want 20", inv[0])
	}
}

func TestHeroItemsRoundTrip(t *testing.T) {
	// given a hero item set for the Diablo stride
	items := heroItemsFixture(persist.ItemSaveSizeDiablo)
	encoded, err := items.Encode(persist.ItemSaveSizeDiablo)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	// when it is decoded back
	decoded, err := persist.DecodeHeroItems(encoded, persist.ItemSaveSizeDiablo)
	// then all three blocks round-trip
	if err != nil {
		t.Fatalf("DecodeHeroItems: %v", err)
	}

	assertHeroItemsEqual(t, decoded, items)
}

func TestHeroItemsEncodeRejectsWrongRecordCount(t *testing.T) {
	// given a set with a missing body record
	items := heroItemsFixture(persist.ItemSaveSizeDiablo)
	items.Body = items.Body[:persist.NumInvLoc-1]
	// when it is encoded
	_, err := items.Encode(persist.ItemSaveSizeDiablo)
	// then the count mismatch is reported
	if err == nil {
		t.Fatal("encoding a short record set succeeded")
	}
}

func TestDecodeHeroItemsTruncated(t *testing.T) {
	// when a stream cut inside the body block is decoded
	_, err := persist.DecodeHeroItems(make([]byte, 1+persist.ItemSaveSizeDiablo), persist.ItemSaveSizeDiablo)
	// then the truncation is reported
	if err == nil {
		t.Fatal("truncated stream decoded without error")
	}
}

func heroItemsFixture(itemSize int) persist.HeroItems {
	records := func(count, fill int) [][]byte {
		out := make([][]byte, count)
		for i := range out {
			record := make([]byte, itemSize)
			record[0] = byte(fill + i)
			out[i] = record
		}

		return out
	}

	return persist.HeroItems{
		Hellfire:  true,
		Body:      records(persist.NumInvLoc, 1),
		Inventory: records(persist.InventoryGridCells, 20),
		Belt:      records(persist.MaxBeltItems, 60),
	}
}

func assertHeroItemsEqual(t *testing.T, decoded, items persist.HeroItems) {
	t.Helper()

	if !decoded.Hellfire || len(decoded.Body) != persist.NumInvLoc ||
		len(decoded.Inventory) != persist.InventoryGridCells || len(decoded.Belt) != persist.MaxBeltItems {
		t.Fatalf("decoded shape = %d/%d/%d", len(decoded.Body), len(decoded.Inventory), len(decoded.Belt))
	}

	if !bytes.Equal(decoded.Body[3], items.Body[3]) || !bytes.Equal(decoded.Inventory[7], items.Inventory[7]) ||
		!bytes.Equal(decoded.Belt[0], items.Belt[0]) {
		t.Fatal("records changed")
	}
}
