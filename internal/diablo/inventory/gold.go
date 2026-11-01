package inventory

import "github.com/bmcszk/devilutionGo/internal/diablo/entities"

// This file ports the gold half of inv.cpp: stack sizing, room accounting,
// topping-off and the cell-filling order used when gold lands in the grid.

// Gold cursor graphics (tables/itemdat.h:116-118).
const (
	CursorGoldSmall  uint8 = 4
	CursorGoldMedium uint8 = 5
	CursorGoldLarge  uint8 = 6
)

// ids of the gold stack and of the gold ring limits (tables/itemdat.h:25,
// items.h:31; the *Limit values live in enums.Gold*Limit).
const idiGold int16 = 0

const (
	goldSmallLimit  = 1000 // GOLD_SMALL_LIMIT (items.h:31)
	goldMediumLimit = 2500 // GOLD_MEDIUM_LIMIT (items.h:31)
)

// GoldCursor ports GetGoldCursor (items.cpp:2975-2984).
// goldFilter decides which gold piles a drop may drain (upstream's
// skipFullStacks switch, player.cpp:329).
type goldFilter func(item entities.Item, maxGold int) bool

func GoldCursor(value int) uint8 {
	if value >= goldMediumLimit {
		return CursorGoldLarge
	}
	if value <= goldSmallLimit {
		return CursorGoldSmall
	}
	return CursorGoldMedium
}

// CalculateGold ports CalculateGold (inv.cpp:2280-2290).
func (v Inv) CalculateGold() int {
	gold := 0
	for i := range *v.numInv {
		if ItemType(v.list[i].Itype) == ItemTypeGold {
			gold += v.list[i].Ivalue
		}
	}
	return gold
}

// RoomForGold ports RoomForGold (inv.cpp:1484-1503) for this player: every
// free cell takes a full stack and every existing pile tops off to the max.
func (v Inv) RoomForGold() int {
	amount := 0
	for _, itemIndex := range v.grid {
		amount += v.cellRoom(itemIndex)
	}
	return amount
}

// cellRoom is the room one InvGrid cell offers: a full stack for a free cell,
// the top-off of a partial gold pile, and nothing for occupied cells.
func (v Inv) cellRoom(itemIndex int8) int {
	maxGold := v.data.maxGold()
	if itemIndex < 0 {
		return 0
	}
	if itemIndex == 0 {
		return maxGold
	}
	goldItem := v.list[itemIndex-1]
	if ItemType(goldItem.Itype) != ItemTypeGold || goldItem.Ivalue == maxGold {
		return 0
	}
	return maxGold - goldItem.Ivalue
}

// AddGold ports AddGoldToInventory without the network sync and cursor
// updates (inv.cpp:1507-1540). It returns the gold that did not fit.
func (v Inv) AddGold(value int) int {
	value = v.topOffPiles(value)
	// The last row, right to left, then the columns bottom to top, right to
	// left (inv.cpp:1529-1538).
	for i := 39; i >= 30 && value > 0; i-- {
		value = v.createGoldInSlot(i, value)
	}
	for x := 9; x >= 0 && value > 0; x-- {
		for y := 2; y >= 0 && value > 0; y-- {
			value = v.createGoldInSlot(InventoryPitch*y+x, value)
		}
	}
	return value
}

// GoldAutoPlace ports GoldAutoPlace (inv.cpp:1543-1551) minus the cursor and
// player-gold bookkeeping: whatever part of the stack fits is moved into the
// grid, and the stack keeps the remainder.
func (v Inv) GoldAutoPlace(goldStack *entities.Item) bool {
	goldStack.Ivalue = v.AddGold(goldStack.Ivalue)
	goldStack.Curs = GoldCursor(goldStack.Ivalue)
	v.player.Gold = v.CalculateGold()
	return goldStack.Ivalue == 0
}

// topOffPiles ports the first loop of AddGoldToInventory (inv.cpp:1509-1527).
func (v Inv) topOffPiles(value int) int {
	maxGold := v.data.maxGold()
	for i := 0; i < *v.numInv && value > 0; i++ {
		goldItem := &v.list[i]
		if ItemType(goldItem.Itype) != ItemTypeGold || goldItem.Ivalue >= maxGold {
			continue
		}
		if goldItem.Ivalue+value > maxGold {
			value -= maxGold - goldItem.Ivalue
			goldItem.Ivalue = maxGold
		} else {
			goldItem.Ivalue += value
			value = 0
		}
		goldItem.Curs = GoldCursor(goldItem.Ivalue)
	}
	return value
}

