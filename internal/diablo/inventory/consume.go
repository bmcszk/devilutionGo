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
	// MiscStaff is IMISC_STAFF (items.h:70).
	MiscStaff int8 = 17
	// MiscUnique is IMISC_UNIQUE (items.h:72).
	MiscUnique int8 = 44
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
// Rune spell ids, in SpellID order (tables/spelldat.h:81-85). The spell enum
// itself belongs to the spell layer; these numeric ids are what Item.Spell and
// the cast path carry.
const (
	// SpellRuneOfFire is SpellID::RuneOfFire.
	SpellRuneOfFire int32 = 47
	// SpellRuneOfLight is SpellID::RuneOfLight.
	SpellRuneOfLight int32 = 48
	// SpellRuneOfNova is SpellID::RuneOfNova.
	SpellRuneOfNova int32 = 49
	// SpellRuneOfImmolation is SpellID::RuneOfImmolation.
	SpellRuneOfImmolation int32 = 50
	// SpellRuneOfStone is SpellID::RuneOfStone.
	SpellRuneOfStone int32 = 51
)

// Rune misc ids (tables/itemdat.h:480-486).
const (
	// MiscRuneF is IMISC_RUNEF.
	MiscRuneF int8 = 47
	// MiscRuneL is IMISC_RUNEL.
	MiscRuneL int8 = 48
	// MiscGRRuneL is IMISC_GR_RUNEL.
	MiscGRRuneL int8 = 49
	// MiscGRRuneF is IMISC_GR_RUNEF.
	MiscGRRuneF int8 = 50
	// MiscRuneS is IMISC_RUNES.
	MiscRuneS int8 = 51
)

// ItemPredicate is the item test the search helpers run (inv.h:255-281).
type ItemPredicate func(item entities.Item) bool

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
// belt, while scrolls and runes stay until ConsumeScroll removes one. The
// IMISC_NOTE branch (inv.cpp:2218-2221) is a quest-text window and stays with
// the UI layer, so a note is treated like any other consumable here.
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
// CanUseStaff ports CanUseStaff (inv.cpp:1079-1085): a charged staff (or

// HasInventoryItem ports HasInventoryItem (inv.h:261-267): any live InvList
// slot up to NumInv matches.
func (v Inv) HasInventoryItem(predicate ItemPredicate) bool {
	for i := range *v.numInv {
		if predicate(v.list[i]) {
			return true
		}
	}
	return false
}

// HasBeltItem ports HasBeltItem (inv.h:268-274): any of the eight belt slots
// matches.
func (v Inv) HasBeltItem(predicate ItemPredicate) bool {
	for i := range BeltSlots {
		if predicate(v.belt[i]) {
			return true
		}
	}
	return false
}

// HasInventoryOrBeltItem ports HasInventoryOrBeltItem (inv.h:277-282).
func (v Inv) HasInventoryOrBeltItem(predicate ItemPredicate) bool {
	return v.HasInventoryItem(predicate) || v.HasBeltItem(predicate)
}

// IsScrollOf ports Item::isScrollOf (items.h:396-399).
func IsScrollOf(item entities.Item, spell int32) bool {
	return IsScroll(item) && item.Spell == spell
}

// IsRuneOf ports Item::isRuneOf (items.h:406-426).
func IsRuneOf(item entities.Item, spell int32) bool {
	if !IsRune(item) {
		return false
	}
	switch item.MiscId {
	case MiscRuneF:
		return spell == SpellRuneOfFire
	case MiscRuneL:
		return spell == SpellRuneOfLight
	case MiscGRRuneL:
		return spell == SpellRuneOfNova
	case MiscGRRuneF:
		return spell == SpellRuneOfImmolation
	case MiscRuneS:
		return spell == SpellRuneOfStone
	}
	return false
}

// CanUseScroll ports CanUseScroll (inv.cpp:2054-2062): the spell must be
// castable where the player stands and a scroll or rune of it must be carried.
// The town check and the spell table are injected, so this layer stays free of
// level and spell data.
func (v Inv) CanUseScroll(spell int32) bool {
	if v.inTown() && !v.spellAllowedInTown(spell) {
		return false
	}
	return v.HasInventoryOrBeltItem(func(item entities.Item) bool {
		return IsScrollOf(item, spell) || IsRuneOf(item, spell)
	})
}

// inTown reports the injected level check; an unwired hook means "not in town".
func (v Inv) inTown() bool {
	return v.data.InTown != nil && v.data.InTown()
}

// spellAllowedInTown reports the injected spell-data check; an unwired hook
// allows nothing, matching a missing spell table.
func (v Inv) spellAllowedInTown(spell int32) bool {
	return v.data.SpellAllowedInTown != nil && v.data.SpellAllowedInTown(spell)
}

// unique) whose stored spell is the one being cast.
func CanUseStaff(staff entities.Item, spell int32) bool {
	if IsEmpty(staff) {
		return false
	}
	if staff.MiscId != MiscStaff && staff.MiscId != MiscUnique {
		return false
	}
	return staff.Spell == spell && staff.Charges > 0
}

func isSpellSource(item entities.Item, spell int32) bool {
	return IsScroll(item) && item.Spell == spell
}
