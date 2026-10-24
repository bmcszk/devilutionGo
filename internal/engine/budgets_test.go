package engine_test

import (
	"github.com/bmcszk/devilutionGo/internal/engine"
	"testing"
	"time"
)

func TestFindPathBudget10kOn40x40(t *testing.T) {
	const calls = 10_000
	grid, start, dest := budgetFixture()
	path := make([]int8, engine.MaxPathLengthPlayer)
	if engine.FindPath(alwaysCanStep, walkablePosOk(grid), start, dest, path) == 0 {
		t.Fatal("seeded grid must contain a path for the budget run")
	}
	// Calibration probe: one worst-case search times the box; budget =
	// 2x(10k x probe). Idle floor ≈ 150ms (7.5µs/search, BenchmarkFindPath40x40);
	// a loaded desktop raises probe and budget together, while a 2x-slower
	// implementation still fails on any box. Measured 2026-10-17.
	probeBegin := time.Now()
	engine.FindPath(alwaysCanStep, walkablePosOk(grid), start, dest, path)
	probe := time.Since(probeBegin)

	// when
	begin := time.Now()
	for range calls {
		engine.FindPath(alwaysCanStep, walkablePosOk(grid), start, dest, path)
	}
	elapsed := time.Since(begin)

	// then
	budget := 2 * time.Duration(calls) * probe
	if elapsed > budget {
		t.Fatalf("%d findPath calls took %v, budget %v (probe %v)", calls, elapsed, budget, probe)
	}
}

func TestCrawlBudget100k(t *testing.T) {
	const sweeps = 100_000
	grid, origin := crawlPocketFixture()
	visit := func(d engine.Displacement) (engine.Displacement, bool) {
		p := origin.Plus(engine.Point{X: d.DeltaX, Y: d.DeltaY})
		if grid[p.X][p.Y] {
			return d, true
		}
		return engine.Displacement{}, false
	}
	if _, found := engine.Crawl(0, 8, visit); !found {
		t.Fatal("pocketed grid must yield a walkable ring-4 tile for the budget run")
	}
	// Same self-calibration as the path budget: probe sweep times the box;
	// budget = 2.5x(100k x probe). Idle floor ≈ 30ms (46 visits/sweep,
	// BenchmarkCrawl100k).
	probeBegin := time.Now()
	engine.Crawl(0, 8, visit)
	probe := time.Since(probeBegin)

	// when
	begin := time.Now()
	for range sweeps {
		engine.Crawl(0, 8, visit)
	}
	elapsed := time.Since(begin)

	// then
	budget := time.Duration(2.5 * float64(sweeps) * probe.Seconds() * float64(time.Second))
	if elapsed > budget {
		t.Fatalf("%d crawl sweeps took %v, budget %v (probe %v)", sweeps, elapsed, budget, probe)
	}
}

func budgetFixture() (grid [][]bool, start, dest engine.Point) {
	grid = makeObstacleGrid(42, 40, 20)
	start = engine.Point{X: 1, Y: 1}
	dest = engine.Point{X: 38, Y: 38}
	grid[start.X][start.Y] = true
	grid[dest.X][dest.Y] = true
	return grid, start, dest
}

// crawlPocketFixture clears a 7x7 pocket around (20, 20) on the budget grid so
// each crawl sweep crosses rings 0..3 (46 visits) before the first walkable
// ring-4 tile ends it; radius stays within the 40x40 bounds by construction.
func crawlPocketFixture() (grid [][]bool, origin engine.Point) {
	grid, _, _ = budgetFixture()
	origin = engine.Point{X: 20, Y: 20}
	for dx := -3; dx <= 3; dx++ {
		for dy := -3; dy <= 3; dy++ {
			grid[origin.X+dx][origin.Y+dy] = false
		}
	}
	return grid, origin
}
