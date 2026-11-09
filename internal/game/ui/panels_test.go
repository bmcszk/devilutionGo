// Package ui_test proves the control-panel chrome: the panel areas the layout
// rule computes, the main-panel button rects, the panel-box draws and the
// side-panel state machine (Source/control/control_panel.cpp).
package ui_test

import (
	"fmt"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game/ui"
	"github.com/bmcszk/devilutionGo/internal/render"
)

const (
	// screenWidth and screenHeight are the 640x480 world logical surface the
	// panel boxes land on (docs: G10 640x480 world surface).
	screenWidth  = 640
	screenHeight = 480
	// pal16GrayIndex is PAL16_GRAY, the first background index InvDrawSlotBack
	// shifts (engine/palette.h:40).
	pal16GrayIndex = 240
)

// mainOrigin is the main panel position on a 640x480 screen
// (control_panel.cpp:245-248).
var mainOrigin = ui.Point{X: 0, Y: 352}

// recordSink keeps every panel box in emission order.
type recordSink struct{ boxes []ui.PanelBox }

// DrawPanelBox appends the box to the recording.
func (s *recordSink) DrawPanelBox(box ui.PanelBox) { s.boxes = append(s.boxes, box) }

func TestCalculatePanelAreasMatchesUpstreamLayout(t *testing.T) {
	// when
	areas := ui.CalculatePanelAreas(640, 480, ui.ControlsKeyboardAndMouse)

	// then
	expectRect(t, "main", areas.Main, ui.Rect{
		Position: mainOrigin,
		Size:     ui.Size{Width: 640, Height: 128},
	})
	expectRect(t, "left", areas.Left, ui.Rect{
		Position: ui.Point{X: 0, Y: 0},
		Size:     ui.Size{Width: 320, Height: 352},
	})
	expectRect(t, "right", areas.Right, ui.Rect{
		Position: ui.Point{X: 320, Y: 0},
		Size:     ui.Size{Width: 320, Height: 352},
	})
}

func TestCalculatePanelAreasCentresSidePanelsOnWideScreens(t *testing.T) {
	// when: 800px leaves room for both side panels beside the main panel, so the
	// left panel shifts right by half the surplus (control_panel.cpp:261-263)
	areas := ui.CalculatePanelAreas(800, 480, ui.ControlsKeyboardAndMouse)

	// then
	expectPoint(t, "main", areas.Main.Position, ui.Point{X: 80, Y: 352})
	expectPoint(t, "left", areas.Left.Position, ui.Point{X: 0, Y: 0})
	expectPoint(t, "right", areas.Right.Position, ui.Point{X: 480, Y: 0})
}

func TestCalculatePanelAreasLaysOutVirtualGamepad(t *testing.T) {
	// when: the virtual gamepad pins the side panels to the screen centre
	// (control_panel.cpp:258-269)
	areas := ui.CalculatePanelAreas(800, 480, ui.ControlsVirtualGamepad)

	// then
	expectPoint(t, "left", areas.Left.Position, ui.Point{X: 80, Y: 0})
	expectPoint(t, "right", areas.Right.Position, ui.Point{X: 400, Y: 0})
}

func TestPanelPositionRoutesByPanel(t *testing.T) {
	// given
	areas := ui.CalculatePanelAreas(640, 480, ui.ControlsKeyboardAndMouse)
	offset := ui.Point{X: 7, Y: 11}

	// when
	got := map[ui.Panel]ui.Point{
		ui.PanelMain:      ui.PanelPosition(areas, ui.PanelMain, offset),
		ui.PanelQuest:     ui.PanelPosition(areas, ui.PanelQuest, offset),
		ui.PanelCharacter: ui.PanelPosition(areas, ui.PanelCharacter, offset),
		ui.PanelStash:     ui.PanelPosition(areas, ui.PanelStash, offset),
		ui.PanelInventory: ui.PanelPosition(areas, ui.PanelInventory, offset),
		ui.PanelSpell:     ui.PanelPosition(areas, ui.PanelSpell, offset),
	}

	// then: main, the left-side panels and the right-side panels each take their
	// own origin (control_panel.cpp:332-350)
	expectPoint(t, "main", got[ui.PanelMain], ui.Point{X: 7, Y: 363})
	expectPoint(t, "quest log", got[ui.PanelQuest], ui.Point{X: 7, Y: 11})
	expectPoint(t, "character", got[ui.PanelCharacter], ui.Point{X: 7, Y: 11})
	expectPoint(t, "stash", got[ui.PanelStash], ui.Point{X: 7, Y: 11})
	expectPoint(t, "inventory", got[ui.PanelInventory], ui.Point{X: 327, Y: 11})
	expectPoint(t, "spellbook", got[ui.PanelSpell], ui.Point{X: 327, Y: 11})
}

