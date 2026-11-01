package stores

import (
	"github.com/bmcszk/devilutionGo/internal/diablo/entities"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
	"github.com/bmcszk/devilutionGo/internal/diablo/inventory"
)

// Env is the game-mode state the store rules branch on: gbIsHellfire and the
// spell table size IsValidSpell needs (spells.cpp:72-75).
type Env struct {
	// Hellfire is gbIsHellfire.
	Hellfire bool
	// SpellCount is SpellsData.size().
	SpellCount int
}

// IsValidSpell ports IsValidSpell (spells.cpp:72-75): a spell id above Null
// and below the table size.
func (e Env) IsValidSpell(spell int32) bool {
	return spell > 0 && int(spell) < e.SpellCount
}

// item_misc_id and item_class values the buy rules compare against
// (tables/itemdat.h:433-489, 89-96); the ordinals are the enum positions.
const (
	// miscStaff is IMISC_STAFF (tables/itemdat.h:457).
	miscStaff int8 = 23
	// miscUnique is IMISC_UNIQUE (tables/itemdat.h:461).
	miscUnique int8 = 27
	// miscOilFirst is IMISC_OILFIRST (tables/itemdat.h:463).
	miscOilFirst int8 = 29
	// miscOilLast is IMISC_OILLAST (tables/itemdat.h:475).
	miscOilLast int8 = 41
	// classQuest is ICLASS_QUEST (tables/itemdat.h:95).
	classQuest = 5
)

// _item_indexes values the buy rules compare against
// (tables/itemdat.h, IDI_ enum ordinal).
const (
	// idiFirstQuest is IDI_FIRSTQUEST (tables/itemdat.h:32).
	idiFirstQuest = 6
	// idiLastQuest is IDI_LASTQUEST (tables/itemdat.h:49).
	idiLastQuest = 22
	// idiLazStaff is IDI_LAZSTAFF (tables/itemdat.h:60).
	idiLazStaff = 33
)

// qualityNormal is ITEM_QUALITY_NORMAL (items.h:41).
const qualityNormal uint8 = 0

// IdentifyPrice is the flat fee AddStoreHoldId puts on every unknown item
// (stores.cpp:1095-1102).
const IdentifyPrice = 100

// SellValue ports the sell-list price (stores.cpp:558-562 and 584-588): an
// identified magic item sells for its base value, then a quarter of it, floored
// at one gold.
func SellValue(item entities.Item) int {
	value := item.Value
	if item.Magical != qualityNormal && item.Identified {
		value = item.Ivalue
	}

	if value/4 < 1 {
		return 1
	}

	return value / 4
}

// RepairPrice ports AddStoreHoldRepair's price (stores.cpp:2218-2233). The
// second result is false for the magic-item branch that computes zero and
// offers no repair line at all.
func RepairPrice(item entities.Item) (int, bool) {
	due := item.MaxDur - item.Durability

	if item.Magical != qualityNormal && item.Identified {
		if item.MaxDur <= 0 {
			return 0, false
		}

		price := 30 * item.Ivalue * due / (item.MaxDur * 100 * 2)
		if price == 0 {
			return 0, false
		}

		return price, true
	}

	price := 1
	if item.MaxDur > 0 {
		price = item.Value * due / (item.MaxDur * 2)
	}

	if price < 1 {
		price = 1
	}

	return price, true
}

// RechargePrice ports AddStoreHoldRecharge's price (stores.cpp:851-859): the
// item value plus the spell's staff cost, scaled by the missing charges.
func RechargePrice(item entities.Item, staffCost int) int {
	if item.MaxCharges <= 0 {
		return 0
	}

	value := item.Value + staffCost

	return value * (item.MaxCharges - item.Charges) / (item.MaxCharges * 2)
}

