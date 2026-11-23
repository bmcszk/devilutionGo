package main

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/controls"
	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/game"
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// PlayerMoved is the event kind this lane's walk announces
// (kernel.KindPlayerMoved, the movement verb's only event).
const playerMovedKind = kernel.KindPlayerMoved

// TestDeviceLane_BootWiresKernelAndBridgeSession pins the boot wiring the
// windowed run depends on: the kernel is real (not nil), the world slice behind
// it raised a hero, and the window's submit site exists. A nil kernel here is
// exactly the s6e gap — the window would tick a bus no handler listens to.
func TestDeviceLane_BootWiresKernelAndBridgeSession(t *testing.T) {
	// given the windowed boot's wiring on a scripted frame clock
	src := &engine.MockTickSource{}
	g := game.NewGame(src)

	// when the boot wires the lane
	lane := bootDeviceLane(src, g)

	// then the kernel is wired
	if g.Kernel == nil {
		t.Fatal("boot left g.Kernel nil, want the movement verbs' kernel")
	}

	// and the bridge session exists
	if lane.input == nil {
		t.Fatal("boot left the bridge session nil, want the window's submit site")
	}

	// and a hero stands on the spawn tile
	pack, tile, ok := lane.state.RaisedPack()
	if !ok {
		t.Fatal("boot raised no hero, want the spawn raise")
	}

	if tile != (world.Tile{X: heroSpawnX, Y: heroSpawnY}) {
		t.Fatalf("hero tile = %v, want the spawn tile (%d,%d)", tile, heroSpawnX, heroSpawnY)
	}

	if pack != (world.StartingPack{Class: 0}) {
		t.Fatalf("raised pack = %+v, want the warrior pack", pack)
	}
}

// TestDeviceLane_StickPushThroughTheGamePassMovesTheHero is the lane's real
// walk: a stick push queued by the window layer lands on the kernel through the
// bridge, and the Game's own settled-tick pass dispatches it — the hero tile
// changes and PlayerMoved lands in the kernel's trace, stamped with the game's
// tick, not a frame counter.
func TestDeviceLane_StickPushThroughTheGamePassMovesTheHero(t *testing.T) {
	// given the booted lane
	src := &engine.MockTickSource{}
	g := game.NewGame(src)
	lane := bootDeviceLane(src, g)
	requireKernel(t, g)
	start := lane.state.Hero()

	// when a firm east push is queued
	if !lane.PushStick(controls.PushFirm, eastPush()) {
		t.Fatal("gate refused the firm east push, want it accepted")
	}

	// then queuing dispatched nothing yet: the window layer only submits, and the
	// frame's own settled pass is the single pass driver (review s6e F2)
	if events := g.Kernel.Trace(); len(events) != 0 {
		t.Fatalf("queuing the push emitted %d events before any settled tick, want 0", len(events))
	}

	if got := lane.state.Hero(); got != start {
		t.Fatalf("hero tile = %v before the settled tick, want the untouched %v", got, start)
	}

	// when the frame settles
	lane.frame(func() {})

	// then the hero stepped one tile east
	want := world.Tile{X: start.X + 1, Y: start.Y - 1}
	if got := lane.state.Hero(); got != want {
		t.Fatalf("hero tile = %v, want %v after the east push", got, want)
	}

	// and the kernel recorded the move at the game's tick
	if got := g.TickCount; got == 0 {
		t.Fatal("TickCount = 0, want the settled pass to advance it")
	}

	tick := g.TickCount - 1
	assertPlayerMovedOnce(t, g.Kernel, tick, want)

	// and the command carried the tick its pass ran: the window layer stamps no
	// frame counter of its own (review s6e F2)
	commands := g.Kernel.TraceCommandsAt(tick)
	if len(commands) != 1 {
		t.Fatalf("tick %d dispatched %d commands, want the one queued push", tick, len(commands))
	}

	if commands[0].Tick != tick {
		t.Fatalf("command tick = %d, want the pass tick %d", commands[0].Tick, tick)
	}
}

