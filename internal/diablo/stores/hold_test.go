package stores_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/entities"
	"github.com/bmcszk/devilutionGo/internal/diablo/inventory"
	"github.com/bmcszk/devilutionGo/internal/diablo/stores"
)

// TestSmithSellHolds pins Griswold's buy-from-player list: inventory then belt,
// every price at a quarter or one gold, and the belt's -(i+1) index
// (stores.cpp:543-604, 1433-1435).
func TestSmithSellHolds(t *testing.T) {
	t.Parallel()

	// given a player holding a sword, a potion and a staff, plus a belt sword
	fx := fixture()
	fx.putInv(0, sword())
	fx.putInv(1, entities.Item{Itype: int8(inventory.ItemTypeMisc)})
	fx.putInv(2, entities.Item{Itype: int8(inventory.ItemTypeStaff)})
	fx.putBelt(0, entities.Item{Itype: int8(inventory.ItemTypeSword), IDidx: 1, Value: 400})

	// when the smith's offer list is built
	holds := stores.SmithSellHolds(fx.player, stores.Env{})

	// then only the two swords are offered, priced by SellValue
	if len(holds) != 2 {
		t.Fatalf("got %d holds, want 2", len(holds))
	}

	if holds[0].Index != 0 || holds[0].Item.Value != 250 || holds[0].Item.Ivalue != 250 {
		t.Fatalf("inventory sword = %+v, want index 0 price 250", holds[0])
	}

	if holds[1].Index != -1 || holds[1].Item.Value != 100 {
		t.Fatalf("belt sword = %+v, want index -1 price 100", holds[1])
	}
}

// TestWitchSellHolds pins Adria's list: misc sells, a potion with the 30..40
// misc id does not (stores.cpp:2888-2908).
func TestWitchSellHolds(t *testing.T) {
	t.Parallel()

	fx := fixture()
	fx.putInv(0, entities.Item{Itype: int8(inventory.ItemTypeMisc), Value: 40})
	fx.putInv(1, entities.Item{Itype: int8(inventory.ItemTypeMisc), MiscId: 30, Value: 40})
	fx.putInv(2, entities.Item{Itype: int8(inventory.ItemTypeStaff), IDidx: 1, Value: 800})

	// when the witch is offered the pack and the belt
	holds := stores.WitchSellHolds(fx.player, stores.Env{})

	// then only her kinds are offered
	if len(holds) != 2 {
		t.Fatalf("got %d holds, want 2", len(holds))
	}

	if holds[0].Index != 0 || holds[0].Item.Value != 10 {
		t.Fatalf("potion = %+v, want index 0 price 10", holds[0])
	}

	if holds[1].Index != 2 || holds[1].Item.Value != 200 {
		t.Fatalf("staff = %+v, want index 2 price 200", holds[1])
	}
}

// TestRepairHolds pins the repair list order — head, chest, both hands, then
// the inventory — and the hold ids it carries (stores.cpp:629-664).
func TestRepairHolds(t *testing.T) {
	t.Parallel()

	fx := fixture()
	fx.putBody(inventory.BodyHead, damagedSword())
	fx.putBody(inventory.BodyChest, entities.Item{
		Itype: int8(inventory.ItemTypeLightArmor), Value: 2000, MaxDur: 100, Durability: 100,
	})
	fx.putBody(inventory.BodyHandLeft, damagedSword())
	fx.putBody(inventory.BodyRingLeft, damagedSword())
	fx.putInv(0, damagedSword())
	fx.putInv(1, sword())

	// when the smith is offered the worn gear
	holds := stores.RepairHolds(fx.player)

	// then the damaged pieces come in the upstream order
	if len(holds) != 3 {
		t.Fatalf("got %d holds, want 3", len(holds))
	}

	wantIndex := []int8{stores.HoldHead, stores.HoldHandLeft, 0}
	wantPrice := []int{250, 250, 250}

	for i := range holds {
		if holds[i].Index != wantIndex[i] {
			t.Fatalf("hold %d index = %d, want %d", i, holds[i].Index, wantIndex[i])
		}

		if holds[i].Item.Ivalue != wantPrice[i] {
			t.Fatalf("hold %d price = %d, want %d", i, holds[i].Item.Ivalue, wantPrice[i])
		}
	}
}

