// Package gameplay holds the game-facing entry points the bean contract names
// for this tree. The inventory port itself lives in
// internal/diablo/inventory (AGENTS.md layer 3: Diablo item logic), so this
// package re-exports it instead of duplicating any of it — type aliases and
// function values only, no logic here.
package gameplay

import (
	"github.com/bmcszk/devilutionGo/internal/diablo/entities"
	"github.com/bmcszk/devilutionGo/internal/diablo/inventory"
)

// Inventory constants, in upstream value order.
const (
	// InventoryPitch is INV_SIZEX (inv.cpp:153).
	InventoryPitch = inventory.InventoryPitch
	// InventoryRows is INV_SIZEY (player.h:38).
	InventoryRows = inventory.InventoryRows
	// BeltSlots is MaxBeltItems (player.h:219).
	BeltSlots = inventory.BeltSlots
	// MaxGoldLimit is GOLD_MAX_LIMIT (items.h:31).
	MaxGoldLimit = inventory.MaxGoldLimit
	// InvItemFirst is INVITEM_INV_FIRST (inv.h:59).
	InvItemFirst = inventory.InvItemFirst
	// InvItemLast is INVITEM_INV_LAST (inv.h:68).
	InvItemLast = inventory.InvItemLast
	// BeltItemFirst is INVITEM_BELT_FIRST (inv.h:71).
	BeltItemFirst = inventory.BeltItemFirst
	// BeltItemLast is INVITEM_BELT_LAST (inv.h:72).
	BeltItemLast = inventory.BeltItemLast
	// LocNone is ILOC_NONE (tables/itemdat.h:99).
	LocNone = inventory.LocNone
	// LocOneHand is ILOC_ONEHAND (tables/itemdat.h:100).
	LocOneHand = inventory.LocOneHand
	// LocTwoHand is ILOC_TWOHAND (tables/itemdat.h:101).
	LocTwoHand = inventory.LocTwoHand
	// LocArmor is ILOC_ARMOR (tables/itemdat.h:102).
	LocArmor = inventory.LocArmor
	// LocHelm is ILOC_HELM (tables/itemdat.h:103).
	LocHelm = inventory.LocHelm
	// LocRing is ILOC_RING (tables/itemdat.h:104).
	LocRing = inventory.LocRing
	// LocAmulet is ILOC_AMULET (tables/itemdat.h:105).
	LocAmulet = inventory.LocAmulet
	// LocUnequippable is ILOC_UNEQUIPABLE (tables/itemdat.h:106).
	LocUnequippable = inventory.LocUnequippable
	// LocBelt is ILOC_BELT (tables/itemdat.h:107).
	LocBelt = inventory.LocBelt
	// ItemTypeMisc is ItemType::Misc (tables/itemdat.h:290).
	ItemTypeMisc = inventory.ItemTypeMisc
	// ItemTypeLightArmor is ItemType::LightArmor (tables/itemdat.h:299).
	ItemTypeLightArmor = inventory.ItemTypeLightArmor
	// SlotHead is SLOTXY_HEAD (inv.h:48).
	SlotHead = inventory.SlotHead
	// SlotRingLeft is SLOTXY_RING_LEFT (inv.h:50).
	SlotRingLeft = inventory.SlotRingLeft
	// SlotRingRight is SLOTXY_RING_RIGHT (inv.h:51).
	SlotRingRight = inventory.SlotRingRight
	// SlotAmulet is SLOTXY_AMULET (inv.h:52).
	SlotAmulet = inventory.SlotAmulet
	// SlotHandLeft is SLOTXY_HAND_LEFT (inv.h:53).
	SlotHandLeft = inventory.SlotHandLeft
	// SlotHandRight is SLOTXY_HAND_RIGHT (inv.h:54).
	SlotHandRight = inventory.SlotHandRight
	// SlotChest is SLOTXY_CHEST (inv.h:55).
	SlotChest = inventory.SlotChest
)

// Inventory entry points, re-exported from internal/diablo/inventory.
var (
	// New builds the view over a player's inventory arrays.
	New = inventory.New
	// ItemSize is GetInventorySize (inv.cpp:2294-2298).
	ItemSize = inventory.ItemSize
	// IsEmpty is Item::isEmpty (items.h:390).
	IsEmpty = inventory.IsEmpty
	// EmptyItem is a default-constructed Item (ItemType::None).
	EmptyItem = inventory.EmptyItem
	// IsEquipment is Item::isEquipment (items.h:288-307).
	IsEquipment = inventory.IsEquipment
	// CanEquip is CanEquip (inv.cpp:185-198).
	CanEquip = inventory.CanEquip
	// EquipTypeForSlot is GetItemEquipType (inv.cpp:543-559).
	EquipTypeForSlot = inventory.EquipTypeForSlot
	// IsBeltItem is FitsInBeltSlot (inv.cpp:174-177).
	IsBeltItem = inventory.IsBeltItem
	// IsScroll is Item::isScroll (items.h:392-395).
	IsScroll = inventory.IsScroll
	// IsRune is Item::isRune (items.h:401-404).
	IsRune = inventory.IsRune
	// GoldCursor is GetGoldCursor (inv.cpp:1040-1049).
	GoldCursor = inventory.GoldCursor
	// SortItemsBySize is SortItemsBySize (inv.cpp:1417-1437).
	SortItemsBySize = inventory.SortItemsBySize
	// AutoEquipItem is AutoEquip (inv.cpp:283-300).
	AutoEquipItem = func(v inventory.Inv, item entities.Item) bool { return v.AutoEquip(item) }
)

// Inventory types, aliased so both import paths name the same types.
type (
	// Inv is the inventory view over a player (inventory.Inv).
	Inv = inventory.Inv
	// Data is the injected upstream state (inventory.Data).
	Data = inventory.Data
	// Size is the item footprint in grid cells (inventory.Size).
	Size = inventory.Size
	// ItemType is the item class (inventory.ItemType).
	ItemType = inventory.ItemType
	// ItemLoc is item_equip_type (inventory.ItemLoc).
	ItemLoc = inventory.ItemLoc
	// BodyLoc is inv_body_loc (inventory.BodyLoc).
	BodyLoc = inventory.BodyLoc
	// CursorSpriteSize resolves a cursor graphic to its pixel size.
	CursorSpriteSize = inventory.CursorSpriteSize
	// ItemPredicate is the item test the search helpers run.
	ItemPredicate = inventory.ItemPredicate
)
