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
			// given
			g := game.NewGame(&engine.MockTickSource{})
			g.StartupBurst = false

			// when
			got := g.Update(tt.step)

			// then
			if got != tt.want {
				t.Fatalf("Update(%v) ran %d ticks, want %d", tt.step, got, tt.want)
			}
		})
	}
}

func TestGame_Update_RealignsAfterClampWithoutCatchUp(t *testing.T) {
	// given
	fired := 0
	g := game.NewGame(&engine.MockTickSource{})
	g.StartupBurst = false
	g.OnTick = func() { fired++ }
	if got := g.Update(500 * time.Millisecond); got != 10 {
		t.Fatalf("precondition: 500ms step ran %d ticks, want the 10-tick clamp", got)
	}
	fired = 0

	// when
	for range 60 {
		g.Update(16_670_000)
	}

	// then
	if fired != 20 {
		t.Fatalf("second after 500ms clamp ran %d ticks, want exactly 20 with no catch-up", fired)
	}
}

//nolint:revive // table subtests are intentionally branchy
func TestGame_ProgressToNextGameTick(t *testing.T) {
	t.Run("hand computed at 25ms frames", func(t *testing.T) {
		// given
		g := game.NewGame(&engine.MockTickSource{})
		g.StartupBurst = false
		want := []float64{0.5, 0, 0.5, 0}

		// when
		got := make([]float64, 0, len(want))
		for range len(want) {
			g.Update(25 * time.Millisecond)
			got = append(got, g.ProgressToNextGameTick)
		}

		// then
		for i, w := range want {
			if got[i] != w {
				t.Fatalf("frame %d progress = %v, want %v", i, got[i], w)
			}
		}
	})

	t.Run("always within half-open unit interval", func(t *testing.T) {
		// given
		var got []float64
		for _, burst := range []bool{true, false} {
			g := game.NewGame(&engine.MockTickSource{})
			g.StartupBurst = burst

			// when
			for range 120 {
				g.Update(7_300_000)
				got = append(got, g.ProgressToNextGameTick)
			}
		}

		// then
		for i, p := range got {
			if p < 0 || p >= 1 {
				t.Fatalf("sample %d progress = %v outside [0,1)", i, p)
			}
		}
	})
}

//nolint:revive // two subtests, complexity from t.Run closure nesting
func TestGame_StartupBurst_RunsThreeTicksOnFirstUpdate(t *testing.T) {
	t.Run("default game bursts three ticks", func(t *testing.T) {
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
		if got != 3 || fired != 3 {
			t.Fatalf("first Update ran %d ticks (OnTick fired %d), want 3 immediate startup ticks", got, fired)
		}
	})

	t.Run("burst runs only once", func(t *testing.T) {
		// given
		g := game.NewGame(&engine.MockTickSource{})
		g.Update(0)

		// when
		got := g.Update(0)

		// then
		if got != 0 {
			t.Fatalf("second Update ran %d ticks, want 0", got)
		}
	})
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
