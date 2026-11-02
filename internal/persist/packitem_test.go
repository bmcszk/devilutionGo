package persist_test

import (
	"fmt"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/entities"
	"github.com/bmcszk/devilutionGo/internal/diablo/inventory"
	"github.com/bmcszk/devilutionGo/internal/persist"
)

func TestPackItemFromGoldenRows(t *testing.T) {
	for _, row := range loadItemPackGolden(t) {
		t.Run(row.suite+"/"+row.label, func(t *testing.T) {
			// given an item carrying exactly the packable fields of the vector
			// (the golden values are already post-remap, so the remap is identity)
			item := itemFromPack(row.pack)
			// when
			got := persist.PackItem(item, persist.PackEnv{Hellfire: true})
			// then PackItem reproduces the recorded vector byte for byte
			if got != row.pack {
				t.Fatalf("PackItem = %+v, want %+v", got, row.pack)
			}
		})
	}
}

func TestPackItemEmpty(t *testing.T) {
	// given an empty item (pack_test.cpp:829 PackItem_empty)
	item := entities.Item{Itype: int8(inventory.ItemTypeNone)}
	// when
	got := persist.PackItem(item, persist.PackEnv{})
	// then the slot is marked empty
	if got.Idx != persist.PackedItemNone {
		t.Fatalf("Idx = %#x, want %#x", got.Idx, persist.PackedItemNone)
	}
}

func TestPackItemArenaPotion(t *testing.T) {
	// given an arena potion, which does not exist in vanilla (pack.cpp:111-112)
	item := entities.Item{Itype: 3, MiscId: persist.MiscArenaPot, IDidx: 3}
	// when
	got := persist.PackItem(item, persist.PackEnv{})
	// then it is dropped from the save to stay backward compatible
	if got.Idx != persist.PackedItemNone {
		t.Fatalf("Idx = %#x, want %#x", got.Idx, persist.PackedItemNone)
	}
}

func TestPackItemMaxDurClamp(t *testing.T) {
	// given an item whose max durability does not fit a byte (pack.cpp:132-135)
	item := entities.Item{Itype: 3, IDidx: 5, Durability: 400, MaxDur: 300}
	// when
	got := persist.PackItem(item, persist.PackEnv{Hellfire: true})
	// then bMDur saturates at 254 and bDur is clamped to it
	if got.BMDur != 254 || got.BDur != 254 {
		t.Fatalf("BMDur/BDur = %d/%d, want 254/254", got.BMDur, got.BDur)
	}
}

func TestPackItemGoldValue(t *testing.T) {
	// given a gold pile and a non-gold item, both carrying an Ivalue
	gold := entities.Item{Itype: 4, IDidx: persist.ItemIndexGold, Ivalue: 2500}
	other := entities.Item{Itype: 3, IDidx: 5, Ivalue: 2500}
	// when
	gotGold := persist.PackItem(gold, persist.PackEnv{Hellfire: true})
	gotOther := persist.PackItem(other, persist.PackEnv{Hellfire: true})
	// then only gold stores its value in wValue (pack.cpp:145-146)
	if gotGold.Value != 2500 {
		t.Fatalf("gold Value = %d, want 2500", gotGold.Value)
	}

	if gotOther.Value != 0 {
		t.Fatalf("non-gold Value = %d, want 0", gotOther.Value)
	}
}

func TestUnpackItemEar(t *testing.T) {
	// given the ear vector from pack_test.cpp:880
	pack := persist.ItemPack{
		Seed:       1633955154,
		CreateInfo: 17509,
		Idx:        uint16(persist.ItemIndexEar),
		BId:        111,
		BDur:       103,
		BMDur:      117,
		BCh:        101,
		BMCh:       68,
		Value:      19843,
	}
	// when
	item, err := persist.UnpackItem(pack, persist.PackEnv{Hellfire: true})
	// then the hero name and value decode as the C++ test expects
	if err != nil {
		t.Fatalf("UnpackItem: %v", err)
	}

	if got := string(item.IName[:12]); got != "Dead-RogueDM" {
		t.Fatalf("IName = %q, want %q", got, "Dead-RogueDM")
	}

	if item.Ivalue != 3 {
		t.Fatalf("Ivalue = %d, want 3", item.Ivalue)
	}

	if item.Curs != persist.EarCursorSorcerer+2 {
		t.Fatalf("Curs = %d, want %d", item.Curs, persist.EarCursorSorcerer+2)
	}

	if item.IDidx != persist.ItemIndexEar {
		t.Fatalf("IDidx = %d, want %d", item.IDidx, persist.ItemIndexEar)
	}
}

