package game_test

import (
	"testing"
	"time"

	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/game"
)

func TestGame_Update_RunsTwentyTicksOverOneSecond(t *testing.T) {
	// given
	fired := 0
	g := game.NewGame(&engine.MockTickSource{})
	g.OnTick = func() { fired++ }
	var returned int

	// when
	for range 60 {
		returned += g.Update(16_670_000)
	}

	// then
	if fired != 20 {
		t.Fatalf("OnTick fired %d times over one second, want 20 (gnTickDelay=50ms)", fired)
	}
	if returned != 20 {
		t.Fatalf("Update returned %d ticks over one second, want 20", returned)
	}
}

func TestGame_Update_LongStallRunsNoCatchUpTicks(t *testing.T) {
	// given
	fired := 0
	g := game.NewGame(&engine.MockTickSource{})
	g.StartupBurst = false
	g.OnTick = func() { fired++ }

	// when
	if got := g.Update(15 * 50 * time.Millisecond); got != 0 {
		t.Fatalf("Update(750ms) ran %d ticks, want 0 (nthread.cpp:219-233 drops missed ticks)", got)
	}

	// then
	if fired != 0 {
		t.Fatalf("OnTick fired %d times during the stall, want 0", fired)
	}

	if got := g.Update(10 * time.Millisecond); got != 0 {
		t.Fatalf("Update(10ms) after the stall ran %d ticks, want 0", got)
	}
	if got := g.Update(40 * time.Millisecond); got != 1 {
		t.Fatalf("Update(40ms) after the stall ran %d ticks, want 1", got)
	}
}

func TestGame_ProgressToNextGameTick_HandComputedAt25msFrames(t *testing.T) {
	// given
	// 25ms frames: the next due tick advances by 50ms every second frame,
	// so ticksMissing alternates 25,50 -> progress 64,0 (nthread.cpp:246-260).
	g := game.NewGame(&engine.MockTickSource{})
	g.StartupBurst = false
	want := []uint8{64, 0, 64, 0}

	got := make([]uint8, 0, len(want))
	// when
	for range len(want) {
		g.Update(25 * time.Millisecond)
		got = append(got, g.ProgressToNextGameTick)
	}

	// then
	for i, w := range want {
		if got[i] != w {
			t.Fatalf("frame %d progress = %d, want %d", i, got[i], w)
		}
	}
}

func TestGame_ProgressToNextGameTick_Is128WhenTickDue_AndStaysIn0To128(t *testing.T) {
	// given
	for _, burst := range []bool{true, false} {
		g := game.NewGame(&engine.MockTickSource{})
		g.StartupBurst = burst

		for range 120 {
			g.Update(7_300_000)
			if p := g.ProgressToNextGameTick; p > 128 {
				t.Fatalf("burst=%v progress = %d outside 0..128", burst, p)
			}
		}
	}

	// when: capped startup burst
	g := game.NewGame(&engine.MockTickSource{NowValue: 4 * time.Second})
	g.Update(0)

	// then
	if p := g.ProgressToNextGameTick; p != 128 {
		t.Fatalf("progress = %d when a tick is overdue, want 128", p)
	}
}

func TestGame_StartupBurst_DefaultsTrue_AndRunsTicksDueOnFirstUpdate(t *testing.T) {
	// given
	fired := 0
	g := game.NewGame(&engine.MockTickSource{})
	g.OnTick = func() { fired++ }

	// when
	got := g.Update(0)

	// then
	if !g.StartupBurst {
		t.Fatal("NewGame must default StartupBurst to true")
	}
	if got != 1 || fired != 1 {
		t.Fatalf("first Update ran %d ticks (OnTick fired %d), want 1 startup tick (no tick is due yet)", got, fired)
	}
}

func TestGame_StartupBurst_RunsTheTicksDueOnTheClock(t *testing.T) {
	// given
	tests := []struct {
		name  string
		clock time.Duration
		want  int
	}{
		{"no time elapsed", 0, 1},
		{"one tick due", 50 * time.Millisecond, 1},
		{"two ticks due", 100 * time.Millisecond, 2},
		{"capped at nTickRate*3", 3 * time.Second, 60},
		{"clock past the cap", 10 * time.Second, 60},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fired := 0
			g := game.NewGame(&engine.MockTickSource{NowValue: tc.clock})
			g.OnTick = func() { fired++ }

			got := g.Update(0)

			if got != tc.want || fired != tc.want {
				t.Fatalf("clock %v: first Update ran %d ticks (OnTick fired %d), want %d",
					tc.clock, got, fired, tc.want)
			}
		})
	}
}

func TestGame_StartupBurst_AdvancesTheTickClockByOneTickPerGameLogic(t *testing.T) {
	// given
	g := game.NewGame(&engine.MockTickSource{})
	g.Update(0)

	// when
	if got := g.Update(49 * time.Millisecond); got != 0 {
		t.Fatalf("Update(49ms) after the burst ran %d ticks, want 0 (the next tick is due at 100ms)", got)
	}

	// then
	if got := g.Update(51 * time.Millisecond); got != 1 {
		t.Fatalf("Update(51ms) after the burst ran %d ticks, want 1", got)
	}
}

func TestGame_StartupBurst_RunsOnlyOnce(t *testing.T) {
	// given
	g := game.NewGame(&engine.MockTickSource{})
	g.Update(0)

	// when
	got := g.Update(0)

	// then
	if got != 0 {
		t.Fatalf("second Update ran %d ticks, want 0", got)
	}
}

func TestGame_Update_RunsWithoutAllocations(t *testing.T) {
	// given
	g := game.NewGame(&engine.MockTickSource{})
	for range 10 {
		g.Update(16_670_000)
	}

	allocs := testing.AllocsPerRun(100, func() { g.Update(16_670_000) })

	if allocs != 0 {
		t.Fatalf("hot-path Update allocated %v objects per run, want 0", allocs)
	}
}
