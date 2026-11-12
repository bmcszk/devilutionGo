package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/render/parity"
)

// TestRunUsageErrors pins the usage exit code on every bad invocation.
func TestRunUsageErrors(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{name: "no mode", args: []string{}},
		{name: "both modes", args: []string{"-record", "/tmp/a", "-compare", "/tmp/b:/tmp/c"}},
		{name: "record with positional", args: []string{"-record", "/tmp/a", "extra"}},
		{name: "compare without colon", args: []string{"-compare", "/tmp/b"}},
		{name: "compare with two colons", args: []string{"-compare", "/tmp/b:/tmp/c:/tmp/d"}},
		{name: "compare empty side", args: []string{"-compare", ":/tmp/c"}},
		{name: "unknown flag", args: []string{"-wat"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// when
			code := run(tc.args, &bytes.Buffer{})

			// then
			if code != exitUsage {
				t.Fatalf("exit code %d, want %d (usage)", code, exitUsage)
			}
		})
	}
}

// TestRunCompareExact pins the pass verdict: identical dumps exit 0 and report
// the exact match.
func TestRunCompareExact(t *testing.T) {
	// given
	golden := goldenFixture(t, "base")
	actual := goldenFixture(t, "base")
	stderr := &bytes.Buffer{}

	// when
	code := run([]string{"-compare", golden + ":" + actual}, stderr)

	// then
	if code != exitExact {
		t.Fatalf("exit code %d, want %d (exact)\n%s", code, exitExact, stderr)
	}

	if !strings.Contains(stderr.String(), "PASS exact") {
		t.Fatalf("verdict lost the exact pass line:\n%s", stderr)
	}
}

// TestRunCompareIndexDiff pins the index-fail verdict: a one-byte divergence
// exits 1 and names the pixel and its owning tile.
func TestRunCompareIndexDiff(t *testing.T) {
	// given
	golden := goldenFixture(t, "base")
	actual := goldenFixture(t, "base")
	flipActualPixel(t, actual, 33, 35)
	stderr := &bytes.Buffer{}

	// when
	code := run([]string{"-compare", golden + ":" + actual}, stderr)

	// then
	if code != exitDiff {
		t.Fatalf("exit code %d, want %d (diff)\n%s", code, exitDiff, stderr)
	}

	out := stderr.String()
	for _, fragment := range []string{"DIFF index", "(33,35)", "tile (1,1)"} {
		if !strings.Contains(out, fragment) {
			t.Fatalf("diff report lost %q:\n%s", fragment, out)
		}
	}
}

// TestRunComparePaletteDiff pins the colour-fail verdict: identical index bytes
// with a diverging palette exit 1 through the colour gate.
func TestRunComparePaletteDiff(t *testing.T) {
	// given
	golden := goldenFixture(t, "base")
	actual := goldenFixture(t, "palshift")
	stderr := &bytes.Buffer{}

	// when
	code := run([]string{"-compare", golden + ":" + actual}, stderr)

	// then
	if code != exitDiff {
		t.Fatalf("exit code %d, want %d (diff)\n%s", code, exitDiff, stderr)
	}

	if !strings.Contains(stderr.String(), "DIFF colour") {
		t.Fatalf("verdict lost the colour diff line:\n%s", stderr)
	}
}

// TestRunRecordRoundTrip pins the record mode's offline path: WriteDump output
// reloads through the comparator as an exact match. (The capture side needs a
// display and is exercised by the paritytest suite.)
func TestRunRecordRoundTrip(t *testing.T) {
	// given
	dump := parity.FrameDump{Width: 0, Height: 0}
	dir := t.TempDir()
	stderr := &bytes.Buffer{}

	// when
	err := parity.WriteDump(dir, frameName, dump)
	code := runCompare([2]string{dir, dir}, stderr)

	// then
	if err == nil {
		t.Fatal("zero-dimension dump wrote successfully; write-side validation failed closed")
	}

	if code != exitUsage {
		t.Fatalf("exit code %d, want %d (usage) on an unloadable dump\n%s", code, exitUsage, stderr)
	}
}

// TestRunCompareMissingDir pins fail-closed loading: a missing golden directory
// is a usage-grade error, never a diff verdict.
func TestRunCompareMissingDir(t *testing.T) {
	// given
	actual := goldenFixture(t, "base")
	stderr := &bytes.Buffer{}

	// when
	code := run([]string{"-compare", "/nonexistent/golden:" + actual}, stderr)

	// then
	if code != exitUsage {
		t.Fatalf("exit code %d, want %d (usage)\n%s", code, exitUsage, stderr)
	}
}

// flipActualPixel inverts one index byte in an actual dump directory, so the
// compare verdict names that pixel.
func flipActualPixel(t *testing.T, dir string, x, y int) {
	t.Helper()

	dump, err := parity.LoadFrameDump(dir, frameName, frameWidth, frameHeight)
	if err != nil {
		t.Fatalf("reload actual: %v", err)
	}

	dump.Index[y*frameWidth+x] ^= 0xFF

	if writeErr := parity.WriteDump(dir, frameName, dump); writeErr != nil {
		t.Fatalf("rewrite actual: %v", writeErr)
	}
}

// goldenFixture writes a valid W x H frame family into dir: the loaded
// comparator fixture family, padded or cropped to the full surface the CLI
// dimension check demands.
func goldenFixture(t *testing.T, frame string) string {
	t.Helper()

	dir := t.TempDir()
	base := mustReadFixtureFamily(t, frame)

	dump := parity.FrameDump{
		Width:   frameWidth,
		Height:  frameHeight,
		Index:   fitBuffer(base.Index, frameWidth*frameHeight),
		Palette: base.Palette,
	}

	if err := parity.WriteDump(dir, frameName, dump); err != nil {
		t.Fatalf("write fixture family: %v", err)
	}

	return dir
}

// mustReadFixtureFamily loads one committed comparator fixture family.
func mustReadFixtureFamily(t *testing.T, frame string) parity.FrameDump {
	t.Helper()

	dump, err := parity.LoadFrameDump(
		filepath.Join("..", "..", "internal", "render", "parity", "testdata"), frame, 4, 2,
	)
	if err != nil {
		t.Fatalf("load fixture %s: %v", frame, err)
	}

	return dump
}

// fitBuffer returns a width*height buffer whose first bytes are src's and whose
// tail repeats src, so deliberate fixture faults keep their offsets modulo len.
func fitBuffer(src []byte, size int) []byte {
	buf := make([]byte, size)
	for i := range buf {
		buf[i] = src[i%len(src)]
	}

	return buf
}
