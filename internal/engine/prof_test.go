package engine_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/engine"
)

func BenchmarkProfileFindPath(b *testing.B) {
	grid, start, dest := budgetFixture()
	path := make([]int8, engine.MaxPathLengthPlayer)
	for b.Loop() {
		engine.FindPath(alwaysCanStep, walkablePosOk(grid), start, dest, path)
	}
}
