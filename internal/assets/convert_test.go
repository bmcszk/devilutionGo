package assets_test

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/assets"
)

// conversionFixture is one derived legacy-format input plus the C++ golden.
type conversionFixture struct {
	name   string
	input  string
	golden string
	widths []int
}

func TestConvertCL2ToCLXMatchesCXXGolden(t *testing.T) {
	fixtures := conversionFixtures(t, "cl2")
	frames := 0

	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			// given a derived CL2 input and the devilutionX C++ converter output
			input := readHexFixture(t, fixture.input)
			want := readHexFixture(t, fixture.golden)
			frames += len(fixture.widths)

			// when
			got, err := assets.ConvertCL2ToCLX(input, fixture.widths)
			// then
			if err != nil {
				t.Fatalf("ConvertCL2ToCLX: %v", err)
			}
			assertBytesEqual(t, fixture.name, got, want)
		})
	}

	// cl2_to_clx.cpp is byte-parity-tested on at least 20 sampled frames.
	if frames < 20 {
		t.Fatalf("parity corpus covers %d CL2 frames, want at least 20", frames)
	}
}

func TestConvertCELToCLXMatchesCXXGolden(t *testing.T) {
	fixtures := conversionFixtures(t, "cel")
	frames := 0

	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			// given a derived CEL input and the devilutionX C++ converter output
			input := readHexFixture(t, fixture.input)
			want := readHexFixture(t, fixture.golden)
			frames += len(fixture.widths)

			// when
			got, err := assets.ConvertCELToCLX(input, fixture.widths)
			// then
			if err != nil {
				t.Fatalf("ConvertCELToCLX: %v", err)
			}
			assertBytesEqual(t, fixture.name, got, want)
		})
	}

	if frames < 20 {
		t.Fatalf("parity corpus covers %d CEL frames, want at least 20", frames)
	}
}

func TestConvertCL2ToCLXRejectsTruncatedInput(t *testing.T) {
	// given a truncated list header
	input := []byte{2, 0, 0, 0, 12, 0}
	// when
	_, err := assets.ConvertCL2ToCLX(input, []int{4})
	// then
	if err == nil {
		t.Fatal("expected an error for a truncated frame list")
	}
}

func TestConvertCELToCLXRejectsTruncatedInput(t *testing.T) {
	// given a list header whose single frame runs past the end of the buffer
	input := []byte{1, 0, 0, 0, 12, 0, 0, 0, 200, 0, 0, 0}
	// when
	_, err := assets.ConvertCELToCLX(input, []int{4})
	// then
	if err == nil {
		t.Fatal("expected an error for a frame that ends past the buffer")
	}
}

func TestConvertCL2ToCLXRequiresOneWidthOrOnePerFrame(t *testing.T) {
	input := []byte{1, 0, 0, 0, 12, 0, 0, 0, 16, 0, 0, 0, 4, 0, 0, 0}

	// when
	_, err := assets.ConvertCL2ToCLX(input, nil)
	// then
	if err == nil {
		t.Fatal("expected an error when no width is supplied")
	}

	// when
	_, err = assets.ConvertCL2ToCLX(input, []int{4, 5})
	// then
	if err == nil {
		t.Fatal("expected an error when fewer widths than frames are supplied")
	}
}

func TestConvertSingleWidthAppliesToEveryFrame(t *testing.T) {
	// given a multi-frame fixture whose frames share one width
	fixture := findFixture(t, "cl2", "data_dirtybuc")
	input := readHexFixture(t, fixture.input)
	want := readHexFixture(t, fixture.golden)

	// when the width list holds one value (cl2_to_clx.cpp:58-59)
	got, err := assets.ConvertCL2ToCLX(input, fixture.widths[:1])
	// then
	if err != nil {
		t.Fatalf("ConvertCL2ToCLX: %v", err)
	}
	assertBytesEqual(t, fixture.name, got, want)
}

// assertBytesEqual reports the first differing byte, which is far more useful
// than "not equal" when an encoder run boundary moves.
func assertBytesEqual(t *testing.T, name string, got, want []byte) {
	t.Helper()

	if len(got) == len(want) && string(got) == string(want) {
		return
	}

	limit := min(len(got), len(want))
	for i := range limit {
		if got[i] != want[i] {
			t.Fatalf("%s: byte %d = 0x%02X, want 0x%02X (got %d bytes, want %d)",
				name, i, got[i], want[i], len(got), len(want))
		}
	}

	t.Fatalf("%s: length %d, want %d", name, len(got), len(want))
}

// findFixture resolves one named fixture, so a case cannot silently disappear.
func findFixture(tb testing.TB, kind, name string) conversionFixture {
	tb.Helper()

	for _, fixture := range conversionFixtures(tb, kind) {
		if fixture.name == name {
			return fixture
		}
	}
	tb.Fatalf("no %s fixture named %s", kind, name)

	return conversionFixture{}
}

// conversionFixtures lists every committed fixture pair under
// internal/testdata/gfx/<kind>. The producer is
// internal/testdata/gfx/producer/make_fixtures.py; a missing golden fails.
func conversionFixtures(tb testing.TB, kind string) []conversionFixture {
	tb.Helper()

	inputs, err := filepath.Glob(filepath.Join(gfxFixtureDir, kind, "*."+kind+".hex"))
	if err != nil {
		tb.Fatalf("glob %s fixtures: %v", kind, err)
	}
	if len(inputs) == 0 {
		tb.Fatalf("no %s fixtures under %s", kind, filepath.Join(gfxFixtureDir, kind))
	}

	fixtures := make([]conversionFixture, 0, len(inputs))
	for _, input := range inputs {
		name := strings.TrimSuffix(filepath.Base(input), "."+kind+".hex")
		fixtures = append(fixtures, conversionFixture{
			name:   name,
			input:  input,
			golden: filepath.Join(gfxFixtureDir, kind, name+"."+kind+".clx.hex"),
			widths: readWidthsFixture(tb, filepath.Join(gfxFixtureDir, kind, name+".widths")),
		})
	}

	return fixtures
}

// readWidthsFixture reads a per-frame width list; one value per line frame.
func readWidthsFixture(tb testing.TB, path string) []int {
	tb.Helper()

	dump, err := os.ReadFile(path)
	if err != nil {
		tb.Fatalf("read widths %s: %v", path, err)
	}

	widths := make([]int, 0, strings.Count(string(dump), ",")+1)
	for _, field := range strings.Split(strings.TrimSpace(string(dump)), ",") {
		width, convErr := strconv.Atoi(strings.TrimSpace(field))
		if convErr != nil {
			tb.Fatalf("parse widths %s: %v", path, convErr)
		}
		widths = append(widths, width)
	}

	return widths
}