// TestRechargeHolds pins the recharge list: the left hand first, then the
// inventory, priced with the spell's staff cost (stores.cpp:861-886).
func TestRechargeHolds(t *testing.T) {
	t.Parallel()

	fx := fixture()
	fx.putBody(inventory.BodyHandLeft, entities.Item{
		Itype: int8(inventory.ItemTypeStaff), Value: 1000, Charges: 5, MaxCharges: 20,
	})
	fx.putInv(0, entities.Item{
		Itype: int8(inventory.ItemTypeStaff), Value: 1000, Charges: 10, MaxCharges: 20,
	})
	fx.putInv(1, entities.Item{Itype: int8(inventory.ItemTypeStaff), Charges: 20, MaxCharges: 20})

	staffCost := func(int32) int { return 600 }

	// when the witch is offered the staves
	holds := stores.RechargeHolds(fx.player, staffCost)

	// then only the spent ones are offered, left hand first
	if len(holds) != 2 {
		t.Fatalf("got %d holds, want 2", len(holds))
	}

	if holds[0].Index != stores.HoldHandLeft || holds[0].Item.Ivalue != 600 {
		t.Fatalf("left hand = %+v, want index -3 price 600", holds[0])
	}

	if holds[1].Index != 0 || holds[1].Item.Ivalue != 400 {
		t.Fatalf("inventory staff = %+v, want index 0 price 400", holds[1])
	}
}

// TestIdentifyHolds pins Cain's list order and flat fee (stores.cpp:1104-1160,
// 1095-1102).
func TestIdentifyHolds(t *testing.T) {
	t.Parallel()

	magic := func() entities.Item {
		item := sword()
		item.Magical = 1

		return item
	}

	fx := fixture()
	fx.putBody(inventory.BodyHead, magic())
	fx.putBody(inventory.BodyRingLeft, magic())
	fx.putBody(inventory.BodyRingRight, magic())
	identified := magic()
	identified.Identified = true
	fx.putBody(inventory.BodyAmulet, identified)
	fx.putInv(0, magic())

	// when Cain is offered the unidentified magic items
	holds := stores.IdentifyHolds(fx.player)

	// then they come in the upstream order only
	wantIndex := []int8{stores.HoldHead, stores.HoldRingLeft, stores.HoldRingRight, 0}
	if len(holds) != len(wantIndex) {
		t.Fatalf("got %d holds, want %d", len(holds), len(wantIndex))
	}

	for i := range holds {
		if holds[i].Index != wantIndex[i] {
			t.Fatalf("hold %d index = %d, want %d", i, holds[i].Index, wantIndex[i])
		}

		if holds[i].Item.Ivalue != stores.IdentifyPrice || holds[i].Item.Value != stores.IdentifyPrice {
			t.Fatalf("hold %d price = %d/%d, want %d", i, holds[i].Item.Value, holds[i].Item.Ivalue, stores.IdentifyPrice)
		}
	}
}

// TestBodySlotMapping pins the hold id to inv_body_loc mapping (player.h:73-82).
func TestBodySlotMapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		hold int8
		want inventory.BodyLoc
	}{
		{stores.HoldHead, inventory.BodyHead},
		{stores.HoldChest, inventory.BodyChest},
		{stores.HoldHandLeft, inventory.BodyHandLeft},
		{stores.HoldHandRight, inventory.BodyHandRight},
		{stores.HoldRingLeft, inventory.BodyRingLeft},
		{stores.HoldRingRight, inventory.BodyRingRight},
		{stores.HoldAmulet, inventory.BodyAmulet},
	}

	for _, tt := range tests {
		// when the hold id is resolved
		got, ok := stores.BodySlot(tt.hold)

		// then it names the case's body slot
		if !ok || got != tt.want {
			t.Fatalf("BodySlot(%d) = %v, %v; want %v", tt.hold, got, ok, tt.want)
		}
	}

	// when a non-body id is resolved
	_, ok := stores.BodySlot(0)

	// then it reports no body slot
	if ok {
		t.Fatal("BodySlot(0) resolved, want no body slot")
	}
}
