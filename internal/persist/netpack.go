package persist

import (
	"github.com/bmcszk/devilutionGo/internal/diablo/entities"
	"github.com/bmcszk/devilutionGo/internal/diablo/inventory"
)

// Net-pack record sizes from Source/pack.h:84-146 under #pragma pack(push,1):
const (
	// ItemNetDefSize is TItemDef (Source/msg.h:528-532): 2+2+4.
	ItemNetDefSize = 8
	// ItemNetPackSize is union ItemNetPack (pack.h:84-88): the widest member
	// is TEar (Source/msg.h:549-555), 2+2+4+1+17 = 26.
	ItemNetPackSize = 26
	// TItemTailSize is the TItem body after the 8-byte TItemDef head
	// (msg.h:534-547): bId..dwBuff + wToHit + wMaxDam = 15, padded to 18.
	TItemTailSize = 18
	// TEarBodySize is the TEar body after the head: bCursval + heroname[17].
	TEarBodySize = 18
	// NetSpellLevels is MAX_SPELLS (pack.h:18).
	NetSpellLevels = 52
	// PlayerNetPackSize is struct PlayerNetPack (pack.h:90-146):
	// 4+32+7+4+16+52+8+7*26+40*26+40+1+8*26+1+2+1+1+1 + 23*4.
	PlayerNetPackSize = 1692
	// heroNameBytes is the max copied heroname chars (TEar heroname[17], msg.h:554; byte 17 is the NUL).
	heroNameBytes = 16
	// earCursorMask and earCursorShift split the packed cursor byte
	// (items.cpp:3582-3583).
	earCursorMask  = 0x3F
	earCursorShift = 6
)

// ItemNetPack is union ItemNetPack (Source/pack.h:84-88): TItemDef, TItem and
// TEar share the 8-byte TItemDef head, and TEar (26 bytes) sets the size.
type ItemNetPack struct {
	// Idx is wIndx (msg.h:529): the itemdat row, 0xFFFF when empty.
	Idx int16
	// CreateInfo is wCI (msg.h:530).
	CreateInfo uint16
	// Seed is dwSeed (msg.h:531).
	Seed uint32
	// Body is the union area after the shared head.
	Body [ItemNetPackSize - ItemNetDefSize]byte
}

// NetItemStats is the PrepareItemForNetwork result (items.cpp), which the
// item tables compute; it is injected here so the record stays table-free.
type NetItemStats struct {
	ToHit  int
	MaxDam int
}

// Encode returns the 26-byte union image.
func (p ItemNetPack) Encode() []byte {
	var out recordWriter

	out.le16(uint16(p.Idx))
	out.le16(p.CreateInfo)
	out.le32(p.Seed)
	out.bytes(p.Body[:])

	return out.buf
}

// DecodeItemNetPack reads a union image produced by Encode.
func DecodeItemNetPack(buf []byte) (ItemNetPack, error) {
	in := recordReader{buf: buf}
	packed := ItemNetPack{
		Idx:        int16(in.le16()),
		CreateInfo: in.le16(),
		Seed:       in.le32(),
	}
	copy(packed.Body[:], in.bytes(len(packed.Body)))

	if in.err != nil {
		return ItemNetPack{}, in.err
	}

	return packed, nil
}

// PackNetItem packs an item for the network (Source/pack.cpp:205-217).
func PackNetItem(item entities.Item, stats NetItemStats, env PackEnv) ItemNetPack {
	if inventory.IsEmpty(item) {
		return ItemNetPack{Idx: -1}
	}

	packed := ItemNetPack{Idx: item.IDidx, CreateInfo: item.CreateInfo, Seed: item.Seed}
	writeTItemBody(&packed.Body, item, stats, env)

	return packed
}

// PackNetEar packs an ear for the network (Source/pack.cpp:212-216 with
// PrepareEarForNetwork).
func PackNetEar(packed ItemNetPack, cursval uint8, heroName []byte) ItemNetPack {
	packed.Body[0] = cursval

	for i := 0; i < heroNameBytes && i < len(heroName); i++ {
		packed.Body[1+i] = heroName[i]
	}

	return packed
}

// EarCursorValue rebuilds the packed cursor byte from a cursor graphic and
// value (items.cpp:3582-3583).
func EarCursorValue(cursor uint8, value int) uint8 {
	return uint8(value)&earCursorMask | (cursor-EarCursorSorcerer)<<earCursorShift
}

// writeTItemBody writes the TItem tail (msg.h:538-547).
func writeTItemBody(body *[TItemTailSize]byte, item entities.Item, stats NetItemStats, env PackEnv) {
	var out recordWriter

	packed := PackItem(item, env)
	out.u8(packed.BId)
	out.u8(packed.BDur)
	out.u8(packed.BMDur)
	out.u8(packed.BCh)
	out.u8(packed.BMCh)
	out.le16(packed.Value)
	out.le32(packed.Buff)
	out.le16(uint16(stats.ToHit))
	out.le16(uint16(stats.MaxDam))

	copy(body[:], out.buf)
}