// createGoldInSlot ports CreateGoldItemInInventorySlot (inv.cpp:1108-1125).
func (v Inv) createGoldInSlot(slotIndex, value int) int {
	if v.grid[slotIndex] != 0 {
		return value
	}
	goldItem := &v.list[*v.numInv]
	*v.numInv++
	v.grid[slotIndex] = int8(*v.numInv)
	fill := value
	if maxGold := v.data.maxGold(); fill > maxGold {
		fill = maxGold
	}
	makeGoldStack(goldItem, fill)
	return value - goldItem.Ivalue
}

// makeGoldStack ports MakeGoldStack (items.cpp:4685-4692) for the fields the
// grid logic reads. InitializeItem/GenerateNewSeed belong to item generation,
// so only the stack identity (IDidx, type, stat flag), value and cursor are
// set here.
func makeGoldStack(goldItem *entities.Item, value int) {
	*goldItem = entities.Item{
		Itype:    int8(ItemTypeGold),
		IDidx:    idiGold,
		StatFlag: true,
		Ivalue:   value,
	}
	goldItem.Curs = GoldCursor(value)
}

// SplitGold ports RemoveGold (control/control_gold.cpp:22-36): the split-off
// amount moves onto the cursor as a fresh stack, the pile keeps the rest and
// is removed when it empties. The cursor redraw and the network sync stay with
// the caller.
func (v Inv) SplitGold(itemIndex, amount int) {
	pile := &v.list[itemIndex]
	pile.Ivalue -= amount
	if pile.Ivalue > 0 {
		pile.Curs = GoldCursor(pile.Ivalue)
	} else {
		v.Remove(itemIndex)
	}
	makeGoldStack(v.hold, amount)
	v.player.Gold = v.CalculateGold()
}

// drainable is the skipFullStacks false rule: every gold pile counts.
func drainable(item entities.Item, _ int) bool {
	return ItemType(item.Itype) == ItemTypeGold
}

// drainableNotFull is the skipFullStacks true rule: piles at the cap are left
// alone (player.cpp:331).
func drainableNotFull(item entities.Item, maxGold int) bool {
	return ItemType(item.Itype) == ItemTypeGold && item.Ivalue != maxGold
}

// DropGold ports DropGold (player.cpp:328-347) with skipFullStacks false: drain
// the requested amount out of the piles in InvList order and spawn each drained
// stack through the DeadItem sink. It returns the gold that did not come out.
func (v Inv) DropGold(amount int) int {
	return v.drainGold(amount, drainable)
}

// DropGoldSkippingFullStacks is DropGold with upstream's skipFullStacks true
// (player.cpp:331): piles already at the cap are left alone.
func (v Inv) DropGoldSkippingFullStacks(amount int) int {
	return v.drainGold(amount, drainableNotFull)
}

// drainGold is the shared body; the filter keeps the bool out of the public API.
func (v Inv) drainGold(amount int, drain goldFilter) int {
	maxGold := v.data.maxGold()
	for i := 0; i < *v.numInv && amount > 0; i++ {
		item := &v.list[i]
		if !drain(*item, maxGold) {
			continue
		}
		if amount < item.Ivalue {
			stack := entities.Item{}
			makeGoldStack(&stack, amount)
			v.spawnDeadItem(stack)
			item.Ivalue -= amount
			return 0
		}
		amount -= item.Ivalue
		v.spawnDeadItem(*item)
		v.Remove(i)
		i = -1 // indices shifted down; restart the scan (player.cpp:346)
	}
	return amount
}

// spawnDeadItem hands a drained stack to the floor-item sink (DeadItem,
// player.cpp:340,344). An unwired sink drops the stack silently.
func (v Inv) spawnDeadItem(item entities.Item) {
	if v.data.DeadItem == nil {
		return
	}
	v.data.DeadItem(item)
}
