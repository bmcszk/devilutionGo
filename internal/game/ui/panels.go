// Package ui is the control-panel layer: panel layout, main-panel button
// rects, flasks, belt and the side-panel state machine, ported from
// Source/control/control_panel.cpp and Source/control/control_flasks.cpp.
//
// The layer is presentation-only and sim-pure (ADR-017, scripts/simpurity): it
// computes geometry and emits PanelBox draw commands, and never touches a
// surface itself. The engine-side painter replays the boxes into the index
// framebuffer.
package ui

// Panel chrome geometry, all in main-panel-local pixels.
const (
	// MainPanelWidth and MainPanelHeight are MainPanelSize (control_panel.cpp:243).
	MainPanelWidth  = 640
	MainPanelHeight = 128
	// SidePanelWidth and SidePanelHeight are SidePanelSize (control.hpp:38).
	SidePanelWidth  = 320
	SidePanelHeight = 352
	// PanelPaddingHeight is the chat row the composed panel buffer reserves
	// below the chrome (control_panel.cpp:105).
	PanelPaddingHeight = 16
	// FlaskBufferSize is the side of pLifeBuff and pManaBuff (control_panel.cpp:353).
	FlaskBufferSize = 88
	// SinglePlayerMainPanelButtons is TotalSpMainPanelButtons
	// (control_panel.cpp:122).
	SinglePlayerMainPanelButtons = 6
	// MultiPlayerMainPanelButtons is TotalMpMainPanelButtons
	// (control_panel.cpp:123).
	MultiPlayerMainPanelButtons = 8
	// MainPanelButtonCount is PanelButtonLast + 1 (control_panel.cpp:131-142).
	MainPanelButtonCount = 8
)

// The two main-panel button sizes (control_panel.cpp:81-82).
const (
	widePanelButtonWidth  = 71
	widePanelButtonHeight = 20
	panelButtonWidth      = 33
	panelButtonHeight     = 32
)

// The panel set, in upstream enum order (UiPanels, panels/ui_panels.hpp:7-14).
const (
	PanelMain Panel = iota
	PanelQuest
	PanelCharacter
	PanelSpell
	PanelInventory
	PanelStash
)

// The buffers a panel box may read.
const (
	BufferPanel PanelBuffer = iota
	BufferFlask
	BufferCursor
)

// The panel-box blit flavours, plus the in-place recolour.
const (
	// BlitCopy is Surface::BlitFrom (engine/surface.cpp:44-47).
	BlitCopy BlitFlavour = iota
	// BlitSkipZero is Surface::BlitFromSkipColorIndexZero (engine/surface.cpp:49-52).
	BlitSkipZero
	// BlitTint recolours the target pixels in place instead of copying:
	// InvDrawSlotBack shifts every index at or above PAL16_GRAY by the slot's
	// quality (inv.cpp:1129-1159).
	BlitTint
)

// The button set, in upstream enum order (panel_button_id,
// control_panel.cpp:131-142).
const (
	MainPanelButtonCharinfo MainPanelButton = iota
	MainPanelButtonQlog
	MainPanelButtonAutomap
	MainPanelButtonMainmenu
	MainPanelButtonInventory
	MainPanelButtonSpellbook
	MainPanelButtonSendmsg
	MainPanelButtonFriendly
)

// The control schemes the panel layout distinguishes
// (ControlTypes, controls/control_mode.hpp:10-15).
const (
	ControlsNone Controls = iota
	ControlsKeyboardAndMouse
	ControlsGamepad
	ControlsVirtualGamepad
)

// Point is a target-buffer coordinate.
type Point struct{ X, Y int }

// Size is a width and height in pixels.
type Size struct{ Width, Height int }

// Rect is a positioned rectangle, the shape of upstream's Rectangle
// (engine/rectangle.hpp:9-12).
type Rect struct {
	Position Point
	Size     Size
}

// Panel names one UI panel (UiPanels, panels/ui_panels.hpp:7-14).
type Panel uint8

