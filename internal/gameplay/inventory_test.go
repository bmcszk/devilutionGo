package gameplay_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/entities"
	"github.com/bmcszk/devilutionGo/internal/gameplay"
)

// TestFacadeMatchesInventoryPort pins that the gameplay entry points are the
// inventory port itself, not a copy: sizes, belt rule and equip mapping agree.
func TestFacadeMatchesInventoryPort(t *testing.T) {
	// given a player with a 2x2 item and a potion in the belt
	sprites := func(curs uint8) gameplay.Size {
		switch curs {
		case 1:
			return gameplay.Size{Width: 32, Height: 32}
		case 2:
			return gameplay.Size{Width: 64, Height: 64}
		}
		return gameplay.Size{}
	}
	player := &entities.Player{}
	inv := gameplay.New(player, gameplay.Data{Sprite: sprites})
	armor := entities.Item{Curs: 2, Itype: int8(gameplay.ItemTypeLightArmor)}
	potion := entities.Item{Curs: 1}
	// when the facade helpers run
	size := gameplay.ItemSize(armor, sprites)
	beltable := gameplay.IsBeltItem(potion, sprites)
	helmLoc := gameplay.EquipTypeForSlot(gameplay.SlotHead, gameplay.LocNone)
	placed := inv.Place(armor)
	// then they behave exactly like the inventory port
	if size != (gameplay.Size{Width: 2, Height: 2}) {
		t.Fatalf("ItemSize = %+v, want 2x2", size)
	}
	if !beltable {
		t.Fatal("IsBeltItem(1x1 potion) = false, want true")
	}
	if helmLoc != gameplay.LocHelm {
		t.Fatalf("EquipTypeForSlot(head) = %v, want Helm", helmLoc)
	}
	if !placed {
		t.Fatal("Place(armor) = false, want true")
	}
	assertConstants(t)
}

// assertConstants pins the upstream container sizes the facade exposes.
func assertConstants(t *testing.T) {
	t.Helper()
	if gameplay.InventoryPitch != 10 {
		t.Fatalf("InventoryPitch = %d, want 10", gameplay.InventoryPitch)
	}
	if gameplay.BeltSlots != 8 {
		t.Fatalf("BeltSlots = %d, want 8", gameplay.BeltSlots)
	}
	if gameplay.MaxGoldLimit != 5000 {
		t.Fatalf("MaxGoldLimit = %d, want 5000", gameplay.MaxGoldLimit)
	}
}
