package persist

import "fmt"

// MaxItems is MAXITEMS (Source/items.h:26): the item array size the dropped
// item block is padded to.
const MaxItems = 127

// DroppedItems is the dropped-item section of a level stream
// (Source/loadsave.cpp:1835-1855): two MaxItems compatibility index tables
// followed by the SaveItem records of the active items.
type DroppedItems struct {
	// ActiveCount is ActiveItemCount.
	ActiveCount int
	// Records holds one SaveItem image per active item.
	Records [][]byte
}

// Size is the block length for an item stride: two index tables plus the
// records (loadsave.cpp:1838-1852).
func (d DroppedItems) Size(itemSize int) int {
	return 2*MaxItems + len(d.Records)*itemSize
}

// Encode returns the dropped-item block; every record must be itemSize bytes
// and ActiveCount must match the record count.
func (d DroppedItems) Encode(itemSize int) ([]byte, error) {
	if d.ActiveCount != len(d.Records) {
		return nil, fmt.Errorf("persist: dropped items active count %d != %d records: %w", d.ActiveCount,
			len(d.Records), ErrShortRecord)
	}

	if d.ActiveCount > MaxItems {
		return nil, fmt.Errorf("persist: dropped items active count %d > %d: %w", d.ActiveCount, MaxItems,
			ErrShortRecord)
	}

	out := make([]byte, 0, d.Size(itemSize))
	out = append(out, activeIndexTable()...)
	out = append(out, availableIndexTable(d.ActiveCount)...)

	for i, record := range d.Records {
		if len(record) != itemSize {
			return nil, fmt.Errorf("persist: dropped item %d len %d != %d: %w", i, len(record), itemSize,
				ErrShortRecord)
		}

		out = append(out, record...)
	}

	return out, nil
}

// DecodeDroppedItems parses the dropped-item block.
func DecodeDroppedItems(buf []byte, itemSize int) (DroppedItems, error) {
	if len(buf) < 2*MaxItems {
		return DroppedItems{}, fmt.Errorf("persist: dropped items block: %w", ErrShortRecord)
	}

	rest := buf[2*MaxItems:]
	if len(rest)%itemSize != 0 {
		return DroppedItems{}, fmt.Errorf("persist: dropped items tail len %d not a multiple of %d: %w", len(rest),
			itemSize, ErrShortRecord)
	}

	save := DroppedItems{ActiveCount: len(rest) / itemSize}

	for offset := 0; offset < len(rest); offset += itemSize {
		record := make([]byte, itemSize)
		copy(record, rest[offset:offset+itemSize])

		save.Records = append(save.Records, record)
	}

	return save, nil
}

// DroppedItemLocations is the dItem grid remapped to save positions
// (Source/loadsave.cpp:1858-1866): one byte per cell.
func DroppedItemLocations(cells [dungeonGridCells]uint8) []byte {
	out := make([]byte, dungeonGridCells)
	copy(out, cells[:])

	return out
}

// activeIndexTable is the ActiveItems compatibility table: 0..MAXITEMS-1
// (loadsave.cpp:1838-1839).
func activeIndexTable() []byte {
	out := make([]byte, MaxItems)

	for i := range out {
		out[i] = uint8(i)
	}

	return out
}

// availableIndexTable is the AvailableItems table, offset by ActiveItemCount
// (loadsave.cpp:1840-1841).
func availableIndexTable(activeCount int) []byte {
	out := make([]byte, MaxItems)

	for i := range out {
		out[i] = uint8((i + activeCount) % MaxItems)
	}

	return out
}
