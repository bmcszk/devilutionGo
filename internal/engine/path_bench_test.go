package engine_test

import (
	"github.com/bmcszk/devilutionGo/internal/engine"
	"testing"
)

// Comparative tool for the TestFindPathBudget10kOn40x40 wall-clock gate in budgets_test.go.
func BenchmarkFindPath40x40(b *testing.B) {
	// given
	grid, start, dest := budgetFixture()
	path := make([]int8, engine.MaxPathLengthPlayer)
	b.ReportAllocs()

	// when
	b.ResetTimer()
	for b.Loop() {
		engine.FindPath(alwaysCanStep, walkablePosOk(grid), start, dest, path)
	}
}

// Comparative tool for the TestCrawlBudget100k wall-clock gate in budgets_test.go.
func BenchmarkCrawl100k(b *testing.B) {
	// given
	grid, origin := crawlPocketFixture()
	visit := func(d engine.Displacement) (engine.Displacement, bool) {
		p := origin.Plus(engine.Point{X: d.DeltaX, Y: d.DeltaY})
		if grid[p.X][p.Y] {
			return d, true
		}
		return engine.Displacement{}, false
	}
	b.ReportAllocs()

	// when
	b.ResetTimer()
	for b.Loop() {
		engine.Crawl(0, 8, visit)
	}
}
