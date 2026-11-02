package persist

import (
	"github.com/bmcszk/devilutionGo/internal/diablo/entities"
	"github.com/bmcszk/devilutionGo/internal/diablo/inventory"
)

const (
	// ItemIndexGold is IDI_GOLD (Source/tables/itemdat.h:25), the first row.
	ItemIndexGold int16 = 0
	// ItemIndexEar is IDI_EAR (Source/tables/itemdat.h:50).
	ItemIndexEar int16 = 23
	// MiscArenaPot is IMISC_ARENAPOT (Source/tables/itemdat.h:489).
	MiscArenaPot int8 = 55
	// SpellNova is SpellID::Nova (Source/tables/spelldat.h:49).
	SpellNova = 19
	// SpellApocalypse is SpellID::Apocalypse (Source/tables/spelldat.h:55).
	SpellApocalypse = 25
	// EarCursorSorcerer is ICURS_EAR_SORCERER, the first ear cursor graphic.
	EarCursorSorcerer uint8 = 19
	// earNameBytes is the 16 name bytes packed into an ear (items.cpp:3573-3586).
	earNameBytes = 16
	// maxDurabilityByte is the byte ceiling used by PackItem
	// (Source/pack.cpp:132-135 clamps _iMaxDur > 255 to 254).
	maxDurabilityByte = 254
	// durabilityByteLimit is 255, the largest _iMaxDur that fits a byte.
	durabilityByteLimit = 255
	// identifiedBit is the low bit of bId (pack.cpp:131).
	identifiedBit = 1
)

// PackEnv carries the mode branches PackItem and UnpackItem take in upstream
// (pack.cpp:121-127): the hellfire flag and the two item-index remaps, which
// live in the itemdat layer and are injected here.
type PackEnv struct {
	// Hellfire is gbIsHellfire.
	Hellfire bool
	// Spawn is gbIsSpawn.
	Spawn bool
	// Vanilla is gbVanilla: PackPlayer then duplicates the tile into targx/y
	// (pack.cpp:162-165).
	Vanilla bool
	// RemapToDiablo is RemapItemIdxToDiablo; nil means identity.
	RemapToDiablo func(idx int16) int16
	// RemapToSpawn is RemapItemIdxToSpawn; nil means identity.
	RemapToSpawn func(idx int16) int16
	// Available is IsItemAvailable (items.cpp:2373-2404); nil means every row
	// is available. It is the gbIsSpawn/gbIsHellfire item-strip guard
	// UnPackItem applies after the remaps (pack.cpp:305-308).
	Available func(idx int16) bool
}

// PackItem minifies an item into its save record (Source/pack.cpp:108-150).
func PackItem(item entities.Item, env PackEnv) ItemPack {
	var packed ItemPack

	// Arena potions don't exist in vanilla so they are not saved
	// (pack.cpp:111-114).
	if item.Itype == int8(inventory.ItemTypeNone) || item.MiscId == MiscArenaPot {
		packed.Idx = PackedItemNone
		return packed
	}

	packed.Idx = uint16(env.remapIndex(item.IDidx))

	if item.IDidx == ItemIndexEar {
		packEar(&packed, item)
		return packed
	}

	packRegular(&packed, item)

	return packed
}

// UnpackItem expands a save record back into item fields
// (Source/pack.cpp:285-360). Item data derived from the item tables
// (RecreateItem, items.cpp) belongs to the itemgen layer and is not repeated
// here: fields the record does not store stay zero.
func UnpackItem(packed ItemPack, env PackEnv) (entities.Item, error) {
	if packed.Idx == PackedItemNone {
		return inventory.EmptyItem(), nil
	}

	idx := env.unremapIndex(int16(packed.Idx))
	if !env.available(idx) {
		return inventory.EmptyItem(), nil
	}

	item := entities.Item{IDidx: idx}
	if idx == ItemIndexEar {
		unpackEar(&item, packed)
		return item, nil
	}

	unpackRegular(&item, packed)

	return item, nil
}

// available is the IsItemAvailable guard of pack.cpp:305-308.
func (e PackEnv) available(idx int16) bool {
	if idx < 0 {
		return false
	}

	if e.Available == nil {
		return true
	}

	return e.Available(idx)
}

