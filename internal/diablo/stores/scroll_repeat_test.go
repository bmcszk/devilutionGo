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
