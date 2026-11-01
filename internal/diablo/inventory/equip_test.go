package inventory_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/entities"
	"github.com/bmcszk/devilutionGo/internal/diablo/inventory"
)

// equipCase is the body-slot setup one CanEquipAt row needs.
type equipCase struct {
	set      func(p *playerFixture)
	headBusy bool
}

func TestIsEquipment(t *testing.T) {
	tests := []struct {
		name string
		item entities.Item
		want bool
	}{
		{name: "helm", item: equip(inventory.LocHelm), want: true},
		{name: "armor", item: equip(inventory.LocArmor), want: true},
		{name: "two-handed sword", item: equip(inventory.LocTwoHand), want: true},
		{name: "belt item", item: equip(inventory.LocBelt), want: false},
		{name: "empty slot", item: inventory.EmptyItem(), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// when
			got := inventory.IsEquipment(tt.item)
			// then
			if got != tt.want {
				t.Fatalf("IsEquipment = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCanEquipItem(t *testing.T) {
	tests := []struct {
		name string
		item entities.Item
		want bool
	}{
		{name: "satisfied requirements", item: equip(inventory.LocHelm), want: true},
		{name: "stat flag not set", item: withStatFlag(equip(inventory.LocHelm), false), want: false},
		{name: "not equipment", item: equip(inventory.LocBelt), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// when
			got := inventory.CanEquip(tt.item)
			// then
			if got != tt.want {
				t.Fatalf("CanEquip = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCanEquipAt(t *testing.T) {
	tests := []struct {
		name     string
		item     entities.Item
		loc      inventory.BodyLoc
		headBusy bool
		set      func(p *playerFixture)
		want     bool
	}{
		{name: "helm into the head slot", item: equip(inventory.LocHelm), loc: inventory.BodyHead, want: true},
		{name: "helm into the chest slot", item: equip(inventory.LocHelm), loc: inventory.BodyChest, want: false},
		{name: "occupied slot", item: equip(inventory.LocHelm), loc: inventory.BodyHead, headBusy: true, want: false},
		{name: "ring into a ring slot", item: equip(inventory.LocRing), loc: inventory.BodyRingLeft, want: true},
		{name: "amulet into the amulet slot", item: equip(inventory.LocAmulet), loc: inventory.BodyAmulet, want: true},
		{name: "two-hander into an empty pair", item: equip(inventory.LocTwoHand), loc: inventory.BodyHandLeft, want: true},
		{
			name: "two-hander with one hand busy",
			item: equip(inventory.LocTwoHand),
			loc:  inventory.BodyHandLeft,
			set:  func(p *playerFixture) { p.InvBody[inventory.BodyHandRight] = equip(inventory.LocOneHand) },
			want: false,
		},
		{
			name: "second one-hander of another class",
			item: withClass(equip(inventory.LocOneHand), 1),
			loc:  inventory.BodyHandLeft,
			set: func(p *playerFixture) {
				p.InvBody[inventory.BodyHandRight] = withClass(equip(inventory.LocOneHand), 2)
			},
			want: true,
		},
		{
			name: "second one-hander of the same class",
			item: withClass(equip(inventory.LocOneHand), 1),
			loc:  inventory.BodyHandLeft,
			set: func(p *playerFixture) {
				p.InvBody[inventory.BodyHandRight] = withClass(equip(inventory.LocOneHand), 1)
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			_, inv := prepareEquipAt(t, equipCase{set: tt.set, headBusy: tt.headBusy})
			// when
			got := inv.CanEquipAt(tt.item, tt.loc)
			// then
			if got != tt.want {
				t.Fatalf("CanEquipAt(%v) = %v, want %v", tt.loc, got, tt.want)
			}
		})
	}
}

func TestCanWieldDualWieldClass(t *testing.T) {
	// given a Bard (dual-wield class) holding a one-handed sword in the right
	// hand
	p, _ := newInv(t)
	inv := inventory.New(p, inventory.Data{
		Sprite:    testSprites,
		Usable:    func(int16) bool { return true },
		DualWield: func(uint8) bool { return true },
		MaxGold:   inventory.MaxGoldLimit,
	})
	sword := withType(equip(inventory.LocOneHand), inventory.ItemTypeSword)
	p.InvBody[inventory.BodyHandRight] = sword
	// when a one-handed mace is checked against the free hand
	mace := withType(equip(inventory.LocOneHand), inventory.ItemTypeMace)
	got := inv.CanWield(mace)
	// then
	if !got {
		t.Fatal("a dual-wield class may hold a second one-hander")
	}
}

func TestAutoEquip(t *testing.T) {
	// given a helm and a player with empty body slots
	p, inv := newInv(t)
	helm := equip(inventory.LocHelm)
	// when
	ok := inv.AutoEquip(helm)
	// then
	if !ok {
		t.Fatal("AutoEquip(helm) = false, want true")
	}
	if inventory.IsEmpty(p.InvBody[inventory.BodyHead]) {
		t.Fatal("the helm should be in the head slot")
	}
	if !inventory.IsEmpty(p.InvBody[inventory.BodyChest]) {
		t.Fatal("no other body slot may be touched")
	}
}

func TestAutoEquipNonEquipment(t *testing.T) {
	// given a belt item
	_, inv := newInv(t)
	// when
	ok := inv.AutoEquip(equip(inventory.LocBelt))
	// then
	if ok {
		t.Fatal("AutoEquip(belt item) = true, want false")
	}
}

// prepareEquipAt seeds the body slots a CanEquipAt case needs.

func TestEquipTypeForSlot(t *testing.T) {
	tests := []struct {
		name    string
		slot    int
		desired inventory.ItemLoc
		want    inventory.ItemLoc
	}{
		{name: "head", slot: inventory.SlotHead, want: inventory.LocHelm},
		{name: "left ring", slot: inventory.SlotRingLeft, want: inventory.LocRing},
		{name: "right ring", slot: inventory.SlotRingRight, want: inventory.LocRing},
		{name: "amulet", slot: inventory.SlotAmulet, want: inventory.LocAmulet},
		{
			name:    "two-handed hand",
			slot:    inventory.SlotHandLeft,
			desired: inventory.LocTwoHand,
			want:    inventory.LocTwoHand,
		},
		{name: "one-handed hand", slot: inventory.SlotHandRight, want: inventory.LocOneHand},
		{name: "chest", slot: inventory.SlotChest, want: inventory.LocArmor},
		{name: "belt first", slot: int(inventory.BeltItemFirst), want: inventory.LocBelt},
		{name: "belt last", slot: int(inventory.BeltItemLast), want: inventory.LocBelt},
		{
			name: "grid slot is not a body slot",
			slot: int(inventory.InvItemFirst),
			want: inventory.LocUnequippable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// when
			got := inventory.EquipTypeForSlot(tt.slot, tt.desired)
			// then
			if got != tt.want {
				t.Fatalf("EquipTypeForSlot(%d, %v) = %v, want %v", tt.slot, tt.desired, got, tt.want)
			}
		})
	}
}

func prepareEquipAt(t *testing.T, c equipCase) (*playerFixture, inventory.Inv) {
	t.Helper()
	p, inv := newInv(t)
	if c.set != nil {
		c.set(p)
	}
	if c.headBusy {
		p.InvBody[inventory.BodyHead] = equip(inventory.LocHelm)
	}
	return p, inv
}

// equip is an equippable item with its stat flag satisfied.
func equip(loc inventory.ItemLoc) entities.Item {
	item := newItem(cursArmor)
	item.Loc = int8(loc)
	item.StatFlag = true
	return item
}

// withStatFlag flips the stat flag on an item.
func withStatFlag(item entities.Item, flag bool) entities.Item {
	item.StatFlag = flag
	return item
}

// withType sets the item's ItemType.
func withType(item entities.Item, itype inventory.ItemType) entities.Item {
	item.Itype = int8(itype)
	return item
}

// withClass sets the item's class, which the one-hand compatibility rule reads.
func withClass(item entities.Item, class uint8) entities.Item {
	item.Class = class
	return item
}
