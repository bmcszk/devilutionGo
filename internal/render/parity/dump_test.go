package parity_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/render/parity"
)

// TestLoadFrameDump pins the loader byte-for-byte against the committed base
// family: the 4x2 index buffer, the 256x4 palette rows and the PPM colour
// reference.
func TestLoadFrameDump(t *testing.T) {
	// given
	paletteBytes := loadFixture(t, "base.pal.bin")
	var wantPalette parity.Palette
	for entry := range wantPalette {
		offset := entry * 4
		wantPalette[entry] = parity.RGBA{
			R: paletteBytes[offset], G: paletteBytes[offset+1],
			B: paletteBytes[offset+2], A: paletteBytes[offset+3],
		}
	}

	// when
	dump := mustLoadDump(t, "base", 4, 2)

	// then
	expectDumpEqual(t, dump, parity.FrameDump{
		Width: 4, Height: 2,
		Index:   loadFixture(t, "base.idx.bin"),
		Palette: wantPalette,
	})
}

// TestLoadFrameDumpSizeViolations pins strict size validation: every short or
// malformed file fails with the first violating offset named.
func TestLoadFrameDumpSizeViolations(t *testing.T) {
	cases := []struct {
		name     string
		frame    string
		evidence []string
	}{
		{
			name:     "index buffer short",
			frame:    "short",
			evidence: []string{"index buffer", "have 7 bytes", "want 8", "offset 7"},
		},
		{
			name:     "palette short",
			frame:    "shortpal",
			evidence: []string{"palette", "have 1023 bytes", "want 1024", "offset 1023"},
		},
		{
			name:     "ppm raster short",
			frame:    "shortppm",
			evidence: []string{"ppm", "have 21 bytes", "want 24", "offset 32"},
		},
		{
			name:     "ppm magic",
			frame:    "badmagic",
			evidence: []string{"ppm", "offset 0", `"P6"`},
		},
		{
			name:     "ppm maxval",
			frame:    "badmaxval",
			evidence: []string{"ppm", "offset 7", "255"},
		},
		{
			name:     "ppm dimensions",
			frame:    "dimmismatch",
			evidence: []string{"ppm", "height: have 3", "want 2", "offset 5"},
		},
		{
			name:     "ppm pixel mismatch",
			frame:    "ppmmismatch",
			evidence: []string{"ppm", "(1,1)", "offset 27"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// when
			_, err := parity.LoadFrameDump("testdata", tc.frame, 4, 2)

			// then
			expectViolation(t, err, tc.evidence...)
		})
	}
}

// TestWriteDumpRoundTrip pins the golden codec: write then read restores every
// field byte-for-byte.
func TestWriteDumpRoundTrip(t *testing.T) {
	// given
	dump := mustLoadDump(t, "base", 4, 2)

	// when
	reloaded := roundTripDump(t, dump)

	// then
	expectDumpEqual(t, reloaded, dump)
}

// TestWriteDumpRoundTripDropsLightLevels pins the light-level contract: the
// fields are diagnostic capture evidence, and the codec never persists them.
func TestWriteDumpRoundTripDropsLightLevels(t *testing.T) {
	// given
	dump := mustLoadDump(t, "wide-base", 40, 40)
	dump.LightLevels = lightLevelsFor(dump)

	// when
	reloaded := roundTripDump(t, dump)

	// then
	if reloaded.LightLevels != nil {
		t.Fatal("light levels persisted; the codec treats them as diagnostic only")
	}

	dump.LightLevels = nil
	expectDumpEqual(t, reloaded, dump)
}

// TestWriteDumpMatchesFixtureBytes pins the writer's byte format against the
// committed family: a written dump reproduces every file verbatim, so Go-side
// records are interchangeable with the C++ dumper's.
func TestWriteDumpMatchesFixtureBytes(t *testing.T) {
	// given
	dump := mustLoadDump(t, "base", 4, 2)
	dir := t.TempDir()

	// when
	err := parity.WriteDump(dir, "base", dump)
	// then
	if err != nil {
		t.Fatalf("write dump: %v", err)
	}

	for _, suffix := range []string{"idx.bin", "pal.bin", "ppm"} {
		written, readErr := os.ReadFile(filepath.Join(dir, "base."+suffix))
		if readErr != nil {
			t.Fatalf("read written %s: %v", suffix, readErr)
		}

		if !bytes.Equal(written, loadFixture(t, "base."+suffix)) {
			t.Fatalf("written %s diverges from the committed golden", suffix)
		}
	}
}

