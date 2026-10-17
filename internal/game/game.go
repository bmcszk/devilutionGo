package game

import (
	"time"

	"github.com/bmcszk/devilutionGo/internal/engine"
)

// gnTickDelay (diablo.cpp:146)
const gnTickDelay = 50 * time.Millisecond

// maxTicksPerUpdate (nthread.cpp:214-243)
const maxTicksPerUpdate = 10

// startupBurstTicks (diablo.cpp:3490)
const startupBurstTicks = 3

type Game struct {
	Tick                   engine.TickSource
	StartupBurst           bool
	ProgressToNextGameTick float64
	OnTick                 func()

	now      time.Duration
	lastTick time.Duration
	burstRun bool
}

func NewGame(tick engine.TickSource) *Game {
	return &Game{Tick: tick, now: tick.Now(), StartupBurst: true}
}

// runStartupBurst advances lastTick as if 3 ticks elapsed before the first
// Update (diablo.cpp:3490 runs GameLogic 3x before the main loop).
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

// GameLogic mirrors diablo.cpp:1523-1568 phase order; phases are no-ops until ported.
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

// updateProgressToNextGameTick mirrors nthread.cpp:246 progressToNextGameTick.
func (g *Game) updateProgressToNextGameTick() {
	g.ProgressToNextGameTick = max(float64(g.now-g.lastTick)/float64(gnTickDelay), 0)
}

// ProcessInput (diablo.cpp:1523).
func (*Game) ProcessInput() {}

// ProcessPlayers (diablo.cpp:1529).
func (*Game) ProcessPlayers() {}

// ProcessMonsters (diablo.cpp:1533).
func (*Game) ProcessMonsters() {}

// ProcessObjects (diablo.cpp:1536).
func (*Game) ProcessObjects() {}

// ProcessMissiles (diablo.cpp:1538).
func (*Game) ProcessMissiles() {}

// ProcessItems (diablo.cpp:1540).
func (*Game) ProcessItems() {}

// ProcessLightList (diablo.cpp:1542).
func (*Game) ProcessLightList() {}

// ProcessVisionList (diablo.cpp:1544).
func (*Game) ProcessVisionList() {}

// ProcessSound (diablo.cpp:1546).
func (*Game) ProcessSound() {}

// ProcessTriggers (diablo.cpp:1550).
func (*Game) ProcessTriggers() {}

// ProcessQuests (diablo.cpp:1556).
func (*Game) ProcessQuests() {}
