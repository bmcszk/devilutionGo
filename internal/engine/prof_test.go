package engine_test

import (
	"github.com/bmcszk/devilutionGo/internal/engine"
	"testing"
)

// Profiling entry point for the pathfinder: captures CPU and memory profiles
// of the worst-case 40x40 search on the budget grid.
func BenchmarkProfileFindPath(b *testing.B) {
	// given
	grid, start, dest := budgetFixture()
	path := make([]int8, engine.MaxPathLengthPlayer)

	// when
	for b.Loop() {
		engine.FindPath(alwaysCanStep, walkablePosOk(grid), start, dest, path)
	}
}
