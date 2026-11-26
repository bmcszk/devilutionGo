package game

import (
	"time"

	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
)

// gnTickDelay (diablo.cpp:146)
// maxCatchUpWindow (nthread.cpp:219-233) is the stall length past which the
// tick clock is reset instead of catching up on the missed ticks.
// nTickRate (diablo.cpp:3492 sgGameInitInfo.nTickRate) is the game tick rate
// in ticks per second; gnTickDelay = 1000 / nTickRate.
// startupBurstTicks (diablo.cpp:3490) is the number of game ticks the startup
// burst may run: nTickRate * 3.
const (
	gnTickDelay       = 50 * time.Millisecond
	maxCatchUpWindow  = 10 * gnTickDelay
	nTickRate         = 20
	startupBurstTicks = nTickRate * 3
)

// LevelTypeTown is the town level type (dungeontype.h DTYPE_TOWN = 0).
const LevelTypeTown LevelType = 0

// LevelType mirrors leveltype (dungeontype.h); the values are the dungeon
// type ids, DTYPE_TOWN is 0.
type LevelType int

// Game is the shared game state that sits on the engine core: it advances
// the tick clock and runs the game-logic phase order.
type Game struct {
	Tick                   engine.TickSource
	StartupBurst           bool
	ProgressToNextGameTick uint8
	OnTick                 func()

	// LevelType is the type of the current level; the zero value is the town
	// (diablo.cpp:1532 branches on leveltype != DTYPE_TOWN).
	LevelType LevelType

	// Doom is the map of the stars portal; nil until the item-use path raises
	// it (items.cpp:4319-4321 IMISC_MAPOFDOOM) and cleared by Escape
	// (diablo.cpp:355-356) or the chat log opening (chatlog.cpp:114).
	Doom *DoomPortal
	// Debug is the cheat router; nil unless the build opted in through the
	// --debug flag (NewDebugRouterFromFlag). A nil router makes every cheat a
	// no-op.
	Debug *DebugRouter

	// Kernel is the deterministic command/event kernel the settled ticks feed
	// (ADR-015, ADR-018). A nil kernel runs the phase stubs with no command
	// pass, so an unwired Game behaves as before.
	Kernel *kernel.Kernel
	// TickCount is the number of settled ticks handed to the kernel: the tick
	// stamped on commands and events (ADR-018 law 1).
	TickCount uint32

	// MenuOpen reports whether the in-game menu is open (gmenu_is_active).
	// Presentation-owned display state read by ProcessInput
	// (X:diablo.cpp:229-232); nil means never open.
	MenuOpen func() bool

	// Multiplayer mirrors gbIsMultiplayer: upstream does not pause in MP
	// (X:diablo.cpp:231).
	Multiplayer bool

	now      time.Duration
	lastTick time.Duration
	burstRun bool
	// tickDelay is the live gnTickDelay (X:diablo.h:87): the SetTickRate
	// handler writes it (X:gamemenu.cpp:277); zero means the default const.
	tickDelay time.Duration
}

// NewGame creates a Game driven by the given tick source.
func NewGame(tick engine.TickSource) *Game {
	return &Game{Tick: tick, now: tick.Now(), StartupBurst: true}
}

// Update advances the clock by dt and runs the due game ticks; it returns
// the number of ticks executed.
func (g *Game) Update(dt time.Duration) int {
	if g.StartupBurst && !g.burstRun {
		n := g.runStartupBurst()
		g.now += dt
		g.updateProgressToNextGameTick()

		return n
	}

	g.now += dt
	elapsed := max(g.now-g.lastTick, 0)

	acc := elapsed / g.TickDelay()
	if elapsed > g.catchUpWindow() {
		// nthread.cpp:219-233: in singleplayer a stall longer than ten ticks
		// resets last_tick to now and drops the elapsed ticks, so no catch-up
		// tick is run.
		g.lastTick = g.now
		acc = 0
	} else if acc > 0 {
		g.lastTick += acc * g.TickDelay()
	}

	for range acc {
		g.GameLogic()
	}

	g.updateProgressToNextGameTick()

	return int(acc)
}