func TestMainPanelButtonRectsMatchUpstream(t *testing.T) {
	// given: MainPanelButtonRect (control_panel.cpp:84-96), wide buttons 71x20
	// and the two small 33x32 ones (:81-82)
	want := []ui.Rect{
		{Position: ui.Point{X: 9, Y: 9}, Size: ui.Size{Width: 71, Height: 20}},
		{Position: ui.Point{X: 9, Y: 35}, Size: ui.Size{Width: 71, Height: 20}},
		{Position: ui.Point{X: 9, Y: 75}, Size: ui.Size{Width: 71, Height: 20}},
		{Position: ui.Point{X: 9, Y: 101}, Size: ui.Size{Width: 71, Height: 20}},
		{Position: ui.Point{X: 560, Y: 9}, Size: ui.Size{Width: 71, Height: 20}},
		{Position: ui.Point{X: 560, Y: 35}, Size: ui.Size{Width: 71, Height: 20}},
		{Position: ui.Point{X: 87, Y: 91}, Size: ui.Size{Width: 33, Height: 32}},
		{Position: ui.Point{X: 527, Y: 91}, Size: ui.Size{Width: 33, Height: 32}},
	}

	// when
	got := make([]ui.Rect, ui.MainPanelButtonCount)
	for button := range got {
		got[button] = ui.MainPanelButtonRect(ui.MainPanelButton(button))
	}

	// then
	for button := range got {
		id := ui.MainPanelButton(button)
		expectRect(t, id.String(), got[button], want[button])
	}
}

func TestMainPanelButtonRectPastLastIsZero(t *testing.T) {
	// when: an id past PanelButtonLast addresses no rect (control_panel.cpp:131-142)
	rect := ui.MainPanelButtonRect(ui.MainPanelButtonFriendly + 1)

	// then
	expectRect(t, "unknown button", rect, ui.Rect{})
}

func TestDrawMainPanelCopiesPanelChrome(t *testing.T) {
	// given
	sink := &recordSink{}
	areas := ui.CalculatePanelAreas(640, 480, ui.ControlsKeyboardAndMouse)

	// when
	ui.DrawMainPanel(sink, areas)

	// then: the panel chrome sits one padding row into the composed buffer
	// (control_panel.cpp:435)
	if len(sink.boxes) != 1 {
		t.Fatalf("DrawMainPanel emitted %d boxes, want 1", len(sink.boxes))
	}

	expectRect(t, "chrome source", sink.boxes[0].Source, ui.Rect{
		Position: ui.Point{X: 0, Y: ui.PanelPaddingHeight},
		Size:     ui.Size{Width: 640, Height: 128},
	})
	expectPoint(t, "chrome target", sink.boxes[0].Target, mainOrigin)
}

func TestDrawMainPanelButtonsCopiesEveryButtonBox(t *testing.T) {
	// given
	areas := ui.CalculatePanelAreas(640, 480, ui.ControlsKeyboardAndMouse)
	sink := &recordSink{}

	// when
	ui.DrawMainPanelButtons(sink, areas, ui.PanelButtonStates{})

	// then: one box per single-player button, each at its own rect with the
	// source one padding row deeper in the buffer (control_panel.cpp:445)
	want := ui.SinglePlayerMainPanelButtons
	if len(sink.boxes) != want {
		t.Fatalf("emitted %d boxes, want %d", len(sink.boxes), want)
	}

	for button := range want {
		rect := ui.MainPanelButtonRect(ui.MainPanelButton(button))

		// the source rect is one row deeper and one row taller than the button
		// (control_panel.cpp:445)
		source := rect.Shift(ui.Point{X: 0, Y: ui.PanelPaddingHeight})
		source.Size.Height++

		expectRect(t, ui.MainPanelButton(button).String()+" source", sink.boxes[button].Source, source)
		expectPoint(t, ui.MainPanelButton(button).String()+" target", sink.boxes[button].Target,
			rect.Shift(mainOrigin).Position)
	}
}