// TestDeviceLane_EachFrameSubmitsUnderItsOwnTick pins the single pass driver:
// two frames queue two pushes and each lands in its own settled tick. A second
// pass driver in the window layer suppresses one of them (the s6e F2 defect).
func TestDeviceLane_EachFrameSubmitsUnderItsOwnTick(t *testing.T) {
	// given the booted lane
	src := &engine.MockTickSource{}
	g := game.NewGame(src)
	lane := bootDeviceLane(src, g)
	requireKernel(t, g)
	start := lane.state.Hero()

	// when two frames each queue one firm east push
	for range 2 {
		if !lane.PushStick(controls.PushFirm, eastPush()) {
			t.Fatal("gate refused the firm east push, want it accepted")
		}

		lane.frame(func() {})
	}

	// then each settled tick dispatched its own push
	for tick := range uint32(2) {
		assertPlayerMovedOnce(t, g.Kernel, tick, world.Tile{X: start.X + int(tick) + 1, Y: start.Y - int(tick) - 1})
	}

	// and the hero walked both steps
	if got, want := lane.state.Hero(), (world.Tile{X: start.X + 2, Y: start.Y - 2}); got != want {
		t.Fatalf("hero tile = %v, want %v after two east pushes", got, want)
	}
}

// TestDeviceLane_QuietFrameMovesNobody pins the complement: a frame with no
// decision settles business as usual and walks nobody, so the lane cannot pass
// its assertion by moving unconditionally.
func TestDeviceLane_QuietFrameMovesNobody(t *testing.T) {
	// given the booted lane
	src := &engine.MockTickSource{}
	g := game.NewGame(src)
	lane := bootDeviceLane(src, g)
	requireKernel(t, g)
	start := lane.state.Hero()

	// when two frames settle with no stick input
	lane.frame(func() {})
	lane.frame(func() {})

	// then nobody moved and no PlayerMoved was announced
	if got := lane.state.Hero(); got != start {
		t.Fatalf("hero tile = %v, want the untouched %v", got, start)
	}

	for _, ev := range g.Kernel.Trace() {
		if ev.Kind == playerMovedKind {
			t.Fatalf("quiet frame announced %s at %v, want silence", ev.Kind, (world.Tile{X: int(ev.A), Y: int(ev.B)}))
		}
	}
}

// TestDeviceLane_FrameReadsInputBeforeItSettles pins the frame order: the read
// callback queues the frame's decision before the clock steps, so the decision
// rides the settled pass of that same frame. Reverse the order and the push
// lands a frame late — a visible input lag and, with the gate's cooldown armed,
// a dropped walk (review s6e F2's ordering half).
func TestDeviceLane_FrameReadsInputBeforeItSettles(t *testing.T) {
	// given the booted lane
	src := &engine.MockTickSource{}
	g := game.NewGame(src)
	lane := bootDeviceLane(src, g)
	requireKernel(t, g)
	start := lane.state.Hero()

	// when ONE frame reads a firm east push as its input
	lane.frame(func() {
		if !lane.PushStick(controls.PushFirm, eastPush()) {
			t.Fatal("gate refused the firm east push, want it accepted")
		}
	})

	// then that same frame already walked: the read preceded the settled pass
	if got, want := lane.state.Hero(), (world.Tile{X: start.X + 1, Y: start.Y - 1}); got != want {
		t.Fatalf("hero tile = %v after one frame, want %v (the read must precede the settled tick)", got, want)
	}
}

// requireKernel stops a test at a named failure when the boot left the kernel
// nil, so a wiring regression reports itself instead of panicking inside the
// kernel's accessors (review s6e F1's "assert the boot" clause).
func requireKernel(t *testing.T, g *game.Game) {
	t.Helper()

	if g.Kernel == nil {
		t.Fatal("boot left g.Kernel nil, want the movement verbs' kernel")
	}
}

// eastPush is the AxisDirection a firm east stick push carries
// (axis_direction.h:7-16 ordinals).
func eastPush() controls.AxisDirection {
	return controls.AxisDirection{X: controls.AxisXRight, Y: controls.AxisYNone}
}

// assertPlayerMovedOnce asserts tick's recorded events are exactly one
// PlayerMoved landing on want, the payload the movement verb emits
// (internal/game/world/movement.go step()).
func assertPlayerMovedOnce(t *testing.T, k *kernel.Kernel, tick uint32, want world.Tile) {
	t.Helper()

	events := k.TraceEventsAt(tick)
	if len(events) != 1 {
		t.Fatalf("tick %d recorded %d events, want exactly one %s", tick, len(events), playerMovedKind)
	}

	if got := events[0]; got.Kind != playerMovedKind || got.A != int32(want.X) || got.B != int32(want.Y) {
		t.Fatalf("tick %d event = %+v, want %s at %v", tick, got, playerMovedKind, want)
	}
}
