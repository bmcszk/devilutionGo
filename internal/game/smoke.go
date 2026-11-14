package game

import (
	"errors"
	"log/slog"

	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/engine/demo"
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// smokeTilePixels is the fixed smoke camera: one 64px square per grid tile,
// world at origin — the mapping the Click resolver and the frame renderer
// share so injected coordinates and pixels agree.
const smokeTilePixels = 64

// SmokeHarness wires one headless running game: the loop, the kernel and the
// movement slice, with structured run logs. The smoke tests and the later
// --smoke debug driver both boot through it.
type SmokeHarness struct {
	game   *Game
	kernel *kernel.Kernel
	state  *world.State
	clock  *engine.MockTickSource
	logger *slog.Logger
	// lastClick tracks the most recent submitted click so RunTicks can log the
	// handler-level verdict (a gate pass that no handler accepted is a
	// silent no-op under law 2 — the log must still name it).
	lastClick clickRecord
}

// clickRecord is one submitted click awaiting its pass verdict.
type clickRecord struct {
	kind    string
	applied bool
}

// passLogger is the kernel publisher that writes the per-pass summary line:
// the command count submitted this tick and the events the pass raised.
type passLogger struct {
	h *SmokeHarness
}

// Publish implements kernel.Publisher: one structured line per finished pass;
// a PlayerMoved event marks the pending click applied.
func (p passLogger) Publish(events []kernel.Event) {
	for _, e := range events {
		if e.Kind == kernel.KindPlayerMoved {
			p.h.lastClick.applied = true

			break
		}
	}

	p.h.logger.Info("kernel pass", "events", len(events))
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

	g := NewGame(clock)
	g.Kernel = k
	g.OnTick = func() { logger.Info("kernel pass", "tick", g.TickCount-1) }

	h := &SmokeHarness{game: g, kernel: k, state: state, clock: clock, logger: logger}
	k.SetPublisher(passLogger{h: h})

	logger.Info("smoke boot", "kernel", "wired", "movement", "wired", "tick", 0)

	return h, nil
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

// RunTicks drives n game ticks through the loop at the tick rate.
func (h *SmokeHarness) RunTicks(n int) {
	for range n {
		h.clock.NowValue += gnTickDelay
		h.game.Update(gnTickDelay)
	}

	h.logClickVerdict()
}

// Click injects one left-button press at screen coordinates, like a driver
// would: the click routes through ClickRouter into the kernel, whose gate and
// movement handler own the verdict. The result lands in the log.
func (h *SmokeHarness) Click(screenX, screenY int) {
	router := NewClickRouter(h.state, func(x, y uint16) world.Tile {
		return world.Tile{X: int(x) / smokeTilePixels, Y: int(y) / smokeTilePixels}
	})

	rec := demo.MouseButtonRecord{Button: 1, X: uint16(screenX), Y: uint16(screenY), Down: true}

	cmd, ok := router.Command(rec)
	if !ok {
		h.logger.Info("click dropped", "reason", "not-a-left-press")

		return
	}

	cmd.Tick = h.game.TickCount

	if h.kernel.Submit(cmd) {
		h.logger.Info("click submitted", "kind", string(cmd.Kind), "verdict", "accepted", "x", screenX, "y", screenY)
		h.lastClick = clickRecord{kind: string(cmd.Kind), applied: false}

		return
	}

	h.logger.Info("click refused", "kind", string(cmd.Kind), "verdict", "refused", "x", screenX, "y", screenY)
}

// logClickVerdict names the pass verdict of the last submitted click: applied
// when a PlayerMoved event landed, refused when the pass produced none.
func (h *SmokeHarness) logClickVerdict() {
	if h.lastClick.kind == "" {
		return
	}

	verdict := "refused"
	if h.lastClick.applied {
		verdict = "applied"
	}

	h.logger.Info("click verdict", "kind", h.lastClick.kind, "verdict", verdict)
	h.lastClick = clickRecord{}
}
