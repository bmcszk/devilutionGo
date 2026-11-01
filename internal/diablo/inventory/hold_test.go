package inventory_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/inventory"
)

func TestPutInGridEmptySlot(t *testing.T) {
	// given a held sword and an empty grid
	p, inv := newInv(t)
	p.HoldItem = newItem(cursSword)
	// when
	ok := inv.PutInGrid(0, inventory.Size{Width: 1, Height: 3})
	// then
	if !ok {
		t.Fatal("PutInGrid = false, want true")
	}
	if !inventory.IsEmpty(p.HoldItem) {
		t.Fatal("the held item must move out of the cursor")
	}
	if inventory.IsEmpty(p.InvList[0]) {
		t.Fatal("InvList[0] should hold the sword")
	}
	if p.InvGrid[0] != -1 || p.InvGrid[20] != 1 {
		t.Fatalf("placeholders = (%d, %d), want (-1, 1)", p.InvGrid[0], p.InvGrid[20])
	}
}

func TestPutInGridSwaps(t *testing.T) {
	// given a 1x1 item in the grid and a different item held
	p, inv := newInv(t)
	inv.Place(newItem(cursConsumable))
	p.HoldItem = newItem(cursArmor)
	// when
	ok := inv.PutInGrid(30, inventory.Size{Width: 1, Height: 1})
	// then
	if !ok {
		t.Fatal("PutInGrid = false, want true")
	}
	if p.HoldItem.Curs != cursConsumable {
		t.Fatalf("held curs = %d, want the swapped-out %d", p.HoldItem.Curs, cursConsumable)
	}
	if p.InvList[0].Curs != cursArmor {
		t.Fatalf("InvList[0].Curs = %d, want %d", p.InvList[0].Curs, cursArmor)
	}
	if p.InvGrid[30] != 1 {
		t.Fatalf("InvGrid[30] = %d, want 1", p.InvGrid[30])
	}
}

func TestPutInGridRejectsTwoItems(t *testing.T) {
	// given a 2x3 footprint straddling two different items
	p, inv := newInv(t)
	inv.Place(newItem(cursConsumable)) // lands on 30
	p.InvGrid[31] = -2
	p.InvList[1] = newItem(cursConsumable)
	p.NumInv = 2
	p.HoldItem = newItem(cursBow)
	// when
	ok := inv.PutInGrid(30, inventory.Size{Width: 2, Height: 1})
	// then
	if ok {
		t.Fatal("PutInGrid over two items = true, want false")
	}
	if p.HoldItem.Curs != cursBow {
		t.Fatal("a rejected paste must keep the item on the cursor")
	}
}

func TestGoldMergeIntoCell(t *testing.T) {
	// given a gold pile in the grid and more gold held
	p, inv := newInv(t)
	stack(p, 0, 4000)
	p.InvGrid[30] = 1
	p.NumInv = 1
	held := newGold()
	held.Ivalue = 500
	p.HoldItem = held
	// when
	ok := inv.PutInGrid(30, inventory.Size{Width: 1, Height: 1})
	// then
	if !ok {
		t.Fatal("PutInGrid(gold) = false, want true")
	}
	if p.InvList[0].Ivalue != 4500 {
		t.Fatalf("pile = %d, want 4500", p.InvList[0].Ivalue)
	}
	if !inventory.IsEmpty(p.HoldItem) {
		t.Fatal("held gold must be cleared when it all fits")
	}
	if p.Gold != 500 {
		t.Fatalf("player gold = %d, want 500", p.Gold)
	}
}

func TestGoldMergeOverflow(t *testing.T) {
	// given a nearly full pile and a held stack that overflows it
	p, inv := newInv(t)
	stack(p, 0, 4900)
	p.InvGrid[30] = 1
	p.NumInv = 1
	held := newGold()
	held.Ivalue = 900
	p.HoldItem = held
	// when
	inv.PutInGrid(30, inventory.Size{Width: 1, Height: 1})
	// then
	if p.InvList[0].Ivalue != inventory.MaxGoldLimit {
		t.Fatalf("pile = %d, want %d", p.InvList[0].Ivalue, inventory.MaxGoldLimit)
	}
	if p.InvList[0].Curs != inventory.CursorGoldLarge {
		t.Fatalf("pile cursor = %d, want the large gold cursor", p.InvList[0].Curs)
	}
	if p.HoldItem.Ivalue != 800 {
		t.Fatalf("held = %d, want the 800 remainder", p.HoldItem.Ivalue)
	}
	if p.Gold != 100 {
		t.Fatalf("player gold = %d, want the 100 that fit", p.Gold)
	}
}

func TestGoldIntoEmptyCell(t *testing.T) {
	// given held gold and an empty cell
	p, inv := newInv(t)
	held := newGold()
	held.Ivalue = 300
	p.HoldItem = held
	// when
	ok := inv.PutInGrid(30, inventory.Size{Width: 1, Height: 1})
	// then
	if !ok {
		t.Fatal("PutInGrid(gold, empty) = false, want true")
	}
	if p.InvGrid[30] != 1 {
		t.Fatalf("InvGrid[30] = %d, want 1", p.InvGrid[30])
	}
	if p.InvList[0].Ivalue != 300 {
		t.Fatalf("InvList[0].Ivalue = %d, want 300", p.InvList[0].Ivalue)
	}
	if p.Gold != 300 {
		t.Fatalf("player gold = %d, want 300", p.Gold)
	}
}

func TestPutInBeltDropsHeldItem(t *testing.T) {
	// given an empty belt slot and a held potion
	p, inv := newInv(t)
	p.HoldItem = newItem(cursConsumable)
	// when
	inv.PutInBelt(1)
	// then
	if inventory.IsEmpty(p.SpdList[1]) {
		t.Fatal("SpdList[1] should hold the potion")
	}
	if !inventory.IsEmpty(p.HoldItem) {
		t.Fatal("the cursor should be empty after the drop")
	}
}

func TestPutInBeltSwapsHeldGold(t *testing.T) {
	// given gold in a belt slot and a potion held
	p, inv := newInv(t)
	stack(p, 0, 200)
	p.SpdList[2] = p.InvList[0]
	p.InvList[0] = inventory.EmptyItem()
	p.HoldItem = newItem(cursConsumable)
	// when
	inv.PutInBelt(2)
	// then
	if p.HoldItem.Ivalue != 200 {
		t.Fatalf("held = %+v, want the swapped-out gold", p.HoldItem)
	}
	if p.SpdList[2].Curs != cursConsumable {
		t.Fatalf("SpdList[2].Curs = %d, want %d", p.SpdList[2].Curs, cursConsumable)
	}
	if p.Gold != 0 {
		t.Fatalf("player gold = %d, want 0 (no gold left in the inventory)", p.Gold)
	}
}