// PanelBuffer names the composed buffer a panel box reads from: BottomBuffer, one
// of the flask buffers, or the cursor cel a belt item draws
// (control_panel.cpp:352-366, inv.cpp:1290).
type PanelBuffer uint8

// BlitFlavour is which Surface copy, or which in-place recolour, a panel box
// uses.
type BlitFlavour uint8

// MainPanelButton names one main-panel button (panel_button_id,
// control_panel.cpp:131-142).
type MainPanelButton uint8

// Controls is the active control scheme (ControlTypes,
// controls/control_mode.hpp:10-15).
type Controls uint8

// Areas is the three panel rects the layout rule produces: MainPanel, LeftPanel
// and RightPanel (control_panel.cpp:241-279).
type Areas struct {
	Main  Rect
	Left  Rect
	Right Rect
}

// PanelBox is one panel composite draw: copy Source of Buffer to Target with
// its blit flavour, or recolour Target in place for BlitTint. It carries the
// whole operation so the painter needs no upstream state.
type PanelBox struct {
	Buffer  PanelBuffer
	Source  Rect
	Target  Point
	Flavour BlitFlavour
	// Quality is the tint InvDrawSlotBack re-shifts the slot background by
	// (inv.cpp:1137-1147).
	Quality ItemQuality
}

// PanelButtonStates is the held-down flag per main-panel button
// (MainPanelButtons, control_panel.cpp:144).
type PanelButtonStates [MainPanelButtonCount]bool

// PanelSink receives the panel boxes in emission order. The index painter and
// the recording test are the two implementations.
type PanelSink interface {
	DrawPanelBox(PanelBox)
}

// PanelState is the open/closed set of the side panels
// (control_panel.cpp:32-39).
type PanelState struct {
	character bool
	quest     bool
	stash     bool
	inventory bool
	spell     bool
}

// Contains reports whether p lies inside r, upper bound excluded
// (RectangleOf::contains, engine/rectangle.hpp:46-53).
func (r Rect) Contains(p Point) bool {
	return p.X >= r.Position.X && p.X < r.Position.X+r.Size.Width &&
		p.Y >= r.Position.Y && p.Y < r.Position.Y+r.Size.Height
}

// Shift moves the rect by delta.
func (r Rect) Shift(delta Point) Rect {
	r.Position.X += delta.X
	r.Position.Y += delta.Y

	return r
}

// CalculatePanelAreas recomputes MainPanel, LeftPanel and RightPanel for the
// screen size (control_panel.cpp:241-279).
func CalculatePanelAreas(screenWidth, screenHeight int, controls Controls) Areas {
	areas := Areas{
		Main: Rect{
			Position: Point{
				X: (screenWidth - MainPanelWidth) / 2,
				Y: screenHeight - MainPanelHeight,
			},
			Size: Size{Width: MainPanelWidth, Height: MainPanelHeight},
		},
		Left:  Rect{Size: Size{Width: SidePanelWidth, Height: SidePanelHeight}},
		Right: Rect{Size: Size{Width: SidePanelWidth, Height: SidePanelHeight}},
	}

	areas.Left.Position.X = sidePanelX(screenWidth, controls)

	areas.Left.Position.Y = (screenHeight - SidePanelHeight - MainPanelHeight) / 2

	areas.Right.Position.X = rightPanelX(screenWidth, areas.Left.Position.X, controls)

	areas.Right.Position.Y = areas.Left.Position.Y

	return areas
}

// PanelPosition maps a panel plus a panel-local offset to a screen position
// (GetPanelPosition, control_panel.cpp:332-350).
func PanelPosition(areas Areas, panel Panel, offset Point) Point {
	return panelOrigin(areas, panel).add(offset)
}

