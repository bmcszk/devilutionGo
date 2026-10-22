package engine_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/engine"
)

// Comparative tool for the TestFindPathBudget10kOn40x40 wall-clock gate in budgets_test.go.
func BenchmarkFindPath40x40(b *testing.B) {
	// given
	// when
	grid, start, dest := budgetFixture()
	path := make([]int8, engine.MaxPathLengthPlayer)
	b.ReportAllocs()

	// when
	b.ResetTimer()
	for b.Loop() {
		engine.FindPath(alwaysCanStep, walkablePosOk(grid), start, dest, path)
	}

	// then
	// The framework-reported ns/op and allocs/op are the comparative figures.
}

// Comparative tool for the TestCrawlBudget100k wall-clock gate in budgets_test.go.
func BenchmarkCrawl100k(b *testing.B) {
	// given
	// when
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

	// then
	// The framework-reported ns/op and allocs/op are the comparative figures.
}
