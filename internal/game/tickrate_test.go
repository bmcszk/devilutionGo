package game_test

import (
	"testing"
	"time"

	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/game"
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
)

func TestGameLogicSkipsWhileMenuOpen(t *testing.T) {
	// given: a single-player game with the game menu open
	g := newMenuGame(func() bool { return true })

	// when: one second of frames runs
	for range 60 {
		g.Update(16_670_000)
	}

	// then: no tick is processed (X:diablo.cpp:229-232 ProcessInput returns
	// false while the menu is open in single-player; diablo.cpp:1524-1526
	// GameLogic bails before any phase) — the frame clock keeps running, the
	// tick WORK does not. TickCount is the settled-tick counter.
	if g.TickCount != 0 {
		t.Fatalf("TickCount advanced %d with the menu open, want 0", g.TickCount)
	}
}

func TestGameLogicRunsWhileMenuOpenInMultiplayer(t *testing.T) {
	// given: a multiplayer game with the menu open (upstream does not pause
	// in MP, X:diablo.cpp:231)
	g := newMenuGame(func() bool { return true })
	g.Multiplayer = true

	// when: one second of frames runs
	for range 60 {
		g.Update(16_670_000)
	}

	// then: the ticks run at the default 20 per second
	if g.TickCount != 20 {
		t.Fatalf("TickCount advanced %d in multiplayer with the menu open, want 20", g.TickCount)
	}
}

func TestUpdateWhilePaused_AdvancesOneTickPerFrame(t *testing.T) {
	// given: a single-player game with the Esc menu open, its clock settled on
	// a tick boundary
	g := newMenuGame(func() bool { return true })
	worldTicks := 0
	g.OnTick = func() { worldTicks++ }

	// when: four 50ms frames run through the paused clock (the production
	// frameDuration)
	const frames = 4
	for range frames {
		g.UpdateWhilePaused(50 * time.Millisecond)
	}

	// then: one kernel pass per DUE tick, not one per second frame (review
	// F10: the in-loop advance moved lastTick twice per due tick, so the paused
	// pass fired every other frame)
	if g.TickCount != frames {
		t.Fatalf("TickCount = %d after %d paused frames, want %d (one pass per due tick)",
			g.TickCount, frames, frames)
	}

	// and the world itself never ticks while the menu holds it
	// (X:diablo.cpp:229-232)
	if worldTicks != 0 {
		t.Fatalf("world ticks = %d across %d paused frames, want 0", worldTicks, frames)
	}
}

func TestUpdateWhilePaused_ResumeRunsOneTickPerFrame(t *testing.T) {
	// given: an ODD number of paused frames from a settled state — the parity
	// where the double advance leaves lastTick AHEAD of now, so the resume
	// clamps elapsed to 0 and stalls (review F10's trace)
	g := newMenuGame(func() bool { return true })
	worldTicks := 0
	g.OnTick = func() { worldTicks++ }

	const pausedFrames = 3
	for range pausedFrames {
		g.UpdateWhilePaused(50 * time.Millisecond)
	}

	pausedTicks := g.TickCount

	// when: the menu closes and the first world frame runs
	g.MenuOpen = nil
	ran := g.Update(50 * time.Millisecond)

	// then: exactly one tick — the paused clock left no debt to stall on
	// (now-lastTick == 0, review F10's 50ms-per-two-frames drift) and no
	// catch-up burst fires
	if ran != 1 {
		t.Fatalf("first resumed Update(50ms) ran %d ticks, want 1 (paused clock left lastTick behind)", ran)
	}

	if g.TickCount != pausedTicks+1 {
		t.Fatalf("TickCount = %d after the resume, want %d (exactly one settled tick)",
			g.TickCount, pausedTicks+1)
	}

	if worldTicks != 1 {
		t.Fatalf("world ticks = %d after the resume, want 1 (one GameLogic pass)", worldTicks)
	}
}

