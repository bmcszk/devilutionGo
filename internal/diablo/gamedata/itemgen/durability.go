package itemgen

// The two equipment hands, in upstream order (player.h inv_body_loc).
const (
	HandLeft  HandSlot = "left"
	HandRight HandSlot = "right"
)

// The four outcomes of the wear walk.
const (
	WearDecayed WearKind = iota + 1 // _iPLDam -= 5, item stays equipped (:447)
	WearDamaged                     // durability lost one point (:485)
	WearBroken                      // durability hit zero: RemoveEquipment (:481)
)

// Equipment class and item type names the wear flow compares.
const (
	classWeapon = "Weapon"
	typeShield  = "Shield"
)

// HandSlot is an equipment hand (inv_body_loc INVLOC_HAND_LEFT/RIGHT).
type HandSlot string

// WearKind is what one DamageWeapon pass did to a slot (player.cpp:457-524).
type WearKind int

// HandState is the slot state DamageWeapon reads.
type HandState struct {
	Empty      bool
	ItemType   string // ItemType::Shield is the only type name the flow compares
	Class      string // ICLASS_WEAPON
	Durability int    // _iDurability
	PLDam      int    // _iPLDam
	Decay      bool   // HasAnyOf(_iDamAcFlags, ItemSpecialEffectHf::Decay)
}

// WearAction is one mutation the caller applies to its own equipment state.
type WearAction struct {
	Slot       HandSlot
	Kind       WearKind
	PLDam      int  // new _iPLDam for WearDecayed
	Durability int  // new _iDurability for WearDamaged and WearBroken
	CalcPlrInv bool // upstream re-derives stats here (player.cpp:451, 482, 494, 505, 516)
}

// WeaponState is the pair of hands a DamageWeapon pass inspects.
type WeaponState struct {
	Left  HandState
	Right HandState
}

// WearResult is the ordered action list of one DamageWeapon pass. Upstream
// applies each action to the live Player as it walks; each action here is the
// same mutation, so the caller can replay them in order to get the same state.
type WearResult struct {
	Actions []WearAction
}

// DamageWeapon ports DamageWeapon (player.cpp:457-524): the decay pass first,
// then the single durability hit gated by FlipCoin(damageFrequency).
func DamageWeapon(state WeaponState, damageFrequency int, rnd func(int32) int32) (WearResult, error) {
	actions, broken := decayPass(state)
	if broken || rnd(int32(damageFrequency)) != 0 {
		return WearResult{Actions: actions}, nil
	}
	handActions, _ := damageHand(state)
	actions = append(actions, handActions...)

	return WearResult{Actions: actions}, nil
}

// decayPass is the WeaponDecay(left)/WeaponDecay(right) pair: a broken weapon
// ends the whole pass, a decayed one keeps it going (:469-470).
func decayPass(state WeaponState) ([]WearAction, bool) {
	var actions []WearAction

	for _, slot := range []HandSlot{HandLeft, HandRight} {
		action, decayed := decayHand(handOf(state, slot), slot)
		if !decayed {
			continue
		}
		actions = append(actions, action)
		if action.Kind == WearBroken {
			return actions, true
		}
	}

	return actions, false
}

// decayHand is WeaponDecay (player.cpp:444-457): each hit costs the decaying
// weapon 5% damage until it reaches -100 and is removed.
func decayHand(hand HandState, slot HandSlot) (WearAction, bool) {
	if hand.Empty || hand.Class != classWeapon || !hand.Decay {
		return WearAction{}, false
	}
	dam := hand.PLDam - 5
	if dam <= -100 {
		return WearAction{Slot: slot, Kind: WearBroken, PLDam: dam, CalcPlrInv: true}, true
	}

	return WearAction{Slot: slot, Kind: WearDecayed, PLDam: dam, CalcPlrInv: true}, true
}

// damageHand ports the four durability blocks (player.cpp:478-521): the left
// weapon, the right weapon, then the shield in whichever hand is free. A
// non-breaking wear does NOT stop the pass (no return unless broken), so a
// dual-wield wears both weapons; an indestructible match ends the pass.
func damageHand(state WeaponState) ([]WearAction, bool) {
	var actions []WearAction
	for _, hand := range []struct {
		hand HandState
		slot HandSlot
	}{
		{state.Left, HandLeft},
		{state.Right, HandRight},
	} {
		action, stop := weaponWear(hand.hand, hand.slot)
		if stop {
			return actions, action.Kind == WearBroken
		}
		actions, stop = appendWear(actions, action, hand.slot)
		if stop {
			return actions, true
		}
	}
	if action, ok := damageShield(state); ok {
		actions = append(actions, action)
	}
	return actions, false
}

// weaponWear is one weapon block (player.cpp:478-499): an empty hand skips,
// an indestructible weapon stops the pass (broken=false), a live weapon wears.
func weaponWear(hand HandState, slot HandSlot) (WearAction, bool) {
	if !isWeapon(hand) {
		return WearAction{}, false
	}
	if hand.Durability == DurIndestructible {
		return WearAction{}, true
	}
	return wearAction(hand, slot), false
}

// appendWear appends a wear action for its slot; it reports whether the
// action broke the item.
func appendWear(actions []WearAction, action WearAction, slot HandSlot) ([]WearAction, bool) {
	if action.Slot != slot {
		return actions, false
	}
	actions = append(actions, action)
	return actions, action.Kind == WearBroken
}

// damageShield ports the two shield-in-free-hand blocks
// (player.cpp:501-521), including the DUR_INDESTRUCTIBLE early return.
func damageShield(state WeaponState) (WearAction, bool) {
	switch {
	case state.Left.Empty && state.Right.ItemType == typeShield:
		return wearShield(state.Right, HandRight), true
	case state.Right.Empty && state.Left.ItemType == typeShield:
		return wearShield(state.Left, HandLeft), true
	default:
		return WearAction{}, false
	}
}

// wearShield ports one shield durability point (:506-510, 516-520).
func wearShield(hand HandState, slot HandSlot) WearAction {
	if hand.Durability == DurIndestructible {
		return WearAction{}
	}
	durability := hand.Durability - 1
	if durability <= 0 {
		return WearAction{Slot: slot, Kind: WearBroken, Durability: durability, CalcPlrInv: true}
	}
	return WearAction{Slot: slot, Kind: WearDamaged, Durability: durability}
}

// wearAction applies one durability point (:481-486) to a weapon hand.
func wearAction(hand HandState, slot HandSlot) WearAction {
	durability := hand.Durability - 1
	if durability <= 0 {
		return WearAction{Slot: slot, Kind: WearBroken, Durability: durability, CalcPlrInv: true}
	}
	return WearAction{Slot: slot, Kind: WearDamaged, Durability: durability}
}

func handOf(state WeaponState, slot HandSlot) HandState {
	if slot == HandLeft {
		return state.Left
	}

	return state.Right
}

func isWeapon(hand HandState) bool {
	return !hand.Empty && hand.Class == classWeapon
}
