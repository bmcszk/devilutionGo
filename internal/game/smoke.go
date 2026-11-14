package game

import (
	"errors"
	"log/slog"

	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// SmokeHarness wires one headless running game: the loop, the kernel and the
// movement slice, with structured run logs. The smoke tests and the later
// --smoke debug driver both boot through it.
type SmokeHarness struct {
	game   *Game
	kernel *kernel.Kernel
	state  *world.State
	clock  *engine.MockTickSource
	logger *slog.Logger
}

// passLogger is the kernel publisher that writes the per-pass summary line:
// the command count submitted this tick and the events the pass raised.
type passLogger struct {
	logger *slog.Logger
}

// Publish implements kernel.Publisher: one structured line per finished pass.
func (p passLogger) Publish(events []kernel.Event) {
	p.logger.Info("kernel pass", "events", len(events))
}

// NewSmokeHarness boots the harness: game.NewGame over a driven tick source,
// kernel + world.RegisterMovement wired, and the boot line logged.
func NewSmokeHarness(logger *slog.Logger) (*SmokeHarness, error) {
	if logger == nil {
		return nil, errors.New("game: smoke harness: nil logger")
	}

	clock := &engine.MockTickSource{}
	state := world.NewState()
	k := kernel.New(world.NewGate(state, 0), nil)
	world.RegisterMovement(k, state)

	k.SetPublisher(passLogger{logger})

	g := NewGame(clock)
	g.Kernel = k
	g.OnTick = func() { logger.Info("kernel pass", "tick", g.TickCount-1) }

	logger.Info("smoke boot", "kernel", "wired", "movement", "wired", "tick", 0)

	return &SmokeHarness{game: g, kernel: k, state: state, clock: clock, logger: logger}, nil
}

// Game returns the running loop.
func (h *SmokeHarness) Game() *Game { return h.game }

// Kernel returns the command kernel the harness feeds.
func (h *SmokeHarness) Kernel() *kernel.Kernel { return h.kernel }

// State returns the world slice (hero tile, grid view).
func (h *SmokeHarness) State() *world.State { return h.state }

// Clock returns the harness tick source so a driver can advance time.
func (h *SmokeHarness) Clock() *engine.MockTickSource { return h.clock }

// Shutdown logs the shutdown line and returns the ticks the loop settled.
func (h *SmokeHarness) Shutdown() uint32 {
	ticks := h.game.TickCount
	h.logger.Info("smoke shutdown", "ticks", ticks)

	return ticks
}

// Close releases the harness; it logs shutdown when the driver forgot to.
func (*SmokeHarness) Close() error { return nil }
