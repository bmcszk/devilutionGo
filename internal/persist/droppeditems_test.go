package persist_test

import (
	"bytes"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

func TestDroppedItemsIndexTables(t *testing.T) {
	// given no active items
	items := persist.DroppedItems{}
	// when the block is encoded (loadsave.cpp:1838-1841)
	image, err := items.Encode(persist.ItemSaveSizeDiablo)
	// then it is two MAXITEMS tables: 0..126 and 0..126
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	if len(image) != 2*persist.MaxItems {
		t.Fatalf("len = %d, want %d", len(image), 2*persist.MaxItems)
	}

	for i := range persist.MaxItems {
		if image[i] != uint8(i) || image[persist.MaxItems+i] != uint8(i) {
			t.Fatalf("index %d = %d/%d", i, image[i], image[persist.MaxItems+i])
		}
	}
}

func TestDroppedItemsAvailableTableOffset(t *testing.T) {
	// given three active items
	records := [][]byte{
		make([]byte, persist.ItemSaveSizeDiablo), make([]byte, persist.ItemSaveSizeDiablo),
		make([]byte, persist.ItemSaveSizeDiablo),
	}
	items := persist.DroppedItems{ActiveCount: 3, Records: records}
	// when the block is encoded
	image, err := items.Encode(persist.ItemSaveSizeDiablo)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	// then the AvailableItems table starts at ActiveItemCount (loadsave.cpp:1840)
	for i := range persist.MaxItems {
		want := uint8((i + 3) % persist.MaxItems)
		if got := image[persist.MaxItems+i]; got != want {
			t.Fatalf("available[%d] = %d, want %d", i, got, want)
		}
	}
}

func TestDroppedItemsRoundTrip(t *testing.T) {
	// given two item records with distinct markers
	first := make([]byte, persist.ItemSaveSizeHellfire)
	first[0] = 7
	second := make([]byte, persist.ItemSaveSizeHellfire)
	second[0] = 9

	items := persist.DroppedItems{ActiveCount: 2, Records: [][]byte{first, second}}
	image, err := items.Encode(persist.ItemSaveSizeHellfire)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	// when the block is decoded back
	decoded, err := persist.DecodeDroppedItems(image, persist.ItemSaveSizeHellfire)
	// then both records survive with their stride
	if err != nil {
		t.Fatalf("DecodeDroppedItems: %v", err)
	}

	if decoded.ActiveCount != 2 || len(decoded.Records) != 2 {
		t.Fatalf("decoded = %+v", decoded)
	}

	if !bytes.Equal(decoded.Records[0], first) || !bytes.Equal(decoded.Records[1], second) {
		t.Fatal("records changed")
	}
}

func TestDroppedItemsRejectsMismatch(t *testing.T) {
	// given an active count that does not match the record list
	items := persist.DroppedItems{ActiveCount: 2, Records: [][]byte{make([]byte, persist.ItemSaveSizeDiablo)}}
	// when it is encoded
	_, err := items.Encode(persist.ItemSaveSizeDiablo)
	// then the mismatch is reported
	if err == nil {
		t.Fatal("mismatched counts encoded")
	}
}

func TestDroppedItemLocations(t *testing.T) {
	// given a dItem grid with two dropped items
	var cells [persist.DungeonWidth * persist.DungeonHeight]uint8
	cells[0] = 1
	cells[persist.DungeonWidth+5] = 2
	// when the grid is encoded (loadsave.cpp:1858-1866)
	image := persist.DroppedItemLocations(cells)
	// then it is one byte per cell with the positions preserved
	if len(image) != persist.DungeonWidth*persist.DungeonHeight {
		t.Fatalf("len = %d", len(image))
	}

	if image[0] != 1 || image[persist.DungeonWidth+5] != 2 {
		t.Fatalf("cells = %d/%d", image[0], image[persist.DungeonWidth+5])
	}
}