func TestDrawMainPanelButtonsPaintsButtonBoxesInTheIndexBuffer(t *testing.T) {
	// given: a solid panel buffer of index 137, so every copied button box shows
	// as that index in the output
	sources := map[ui.PanelBuffer]render.Surface{
		ui.BufferPanel: solidSurface(ui.MainPanelWidth, ui.MainPanelHeight+ui.PanelPaddingHeight, 137),
	}
	sink := &recordSink{}
	areas := ui.CalculatePanelAreas(640, 480, ui.ControlsKeyboardAndMouse)

	// when
	ui.DrawMainPanelButtons(sink, areas, ui.PanelButtonStates{})

	// then
	out := paintBoxes(t, sink.boxes, sources)
	rect := ui.MainPanelButtonRect(ui.MainPanelButtonCharinfo).Shift(mainOrigin)
	expectIndexes(t, out, rect, 137)
}

func TestDrawMainPanelButtonsSkipsPressedButton(t *testing.T) {
	// given: the character button is held down, so upstream draws the pressed
	// sprite instead of the panel box (control_panel.cpp:442-446)
	areas := ui.CalculatePanelAreas(640, 480, ui.ControlsKeyboardAndMouse)
	pressed := ui.PanelButtonStates{}
	pressed[ui.MainPanelButtonCharinfo] = true
	sink := &recordSink{}

	// when
	ui.DrawMainPanelButtons(sink, areas, pressed)

	// then: one box per remaining button, and none of them covers the held one
	want := ui.SinglePlayerMainPanelButtons - 1
	if len(sink.boxes) != want {
		t.Fatalf("emitted %d boxes, want %d", len(sink.boxes), want)
	}

	held := ui.MainPanelButtonRect(ui.MainPanelButtonCharinfo).Shift(mainOrigin)
	for _, box := range sink.boxes {
		if box.Target == held.Position && box.Source.Size == held.Size {
			t.Fatalf("pressed button still emitted its box at %v", box.Target)
		}
	}
}

func TestToggleOpensAndClosesEachPanel(t *testing.T) {
	for _, panel := range []ui.Panel{ui.PanelCharacter, ui.PanelQuest, ui.PanelInventory, ui.PanelSpell} {
		// given
		state := ui.PanelState{}

		// when
		state.Toggle(panel)
		openedAfterFirst := state.IsOpen(panel)
		state.Toggle(panel)
		openedAfterSecond := state.IsOpen(panel)

		// then
		if !openedAfterFirst || openedAfterSecond {
			t.Fatalf("Toggle(%v): open after first=%v, after second=%v, want true then false",
				panel, openedAfterFirst, openedAfterSecond)
		}
	}
}

func TestOpeningInventoryClosesSpellbook(t *testing.T) {
	// given
	state := ui.PanelState{}
	state.Toggle(ui.PanelSpell)

	// when
	state.Toggle(ui.PanelInventory)

	// then: both right-side panels share one slot (diablo.cpp:1644-1664)
	expectOpen(t, state, ui.PanelInventory)
	expectClosed(t, state, ui.PanelSpell)
}

func TestOpeningSpellbookClosesInventory(t *testing.T) {
	// given
	state := ui.PanelState{}
	state.Toggle(ui.PanelInventory)

	// when
	state.Toggle(ui.PanelSpell)

	// then (diablo.cpp:1731-1748)
	expectOpen(t, state, ui.PanelSpell)
	expectClosed(t, state, ui.PanelInventory)
}