func TestUnpackItemEmpty(t *testing.T) {
	// given the empty vector from pack_test.cpp:818
	pack := persist.ItemPack{Idx: persist.PackedItemNone}
	// when
	item, err := persist.UnpackItem(pack, persist.PackEnv{})
	// then the item decodes as an empty slot
	if err != nil {
		t.Fatalf("UnpackItem: %v", err)
	}

	if !inventory.IsEmpty(item) {
		t.Fatalf("item = %+v, want empty", item)
	}
}

func TestUnpackItemRoundTrip(t *testing.T) {
	for _, row := range loadItemPackGolden(t) {
		t.Run(row.suite+"/"+row.label, func(t *testing.T) {
			// when the recorded vector is unpacked and packed again
			err := repackMatches(row.pack)
			// then the byte image is stable
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestPackItemRemapBranches(t *testing.T) {
	// given a diablo-mode env whose remaps bump the item index by one
	env := persist.PackEnv{
		RemapToDiablo: func(idx int16) int16 { return idx + 1 },
		RemapToSpawn:  func(idx int16) int16 { return idx + 2 },
	}
	item := entities.Item{Itype: 3, IDidx: 5}
	// when
	diablo := persist.PackItem(item, env)
	spawn := persist.PackItem(item, persist.PackEnv{
		Spawn:         true,
		RemapToDiablo: env.RemapToDiablo,
		RemapToSpawn:  env.RemapToSpawn,
	})
	hellfire := persist.PackItem(item, persist.PackEnv{Hellfire: true})
	// then hellfire skips the diablo remap and spawn applies after it
	// (pack.cpp:121-127)
	if diablo.Idx != 6 {
		t.Fatalf("diablo Idx = %d, want 6", diablo.Idx)
	}

	if spawn.Idx != 8 {
		t.Fatalf("spawn Idx = %d, want 8", spawn.Idx)
	}

	if hellfire.Idx != 5 {
		t.Fatalf("hellfire Idx = %d, want 5", hellfire.Idx)
	}
}

// repackMatches checks that UnpackItem followed by PackItem is a no-op.
func repackMatches(pack persist.ItemPack) error {
	item, err := persist.UnpackItem(pack, persist.PackEnv{Hellfire: true})
	if err != nil {
		return fmt.Errorf("unpack: %w", err)
	}

	if got := persist.PackItem(item, persist.PackEnv{Hellfire: true}); got != pack {
		return fmt.Errorf("repack = %+v, want %+v", got, pack)
	}

	return nil
}

// itemFromPack rebuilds the packable fields of an item from a recorded vector.
func itemFromPack(pack persist.ItemPack) entities.Item {
	item := entities.Item{
		Seed:       pack.Seed,
		CreateInfo: pack.CreateInfo,
		IDidx:      int16(pack.Idx),
		Magical:    pack.BId >> 1,
		Identified: pack.BId&1 != 0,
		MaxDur:     int(pack.BMDur),
		Durability: int(pack.BDur),
		Charges:    int(pack.BCh),
		MaxCharges: int(pack.BMCh),
		Ivalue:     int(pack.Value),
		Buff:       pack.Buff,
		Itype:      3,
	}

	if int16(pack.Idx) == persist.ItemIndexEar {
		item.Itype = 3
		item.IName = earNameFromPack(pack)
		item.Curs = persist.EarCursorSorcerer + uint8(pack.Value>>6&3)
	}

	return item
}

// earNameFromPack rebuilds the 16-byte ear hero name (items.cpp:3573-3586).
func earNameFromPack(pack persist.ItemPack) [entities.ItemNameLength]byte {
	var name [entities.ItemNameLength]byte

	name[0] = byte(pack.CreateInfo>>8) & 0x7F
	name[1] = byte(pack.CreateInfo) & 0x7F
	name[2] = byte(pack.Seed>>24) & 0x7F
	name[3] = byte(pack.Seed>>16) & 0x7F
	name[4] = byte(pack.Seed>>8) & 0x7F
	name[5] = byte(pack.Seed) & 0x7F
	name[6] = pack.BId & 0x7F
	name[7] = pack.BDur & 0x7F
	name[8] = pack.BMDur & 0x7F
	name[9] = pack.BCh & 0x7F
	name[10] = pack.BMCh & 0x7F
	name[11] = byte(pack.Value>>8) & 0x7F
	name[12] = byte(pack.Buff>>24) & 0x7F
	name[13] = byte(pack.Buff>>16) & 0x7F
	name[14] = byte(pack.Buff>>8) & 0x7F
	name[15] = byte(pack.Buff) & 0x7F

	return name
}