// UpdateWhilePaused advances the clock by dt while the Esc menu holds the
// world: NO game tick runs (GameLogic's ProcessInput gate is untouched —
// X:diablo.cpp:229-232), but the kernel pass runs once per due tick so the
// command queue drains and a menu decision dispatches on a tick that runs
// (review F1 fix shape b1). The lastTick bookkeeping follows the same
// cadence Update uses, so resuming the world continues the schedule.
func (g *Game) UpdateWhilePaused(dt time.Duration) {
	if g.StartupBurst && !g.burstRun {
		// The startup burst still owes its ticks: run it as Update would, so
		// the paused frames do not skip the boot's opening cadence.
		g.runStartupBurst()
		g.now += dt
		g.updateProgressToNextGameTick()

		return
	}

	g.now += dt
	elapsed := max(g.now-g.lastTick, 0)

	acc := elapsed / g.TickDelay()
	if elapsed > g.catchUpWindow() {
		g.lastTick = g.now
		acc = 0
	} else if acc > 0 {
		g.lastTick += acc * g.TickDelay()
	}

	for range acc {
		g.RunKernelPassWhilePaused()
		g.lastTick += g.TickDelay()
	}

	g.updateProgressToNextGameTick()
}

// GameLogic mirrors diablo.cpp:1523-1568; the phases are no-ops until ported.
// The order is: bail out when ProcessInput reports no game tick, run the
// player phase, then either the dungeon phases (monsters, objects, missiles,
// items, light, vision) or the town phases (towners, items, missiles),
// followed by the sound, trigger and quest phases.
func (g *Game) GameLogic() {
	if !g.ProcessInput() {
		return
	}
	g.runGameTick()
}

// RunKernelPassWhilePaused hands the settled tick to the kernel WITHOUT the
// world/sim phases: the Esc-menu pause gate (ProcessInput, X:diablo.cpp:229-232)
// stays in force for the world, but the kernel's own pass drains the command
// queue, so a menu command (SetTickRate) dispatches on a tick that runs
// (review hl4 F1, fix shape b1). The tick counter advances only here.
func (g *Game) RunKernelPassWhilePaused() {
	g.runKernelPass()
}

// ProcessInput (diablo.cpp:222) reports whether a game tick should be
// processed: never in multiplayer-with-menu-pause disagreement — multiplayer
// always processes (X:diablo.cpp:231), a single-player game with the menu open
// does not (:229-232), anything else does. The PauseMode == 2 focus-pause arm
// upstream carries has no counterpart yet (deferred, not a deviation).
func (g *Game) ProcessInput() bool {
	if g.Multiplayer {
		return true
	}

	if g.MenuOpen != nil && g.MenuOpen() {
		return false
	}

	return true
}

// TickDelay returns the live tick delay: gnTickDelay when no rate change has
// landed yet (X:diablo.cpp:146 gnTickDelay = 50), else 1000/rate
// (X:gamemenu.cpp:277).
func (g *Game) TickDelay() time.Duration {
	if g.tickDelay == 0 {
		return gnTickDelay
	}

	return g.tickDelay
}

// ProcessPlayers (diablo.cpp:1530).
func (*Game) ProcessPlayers() {}

// ProcessMonsters (diablo.cpp:1537).
func (*Game) ProcessMonsters() {}

// ProcessObjects (diablo.cpp:1539).
func (*Game) ProcessObjects() {}

// ProcessMissiles (diablo.cpp:1541 dungeon branch, diablo.cpp:1552 town branch).
func (*Game) ProcessMissiles() {}

// ProcessItems (diablo.cpp:1543 dungeon branch, diablo.cpp:1550 town branch).
func (*Game) ProcessItems() {}

// ProcessLightList (diablo.cpp:1544).
func (*Game) ProcessLightList() {}