func TestOpeningCharacterClosesQuestLog(t *testing.T) {
	// given
	state := ui.PanelState{}
	state.Toggle(ui.PanelQuest)

	// when
	state.Toggle(ui.PanelCharacter)

	// then: the left-side panels are mutually exclusive too (diablo.cpp:1666-1682)
	expectOpen(t, state, ui.PanelCharacter)
	expectClosed(t, state, ui.PanelQuest)
}

func TestOpeningQuestLogClosesCharacter(t *testing.T) {
	// given
	state := ui.PanelState{}
	state.Toggle(ui.PanelCharacter)

	// when
	state.Toggle(ui.PanelQuest)

	// then (diablo.cpp:1689-1713)
	expectOpen(t, state, ui.PanelQuest)
	expectClosed(t, state, ui.PanelCharacter)
}

func TestInventoryDoesNotCloseCharacterSheet(t *testing.T) {
	// given
	state := ui.PanelState{}
	state.Toggle(ui.PanelCharacter)

	// when: the two sides are independent (test/panel_state_test.cpp,
	// InventoryDoesNotCloseCharSheet)
	state.Toggle(ui.PanelInventory)

	// then
	expectOpen(t, state, ui.PanelInventory)
	expectOpen(t, state, ui.PanelCharacter)
}

func TestSidePanelsReportOpenSides(t *testing.T) {
	for _, row := range []struct {
		panel      ui.Panel
		left, rite bool
	}{
		{panel: ui.PanelCharacter, left: true},
		{panel: ui.PanelQuest, left: true},
		{panel: ui.PanelInventory, rite: true},
		{panel: ui.PanelSpell, rite: true},
	} {
		// given
		state := ui.PanelState{}

		// when
		state.Toggle(row.panel)

		// then: IsLeftPanelOpen/IsRightPanelOpen (control_panel.cpp:63-70)
		if state.LeftOpen() != row.left || state.RightOpen() != row.rite {
			t.Fatalf("open %v: left=%v right=%v, want left=%v right=%v",
				row.panel, state.LeftOpen(), state.RightOpen(), row.left, row.rite)
		}
	}
}

func TestOpeningStashAlsoShowsInventory(t *testing.T) {
	// given
	state := ui.PanelState{}

	// when: receiving the stash opens the inventory beside it
	// (stores.cpp:2032-2034)
	state.Toggle(ui.PanelStash)

	// then
	expectOpen(t, state, ui.PanelStash)
	expectOpen(t, state, ui.PanelInventory)
}

func TestCloseInventoryClosesStashToo(t *testing.T) {
	// given
	state := ui.PanelState{}
	state.Toggle(ui.PanelStash)

	// when: CloseInventory also closes the stash (inv.cpp:2231-2237)
	state.Close(ui.PanelInventory)

	// then
	expectClosed(t, state, ui.PanelInventory)
	expectClosed(t, state, ui.PanelStash)
}

// paintBoxes replays boxes onto a fresh 640x480 world surface, honouring
// each box's blit flavour.
func paintBoxes(t *testing.T, boxes []ui.PanelBox, sources map[ui.PanelBuffer]render.Surface) render.Surface {
	t.Helper()

	out := render.NewSurface(screenWidth, screenHeight)

	for _, box := range boxes {
		replayBox(out, box, sources)
	}

	return out
}

// paintBoxesOn replays boxes onto a given surface, honouring each blit flavour.
func paintBoxesOn(out render.Surface, boxes []ui.PanelBox, sources map[ui.PanelBuffer]render.Surface) render.Surface {
	for _, box := range boxes {
		replayBox(out, box, sources)
	}

	return out
}

// replayBox applies one panel box to out the way the engine painter does:
// a copy or a zero-keyed copy of the source rect, or an in-place tint.
func replayBox(out render.Surface, box ui.PanelBox, sources map[ui.PanelBuffer]render.Surface) {
	if box.Flavour == ui.BlitTint {
		tintRect(out, ui.Rect{Position: box.Target, Size: box.Source.Size}, box.Quality)

		return
	}

	source := cropSurface(sources[box.Buffer], box.Source)

	if box.Flavour == ui.BlitSkipZero {
		keepUnderZeroed(out, box, source)
	}

	out.DrawDirect(render.Clip{
		X:      box.Target.X,
		Y:      box.Target.Y,
		Width:  box.Source.Size.Width,
		Height: box.Source.Size.Height,
		Pixels: source,
	})
}

