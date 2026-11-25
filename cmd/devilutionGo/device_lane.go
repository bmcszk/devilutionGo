package main

import (
	"github.com/bmcszk/devilutionGo/internal/diablo/controls"
	"github.com/bmcszk/devilutionGo/internal/diablo/inputbridge"
	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/game"
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// Device lane boot constants: the kernel seed and stream name are fixed so two
// boots draw the same walk, and the spawn tile is the bootstrap shortcut's
// fixed raise (see bootDeviceLane).
const (
	kernelSeed     uint32 = 0xdeadbeef
	movementStream        = "movement"
	// heroSpawnX/heroSpawnY raise the boot hero on the first-scene hero tile
	// (75,68 — ViewPosition of ENTRY_MAIN, town.cpp:363-365): the live view
	// composes at the state's hero tile (devilGo-njzu), so the spawn IS the
	// baseline tile and the validated first frame stays byte-identical.
	heroSpawnX = 75
	heroSpawnY = 68
)

// deviceLane is the device lane's boot: the kernel-wired game, the world slice
// its movement verbs write, the one bridge session the window layer submits
// through, and the frame clock that drives the game's settled ticks.
//
// It lives in the untagged build on purpose: the device-lane test drives this
// exact wiring (device_lane_test.go), which it cannot do through the ebiten
// window type. The kernel wiring itself is ebiten-free — v9xy's dep-graph guard
// pins that this package still links no display stack.
//
// The lane never passes the kernel. The Game is the single pass driver
// (game.Game.runKernelPass, internal/game/game.go:217-223): a frame queues its
// decisions and then runs the settled ticks, whose pass drains the queue. That
// is ADR-018 laws 1 and 3, and it is the fix for review s6e F2.
// lanePackSource is the PackSource the create verb serves: the windowed boot
// installs the data-built classdat source (classdat_pack_source.go); the
// variable indirection keeps this untagged file out of the scene dependency.
var lanePackSource world.PackSource = bootPackSource{}

type deviceLane struct {
	game  *game.Game
	state *world.State
	input *inputbridge.Session
	clock *frameClock

	// onCreateEvent is the callback the selhero create flow registers: the
	// lane reads the kernel trace tail after each frame's settled pass and
	// invokes it with the first new event, so the store row lands only when
	// the kernel actually raised the hero (review F5). Nil = no hook.
	onCreateEvent func(kernel.Event)

	// tracedNext is the next tick whose events the hook has not yet read:
	// the kernel's pass records events under the tick it ran as, which is
	// TickCount BEFORE the pass increments it.
	tracedNext uint32

	// pendingCreate is set by the selhero create flow when it submits its
	// CreatePack decision: the lane then reports the pass's outcome — the
	// CreatedPack event, or a refusal when the command dispatched and the
	// handler stayed silent — exactly once (review F5).
	pendingCreate bool
}

// bootDeviceLane wires the kernel exactly as the windowed run boots it: a world
// slice, the session gate over it, a seeded kernel, the movement verbs, and the
// bridge session the window submits through.
//
// Disclosures (review s6e F3/F4/F5 — recorded here, updated devilGo-5fr4):
//   - The hero raise was DIRECT state mutation; the selhero slice registers the
//     event-sourced CreatePack verb (world.RegisterCreatePack over
//     bootPackSource) and the selhero screen submits KindCreatePack. The boot
//     itself keeps the bootstrap raise so the live-view baseline tile stays
//     byte-identical before the front flow ever reaches selhero.
//   - The mouse lane stays unwired: inputbridge.Session.MouseDecision
//     (internal/diablo/inputbridge/session.go:49) has no window caller yet, so a
//     click does not walk.
//   - The sim walks the bare empty grid world.NewState builds while the presenter
//     shows the static retail-town scene: sim and picture are unrelated until the
//     town-consumption slice feeds one from the other (pre-existing).
//   - One settled tick per frame: the clock steps a fixed 50ms per frame
//     (frameDuration, window_shared.go) and the game settles one tick per 50ms,
//     so walking runs at frame rate rather than at a 20 TPS budget (cadence
//     record).
//
// bootPackSource is the PackSource the lane's create verb serves: the warrior
// class's starting pack, the only class the first demo raises. The diablo
// loaders own the numbers (player.cpp:2287-2330 calculateBaseLife/
// calculateBaseMana); the values below are the scripted walk's frozen warrior
// pack (internal/integration/session_stage_test.go:warriorStartingPack).
type bootPackSource struct{}

// StartingPack implements world.PackSource: the warrior class and no other.
func (bootPackSource) StartingPack(class int) (world.StartingPack, bool) {
	if class != 0 {
		return world.StartingPack{}, false
	}

	return world.StartingPack{
		Class: 0, BaseStr: 30, BaseMag: 10, BaseDex: 20, BaseVit: 25,
		HPBase: 96, MaxHPBase: 128, ManaBase: 8, MaxManaBase: 8,
	}, true
}

func bootDeviceLane(src *engine.MockTickSource, g *game.Game) *deviceLane {
	state := world.NewState()
	state.CreateFromPack(world.StartingPack{Class: 0}, world.Tile{X: heroSpawnX, Y: heroSpawnY})

	k := kernel.New(world.NewSessionGate(state, state, 0), kernel.NewStreams(kernelSeed, movementStream))
	world.RegisterMovement(k, state)
	world.RegisterCreatePack(k, state, lanePackSource)
	g.Kernel = k

	return &deviceLane{game: g, state: state, input: inputbridge.NewSession(k), clock: newFrameClock(src, g)}
}

// PushStick hands one classified left-stick decision to the submit site and
// reports whether the gate accepted it. The command carries the tick the next
// settled pass will run — the game advances TickCount in that same pass — so the
// window layer never invents a tick number of its own (ADR-018 law 1; the review
// s6e F2 stamping defect).
//
// inputbridge.Session.Tick is deliberately not used here: it also runs a kernel
// pass (internal/diablo/inputbridge/session.go:65-73), which would make the
// window layer a second pass driver. The decision goes through the bridge's own
// adapter and the session's submit contract instead.
func (l *deviceLane) PushStick(push controls.StickPush, dir controls.AxisDirection) bool {
	cmd, ok := inputbridge.StickCommand(push, dir)
	if !ok {
		return false
	}

	cmd.Tick = l.game.TickCount

	return l.input.SubmitDecision(cmd)
}

// frame runs one window frame in the order the single pass driver requires:
// readInput queues the frame's decisions first (StickDecision buffers, nothing
// dispatches), then the clock steps and the game's settled ticks — and only
// their passes — drain the queue. The window loop hands its raw key read in as
// readInput (windowGame.Update), so the ordering itself is part of this tested
// seam rather than the untested ebiten file (review s6e F2).
//
// The fixed 50ms step (frameDuration, window_shared.go) is why one settled tick
// per frame is the norm — the cadence the town slice will change.
func (l *deviceLane) frame(readInput func()) {
	readInput()
	l.clock.step(frameDuration)
	l.deliverCreateEvents()
}

// deliverCreateEvents feeds the create hook the first event of each tick the
// settled pass produced: the selhero flow reads CreatedPack (persist the row)
// or nothing (the kernel refused — no row, the refusal surfaces on the
// screen). At most one create decision is in flight.
func (l *deviceLane) deliverCreateEvents() {
	if l.onCreateEvent == nil {
		// Nothing to deliver, but the read cursor still advances so a hook
		// registered later never reads a stale tick's events as its own.
		l.tracedNext = l.game.TickCount

		return
	}

	for tick := l.tracedNext; tick < l.game.TickCount; tick++ {
		if ev, dispatched := l.createOutcomeAt(tick); dispatched {
			l.pendingCreate = false
			l.onCreateEvent(ev)

			break
		}
	}

	l.tracedNext = l.game.TickCount
}

// createOutcomeAt reports the create decision's outcome for one passed tick:
// the CreatedPack event when the handler raised the hero, or a zero event when
// the command dispatched and the handler refused it. dispatched is false while
// the tick carried no create decision.
func (l *deviceLane) createOutcomeAt(tick uint32) (kernel.Event, bool) {
	if !l.pendingCreate || !dispatchedCreate(l.game.Kernel.TraceCommandsAt(tick)) {
		return kernel.Event{}, false
	}

	return createdPackEvent(l.game.Kernel.TraceEventsAt(tick)), true
}

// dispatchedCreate reports whether a passed tick carried the create decision.
func dispatchedCreate(cmds []kernel.Command) bool {
	for _, cmd := range cmds {
		if cmd.Kind == kernel.KindCreatePack {
			return true
		}
	}

	return false
}

// createdPackEvent reports the tick's CreatedPack event, or a zero event while
// the handler stayed silent (the refusal the create flow reports).
func createdPackEvent(events []kernel.Event) kernel.Event {
	for _, ev := range events {
		if ev.Kind == kernel.KindCreatedPack {
			return ev
		}
	}

	return kernel.Event{}
}
