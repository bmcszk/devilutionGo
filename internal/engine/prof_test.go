package engine_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/engine"
)

// Profiling entry point for the pathfinder: captures CPU and memory profiles
// of the worst-case 40x40 search on the budget grid.
func BenchmarkProfileFindPath(b *testing.B) {
	// given
	// when
	grid, start, dest := budgetFixture()
	path := make([]int8, engine.MaxPathLengthPlayer)

	// when
	for b.Loop() {
		engine.FindPath(alwaysCanStep, walkablePosOk(grid), start, dest, path)
	}

	// then
	// The framework-reported ns/op is the profiling figure.
}
