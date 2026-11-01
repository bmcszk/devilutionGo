package stores

import "slices"

// The premium shelf: Griswold's rotating special stock. The item generation
// itself is items.cpp's SpawnOnePremium, which the frozen itemgen layer owns,
// so the shelf is generic over the item type and takes the spawn as a
// callback: this file ports the slot count, the level deltas and the shift.

// Premium shelf sizes (stores.h:29-30).
const (
	// NumSmithItems is NumSmithItems: the vanilla premium slot count.
	NumSmithItems = 6
	// NumSmithItemsHF is NumSmithItemsHf: the Hellfire premium slot count.
	NumSmithItemsHF = 15
)

// NumSmithItemsHfSlots is the vanilla delta table's length; it equals
// NumSmithItems and is separate only so the array bound reads as the table's.
const NumSmithItemsHfSlots = NumSmithItems

// BoySlotCap is MaxBoyValue, the vanilla cap SpawnBoy's roll rejects above
// (items.h:39).
const BoySlotCap = 90000

// BoySlotCapHf is MaxBoyValueHf (items.h:40).
const BoySlotCapHf = 200000

// itemLevelAddHf is itemLevelAddHf (items.cpp:372-388): the per-slot level
// delta added to the shelf level for Hellfire. It stays unexported so only
// PremiumDeltas hands it out, as a copy.
var itemLevelAddHf = [NumSmithItemsHF]int{-1, -1, -1, 0, 0, 0, 0, 1, 1, 1, 1, 2, 2, 3, 3}

// itemLevelAdd is itemLevelAdd (items.cpp:362-370): the per-slot level delta
// added to the shelf level for vanilla.
var itemLevelAdd = [NumSmithItemsHfSlots]int{-1, -1, 0, 0, 1, 2}

// PremiumStock is PremiumItems plus PremiumItemLevel (stores.cpp:50-53).
type PremiumStock[T any] struct {
	// Items is the shelf, in slot order.
	Items []T
	// Level is PremiumItemLevel, the shelf's quality level.
	Level int
}

// ResetPremium ports InitStores' premium lines (stores.cpp:2145-2161): the
// shelf empties and its level returns to 1, so the next SpawnPremium starts
// from the bottom slot deltas. The countdown counter it also clears is the
// savegame field PremiumItemCount (:52), which no rule reads back.
func ResetPremium[T any](stock *PremiumStock[T]) {
	stock.Items = nil
	stock.Level = 1
}

// NumberPremiumItems is the shelf size rule: gbIsHellfire ? NumSmithItemsHf :
// NumSmithItems (items.cpp:4445, the same switch the savegame loader makes in
// loadsave.cpp:2489-2494).
func NumberPremiumItems(env Env) int {
	if env.Hellfire {
		return NumSmithItemsHF
	}

	return NumSmithItems
}

// PremiumDeltas is the level delta table of the game mode (items.cpp:4448),
// copied out so no caller can rewrite the shared table.
func PremiumDeltas(env Env) []int {
	if env.Hellfire {
		return slices.Clone(itemLevelAddHf[:])
	}

	return slices.Clone(itemLevelAdd[:])
}

// SpawnPremium ports SpawnPremium (items.cpp:4442-4470): fill the shelf, then
// raise it to the character level, shifting the old entries down and respawning
// the last slots at each step.
func SpawnPremium[T any](stock *PremiumStock[T], charLevel int, env Env, spawn func(plvl int) T) {
	maxItems := NumberPremiumItems(env)
	deltas := PremiumDeltas(env)

	for len(stock.Items) < maxItems {
		stock.Items = append(stock.Items, spawn(stock.Level+deltas[len(stock.Items)]))
	}

	for stock.Level < charLevel {
		stock.Level++

		if env.Hellfire {
			copy(stock.Items[0:10], stock.Items[3:13])
			stock.Items[11] = stock.Items[13]
			stock.Items[13] = stock.Items[14]
			stock.Items[10] = spawn(stock.Level + deltas[10])
			stock.Items[12] = spawn(stock.Level + deltas[12])
			stock.Items[14] = spawn(stock.Level + deltas[14])

			continue
		}

		copy(stock.Items[0:3], stock.Items[2:5])
		stock.Items[4] = stock.Items[5]
		stock.Items[3] = spawn(stock.Level + deltas[3])
		stock.Items[5] = spawn(stock.Level + deltas[5])
	}
}

// ReplacePremium ports ReplacePremium (items.cpp:4435-4440): a bought slot is
// refilled at the character level plus that slot's delta.
func ReplacePremium[T any](stock *PremiumStock[T], idx, charLevel int, env Env, spawn func(plvl int) T) bool {
	if idx < 0 || idx >= NumberPremiumItems(env) {
		return false
	}

	stock.Items[idx] = spawn(PremiumDeltas(env)[idx] + charLevel)

	return true
}

// BoyNeedsRestock ports SpawnBoy's early return: Wirt keeps his offer until the
// shelf level falls half a character level behind (items.cpp:4555-4556).
func BoyNeedsRestock(boyLevel, charLevel int, boyItemEmpty bool) bool {
	return boyLevel < charLevel/2 || boyItemEmpty
}

// BoyValueCap returns the value cap SpawnBoy enforces per game mode
// (items.cpp:4566-4572).
func BoyValueCap(env Env) int {
	if env.Hellfire {
		return BoySlotCapHf
	}

	return BoySlotCap
}
