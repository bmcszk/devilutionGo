// Package inventory ports the player inventory core from inv.cpp: 40-cell
// grid packing, the 8-slot belt, gold stacks and the cursor hold-item
// machine. It is Diablo-specific item logic (AGENTS.md layer 3) and depends
// on the entity data port only — no renderer, UI, audio or network imports.
//
// Upstream's network sends (NetSendCmdChInvItem and friends) and UI redraw
// calls are dropped here: they are side effects of the game loop, not
// inventory state. Every ported function cites its upstream line.
package inventory

import "github.com/bmcszk/devilutionGo/internal/diablo/entities"

const (
	// InventorySlotSizeInPixels is the size of one inventory cell (inv.cpp:2294).
	InventorySlotSizeInPixels = 32
	// InventoryPitch is the InvGrid row pitch (inv.cpp:153).
	InventoryPitch = 10
	// InventoryRows is the InvGrid row count (player.h:38: 10x4 cells).
	InventoryRows = entities.InventoryGridCells / InventoryPitch
	// BeltSlots is MaxBeltItems: 4 visible slots plus 4 extended ones
	// (player.h:39, 219).
	BeltSlots = entities.MaxBeltItems
	// MaxGoldLimit is GOLD_MAX_LIMIT (items.h:31). The runtime MaxGold global
	// doubles it for the Auric Amulet (enums.GoldMaxLimitAuric).
	MaxGoldLimit = 5000
)

// The upstream ItemType values in declaration order (tables/itemdat.h:290-306).
const (
	ItemTypeMisc ItemType = iota
	ItemTypeSword
	ItemTypeAxe
	ItemTypeBow
	ItemTypeMace
	ItemTypeShield
	ItemTypeLightArmor
	ItemTypeHelm
	ItemTypeMediumArmor
	ItemTypeHeavyArmor
	ItemTypeStaff
	ItemTypeGold
	ItemTypeRing
	ItemTypeAmulet
	// ItemTypeNone marks an empty item slot (tables/itemdat.h:305).
	ItemTypeNone ItemType = -1
)

// ItemType mirrors ItemType (tables/itemdat.h:290-306); int8 per the _itype
// field width (items.h:193).
type ItemType int8

// Size is an item footprint or grid extent in inventory cells (engine Size).
type Size struct {
	Width  int
	Height int
}

// CursorSpriteSize resolves a cursor graphic to its pixel size; the asset
// layer supplies it (cursor.cpp:476 GetInvItemSize reads pCursCels).
type CursorSpriteSize func(curs uint8) Size

// Data carries the table lookups inv.cpp reaches through globals. Both hooks
// are required: a nil Sprite makes ItemSize report a 0x0 footprint and a nil
// Usable makes every item unusable, so the wiring layer must fill them.
type Data struct {
	// Sprite is pCursCels (cursor.cpp:476).
	Sprite CursorSpriteSize
	// Usable is AllItemsList[idIdx].iUsable (items.cpp:4810).
	Usable func(idIdx int16) bool
	// MaxGold is the runtime MaxGold global (items.cpp:592); zero means
	// MaxGoldLimit.
	MaxGold int
}

// Inv is a view over the inventory arrays of a Player (player.h:217-220,298).
// It owns no state: the caller keeps the player, so the port is testable
// without constructing a full Player (the C++ globals become injected inputs).
type Inv struct {
	body   *[entities.NumInvLoc]entities.Item
	list   *[entities.InventoryGridCells]entities.Item
	grid   *[entities.InventoryGridCells]int8
	belt   *[entities.MaxBeltItems]entities.Item
	hold   *entities.Item
	numInv *int
	player *entities.Player
	data   Data
}

// New binds an inventory view to a player's inventory fields.
func New(p *entities.Player, data Data) Inv {
	return Inv{
		body:   &p.InvBody,
		list:   &p.InvList,
		grid:   &p.InvGrid,
		belt:   &p.SpdList,
		hold:   &p.HoldItem,
		numInv: &p.NumInv,
		player: p,
		data:   data,
	}
}

// IsEmpty ports Item::isEmpty (items.h:283-286).
func IsEmpty(item entities.Item) bool {
	return ItemType(item.Itype) == ItemTypeNone
}

// EmptyItem is the port's empty item. C++ default-constructs Item with
// _itype = ItemType::None (items.h:194), which the Go zero value cannot
// express, so the player-construction layer must seed its slots with
// EmptyItem.
func EmptyItem() entities.Item {
	return entities.Item{Itype: int8(ItemTypeNone)}
}

// ItemSize ports GetInventorySize (inv.cpp:2292-2297): the item's cursor
// sprite divided by the inventory slot size.
func ItemSize(item entities.Item, sprite CursorSpriteSize) Size {
	px := sprite(item.Curs)
	return Size{
		Width:  px.Width / InventorySlotSizeInPixels,
		Height: px.Height / InventorySlotSizeInPixels,
	}
}

// clearItem ports Item::clear (items.h:274-277).
func clearItem(item *entities.Item) {
	item.Itype = int8(ItemTypeNone)
}

// pop ports Item::pop (items.h:264-269): return the item and empty the slot.
func pop(item *entities.Item) entities.Item {
	old := *item
	clearItem(item)
	return old
}

// maxGold resolves Data.MaxGold with the base-game default.
func (d Data) maxGold() int {
	if d.MaxGold == 0 {
		return MaxGoldLimit
	}
	return d.MaxGold
}

// sizeOf is ItemSize through the view's own sprite table.
func (v Inv) sizeOf(item entities.Item) Size {
	return ItemSize(item, v.data.Sprite)
}
