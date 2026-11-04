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

const (
	artifactWidth  = 2
	artifactHeight = 3
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

func TestParseDunArtifactMirrorsProbeLayout(t *testing.T) {
	// given a 2x3 artifact whose layers are both x-major, plus a
	// view/leveltype meta sidecar
	dun, meta := xMajorArtifact()
	// when the pair is parsed
	pair, err := testsupport.ParseDunArtifact(dun, meta)
	// then the header, both layers and the meta are laid out as [y][x]
	if err != nil {
		t.Fatalf("ParseDunArtifact: %v", err)
	}

	if pair.Width != artifactWidth || pair.Height != artifactHeight {
		t.Fatalf("dims = %dx%d, want %dx%d", pair.Width, pair.Height, artifactWidth, artifactHeight)
	}

	checkArtifactGrid(t, "tile", pair.TileLayer, artifactWidth, artifactHeight, func(x, y int) uint8 {
		return uint8(10 + x*artifactHeight + y)
	})
	checkArtifactGrid(t, "trans", pair.TransparencyLayer, artifactWidth*2, artifactHeight*2, func(x, y int) uint8 {
		return uint8(40 + x*artifactHeight*2 + y)
	})

	if want := (testsupport.DunMeta{ViewX: 300, ViewY: 7, LevelType: 2}); pair.Meta != want {
		t.Errorf("meta = %+v, want %+v", pair.Meta, want)
	}
}

func TestParseDunArtifactRoundTripsThroughCorpusEncoder(t *testing.T) {
	// given the smallest artifact the probe can emit: 1x1 tiles and a 2x2
	// transparency window, both dumped in the probe's own order
	dun := []byte{1, 0, 1, 0, 22, 9, 8, 7, 6}
	// when it is parsed and re-encoded
	pair, err := testsupport.ParseDunArtifact(dun, []byte{3, 0, 4, 0, 1})
	if err != nil {
		t.Fatalf("ParseDunArtifact: %v", err)
	}

	encoded := testsupport.EncodeDunCSV(pair)
	// then the CSV carries the four records the corpus format defines
	want := "1,1\n22\n9,8,7,6\n3,4,1\n"
	if string(encoded) != want {
		t.Fatalf("encoded = %q, want %q", encoded, want)
	}
}

func TestParseDunArtifactRejectsShortTileLayer(t *testing.T) {
	// given a header promising 2x3 tiles but only one tile of payload
	dun := []byte{2, 0, 3, 0, 22}
	// when
	_, err := testsupport.ParseDunArtifact(dun, make([]byte, 5))
	// then the truncation is reported
	if err == nil {
		t.Fatal("truncated tile layer accepted")
	}
}

func TestParseDunArtifactRejectsShortTransparencyWindow(t *testing.T) {
	// given a complete tile layer with a truncated transparency window
	dun := make([]byte, 4+2*2+2)
	// when
	_, err := testsupport.ParseDunArtifact(dun, make([]byte, 5))
	// then the truncation is reported
	if err == nil {
		t.Fatal("truncated transparency window accepted")
	}
}

func TestParseDunArtifactRejectsMalformedMeta(t *testing.T) {
	// given a valid dun payload and a 4-byte sidecar
	dun := []byte{1, 0, 1, 0, 22, 9}
	// when
	_, err := testsupport.ParseDunArtifact(dun, make([]byte, 4))
	// then the fixed 5-byte meta size is enforced
	if err == nil {
		t.Fatal("4-byte meta accepted")
	}
}

// xMajorArtifact builds a synthetic `.dun` artifact pair: a 2x3 tile grid and a
// 4x6 transparency window, each holding its x-major cell index offset by a
// layer-specific bias, plus a view/leveltype meta sidecar.
func xMajorArtifact() (dun, meta []byte) {
	dun = make([]byte, 0, 4+artifactWidth*artifactHeight+4*artifactWidth*artifactHeight)
	dun = append(dun, byte(artifactWidth), 0, byte(artifactHeight), 0)
	for col := range artifactWidth {
		for row := range artifactHeight {
			dun = append(dun, uint8(10+col*artifactHeight+row))
		}
	}

	for col := range artifactWidth * 2 {
		for row := range artifactHeight * 2 {
			dun = append(dun, uint8(40+col*artifactHeight*2+row))
		}
	}

	return dun, []byte{0x2c, 0x01, 0x07, 0x00, 0x02}
}

// checkArtifactGrid compares one parsed grid against the value the artifact
// must hold at (x, y).
func checkArtifactGrid(t *testing.T, name string, grid [][]uint8, width, height int, want func(x, y int) uint8) {
	t.Helper()

	if len(grid) != height || len(grid[0]) != width {
		t.Fatalf("%s grid = %dx%d, want %dx%d", name, len(grid[0]), len(grid), width, height)
	}

	for y := range height {
		checkArtifactRow(t, name, grid[y], y, width, want)
	}
}

// checkArtifactRow compares one parsed row against the artifact values.
func checkArtifactRow(t *testing.T, name string, row []uint8, y, width int, want func(x, y int) uint8) {
	t.Helper()

	for x := range width {
		if got, expected := row[x], want(x, y); got != expected {
			t.Errorf("%s[%d][%d] = %d, want %d", name, y, x, got, expected)
		}
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
