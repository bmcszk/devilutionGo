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
type deviceLane struct {
	game  *game.Game
	state *world.State
	input *inputbridge.Session
	clock *frameClock
}

// bootDeviceLane wires the kernel exactly as the windowed run boots it: a world
// slice, the session gate over it, a seeded kernel, the movement verbs, and the
// bridge session the window submits through.
//
// Disclosures (review s6e F3/F4/F5 — recorded here, deliberately not fixed in
// this slice):
//   - The hero raise is DIRECT state mutation, not the event-sourced CreatePack
//     verb this same kernel already carries (world.RegisterCreatePack,
//     internal/game/world/session.go:226-228). It is the bootstrap shortcut for
//     this slice and its bool return is deliberately ignored: a refused raise
//     (a solid spawn tile) would leave the session hero-less and boot silently
//     without one. The selhero slice replaces the call with KindCreatePack.
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
func bootDeviceLane(src *engine.MockTickSource, g *game.Game) *deviceLane {
	state := world.NewState()
	state.CreateFromPack(world.StartingPack{Class: 0}, world.Tile{X: heroSpawnX, Y: heroSpawnY})

	k := kernel.New(world.NewSessionGate(state, state, 0), kernel.NewStreams(kernelSeed, movementStream))
	world.RegisterMovement(k, state)
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
}
