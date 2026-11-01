package inventory

// This file ports the cursor hold-item machine: pasting the item under the
// cursor into an inventory cell or a belt slot, including the gold merge path
// and the one-item overlap rule.

// PutInGrid ports ChangeInvItem (inv.cpp:421-522). slotIndex is the grid cell
// (upstream's panel slot minus SLOTXY_INV_FIRST). It reports false when the
// footprint overlaps two different items.
func (v Inv) PutInGrid(slotIndex int, itemSize Size) bool {
	prevItemID := v.prevItemID(slotIndex, itemSize)
	if prevItemID < 0 {
		return false
	}
	if ItemType(v.hold.Itype) == ItemTypeGold && prevItemID == 0 {
		v.pasteGold(slotIndex)
		return true
	}
	v.pasteItem(slotIndex, itemSize, prevItemID)
	return true
}

// PutInBelt ports ChangeBeltItem (inv.cpp:524-544). slot is the belt index
// (upstream's panel slot minus SLOTXY_BELT_FIRST).
func (v Inv) PutInBelt(slot int) {
	if IsEmpty(v.belt[slot]) {
		v.belt[slot] = pop(v.hold)
		return
	}
	swapped := v.belt[slot]
	v.belt[slot] = *v.hold
	*v.hold = swapped
	if ItemType(v.hold.Itype) == ItemTypeGold {
		v.player.Gold = v.CalculateGold()
	}
}

// pasteGold ports the gold branch of ChangeInvItem (inv.cpp:435-459): top off
// the pile under the cursor, or start a new pile in the empty cell.
func (v Inv) pasteGold(slotIndex int) {
	if v.grid[slotIndex] > 0 {
		v.mergeGoldInto(int(v.grid[slotIndex]) - 1)
		return
	}
	invIndex := *v.numInv
	v.player.Gold += v.hold.Ivalue
	v.list[invIndex] = pop(v.hold)
	*v.numInv++
	v.grid[slotIndex] = int8(*v.numInv)
}

// mergeGoldInto ports the in-cell merge: a full hold empties the cursor, an
// overflow leaves the remainder on it (inv.cpp:438-452).
func (v Inv) mergeGoldInto(invIndex int) {
	maxGold := v.data.maxGold()
	pile := &v.list[invIndex]
	combined := v.hold.Ivalue + pile.Ivalue
	if combined <= maxGold {
		pile.Ivalue = combined
		pile.Curs = GoldCursor(combined)
		v.player.Gold += v.hold.Ivalue
		clearItem(v.hold)
		return
	}
	fitting := maxGold - pile.Ivalue
	v.player.Gold += fitting
	v.hold.Ivalue -= fitting
	v.hold.Curs = GoldCursor(v.hold.Ivalue)
	pile.Ivalue = maxGold
	pile.Curs = CursorGoldLarge
}

// pasteItem ports the non-gold branch of ChangeInvItem (inv.cpp:481-519):
// append to InvList when the cell is free, otherwise swap with the item that
// is there and rewrite its grid footprint.
func (v Inv) pasteItem(slotIndex int, itemSize Size, prevItemID int) {
	if prevItemID == 0 {
		v.list[*v.numInv] = pop(v.hold)
		*v.numInv++
		prevItemID = *v.numInv
	} else {
		invIndex := prevItemID - 1
		if ItemType(v.hold.Itype) == ItemTypeGold {
			v.player.Gold += v.hold.Ivalue
		}
		swapped := v.list[invIndex]
		v.list[invIndex] = *v.hold
		*v.hold = swapped
		if ItemType(v.hold.Itype) == ItemTypeGold {
			v.player.Gold = v.CalculateGold()
		}
		v.clearPlaceholders(prevItemID)
	}
	v.addToGrid(slotIndex, prevItemID, itemSize)
}

// clearPlaceholders drops every grid reference to the swapped-out item
// (inv.cpp:513-520).
func (v Inv) clearPlaceholders(itemID int) {
	for i := range v.grid {
		if v.grid[i] == int8(itemID) || v.grid[i] == int8(-itemID) {
			v.grid[i] = 0
		}
	}
}

// prevItemID ports GetPrevItemId (inv.cpp:440-452).
func (v Inv) prevItemID(slotIndex int, itemSize Size) int {
	if ItemType(v.hold.Itype) != ItemTypeGold {
		return v.checkOverlappingItems(slotIndex, itemSize)
	}
	first := v.grid[slotIndex]
	switch {
	case first == 0:
		return 0
	case first < 0:
		return -int(first)
	case ItemType(v.list[first-1].Itype) != ItemTypeGold:
		return int(first)
	default:
		return 0
	}
}

// checkOverlappingItems ports CheckOverlappingItems (inv.cpp:421-449): the
// footprint may cover empty cells plus, at most, one item.
func (v Inv) checkOverlappingItems(slotIndex int, itemSize Size) int {
	overlappingID := int8(0)
	for row := range itemSize.Height {
		for col := range itemSize.Width {
			next, ok := mergeItemID(overlappingID, v.grid[slotIndex+InventoryPitch*row+col])
			if !ok {
				// Two different items would be displaced: reject the paste.
				return -1
			}
			overlappingID = next
		}
	}
	return int(overlappingID)
}

// mergeItemID folds one grid cell into the running overlap id. It reports
// false when the cell belongs to a second, different item.
func mergeItemID(running, cell int8) (int8, bool) {
	if cell == 0 {
		return running, true
	}
	cellID := int8(abs8(cell))
	if running != 0 && running != cellID {
		return running, false
	}
	return cellID, true
}
