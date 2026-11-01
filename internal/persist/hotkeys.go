package persist

// HotkeysEntryName is the archive entry SaveHotkeys writes
// (Source/loadsave.cpp:2382).
const HotkeysEntryName = "hotkeys"

// NumHotkeys is the spell-hotkey count (Source/player.h:45).
const NumHotkeys = 12

// hotkeyTrailerSize is the active-spell block: _pRSpell int32 plus
// _pRSplType byte (loadsave.cpp:2332-2337).
const hotkeyTrailerSize = 5

// SpellHotkey is one _pSplHotKey/_pSplTHotKey pair
// (Source/loadsave.cpp:2386-2392).
type SpellHotkey struct {
	// SpellID is written as int32 of the int8 spell id.
	SpellID int8
	// SpellType is _pSplTHotKey.
	SpellType uint8
}

// HeroHotkeys is the hotkeys entry: the hotkey list plus the selected spell
// (Source/loadsave.cpp:2380-2397).
type HeroHotkeys struct {
	// Hotkeys holds up to NumHotkeys pairs; shorter slices are zero-filled.
	Hotkeys []SpellHotkey
	// ActiveSpell is _pRSpell.
	ActiveSpell int8
	// ActiveType is _pRSplType.
	ActiveType uint8
}

// HotkeysSize is HotkeysSize(count) (Source/loadsave.cpp:2332-2337): header
// byte, an int32 spell id and a byte spell type per hotkey, then the active
// spell plus its type.
func HotkeysSize(count int) int {
	return 1 + count*4 + count*1 + hotkeyTrailerSize
}

// Encode returns the hotkeys stream, one header byte carrying the written
// hotkey count (loadsave.cpp:2384-2397).
func (h HeroHotkeys) Encode() []byte {
	count := len(h.Hotkeys)
	if count > NumHotkeys {
		count = NumHotkeys
	}

	var out recordWriter

	out.u8(uint8(count))
	out.bytes(spellIDBytes(h.Hotkeys, count))
	out.bytes(spellTypeBytes(h.Hotkeys, count))
	out.i32(int32(h.ActiveSpell))
	out.u8(h.ActiveType)

	return out.buf
}

// DecodeHotkeys parses a hotkeys stream, honouring the stored hotkey count
// (loadsave.cpp:2342-2365).
func DecodeHotkeys(buf []byte) (HeroHotkeys, error) {
	in := recordReader{buf: buf}

	count := int(in.u8())
	if count > NumHotkeys {
		return HeroHotkeys{}, ErrInvalidHotkeys
	}

	hotkeys := make([]SpellHotkey, count)
	for i := range hotkeys {
		hotkeys[i].SpellID = int8(in.le32())
	}

	for i := range hotkeys {
		hotkeys[i].SpellType = in.u8()
	}

	active := int8(in.le32())
	activeType := in.u8()

	if in.err != nil {
		return HeroHotkeys{}, in.err
	}

	return HeroHotkeys{Hotkeys: hotkeys, ActiveSpell: active, ActiveType: activeType}, nil
}

// spellIDBytes writes the _pSplHotKey block as int32 values of the int8 ids
// (loadsave.cpp:2387-2389).
func spellIDBytes(hotkeys []SpellHotkey, count int) []byte {
	var out recordWriter

	for i := range count {
		out.i32(int32(hotkeys[i].SpellID))
	}

	return out.buf
}

// spellTypeBytes writes the _pSplTHotKey block (loadsave.cpp:2390-2392).
func spellTypeBytes(hotkeys []SpellHotkey, count int) []byte {
	out := make([]byte, count)

	for i := range count {
		out[i] = hotkeys[i].SpellType
	}

	return out
}
