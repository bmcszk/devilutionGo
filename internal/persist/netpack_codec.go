package persist

// encodeHead writes plrlevel..pMemSpells (pack.h:91-110).
func (p PlayerNetPack) encodeHead(out *recordWriter) {
	out.u8(p.PlrLevel)
	out.u8(p.Px)
	out.u8(p.Py)
	out.u8(p.Dir)
	out.bytes(p.Name[:])
	out.u8(p.Class)
	out.u8(p.BaseStr)
	out.u8(p.BaseMag)
	out.u8(p.BaseDex)
	out.u8(p.BaseVit)
	out.u8(uint8(p.Level))
	out.u8(p.StatPts)
	out.le32(p.Exp)
	out.le32(uint32(p.HPBase))
	out.le32(uint32(p.MaxHPBase))
	out.le32(uint32(p.ManaBase))
	out.le32(uint32(p.MaxManaBase))
	out.bytes(p.SplLvl[:])
	out.le64(p.MemSpells)
}

// encodeItems writes the item blocks and inventory grid (pack.h:111-115).
func (p PlayerNetPack) encodeItems(out *recordWriter) {
	writeNetItems(out, p.InvBody[:])
	writeNetItems(out, p.InvList[:])

	for _, cell := range p.InvGrid {
		out.u8(uint8(cell))
	}

	out.u8(p.NumInv)
	writeNetItems(out, p.SpdList[:])
}

// encodeTail writes the tail flags and validation words (pack.h:116-146).
func (p PlayerNetPack) encodeTail(out *recordWriter) {
	out.u8(p.ManaShield)
	out.le16(p.Reflections)
	out.u8(p.KillLevel)
	out.u8(p.FriendlyMode)
	out.u8(p.IsOnSetLevel)

	for _, value := range p.Validation {
		out.le32(uint32(value))
	}
}

// decodeHead reads plrlevel..pMemSpells (pack.h:91-110).
func (p *PlayerNetPack) decodeHead(in *recordReader) {
	p.PlrLevel = in.u8()
	p.Px = in.u8()
	p.Py = in.u8()
	p.Dir = in.u8()
	copy(p.Name[:], in.bytes(len(p.Name)))
	p.Class = in.u8()
	p.BaseStr = in.u8()
	p.BaseMag = in.u8()
	p.BaseDex = in.u8()
	p.BaseVit = in.u8()
	p.Level = int8(in.u8())
	p.StatPts = in.u8()
	p.Exp = in.le32()
	p.HPBase = int32(in.le32())
	p.MaxHPBase = int32(in.le32())
	p.ManaBase = int32(in.le32())
	p.MaxManaBase = int32(in.le32())
	copy(p.SplLvl[:], in.bytes(len(p.SplLvl)))
	p.MemSpells = in.le64()
}

// decodeItems reads the item blocks and inventory grid (pack.h:111-115).
func (p *PlayerNetPack) decodeItems(in *recordReader) {
	readNetItems(in, p.InvBody[:])
	readNetItems(in, p.InvList[:])

	for i := range p.InvGrid {
		p.InvGrid[i] = int8(in.u8())
	}

	p.NumInv = in.u8()
	readNetItems(in, p.SpdList[:])
}

// decodeTail reads the tail flags and validation words (pack.h:116-146).
func (p *PlayerNetPack) decodeTail(in *recordReader) {
	p.ManaShield = in.u8()
	p.Reflections = in.le16()
	p.KillLevel = in.u8()
	p.FriendlyMode = in.u8()
	p.IsOnSetLevel = in.u8()

	for i := range p.Validation {
		p.Validation[i] = int32(in.le32())
	}
}

// writeNetItems appends a net-pack item array.
func writeNetItems(out *recordWriter, packs []ItemNetPack) {
	for _, pack := range packs {
		out.bytes(pack.Encode())
	}
}

// readNetItems fills a net-pack item array from the stream.
func readNetItems(in *recordReader, packs []ItemNetPack) {
	for i := range packs {
		pack := ItemNetPack{
			Idx:        int16(in.le16()),
			CreateInfo: in.le16(),
			Seed:       in.le32(),
		}
		copy(pack.Body[:], in.bytes(len(pack.Body)))
		packs[i] = pack
	}
}
