// Package testsupport_test: devilGo-xcdz slice-1 acceptance tests for the
// dun fixture parser and the golden diff reporter. The corpus requirement is
// every refreshed local test/fixtures/generated/dun/*.dun (never a
// hardcoded list).

package testsupport_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/testsupport"
	gametestsupport "github.com/bmcszk/devilutionGo/internal/game/testsupport"
)

func TestLoadDunFixtureParsesEveryCommittedFile(t *testing.T) {
	// given the committed dun corpus, whose shape is pinned before loading
	checkCorpusShape(t)
	seeds := committedSeedSet(t)
	// when one fixture is loaded per committed seed
	fixtures := loadDunFixtures(t, seeds)
	// then every variant pair carries valid layers and metadata
	for seed, fixture := range fixtures {
		t.Run(fmt.Sprintf("seed%d", seed), func(t *testing.T) {
			for base, pair := range fixture.Pairs {
				checkPair(t, base, pair)
			}
		})
	}
}

func TestCompareGrids(t *testing.T) {
	// given equal grids, two cells apart, and a dimension mismatch
	same := [][]uint8{{1, 2}, {3, 4}}
	got := [][]uint8{{1, 9}, {3, 4}}
	want := [][]uint8{{1, 2}, {8, 4}}
	// when each pair is compared
	equalDiffs := gametestsupport.CompareGrids(same, same)
	diffs := gametestsupport.CompareGrids(got, want)
	dimDiffs := gametestsupport.CompareGrids(got, [][]uint8{{1}})
	// then no diffs for equal grids and one entry per differing cell
	if len(equalDiffs) != 0 {
		t.Fatalf("equal grids: diffs = %v, want none", equalDiffs)
	}

	if len(diffs) != 2 {
		t.Fatalf("diffs = %v, want 2 entries", diffs)
	}

	if diffs[0] != "[0][1]: got 9, want 2" {
		t.Errorf("diffs[0] = %q", diffs[0])
	}

	if diffs[1] != "[1][0]: got 3, want 8" {
		t.Errorf("diffs[1] = %q", diffs[1])
	}

	if len(dimDiffs) == 0 {
		t.Error("dim mismatch must be reported")
	}
}

func TestGoldenRoundTripL1Main2588(t *testing.T) {
	// given
	original, err := os.ReadFile(filepath.Join(testsupport.DunDir, "l1-main-2588.csv"))
	if err != nil {
		t.Fatalf("read committed dun fixture: %v", err)
	}
	// when
	fixture := testsupport.LoadDunFixture(t, 2588)
	pair := goldenPair(t, fixture, "l1-main")
	encoded := testsupport.EncodeDunCSV(pair)
	// then
	if !bytes.Equal(encoded, original) {
		n := min(len(original), len(encoded))
		for i := range n {
			if encoded[i] != original[i] {
				t.Fatalf("first byte diff at %d: encoded %d, original %d", i, encoded[i], original[i])
			}
		}

		t.Fatalf("length mismatch: encoded %d, original %d", len(encoded), len(original))
	}
}

// loadDunFixtures loads one fixture per committed seed.
func loadDunFixtures(t *testing.T, seeds map[uint32][]string) map[uint32]*testsupport.DunFixture {
	t.Helper()

	fixtures := make(map[uint32]*testsupport.DunFixture, len(seeds))
	for seed := range seeds {
		fixtures[seed] = testsupport.LoadDunFixture(t, seed)
	}

	return fixtures
}

func checkCorpusShape(t *testing.T) {
	t.Helper()
	paths := committedDunFiles(t)

	bySeed := committedSeedSet(t)
	if len(bySeed) != 8 {
		t.Fatalf("distinct seeds = %d, want 8", len(bySeed))
	}

	if len(paths) != 128 {
		t.Fatalf("generated pairs = %d, want 128 (16 variants x 8 seeds)", len(paths))
	}
}

func committedSeedSet(t *testing.T) map[uint32][]string {
	t.Helper()

	bySeed := map[uint32][]string{}

	for _, p := range committedDunFiles(t) {
		seed, _ := testsupport.SeedFromFilename(p)
		bySeed[seed] = append(bySeed[seed], p)
	}

	return bySeed
}

func hasNonZeroCell(layer [][]uint8) bool {
	for _, row := range layer {
		for _, cell := range row {
			if cell != 0 {
				return true
			}
		}
	}

	return false
}

func checkPair(t *testing.T, base string, pair *testsupport.DunPair) {
	t.Helper()
	checkPairDims(t, base, pair)
	checkTransparencyDims(t, base, pair)

	if !hasNonZeroCell(pair.TileLayer) {
		t.Fatalf("%s: tile layer all zero", base)
	}

	if pair.Meta.ViewX == 0 && pair.Meta.ViewY == 0 && pair.Meta.LevelType == 0 {
		t.Fatalf("%s: meta all zero", base)
	}
}

// checkPairDims pins the tile layer shape and the [40,112] dimension bounds.
func checkPairDims(t *testing.T, base string, pair *testsupport.DunPair) {
	t.Helper()

	if got := len(pair.TileLayer); got != pair.Height {
		t.Fatalf("%s: tile rows = %d, want %d", base, got, pair.Height)
	}

	if got := len(pair.TileLayer[0]); got != pair.Width {
		t.Fatalf("%s: tile cols = %d, want %d", base, got, pair.Width)
	}

	if pair.Width < 40 || pair.Width > 112 || pair.Height < 40 || pair.Height > 112 {
		t.Fatalf("%s: dims %dx%d outside [40,112]", base, pair.Width, pair.Height)
	}
}

// checkTransparencyDims pins the doubled transparency layer shape.
func checkTransparencyDims(t *testing.T, base string, pair *testsupport.DunPair) {
	t.Helper()

	if len(pair.TransparencyLayer) != pair.Width*2 || len(pair.TransparencyLayer[0]) != pair.Height*2 {
		t.Fatalf("%s: transparency dims %dx%d, want %dx%d",
			base, len(pair.TransparencyLayer), len(pair.TransparencyLayer[0]),
			pair.Width*2, pair.Height*2)
	}
}

// goldenPair extracts one named pair from a fixture or fails the test.
func goldenPair(t *testing.T, fixture *testsupport.DunFixture, name string) *testsupport.DunPair {
	t.Helper()

	pair, ok := fixture.Pairs[name]
	if !ok {
		t.Fatalf("%s pair missing", name)
	}

	return pair
}

func committedDunFiles(t *testing.T) []string {
	t.Helper()

	paths, err := filepath.Glob(filepath.Join(testsupport.DunDir, "*.csv"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}

	if len(paths) == 0 {
		t.Fatalf("committed dun fixtures absent under %s", testsupport.DunDir)
	}

	return paths
}
