// Package ui_test proves the widget kit: Button, MenuList, ScreenStack,
// focus model, and keyboard navigation.
package ui_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game/ui"
)

// ============================================================================
// Button tests
// ============================================================================

func TestButtonCreation(t *testing.T) {
	// given: a button label
	label := "Test Button"

	// when: creating a button
	btn := ui.NewButton(label)

	// then: button has the label and is not focused initially
	if btn.Label() != label {
		t.Errorf("Button label = %q, want %q", btn.Label(), label)
	}
	if btn.Focused() {
		t.Error("Button should not be focused initially")
	}
}

func TestButtonFocus(t *testing.T) {
	// given: a button
	btn := ui.NewButton("Test")

	// when: focusing it
	btn.Focus()

	// then: button is focused
	if !btn.Focused() {
		t.Error("Button should be focused after Focus()")
	}

	// when: blurring it
	btn.Blur()

	// then: button is not focused
	if btn.Focused() {
		t.Error("Button should not be focused after Blur()")
	}
}

func TestButtonKeyActivationEnter(t *testing.T) {
	// given: a button with a callback
	activated := false
	btn := ui.NewButton("Test")
	btn.OnActivate(func() { activated = true })

	// when: handling Enter key
	btn.HandleKey(ui.KeyEnter)

	// then: callback was invoked
	if !activated {
		t.Error("Button OnActivate callback should be called on Enter")
	}
}

func TestButtonKeyActivationSpace(t *testing.T) {
	// given: a button with a callback
	activated := false
	btn := ui.NewButton("Test")
	btn.OnActivate(func() { activated = true })

	// when: handling Space key
	btn.HandleKey(ui.KeySpace)

	// then: callback was invoked
	if !activated {
		t.Error("Button OnActivate callback should be called on Space")
	}
}

func TestButtonKeyActivationIgnoresOtherKeys(t *testing.T) {
	// given: a button with a callback
	activated := false
	btn := ui.NewButton("Test")
	btn.OnActivate(func() { activated = true })

	// when: handling a non-activation key
	btn.HandleKey(ui.KeyArrowUp)

	// then: callback was not invoked
	if activated {
		t.Error("Button OnActivate callback should not be called on ArrowUp")
	}
}

// ============================================================================
// MenuList tests
// ============================================================================

func TestMenuListCreation(t *testing.T) {
	// given: a list of items
	items := []string{"Item 1", "Item 2", "Item 3"}

	// when: creating a menu list
	ml := ui.NewMenuList(items)

	// then: menu list has the items and first item is focused
	if ml.ItemCount() != len(items) {
		t.Errorf("MenuList item count = %d, want %d", ml.ItemCount(), len(items))
	}
	if ml.FocusedIndex() != 0 {
		t.Errorf("MenuList focused index = %d, want 0", ml.FocusedIndex())
	}
	if ml.FocusedItem() != items[0] {
		t.Errorf("MenuList focused item = %q, want %q", ml.FocusedItem(), items[0])
	}
}

func TestMenuListFocusDown(t *testing.T) {
	// given: a menu list with 3 items
	ml := ui.NewMenuList([]string{"Item 1", "Item 2", "Item 3"})

	// when: pressing Down arrow
	ml.HandleKey(ui.KeyArrowDown)

	// then: focus moves to second item
	if ml.FocusedIndex() != 1 {
		t.Errorf("MenuList focused index after Down = %d, want 1", ml.FocusedIndex())
	}
}

func TestMenuListFocusUp(t *testing.T) {
	// given: a menu list with 3 items, focused on second
	ml := ui.NewMenuList([]string{"Item 1", "Item 2", "Item 3"})
	ml.HandleKey(ui.KeyArrowDown) // move to index 1

	// when: pressing Up arrow
	ml.HandleKey(ui.KeyArrowUp)

	// then: focus moves back to first item
	if ml.FocusedIndex() != 0 {
		t.Errorf("MenuList focused index after Up = %d, want 0", ml.FocusedIndex())
	}
}

func TestMenuListFocusWrapsDown(t *testing.T) {
	// given: a menu list with 3 items, focused on last
	ml := ui.NewMenuList([]string{"Item 1", "Item 2", "Item 3"})
	ml.HandleKey(ui.KeyArrowDown) // index 1
	ml.HandleKey(ui.KeyArrowDown) // index 2

	// when: pressing Down arrow past last item
	ml.HandleKey(ui.KeyArrowDown)

	// then: focus wraps to first item
	if ml.FocusedIndex() != 0 {
		t.Errorf("MenuList focused index after wrap Down = %d, want 0", ml.FocusedIndex())
	}
}