// remapIndex applies the diablo and spawn remaps in upstream order
// (pack.cpp:121-127).
func (e PackEnv) remapIndex(idx int16) int16 {
	if !e.Hellfire && e.RemapToDiablo != nil {
		idx = e.RemapToDiablo(idx)
	}

	if e.Spawn && e.RemapToSpawn != nil {
		idx = e.RemapToSpawn(idx)
	}

	return idx
}

// unremapIndex reverses remapIndex (pack.cpp:296-303 applies spawn first).
func (e PackEnv) unremapIndex(idx int16) int16 {
	if e.Spawn && e.RemapToSpawn != nil {
		idx = reverseRemap(idx, e.RemapToSpawn)
	}

	if !e.Hellfire && e.RemapToDiablo != nil {
		idx = reverseRemap(idx, e.RemapToDiablo)
	}

	return idx
}

// packEar stores an ear's data in the packed name bytes (pack.cpp:129-139).
func packEar(packed *ItemPack, item entities.Item) {
	name := item.IName

	packed.CreateInfo = uint16(name[1]) | uint16(name[0])<<8
	packed.Seed = loadBE32(name[2:])
	packed.BId = name[6]
	packed.BDur = name[7]
	packed.BMDur = name[8]
	packed.BCh = name[9]
	packed.BMCh = name[10]
	packed.Value = uint16(uint8(item.Ivalue)) |
		uint16(name[11])<<8 |
		uint16(item.Curs-EarCursorSorcerer)<<6
	packed.Buff = loadBE32(name[12:])
}

// packRegular stores the ordinary item fields (pack.cpp:140-148).
func packRegular(packed *ItemPack, item entities.Item) {
	packed.Seed = item.Seed
	packed.CreateInfo = item.CreateInfo
	packed.BId = item.Magical << 1
	if item.Identified {
		packed.BId |= identifiedBit
	}

	if item.MaxDur > durabilityByteLimit {
		packed.BMDur = maxDurabilityByte
	} else {
		packed.BMDur = uint8(item.MaxDur)
	}

	packed.BDur = minUint8(item.Durability, int(packed.BMDur))
	packed.BCh = uint8(item.Charges)
	packed.BMCh = uint8(item.MaxCharges)

	if item.IDidx == ItemIndexGold {
		packed.Value = uint16(item.Ivalue)
	}

	packed.Buff = item.Buff
}

// unpackEar recovers the ear name, value and cursor (pack.cpp:322-348 plus
// RecreateEar, items.cpp:3573-3586).
func unpackEar(item *entities.Item, packed ItemPack) {
	var name [earNameBytes]byte

	name[0] = byte(packed.CreateInfo>>8) & nameMask
	name[1] = byte(packed.CreateInfo) & nameMask
	name[2] = byte(packed.Seed>>24) & nameMask
	name[3] = byte(packed.Seed>>16) & nameMask
	name[4] = byte(packed.Seed>>8) & nameMask
	name[5] = byte(packed.Seed) & nameMask
	name[6] = packed.BId & nameMask
	name[7] = packed.BDur & nameMask
	name[8] = packed.BMDur & nameMask
	name[9] = packed.BCh & nameMask
	name[10] = packed.BMCh & nameMask
	name[11] = byte(packed.Value>>8) & nameMask
	name[12] = byte(packed.Buff>>24) & nameMask
	name[13] = byte(packed.Buff>>16) & nameMask
	name[14] = byte(packed.Buff>>8) & nameMask
	name[15] = byte(packed.Buff) & nameMask

	copy(item.IName[:], name[:])
	applyEarFields(item, packed)
}

// unpackRegular recovers the ordinary item fields (pack.cpp:349-360).
func unpackRegular(item *entities.Item, packed ItemPack) {
	item.Seed = packed.Seed
	item.CreateInfo = packed.CreateInfo
	item.Magical = packed.BId >> 1
	item.Identified = packed.BId&identifiedBit != 0
	item.MaxDur = int(packed.BMDur)
	item.Durability = int(packed.BDur)
	item.Charges = int(packed.BCh)
	item.MaxCharges = int(packed.BMCh)
	item.Buff = packed.Buff

	if item.IDidx == ItemIndexGold {
		item.Ivalue = int(packed.Value)
	}
}
