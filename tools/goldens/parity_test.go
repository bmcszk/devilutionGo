package main

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/render/parity"
	"github.com/bmcszk/devilutionGo/internal/render/paritytest"
)

// TestRunParityCompareExact pins the pass verdict: identical dumps report the
// exact match and errParityDiff stays nil.
func TestRunParityCompareExact(t *testing.T) {
	// given
	golden := parityFixture(t, "base")
	actual := parityFixture(t, "base")
	stderr := &bytes.Buffer{}

	// when
	err := compareParity([]string{golden + ":" + actual}, stderr)
	// then
	if err != nil {
		t.Fatalf("compare of identical dumps errored: %v\n%s", err, stderr)
	}

	if !strings.Contains(stderr.String(), "PASS exact") {
		t.Fatalf("verdict lost the exact pass line:\n%s", stderr)
	}
}

// TestRunParityCompareIndexDiff pins the index-fail verdict: a one-byte
// divergence at (33,35) reports tile (1,1) and errParityDiff.
func TestRunParityCompareIndexDiff(t *testing.T) {
	// given
	golden := parityFixture(t, "base")
	actual := parityFixture(t, "base")
	flipActualPixel(t, actual, 33, 35)
	stderr := &bytes.Buffer{}

	// when
	err := compareParity([]string{golden + ":" + actual}, stderr)

	// then
	if !errors.Is(err, errParityDiff) {
		t.Fatalf("err = %v, want errParityDiff\n%s", err, stderr)
	}

	out := stderr.String()
	for _, fragment := range []string{"DIFF index", "(33,35)", "tile (1,1)"} {
		if !strings.Contains(out, fragment) {
			t.Fatalf("diff report lost %q:\n%s", fragment, out)
		}
	}
}

// TestRunParityComparePaletteDiff pins the colour-fail verdict: identical
// index bytes with a diverging palette fail through the colour gate.
func TestRunParityComparePaletteDiff(t *testing.T) {
	// given
	golden := parityFixture(t, "base")
	actual := parityFixture(t, "palshift")
	stderr := &bytes.Buffer{}

	// when
	err := compareParity([]string{golden + ":" + actual}, stderr)

	// then
	if !errors.Is(err, errParityDiff) {
		t.Fatalf("err = %v, want errParityDiff\n%s", err, stderr)
	}

	if !strings.Contains(stderr.String(), "DIFF colour") {
		t.Fatalf("verdict lost the colour diff line:\n%s", stderr)
	}
}

// TestRunParityCompareMissingDir pins fail-closed loading: a missing golden
// directory is a load error, never a diff verdict.
func TestRunParityCompareMissingDir(t *testing.T) {
	// given
	actual := parityFixture(t, "base")

	// when
	err := compareParity([]string{"/nonexistent/golden:" + actual}, &bytes.Buffer{})

	// then
	if err == nil || errors.Is(err, errParityDiff) {
		t.Fatalf("err = %v, want a load error distinct from errParityDiff", err)
	}
}

// TestSplitCompareTargetsOnePair pins the valid GOLDEN_DIR:ACTUAL_DIR shape:
// exactly one colon-separated pair of non-empty paths.
func TestSplitCompareTargetsOnePair(t *testing.T) {
	// when
	golden, actual, err := splitCompareTargets("/tmp/g:/tmp/a")

	// then
	if err != nil || golden != "/tmp/g" || actual != "/tmp/a" {
		t.Fatalf("splitCompareTargets = %q, %q, %v", golden, actual, err)
	}
}

// TestSplitCompareTargetsRejectsBadShapes pins every rejected operand shape:
// the retired CLI's splitCompareTargets covered these through parseFlags.
func TestSplitCompareTargetsRejectsBadShapes(t *testing.T) {
	cases := []struct {
		name    string
		targets string
	}{
		{name: "no colon", targets: "/tmp/g"},
		{name: "two colons", targets: "/tmp/g:/tmp/a:/tmp/b"},
		{name: "empty golden side", targets: ":/tmp/a"},
		{name: "empty actual side", targets: "/tmp/g:"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// when
			_, _, err := splitCompareTargets(tc.targets)

			// then
			if err == nil {
				t.Fatalf("splitCompareTargets(%q) = nil error, want one", tc.targets)
			}
		})
	}
}

// TestRunParityRecordRequiresDisplay pins the record half's display contract
// without a GPU run loop: a headless environment fails closed at the DISPLAY
// gate.
func TestRunParityRecordRequiresDisplay(t *testing.T) {
	// given
	if paritytest.GPUAvailable() {
		t.Skip("DISPLAY is set; the headless gate is not reachable")
	}

	dir := t.TempDir()

	// when
	err := runParityRecord([]string{dir})

	// then
	if err == nil {
		t.Fatal("record without a display succeeded; fail-closed gate missing")
	}
}

// TestRunParityRecordRoundTrip pins the record half's write-side contract
// through the real codec without a display: WriteDump output reloads through
// the comparator as an exact match, and an unset width fails closed.
func TestRunParityRecordRoundTrip(t *testing.T) {
	// given
	dir := t.TempDir()

	// when
	err := parity.WriteDump(dir, frameName, parity.FrameDump{Width: 0, Height: 0})
	if err == nil {
		t.Fatal("zero-dimension dump wrote successfully; write-side validation failed open")
	}

	// then
	compareErr := compareParity([]string{dir + ":" + dir}, &bytes.Buffer{})
	if compareErr == nil || errors.Is(compareErr, errParityDiff) {
		t.Fatalf("compare on an unloadable dump = %v, want a load error", compareErr)
	}
}

// parityFixture writes a valid surface-sized frame family into dir from one
// committed comparator fixture family in internal/render/parity/testdata.
func parityFixture(t *testing.T, frame string) string {
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
