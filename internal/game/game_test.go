package game_test

import (
	"testing"
	"time"

	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/game"
)

func TestGame_Update_RunsTwentyTicksOverOneSecond(t *testing.T) {
	fired := 0
	g := game.NewGame(&engine.MockTickSource{})
	g.OnTick = func() { fired++ }
	var returned int

	for range 60 {
		returned += g.Update(16_670_000)
	}

	if fired != 20 {
		t.Fatalf("OnTick fired %d times over one second, want 20 (gnTickDelay=50ms)", fired)
	}
	if returned != 20 {
		t.Fatalf("Update returned %d ticks over one second, want 20", returned)
	}
}

func TestGame_Update_ClampsLargeStepsToTenTicks(t *testing.T) {
	tests := []struct {
		name string
		step time.Duration
		want int
	}{
		{"exactly ten ticks at 500ms", 500 * time.Millisecond, 10},
		{"clamped from eleven at 550ms", 550 * time.Millisecond, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame(&engine.MockTickSource{})
			g.StartupBurst = false

			got := g.Update(tt.step)

			if got != tt.want {
				t.Fatalf("Update(%v) ran %d ticks, want %d", tt.step, got, tt.want)
			}
		})
	}
}

func TestGame_Update_RealignsAfterClampWithoutCatchUp(t *testing.T) {
	fired := 0
	g := game.NewGame(&engine.MockTickSource{})
	g.StartupBurst = false
	g.OnTick = func() { fired++ }
	if got := g.Update(500 * time.Millisecond); got != 10 {
		t.Fatalf("precondition: 500ms step ran %d ticks, want the 10-tick clamp", got)
	}
	fired = 0

	for range 60 {
		g.Update(16_670_000)
	}

	if fired != 20 {
		t.Fatalf("second after 500ms clamp ran %d ticks, want exactly 20 with no catch-up", fired)
	}
}

func TestGame_ProgressToNextGameTick_HandComputedAt25msFrames(t *testing.T) {
	g := game.NewGame(&engine.MockTickSource{})
	g.StartupBurst = false
	want := []float64{0.5, 0, 0.5, 0}

	got := make([]float64, 0, len(want))
	for range len(want) {
		g.Update(25 * time.Millisecond)
		got = append(got, g.ProgressToNextGameTick)
	}

	for i, w := range want {
		if got[i] != w {
			t.Fatalf("frame %d progress = %v, want %v", i, got[i], w)
		}
	}
}

func TestGame_ProgressToNextGameTick_StaysWithinHalfOpenUnitInterval(t *testing.T) {
	for _, burst := range []bool{true, false} {
		g := game.NewGame(&engine.MockTickSource{})
		g.StartupBurst = burst

		for range 120 {
			g.Update(7_300_000)
			if p := g.ProgressToNextGameTick; p < 0 || p >= 1 {
				t.Fatalf("burst=%v progress = %v outside [0,1)", burst, p)
			}
		}
	}
}

func TestGame_StartupBurst_DefaultsTrue_AndRunsThreeTicksOnFirstUpdate(t *testing.T) {
	fired := 0
	g := game.NewGame(&engine.MockTickSource{})
	g.OnTick = func() { fired++ }

	got := g.Update(0)

	if !g.StartupBurst {
		t.Fatal("NewGame must default StartupBurst to true")
	}
	if got != 3 || fired != 3 {
		t.Fatalf("first Update ran %d ticks (OnTick fired %d), want 3 immediate startup ticks", got, fired)
	}
}

func TestGame_StartupBurst_RunsOnlyOnce(t *testing.T) {
	g := game.NewGame(&engine.MockTickSource{})
	g.Update(0)

	got := g.Update(0)

	if got != 0 {
		t.Fatalf("second Update ran %d ticks, want 0", got)
	}
}

func TestGame_Update_RunsWithoutAllocations(t *testing.T) {
	g := game.NewGame(&engine.MockTickSource{})
	for range 10 {
		g.Update(16_670_000)
	}

	allocs := testing.AllocsPerRun(100, func() { g.Update(16_670_000) })

	if allocs != 0 {
		t.Fatalf("hot-path Update allocated %v objects per run, want 0", allocs)
	}
}
