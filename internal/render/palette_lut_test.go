package render_test

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/assets"
	"github.com/bmcszk/devilutionGo/internal/render"
)

// fixtureDir holds the committed render goldens.
const fixtureDir = "../testdata/render"

// lutCases covers the three blend ranges LoadPaletteAndInitBlending selects
// (palette.cpp:188-200): town and cathedral blend the whole palette, caves and
// crypt skip 1..31, the nest skips 1..15.
var lutCases = []lutCase{
	{name: "town", palette: "town.pal.hex", skipFrom: -1, skipTo: -1, lut: "town.txt"},
	{name: "cathedral", palette: "l1_1.pal.hex", skipFrom: -1, skipTo: -1, lut: "l1_1.txt"},
	{name: "crypt", palette: "l5base.pal.hex", skipFrom: 1, skipTo: 31, lut: "l5base.txt"},
}

// lutCase is one palette + blend-range combination with its golden LUT file.
type lutCase struct {
	name     string
	palette  string
	skipFrom int
	skipTo   int
	lut      string
}

// TestBlendingLUTMatchesUpstreamGolden pins every byte of the blend LUT against a
// dump of paletteTransparencyLookup produced by the pinned upstream
// GenerateBlendedLookupTable (palette_blending.cpp:44-60).
func TestBlendingLUTMatchesUpstreamGolden(t *testing.T) {
	for _, tc := range lutCases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			pal := loadPalette(t, tc.palette)
			want := readFlatBytes(t, filepath.Join(fixtureDir, "lut", tc.lut), 256*256)

			// when
			lut := render.BuildBlendingLUT(pal, tc.skipFrom, tc.skipTo)

			// then
			expectBytesEqual(t, flattenLUT(lut), want)
		})
	}
}

// TestBlendingLUTLaws pins the two laws GenerateBlendedLookupTable builds in rather
// than derives: the diagonal is the identity (palette_blending.cpp:47) and the table
// is symmetric because the loop copies the lower triangle from the upper one
// (palette_blending.cpp:50-51).
func TestBlendingLUTLaws(t *testing.T) {
	// given
	pal := loadPalette(t, "town.pal.hex")

	// when
	lut := render.BuildBlendingLUT(pal, -1, -1)

	// then
	expectDiagonalIsIdentity(t, lut)
	expectSymmetric(t, lut)
}

// TestBlendingLUTSkipRange pins the skip semantics: no blend of two distinct
// colours lands in the range the level art reserves (palette_blending.cpp:48-57).
func TestBlendingLUTSkipRange(t *testing.T) {
	// given
	pal := loadPalette(t, "l5base.pal.hex")
	const skipFrom, skipTo = 1, 31

	// when
	lut := render.BuildBlendingLUT(pal, skipFrom, skipTo)

	// then
	for i := range lut {
		if first := findSkippedBlend(lut[i], i, skipFrom, skipTo); first >= 0 {
			t.Fatalf("blend(%d,%d) = %d, inside the skipped range %d..%d",
				i, first, lut[i][first], skipFrom, skipTo)
		}
	}
}

// TestBlendingLUTPicksNearestColour pins the search itself: every table entry is the
// palette colour closest to the 50% blend of its two indices. Which of two equally
// distant colours wins is the tree-order tie-break pinned by the golden bytes.
func TestBlendingLUTPicksNearestColour(t *testing.T) {
	// given
	pal := loadPalette(t, "l1_1.pal.hex")
	const skipFrom, skipTo = 1, 31

	// when
	lut := render.BuildBlendingLUT(pal, skipFrom, skipTo)

	// then
	for i := range lut {
		expectRowIsNearest(t, pal, lut[i], i, skipFrom, skipTo)
	}
}

// TestBlendingLUTDeterministic pins the LUT as a function of its inputs only.
func TestBlendingLUTDeterministic(t *testing.T) {
	// given
	pal := loadPalette(t, "l1_1.pal.hex")

	// when
	first := render.BuildBlendingLUT(pal, -1, -1)
	second := render.BuildBlendingLUT(pal, -1, -1)

	// then
	if hashHex(flattenLUT(first)) != hashHex(flattenLUT(second)) {
		t.Fatal("two builds of the same palette produced different bytes")
	}
}

// expectDiagonalIsIdentity checks lut[i][i] == i (palette_blending.cpp:47).
func expectDiagonalIsIdentity(t *testing.T, lut render.BlendLUT) {
	t.Helper()

	for i := range lut {
		if got := lut[i][i]; got != uint8(i) { //nolint:gosec // i < paletteSize.
			t.Fatalf("diagonal [%d][%d] = %d, want %d", i, i, got, i)
		}
	}
}

