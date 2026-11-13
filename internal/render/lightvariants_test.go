package render_test

import (
	"encoding/hex"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/assets"
	"github.com/bmcszk/devilutionGo/internal/light"
	"github.com/bmcszk/devilutionGo/internal/render"
)

// cycleStep is one row of the committed cycle after-state fixture.
type cycleStep struct {
	name     string
	palette  []byte
	variants string
	lut      string
}

// TestVariantsMatchLightTableMapping pins the variant rows to the mapping they
// claim: variant[shade][index] = palette[lightTables[shade][index]]
// (docs/research/dg-graphics.md G4; light_render.hpp:30-33).
func TestVariantsMatchLightTableMapping(t *testing.T) {
	// given
	pal := loadPalette(t, "l1_1.pal.hex")
	tables := light.MakeLightTables(light.DTYPE_CATHEDRAL)
	// when
	state := render.NewPaletteState(pal, tables, assets.BlendingFor(0))

	// then
	for shade := range state.Variants {
		for index := range state.Variants[shade] {
			want := render.RGB(pal[tables.Tables[shade][index]])
			if got := state.Variants[shade][index]; got != want {
				t.Fatalf("variant[%d][%d] = %v, want %v (index %d)", shade, index, got, want,
					tables.Tables[shade][index])
			}
		}
	}
}

// TestVariantsBrightestShadeKeepsThePalette pins the shade-0 law: the fully lit table
// maps every index to itself except the white entry, which MakeLightTable remaps to
// black (light_tables.cpp:19-48).
func TestVariantsBrightestShadeKeepsThePalette(t *testing.T) {
	// given
	const white = 255

	pal := loadPalette(t, "l1_1.pal.hex")
	tables := light.MakeLightTables(light.DTYPE_CATHEDRAL)
	// when
	state := render.NewPaletteState(pal, tables, assets.BlendingFor(0))

	// then
	for index := range pal {
		want := render.RGB(pal[index])
		if index == white {
			want = render.RGB{}
		}

		if got := state.Variants[0][index]; got != want {
			t.Fatalf("variant[0][%d] = %v, want %v", index, got, want)
		}
	}
}

// TestVariantsGolden pins all 16 rows of 256 colours against a dump of the same
// mapping computed by the pinned C++ code.
func TestVariantsGolden(t *testing.T) {
	// given
	pal := loadPalette(t, "diablo.pal.hex")
	tables := light.MakeLightTables(light.DTYPE_CATHEDRAL)
	want := readFlatBytes(t, filepath.Join(fixtureDir, "lut", "variants_diablo.txt"), 16*256*3)

	// when
	state := render.NewPaletteState(pal, tables, assets.BlendingFor(0))

	// then
	expectBytesEqual(t, flattenVariants(state.Variants), want)
}

// TestCyclePaletteGolden pins the cycle after-state: the rotated palette bytes, the
// light variants re-derived from it and the rotated blend LUT (palette.cpp:52-90).
func TestCyclePaletteGolden(t *testing.T) {
	pal := loadPalette(t, "diablo.pal.hex")
	tables := light.MakeLightTables(light.DTYPE_CATHEDRAL)

	for _, step := range readCycleSteps(t) {
		t.Run(step.name, func(t *testing.T) {
			// when the state reached by replaying every step up to this one is rebuilt
			state := replaySteps(t, pal, tables, step.name)

			// then
			expectBytesEqual(t, flattenPalette(state.Palette), step.palette)
			expectHash(t, "lut", hashHex(flattenLUT(state.LUT)), step.lut)
			expectHash(t, "variants", hashHex(flattenVariants(state.Variants)), step.variants)
		})
	}
}

// TestCyclePaletteRotatesRangeOnly pins the range boundary: a cycle over from..to
// leaves every index outside it untouched in the palette and in the LUT, so two
// animations on disjoint ranges cannot corrupt each other (palette.cpp:52-90).
func TestCyclePaletteRotatesRangeOnly(t *testing.T) {
	// given
	const from, to = 60, 90

	before := newState(t, "l1_1.pal.hex")
	after := before

	// when
	after.CyclePalette(from, to)

	// then
	for index := range before.Palette {
		if outsideRange(index, from, to) && after.Palette[index] != before.Palette[index] {
			t.Fatalf("palette[%d] changed outside the cycled range", index)
		}
	}

	for i := range before.LUT {
		expectRowUnchangedOutside(t, before.LUT, after.LUT, i, from, to)
	}
}

