package persist

// PlayerPackSize is the byte size of struct PlayerPack (Source/pack.h:33-88),
// asserted by TestPlayerPackSize: the split below is the pack.h field order,
// 8+8+32+7+4+4+4+28+37+8+133+760+40+1+152+8+2+2+10+2+4+4+4+20.
const PlayerPackSize = 1266

// The pack.h:58-88 constants PlayerPack depends on.
const (
	// SpellLevelsPacked is the pSplLvl array length; "Should be MAX_SPELLS but
	// set to 37 to make save games compatible" (Source/pack.h:58).
	SpellLevelsPacked = 37
	// SpellLevelsPackedHf is pSplLvl2: the Hellfire spells 37..46
	// (Source/pack.h:82).
	SpellLevelsPackedHf = 10
)

// PlayerPack is struct PlayerPack (Source/pack.h:33-88): the hero-file record,
// written as one fixed-layout blob and then run through the codec.
type PlayerPack struct {
	// LowDateTime/HighDateTime are the FILETIME pair (pack.h:34-35).
	LowDateTime  uint32
	HighDateTime uint32

	DestAction int8
	DestParam1 int8
	DestParam2 int8
	PlrLevel   uint8
	Px         uint8
	Py         uint8
	TargX      uint8
	TargY      uint8

	// Name is pName[PlayerNameLength] (pack.h:44).
	Name [32]byte

	Class    uint8
	BaseStr  uint8
	BaseMag  uint8
	BaseDex  uint8
	BaseVit  uint8
	Level    uint8
	StatPts  uint8
	Exp      uint32
	Gold     int32
	HPBase   int32
	MaxHP    int32
	ManaBase int32
	MaxMana  int32

	// SplLvl is pSplLvl[37] (pack.h:58), the vanilla-compatible spell levels.
	SplLvl [SpellLevelsPacked]uint8
	// MemSpells is pMemSpells (pack.h:59).
	MemSpells uint64

	// InvBody is InvBody[NUM_INVLOC] (pack.h:60).
	InvBody [7]ItemPack
	// InvList is InvList[InventoryGridCells] (pack.h:61).
	InvList [40]ItemPack
	// InvGrid is InvGrid[InventoryGridCells] (pack.h:62).
	InvGrid [40]int8
	// NumInv is _pNumInv (pack.h:63).
	NumInv uint8
	// SpdList is SpdList[MaxBeltItems] (pack.h:64).
	SpdList [8]ItemPack

	TownWarps  int8
	DungMsgs   int8
	LvlLoad    int8
	BattleNet  uint8
	ManaShield uint8
	DungMsgs2  uint8
	// IsHellfire is bIsHellfire (pack.h:70): 0 = Diablo, 1 = Hellfire.
	IsHellfire int8
	Reserved   uint8

	Reflections uint16
	// Reserved2 is the pack.h:73 padding pair.
	Reserved2 [2]uint8
	// SplLvl2 is pSplLvl2[10] (pack.h:82): the Hellfire spell levels.
	SplLvl2 [SpellLevelsPackedHf]uint8
	// Reserved8 is wReserved8 (pack.h:83).
	Reserved8 int16

	DiabloKillLevel uint32
	Difficulty      uint32
	DamAcFlags      uint32
	// Reserved3 is the 20 trailing spare bytes (pack.h:87).
	Reserved3 [20]uint8
}

// Encode returns the hero-file record image.
func (p PlayerPack) Encode() []byte {
	var out recordWriter

	p.encodeHead(&out)
	p.encodeStats(&out)
	p.encodeSlots(&out)
	p.encodeTail(&out)

	return out.buf
}

// encodeHead writes the timestamp, destination, position and name block
// (pack.h:34-45).
func (p PlayerPack) encodeHead(out *recordWriter) {
	out.le32(p.LowDateTime)
	out.le32(p.HighDateTime)
	out.u8(uint8(p.DestAction))
	out.u8(uint8(p.DestParam1))
	out.u8(uint8(p.DestParam2))
	out.u8(p.PlrLevel)
	out.u8(p.Px)
	out.u8(p.Py)
	out.u8(p.TargX)
	out.u8(p.TargY)
	out.bytes(p.Name[:])
}

// encodeStats writes the class, stat and spell block (pack.h:46-59).
func (p PlayerPack) encodeStats(out *recordWriter) {
	out.u8(p.Class)
	out.u8(p.BaseStr)
	out.u8(p.BaseMag)
	out.u8(p.BaseDex)
	out.u8(p.BaseVit)
	out.u8(p.Level)
	out.u8(p.StatPts)
	out.le32(p.Exp)
	out.le32(uint32(p.Gold))
	out.le32(uint32(p.HPBase))
	out.le32(uint32(p.MaxHP))
	out.le32(uint32(p.ManaBase))
	out.le32(uint32(p.MaxMana))
	out.bytes(p.SplLvl[:])
	out.le64(p.MemSpells)
}

// encodeSlots writes the item and inventory-grid block (pack.h:60-64).
func (p PlayerPack) encodeSlots(out *recordWriter) {
	writeItemPacks(out, p.InvBody[:])
	writeItemPacks(out, p.InvList[:])

	for _, cell := range p.InvGrid {
		out.u8(uint8(cell))
	}

	out.u8(p.NumInv)
	writeItemPacks(out, p.SpdList[:])
}

// encodeTail writes the message, Hellfire and reserved block (pack.h:65-87).
func (p PlayerPack) encodeTail(out *recordWriter) {
	out.u8(uint8(p.TownWarps))
	out.u8(uint8(p.DungMsgs))
	out.u8(uint8(p.LvlLoad))
	out.u8(p.BattleNet)
	out.u8(p.ManaShield)
	out.u8(p.DungMsgs2)
	out.u8(uint8(p.IsHellfire))
	out.u8(p.Reserved)
	out.le16(p.Reflections)
	out.bytes(p.Reserved2[:])
	out.bytes(p.SplLvl2[:])
	out.le16(uint16(p.Reserved8))
	out.le32(p.DiabloKillLevel)
	out.le32(p.Difficulty)
	out.le32(p.DamAcFlags)
	out.bytes(p.Reserved3[:])
}
