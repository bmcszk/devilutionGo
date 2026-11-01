package inventory

import "github.com/bmcszk/devilutionGo/internal/diablo/entities"

// This file ports inventory-side consumption: what a belt hotkey uses up, how
// an inventory item is removed, and how the executed spell finds and consumes
// its scroll or rune (inv.cpp:2018-2052, player.cpp:1532-1575).

// item_misc_id values the consumption rules branch on (tables/itemdat.h:433-489).
const (
	MiscNone      int8 = 0
	MiscScroll    int8 = 21
	MiscScrollT   int8 = 22
	MiscMapOfDoom int8 = 42
	MiscRuneFirst int8 = 46
	MiscRuneLast  int8 = 52
)

// Inventory and belt item locations of the executed spell's source item
// (inv.h:28-38).
const (
	InvItemFirst  int8 = 7
	InvItemLast   int8 = 46
	BeltItemFirst int8 = 47
	BeltItemLast  int8 = 54
)

// IsScroll ports Item::isScroll (items.h:391-394).
func IsScroll(item entities.Item) bool {
	return item.MiscId == MiscScroll || item.MiscId == MiscScrollT
}

// IsRune ports Item::isRune (items.h:401-404).
func IsRune(item entities.Item) bool {
	return item.MiscId > MiscRuneFirst && item.MiscId < MiscRuneLast
}

// Remove ports Player::RemoveInvItem without the network send and the
// CalcScrolls recompute (player.cpp:1532-1575). Later items shift down so the
// InvList order is kept, and every grid reference to the removed item is
// cleared.
func (v Inv) Remove(itemIndex int) {
	v.clearPlaceholders(itemIndex + 1)
	clearItem(&v.list[itemIndex])
	*v.numInv--
	if *v.numInv > 0 && *v.numInv != itemIndex {
		for index := itemIndex; index < *v.numInv; index++ {
			v.list[index] = pop(&v.list[index+1])
		}
		v.shiftPlaceholders(itemIndex)
	}
}

// shiftPlaceholders renumbers the grid after an item was shifted down
// (player.cpp:1566-1573).
func (v Inv) shiftPlaceholders(itemIndex int) {
	for i := range v.grid {
		if v.grid[i] > int8(itemIndex+1) {
			v.grid[i]--
		}
		if v.grid[i] < int8(-(itemIndex + 1)) {
			v.grid[i]++
		}
	}
}

// ConsumeBeltSlot ports the belt tail of UseInvItem (inv.cpp:2216-2226): after
// the item's effect ran, potions and other non-scroll consumables leave the
// belt, while scrolls and runes stay until ConsumeScroll removes one.
func (v Inv) ConsumeBeltSlot(slot int) {
	item := v.belt[slot]
	if IsScroll(item) || IsRune(item) {
		return
	}
	v.RemoveBeltSlot(slot)
}

// ConsumeScroll ports ConsumeScroll (inv.cpp:2018-2052) minus the app_fatal on
// an invalid index: it first tries the source slot the spell was cast from and
// otherwise removes the first matching scroll or rune.
func (v Inv) ConsumeScroll(spellFrom int8, spell int32) {
	if v.consumeFromSource(spellFrom, spell) {
		return
	}
	v.removeFirstScroll(spell)
}

// consumeFromSource tries the recorded source slot of the executed spell
// (inv.cpp:2028-2047).
func (v Inv) consumeFromSource(spellFrom int8, spell int32) bool {
	switch {
	case spellFrom >= InvItemFirst && spellFrom <= InvItemLast:
		return v.consumeInventorySlot(int(spellFrom-InvItemFirst), spell)
	case spellFrom >= BeltItemFirst && spellFrom <= BeltItemLast:
		return v.consumeBeltSlotFor(spellFrom-BeltItemFirst, spell)
	default:
		return false
	}
}

// consumeInventorySlot removes the scroll in an inventory slot when it matches
// the executed spell.
func (v Inv) consumeInventorySlot(itemIndex int, spell int32) bool {
	if IsEmpty(v.list[itemIndex]) || !isSpellSource(v.list[itemIndex], spell) {
		return false
	}
	v.Remove(itemIndex)
	return true
}

// consumeBeltSlotFor is consumeInventorySlot for the belt.
func (v Inv) consumeBeltSlotFor(slot int8, spell int32) bool {
	if IsEmpty(v.belt[slot]) || !isSpellSource(v.belt[slot], spell) {
		return false
	}
	v.RemoveBeltSlot(int(slot))
	return true
}

// removeFirstScroll ports RemoveInventoryOrBeltItem with the isCurrentSpell
// predicate (inv.h:320-352): inventory first, then the belt.
func (v Inv) removeFirstScroll(spell int32) {
	for i := range *v.numInv {
		if isSpellSource(v.list[i], spell) {
			v.Remove(i)
			return
		}
	}
	for i := range v.belt {
		if isSpellSource(v.belt[i], spell) {
			v.RemoveBeltSlot(i)
			return
		}
	}
}

// isSpellSource ports the isCurrentSpell lambda (inv.cpp:2021-2023). The rune
// half needs the rune-to-spell table (items.h:406-420), which lands with the
// spell layer, so only the scroll form is matched here.
func isSpellSource(item entities.Item, spell int32) bool {
	return IsScroll(item) && item.Spell == spell
}
