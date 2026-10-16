package game

import (
	"time"

	"github.com/bmcszk/devilutionGo/internal/engine"
)

// gnTickDelay (diablo.cpp:146): the game logic tick length. All game state
// advances at 1000/50 = 20 ticks per second.
const gnTickDelay = 50 * time.Millisecond

// maxTicksPerUpdate (nthread.cpp:214-243): accumulated ticks are clamped to
// this and last_tick realigned, so a long stall can never cause a catch-up
// avalanche of GameLogic passes.
const maxTicksPerUpdate = 10

// startupBurstTicks (diablo.cpp:3490, game_loop): the first frame runs this
// many immediate ticks (nTickRate*3 in devilutionX terms) behind a bStartup
// flag so the game enters a valid state before pacing begins.
const startupBurstTicks = 3

// Game drives the fixed-timestep game logic loop: it converts elapsed time
// into discrete ticks and runs one GameLogic phase pass per tick.
type Game struct {
	// TickSource anchors the game timeline; NewGame seeds the timeline from
	// it and Update advances the timeline by dt each frame.
	Tick engine.TickSource
	// StartupBurst mirrors devilutionX game_loop (diablo.cpp:3490): the first
	// Update runs three immediate ticks, charged against the first
	// 150ms of accumulated time so the 20 ticks/sec average holds.
	StartupBurst bool
	// ProgressToNextGameTick (nthread.cpp:246, UpdateProgressToNextGameTick):
	// fraction of the current tick already elapsed, always in [0,1); render
	// interpolation reads this.
	ProgressToNextGameTick float64
	// OnTick, when set, is invoked once per GameLogic pass (test hook).
	OnTick func()

	now      time.Duration
	lastTick time.Duration
	burstRun bool
}

// NewGame returns a Game whose timeline starts at tick's current reading and
// with the devilutionX startup burst enabled.
func NewGame(tick engine.TickSource) *Game {
	return &Game{Tick: tick, now: tick.Now(), StartupBurst: true}
}

// Update consumes dt of elapsed time, runs the accumulated game logic ticks
// (clamped as nthread.cpp:214-243 does), and returns how many ticks ran.
// runStartupBurst prepays three ticks of game time at the current timeline
// position, so a one-second window still yields exactly 20 GameLogic passes
// (diablo.cpp:3490).
func (g *Game) runStartupBurst() int {
	g.burstRun = true
	g.lastTick += startupBurstTicks * gnTickDelay
	for range startupBurstTicks {
		g.GameLogic()
	}
	return startupBurstTicks
}

func (g *Game) Update(dt time.Duration) int {
	if g.StartupBurst && !g.burstRun {
		n := g.runStartupBurst()
		g.now += dt
		g.updateProgressToNextGameTick()
		return n
	}

	g.now += dt
	acc := (g.now - g.lastTick) / gnTickDelay
	if acc < 0 {
		acc = 0
	}
	if acc > maxTicksPerUpdate {
		acc = maxTicksPerUpdate
		g.lastTick = g.now
	} else if acc > 0 {
		g.lastTick += acc * gnTickDelay
	}
	for range acc {
		g.GameLogic()
	}
	g.updateProgressToNextGameTick()
	return int(acc)
}

// GameLogic runs one full game logic pass in the devilutionX order
// (diablo.cpp:1523-1568). Each phase is a named no-op until ported.
func (g *Game) GameLogic() {
	g.ProcessInput()
	g.ProcessPlayers()
	g.ProcessMonsters()
	g.ProcessObjects()
	g.ProcessMissiles()
	g.ProcessItems()
	g.ProcessLightList()
	g.ProcessVisionList()
	g.ProcessSound()
	g.ProcessTriggers()
	g.ProcessQuests()
	if g.OnTick != nil {
		g.OnTick()
	}
}

// updateProgressToNextGameTick mirrors UpdateProgressToNextGameTick
// (nthread.cpp:246): the fraction of the current tick already elapsed,
// clamped into [0,1). It never reaches 1 because lastTick advances in whole
// ticks up to now.
func (g *Game) updateProgressToNextGameTick() {
	g.ProgressToNextGameTick = max(float64(g.now-g.lastTick)/float64(gnTickDelay), 0)
}

// ProcessInput runs the per-tick input pass (diablo.cpp:1523).
func (*Game) ProcessInput() {}

// ProcessPlayers runs the per-tick player pass (diablo.cpp:1529).
func (*Game) ProcessPlayers() {}

// ProcessMonsters runs the per-tick monster AI pass (diablo.cpp:1533).
func (*Game) ProcessMonsters() {}

// ProcessObjects runs the per-tick object pass (diablo.cpp:1536).
func (*Game) ProcessObjects() {}

// ProcessMissiles runs the per-tick missile pass (diablo.cpp:1538).
func (*Game) ProcessMissiles() {}

// ProcessItems runs the per-tick item pass (diablo.cpp:1540).
func (*Game) ProcessItems() {}

// ProcessLightList runs the per-tick light refresh (diablo.cpp:1542).
func (*Game) ProcessLightList() {}

// ProcessVisionList runs the per-tick vision refresh (diablo.cpp:1544).
func (*Game) ProcessVisionList() {}

// ProcessSound runs the per-tick sound maintenance (diablo.cpp:1546).
func (*Game) ProcessSound() {}

// ProcessTriggers runs the per-tick trigger checks (diablo.cpp:1550).
func (*Game) ProcessTriggers() {}

// ProcessQuests runs the per-tick quest pass (diablo.cpp:1556).
func (*Game) ProcessQuests() {}
