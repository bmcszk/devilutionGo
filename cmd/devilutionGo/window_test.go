package main

import (
	"testing"
	"time"

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
	loop.step(50 * time.Millisecond)
	loop.step(50 * time.Millisecond)

	// then the game advanced on the stepped clock
	if src.NowValue != 100*time.Millisecond {
		t.Fatalf("tick source now = %v, want 100ms", src.NowValue)
	}
}

// TestWindowBootQuitKeys pins the quit surface: ESC and q both end the
// window boot, any other key does not (devilGo-v9xy, upstream ESC in
// diablo.cpp:355-356).
func TestWindowBootQuitKeys(t *testing.T) {
	cases := []struct {
		name string
		key  keyEvent
		want bool
	}{
		{name: "escape", key: keyEscape, want: true},
		{name: "q", key: keyQ, want: true},
		{name: "space", key: keySpace, want: false},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			// when + then
			if got := isQuitKey(test.key); got != test.want {
				t.Fatalf("isQuitKey(%v) = %v, want %v", test.key, got, test.want)
			}
		})
	}
}