func TestMenuListFocusWrapsUp(t *testing.T) {
	// given: a menu list with 3 items, focused on first
	ml := ui.NewMenuList([]string{"Item 1", "Item 2", "Item 3"})

	// when: pressing Up arrow past first item
	ml.HandleKey(ui.KeyArrowUp)

	// then: focus wraps to last item
	if ml.FocusedIndex() != 2 {
		t.Errorf("MenuList focused index after wrap Up = %d, want 2", ml.FocusedIndex())
	}
}

func TestMenuListSelectionEnter(t *testing.T) {
	// given: a menu list with a selection callback
	selectedIndex := -1
	ml := ui.NewMenuList([]string{"Item 1", "Item 2", "Item 3"})
	ml.OnSelect(func(index int) { selectedIndex = index })

	// when: pressing Enter
	ml.HandleKey(ui.KeyEnter)

	// then: callback invoked with focused index
	if selectedIndex != 0 {
		t.Errorf("MenuList OnSelect callback index = %d, want 0", selectedIndex)
	}
}

func TestMenuListSelectionEscape(t *testing.T) {
	// given: a menu list with a cancel callback
	cancelled := false
	ml := ui.NewMenuList([]string{"Item 1", "Item 2", "Item 3"})
	ml.OnCancel(func() { cancelled = true })

	// when: pressing Escape
	ml.HandleKey(ui.KeyEscape)

	// then: callback was invoked
	if !cancelled {
		t.Error("MenuList OnCancel callback should be called on Escape")
	}
}

func TestMenuListEmptyList(t *testing.T) {
	// given: an empty menu list
	ml := ui.NewMenuList([]string{})

	// when: handling keys on the empty list
	ml.HandleKey(ui.KeyArrowDown)
	ml.HandleKey(ui.KeyArrowUp)
	ml.HandleKey(ui.KeyEnter)
	ml.HandleKey(ui.KeyEscape)

	// then: nothing panicked and the list stayed empty and unfocused
	if ml.ItemCount() != 0 {
		t.Errorf("Empty MenuList item count = %d, want 0", ml.ItemCount())
	}
	if ml.FocusedIndex() != -1 {
		t.Errorf("Empty MenuList focused index = %d, want -1", ml.FocusedIndex())
	}
	if ml.FocusedItem() != "" {
		t.Errorf("Empty MenuList focused item = %q, want empty string", ml.FocusedItem())
	}
}

// ============================================================================
// ScreenStack tests
// ============================================================================

func TestScreenStackCreation(t *testing.T) {
	// when: creating a screen stack
	stack := ui.NewScreenStack()

	// then: stack is empty
	if stack.Size() != 0 {
		t.Errorf("New ScreenStack size = %d, want 0", stack.Size())
	}
	if stack.Top() != nil {
		t.Error("New ScreenStack top should be nil")
	}
}

func TestScreenStackPush(t *testing.T) {
	// given: a screen stack and a screen
	stack := ui.NewScreenStack()
	screen := ui.NewWidgetScreen(ui.NewFocusModel())

	// when: pushing the screen
	stack.Push(screen)

	// then: stack size is 1 and top is the screen
	if stack.Size() != 1 {
		t.Errorf("ScreenStack size after push = %d, want 1", stack.Size())
	}
	if stack.Top() != screen {
		t.Error("ScreenStack top should be the pushed screen")
	}
}

func TestScreenStackPushMultiple(t *testing.T) {
	// given: a screen stack
	stack := ui.NewScreenStack()
	screen1 := ui.NewWidgetScreen(ui.NewFocusModel())
	screen2 := ui.NewWidgetScreen(ui.NewFocusModel())
	screen3 := ui.NewWidgetScreen(ui.NewFocusModel())

	// when: pushing multiple screens
	stack.Push(screen1)
	stack.Push(screen2)
	stack.Push(screen3)

	// then: stack size is 3 and top is the last pushed
	if stack.Size() != 3 {
		t.Errorf("ScreenStack size after 3 pushes = %d, want 3", stack.Size())
	}
	if stack.Top() != screen3 {
		t.Error("ScreenStack top should be the last pushed screen")
	}
}

