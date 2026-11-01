package persist_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

func TestHotkeysSize(t *testing.T) {
	// when the entry size is computed (loadsave.cpp:2332-2337)
	got := persist.HotkeysSize(persist.NumHotkeys)
	// then it is 1 + 12*4 + 12*1 + 4 + 1 for the default hotkey count
	if got != 66 {
		t.Fatalf("HotkeysSize(12) = %d, want 66", got)
	}

	if got := persist.HotkeysSize(4); got != 26 {
		t.Fatalf("HotkeysSize(4) = %d, want 26 (old save format)", got)
	}
}

func TestHotkeysEncodeGolden(t *testing.T) {
	// given two hotkeys and a selected spell
	hotkeys := persist.HeroHotkeys{
		Hotkeys:     []persist.SpellHotkey{{SpellID: 3, SpellType: 1}, {SpellID: 19, SpellType: 2}},
		ActiveSpell: 19,
		ActiveType:  2,
	}
	// when the stream is encoded (loadsave.cpp:2384-2397)
	image := hotkeys.Encode()
	// then the header carries the count, ids are int32 of int8, types follow,
	// and the active spell closes the stream
	want := []byte{
		0x02,                   // count
		0x03, 0x00, 0x00, 0x00, // hotkey 0 spell id
		0x13, 0x00, 0x00, 0x00, // hotkey 1 spell id (19)
		0x01, 0x02, // spell types
		0x13, 0x00, 0x00, 0x00, // _pRSpell
		0x02, // _pRSplType
	}
	if !bytes.Equal(image, want) {
		t.Fatalf("image = %x, want %x", image, want)
	}
}

func TestHotkeysRoundTrip(t *testing.T) {
	// given a hotkey list with a negative spell id
	hotkeys := persist.HeroHotkeys{
		Hotkeys:     []persist.SpellHotkey{{SpellID: -1, SpellType: 0}, {SpellID: 25, SpellType: 4}},
		ActiveSpell: -1,
		ActiveType:  0,
	}
	// when it is encoded and decoded back
	decoded, err := persist.DecodeHotkeys(hotkeys.Encode())
	// then the ids, types and active spell come back unchanged
	if err != nil {
		t.Fatalf("DecodeHotkeys: %v", err)
	}

	if len(decoded.Hotkeys) != 2 || decoded.Hotkeys[0] != hotkeys.Hotkeys[0] ||
		decoded.Hotkeys[1] != hotkeys.Hotkeys[1] {
		t.Fatalf("hotkeys = %+v", decoded.Hotkeys)
	}

	if decoded.ActiveSpell != -1 || decoded.ActiveType != 0 {
		t.Fatalf("active = %d/%d", decoded.ActiveSpell, decoded.ActiveType)
	}
}

func TestHotkeysEncodeClampsToNumHotkeys(t *testing.T) {
	// given more hotkeys than the format allows
	hotkeys := persist.HeroHotkeys{Hotkeys: make([]persist.SpellHotkey, persist.NumHotkeys+3)}
	// when it is encoded
	image := hotkeys.Encode()
	// then only NumHotkeys pairs are written and the size matches
	if len(image) != persist.HotkeysSize(persist.NumHotkeys) {
		t.Fatalf("len = %d, want %d", len(image), persist.HotkeysSize(persist.NumHotkeys))
	}

	if image[0] != persist.NumHotkeys {
		t.Fatalf("count byte = %d, want %d", image[0], persist.NumHotkeys)
	}
}

func TestDecodeHotkeysRejectsCount(t *testing.T) {
	// given a stream claiming more hotkeys than NumHotkeys (loadsave.cpp:2346-2352)
	buf := make([]byte, persist.HotkeysSize(persist.NumHotkeys+1))
	buf[0] = persist.NumHotkeys + 1
	// when it is decoded
	_, err := persist.DecodeHotkeys(buf)
	// then the count is rejected
	if !errors.Is(err, persist.ErrInvalidHotkeys) {
		t.Fatalf("err = %v, want ErrInvalidHotkeys", err)
	}
}
