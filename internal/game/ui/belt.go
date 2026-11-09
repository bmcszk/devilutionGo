package ui

// The belt module ports DrawInvBelt (Source/inv.cpp:1272-1317) over the panel
// geometry in control_panel.cpp:99-101: one backdrop copy of the panel's belt
// strip, then a slot background and the potion art per occupied slot.

// BeltSlotCount is MaxBeltItems and BeltItems, the belt width in slots
// (player.h:40, control_panel.cpp:99).
const BeltSlotCount = 8

// InventorySlotHeight is INV_SLOT_SIZE_PX, the belt grid step (inv.h:20).
const InventorySlotHeight = 28

// BeltSlotWidth is the drawn slot width, one pixel wider than the grid step
// (inv.cpp:122, control_panel.cpp:100). It is also the per-slot x step, the
// gutter included (control_panel.cpp:100).
const BeltSlotWidth = InventorySlotHeight + 1

// NoBeltHighlight is the highlight input of a cursor that hovers no belt slot.
const NoBeltHighlight = -1

// The palette bases the slot-background shift is measured from
// (engine/palette.h:35-40).
const (
	pal16Beige  = 160
	pal16Blue   = 176
	pal16Yellow = 192
	pal16Gray   = 240
)

// The item quality classes (item_quality, items.h:42-46).
const (
	ItemQualityNormal ItemQuality = iota
	ItemQualityMagic
	ItemQualityUnique
)

// BeltRect is the belt strip on the main panel (control_panel.cpp:101).
var BeltRect = Rect{
	Position: Point{X: 205, Y: 5},
	Size:     Size{Width: BeltSlotWidth * BeltSlotCount, Height: InventorySlotHeight},
}

// BeltSlots are the belt cell positions, InvRect[SLOTXY_BELT_FIRST..]
// (inv.cpp:122-130, inv.h:71).
var BeltSlots = func() [BeltSlotCount]Rect {
	var slots [BeltSlotCount]Rect

	for slot := range slots {
		slots[slot] = Rect{
			Position: Point{X: BeltRect.Position.X + slot*BeltSlotWidth, Y: BeltRect.Position.Y},
			Size:     Size{Width: BeltSlotWidth, Height: BeltSlotWidth},
		}
	}

	return slots
}()

// ItemQuality is item_quality, the colour class the slot background shifts by
// (items.h:42-46).
type ItemQuality uint8

// BeltItem is one belt slot's contents. Items themselves are the inventory
// bean's; this layer only needs the sprite and the quality tint.
type BeltItem struct {
	Present bool
	Sprite  int
	Quality ItemQuality
}

// BeltContents is the eight belt slots (Player::SpdList, player.h:40).
type BeltContents [BeltSlotCount]BeltItem

// BeltSlotRect returns the belt-local rect of one slot. A slot past the last
// addresses no rect, as upstream's array bound does.
func BeltSlotRect(slot int) Rect {
	if slot < 0 || slot >= BeltSlotCount {
		return Rect{}
	}

	return BeltSlots[slot]
}

// QualityShift is the amount InvDrawSlotBack subtracts from every background
// index at or above PAL16_GRAY for an item of this quality, inspecting nothing
// (inv.cpp:1136-1147). Every class but magic and unique takes the beige shift.
func QualityShift(quality ItemQuality) uint8 {
	switch quality {
	case ItemQualityMagic:
		return pal16Gray - pal16Blue - 1
	case ItemQualityUnique:
		return pal16Gray - pal16Yellow - 1
	case ItemQualityNormal:
		return pal16Gray - pal16Beige - 1
	}

	return pal16Gray - pal16Beige - 1
}

// DrawBelt emits the belt: the panel's own strip copied back over itself, then
// each occupied slot's background and potion art (DrawInvBelt,
// inv.cpp:1272-1317).
func DrawBelt(sink PanelSink, areas Areas, belt BeltContents, highlight int) {
	sink.DrawPanelBox(PanelBox{
		Buffer: BufferPanel,
		Source: Rect{
			Position: Point{X: BeltRect.Position.X, Y: BeltRect.Position.Y + PanelPaddingHeight},
			Size:     BeltRect.Size,
		},
		Target: Point{
			X: areas.Main.Position.X + BeltRect.Position.X,
			Y: areas.Main.Position.Y + BeltRect.Position.Y,
		},
		Flavour: BlitCopy,
	})

	for slot, item := range belt {
		if !item.Present {
			continue
		}

		drawBeltSlot(sink, areas, slot, item, highlight)
	}
}

// drawBeltSlot emits one occupied slot: its quality-tinted background, its art
// and, for the hovered slot, the outline (inv.cpp:1286-1300).
func drawBeltSlot(sink PanelSink, areas Areas, slot int, item BeltItem, highlight int) {
	// The item grid row sits one slot height below the belt strip the backdrop
	// copies (inv.cpp:1288).
	cell := BeltSlotRect(slot).Shift(areas.Main.Position).Shift(Point{X: 0, Y: InventorySlotHeight})
	art := Rect{Position: Point{X: 0, Y: 0}, Size: Size{Width: BeltSlotWidth, Height: BeltSlotWidth}}

	// InvDrawSlotBack recolours the pixels already on the target, so the tint
	// box carries a size and no source buffer (inv.cpp:1129-1159).
	sink.DrawPanelBox(PanelBox{
		Source:  Rect{Size: cell.Size},
		Target:  cell.Position,
		Flavour: BlitTint,
		Quality: item.Quality,
	})

	sink.DrawPanelBox(PanelBox{Buffer: BufferCursor, Source: art, Target: cell.Position, Flavour: BlitCopy})

	if slot != highlight {
		return
	}

	sink.DrawPanelBox(PanelBox{
		Buffer:  BufferCursor,
		Source:  art,
		Target:  cell.Position,
		Flavour: BlitSkipZero,
		Quality: item.Quality,
	})
}
