package inventory

import "github.com/bmcszk/devilutionGo/internal/diablo/entities"

// This file ports the belt half of inv.cpp: which items fit a 1x1 belt slot,
// the placement sink used by pickup and stash-close, and slot removal.

// IsBeltItem ports FitsInBeltSlot (inv.cpp:174-177): a belt slot holds only
// items whose inventory footprint is a single cell.
func IsBeltItem(item entities.Item, sprite CursorSpriteSize) bool {
	return ItemSize(item, sprite) == (Size{Width: 1, Height: 1})
}

// CanUse ports Player::CanUseItem's requirement check (player.cpp:1527-1529).
// Upstream also runs IsItemValid, the multiplayer item-legitimacy check
// (items/validation.cpp:158), which belongs to the net layer and is not part
// of this port.
func (v Inv) CanUse(item entities.Item) bool {
	return v.player.Strength >= int(item.MinStr) &&
		v.player.Magic >= int(item.MinMag) &&
		v.player.Dexterity >= int(item.MinDex)
}

// CanPlaceOnBelt ports CanBePlacedOnBelt (inv.cpp:1161-1167).
func (v Inv) CanPlaceOnBelt(item entities.Item) bool {
	return IsBeltItem(item, v.data.Sprite) &&
		ItemType(item.Itype) != ItemTypeGold &&
		v.CanUse(item) &&
		v.isUsable(item)
}

// isUsable ports Item::isUsable (items.cpp:4806-4811) through the injected
// item table: an unwired table makes nothing usable, as Data documents.
func (v Inv) isUsable(item entities.Item) bool {
	if v.data.Usable == nil {
		return false
	}
	return v.data.Usable(item.IDidx)
}

// PlaceInBelt ports AutoPlaceItemInBelt without the persistItem flag
// (inv.cpp:1328-1350): store the item in the first empty belt slot. It
// reports false when every slot is taken.
func (v Inv) PlaceInBelt(item entities.Item) bool {
	if !v.CanPlaceOnBelt(item) {
		return false
	}
	for i := range v.belt {
		if IsEmpty(v.belt[i]) {
			v.belt[i] = item
			return true
		}
	}
	return false
}

// RemoveBeltSlot ports Player::RemoveSpdBarItem without the belt redraw and
// the CalcScrolls recompute (player.cpp:1577-1584); scroll bookkeeping belongs
// to the spell layer.
func (v Inv) RemoveBeltSlot(slot int) {
	clearItem(&v.belt[slot])
}