func TestSetTickRateSpeedsTheClockLive(t *testing.T) {
	// given: a game with the tick-rate verb registered
	g := newMenuGame(nil)
	k := kernel.New(nil, nil)
	game.RegisterTickRate(k, g, nil)
	g.Kernel = k

	// when: a rate-50 decision is submitted and one frame settles
	if !k.Submit(kernel.Command{Kind: kernel.KindSetTickRate, A: 50}) {
		t.Fatal("gate refused KindSetTickRate A:50")
	}
	g.Update(50 * time.Millisecond)

	// then: exactly one TickRateSet event, and the NEXT 50ms frame runs two
	// ticks (gnTickDelay = 1000/50 = 20ms)
	if events := tickRateSetEvents(k, 0); events != 1 {
		t.Fatalf("TickRateSet events = %d, want exactly 1", events)
	}
	if got := g.Update(50 * time.Millisecond); got != 2 {
		t.Fatalf("Update(50ms) after the rate change ran %d ticks, want 2", got)
	}
	if g.TickDelay() != 20*time.Millisecond {
		t.Fatalf("TickDelay() = %v, want 20ms (1000/50)", g.TickDelay())
	}
}

func TestSetTickRateEmitsNothingWhenTheRateIsUnchanged(t *testing.T) {
	// given: the rate already at 50
	g := newMenuGame(nil)
	k := kernel.New(nil, nil)
	game.RegisterTickRate(k, g, nil)
	g.Kernel = k
	k.Submit(kernel.Command{Kind: kernel.KindSetTickRate, A: 50})
	g.Update(50 * time.Millisecond)

	// when: the same rate is submitted again and the frame settles
	k.Submit(kernel.Command{Kind: kernel.KindSetTickRate, A: 50})
	g.Update(50 * time.Millisecond)

	// then: no second event fires (law 2: no change, no event)
	for i := uint32(1); i < g.TickCount; i++ {
		for _, ev := range k.TraceEventsAt(i) {
			if ev.Kind == kernel.KindTickRateSet {
				t.Fatalf("TickRateSet fired again at tick %d for an unchanged rate", i)
			}
		}
	}
}

func TestSetTickRateCallsOnRateExactlyOnce(t *testing.T) {
	// given: an onRate sink
	g := newMenuGame(nil)
	k := kernel.New(nil, nil)
	calls := 0
	game.RegisterTickRate(k, g, func(rate int) { calls += rate })
	g.Kernel = k

	// when: one rate change settles
	k.Submit(kernel.Command{Kind: kernel.KindSetTickRate, A: 50})
	g.Update(50 * time.Millisecond)

	// then: onRate saw 50 exactly once
	if calls != 50 {
		t.Fatalf("onRate calls sum = %d, want 50 (exactly once)", calls)
	}
}

func TestSetTickRateDefaultPathStaysAtTwentyPerSecond(t *testing.T) {
	// given: a game with no rate change submitted
	g := newMenuGame(nil)
	k := kernel.New(nil, nil)
	game.RegisterTickRate(k, g, nil)
	g.Kernel = k

	// when: one second of frames runs
	for range 60 {
		g.Update(16_670_000)
	}

	// then: the default 20-per-second pace holds
	if g.TickCount != 20 {
		t.Fatalf("TickCount advanced %d at the default, want 20", g.TickCount)
	}
}

// newMenuGame builds a non-burst game over a mock tick source with the given
// menu-open state.
func newMenuGame(menuOpen func() bool) *game.Game {
	g := game.NewGame(&engine.MockTickSource{})
	g.StartupBurst = false
	g.MenuOpen = menuOpen

	return g
}

// tickRateSetEvents counts the tick's TickRateSet events carrying rate.
func tickRateSetEvents(k *kernel.Kernel, tick uint32) int {
	events := 0
	for _, ev := range k.TraceEventsAt(tick) {
		if ev.Kind == kernel.KindTickRateSet && ev.A == 50 {
			events++
		}
	}

	return events
}