// BoyPrice ports the Wirt haggling price (stores.cpp:1821-1824): a quarter off
// in Hellfire, half again otherwise.
func BoyPrice(item entities.Item, env Env) int {
	if env.Hellfire {
		return item.Ivalue - item.Ivalue/4
	}

	return item.Ivalue + item.Ivalue/2
}

// TotalGold ports TotalPlayerGold (stores.cpp:459-462): carried gold plus the
// shared stash.
func TotalGold(playerGold, stashGold int) int {
	return playerGold + stashGold
}

// CanAfford ports PlayerCanAfford (stores.cpp:2861-2864). The upstream uint32
// cast makes a negative price unaffordable, so that case is spelled out.
func CanAfford(gold, price int) bool {
	if price < 0 {
		return false
	}

	return gold >= price
}

// GoldFit ports StoreGoldFit (stores.cpp:1414-1426): the gold a sale adds must
// fit in the item's own inventory cells, or in the free cells as well.
func GoldFit(cost, itemRoomForGold, roomForGold int) bool {
	if cost <= itemRoomForGold {
		return true
	}

	return cost <= itemRoomForGold+roomForGold
}

// SmithWillBuy ports SmithWillBuy (stores.cpp:2866-2886).
func SmithWillBuy(item entities.Item, env Env) bool {
	if inventory.IsEmpty(item) {
		return false
	}

	if item.MiscId > miscOilFirst && item.MiscId < miscOilLast {
		return true
	}

	if inventory.ItemType(item.Itype) == inventory.ItemTypeMisc {
		return false
	}

	if inventory.ItemType(item.Itype) == inventory.ItemTypeGold {
		return false
	}

	if inventory.ItemType(item.Itype) == inventory.ItemTypeStaff && (!env.Hellfire || env.IsValidSpell(item.Spell)) {
		return false
	}

	if item.Class == classQuest {
		return false
	}

	return int(item.IDidx) != idiLazStaff
}

// WitchWillBuy ports WitchWillBuy (stores.cpp:2888-2908).
func WitchWillBuy(item entities.Item, env Env) bool {
	if inventory.IsEmpty(item) {
		return false
	}

	rv := inventory.ItemType(item.Itype) == inventory.ItemTypeMisc

	if item.MiscId > 29 && item.MiscId < 41 {
		rv = false
	}

	if item.Class == classQuest {
		rv = false
	}

	if inventory.ItemType(item.Itype) == inventory.ItemTypeStaff && (!env.Hellfire || env.IsValidSpell(item.Spell)) {
		rv = true
	}

	if idi := int(item.IDidx); idi >= idiFirstQuest && idi <= idiLastQuest {
		rv = false
	}

	if int(item.IDidx) == idiLazStaff {
		rv = false
	}

	return rv
}

// Repairable ports SmithRepairOk (stores.cpp:610-626): an inventory item that
// has lost durability and is not indestructible.
func Repairable(item entities.Item) bool {
	if inventory.IsEmpty(item) {
		return false
	}

	if inventory.ItemType(item.Itype) == inventory.ItemTypeMisc {
		return false
	}

	if inventory.ItemType(item.Itype) == inventory.ItemTypeGold {
		return false
	}

	if item.Durability == item.MaxDur {
		return false
	}

	return item.MaxDur != itemgen.DurIndestructible
}

// Rechargeable ports WitchRechargeOk (stores.cpp:836-849): a staff or unique
// with charges missing.
func Rechargeable(item entities.Item) bool {
	if inventory.ItemType(item.Itype) == inventory.ItemTypeStaff && item.Charges != item.MaxCharges {
		return true
	}

	if (item.MiscId == miscUnique || item.MiscId == miscStaff) && item.Charges < item.MaxCharges {
		return true
	}

	return false
}

// Identifiable ports IdItemOk (stores.cpp:1084-1093): a magic item the player
// has not identified yet.
func Identifiable(item entities.Item) bool {
	if inventory.IsEmpty(item) {
		return false
	}

	if item.Magical == qualityNormal {
		return false
	}

	return !item.Identified
}
