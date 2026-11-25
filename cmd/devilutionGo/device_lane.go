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

var lanePackSource world.PackSource = bootPackSource{}

// outcomeFlow is one in-flight decision outcome: a command kind, the event
// kind its handler announces, the pending flag the submitting screen sets,
// and the hook the lane invokes with the settled event exactly once.
type outcomeFlow struct {
	kind      kernel.Kind
	eventKind kernel.Kind
	pending   bool
	hook      func(kernel.Event)
}

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
//
// lanePackSource is the PackSource the create verb serves: the untagged
// default is the warrior-only boot source; the windowed boot installs the
// data-built classdat source (classdat_pack_source.go, review R1 F9). The
// variable indirection keeps this untagged file out of the scene dependency.
type deviceLane struct {
	game  *game.Game
	state *world.State
	input *inputbridge.Session
	clock *frameClock

	// tracedNext is the next tick whose events the hooks have not yet read:
	// the kernel's pass records events under the tick it ran as, which is
	// TickCount BEFORE the pass increments it.
	tracedNext uint32

	// outcomes are the decision-outcome flows the screens arm, in delivery
	// order: create (KindCreatePack -> KindCreatedPack), load (KindLoadSlot
	// -> KindLoadedSlot) and save (KindSaveSlot -> KindSavedSlot, the
	// game-menu row). Each carries one in-flight decision: the screen sets
	// pending when it submits and the hook fires exactly once with the
	// settled event — or a zero event on a refusal (the command dispatched,
	// the handler stayed silent). The third flow was the extraction point
	// (coding-rules 3rd occurrence).
	outcomes []outcomeFlow
}

// bootDeviceLane wires the kernel exactly as the windowed run boots it: a world
// slice, the session gate over it, a seeded kernel, the movement verbs, and the
// bridge session the window submits through.
//
// Disclosures (review s6e F3/F4/F5 — recorded here, updated devilGo-5fr4;
// rewritten at the S8 load slice):
//   - The hero raise was DIRECT state mutation; the selhero slice registers the
//     event-sourced CreatePack verb (world.RegisterCreatePack over
//     bootPackSource) and the selhero screen submits KindCreatePack. The boot
//     itself keeps the bootstrap raise so the live-view baseline tile stays
//     byte-identical before the front flow ever reaches selhero.
//   - The mouse lane IS wired: windowMovement (window_scene.go) reads the
//     arrows into PushStick, so a stick step walks. The mouse click-to-walk
//     path (inputbridge.Session.MouseDecision) has no window caller yet.
//   - The SaveLoad verbs register through lane.attachSaveLoad, called after
//     the boot opens the store (buildWindowGame): the archive is the per-hero
//     store, the LevelSource is town-only — a record saved inside a dungeon
//     restores nothing (the handler stays silent, ADR-018 law 2) until a
//     production dungeon corpus loader exists.
//   - The front flow owns the screen; FrontGame is entered only from the
//     LoadedSlot/CreatedPack outcome, never from a key.
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

// outcomeFor reports the lane's outcome flow for kind, or nil.
func (l *deviceLane) outcomeFor(kind kernel.Kind) *outcomeFlow {
	for i := range l.outcomes {
		if l.outcomes[i].kind == kind {
			return &l.outcomes[i]
		}
	}

	return nil
}

// armOutcome arms or replaces the outcome flow for kind: the screen sets the
// hook that fires on the settled pass's event.
func (l *deviceLane) armOutcome(kind, eventKind kernel.Kind, hook func(kernel.Event)) {
	if flow := l.outcomeFor(kind); flow != nil {
		flow.eventKind, flow.pending, flow.hook = eventKind, true, hook

		return
	}

	l.outcomes = append(l.outcomes, outcomeFlow{kind: kind, eventKind: eventKind, pending: true, hook: hook})
}

// attachSaveLoad registers the session's save and load verbs on the lane's
// kernel: the archive the boot opens (the production persist.SessionArchive)
// and the level source the restore asks for. Kernel.Register is a map write,
// legal any time before the first frame's pass, so the windowed boot calls
// this after the store opens without reordering bootDeviceLane.
func (l *deviceLane) attachSaveLoad(archive world.SessionArchive, levels world.LevelSource) {
	world.RegisterSaveLoad(l.game.Kernel, l.state, archive, levels)
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
	l.deliverOutcomes()

	// The read cursor advances once, after both flows have iterated: a hook
	// registered later never reads a stale tick's events as its own (review
	// F5), and neither flow consumes the other's ticks first.
	l.tracedNext = l.game.TickCount
}

// deliverOutcomes feeds every armed outcome hook the first event of the ticks
// the settled pass produced: the submitting screen reads the announced event
// (the decision landed) or a zero event (the kernel refused — the screen
// surfaces the refusal). At most one decision per kind is in flight.
func (l *deviceLane) deliverOutcomes() {
	for tick := l.tracedNext; tick < l.game.TickCount; tick++ {
		cmds := l.game.Kernel.TraceCommandsAt(tick)

		for i := range l.outcomes {
			flow := &l.outcomes[i]
			if !flow.pending || !dispatchedKind(cmds, flow.kind) {
				continue
			}

			flow.pending = false
			flow.hook(eventOfKind(l.game.Kernel.TraceEventsAt(tick), flow.eventKind))

			break
		}
	}
}

// dispatchedKind reports whether a passed tick carried a decision of kind.
func dispatchedKind(cmds []kernel.Command, kind kernel.Kind) bool {
	for _, cmd := range cmds {
		if cmd.Kind == kind {
			return true
		}
	}

	return false
}

// eventOfKind reports the tick's event of kind, or a zero event while the
// handler stayed silent (the refusal the submitting screen reports).
func eventOfKind(events []kernel.Event, kind kernel.Kind) kernel.Event {
	for _, ev := range events {
		if ev.Kind == kind {
			return ev
		}
	}

	return kernel.Event{}
}
