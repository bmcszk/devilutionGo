package levels_test

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/levels"
)

// lightDir is the committed dLight corpus (internal/testdata/drlg/light).
const lightDir = drlgDir + "/light"

// lightVariants is the matrix the dLight corpus was produced for: the two level
// types whose static light pass the port owns — DTYPE_NEST levels 17-20
// (PlaceHiveLights, drlg_l3.cpp:2169-2178) and DTYPE_CRYPT levels 21-24
// (PlaceCryptLights, crypt.cpp:817-838) — both entries.
var lightVariants = []lightVariant{
	{17, levels.EntryMain, "main"},
	{17, levels.EntryPrev, "prev"},
	{18, levels.EntryMain, "main"},
	{18, levels.EntryPrev, "prev"},
	{19, levels.EntryMain, "main"},
	{19, levels.EntryPrev, "prev"},
	{20, levels.EntryMain, "main"},
	{20, levels.EntryPrev, "prev"},
	{21, levels.EntryMain, "main"},
	{21, levels.EntryPrev, "prev"},
	{22, levels.EntryMain, "main"},
	{22, levels.EntryPrev, "prev"},
	{23, levels.EntryMain, "main"},
	{23, levels.EntryPrev, "prev"},
	{24, levels.EntryMain, "main"},
	{24, levels.EntryPrev, "prev"},
}

// lightVariant is one level/entry pair of the dLight matrix.
type lightVariant struct {
	level int
	entry levels.Entry
	name  string
}

// lightDiff is one differing dLight cell.
type lightDiff struct {
	x, y int
	got  uint8
	want uint8
}

// TestLightParity is the lighting acceptance gate: for every committed dLight
// fixture the Go port must build the same level and stamp the same light grid,
// shade for shade.
func TestLightParity(t *testing.T) {
	// given: the two mega-tile tables the matrix's level families install.
	megaTiles := map[string][]levels.MegaTile{
		"l5": loadMegaTiles(t, "l5"),
		"l6": loadMegaTiles(t, "l6"),
	}

	// when: every variant of every pinned seed is generated, expanded and lit.
	covered := runLightMatrix(t, megaTiles)

	// then: the corpus holds exactly one fixture per variant and pinned seed.
	if want := len(lightVariants) * len(committedSeeds); covered != want {
		t.Fatalf("light corpus covers %d variant/seed pairs, want %d: run `dvg_probe light "+
			"<out> tools/probe/seeds.txt` then `go run ./tools/goldens light <out>/light "+
			"internal/testdata/drlg/light` (internal/testdata/drlg/README.md)", covered, want)
	}
}

// runLightMatrix runs one sub-test per variant and seed and reports how many
// had a committed fixture.
func runLightMatrix(t *testing.T, megaTiles map[string][]levels.MegaTile) int {
	t.Helper()

	covered := 0
	for _, seed := range committedSeeds {
		paths := lightGoldenPaths(t, seed)
		for _, variant := range lightVariants {
			if runLightVariant(t, variant, seed, paths, megaTiles) {
				covered++
			}
		}
	}

	return covered
}

// runLightVariant runs one variant/seed sub-test and reports whether its
// fixture was found.
func runLightVariant(t *testing.T, variant lightVariant, seed uint32, paths map[string]string,
	megaTiles map[string][]levels.MegaTile,
) bool {
	t.Helper()

	name := lightFixtureName(variant)
	path := paths[name]
	found := path != ""

	t.Run(fmt.Sprintf("%s-seed-%d", name, seed), func(t *testing.T) {
		// given: this variant's committed fixture.
		if !found {
			t.Fatalf("no committed fixture for %s seed %d (expected %s/...csv)", name, seed, lightDir)
		}

		// when: this variant is generated, expanded and lit.
		got := lightVariantGrid(t, variant, seed, megaTiles)

		// then: dLight must match the fixture, cell for cell.
		checkLightCells(t, got, loadLightGolden(t, path), path)
	})

	return found
}

// lightVariantGrid builds one variant's level the way CreateDungeon does and
// returns its dLight grid.
func lightVariantGrid(t *testing.T, variant lightVariant, seed uint32,
	megaTiles map[string][]levels.MegaTile,
) *[levels.MAXDUNX][levels.MAXDUNY]uint8 {
	t.Helper()

	dungeon := levels.NewDungeon(lightLevelType(variant.level), variant.level)
	dungeon.Rnd.Seed(seed)
	dungeon.InitLight()

	if err := lightVariantDungeon(t, dungeon, variant, megaTiles); err != nil {
		t.Fatalf("generate %s seed %d: %v", lightFixtureName(variant), seed, err)
	}

	return &dungeon.Light
}

