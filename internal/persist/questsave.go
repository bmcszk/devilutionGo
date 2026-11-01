package persist

// Quest save record sizes (Source/loadsave.cpp:1769-1799): the Diablo record
// stores _qmsg as one byte, the Hellfire record widens it to int32.
const (
	// QuestSaveSizeDiablo is the record length without Hellfire.
	QuestSaveSizeDiablo = 44
	// QuestSaveSizeHellfire is the record length with Hellfire.
	QuestSaveSizeHellfire = 48
)

// QuestSave is one SaveQuest record (Source/loadsave.cpp:1769-1799). The
// trailing return-level block is global state upstream writes per quest.
type QuestSave struct {
	// Level is _qlevel.
	Level uint8
	// Idx is _qidx, written twice (:1770 and :1772).
	Idx uint8
	// Active is _qactive.
	Active uint8
	// LvlType is _qlvltype.
	LvlType uint8
	// X and Y are quest.position.
	X int32
	Y int32
	// Slvl is _qslvl.
	Slvl uint8
	// Msg is _qmsg: one byte without Hellfire, int32 with it.
	Msg int32
	// Var1 and Var2 are _qvar1/_qvar2.
	Var1 uint8
	Var2 uint8
	// Log is _qlog.
	Log uint32
	// ReturnX and ReturnY are ReturnLvlPosition, big-endian.
	ReturnX int32
	ReturnY int32
	// ReturnLevel and ReturnLevelType are big-endian.
	ReturnLevel     int32
	ReturnLevelType int32
}

// QuestSaveSize is the record length of a mode (loadsave.cpp:1801-1807).
func QuestSaveSize(mode GameMode) int {
	if mode.Hellfire {
		return QuestSaveSizeHellfire
	}

	return QuestSaveSizeDiablo
}

// Encode returns the quest record image.
func (q QuestSave) Encode(mode GameMode) []byte {
	var out recordWriter

	out.u8(q.Level)
	out.u8(q.Idx)
	out.u8(q.Active)
	out.u8(q.LvlType)
	out.i32(q.X)
	out.i32(q.Y)
	out.u8(q.Slvl)
	out.u8(q.Idx)

	q.encodeMessage(&out, mode)

	out.u8(q.Var1)
	out.u8(q.Var2)
	out.zeroes(2)

	if !mode.Hellfire {
		out.zeroes(1) // Alignment
	}

	out.u32(q.Log)

	out.be32(uint32(q.ReturnX))
	out.be32(uint32(q.ReturnY))
	out.be32(uint32(q.ReturnLevel))
	out.be32(uint32(q.ReturnLevelType))
	out.zeroes(4) // DoomQuestState

	return out.buf
}

// encodeMessage writes _qmsg widened to int32 under Hellfire
// (loadsave.cpp:1774-1780).
func (q QuestSave) encodeMessage(out *recordWriter, mode GameMode) {
	if mode.Hellfire {
		out.zeroes(2) // Alignment
		out.i32(q.Msg)

		return
	}

	out.u8(uint8(q.Msg))
}

// DecodeQuestSave parses a quest record image.
func DecodeQuestSave(buf []byte, mode GameMode) (QuestSave, error) {
	in := recordReader{buf: buf}

	quest := QuestSave{
		Level:   in.u8(),
		Idx:     in.u8(),
		Active:  in.u8(),
		LvlType: in.u8(),
		X:       int32(in.le32()),
		Y:       int32(in.le32()),
		Slvl:    in.u8(),
	}
	in.u8() // _qidx repeated for compatibility

	quest = quest.withMessage(&in, mode)

	quest.Var1 = in.u8()
	quest.Var2 = in.u8()
	in.skip(2)

	if !mode.Hellfire {
		in.skip(1)
	}

	quest.Log = in.le32()
	quest.ReturnX = int32(in.be32())
	quest.ReturnY = int32(in.be32())
	quest.ReturnLevel = int32(in.be32())
	quest.ReturnLevelType = int32(in.be32())
	in.skip(4)

	if in.err != nil {
		return QuestSave{}, in.err
	}

	return quest, nil
}

// withMessage reads _qmsg at its mode width (loadsave.cpp:1774-1780).
func (q QuestSave) withMessage(in *recordReader, mode GameMode) QuestSave {
	if mode.Hellfire {
		in.skip(2)
		q.Msg = int32(in.le32())

		return q
	}

	q.Msg = int32(in.u8())

	return q
}