// TestCyclePaletteRoundTrip pins the two directions as inverses: a forward cycle
// followed by a reverse one restores the palette, the variants and the LUT.
func TestCyclePaletteRoundTrip(t *testing.T) {
	// given
	const from, to = 1, 31

	before := newState(t, "l1_1.pal.hex")
	state := before

	// when
	state.CyclePalette(from, to)
	if state.Palette == before.Palette {
		t.Fatal("forward cycle left the palette unchanged")
	}
	state.CyclePaletteReverse(from, to)

	// then
	expectBytesEqual(t, flattenVariants(state.Variants), flattenVariants(before.Variants))
	expectBytesEqual(t, flattenLUT(state.LUT), flattenLUT(before.LUT))
}

// newState builds the palette state a cathedral level starts with.
func newState(t *testing.T, palette string) render.PaletteState {
	t.Helper()

	tables := light.MakeLightTables(light.DTYPE_CATHEDRAL)

	return render.NewPaletteState(loadPalette(t, palette), tables, assets.BlendingFor(0))
}

// outsideRange reports whether index lies outside from..to.
func outsideRange(index, from, to int) bool { return index < from || index > to }

// expectRowUnchangedOutside checks one LUT row keeps every entry whose two axes both
// lie outside the cycled range.
func expectRowUnchangedOutside(
	t *testing.T,
	before, after render.BlendLUT,
	row, from, to int,
) {
	t.Helper()

	for j := range before[row] {
		if outsideRange(row, from, to) && outsideRange(j, from, to) && after[row][j] != before[row][j] {
			t.Fatalf("lut[%d][%d] changed outside the cycled range", row, j)
		}
	}
}

// expectHash compares one after-state hash with the fixture's.
func expectHash(t *testing.T, name, got, want string) {
	t.Helper()

	if got != want {
		t.Fatalf("%s hash = %s, want %s", name, got, want)
	}
}

// replaySteps rebuilds the state named by target by applying the fixture's own step
// labels from the first row through it, so every row is checked along its path.
func replaySteps(
	t *testing.T,
	pal assets.Palette,
	tables light.LightTables,
	target string,
) render.PaletteState {
	t.Helper()

	state := render.NewPaletteState(pal, tables, assets.BlendingFor(0))

	for _, step := range readCycleSteps(t) {
		applyCycleStep(t, &state, step.name)

		if step.name == target {
			return state
		}
	}

	t.Fatalf("cycle step %q not found in the fixture", target)

	return state
}

// applyCycleStep runs the rotation a fixture label names.
func applyCycleStep(t *testing.T, state *render.PaletteState, label string) {
	t.Helper()

	switch label {
	case "init":
	case "cycle 1..31":
		state.CyclePalette(1, 31)
	case "cycle 16..31 reverse":
		state.CyclePaletteReverse(16, 31)
	case "cycle 1..8":
		state.CyclePalette(1, 8)
	default:
		t.Fatalf("unknown cycle step %q", label)
	}
}

// readCycleSteps decodes the cycle after-state fixture.
func readCycleSteps(t *testing.T) []cycleStep {
	t.Helper()

	body := readFixtureText(t, filepath.Join(fixtureDir, "cycle", "cycle_steps.tsv"))

	var steps []cycleStep
	for line := range strings.SplitSeq(body, "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		steps = append(steps, parseCycleStep(t, line))
	}

	if len(steps) == 0 {
		t.Fatal("cycle fixture holds no rows")
	}

	return steps
}

// parseCycleStep decodes one fixture row.
func parseCycleStep(t *testing.T, line string) cycleStep {
	t.Helper()

	fields := strings.Split(line, "\t")
	if len(fields) != 4 {
		t.Fatalf("cycle fixture: %d fields in %q, want 4", len(fields), line)
	}

	palette, err := hex.DecodeString(fields[1])
	if err != nil {
		t.Fatalf("cycle fixture palette hex: %v", err)
	}

	return cycleStep{name: fields[0], palette: palette, variants: fields[2], lut: fields[3]}
}

// flattenVariants returns the variant rows row-major as RGB triples.
func flattenVariants(variants render.Variants) []byte {
	flat := make([]byte, 0, len(variants)*len(variants[0])*3)
	for _, row := range variants {
		for _, color := range row {
			flat = append(flat, color[:]...)
		}
	}

	return flat
}

// flattenPalette returns the palette as RGB triples.
func flattenPalette(pal assets.Palette) []byte {
	flat := make([]byte, 0, len(pal)*3)
	for _, color := range pal {
		flat = append(flat, color[:]...)
	}

	return flat
}