// expectSymmetric checks lut[i][j] == lut[j][i] (palette_blending.cpp:50-51).
func expectSymmetric(t *testing.T, lut render.BlendLUT) {
	t.Helper()

	for i := range lut {
		for j := range lut {
			if got, want := lut[i][j], lut[j][i]; got != want {
				t.Fatalf("asymmetric at [%d][%d]=%d vs [%d][%d]=%d", i, j, got, j, i, want)
			}
		}
	}
}

// findSkippedBlend returns the first column whose blend lands in the skipped range,
// or -1 when the row is clean.
func findSkippedBlend(row render.BlendRow, index, skipFrom, skipTo int) int {
	for j := range row {
		if blend := int(row[j]); index != j && blend >= skipFrom && blend <= skipTo {
			return j
		}
	}

	return -1
}

// expectRowIsNearest checks one LUT row against a brute-force nearest-colour scan.
func expectRowIsNearest(
	t *testing.T,
	pal assets.Palette,
	row render.BlendRow,
	index, skipFrom, skipTo int,
) {
	t.Helper()

	for j := range row {
		if index == j {
			continue // the diagonal is written as the identity, not searched (:47).
		}

		target := blendOf(pal[index], pal[j])
		got := row[j]
		if want := nearestDistance(pal, target, skipFrom, skipTo); distance(pal[got], target) != want {
			t.Fatalf("blend(%d,%d) = %d at distance %d, nearest is %d", index, j, got,
				distance(pal[got], target), want)
		}
	}
}

// nearestDistance is the brute-force reference distance the k-d tree must match.
func nearestDistance(pal assets.Palette, target [3]uint8, skipFrom, skipTo int) int {
	best := distance(pal[0], target)

	for index := range pal {
		if index >= skipFrom && index <= skipTo {
			continue
		}

		if d := distance(pal[index], target); d < best {
			best = d
		}
	}

	return best
}

// distance is the squared euclidean distance between two colours.
func distance(a, b [3]uint8) int {
	sum := 0
	for component := range a {
		delta := int(a[component]) - int(b[component])
		sum += delta * delta
	}

	return sum
}

// blendOf is the integer-halving average of two palette colours.
func blendOf(a, b [3]uint8) [3]uint8 {
	return [3]uint8{
		uint8((int(a[0]) + int(b[0])) / 2),
		uint8((int(a[1]) + int(b[1])) / 2),
		uint8((int(a[2]) + int(b[2])) / 2),
	}
}

// loadPalette reads a committed .pal hexdump, from the render fixtures first and the
// earlier gfx fixtures second (both are commit-pinned dumps, not MPQ reads).
func loadPalette(t *testing.T, name string) assets.Palette {
	t.Helper()

	path := filepath.Join(fixtureDir, "pal", name)
	if _, err := os.Stat(path); err != nil {
		path = filepath.Join("..", "testdata", "gfx", "pal", name)
	}

	pal, err := assets.LoadPaletteFromBytes(readFlatBytes(t, path, 768))
	if err != nil {
		t.Fatalf("load palette %s: %v", name, err)
	}

	return *pal
}

// flattenLUT returns the LUT row-major, the layout upstream's C array has.
func flattenLUT(lut render.BlendLUT) []byte {
	flat := make([]byte, 0, len(lut)*len(lut[0]))
	for _, row := range lut {
		flat = append(flat, row[:]...)
	}

	return flat
}

// hashHex returns the lowercase sha256 of a byte slice as the fixtures store it.
func hashHex(data []byte) string {
	sum := sha256.Sum256(data)

	return hex.EncodeToString(sum[:])
}

// readFlatBytes decodes a whitespace-wrapped hex fixture and refuses a short one.
func readFlatBytes(t *testing.T, path string, want int) []byte {
	t.Helper()

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}

	decoded, err := hex.DecodeString(strings.Join(strings.Fields(string(body)), ""))
	if err != nil {
		t.Fatalf("decode fixture %s: %v", path, err)
	}

	if len(decoded) != want {
		t.Fatalf("fixture %s: %d bytes, want %d", path, len(decoded), want)
	}

	return decoded
}

// readFixtureText reads a committed fixture as text.
func readFixtureText(t *testing.T, path string) string {
	t.Helper()

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}

	return string(body)
}

// expectBytesEqual fails with the first differing offset.
func expectBytesEqual(t *testing.T, got, want []byte) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("length: got %d want %d", len(got), len(want))
	}

	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("byte %d: got 0x%02x want 0x%02x", i, got[i], want[i])
		}
	}
}
