package levels_test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/levels"
)

// hiveVariants is the nest matrix the hive dPiece corpus was produced for:
// DTYPE_NEST levels 17-20 (gendung.cpp:335), both entries, expanded through the
// hive's own nlevels/l6data/l6.til (166 mega tiles).
var hiveVariants = []hiveVariant{
	{17, levels.EntryMain, "main"},
	{17, levels.EntryPrev, "prev"},
	{18, levels.EntryMain, "main"},
	{18, levels.EntryPrev, "prev"},
	{19, levels.EntryMain, "main"},
	{19, levels.EntryPrev, "prev"},
	{20, levels.EntryMain, "main"},
	{20, levels.EntryPrev, "prev"},
}

// hiveVariant is one level/entry pair of the nest matrix.
type hiveVariant struct {
	level int
	entry levels.Entry
	name  string
}

// TestHiveParity is the hive port's parity gate: it drives GenerateNest (the
// CreateL3Dungeon nest branch, drlg_l3.cpp:2192-2200) and compares the expanded
// dPiece grid against the 64 committed nest goldens.
//
// Green since the port landed in hive.go: GenerateNest runs the caves layout
// with the nest stairs, slime pool and decoration branches, and Pass3 expands
// the layer through the shared ExpandPieces path.
func TestHiveParity(t *testing.T) {
	// given: the hive mega-tile table and every seed's committed nest goldens.
	megaTiles := loadMegaTiles(t, "l6")

	for _, seed := range committedSeeds {
		goldens := loadPieceGoldens(t, seed)

		// when: each nest variant runs GenerateNest and Pass3 expands the layer.
		for _, variant := range hiveVariants {
			name := hiveFixtureName(variant)
			// then: dPiece must match the fixture, cell for cell.
			t.Run(fmt.Sprintf("%s-seed-%d", name, seed), func(t *testing.T) {
				checkHiveVariant(t, variant, seed, goldens[name], megaTiles)
			})
		}
	}
}

// TestHiveCorpus pins the corpus itself, so a truncated or renamed nest golden
// fails here rather than silently shrinking the matrix above.
func TestHiveCorpus(t *testing.T) {
	// when: the committed nest dPiece fixtures are globbed.
	names := hiveGoldenNames(t)

	// then: the corpus holds one fixture per variant and seed.
	want := len(hiveVariants) * len(committedSeeds)
	if got := len(names); got != want {
		t.Fatalf("nest dPiece corpus holds %d fixtures, want %d", got, want)
	}

	// and: every matrix entry has its own fixture.
	for _, variant := range hiveVariants {
		for _, seed := range committedSeeds {
			key := fmt.Sprintf("l%d-dpiece-%s-%d", variant.level, variant.name, seed)
			if _, ok := names[key]; !ok {
				t.Errorf("missing fixture %s.csv", key)
			}
		}
	}
}

// checkHiveVariant runs one nest variant and compares it with its golden.
func checkHiveVariant(t *testing.T, variant hiveVariant, seed uint32, golden *pieceGolden,
	megaTiles []levels.MegaTile,
) {
	t.Helper()

	// given: the committed fixture of this level/entry/seed variant.
	if golden == nil {
		t.Fatalf("missing committed fixture for %s seed %d", hiveFixtureName(variant), seed)
	}

	dungeon := levels.NewDungeon(levels.Nest, variant.level)
	dungeon.Rnd.Seed(seed)

	// when: the hive generator runs and Pass3 expands the layer.
	if err := levels.GenerateNest(dungeon, levels.Options{Entry: variant.entry}); err != nil {
		t.Fatalf("generate nest %s: %v", hiveFixtureName(variant), err)
	}

	// then: dPiece matches the fixture, cell for cell.
	comparePieces(t, levels.ExpandPieces(dungeon, megaTiles), golden)
}

// hiveFixtureName is the variant's golden name ("l17-main") and test label.
func hiveFixtureName(variant hiveVariant) string {
	return fmt.Sprintf("l%d-%s", variant.level, variant.name)
}

// hiveGoldenNames indexes the committed nest dPiece fixtures by their base name
// ("l17-dpiece-main-2588").
func hiveGoldenNames(t *testing.T) map[string]string {
	t.Helper()

	names := make(map[string]string, len(hiveVariants)*len(committedSeeds))
	for _, path := range pieceGoldenPaths(t) {
		base := strings.TrimSuffix(filepath.Base(path), ".csv")
		if isHiveFixture(base) {
			names[base] = path
		}
	}

	return names
}

// isHiveFixture reports whether a fixture base name belongs to a nest level.
func isHiveFixture(base string) bool {
	for _, prefix := range []string{"l17-dpiece-", "l18-dpiece-", "l19-dpiece-", "l20-dpiece-"} {
		if strings.HasPrefix(base, prefix) {
			return true
		}
	}

	return false
}
