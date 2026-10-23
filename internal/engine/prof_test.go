//nolint:testpackage // internal-state test: pokes unexported identifiers
//nolint:testpackage // internal-state test: pokes unexported identifiers
package engine

import (
	"testing"
)

// Profiling entry point for the pathfinder: captures CPU and memory profiles
// of the worst-case 40x40 search on the budget grid.
func BenchmarkProfileFindPath(b *testing.B) {
	// given
	grid, start, dest := budgetFixture()
	path := make([]int8, MaxPathLengthPlayer)

	// when
	for b.Loop() {
		findPath(alwaysCanStep, walkablePosOk(grid), start, dest, path)
	}
}
