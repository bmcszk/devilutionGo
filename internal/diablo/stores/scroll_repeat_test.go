package stores_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/stores"
)

// TestScrollRepeatArmsAndCountsDown pins the scrollbar push-repeat counters
// (stores.cpp:2803-2818): a spent counter fires the step and rearms to ten,
// every value in between only counts down, so a held button steps once per ten
// clicks' worth of frames.
func TestScrollRepeatArmsAndCountsDown(t *testing.T) {
	t.Parallel()

	// given idle counters
	scroll := stores.NewScrollRepeat()

	// when the up button is clicked the first time
	firstFired := scroll.PressUp()

	// then it steps and rearms
	assertValue(t, firstFired, true)
	assertValue(t, scroll.Up, stores.ScrollRepeatFrames)

	// when it is clicked ten more times
	fired := 0

	for range 10 {
		if scroll.PressUp() {
			fired++
		}
	}

	// then none of those step and the counter walked down to zero
	assertValue(t, fired, 0)
	assertValue(t, scroll.Up, 0)

	// when the next click lands
	eleventh := scroll.PressUp()

	// then it steps again
	assertValue(t, eleventh, true)
	assertValue(t, scroll.Up, stores.ScrollRepeatFrames)
}

// TestScrollRepeatDownIsIndependent pins that the two counters do not share
// state (stores.cpp:196-199).
func TestScrollRepeatDownIsIndependent(t *testing.T) {
	t.Parallel()

	// given a counter pair where only the up button was used
	scroll := stores.NewScrollRepeat()
	scroll.PressUp()

	// when the down button is clicked
	downFired := scroll.PressDown()

	// then it steps on its own counter and the up counter is untouched
	assertValue(t, downFired, true)
	assertValue(t, scroll.Down, stores.ScrollRepeatFrames)
	assertValue(t, scroll.Up, stores.ScrollRepeatFrames)
}

// TestScrollRepeatPushedState pins DrawSSlider's pushed-button test: any value
// but -1 draws the button pushed (stores.cpp:282-289).
func TestScrollRepeatPushedState(t *testing.T) {
	t.Parallel()

	// given idle counters
	scroll := stores.NewScrollRepeat()

	// when both pushed checks run on the idle pair
	idleUp := scroll.IsPushedUp()
	idleDown := scroll.IsPushedDown()

	// then neither button is pushed
	assertValue(t, idleUp, false)
	assertValue(t, idleDown, false)

	// when the up button is armed
	scroll.PressUp()

	// then only the up button draws pushed
	assertValue(t, scroll.IsPushedUp(), true)
	assertValue(t, scroll.IsPushedDown(), false)
}

// TestScrollRepeatReleaseIdlesBoth pins ReleaseStoreBtn (stores.cpp:2848-2852).
func TestScrollRepeatReleaseIdlesBoth(t *testing.T) {
	t.Parallel()

	// given both counters armed
	scroll := stores.NewScrollRepeat()
	scroll.PressUp()
	scroll.PressDown()

	// when the button is released
	scroll.Release()

	// then both counters return to -1 and draw unpushed
	assertValue(t, scroll.Up, stores.NoneScrollCountdown)
	assertValue(t, scroll.Down, stores.NoneScrollCountdown)
	assertValue(t, scroll.IsPushedUp(), false)
	assertValue(t, scroll.IsPushedDown(), false)
}

// TestMoveUpWrapsAndSkipsLabels pins StoreUp's page-less branch
// (stores.cpp:2584-2591): the cursor wraps at the top of the panel and steps
// over the non-selectable lines.
func TestMoveUpWrapsAndSkipsLabels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		from int
		want int
	}{
		{"wraps to the last selectable line", 0, 18},
		{"steps over a label row", 12, 10},
		{"plain step", 10, 8},
		{"empty cursor stays empty", stores.NoneTextLine, stores.NoneTextLine},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given a page whose only selectable lines are even
			d := labelledDialog()
			d.CurrentTextLine = tt.from

			// when the cursor steps up
			d.MoveUp()

			// then it lands where the wraparound rule says
			assertValue(t, d.CurrentTextLine, tt.want)
		})
	}
}

