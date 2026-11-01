package persist

// ItemPackSize is the byte size of struct ItemPack
// (Source/pack.h:20-31): 4+2+2+1+1+1+1+1+2+4.
const ItemPackSize = 19

// PackedItemNone is the idx of an empty slot (Source/pack.cpp:118 sets
// packedItem.idx = 0xFFFF for empty items).
const PackedItemNone = 0xFFFF

// ItemPack is struct ItemPack (Source/pack.h:20-31), the minified item record
// stored in hero files and game saves. Split-sizes are asserted by
// TestItemPackSize.
type ItemPack struct {
	// Seed is iSeed (pack.h:21): the random item seed.
	Seed uint32
	// CreateInfo is iCreateInfo (pack.h:22): CF_* creation flags + level.
	CreateInfo uint16
	// Idx is idx (pack.h:23): the itemdat row, 0xFFFF for an empty slot.
	Idx uint16
	// BId is bId (pack.h:24): magical << 1 | identified.
	BId uint8
	// BDur is bDur (pack.h:25): durability.
	BDur uint8
	// BMDur is bMDur (pack.h:26): max durability.
	BMDur uint8
	// BCh is bCh (pack.h:27): charges.
	BCh uint8
	// BMCh is bMCh (pack.h:28): max charges.
	BMCh uint8
	// Value is wValue (pack.h:29): gold amount, or the packed ear cursor.
	Value uint16
	// Buff is dwBuff (pack.h:30): item flags and spell/affix effects.
	Buff uint32
}

// Encode returns the record image written into save data.
func (p ItemPack) Encode() []byte {
	var out recordWriter

	out.le32(p.Seed)
	out.le16(p.CreateInfo)
	out.le16(p.Idx)
	out.u8(p.BId)
	out.u8(p.BDur)
	out.u8(p.BMDur)
	out.u8(p.BCh)
	out.u8(p.BMCh)
	out.le16(p.Value)
	out.le32(p.Buff)

	return out.buf
}

// DecodeItemPack reads a record image produced by Encode.
func DecodeItemPack(buf []byte) (ItemPack, error) {
	var in recordReader

	in.buf = buf
	packed := ItemPack{
		Seed:       in.le32(),
		CreateInfo: in.le16(),
		Idx:        in.le16(),
		BId:        in.u8(),
		BDur:       in.u8(),
		BMDur:      in.u8(),
		BCh:        in.u8(),
		BMCh:       in.u8(),
		Value:      in.le16(),
		Buff:       in.le32(),
	}

	if in.err != nil {
		return ItemPack{}, in.err
	}

	return packed, nil
}

// writeItemPacks appends a fixed-size item-record array.
func writeItemPacks(out *recordWriter, packs []ItemPack) {
	for _, pack := range packs {
		out.bytes(pack.Encode())
	}
}

// readItemPacks reads a fixed-size item-record array.
func readItemPacks(in *recordReader, count int) []ItemPack {
	packs := make([]ItemPack, count)
	for i := range packs {
		packs[i] = readItemPack(in)
	}

	return packs
}

// readItemPack reads one record image from the stream.
func readItemPack(in *recordReader) ItemPack {
	return ItemPack{
		Seed:       in.le32(),
		CreateInfo: in.le16(),
		Idx:        in.le16(),
		BId:        in.u8(),
		BDur:       in.u8(),
		BMDur:      in.u8(),
		BCh:        in.u8(),
		BMCh:       in.u8(),
		Value:      in.le16(),
		Buff:       in.le32(),
	}
}
