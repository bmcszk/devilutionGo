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
	stall := g.Update(15 * 50 * time.Millisecond)
	afterStall := g.Update(10 * time.Millisecond)
	nextDue := g.Update(40 * time.Millisecond)

	// then
	if stall != 0 {
		t.Fatalf("Update(750ms) ran %d ticks, want 0 (nthread.cpp:219-233 drops missed ticks)", stall)
	}

	if afterStall != 0 {
		t.Fatalf("Update(10ms) after the stall ran %d ticks, want 0", afterStall)
	}

	if nextDue != 1 {
		t.Fatalf("Update(40ms) after the stall ran %d ticks, want 1", nextDue)
	}

	if fired != 1 {
		t.Fatalf("OnTick fired %d times across the three updates, want 1 (only the final due tick)", fired)
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
	for range want {
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
	// given one game per startup-burst setting, plus a game 4s ahead of its clock
	bursts := []bool{true, false}
	games := make([]*game.Game, len(bursts))
	for i, burst := range bursts {
		games[i] = game.NewGame(&engine.MockTickSource{})
		games[i].StartupBurst = burst
	}

	overdue := game.NewGame(&engine.MockTickSource{NowValue: 4 * time.Second})

	// when each game runs 120 frames and the overdue one consumes its burst
	frames := make([][]uint8, len(games))
	for i, g := range games {
		frames[i] = driveProgress(g, 120, 7_300_000)
	}

	overdue.Update(0)

	// then every reading stays in 0..128 and an overdue tick reports 128
	t.Run("progress stays in 0..128 across frames", func(t *testing.T) {
		for i, burst := range bursts {
			assertProgressInsideRange(t, burst, frames[i])
		}
	})

	t.Run("progress is 128 when a tick is overdue", func(t *testing.T) {
		if p := overdue.ProgressToNextGameTick; p != 128 {
			t.Fatalf("progress = %d when a tick is overdue, want 128", p)
		}
	})
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

	// when every case runs its first Update
	gots := make([]int, len(tests))
	fireds := make([]int, len(tests))
	for i, tc := range tests {
		fired := 0
		g := game.NewGame(&engine.MockTickSource{NowValue: tc.clock})
		g.OnTick = func() { fired++ }

		gots[i] = g.Update(0)
		fireds[i] = fired
	}

	// then each case runs exactly the ticks due on its clock
	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assertStartupTicks(t, tc.clock, gots[i], fireds[i], tc.want)
		})
	}
}

func TestGame_StartupBurst_AdvancesTheTickClockByOneTickPerGameLogic(t *testing.T) {
	// given
	g := game.NewGame(&engine.MockTickSource{})
	g.Update(0)

	// when
	after49 := g.Update(49 * time.Millisecond)
	after100 := g.Update(51 * time.Millisecond)

	// then
	if after49 != 0 {
		t.Fatalf("Update(49ms) after the burst ran %d ticks, want 0 (the next tick is due at 100ms)", after49)
	}

	if after100 != 1 {
		t.Fatalf("Update(51ms) after the burst ran %d ticks, want 1", after100)
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

	// when
	allocs := testing.AllocsPerRun(100, func() { g.Update(16_670_000) })

	// then
	if allocs != 0 {
		t.Fatalf("hot-path Update allocated %v objects per run, want 0", allocs)
	}
}

// driveProgress runs frames of dt nanoseconds and returns every progress
// reading.
func driveProgress(g *game.Game, frames int, dt time.Duration) []uint8 {
	got := make([]uint8, 0, frames)

	for range frames {
		g.Update(dt)
		got = append(got, g.ProgressToNextGameTick)
	}

	return got
}

// assertProgressInsideRange pins every progress reading inside 0..128.
func assertProgressInsideRange(t *testing.T, burst bool, got []uint8) {
	t.Helper()

	for i, p := range got {
		if p > 128 {
			t.Fatalf("startupBurst=%v frame %d progress = %d outside 0..128", burst, i, p)
		}
	}
}

// assertStartupTicks pins the ticks run and OnTick firings for one clock.
func assertStartupTicks(t *testing.T, clock time.Duration, got, fired, want int) {
	t.Helper()

	if got != want || fired != want {
		t.Fatalf("clock %v: first Update ran %d ticks (OnTick fired %d), want %d",
			clock, got, fired, want)
	}
}
