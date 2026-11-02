package persist

// Item index remaps between the three itemdat orderings a save can be written
// in: the current (Hellfire) table, the Diablo-only table and the spawn table.
//
// Row indices are the _item_indexes enum values (tables/itemdat.h:24-86); the
// arithmetic below is a verbatim port of loadsave.cpp:2204-2295.

// ItemSorcererRow is IDI_SORCERER (tables/itemdat.h:30).
const ItemSorcererRow int16 = 5

// ItemSorcererDiabloRow is IDI_SORCERER_DIABLO (tables/itemdat.h:82).
const ItemSorcererDiabloRow int16 = 166

// itemIndexNone is IDI_NONE (tables/itemdat.h:86), the "no such row" sentinel
// a remap returns; PackItem turns it into PackedItemNone.
const itemIndexNone int16 = -1

// RemapItemIdxFromDiablo maps a Diablo-only table row to the current table
// (loadsave.cpp:2204-2222).
func RemapItemIdxFromDiablo(i int16) int16 {
	if i == ItemSorcererRow {
		return ItemSorcererDiabloRow
	}

	if i >= 156 {
		i += 5 // Hellfire exclusive items
	}

	if i >= 88 {
		i++ // Scroll of Search
	}

	if i >= 83 {
		i += 4 // Oils
	}

	return i
}

// RemapItemIdxToDiablo maps a current table row to the Diablo-only table,
// returning ItemNone for rows a Diablo build does not have
// (loadsave.cpp:2226-2246).
func RemapItemIdxToDiablo(i int16) int16 {
	if i == ItemSorcererDiabloRow {
		return ItemSorcererRow
	}

	if (i >= 83 && i <= 86) || i == 92 || i >= 161 {
		return itemIndexNone // Hellfire exclusive items
	}

	if i >= 93 {
		i-- // Scroll of Search
	}

	if i >= 87 {
		i -= 4 // Oils
	}

	return i
}

// RemapItemIdxFromSpawn maps a spawn table row to the current table
// (loadsave.cpp:2248-2277).
func RemapItemIdxFromSpawn(i int16) int16 {
	if i >= 62 {
		i += 9 // Medium and heavy armors
	}

	if i >= 96 {
		i++ // Scroll of Stone Curse
	}

	if i >= 98 {
		i++ // Scroll of Guardian
	}

	if i >= 99 {
		i++ // Scroll of ...
	}

	if i >= 101 {
		i++ // Scroll of Golem
	}

	if i >= 102 {
		i++ // Scroll of None
	}

	if i >= 104 {
		i++ // Scroll of Apocalypse
	}

	return i
}

// RemapItemIdxToSpawn maps a current table row to the spawn table
// (loadsave.cpp:2279-2302).
func RemapItemIdxToSpawn(i int16) int16 {
	if i >= 104 {
		i-- // Scroll of Apocalypse
	}

	if i >= 102 {
		i-- // Scroll of None
	}

	if i >= 101 {
		i-- // Scroll of Golem
	}

	if i >= 99 {
		i-- // Scroll of ...
	}

	if i >= 98 {
		i-- // Scroll of Guardian
	}

	if i >= 96 {
		i-- // Scroll of Stone Curse
	}

	if i >= 71 {
		i -= 9 // Medium and heavy armors
	}

	return i
}
