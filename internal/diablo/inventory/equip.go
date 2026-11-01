package inventory

import "github.com/bmcszk/devilutionGo/internal/diablo/entities"

// This file ports equipping: the equipment test, the hand-wield rules and the
// auto-equip sweep inv.cpp runs on pickup and stash close.

// The body slots in upstream order (inv_body_loc, player.h:73-82); AutoEquip
// sweeps them in this order.
const (
	BodyHead BodyLoc = iota
	BodyRingLeft
	BodyRingRight
	BodyAmulet
	BodyHandLeft
	BodyHandRight
	BodyChest
	BodyLocCount
)

// The equipment locations in upstream order (item_equip_type,
// tables/itemdat.h:98-109), i.e. the values stored in Item.Loc.
const (
	LocNone ItemLoc = iota
	LocOneHand
	LocTwoHand
	LocArmor
	LocHelm
	LocRing
	LocAmulet
	LocUnequippable
	LocBelt
	LocInvalid ItemLoc = -1
)

// Hero classes GetItemLocation and the dual-wield rule branch on
// (tables/playerdat.hpp:21-26).
const (
	HeroBarbarian uint8 = 5
)

// Equipped body slot ids (inv.h:48-56).
const (
	// SlotHead is SLOTXY_HEAD.
	SlotHead = 0
	// SlotRingLeft is SLOTXY_RING_LEFT.
	SlotRingLeft = 1
	// SlotRingRight is SLOTXY_RING_RIGHT.
	SlotRingRight = 2
	// SlotAmulet is SLOTXY_AMULET.
	SlotAmulet = 3
	// SlotHandLeft is SLOTXY_HAND_LEFT.
	SlotHandLeft = 4
	// SlotHandRight is SLOTXY_HAND_RIGHT.
	SlotHandRight = 5
	// SlotChest is SLOTXY_CHEST.
	SlotChest = 6
)

// BodyLoc is inv_body_loc (player.h:73-82).
type BodyLoc uint8

// ItemLoc is item_equip_type (tables/itemdat.h:98-109).
type ItemLoc int8

// IsEquipment ports Item::isEquipment (items.h:288-307).
func IsEquipment(item entities.Item) bool {
	if IsEmpty(item) {
		return false
	}
	switch ItemLoc(item.Loc) {
	case LocAmulet, LocArmor, LocHelm, LocOneHand, LocRing, LocTwoHand:
		return true
	default:
		return false
	}
}

// CanEquip ports CanEquip (inv.cpp:185-188): the item is equipment and its
// stat requirements are met.
func CanEquip(item entities.Item) bool {
	return IsEquipment(item) && item.StatFlag
}

// itemLocation ports Player::GetItemLocation (player.h:557-563): a Barbarian
// wields two-handed swords and maces in one hand.
func (v Inv) itemLocation(item entities.Item) ItemLoc {
	loc := ItemLoc(item.Loc)
	itype := ItemType(item.Itype)
	twoHandWeapon := loc == LocTwoHand && (itype == ItemTypeSword || itype == ItemTypeMace)
	if v.player.Class == HeroBarbarian && twoHandWeapon {
		return LocOneHand
	}
	return loc
}

// CanWield ports CanWield (inv.cpp:199-235): whether either hand can take the
// weapon right now.
func (v Inv) CanWield(item entities.Item) bool {
	loc := v.itemLocation(item)
	if !CanEquip(item) || (loc != LocOneHand && loc != LocTwoHand) {
		return false
	}
	return v.handsAccept(item, loc)
}

// handsAccept applies the hand rules (inv.cpp:205-234).
func (v Inv) handsAccept(item entities.Item, loc ItemLoc) bool {
	left, right := &v.body[BodyHandLeft], &v.body[BodyHandRight]
	if IsEmpty(*left) && IsEmpty(*right) {
		return true
	}
	if !IsEmpty(*left) && !IsEmpty(*right) {
		return false
	}
	occupied := *left
	if IsEmpty(*left) {
		occupied = *right
	}
	if v.canDualWield(item, occupied) {
		return true
	}
	// Two different one-handers may share the hands; the same class may not.
	return loc == LocOneHand && v.itemLocation(occupied) == LocOneHand && item.Class != occupied.Class
}

// canDualWield ports the Bard dual-wield branch (inv.cpp:214-228).
func (v Inv) canDualWield(item, occupied entities.Item) bool {
	if v.data.DualWield == nil || !v.data.DualWield(v.player.Class) {
		return false
	}
	return v.isOneHandSwordOrMace(occupied) && v.isOneHandSwordOrMace(item)
}

// isOneHandSwordOrMace is the shared hand-weapon test of the dual-wield rule.
func (v Inv) isOneHandSwordOrMace(item entities.Item) bool {
	itype := ItemType(item.Itype)
	return v.itemLocation(item) == LocOneHand && (itype == ItemTypeSword || itype == ItemTypeMace)
}

// CanEquipAt ports CanEquip (inv.cpp:245-271) without the _pmode guard: the
// animation-state check belongs to the action layer, and this port's PlrMode
// is a reduced enum that does not carry PM_WALK_SIDEWAYS.
func (v Inv) CanEquipAt(item entities.Item, loc BodyLoc) bool {
	if !CanEquip(item) || !IsEmpty(v.body[loc]) {
		return false
	}
	itemLoc := ItemLoc(item.Loc)
	switch loc {
	case BodyAmulet:
		return itemLoc == LocAmulet
	case BodyChest:
		return itemLoc == LocArmor
	case BodyHandLeft, BodyHandRight:
		return v.CanWield(item)
	case BodyHead:
		return itemLoc == LocHelm
	case BodyRingLeft, BodyRingRight:
		return itemLoc == LocRing
	default:
		return false
	}
}

// AutoEquipAt ports AutoEquip(player, item, bodyLocation, …) (inv.cpp:283-300)
// without the sound and the CalcPlrInv recompute, which belong to the audio
// and stats layers.
func (v Inv) AutoEquipAt(item entities.Item, loc BodyLoc) bool {
	if !v.CanEquipAt(item, loc) {
		return false
	}
	v.body[loc] = item
	return true
}

// AutoEquip ports the AutoEquip sweep (inv.cpp:288-300): the first body slot
// that accepts the item takes it.
func (v Inv) AutoEquip(item entities.Item) bool {
	if !CanEquip(item) {
		return false
	}
	for loc := BodyHead; loc < BodyLocCount; loc++ {
		if v.AutoEquipAt(item, loc) {
			return true
		}
	}
	return false
}

// EquipTypeForSlot ports GetItemEquipType (inv.cpp:543-559): the body location
// a slot accepts, with the hands honouring a two-handed request.
func EquipTypeForSlot(slot int, desiredLocation ItemLoc) ItemLoc {
	switch slot {
	case SlotHead:
		return LocHelm
	case SlotRingLeft, SlotRingRight:
		return LocRing
	case SlotAmulet:
		return LocAmulet
	case SlotHandLeft, SlotHandRight:
		if desiredLocation == LocTwoHand {
			return LocTwoHand
		}
		return LocOneHand
	case SlotChest:
		return LocArmor
	}
	if int8(slot) >= BeltItemFirst && int8(slot) <= BeltItemLast {
		return LocBelt
	}
	return LocUnequippable
}
