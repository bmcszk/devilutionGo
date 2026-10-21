package game

import (
	"time"

	"github.com/bmcszk/devilutionGo/internal/engine"
)

// gnTickDelay (diablo.cpp:146)
const gnTickDelay = 50 * time.Millisecond

// maxCatchUpWindow (nthread.cpp:219-233) is the stall length past which the
// tick clock is reset instead of catching up on the missed ticks.
const maxCatchUpWindow = 10 * gnTickDelay

// nTickRate (diablo.cpp:3492 sgGameInitInfo.nTickRate) is the game tick rate
// in ticks per second; gnTickDelay = 1000 / nTickRate.
const nTickRate = 20

// startupBurstTicks (diablo.cpp:3490) is the number of game ticks the startup
// burst may run: nTickRate * 3.
const startupBurstTicks = nTickRate * 3

// LevelType mirrors leveltype (dungeontype.h); the values are the dungeon
// type ids, DTYPE_TOWN is 0.
type LevelType int

// LevelTypeTown is the town level type (dungeontype.h DTYPE_TOWN = 0).
const LevelTypeTown LevelType = 0

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

	now      time.Duration
	lastTick time.Duration
	burstRun bool
}

// NewGame creates a Game driven by the given tick source.
func NewGame(tick engine.TickSource) *Game {
	return &Game{Tick: tick, now: tick.Now(), StartupBurst: true}
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
		g.lastTick += gnTickDelay
		ticks++
		if g.now-g.lastTick < gnTickDelay {
			break
		}
	}
	return ticks
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
	acc := elapsed / gnTickDelay
	if elapsed > maxCatchUpWindow {
		// nthread.cpp:219-233: in singleplayer a stall longer than ten ticks
		// resets last_tick to now and drops the elapsed ticks, so no catch-up
		// tick is run.
		g.lastTick = g.now
		acc = 0
	} else if acc > 0 {
		g.lastTick += acc * gnTickDelay
	}
	for range acc {
		g.GameLogic()
	}
	g.updateProgressToNextGameTick()
	return int(acc)
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

// updateProgressToNextGameTick mirrors nthread.cpp:246-260: uint8 in
// 0..128, 128 when a tick is due, else (gnTickDelay-ticksMissing)*128/gnTickDelay.
// Upstream last_tick is the next due tick time; our lastTick is the last
// executed tick, so ticksMissing = lastTick + gnTickDelay - now.
func (g *Game) updateProgressToNextGameTick() {
	ticksMissing := g.lastTick + gnTickDelay - g.now
	if ticksMissing <= 0 {
		g.ProgressToNextGameTick = 128
		return
	}
	fraction := (gnTickDelay - ticksMissing) * 128 / gnTickDelay
	g.ProgressToNextGameTick = uint8(min(max(fraction, 0), 128))
}

// ProcessInput (diablo.cpp:222) reports whether a game tick should be
// processed; the stub always returns true.
func (*Game) ProcessInput() bool { return true }

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
