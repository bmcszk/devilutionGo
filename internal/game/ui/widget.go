// Package ui is the widget kit: Button, MenuList, ScreenStack, focus model,
// and keyboard navigation. Presentation-only, sim-pure (ADR-017).
package ui

const (
	KeyUnknown Key = iota
	KeyEnter
	KeySpace
	KeyEscape
	KeyArrowUp
	KeyArrowDown
	KeyArrowLeft
	KeyArrowRight
	KeyTab
	KeyShiftTab
)

// Key represents a keyboard key for widget navigation.
type Key int

// Focusable is a widget that can receive keyboard focus.
type Focusable interface {
	Focused() bool
	Focus()
	Blur()
	HandleKey(Key)
}

// Button is a clickable button widget with label and activation callback.
type Button struct {
	label      string
	focused    bool
	onActivate func()
}

// MenuList is a vertical list of selectable items with keyboard navigation.
type MenuList struct {
	items      []string
	focusedIdx int
	onSelect   func(int)
	onCancel   func()
}

// Screen represents a UI screen that can handle input.
type Screen interface {
	HandleKey(Key)
}

// WidgetScreen is a screen over a focus model: it routes keys to the focused
// widget and leaves every non-widget key unhandled.
type WidgetScreen struct {
	focus *FocusModel
}

// FocusContainer holds an ordered list of focusable widgets.
type FocusContainer struct {
	widgets []Focusable
}

// FocusModel manages keyboard focus across a set of focusable widgets.
type FocusModel struct {
	container  *FocusContainer
	focusedIdx int // index within container, -1 if none
}

// ScreenStack manages a stack of screens, where only the top screen receives input.
type ScreenStack struct {
	screens []Screen
}

// ============================================================================
// Button functions
// ============================================================================

// NewButton creates a new button with the given label.
func NewButton(label string) *Button {
	return &Button{label: label}
}

// Label returns the button's label text.
func (b *Button) Label() string {
	return b.label
}

// Focused reports whether the button has keyboard focus.
func (b *Button) Focused() bool {
	return b.focused
}

// Focus gives the button keyboard focus.
func (b *Button) Focus() {
	b.focused = true
}

// Blur removes keyboard focus from the button.
func (b *Button) Blur() {
	b.focused = false
}

// OnActivate sets the callback invoked when the button is activated.
func (b *Button) OnActivate(fn func()) {
	b.onActivate = fn
}

// HandleKey processes a key press. Enter and Space activate the button.
func (b *Button) HandleKey(key Key) {
	switch key {
	case KeyEnter, KeySpace:
		if b.onActivate != nil {
			b.onActivate()
		}
	default:
	}
}

// ============================================================================
// MenuList functions
// ============================================================================

// NewMenuList creates a new menu list with the given items.
func NewMenuList(items []string) *MenuList {
	ml := &MenuList{
		items:      items,
		focusedIdx: -1,
	}
	if len(items) > 0 {
		ml.focusedIdx = 0
	}
	return ml
}

// ItemCount returns the number of items in the list.
func (ml *MenuList) ItemCount() int {
	return len(ml.items)
}

// FocusedIndex returns the index of the currently focused item, or -1 if empty.
func (ml *MenuList) FocusedIndex() int {
	return ml.focusedIdx
}

// FocusedItem returns the currently focused item text, or empty string if none.
func (ml *MenuList) FocusedItem() string {
	if ml.focusedIdx >= 0 && ml.focusedIdx < len(ml.items) {
		return ml.items[ml.focusedIdx]
	}
	return ""
}

// Focused reports whether the menu list has focus (always true when non-empty).
func (ml *MenuList) Focused() bool {
	return ml.focusedIdx >= 0
}

// Focus gives the menu list keyboard focus at its first item.
func (ml *MenuList) Focus() {
	if len(ml.items) > 0 && ml.focusedIdx == -1 {
		ml.focusedIdx = 0
	}
}

// Blur removes keyboard focus from the menu list.
func (ml *MenuList) Blur() {
	ml.focusedIdx = -1
}

// OnSelect sets the callback invoked when an item is selected (Enter).
func (ml *MenuList) OnSelect(fn func(int)) {
	ml.onSelect = fn
}

// OnCancel sets the callback invoked when the menu is cancelled (Escape).
func (ml *MenuList) OnCancel(fn func()) {
	ml.onCancel = fn
}

// HandleKey processes a key press for navigation and selection.
func (ml *MenuList) HandleKey(key Key) {
	if len(ml.items) == 0 {
		return
	}

	switch key {
	case KeyArrowDown:
		ml.focusedIdx = (ml.focusedIdx + 1) % len(ml.items)
	case KeyArrowUp:
		ml.focusedIdx = (ml.focusedIdx - 1 + len(ml.items)) % len(ml.items)
	case KeyEnter:
		if ml.onSelect != nil {
			ml.onSelect(ml.focusedIdx)
		}
	case KeyEscape:
		if ml.onCancel != nil {
			ml.onCancel()
		}
	default:
	}
}

