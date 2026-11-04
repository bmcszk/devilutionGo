package levels_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/levels"
)

// TestGeneratorTableSizes pins the sizes of the three hard-coded tables the
// decoration passes index by tile id (levels/drlg_l1.cpp:183-293): BaseTypes and
// TileDecorations are [207]Tile in upstream, so a truncated row would silently
// mis-map tile ids; ShadowPatterns is a 37-entry list.
//
// The table data itself is pinned end to end by TestL1CathedralParity: a wrong
// constant shifts the tiles it rewrites, and the committed golden then fails.
func TestGeneratorTableSizes(t *testing.T) {
	t.Parallel()

	// given: the three ported generator tables and their upstream row counts.
	tests := []struct {
		name string
		get  func(base, decorations, shadows int) int
		want int
	}{
		{name: "BaseTypes", get: func(base, _, _ int) int { return base }, want: 207},
		{name: "TileDecorations", get: func(_, decorations, _ int) int { return decorations }, want: 207},
		{name: "ShadowPatterns", get: func(_, _, shadows int) int { return shadows }, want: 37},
	}

	// when: each table's length is read through the test seam.
	base, decorations, shadows := levels.TileTableSizes()

	// then: it is the upstream row count.
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := test.get(base, decorations, shadows); got != test.want {
				t.Errorf("%s table has %d entries, want %d", test.name, got, test.want)
			}
		})
	}
}
