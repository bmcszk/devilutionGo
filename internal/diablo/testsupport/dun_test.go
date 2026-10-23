// Package testsupport_test: devilGo-xcdz slice-1 acceptance tests for the
// dun fixture parser and the golden diff reporter. The corpus requirement is
// every refreshed local test/fixtures/generated/dun/*.dun (never a
// hardcoded list).
//
//nolint:testpackage // internal-state test: pokes unexported identifiers
package testsupport

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	gametestsupport "github.com/bmcszk/devilutionGo/internal/game/testsupport"
)

// committedDunFiles globs every committed CSV dungeon fixture.
func committedDunFiles(t *testing.T) []string {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(dunDir, "*.csv"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(paths) == 0 {
		t.Fatalf("committed dun fixtures absent under %s", dunDir)
	}
	return paths
}

func TestLoadDunFixtureParsesEveryCommittedFile(t *testing.T) {
	// given
	checkCorpusShape(t)
	// when
	for seed := range committedSeedSet(t) {
		t.Run(fmt.Sprintf("seed%d", seed), func(t *testing.T) {
			fixture := loadDunFixture(t, seed)
			for base, pair := range fixture.Pairs {
				// then
				checkPair(t, base, pair)
			}
		})
	}
}

// checkCorpusShape pins the corpus topology: 16 variants x 8 seeds.
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

// committedSeedSet groups the committed corpus by parsed filename seed.
func committedSeedSet(t *testing.T) map[uint32][]string {
	t.Helper()
	bySeed := map[uint32][]string{}
	for _, p := range committedDunFiles(t) {
		seed, _ := seedFromFilename(p)
		bySeed[seed] = append(bySeed[seed], p)
	}
	return bySeed
}

// hasNonZeroCell reports whether any cell of the layer is nonzero.
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

// checkPair validates one parsed pair: dims plausibility (AC 40..112),
// layer-shape agreement, non-empty tile data and a non-zero meta.
func checkPair(t *testing.T, base string, pair *dunPair) {
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
	if len(pair.TransparencyLayer) != pair.Width*2 || len(pair.TransparencyLayer[0]) != pair.Height*2 {
		t.Fatalf("%s: transparency dims %dx%d, want %dx%d",
			base, len(pair.TransparencyLayer), len(pair.TransparencyLayer[0]),
			pair.Width*2, pair.Height*2)
	}
	nonZero := hasNonZeroCell(pair.TileLayer)
	if !nonZero {
		t.Fatalf("%s: tile layer all zero", base)
	}
	if pair.Meta.ViewX == 0 && pair.Meta.ViewY == 0 && pair.Meta.LevelType == 0 {
		t.Fatalf("%s: meta all zero", base)
	}
}

func TestCompareGrids(t *testing.T) {
	// when equal grids are compared, no diffs are reported
	same := [][]uint8{{1, 2}, {3, 4}}
	if diffs := gametestsupport.CompareGrids(same, same); len(diffs) != 0 {
		t.Fatalf("equal grids: diffs = %v, want none", diffs)
	}

	got := [][]uint8{{1, 9}, {3, 4}}
	want := [][]uint8{{1, 2}, {8, 4}}
	// when two cells differ, each differing cell is reported in row-major order
	diffs := gametestsupport.CompareGrids(got, want)
	// then
	if len(diffs) != 2 {
		t.Fatalf("diffs = %v, want 2 entries", diffs)
	}
	if diffs[0] != "[0][1]: got 9, want 2" {
		t.Errorf("diffs[0] = %q", diffs[0])
	}
	if diffs[1] != "[1][0]: got 3, want 8" {
		t.Errorf("diffs[1] = %q", diffs[1])
	}

	if diffs := gametestsupport.CompareGrids(got, [][]uint8{{1}}); len(diffs) == 0 {
		t.Error("dim mismatch must be reported")
	}
}

func TestGoldenRoundTripL1Main2588(t *testing.T) {
	// given
	original, err := os.ReadFile(filepath.Join(dunDir, "l1-main-2588.csv"))
	if err != nil {
		t.Fatalf("read committed dun fixture: %v", err)
	}
	// when
	fixture := loadDunFixture(t, 2588)
	pair, ok := fixture.Pairs["l1-main"]
	if !ok {
		t.Fatal("l1-main pair missing")
	}
	encoded := encodeDunCSV(pair)
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
