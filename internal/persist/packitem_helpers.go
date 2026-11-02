package persist

import "github.com/bmcszk/devilutionGo/internal/diablo/entities"

// remapSearchLimit bounds reverseRemap: the full itemdat row count.
const remapSearchLimit = 200

// nameMask masks each ear name byte to 7 bits (items.cpp:3578-3583 uses
// & 0x7F so names never carry the high bit).
const nameMask = 0x7F

// loadBE32 reads a big-endian uint32 out of the ear name bytes
// (Source/utils/endian_read.hpp:32-36, called from pack.cpp:131).
func loadBE32(block []byte) uint32 {
	return uint32(block[0])<<24 | uint32(block[1])<<16 | uint32(block[2])<<8 | uint32(block[3])
}

// minUint8 is std::min over durability bytes (pack.cpp:136).
func minUint8(value, limit int) uint8 {
	if value < 0 {
		return 0
	}

	if value > int(limit) {
		return uint8(limit)
	}

	return uint8(value)
}

// reverseRemap finds the record index whose remapped value is idx. The
// upstream remaps are item-table permutations (RemapItemIdxToDiablo /
// RemapItemIdxToSpawn, items.cpp), so the inverse is a table scan; the scan
// bound is the itemdat row count the caller's remap can emit.
func reverseRemap(idx int16, remap func(int16) int16) int16 {
	for candidate := int16(0); int(candidate) < remapSearchLimit; candidate++ {
		if remap(candidate) == idx {
			return candidate
		}
	}

	return idx
}

// applyEarFields stores the fields RecreateEar derives from the packed cursor
// byte (items.cpp:3582-3585).
func applyEarFields(item *entities.Item, packed ItemPack) {
	cursor := uint8(packed.Value) & 0xFF
	item.Curs = cursor>>6&3 + EarCursorSorcerer
	item.Ivalue = int(cursor & 0x3F)
	item.CreateInfo = packed.CreateInfo
	item.Seed = packed.Seed
}