// keepUnderZeroed replaces every zero source index with what is already at the
// target, the SkipColorIndexZero blit (engine/surface.cpp:26-33).
func keepUnderZeroed(out render.Surface, box ui.PanelBox, source []byte) {
	for i, index := range source {
		if index != 0 {
			continue
		}

		source[i] = out.At(box.Target.X+i%box.Source.Size.Width, box.Target.Y+i/box.Source.Size.Width)
	}
}

// tintRect applies InvDrawSlotBack's shift in place (inv.cpp:1149-1158).
func tintRect(out render.Surface, rect ui.Rect, quality ui.ItemQuality) {
	for y := rect.Position.Y; y < rect.Position.Y+rect.Size.Height; y++ {
		for x := rect.Position.X; x < rect.Position.X+rect.Size.Width; x++ {
			index := out.At(x, y)
			if index >= pal16GrayIndex {
				out.Set(x, y, index-ui.QualityShift(quality))
			}
		}
	}
}

// cropSurface cuts rect out of surface as a standalone sprite. Pixels outside
// the source surface read as 0, so a box pointing at a wrong buffer shows as an
// index mismatch instead of panicking.
func cropSurface(surface render.Surface, rect ui.Rect) []byte {
	pixels := make([]byte, 0, rect.Size.Width*rect.Size.Height)

	for y := range rect.Size.Height {
		for x := range rect.Size.Width {
			pixels = append(pixels, pixelAt(surface, rect.Position.X+x, rect.Position.Y+y))
		}
	}

	return pixels
}

// pixelAt reads surface at (x,y), 0 when the coordinate falls outside it.
func pixelAt(surface render.Surface, x, y int) byte {
	if x < 0 || y < 0 || x >= surface.Width || y >= surface.Height {
		return 0
	}

	return surface.At(x, y)
}

// solidSurface builds a width×height surface of one index.
func solidSurface(width, height int, index byte) render.Surface {
	surface := render.NewSurface(width, height)
	for i := range surface.Pixels {
		surface.Pixels[i] = index
	}

	return surface
}

func expectRect(t *testing.T, name string, got, want ui.Rect) {
	t.Helper()

	if got != want {
		t.Fatalf("%s rect = %s, want %s", name, describeRect(got), describeRect(want))
	}
}

func expectPoint(t *testing.T, name string, got, want ui.Point) {
	t.Helper()

	if got != want {
		t.Fatalf("%s point = {%d,%d}, want {%d,%d}", name, got.X, got.Y, want.X, want.Y)
	}
}

// expectIndexes asserts that every pixel of rect on surface holds index.
func expectIndexes(t *testing.T, surface render.Surface, rect ui.Rect, index byte) {
	t.Helper()

	for y := rect.Position.Y; y < rect.Position.Y+rect.Size.Height; y++ {
		for x := rect.Position.X; x < rect.Position.X+rect.Size.Width; x++ {
			if got := surface.At(x, y); got != index {
				t.Fatalf("index at (%d,%d) = %d, want %d in rect %s", x, y, got, index, describeRect(rect))
			}
		}
	}
}

func expectOpen(t *testing.T, state ui.PanelState, panel ui.Panel) {
	t.Helper()

	if !state.IsOpen(panel) {
		t.Fatalf("IsOpen(%v) = false, want true", panel)
	}
}

func expectClosed(t *testing.T, state ui.PanelState, panel ui.Panel) {
	t.Helper()

	if state.IsOpen(panel) {
		t.Fatalf("IsOpen(%v) = true, want false", panel)
	}
}

func describeRect(rect ui.Rect) string {
	return fmt.Sprintf("{%d,%d %dx%d}", rect.Position.X, rect.Position.Y, rect.Size.Width, rect.Size.Height)
}
