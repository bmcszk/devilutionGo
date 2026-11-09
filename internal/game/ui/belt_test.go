// Package ui_test proves the belt panel: the eight slot rects and the potion
// art placement (Source/control/control_panel.cpp:99-101, Source/inv.cpp:1272-1300).
package ui_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game/ui"
	"github.com/bmcszk/devilutionGo/internal/render"
)

func TestBeltRectMatchesUpstreamGeometry(t *testing.T) {
	// when: BeltRect is 8 slots of (INV_SLOT_SIZE_PX+1) x INV_SLOT_SIZE_PX at
	// (205,5) (control_panel.cpp:99-101, inv.h:20)
	belt := ui.BeltRect

	// then
	expectRect(t, "belt", belt, ui.Rect{
		Position: ui.Point{X: 205, Y: 5},
		Size:     ui.Size{Width: 232, Height: 28},
	})
}

func TestBeltSlotsAreOneGridStepApart(t *testing.T) {
	// given: the belt slot positions are the eight InvRect belt entries
	// (inv.cpp:122-130, SLOTXY_BELT_FIRST = 47 in inv.h:71)
	want := []ui.Point{
		{X: 205, Y: 5},
		{X: 234, Y: 5},
		{X: 263, Y: 5},
		{X: 292, Y: 5},
		{X: 321, Y: 5},
		{X: 350, Y: 5},
		{X: 379, Y: 5},
		{X: 408, Y: 5},
	}

	// when
	got := make([]ui.Rect, len(ui.BeltSlots))
	for slot := range ui.BeltSlots {
		got[slot] = ui.BeltSlotRect(slot)
	}

	// then
	if len(got) != len(want) {
		t.Fatalf("BeltSlots len = %d, want %d", len(got), len(want))
	}

	for slot, position := range want {
		expectRect(t, "belt slot", got[slot], ui.Rect{
			Position: position,
			Size:     ui.Size{Width: 29, Height: 29},
		})
	}
}

func TestBeltSlotRectPastLastIsZero(t *testing.T) {
	// when: MaxBeltItems is 8, so slot 8 addresses no rect (player.h:40)
	rect := ui.BeltSlotRect(ui.BeltSlotCount)

	// then
	expectRect(t, "unknown slot", rect, ui.Rect{})
}

func TestDrawBeltCopiesTheEmptySlotBackdrop(t *testing.T) {
	// given
	sink := &recordSink{}
	areas := ui.CalculatePanelAreas(640, 480, ui.ControlsKeyboardAndMouse)

	// when: DrawInvBelt first copies the panel's own belt strip back over itself
	// (inv.cpp:1279)
	ui.DrawBelt(sink, areas, ui.BeltContents{}, ui.NoBeltHighlight)

	// then
	if len(sink.boxes) != 1 {
		t.Fatalf("empty belt emitted %d boxes, want 1", len(sink.boxes))
	}

	expectRect(t, "backdrop source", sink.boxes[0].Source, ui.Rect{
		Position: ui.Point{X: 205, Y: 21},
		Size:     ui.Size{Width: 232, Height: 28},
	})
	expectPoint(t, "backdrop target", sink.boxes[0].Target, ui.Point{
		X: areas.Main.Position.X + 205,
		Y: areas.Main.Position.Y + 5,
	})
}

func TestDrawBeltPlacesPotionArtPerOccupiedSlot(t *testing.T) {
	// given: slots 0 and 5 hold potions, the rest are empty
	sink := &recordSink{}
	areas := ui.CalculatePanelAreas(640, 480, ui.ControlsKeyboardAndMouse)
	belt := ui.BeltContents{}
	belt[0] = ui.BeltItem{Present: true, Sprite: 11, Quality: ui.ItemQualityNormal}
	belt[5] = ui.BeltItem{Present: true, Sprite: 12, Quality: ui.ItemQualityMagic}

	// when
	ui.DrawBelt(sink, areas, belt, ui.NoBeltHighlight)

	// then: one backdrop box plus one background box and one art box per potion
	if len(sink.boxes) != 5 {
		t.Fatalf("belt emitted %d boxes, want 5", len(sink.boxes))
	}

	// the item's background and art shift down by one slot height because the
	// belt strip sits above the grid row the slot rect names (inv.cpp:1288)
	occupied := []int{0, 5}
	for i, slot := range occupied {
		cell := ui.BeltSlotRect(slot).Shift(areas.Main.Position)
		background := cell.Shift(ui.Point{X: 0, Y: ui.InventorySlotHeight})

		expectPoint(t, "slot background", sink.boxes[1+i*2].Target, background.Position)
		expectPoint(t, "potion art", sink.boxes[2+i*2].Target, background.Position)
	}
}

func TestBeltItemSpriteIsItsOwnFrame(t *testing.T) {
	// given: a belt of one potion
	areas := ui.CalculatePanelAreas(640, 480, ui.ControlsKeyboardAndMouse)
	belt := ui.BeltContents{}
	belt[2] = ui.BeltItem{Present: true, Sprite: 0, Quality: ui.ItemQualityNormal}
	sink := &recordSink{}

	// when
	ui.DrawBelt(sink, areas, belt, ui.NoBeltHighlight)

	// then: the art box reads the cursor buffer, not the panel buffer
	art := sink.boxes[len(sink.boxes)-1]
	if art.Buffer != ui.BufferCursor {
		t.Fatalf("potion art buffer = %v, want BufferCursor", art.Buffer)
	}

	expectRect(t, "potion art source", art.Source, ui.Rect{
		Position: ui.Point{X: 0, Y: 0},
		Size:     ui.Size{Width: 29, Height: 29},
	})
}

