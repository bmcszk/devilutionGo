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
func (f *frameClock) step(dt time.Duration) {
	f.src.NowValue += dt

	f.g.Update(dt)
}

// isQuitKey reports the keys that end the window boot: none since the front
// flow took over the quit surface (devilGo-uuyh; review F2) — ESC/q were the
// legacy overlay arms that fired on the same frame the flow walked back, so
// the flow's own quit request is the only key-driven exit. Kept as the quit
// surface's seam (window_shared keyEvent contract); always false.
func isQuitKey(_ keyEvent) bool {
	return false
}