// TestWriteDumpRejectsInvalidDump pins write-side validation: a dump whose
// buffers do not match its dimensions fails with the first violation named.
func TestWriteDumpRejectsInvalidDump(t *testing.T) {
	cases := []struct {
		name     string
		dump     parity.FrameDump
		evidence []string
	}{
		{
			name:     "index buffer short",
			dump:     parity.FrameDump{Width: 4, Height: 2, Index: make([]byte, 7)},
			evidence: []string{"index buffer", "have 7 bytes", "want 8", "offset 7"},
		},
		{
			name: "light levels long",
			dump: parity.FrameDump{
				Width: 4, Height: 2,
				Index:       make([]byte, 8),
				LightLevels: make([]byte, 9),
			},
			evidence: []string{"light levels", "have 9 bytes", "want 8", "offset 9"},
		},
		{
			name:     "zero dimensions",
			dump:     parity.FrameDump{Width: 0, Height: 2, Index: []byte{}},
			evidence: []string{"dimensions"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// when
			err := parity.WriteDump(t.TempDir(), "invalid", tc.dump)

			// then
			expectViolation(t, err, tc.evidence...)
		})
	}
}

// roundTripDump writes dump's family into a temp dir and reloads it.
func roundTripDump(t *testing.T, dump parity.FrameDump) parity.FrameDump {
	t.Helper()

	dir := t.TempDir()

	if err := parity.WriteDump(dir, "frame", dump); err != nil {
		t.Fatalf("write dump: %v", err)
	}

	reloaded, err := parity.LoadFrameDump(dir, "frame", dump.Width, dump.Height)
	if err != nil {
		t.Fatalf("reload dump: %v", err)
	}

	return reloaded
}

// loadFixture reads a committed testdata golden.
func loadFixture(t *testing.T, name string) []byte {
	t.Helper()

	body, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}

	return body
}

// mustLoadDump loads one committed testdata family, failing the test on any
// violation.
func mustLoadDump(t *testing.T, frame string, width, height int) parity.FrameDump {
	t.Helper()

	dump, err := parity.LoadFrameDump("testdata", frame, width, height)
	if err != nil {
		t.Fatalf("load %s: %v", frame, err)
	}

	return dump
}

// expectDumpEqual asserts two dumps carry byte-identical fields.
func expectDumpEqual(t *testing.T, got, want parity.FrameDump) {
	t.Helper()

	if got.Width != want.Width || got.Height != want.Height {
		t.Fatalf("dimensions %dx%d, want %dx%d", got.Width, got.Height, want.Width, want.Height)
	}

	if !bytes.Equal(got.Index, want.Index) {
		t.Fatalf("index bytes diverge (got len %d, want len %d)", len(got.Index), len(want.Index))
	}

	if got.Palette != want.Palette {
		t.Fatal("palette rows diverge")
	}

	if !bytes.Equal(got.LightLevels, want.LightLevels) {
		t.Fatalf("light levels diverge (got len %d, want len %d)", len(got.LightLevels), len(want.LightLevels))
	}
}

// expectViolation asserts the load error names its first-violation evidence.
func expectViolation(t *testing.T, err error, evidence ...string) {
	t.Helper()

	if err == nil {
		t.Fatal("load succeeded, want a violation error")
	}

	for _, fragment := range evidence {
		if !strings.Contains(err.Error(), fragment) {
			t.Fatalf("violation error lost %q: %v", fragment, err)
		}
	}
}

// lightLevelsFor returns deterministic per-pixel light levels for a dump.
func lightLevelsFor(dump parity.FrameDump) []byte {
	levels := make([]byte, dump.Width*dump.Height)
	for i := range levels {
		levels[i] = uint8(i % 16)
	}

	return levels
}
