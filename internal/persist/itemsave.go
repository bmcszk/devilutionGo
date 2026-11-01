package persist

// Item-save constants (Source/items.h:36, :185, Source/loadsave.cpp:2114-2115,
// Source/levels/dun_tile.hpp:7, 137-141).
const (
	// SaveItemNameLength is ItemNameLength: the fixed name buffer size.
	SaveItemNameLength = 64
	// SaveItemAnimWidth is ItemAnimWidth, written for vanilla compatibility.
	SaveItemAnimWidth = 96
	// SaveItemSpriteTileCenter is CalculateSpriteTileCenterX(ItemAnimWidth)
	// with TILE_WIDTH 64.
	SaveItemSpriteTileCenter = (SaveItemAnimWidth - 64) / 2
)

// ItemSave is struct Item (items.h) as SaveItem serializes it
// (Source/loadsave.cpp:1164-1260). Names are fixed 64-byte buffers and the
// signed one-byte fields keep upstream's int8_t width.
type ItemSave struct {
	Seed       uint32
	CreateInfo int16
	Type       int32
	PosX       int32
	PosY       int32
	AnimFlag   uint32
	NumFrames  int32
	CurFrame   int32
	Selection  uint8
	PostDraw   uint32
	Identified uint32

	Magical int8
	Name    [SaveItemNameLength]byte
	IName   [SaveItemNameLength]byte

	Loc        int8
	Class      uint8
	Curs       int32
	Value      int32
	IValue     int32
	MinDam     int32
	MaxDam     int32
	AC         int32
	Flags      uint32
	MiscID     int32
	Spell      int32
	Charges    int32
	MaxCharges int32
	Durability int32
	MaxDur     int32
	PLDam      int32
	PLToHit    int32
	PLAC       int32
	PLStr      int32
	PLMag      int32
	PLDex      int32
	PLVit      int32
	PLFR       int32
	PLLR       int32
	PLMR       int32
	PLMana     int32
	PLHP       int32
	PLDamMod   int32
	PLGetHit   int32
	PLLight    int32
	SplLvlAdd  int8
	Request    int8
	UniqueItem int32
	FMinDam    int32
	FMaxDam    int32
	LMinDam    int32
	LMaxDam    int32
	PLEnAc     int32
	PrePower   int8
	SufPower   int8
	VAdd1      int32
	VMult1     int32
	VAdd2      int32
	VMult2     int32
	MinStr     int8
	MinMag     uint8
	MinDex     int8
	StatFlag   uint32
	IDIdx      int32
	DwBuff     uint32
	DamAcFlags uint32
}

// SaveItemSize is the record length of a mode (loadsave.cpp:2114-2115).
func SaveItemSize(mode GameMode) int {
	return ItemSizeFor(mode)
}

// Encode returns the SaveItem image: 368 bytes, plus _iDamAcFlags under
// Hellfire (loadsave.cpp:1259-1260).
func (s ItemSave) Encode(mode GameMode) []byte {
	var out recordWriter

	s.encodeHead(&out)
	out.zeroes(1)       // Alignment after Loc/Class
	s.encodeStats(&out) // curs..PLLight
	s.encodeTail(&out, mode)

	return out.buf
}

// encodeHead writes seed through IName (loadsave.cpp:1176-1201).
func (s ItemSave) encodeHead(out *recordWriter) {
	out.u32(s.Seed)
	out.le16(uint16(s.CreateInfo))
	out.zeroes(2) // Alignment
	out.i32(s.Type)
	out.i32(s.PosX)
	out.i32(s.PosY)
	out.u32(s.AnimFlag)
	out.zeroes(4) // _iAnimData pointer
	out.i32(s.NumFrames)
	out.i32(s.CurFrame + 1)
	out.i32(SaveItemAnimWidth)
	out.i32(SaveItemSpriteTileCenter)
	out.zeroes(4) // _delFlag, unused since 1.02
	out.u8(s.Selection)
	out.zeroes(3) // Alignment
	out.u32(s.PostDraw)
	out.u32(s.Identified)
	out.u8(uint8(s.Magical))
	out.bytes(s.Name[:])
	out.bytes(s.IName[:])
	out.u8(uint8(s.Loc))
	out.u8(s.Class)
}

// encodeStats writes curs through PLLight (loadsave.cpp:1202-1236).
func (s ItemSave) encodeStats(out *recordWriter) {
	out.i32(s.Curs)
	out.i32(s.Value)
	out.i32(s.IValue)
	out.i32(s.MinDam)
	out.i32(s.MaxDam)
	out.i32(s.AC)
	out.u32(s.Flags)
	out.i32(s.MiscID)
	out.i32(s.Spell)
	out.i32(s.Charges)
	out.i32(s.MaxCharges)
	out.i32(s.Durability)
	out.i32(s.MaxDur)
	s.encodePlayerMods(out)
	out.i32(s.PLMana)
	out.i32(s.PLHP)
	out.i32(s.PLDamMod)
	out.i32(s.PLGetHit)
	out.i32(s.PLLight)
}

// encodePlayerMods writes the _iPL* scalar bonus block
// (loadsave.cpp:1218-1233).
func (s ItemSave) encodePlayerMods(out *recordWriter) {
	out.i32(s.PLDam)
	out.i32(s.PLToHit)
	out.i32(s.PLAC)
	out.i32(s.PLStr)
	out.i32(s.PLMag)
	out.i32(s.PLDex)
	out.i32(s.PLVit)
	out.i32(s.PLFR)
	out.i32(s.PLLR)
	out.i32(s.PLMR)
}

// encodeTail writes splLvlAdd through dwBuff, plus damAcFlags under Hellfire
// (loadsave.cpp:1237-1260).
func (s ItemSave) encodeTail(out *recordWriter, mode GameMode) {
	out.u8(uint8(s.SplLvlAdd))
	out.u8(uint8(s.Request))
	out.zeroes(2) // Alignment
	out.i32(s.UniqueItem)
	out.i32(s.FMinDam)
	out.i32(s.FMaxDam)
	out.i32(s.LMinDam)
	out.i32(s.LMaxDam)
	out.i32(s.PLEnAc)
	out.u8(uint8(s.PrePower))
	out.u8(uint8(s.SufPower))
	out.zeroes(2) // Alignment
	out.i32(s.VAdd1)
	out.i32(s.VMult1)
	out.i32(s.VAdd2)
	out.i32(s.VMult2)
	out.u8(uint8(s.MinStr))
	out.u8(s.MinMag)
	out.u8(uint8(s.MinDex))
	out.zeroes(1) // Alignment
	out.u32(s.StatFlag)
	out.i32(s.IDIdx)
	out.u32(s.DwBuff)

	if mode.Hellfire {
		out.u32(s.DamAcFlags)
	}
}