func TestDrawBeltOutlinesTheHighlightedSlot(t *testing.T) {
	// given
	sink := &recordSink{}
	areas := ui.CalculatePanelAreas(640, 480, ui.ControlsKeyboardAndMouse)
	belt := ui.BeltContents{}
	belt[3] = ui.BeltItem{Present: true, Sprite: 5, Quality: ui.ItemQualityNormal}

	// when: the cursor hovers slot 3, so its art is outlined
	// (inv.cpp:1294-1298)
	ui.DrawBelt(sink, areas, belt, 3)

	// then
	outline := sink.boxes[len(sink.boxes)-1]
	if outline.Flavour != ui.BlitSkipZero {
		t.Fatalf("outline flavour = %v, want BlitSkipZero", outline.Flavour)
	}

	if outline.Quality != ui.ItemQualityNormal {
		t.Fatalf("outline quality = %v, want normal", outline.Quality)
	}
}

func TestBeltPaintingPutsPotionArtInTheIndexBuffer(t *testing.T) {
	// given: a panel buffer of index 3 and a cursor cel of index 9
	sources := map[ui.PanelBuffer]render.Surface{
		ui.BufferPanel:  solidSurface(ui.MainPanelWidth, ui.MainPanelHeight+ui.PanelPaddingHeight, 3),
		ui.BufferCursor: solidSurface(64, 64, 9),
	}
	areas := ui.CalculatePanelAreas(640, 480, ui.ControlsKeyboardAndMouse)
	belt := ui.BeltContents{}
	belt[4] = ui.BeltItem{Present: true, Sprite: 0, Quality: ui.ItemQualityNormal}
	sink := &recordSink{}

	// when
	ui.DrawBelt(sink, areas, belt, ui.NoBeltHighlight)

	// then: the art lands one slot height below the belt strip, 29px wide
	art := ui.BeltSlotRect(4).Shift(areas.Main.Position).Shift(ui.Point{X: 0, Y: ui.InventorySlotHeight})
	out := paintBoxes(t, sink.boxes, sources)
	expectIndexes(t, out, art, 9)
}

func TestBeltSlotBackgroundShiftFollowsItemQuality(t *testing.T) {
	// when: the slot background shifts by the quality's palette band
	// (inv.cpp:1136-1147)
	normal := ui.QualityShift(ui.ItemQualityNormal)
	magic := ui.QualityShift(ui.ItemQualityMagic)
	unique := ui.QualityShift(ui.ItemQualityUnique)

	// then: PAL16_GRAY minus the quality's band minus one
	if normal != 240-160-1 {
		t.Fatalf("normal shift = %d, want %d", normal, 240-160-1)
	}

	if magic != 240-176-1 {
		t.Fatalf("magic shift = %d, want %d", magic, 240-176-1)
	}

	if unique != 240-192-1 {
		t.Fatalf("unique shift = %d, want %d", unique, 240-192-1)
	}
}

func TestDrawBeltTintsEachOccupiedSlotBackground(t *testing.T) {
	// given
	sink := &recordSink{}
	areas := ui.CalculatePanelAreas(640, 480, ui.ControlsKeyboardAndMouse)
	belt := ui.BeltContents{}
	belt[0] = ui.BeltItem{Present: true, Quality: ui.ItemQualityNormal}
	belt[1] = ui.BeltItem{Present: true, Quality: ui.ItemQualityMagic}

	// when
	ui.DrawBelt(sink, areas, belt, ui.NoBeltHighlight)

	// then: one in-place tint per occupied slot, carrying that item's quality
	qualities := []ui.ItemQuality{}
	for _, box := range sink.boxes {
		if box.Flavour == ui.BlitTint {
			qualities = append(qualities, box.Quality)
		}
	}

	if len(qualities) != 2 || qualities[0] != ui.ItemQualityNormal || qualities[1] != ui.ItemQualityMagic {
		t.Fatalf("tint qualities = %v, want [normal magic]", qualities)
	}
}

func TestBeltSlotBackgroundTintsTheIndexBuffer(t *testing.T) {
	// given: a screen whose pixels already carry a background index, as the real
	// frame does after DrawMainPanel and the world pass
	sources := map[ui.PanelBuffer]render.Surface{
		ui.BufferPanel:  solidSurface(ui.MainPanelWidth, ui.MainPanelHeight+ui.PanelPaddingHeight, 250),
		ui.BufferCursor: solidSurface(64, 64, 0),
	}
	areas := ui.CalculatePanelAreas(640, 480, ui.ControlsKeyboardAndMouse)

	// when
	belt := ui.BeltContents{}
	belt[2] = ui.BeltItem{Present: true, Quality: ui.ItemQualityNormal}
	sink := &recordSink{}
	ui.DrawBelt(sink, areas, belt, ui.NoBeltHighlight)

	// then: InvDrawSlotBack subtracts the quality shift from every background
	// index it finds on the target (inv.cpp:1149-1158), so the slot cell ends up
	// shifted while the untouched screen keeps its index
	cell := ui.BeltSlotRect(2).Shift(areas.Main.Position).Shift(ui.Point{X: 0, Y: ui.InventorySlotHeight})

	tints := []ui.PanelBox{}
	for _, box := range sink.boxes {
		if box.Flavour == ui.BlitTint {
			tints = append(tints, box)
		}
	}

	out := paintBoxesOn(solidSurface(screenWidth, screenHeight, 250), tints, sources)

	expectIndexes(t, out, cell, 250-ui.QualityShift(ui.ItemQualityNormal))

	if got := out.At(0, 0); got != 250 {
		t.Fatalf("untinted pixel = %d, want 250", got)
	}
}
