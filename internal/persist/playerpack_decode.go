package persist

// DecodePlayerPack reads a hero-file record image produced by
// PlayerPack.Encode.
func DecodePlayerPack(buf []byte) (PlayerPack, error) {
	var packed PlayerPack

	in := recordReader{buf: buf}
	packed.decodeHead(&in)
	packed.decodeStats(&in)
	packed.decodeSlots(&in)
	packed.decodeTail(&in)

	if in.err != nil {
		return PlayerPack{}, in.err
	}

	return packed, nil
}

// decodeHead reads the timestamp, destination, position and name block
// (pack.h:34-45).
func (p *PlayerPack) decodeHead(in *recordReader) {
	p.LowDateTime = in.le32()
	p.HighDateTime = in.le32()
	p.DestAction = int8(in.u8())
	p.DestParam1 = int8(in.u8())
	p.DestParam2 = int8(in.u8())
	p.PlrLevel = in.u8()
	p.Px = in.u8()
	p.Py = in.u8()
	p.TargX = in.u8()
	p.TargY = in.u8()
	copy(p.Name[:], in.bytes(len(p.Name)))
}

// decodeStats reads the class, stat and spell block (pack.h:46-59).
func (p *PlayerPack) decodeStats(in *recordReader) {
	p.Class = in.u8()
	p.BaseStr = in.u8()
	p.BaseMag = in.u8()
	p.BaseDex = in.u8()
	p.BaseVit = in.u8()
	p.Level = in.u8()
	p.StatPts = in.u8()
	p.Exp = in.le32()
	p.Gold = int32(in.le32())
	p.HPBase = int32(in.le32())
	p.MaxHP = int32(in.le32())
	p.ManaBase = int32(in.le32())
	p.MaxMana = int32(in.le32())
	copy(p.SplLvl[:], in.bytes(len(p.SplLvl)))
	p.MemSpells = in.le64()
}

// decodeSlots reads the item and inventory-grid block (pack.h:60-64).
func (p *PlayerPack) decodeSlots(in *recordReader) {
	copy(p.InvBody[:], readItemPacks(in, len(p.InvBody)))
	copy(p.InvList[:], readItemPacks(in, len(p.InvList)))

	for i := range p.InvGrid {
		p.InvGrid[i] = int8(in.u8())
	}

	p.NumInv = in.u8()
	copy(p.SpdList[:], readItemPacks(in, len(p.SpdList)))
}

// decodeTail reads the message, Hellfire and reserved block (pack.h:65-87).
func (p *PlayerPack) decodeTail(in *recordReader) {
	p.TownWarps = int8(in.u8())
	p.DungMsgs = int8(in.u8())
	p.LvlLoad = int8(in.u8())
	p.BattleNet = in.u8()
	p.ManaShield = in.u8()
	p.DungMsgs2 = in.u8()
	p.IsHellfire = int8(in.u8())
	p.Reserved = in.u8()
	p.Reflections = in.le16()
	copy(p.Reserved2[:], in.bytes(len(p.Reserved2)))
	copy(p.SplLvl2[:], in.bytes(len(p.SplLvl2)))
	p.Reserved8 = int16(in.le16())
	p.DiabloKillLevel = in.le32()
	p.Difficulty = in.le32()
	p.DamAcFlags = in.le32()
	copy(p.Reserved3[:], in.bytes(len(p.Reserved3)))
}
