// Headless run mode: the game loop advances on a mock tick source and exits
// after a fixed tick count, with no window (ADR-020, decision 1).
package main

import (
	"log/slog"
	"time"

	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/game"
)

// tickDuration is the fixed step the headless loop advances the clock by.
const tickDuration = 50 * time.Millisecond

// runHeadless drives the tick loop and returns the process exit code.
func runHeadless(ticks int) int {
	src := &engine.MockTickSource{}

	g := game.NewGame(src)
	for range ticks {
		src.NowValue += tickDuration
		g.Update(tickDuration)
	}

	slog.Info("headless run complete", "ticks", ticks)

	return 0
}