func TestScreenStackPop(t *testing.T) {
	// given: a screen stack with two screens
	stack := ui.NewScreenStack()
	screen1 := ui.NewWidgetScreen(ui.NewFocusModel())
	screen2 := ui.NewWidgetScreen(ui.NewFocusModel())
	stack.Push(screen1)
	stack.Push(screen2)

	// when: popping the top screen
	popped := stack.Pop()

	// then: popped screen is screen2, stack size is 1, top is screen1
	if popped != screen2 {
		t.Error("Pop should return the popped screen")
	}
	if stack.Size() != 1 {
		t.Errorf("ScreenStack size after pop = %d, want 1", stack.Size())
	}
	if stack.Top() != screen1 {
		t.Error("ScreenStack top should be the previous screen")
	}
}

func TestScreenStackPopEmpty(t *testing.T) {
	// given: an empty screen stack
	stack := ui.NewScreenStack()

	// when: popping from empty stack
	popped := stack.Pop()

	// then: returns nil, stack remains empty
	if popped != nil {
		t.Error("Pop on empty stack should return nil")
	}
	if stack.Size() != 0 {
		t.Errorf("ScreenStack size after pop on empty = %d, want 0", stack.Size())
	}
}

func TestScreenStackTopScreenReceivesInput(t *testing.T) {
	// given: two screens, each holding a focused button that records activation
	topActivated, bottomActivated := false, false
	topButton := ui.NewButton("top")
	topButton.OnActivate(func() { topActivated = true })
	top := screenWithFocused(topButton)

	bottomButton := ui.NewButton("bottom")
	bottomButton.OnActivate(func() { bottomActivated = true })
	bottom := screenWithFocused(bottomButton)

	stack := ui.NewScreenStack()
	stack.Push(bottom)
	stack.Push(top)

	// when: handling Enter
	stack.HandleKey(ui.KeyEnter)

	// then: only the top screen's button activated
	if !topActivated {
		t.Error("Top screen's button should activate on Enter")
	}
	if bottomActivated {
		t.Error("Lower screen's button should not activate")
	}
}

func TestScreenStackEscapePopsScreen(t *testing.T) {
	// given: a screen stack with two screens
	stack := ui.NewScreenStack()
	screen1 := ui.NewWidgetScreen(ui.NewFocusModel())
	screen2 := ui.NewWidgetScreen(ui.NewFocusModel())
	stack.Push(screen1)
	stack.Push(screen2)

	// when: handling Escape key
	stack.HandleKey(ui.KeyEscape)

	// then: top screen is popped, screen1 is now top
	if stack.Size() != 1 {
		t.Errorf("ScreenStack size after Escape = %d, want 1", stack.Size())
	}
	if stack.Top() != screen1 {
		t.Error("ScreenStack top should be the previous screen after Escape")
	}
}

func TestScreenStackEscapeOnEmptyStack(t *testing.T) {
	// given: an empty screen stack
	stack := ui.NewScreenStack()

	// when: handling Escape key
	stack.HandleKey(ui.KeyEscape)

	// then: stack remains empty (no panic)
	if stack.Size() != 0 {
		t.Errorf("ScreenStack size after Escape on empty = %d, want 0", stack.Size())
	}
}

// ============================================================================
// Focus Model tests
// ============================================================================

func TestFocusModelCreation(t *testing.T) {
	// when: creating a focus model
	fm := ui.NewFocusModel()

	// then: no focused widget initially
	if fm.Focused() != nil {
		t.Error("New FocusModel should have no focused widget")
	}
}

func TestFocusModelKeysInertWithoutFocus(t *testing.T) {
	// given: a focus model over a callback button, with no focus set
	fm, activated := unfocusedButtonModel()

	// when: Enter arrives with no widget focused
	fm.HandleKey(ui.KeyEnter)

	// then: nothing activates and still no widget is focused
	if *activated {
		t.Error("Button OnActivate should not fire without focus")
	}
	if fm.Focused() != nil {
		t.Error("keys must not take focus by themselves")
	}
}

func TestWidgetScreenKeysInertWithoutFocus(t *testing.T) {
	// given: a widget screen over a callback button, with no focus set
	fm, activated := unfocusedButtonModel()

	// when: Enter arrives through the screen with no widget focused
	ui.NewWidgetScreen(fm).HandleKey(ui.KeyEnter)

	// then: nothing activates and still no widget is focused
	if *activated {
		t.Error("Button OnActivate should not fire without focus")
	}
	if fm.Focused() != nil {
		t.Error("keys must not take focus by themselves")
	}
}

