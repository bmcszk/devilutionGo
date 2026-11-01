package persist

import "fmt"

// Hero-item slot counts (Source/player.h: NUM_INVLOC, InventoryGridCells,
// MaxBeltItems) and the entry name (Source/loadsave.cpp:2701).
const (
	// HeroItemsEntryName is the archive entry SaveHeroItems writes.
	HeroItemsEntryName = "heroitems"
	// NumInvLoc is NUM_INVLOC: worn equipment slots.
	NumInvLoc = 7
	// InventoryGridCells is InventoryGridCells: backpack cells.
	InventoryGridCells = 40
	// MaxBeltItems is MaxBeltItems: belt slots.
	MaxBeltItems = 8
	// heroItemCount is the record count of the entry.
	heroItemCount = NumInvLoc + InventoryGridCells + MaxBeltItems
)

// HeroItems is the heroitems entry: the save-game Hellfire flag plus the worn,
// backpack and belt SaveItem records (Source/loadsave.cpp:2697-2710).
type HeroItems struct {
	// Hellfire is gbIsHellfireSaveGame, written as one byte.
	Hellfire bool
	// Body is InvBody: NumInvLoc records.
	Body [][]byte
	// Inventory is InvList: InventoryGridCells records.
	Inventory [][]byte
	// Belt is SpdList: MaxBeltItems records.
	Belt [][]byte
}

// Size is the entry length for a mode's item stride
// (loadsave.cpp:2701-2702).
func (HeroItems) Size(itemSize int) int {
	return 1 + heroItemCount*itemSize
}

// HellfireFlag is the stored gbIsHellfireSaveGame byte
// (loadsave.cpp:2704).
func (h HeroItems) HellfireFlag() uint8 {
	if h.Hellfire {
		return 1
	}

	return 0
}

// Encode returns the heroitems stream; every record must be itemSize bytes.
func (h HeroItems) Encode(itemSize int) ([]byte, error) {
	records := h.records()
	if len(records) != heroItemCount {
		return nil, fmt.Errorf("persist: heroitems has %d records, want %d: %w", len(records), heroItemCount,
			ErrShortRecord)
	}

	out := make([]byte, 0, h.Size(itemSize))
	out = append(out, h.HellfireFlag())

	for i, record := range records {
		if len(record) != itemSize {
			return nil, fmt.Errorf("persist: heroitems record %d len %d != %d: %w", i, len(record), itemSize,
				ErrShortRecord)
		}

		out = append(out, record...)
	}

	return out, nil
}

// DecodeHeroItems parses a heroitems stream; itemSize is the mode's SaveItem
// stride because the stored flag is not trusted for sizing
// (loadsave.cpp:2396-2410).
func DecodeHeroItems(buf []byte, itemSize int) (HeroItems, error) {
	in := recordReader{buf: buf}

	save := HeroItems{Hellfire: in.u8() != 0}
	save.Body = readItemRecords(&in, NumInvLoc, itemSize)
	save.Inventory = readItemRecords(&in, InventoryGridCells, itemSize)
	save.Belt = readItemRecords(&in, MaxBeltItems, itemSize)

	if in.err != nil {
		return HeroItems{}, in.err
	}

	return save, nil
}

// records returns the worn, backpack and belt records in save order.
func (h HeroItems) records() [][]byte {
	out := make([][]byte, 0, heroItemCount)

	for _, block := range [][][]byte{h.Body, h.Inventory, h.Belt} {
		out = append(out, block...)
	}

	return out
}

// readItemRecords reads count fixed-stride records.
func readItemRecords(in *recordReader, count, itemSize int) [][]byte {
	records := make([][]byte, 0, count)

	for range count {
		records = append(records, in.bytes(itemSize))
	}

	return records
}
