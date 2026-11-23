package main

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/game"
)

// TestWindowBootTickLoopDrivesGameTicks pins the wiring the windowed boot
// uses headlessly: the frame clock feeds game.NewGame through the window
// boot's tick shim, so real wall-clock frames advance real game ticks
// (devilGo-v9xy; the headless dep-graph guard pins that this file never
// pulls ebiten into the sim graph).
func TestWindowBootTickLoopDrivesGameTicks(t *testing.T) {
	// given a window boot driven by a scripted frame clock
	src := &engine.MockTickSource{}
	g := game.NewGame(src)
	loop := newFrameClock(src, g)

	// when the loop steps two 50ms frames
	loop.step(frameDuration)
	loop.step(frameDuration)

	// then the game advanced on the stepped clock
	if src.NowValue != 2*frameDuration {
		t.Fatalf("tick source now = %v, want %v", src.NowValue, 2*frameDuration)
	}
}

// TestWindowBootQuitKeys pins the quit surface since the front flow took
// over (devilGo-uuyh; review F2): no key ends the boot through the overlay —
// the flow's own quit request is the only key-driven exit, and the window
// close is the other rc-0 path.
func TestWindowBootQuitKeys(t *testing.T) {
	cases := []struct {
		name string
		key  keyEvent
		want bool
	}{
		{name: "escape", key: keyEscape, want: false},
		{name: "q", key: keyQ, want: false},
		{name: "space", key: keySpace, want: false},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			// when checking isQuitKey with the test key
			got := isQuitKey(test.key)

			// then the result matches the expected quit/continue
			if got != test.want {
				t.Fatalf("isQuitKey(%v) = %v, want %v", test.key, got, test.want)
			}
		})
	}
}