// TestMoveUpScrollsAtTheFirstRow pins the scrollbar branch of StoreUp
// (stores.cpp:2564-2570): at the page's first row the list pages instead of
// the cursor, and it stops at the top of the list.
func TestMoveUpScrollsAtTheFirstRow(t *testing.T) {
	t.Parallel()

	// given a paged list sitting on its first row, scrolled to the top
	d := labelledDialog()
	d.HasScrollbar = true
	d.PreviousScrollPos = 5
	d.NextScrollPos = 21
	d.CurrentTextLine = 5

	// when the cursor steps up
	d.MoveUp()

	// then the cursor stayed and the list did not page below zero
	assertValue(t, d.CurrentTextLine, 5)
	assertValue(t, d.ScrollPos, 0)

	// when the list is scrolled down and the cursor steps up again
	d.ScrollPos = 3
	d.MoveUp()

	// then the list pages up by one
	assertValue(t, d.CurrentTextLine, 5)
	assertValue(t, d.ScrollPos, 2)
}

// TestMoveDownScrollsAtTheLastRow pins the scrollbar branch of StoreDown
// (stores.cpp:2601-2607) and that it never scrolls past NumTextLines.
func TestMoveDownScrollsAtTheLastRow(t *testing.T) {
	t.Parallel()

	// given a paged list sitting on its last row with three lines of slack
	d := labelledDialog()
	d.HasScrollbar = true
	d.PreviousScrollPos = 5
	d.NextScrollPos = 21
	d.NumTextLines = 4
	d.CurrentTextLine = 21

	// when the cursor steps down three times
	for range 3 {
		d.MoveDown()
	}

	// then the list paged to three and the cursor stayed on the last row
	assertValue(t, d.CurrentTextLine, 21)
	assertValue(t, d.ScrollPos, 3)

	// when it steps down past the slack
	d.MoveDown()

	// then the list stops at NumTextLines
	assertValue(t, d.ScrollPos, 4)
}

// TestMovePriorPagesAndJumps pins StorePrior (stores.cpp:2630-2641): from the
// page's first row the list pages up by four and stops at zero, from anywhere
// else the cursor jumps back to that first row, and without a scrollbar it does
// nothing.
func TestMovePriorPagesAndJumps(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		scroll bool
		line   int
		pos    int
		wantLn int
		wantPs int
	}{
		{"pages up from the first row", true, 5, 9, 5, 5},
		{"page up stops at zero", true, 5, 2, 5, 0},
		{"jumps back to the first row", true, 12, 7, 5, 7},
		{"no scrollbar does nothing", false, 12, 7, 12, 7},
		{"empty cursor does nothing", true, stores.NoneTextLine, 7, stores.NoneTextLine, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given a page at the case's cursor and scroll position
			d := labelledDialog()
			d.HasScrollbar = tt.scroll
			d.PreviousScrollPos = 5
			d.NextScrollPos = 21
			d.NumTextLines = 16
			d.CurrentTextLine = tt.line
			d.ScrollPos = tt.pos

			// when the prior key is pressed
			d.MovePrior()

			// then the cursor and the list land where the rules say
			assertValue(t, d.CurrentTextLine, tt.wantLn)
			assertValue(t, d.ScrollPos, tt.wantPs)
		})
	}
}

// TestMoveNextPagesAndJumps pins StoreNext (stores.cpp:2643-2655): the list
// pages down by four but never past NumTextLines.
func TestMoveNextPagesAndJumps(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		line   int
		pos    int
		wantLn int
		wantPs int
	}{
		{"pages down from the last row", 21, 0, 21, 4},
		{"page down stops at NumTextLines", 21, 14, 21, 16},
		{"jumps to the last row", 8, 2, 21, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given a paged list at the case's cursor and scroll position
			d := labelledDialog()
			d.HasScrollbar = true
			d.PreviousScrollPos = 5
			d.NextScrollPos = 21
			d.NumTextLines = 16
			d.CurrentTextLine = tt.line
			d.ScrollPos = tt.pos

			// when the next key is pressed
			d.MoveNext()

			// then the cursor and the list land where the rules say
			assertValue(t, d.CurrentTextLine, tt.wantLn)
			assertValue(t, d.ScrollPos, tt.wantPs)
		})
	}
}

// labelledDialog builds a page whose only selectable lines are 0, 8, 10, 12,
// 14, 16 and 18, the shape of a list page's option rows.
func labelledDialog() *stores.Dialog {
	d := &stores.Dialog{}

	for _, line := range []int{0, 8, 10, 12, 14, 16, 18} {
		d.Lines[line].Kind = stores.LineSelectable
	}

	return d
}
