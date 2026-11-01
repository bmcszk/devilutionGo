package persist_test

import (
	"encoding/binary"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

func TestSaveItemSizeGolden(t *testing.T) {
	// when an item record is encoded for each mode (loadsave.cpp:2114-2115)
	diablo := itemSaveFixture().Encode(persist.GameMode{})
	hellfire := itemSaveFixture().Encode(persist.GameMode{Hellfire: true})
	// then it is 368 bytes, 372 under Hellfire for _iDamAcFlags (:1259-1260)
	if len(diablo) != 368 {
		t.Fatalf("diablo record len = %d, want 368", len(diablo))
	}

	if len(hellfire) != 372 {
		t.Fatalf("hellfire record len = %d, want 372", len(hellfire))
	}

	if got := persist.SaveItemSize(persist.GameMode{Hellfire: true}); got != 372 {
		t.Fatalf("SaveItemSize(hellfire) = %d", got)
	}
}

func TestSaveItemFieldOffsets(t *testing.T) {
	// when an item record with distinct per-field values is encoded
	image := itemSaveFixture().Encode(persist.GameMode{})
	// then each written field lands at the offset the C++ field list implies
	// (loadsave.cpp:1176-1258)
	cases := []struct {
		name   string
		offset int
		want   int32
	}{
		{"seed", 0, 0x11223344},
		{"type", 8, 1},
		{"position.x", 12, 100},
		{"position.y", 16, 110},
		{"animFlag", 20, 1},
		{"numberOfFrames", 28, 8},
		{"currentFrame+1", 32, 3},
		{"ItemAnimWidth", 36, persist.SaveItemAnimWidth},
		{"spriteTileCenterX", 40, persist.SaveItemSpriteTileCenter},
		{"postDraw", 52, 1},
		{"identified", 56, 1},
		{"curs", 192, 4},
		{"ivalue", 200, 500},
		{"maxDam", 208, 6},
		{"ac", 212, -5},
		{"flags", 216, 0x00040005},
		{"miscId", 220, 10},
		{"maxDur", 240, 20},
		{"plDam", 244, 5},
		{"plMR", 280, 14},
		{"plMana", 284, 15},
		{"plLight", 300, -1},
		{"uniqueItemMapping", 308, 0},
		{"plEnAc", 328, 9},
		{"vMult2", 348, 24},
		{"statFlag", 356, 1},
		{"idIdx", 360, 55},
		{"dwBuff", 364, -0x80000000},
	}
	for _, tc := range cases {
		got := int32(binary.LittleEndian.Uint32(image[tc.offset : tc.offset+4]))
		if got != tc.want {
			t.Fatalf("%s at %d = %d, want %d", tc.name, tc.offset, got, tc.want)
		}
	}
}

func TestSaveItemByteFields(t *testing.T) {
	// when the record is encoded
	image := itemSaveFixture().Encode(persist.GameMode{})
	// then the int8 fields keep their signed byte, the names land at 61/125,
	// and the alignment gaps stay zero (loadsave.cpp:1195-1210)
	cases := []struct {
		name   string
		offset int
		want   uint8
	}{
		{"magical", 60, 0xff},
		{"name[0]", 61, 'A'},
		{"name[1]", 62, 'B'},
		{"name end padding", 124, 0},
		{"iname[0]", 125, 'X'},
		{"loc", 189, 2},
		{"class", 190, 3},
		{"alignment after class", 191, 0},
		{"splLvlAdd", 304, 0xff},
		{"request", 305, 1},
		{"prePower", 332, 0xff},
		{"sufPower", 333, 0xfe},
		{"minStr", 352, 25},
		{"minMag", 353, 26},
		{"minDex", 354, 27},
	}
	for _, tc := range cases {
		if image[tc.offset] != tc.want {
			t.Fatalf("%s at %d = %#x, want %#x", tc.name, tc.offset, image[tc.offset], tc.want)
		}
	}
}

func TestSaveItemHellfireTail(t *testing.T) {
	// when the Hellfire record is encoded
	image := itemSaveFixture().Encode(persist.GameMode{Hellfire: true})
	// then _iDamAcFlags is the last field (loadsave.cpp:1259-1260)
	if got := binary.LittleEndian.Uint32(image[368:372]); got != 0x01020304 {
		t.Fatalf("damAcFlags = %#x, want 0x01020304", got)
	}
}

func itemSaveFixture() persist.ItemSave {
	item := itemSaveIdentity()
	applyItemSaveCombat(&item)
	applyItemSaveBonuses(&item)
	item.Name[0], item.Name[1] = 'A', 'B'
	item.IName[0] = 'X'

	return item
}

func itemSaveIdentity() persist.ItemSave {
	return persist.ItemSave{
		Seed: 0x11223344, CreateInfo: -2, Type: 1, PosX: 100, PosY: 110, AnimFlag: 1,
		NumFrames: 8, CurFrame: 2, Selection: 3, PostDraw: 1, Identified: 1, Magical: -1,
		Loc: 2, Class: 3, Curs: 4, Value: 500, IValue: 500, MinDam: 1, MaxDam: 6, AC: -5,
		Flags: 0x00040005, MiscID: 10, Spell: -1, Durability: 20, MaxDur: 20,
	}
}

func applyItemSaveCombat(item *persist.ItemSave) {
	item.PLDam = 5
	item.PLToHit = 6
	item.PLAC = 7
	item.PLStr = 8
	item.PLMag = 9
	item.PLDex = 10
	item.PLVit = 11
	item.PLFR = 12
	item.PLLR = 13
	item.PLMR = 14
	item.PLMana = 15
	item.PLHP = 16
	item.PLDamMod = 17
	item.PLGetHit = 18
	item.PLLight = -1
}

func applyItemSaveBonuses(item *persist.ItemSave) {
	item.SplLvlAdd = -1
	item.Request = 1
	item.FMinDam = 2
	item.FMaxDam = 3
	item.LMinDam = 4
	item.LMaxDam = 5
	item.PLEnAc = 9
	item.PrePower = -1
	item.SufPower = -2
	item.VAdd1 = 21
	item.VMult1 = 22
	item.VAdd2 = 23
	item.VMult2 = 24
	item.MinStr = 25
	item.MinMag = 26
	item.MinDex = 27
	item.StatFlag = 1
	item.IDIdx = 55
	item.DwBuff = 0x80000000
	item.DamAcFlags = 0x01020304
}
