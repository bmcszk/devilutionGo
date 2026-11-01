package stores

import (
	"github.com/bmcszk/devilutionGo/internal/diablo/entities"
	"github.com/bmcszk/devilutionGo/internal/diablo/inventory"
)

// MaxHoldItems is PlayerItems[48] (stores.cpp:44).
const MaxHoldItems = 48

// The equipped-slot ids PlayerItemIndexes carries for repair, recharge and
// identify lists (stores.cpp:637-651, 1108-1139). They are upstream's literal
// -1..-7 literals, *not* inv_body_loc order.
const (
	// HoldHead is INVLOC_HEAD's hold id.
	HoldHead int8 = -1
	// HoldChest is INVLOC_CHEST's hold id.
	HoldChest int8 = -2
	// HoldHandLeft is INVLOC_HAND_LEFT's hold id.
	HoldHandLeft int8 = -3
	// HoldHandRight is INVLOC_HAND_RIGHT's hold id.
	HoldHandRight int8 = -4
	// HoldRingLeft is INVLOC_RING_LEFT's hold id.
	HoldRingLeft int8 = -5
	// HoldRingRight is INVLOC_RING_RIGHT's hold id.
	HoldRingRight int8 = -6
	// HoldAmulet is INVLOC_AMULET's hold id.
	HoldAmulet int8 = -7
)

// bodySlots maps a hold id to its inv_body_loc index (player.h:73-82): head,
// ring left, ring right, amulet, hand left, hand right, chest.
var bodySlots = map[int8]inventory.BodyLoc{
	HoldHead:      inventory.BodyHead,
	HoldRingLeft:  inventory.BodyRingLeft,
	HoldRingRight: inventory.BodyRingRight,
	HoldAmulet:    inventory.BodyAmulet,
	HoldHandLeft:  inventory.BodyHandLeft,
	HoldHandRight: inventory.BodyHandRight,
	HoldChest:     inventory.BodyChest,
}

// BodySlot returns the inv_body_loc index a hold id addresses, and false for a
// non-body hold id.
func BodySlot(hold int8) (inventory.BodyLoc, bool) {
	loc, ok := bodySlots[hold]

	return loc, ok
}

// beltHoldID is the PlayerItemIndexes value for belt slot i: -(i + 1), the
// 4-argument StoreSellItem decode (stores.cpp:1433-1435).
func beltHoldID(i int) int8 {
	return int8(-(i + 1))
}

// Hold is one offered item: a copy of the player's item plus the index that
// points back at it (PlayerItems / PlayerItemIndexes, stores.cpp:44-46).
type Hold struct {
	// Item is the copy the panel prices and prints.
	Item entities.Item
	// Index is PlayerItemIndexes[i].
	Index int8
}

// SmithSellHolds builds the smith's buy-from-player list (stores.cpp:543-604).
func SmithSellHolds(p *entities.Player, env Env) []Hold {
	return sellHolds(p, func(item entities.Item) bool { return SmithWillBuy(item, env) })
}

// WitchSellHolds builds the witch's buy-from-player list (stores.cpp:768-834).
func WitchSellHolds(p *entities.Player, env Env) []Hold {
	return sellHolds(p, func(item entities.Item) bool { return WitchWillBuy(item, env) })
}

// sellHolds is the shared inventory-then-belt sweep of StartSmithSell
// (stores.cpp:552-588) and StartWitchSell (stores.cpp:777-813).
func sellHolds(p *entities.Player, willBuy func(entities.Item) bool) []Hold {
	holds := make([]Hold, 0, MaxHoldItems)

	for i := 0; i < p.NumInv && len(holds) < MaxHoldItems; i++ {
		if !willBuy(p.InvList[i]) {
			continue
		}

		item := p.InvList[i]
		item.Value = SellValue(item)
		item.Ivalue = item.Value
		holds = append(holds, Hold{Item: item, Index: int8(i)})
	}

	for i := 0; i < entities.MaxBeltItems && len(holds) < MaxHoldItems; i++ {
		if !willBuy(p.SpdList[i]) {
			continue
		}

		item := p.SpdList[i]
		item.Value = SellValue(item)
		item.Ivalue = item.Value
		holds = append(holds, Hold{Item: item, Index: beltHoldID(i)})
	}

	return holds
}

