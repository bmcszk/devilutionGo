package inventory

import "github.com/bmcszk/devilutionGo/internal/diablo/entities"

// This file ports the grid half of inv.cpp: cell-fit checks, the per-size
// search orders, InvGrid writing and auto-placement.

// FitsInSlot ports CheckItemFitsInInventorySlot (inv.cpp:660-680): the item's
// footprint must stay inside the 10-wide grid and every covered cell must be
// free, or hold the item we plan to remove (itemIndexToIgnore, 0-based).
func (v Inv) FitsInSlot(slotIndex int, itemSize Size, itemIndexToIgnore int) bool {
	xx := 0
	if slotIndex > 0 {
		xx = slotIndex % InventoryPitch
	}
	yy := InventoryPitch * (slotIndex / InventoryPitch)
	for range itemSize.Height {
		if yy >= entities.InventoryGridCells {
			return false
		}
		if !v.rowFits(xx, yy, itemSize.Width, itemIndexToIgnore) {
			return false
		}
		yy += InventoryPitch
	}
	return true
}

// rowFits checks one row of the footprint: the cells must be in range and
// either free or occupied by the item assumed removed (inv.cpp:667-676).
func (v Inv) rowFits(xx, yy, width, itemIndexToIgnore int) bool {
	for i := range width {
		if xx+i >= InventoryPitch {
			// The item is too wide to fit in the specified column.
			return false
		}
		cell := v.grid[xx+i+yy]
		if cell != 0 && abs8(cell)-1 != itemIndexToIgnore {
			return false
		}
	}
	return true
}

// FindSlot ports FindSlotForItem (inv.cpp:690-735): the first grid index that
// fits an item of the given size, in the upstream search order of its size
// class (1x1, height 2, 1x3, 2x3).
func (v Inv) FindSlot(itemSize Size, itemIndexToIgnore int) (int, bool) {
	switch {
	case itemSize.Height == 1:
		return v.findRowOne(itemSize, itemIndexToIgnore)
	case itemSize.Height == 2:
		return v.findHeightTwo(itemSize, itemIndexToIgnore)
	case itemSize == (Size{Width: 1, Height: 3}):
		return v.findInOrder(itemSize, itemIndexToIgnore, 0, 20)
	case itemSize == (Size{Width: 2, Height: 3}):
		if slot, ok := v.findInOrder(itemSize, itemIndexToIgnore, 0, 9); ok {
			return slot, ok
		}
		return v.findInOrder(itemSize, itemIndexToIgnore, 10, 19)
	default:
		return 0, false
	}
}

// findRowOne ports the 1x1 order: the bottom row first, then the remaining
// grid right to left, bottom to top (inv.cpp:691-703).
func (v Inv) findRowOne(itemSize Size, itemIndexToIgnore int) (int, bool) {
	if slot, ok := v.findInOrder(itemSize, itemIndexToIgnore, 30, 40); ok {
		return slot, ok
	}
	for x := 9; x >= 0; x-- {
		for y := 2; y >= 0; y-- {
			if v.FitsInSlot(InventoryPitch*y+x, itemSize, itemIndexToIgnore) {
				return InventoryPitch*y + x, true
			}
		}
	}
	return 0, false
}

// findHeightTwo ports the height-2 order: columns right to left, rows top to
// bottom (inv.cpp:706-714).
func (v Inv) findHeightTwo(itemSize Size, itemIndexToIgnore int) (int, bool) {
	for i := range InventoryPitch - itemSize.Width + 1 {
		x := InventoryPitch - itemSize.Width - i
		for y := range 3 {
			if v.FitsInSlot(InventoryPitch*y+x, itemSize, itemIndexToIgnore) {
				return InventoryPitch*y + x, true
			}
		}
	}
	return 0, false
}

// findInOrder returns the first index in [from, to) that fits the item
// (inv.cpp:717-733).
func (v Inv) findInOrder(itemSize Size, itemIndexToIgnore, from, to int) (int, bool) {
	for i := from; i < to; i++ {
		if v.FitsInSlot(i, itemSize, itemIndexToIgnore) {
			return i, true
		}
	}
	return 0, false
}

// CanFit ports CanFitItemInInventory (inv.cpp:1465-1468).
func (v Inv) CanFit(item entities.Item) bool {
	_, ok := v.FindSlot(v.sizeOf(item), -1)
	return ok
}

// CouldFit ports CouldFitItemInInventory (inv.cpp:743-746): whether the item
// would fit with inventory item `ignore` (0-based) assumed removed.
func (v Inv) CouldFit(item entities.Item, ignore int) bool {
	_, ok := v.FindSlot(v.sizeOf(item), ignore)
	return ok
}

// addToGrid ports AddItemToInvGrid (inv.cpp:151-172). invListIndex is the
// 1-based InvList index; the last row's first cell keeps the positive index,
// every other covered cell stores it negated as an occupancy placeholder.
func (v Inv) addToGrid(invGridIndex, invListIndex int, itemSize Size) {
	for y := range itemSize.Height {
		rowGridIndex := invGridIndex + InventoryPitch*y
		for x := range itemSize.Width {
			if x == 0 && y == itemSize.Height-1 {
				v.grid[rowGridIndex+x] = int8(invListIndex)
			} else {
				v.grid[rowGridIndex+x] = int8(-invListIndex)
			}
		}
	}
}

// Place ports AutoPlaceItemInInventory (inv.cpp:1470-1487): find a free slot
// and append the item to InvList. It reports false when the grid is full.
func (v Inv) Place(item entities.Item) bool {
	itemSize := v.sizeOf(item)
	slot, ok := v.FindSlot(itemSize, -1)
	if !ok {
		return false
	}
	v.list[*v.numInv] = item
	*v.numInv++
	v.addToGrid(slot, *v.numInv, itemSize)
	return true
}

// abs8 is std::abs for the int8_t InvGrid cells.
func abs8(v int8) int {
	if v < 0 {
		return -int(v)
	}
	return int(v)
}
