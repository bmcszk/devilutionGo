package engine_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/engine"
)

// BenchmarkFindPath40x40 (path.cpp:184-283) — 10k-iteration budget companion;
// the wall-clock gate lives in budgets_test.go, this is the comparative tool.
func BenchmarkFindPath40x40(b *testing.B) {
	grid, start, dest := budgetFixture()
	path := make([]int8, engine.MaxPathLengthPlayer)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		engine.FindPath(alwaysCanStep, walkablePosOk(grid), start, dest, path)
	}
}

// BenchmarkCrawl100k (crawl.cpp:16-49) — 100k ring sweeps on the pocketed
// budget grid; wall-clock gate lives in budgets_test.go.
func BenchmarkCrawl100k(b *testing.B) {
	grid, origin := crawlPocketFixture()
	visit := func(d engine.Displacement) (engine.Displacement, bool) {
		p := origin.Plus(engine.Point{X: d.DeltaX, Y: d.DeltaY})
		if grid[p.X][p.Y] {
			return d, true
		}
		return engine.Displacement{}, false
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		engine.Crawl(0, 8, visit)
	}
}