// ProcessVisionList (diablo.cpp:1545).
func (*Game) ProcessVisionList() {}

// ProcessTowners (diablo.cpp:1548).
func (*Game) ProcessTowners() {}

// ProcessSound (diablo.cpp:1562 sound_update).
func (*Game) ProcessSound() {}

// ProcessTriggers (diablo.cpp:1563 CheckTriggers).
func (*Game) ProcessTriggers() {}

// ProcessQuests (diablo.cpp:1564 CheckQuests).
func (*Game) ProcessQuests() {}

// runGameTick runs ONE game tick's phases after a settled ProcessInput: the
// kernel pass (TickCount advances here, runKernelPass), then the phase order.
// Split from GameLogic so the kernel pass alone has a menu-independent entry:
// RunKernelPassWhilePaused feeds the command queue while the Esc menu holds
// the world (review hl4 F1 fix, shape b1) — the SIM BODY stays gated by
// ProcessInput exactly as upstream (X:diablo.cpp:229-232); only the kernel
// pass runs, so a queued SetTickRate lands on a tick that actually runs.
func (g *Game) runGameTick() {
	g.runKernelPass()

	// diablo.cpp:1528-1531 guards ProcessPlayers with gbProcessPlayers, which
	// is not ported yet, so the player phase always runs.
	g.ProcessPlayers()

	if g.LevelType != LevelTypeTown {
		g.ProcessMonsters()
		g.ProcessObjects()
		g.ProcessMissiles()
		g.ProcessItems()
		g.ProcessLightList()
		g.ProcessVisionList()
	} else {
		g.ProcessTowners()
		g.ProcessItems()
		g.ProcessMissiles()
	}

	g.ProcessSound()
	g.ProcessTriggers()
	g.ProcessQuests()

	if g.OnTick != nil {
		g.OnTick()
	}
}

// catchUpWindow is maxCatchUpWindow derived from the live tick delay
// (nthread.cpp:219-233: ten ticks of stall).
func (g *Game) catchUpWindow() time.Duration {
	return 10 * g.TickDelay()
}

// runStartupBurst mirrors diablo.cpp:3490-3501 game_loop(bStartup): it runs
// GameLogic up to startupBurstTicks times, advancing the tick clock by one
// tick per call, and stops as soon as no further tick is due
// (!nthread_has_500ms_passed, nthread.cpp:214-246).
func (g *Game) runStartupBurst() int {
	g.burstRun = true
	ticks := 0

	for range startupBurstTicks {
		g.GameLogic()
		g.lastTick += g.TickDelay()
		ticks++

		// nthread.cpp:214-246: the burst stops as soon as no further tick is
		// due (now has not caught up to the advanced lastTick)
		if g.now-g.lastTick < g.TickDelay() {
			break
		}
	}

	return ticks
}

// updateProgressToNextGameTick mirrors nthread.cpp:246-260: uint8 in
// 0..128, 128 when a tick is due, else (gnTickDelay-ticksMissing)*128/gnTickDelay.
// Upstream last_tick is the next due tick time; our lastTick is the last
// executed tick, so ticksMissing = lastTick + gnTickDelay - now.
func (g *Game) updateProgressToNextGameTick() {
	delay := g.TickDelay()
	ticksMissing := g.lastTick + delay - g.now
	if ticksMissing <= 0 {
		g.ProgressToNextGameTick = 128
		return
	}

	fraction := (delay - ticksMissing) * 128 / delay
	g.ProgressToNextGameTick = uint8(min(max(fraction, 0), 128))
}

// runKernelPass hands the settled tick to the kernel: exactly one ordered
// pass, whose events land in the kernel's trace for presentation and the tick
// log. The tick counter advances here, so the next tick's commands carry the
// next tick number even when no kernel is wired.
func (g *Game) runKernelPass() {
	if g.Kernel != nil {
		g.Kernel.Pass(g.TickCount)
	}

	g.TickCount++
}