// MainPanelButtonRect returns the panel-local rect of one button
// (MainPanelButtonRect, control_panel.cpp:84-96). An id past the last button
// addresses no rect, as upstream's array bound does.
func MainPanelButtonRect(button MainPanelButton) Rect {
	rects := mainPanelButtonRects()

	if int(button) >= len(rects) {
		return Rect{}
	}

	return rects[button]
}

// DrawMainPanel emits the panel chrome box (DrawMainPanel,
// control_panel.cpp:433-437). The info box is drawn by its own bean.
func DrawMainPanel(sink PanelSink, areas Areas) {
	sink.DrawPanelBox(PanelBox{
		Buffer: BufferPanel,
		Source: Rect{
			Position: Point{X: 0, Y: PanelPaddingHeight},
			Size:     areas.Main.Size,
		},
		Target:  areas.Main.Position,
		Flavour: BlitCopy,
	})
}

// DrawMainPanelButtons emits the resting box of every single-player main-panel
// button (DrawMainPanelButtons, control_panel.cpp:439-446). A held button draws
// its pressed sprite instead, which is the panel-button art bean.
func DrawMainPanelButtons(sink PanelSink, areas Areas, pressed PanelButtonStates) {
	drawButtonBoxes(sink, areas, pressed, SinglePlayerMainPanelButtons)
}

// DrawMainPanelButtonsMultiplayer emits the resting box of all eight buttons,
// the chat and friendly-fire pair included (TotalMpMainPanelButtons,
// control_panel.cpp:123).
func DrawMainPanelButtonsMultiplayer(sink PanelSink, areas Areas, pressed PanelButtonStates) {
	drawButtonBoxes(sink, areas, pressed, MultiPlayerMainPanelButtons)
}

// Toggle flips one panel, applying the mutual-exclusion rule of the key
// handlers (diablo.cpp:1644-1748).
func (s *PanelState) Toggle(panel Panel) {
	switch panel {
	case PanelInventory:
		s.inventory = !s.inventory
		s.spell = false
		s.stash = false
	case PanelSpell:
		s.spell = !s.spell
		s.inventory = false
		s.stash = false
	case PanelCharacter:
		s.character = !s.character
		s.quest = false
	case PanelQuest:
		s.quest = !s.quest
		s.character = false
	case PanelStash:
		// Receiving the stash opens the inventory beside it
		// (stores.cpp:2032-2034).
		s.stash = !s.stash
		s.inventory = s.stash
	default:
	}
}

// Close clears one panel, plus the neighbours its upstream close path clears.
// CloseInventory closes the stash too (inv.cpp:2231-2237).
func (s *PanelState) Close(panel Panel) {
	switch panel {
	case PanelInventory:
		s.inventory = false
		s.stash = false
	case PanelSpell:
		s.spell = false
	case PanelCharacter:
		s.character = false
	case PanelQuest:
		s.quest = false
	case PanelStash:
		s.stash = false
	default:
	}
}

// IsOpen reports whether panel is open.
func (s *PanelState) IsOpen(panel Panel) bool {
	switch panel {
	case PanelInventory:
		return s.inventory
	case PanelSpell:
		return s.spell
	case PanelCharacter:
		return s.character
	case PanelQuest:
		return s.quest
	case PanelStash:
		return s.stash
	default:
		return false
	}
}

// LeftOpen reports IsLeftPanelOpen: a left-side panel is showing
// (control_panel.cpp:63-66).
func (s *PanelState) LeftOpen() bool {
	return s.character || s.quest || s.stash
}

// RightOpen reports IsRightPanelOpen (control_panel.cpp:67-70).
func (s *PanelState) RightOpen() bool {
	return s.inventory || s.spell
}

// String names the panel.
func (p Panel) String() string {
	names := [...]string{"main", "quest", "character", "spell", "inventory", "stash"}
	if int(p) >= len(names) {
		return "unknown"
	}

	return names[p]
}

// String names the button.
func (b MainPanelButton) String() string {
	names := [...]string{"charinfo", "qlog", "automap", "mainmenu", "inventory", "spellbook", "sendmsg", "friendly"}
	if int(b) >= len(names) {
		return "unknown"
	}

	return names[b]
}