func TestFocusModelSetFocus(t *testing.T) {
	// given: a focus model and two widgets in a container
	fm := ui.NewFocusModel()
	btn1 := ui.NewButton("Button 1")
	btn2 := ui.NewButton("Button 2")
	fm.SetContainer(ui.NewFocusContainer([]ui.Focusable{btn1, btn2}))

	// when: setting focus to btn1
	fm.SetFocus(btn1)

	// then: btn1 is focused, btn2 is not
	if !btn1.Focused() {
		t.Error("btn1 should be focused")
	}
	if btn2.Focused() {
		t.Error("btn2 should not be focused")
	}
	if fm.Focused() != btn1 {
		t.Error("FocusModel.Focused() should return btn1")
	}

	// when: setting focus to btn2
	fm.SetFocus(btn2)

	// then: btn2 is focused, btn1 is not
	if !btn2.Focused() {
		t.Error("btn2 should be focused")
	}
	if btn1.Focused() {
		t.Error("btn1 should not be focused")
	}
	if fm.Focused() != btn2 {
		t.Error("FocusModel.Focused() should return btn2")
	}
}

func TestFocusModelClearFocus(t *testing.T) {
	// given: a focus model with a focused widget in a container
	fm := ui.NewFocusModel()
	btn := ui.NewButton("Button")
	fm.SetContainer(ui.NewFocusContainer([]ui.Focusable{btn}))
	fm.SetFocus(btn)

	// when: clearing focus
	fm.ClearFocus()

	// then: no widget is focused
	if fm.Focused() != nil {
		t.Error("FocusModel should have no focused widget after ClearFocus")
	}
	if btn.Focused() {
		t.Error("Widget should not be focused after ClearFocus")
	}
}

func TestFocusModelTabCyclesBothWays(t *testing.T) {
	// given: a focus model over three buttons
	fm := ui.NewFocusModel()
	btn1 := ui.NewButton("Button 1")
	btn2 := ui.NewButton("Button 2")
	btn3 := ui.NewButton("Button 3")
	fm.SetContainer(ui.NewFocusContainer([]ui.Focusable{btn1, btn2, btn3}))

	// when: Tab walks forward past the last widget
	wantForward := []ui.Focusable{btn1, btn2, btn3, btn1}

	for step, want := range wantForward {
		fm.HandleKey(ui.KeyTab)

		// then: focus reaches the expected widget, wrapping at the end
		if got := fm.Focused(); got != want {
			t.Fatalf("after Tab %d focused = %v, want %v", step+1, got, want)
		}
	}

	// when: Shift+Tab walks backward past the first widget
	wantBackward := []ui.Focusable{btn3, btn2, btn1, btn3}

	for step, want := range wantBackward {
		fm.HandleKey(ui.KeyShiftTab)

		// then: focus reaches the expected widget, wrapping at the start
		if got := fm.Focused(); got != want {
			t.Fatalf("after Shift+Tab %d focused = %v, want %v", step+1, got, want)
		}
	}
}

func TestFocusModelArrowNavigation(t *testing.T) {
	// given: a focus model with a menu list in a container
	fm := ui.NewFocusModel()
	ml := ui.NewMenuList([]string{"Item 1", "Item 2", "Item 3"})
	fm.SetContainer(ui.NewFocusContainer([]ui.Focusable{ml}))
	fm.SetFocus(ml)

	// when: pressing Down arrow
	fm.HandleKey(ui.KeyArrowDown)

	// then: menu list's internal focus moved
	if ml.FocusedIndex() != 1 {
		t.Errorf("MenuList focused index after Down = %d, want 1", ml.FocusedIndex())
	}

	// when: pressing Up arrow
	fm.HandleKey(ui.KeyArrowUp)

	// then: menu list's internal focus moved back
	if ml.FocusedIndex() != 0 {
		t.Errorf("MenuList focused index after Up = %d, want 0", ml.FocusedIndex())
	}
}

// screenWithFocused builds a widget screen whose single widget holds focus.
func screenWithFocused(widget ui.Focusable) *ui.WidgetScreen {
	fm := ui.NewFocusModel()
	fm.SetContainer(ui.NewFocusContainer([]ui.Focusable{widget}))
	fm.SetFocus(widget)

	return ui.NewWidgetScreen(fm)
}

// unfocusedButtonModel builds a focus model over a callback button, leaves the
// model unfocused, and reports whether the callback fired.
func unfocusedButtonModel() (*ui.FocusModel, *bool) {
	activated := false
	btn := ui.NewButton("Button")
	btn.OnActivate(func() { activated = true })
	fm := ui.NewFocusModel()
	fm.SetContainer(ui.NewFocusContainer([]ui.Focusable{btn}))

	return fm, &activated
}
