package main

import (
	"time"

	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/game"
)

// frameClock advances the game's tick source by the frame duration and runs
// one game Update: the windowed boot's tick shim. ebiten's TPS loop calls
// step once per frame; the headless tests drive it with scripted durations.
type frameClock struct {
	src *engine.MockTickSource
	g   *game.Game
}

// newFrameClock pairs a tick source with the game it drives.
func newFrameClock(src *engine.MockTickSource, g *game.Game) *frameClock {
	return &frameClock{src: src, g: g}
}

// step advances the clock by dt and runs the due game ticks.
func (f *frameClock) step(dt time.Duration) int {
	f.src.NowValue += dt

	return f.g.Update(dt)
}

// isQuitKey reports the keys that end the window boot: ESC and q
// (diablo.cpp:355-356 Escape opens the menu; here the slice's quit is a
// clean rc-0 exit).
func isQuitKey(key keyEvent) bool {
	return key == keyEscape || key == keyQ
}