// RepairHolds builds the smith's repair list in upstream order — head, chest,
// both hands, then the inventory (stores.cpp:641-664).
func RepairHolds(p *entities.Player) []Hold {
	holds := make([]Hold, 0, MaxHoldItems)

	for _, hold := range []int8{HoldHead, HoldChest, HoldHandLeft, HoldHandRight} {
		loc, _ := BodySlot(hold)
		item := p.InvBody[loc]

		if inventory.IsEmpty(item) || item.Durability == item.MaxDur {
			continue
		}

		if held, ok := repairHold(item, hold); ok {
			holds = append(holds, held)
		}
	}

	for i := 0; i < p.NumInv && len(holds) < MaxHoldItems; i++ {
		if !Repairable(p.InvList[i]) {
			continue
		}

		if held, ok := repairHold(p.InvList[i], int8(i)); ok {
			holds = append(holds, held)
		}
	}

	return holds
}

// repairHold prices one entry the way AddStoreHoldRepair does and reports
// whether it is offered at all (stores.cpp:2211-2233).
func repairHold(item entities.Item, index int8) (Hold, bool) {
	price, ok := RepairPrice(item)
	if !ok {
		return Hold{}, false
	}

	item.Ivalue = price
	item.Value = price

	return Hold{Item: item, Index: index}, true
}

// RechargeHolds builds the witch's recharge list: the left hand first, then the
// inventory (stores.cpp:861-886). staffCost resolves IMisc spell staff cost.
func RechargeHolds(p *entities.Player, staffCost func(spell int32) int) []Hold {
	holds := make([]Hold, 0, MaxHoldItems)

	leftHand := p.InvBody[inventory.BodyHandLeft]
	if (inventory.ItemType(leftHand.Itype) == inventory.ItemTypeStaff || leftHand.MiscId == miscUnique) &&
		leftHand.Charges != leftHand.MaxCharges {
		holds = append(holds, rechargeHold(leftHand, HoldHandLeft, staffCost))
	}

	for i := 0; i < p.NumInv && len(holds) < MaxHoldItems; i++ {
		if !Rechargeable(p.InvList[i]) {
			continue
		}

		holds = append(holds, rechargeHold(p.InvList[i], int8(i), staffCost))
	}

	return holds
}

// rechargeHold prices one recharge entry (stores.cpp:851-859).
func rechargeHold(item entities.Item, index int8, staffCost func(spell int32) int) Hold {
	price := RechargePrice(item, staffCost(item.Spell))
	item.Value = price
	item.Ivalue = price

	return Hold{Item: item, Index: index}
}

// IdentifyHolds builds Cain's identify list in upstream order: head, chest,
// both hands, both rings, amulet, then the inventory (stores.cpp:1104-1160).
func IdentifyHolds(p *entities.Player) []Hold {
	holds := make([]Hold, 0, MaxHoldItems)

	order := []int8{HoldHead, HoldChest, HoldHandLeft, HoldHandRight, HoldRingLeft, HoldRingRight, HoldAmulet}
	for _, hold := range order {
		loc, _ := BodySlot(hold)
		item := p.InvBody[loc]

		if !Identifiable(item) {
			continue
		}

		holds = append(holds, identifyHold(item, hold))
	}

	for i := 0; i < p.NumInv && len(holds) < MaxHoldItems; i++ {
		if !Identifiable(p.InvList[i]) {
			continue
		}

		holds = append(holds, identifyHold(p.InvList[i], int8(i)))
	}

	return holds
}

// identifyHold stamps the flat identify fee on one entry (stores.cpp:1095-1102).
func identifyHold(item entities.Item, index int8) Hold {
	item.Value = IdentifyPrice
	item.Ivalue = IdentifyPrice

	return Hold{Item: item, Index: index}
}
