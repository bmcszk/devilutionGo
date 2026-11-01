package persist_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

func TestQuestSaveSize(t *testing.T) {
	// when a quest record is encoded for each mode (loadsave.cpp:1769-1799)
	diablo := questFixture().Encode(persist.GameMode{})
	hellfire := questFixture().Encode(persist.GameMode{Hellfire: true})
	// then it is 44 bytes, 48 under Hellfire for the widened _qmsg
	if len(diablo) != persist.QuestSaveSizeDiablo {
		t.Fatalf("diablo len = %d, want %d", len(diablo), persist.QuestSaveSizeDiablo)
	}

	if len(hellfire) != persist.QuestSaveSizeHellfire {
		t.Fatalf("hellfire len = %d, want %d", len(hellfire), persist.QuestSaveSizeHellfire)
	}

	if persist.QuestSaveSize(persist.GameMode{}) != 44 || persist.QuestSaveSize(persist.GameMode{Hellfire: true}) != 48 {
		t.Fatal("QuestSaveSize disagrees with the encoded lengths")
	}
}

func TestQuestSaveFieldLayout(t *testing.T) {
	// when the Diablo record is encoded
	image := questFixture().Encode(persist.GameMode{})
	// then the head bytes match the field order (loadsave.cpp:1770-1773)
	head := []byte{2, 5, 1, 3, 100, 0, 0, 0, 110, 0, 0, 0, 7, 5}
	if !bytes.Equal(image[:len(head)], head) {
		t.Fatalf("head = %x, want %x", image[:len(head)], head)
	}
	// and the return-level block is big-endian before the DoomQuestState skip
	if got := int32(binary.BigEndian.Uint32(image[24:28])); got != 99 {
		t.Fatalf("ReturnLvlPosition.x = %d, want 99", got)
	}

	if got := int32(binary.BigEndian.Uint32(image[36:40])); got != 2 {
		t.Fatalf("ReturnLevelType = %d, want 2", got)
	}

	if !bytes.Equal(image[40:44], []byte{0, 0, 0, 0}) {
		t.Fatalf("DoomQuestState skip = %x", image[40:44])
	}
}

func TestQuestSaveRoundTrip(t *testing.T) {
	// given a quest with a message wider than a byte
	quest := questFixture()
	quest.Msg = 300
	// when the Hellfire record is encoded and decoded back
	image := quest.Encode(persist.GameMode{Hellfire: true})

	decoded, err := persist.DecodeQuestSave(image, persist.GameMode{Hellfire: true})
	// then every field round-trips, including the widened message
	if err != nil {
		t.Fatalf("DecodeQuestSave: %v", err)
	}

	if decoded != quest {
		t.Fatalf("decoded = %+v, want %+v", decoded, quest)
	}
}

func TestQuestSaveDiabloMessageTruncates(t *testing.T) {
	// given a message wider than the Diablo record can carry
	quest := questFixture()
	quest.Msg = 0x12e
	// when the Diablo record is encoded and decoded back
	image := quest.Encode(persist.GameMode{})

	decoded, err := persist.DecodeQuestSave(image, persist.GameMode{})
	// then only the low byte survives (loadsave.cpp:1778)
	if err != nil {
		t.Fatalf("DecodeQuestSave: %v", err)
	}

	if decoded.Msg != 0x2e {
		t.Fatalf("msg = %#x, want 0x2e", decoded.Msg)
	}
}

func questFixture() persist.QuestSave {
	return persist.QuestSave{
		Level:           2,
		Idx:             5,
		Active:          1,
		LvlType:         3,
		X:               100,
		Y:               110,
		Slvl:            7,
		Msg:             0x1e,
		Var1:            4,
		Var2:            6,
		Log:             1,
		ReturnX:         99,
		ReturnY:         109,
		ReturnLevel:     8,
		ReturnLevelType: 2,
	}
}