// lightVariantDungeon runs the generator of the variant's level family, then
// the two create-time steps CreateL5Dungeon and CreateL3Dungeon add after
// GenerateLevel: Pass3 (:1305, :2198) and the static light pass.
func lightVariantDungeon(t *testing.T, dungeon *levels.Dungeon, variant lightVariant,
	megaTiles map[string][]levels.MegaTile,
) error {
	t.Helper()

	if err := runLightGenerator(t, dungeon, variant); err != nil {
		return err
	}

	pieces := levels.ExpandPieces(dungeon, megaTiles[lightMegaTileFamily(dungeon.LevelType)])
	levels.ApplyLights(dungeon, pieces)

	return nil
}

// runLightGenerator dispatches the variant to its family's generator. The nest
// levels use GenerateNest (drlg_l3.cpp:2192-2200); the crypt levels share
// CreateL5Dungeon's cathedral-family generator with the crypt branch
// (drlg_l1.cpp:1299-1312).
func runLightGenerator(t *testing.T, dungeon *levels.Dungeon, variant lightVariant) error {
	t.Helper()

	if dungeon.LevelType == levels.Nest {
		if err := levels.GenerateNest(dungeon, levels.Options{Entry: variant.entry}); err != nil {
			return fmt.Errorf("generate nest level %d: %w", variant.level, err)
		}

		return nil
	}

	setPieces := familySetPieces(t, dungeon.LevelType)
	if err := generateFamily(dungeon, dungeon.LevelType, variant.entry, setPieces); err != nil {
		return fmt.Errorf("generate crypt level %d: %w", variant.level, err)
	}

	return nil
}

// lightMegaTileFamily is the lN.til table LoadLvlGFX installs for a level type
// (diablo.cpp:1360-1400).
func lightMegaTileFamily(levelType levels.LevelType) string {
	if levelType == levels.Nest {
		return "l6"
	}

	return "l5"
}

// lightLevelType is the level type the level number belongs to
// (levels/gendung.cpp:326-344), restricted to the families this matrix covers.
func lightLevelType(level int) levels.LevelType {
	if level <= 20 {
		return levels.Nest
	}

	return levels.Crypt
}

// lightFixtureName is the variant's fixture base name without its seed
// ("l17-light-main").
func lightFixtureName(variant lightVariant) string {
	return fmt.Sprintf("l%d-light-%s", variant.level, variant.name)
}

// lightGoldenPaths indexes one seed's committed dLight fixtures by their base
// name without the seed ("l17-light-main").
func lightGoldenPaths(t *testing.T, seed uint32) map[string]string {
	t.Helper()

	paths, err := filepath.Glob(filepath.Join(lightDir, fmt.Sprintf("l*-light-*-%d.csv", seed)))
	if err != nil {
		t.Fatalf("glob light corpus for seed %d: %v", seed, err)
	}

	index := make(map[string]string, len(paths))
	for _, path := range paths {
		base := strings.TrimSuffix(filepath.Base(path), ".csv")
		index[strings.TrimSuffix(base, fmt.Sprintf("-%d", seed))] = path
	}

	return index
}

// loadLightGolden parses one fixture: "112,112", then 112*112 dLight shades,
// x-major.
func loadLightGolden(t *testing.T, path string) *[levels.MAXDUNX][levels.MAXDUNY]uint8 {
	t.Helper()

	records := readCSVRecords(t, path)
	if len(records) != 2 {
		t.Fatalf("%s: %d records, want 2", path, len(records))
	}

	checkPieceHeader(t, path, records[0])

	var grid [levels.MAXDUNX][levels.MAXDUNY]uint8
	checkPieceRecordLen(t, path, len(records[1]))
	for index, field := range records[1] {
		value, err := strconv.ParseUint(field, 10, 8)
		if err != nil {
			t.Fatalf("%s: dLight value %d %q: %v", path, index, field, err)
		}
		grid[index/levels.MAXDUNY][index%levels.MAXDUNY] = uint8(value)
	}

	return &grid
}

// lightDiffs lists every cell where the two grids disagree.
func lightDiffs(got, want *[levels.MAXDUNX][levels.MAXDUNY]uint8) []lightDiff {
	diffs := make([]lightDiff, 0)
	for x := range levels.MAXDUNX {
		for y := range levels.MAXDUNY {
			if got[x][y] != want[x][y] {
				diffs = append(diffs, lightDiff{x: x, y: y, got: got[x][y], want: want[x][y]})
			}
		}
	}

	return diffs
}

// checkLightCells diffs the whole grid, reporting the first differing cells.
func checkLightCells(t *testing.T, got, want *[levels.MAXDUNX][levels.MAXDUNY]uint8, path string) {
	t.Helper()

	diffs := lightDiffs(got, want)
	for _, diff := range diffs[:min(len(diffs), 10)] {
		t.Errorf("%s: dLight[%d][%d] = %d, want %d", path, diff.x, diff.y, diff.got, diff.want)
	}

	if len(diffs) > 0 {
		t.Fatalf("%s: %d of %d dLight cells differ", path, len(diffs), levels.MAXDUNX*levels.MAXDUNY)
	}
}