// add moves a point by delta.
func (p Point) add(delta Point) Point {
	return Point{X: p.X + delta.X, Y: p.Y + delta.Y}
}

// panelOrigin is the screen origin of a panel's own buffer
// (GetPanelPosition, control_panel.cpp:332-350).
func panelOrigin(areas Areas, panel Panel) Point {
	switch panel {
	case PanelQuest, PanelCharacter, PanelStash:
		return areas.Left.Position
	case PanelSpell, PanelInventory:
		return areas.Right.Position
	default:
		return areas.Main.Position
	}
}

// sidePanelX is the horizontal placement of the left panel: at the screen
// centre for the virtual gamepad, centred in the surplus otherwise
// (control_panel.cpp:258-264).
func sidePanelX(screenWidth int, controls Controls) int {
	if controls == ControlsVirtualGamepad {
		return screenWidth/2 - SidePanelWidth
	}

	surplus := screenWidth - SidePanelWidth - SidePanelWidth - MainPanelWidth
	if surplus > 0 {
		return surplus / 2
	}

	return 0
}

// rightPanelX is the horizontal placement of the right panel: mirrored against
// the screen centre for the virtual gamepad, flush to the right edge otherwise
// (control_panel.cpp:267-271).
func rightPanelX(screenWidth, leftX int, controls Controls) int {
	if controls == ControlsVirtualGamepad {
		return screenWidth / 2
	}

	return screenWidth - SidePanelWidth - leftX
}

// drawButtonBoxes emits the resting box of the first count buttons. A held
// button skips its box and draws the pressed sprite instead, which the
// panel-button art bean supplies (control_panel.cpp:442-446).
func drawButtonBoxes(sink PanelSink, areas Areas, pressed PanelButtonStates, count int) {
	for button := range count {
		if pressed[button] {
			continue
		}

		sink.DrawPanelBox(buttonBox(areas, MainPanelButton(button)))
	}
}

// buttonBox is the panel-box draw of one resting button. The source rect sits
// one padding row deeper in the composed buffer than the target, and is one row
// taller than the button (control_panel.cpp:445).
func buttonBox(areas Areas, button MainPanelButton) PanelBox {
	rect := MainPanelButtonRect(button)
	source := rect.Shift(Point{X: 0, Y: PanelPaddingHeight})
	source.Size.Height++

	return PanelBox{
		Buffer:  BufferPanel,
		Source:  source,
		Target:  rect.Shift(areas.Main.Position).Position,
		Flavour: BlitCopy,
	}
}

// mainPanelButtonRects is MainPanelButtonRect (control_panel.cpp:84-96).
func mainPanelButtonRects() [MainPanelButtonCount]Rect {
	return [MainPanelButtonCount]Rect{
		{Position: Point{X: 9, Y: 9}, Size: Size{Width: widePanelButtonWidth, Height: widePanelButtonHeight}},
		{Position: Point{X: 9, Y: 35}, Size: Size{Width: widePanelButtonWidth, Height: widePanelButtonHeight}},
		{Position: Point{X: 9, Y: 75}, Size: Size{Width: widePanelButtonWidth, Height: widePanelButtonHeight}},
		{Position: Point{X: 9, Y: 101}, Size: Size{Width: widePanelButtonWidth, Height: widePanelButtonHeight}},
		{Position: Point{X: 560, Y: 9}, Size: Size{Width: widePanelButtonWidth, Height: widePanelButtonHeight}},
		{Position: Point{X: 560, Y: 35}, Size: Size{Width: widePanelButtonWidth, Height: widePanelButtonHeight}},
		{Position: Point{X: 87, Y: 91}, Size: Size{Width: panelButtonWidth, Height: panelButtonHeight}},
		{Position: Point{X: 527, Y: 91}, Size: Size{Width: panelButtonWidth, Height: panelButtonHeight}},
	}
}