// ============================================================================
// WidgetScreen functions
// ============================================================================

// NewWidgetScreen creates a screen over the given focus model.
func NewWidgetScreen(focus *FocusModel) *WidgetScreen {
	return &WidgetScreen{focus: focus}
}

// HandleKey routes the key to the screen's focused widget.
func (s *WidgetScreen) HandleKey(key Key) {
	s.focus.HandleKey(key)
}

// ============================================================================
// FocusModel functions
// ============================================================================

// NewFocusContainer creates a new focus container with the given widgets.
func NewFocusContainer(widgets []Focusable) *FocusContainer {
	return &FocusContainer{widgets: widgets}
}

// Widgets returns the widgets in focus order.
func (fc *FocusContainer) Widgets() []Focusable {
	return fc.widgets
}

// NewFocusModel creates a new focus model.
func NewFocusModel() *FocusModel {
	return &FocusModel{focusedIdx: -1}
}

// SetContainer sets the focus container.
func (fm *FocusModel) SetContainer(container *FocusContainer) {
	fm.container = container
	fm.focusedIdx = -1
}

// Focused returns the currently focused widget, or nil if none.
func (fm *FocusModel) Focused() Focusable {
	if fm.container == nil || fm.focusedIdx < 0 || fm.focusedIdx >= len(fm.container.widgets) {
		return nil
	}
	return fm.container.widgets[fm.focusedIdx]
}

// HandleKey moves focus on Tab and Shift+Tab; every other key goes to the
// focused widget.
func (fm *FocusModel) HandleKey(key Key) {
	if fm.container == nil || len(fm.container.widgets) == 0 {
		return
	}

	switch key {
	case KeyTab:
		fm.focusNext()
	case KeyShiftTab:
		fm.focusPrev()
	default:
		if focused := fm.Focused(); focused != nil {
			focused.HandleKey(key)
		}
	}
}

// SetFocus focuses the given widget if it's in the container.
func (fm *FocusModel) SetFocus(widget Focusable) {
	if fm.container == nil {
		return
	}
	for i, w := range fm.container.widgets {
		if w == widget {
			fm.clearPreviousFocus()
			fm.focusedIdx = i
			widget.Focus()
			return
		}
	}
}

// ClearFocus clears the current focus.
func (fm *FocusModel) ClearFocus() {
	fm.clearPreviousFocus()
	fm.focusedIdx = -1
}

// focusNext moves focus to the next widget (wraps).
func (fm *FocusModel) focusNext() {
	if fm.container == nil || len(fm.container.widgets) == 0 {
		return
	}

	fm.clearPreviousFocus()

	if fm.focusedIdx < 0 {
		fm.focusedIdx = 0
	} else {
		fm.focusedIdx = (fm.focusedIdx + 1) % len(fm.container.widgets)
	}
	fm.container.widgets[fm.focusedIdx].Focus()
}

// focusPrev moves focus to the previous widget (wraps).
func (fm *FocusModel) focusPrev() {
	if fm.container == nil || len(fm.container.widgets) == 0 {
		return
	}

	fm.clearPreviousFocus()

	if fm.focusedIdx < 0 {
		fm.focusedIdx = len(fm.container.widgets) - 1
	} else {
		fm.focusedIdx = (fm.focusedIdx - 1 + len(fm.container.widgets)) % len(fm.container.widgets)
	}
	fm.container.widgets[fm.focusedIdx].Focus()
}

// clearPreviousFocus blurs the previously focused widget.
func (fm *FocusModel) clearPreviousFocus() {
	if fm.focusedIdx >= 0 && fm.focusedIdx < len(fm.container.widgets) {
		fm.container.widgets[fm.focusedIdx].Blur()
	}
}

// ============================================================================
// ScreenStack functions
// ============================================================================

// NewScreenStack creates a new empty screen stack.
func NewScreenStack() *ScreenStack {
	return &ScreenStack{}
}

// Size returns the number of screens in the stack.
func (ss *ScreenStack) Size() int {
	return len(ss.screens)
}

// Top returns the top screen, or nil if the stack is empty.
func (ss *ScreenStack) Top() Screen {
	if len(ss.screens) == 0 {
		return nil
	}
	return ss.screens[len(ss.screens)-1]
}

// Push adds a screen to the top of the stack.
func (ss *ScreenStack) Push(screen Screen) {
	ss.screens = append(ss.screens, screen)
}

// Pop removes and returns the top screen, or nil if empty.
func (ss *ScreenStack) Pop() Screen {
	if len(ss.screens) == 0 {
		return nil
	}
	screen := ss.screens[len(ss.screens)-1]
	ss.screens = ss.screens[:len(ss.screens)-1]
	return screen
}

// HandleKey forwards the key to the top screen. Escape pops the top screen.
func (ss *ScreenStack) HandleKey(key Key) {
	if key == KeyEscape {
		ss.Pop()
		return
	}
	if top := ss.Top(); top != nil {
		top.HandleKey(key)
	}
}
