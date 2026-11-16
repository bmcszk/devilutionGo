package scene_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/present/scene"
)

// paletteCount is one index/occurrence pair.
type paletteCount struct {
	index uint8
	count int
}

// TestProbeTownBrightness dumps the composed town frame's top palette colours,
// the shade-0 brightness probe for the darkness investigation.
func TestProbeTownBrightness(t *testing.T) {
	// given the retail archives reachable from DG_ASSET_ARCHIVES
	dirs := filepath.SplitList(os.Getenv("DG_ASSET_ARCHIVES"))
	if len(dirs) == 0 {
		t.Skip("no archives")
	}

	// when the town frame composes
	composed, err := scene.ComposeTownFrame(dirs)
	if err != nil {
		t.Fatal(err)
	}

	// then the top palette colours land in the log
	for _, e := range topPaletteCounts(countIndices(composed)) {
		c := composed.State.Variants[0][e.index]
		t.Logf("idx %3d x%-6d RGB(%3d,%3d,%3d)", e.index, e.count, c[0], c[1], c[2])
	}
}

// countIndices tallies every palette index of the composed surface.
func countIndices(composed scene.TownScene) map[uint8]int {
	counts := make(map[uint8]int, 256)
	for _, idx := range composed.Surface.Pixels {
		counts[idx]++
	}

	return counts
}

// topPaletteCounts returns the counts sorted by occurrence, descending, capped
// at the top 12.
func topPaletteCounts(counts map[uint8]int) []paletteCount {
	top := make([]paletteCount, 0, len(counts))
	for index, count := range counts {
		top = append(top, paletteCount{index, count})
	}

	slices.SortFunc(top, func(a, b paletteCount) int { return b.count - a.count })

	if len(top) > 12 {
		top = top[:12]
	}

	return top
}
