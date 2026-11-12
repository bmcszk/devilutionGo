package parity_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/render/parity"
)

// TestDiffIdenticalDumpsAreExact pins the pass branch.
func TestDiffIdenticalDumpsAreExact(t *testing.T) {
	// given
	a, b := loadTwoFamilies(t, "base", "base", 4, 2)

	// when
	report := parity.Diff(a, b, parity.Options{})

	// then
	expectExact(t, report)

	if len(report.ColourDiffs) != 0 {
		t.Fatalf("exact match carries colour diffs: %+v", report.ColourDiffs)
	}

	if len(report.Masked) != 0 {
		t.Fatalf("exact match carries masked pixels: %+v", report.Masked)
	}
}

// TestDiffIndexMismatch pins the fail branch: the first differing index byte
// names its pixel and its owning 32px tile.
func TestDiffIndexMismatch(t *testing.T) {
	// given
	a, b := loadTwoFamilies(t, "base", "flipped", 4, 2)

	// when
	report := parity.Diff(a, b, parity.Options{})

	// then
	expectFirstDiff(t, report, 1, 1)

	if got, want := a.Index[5], b.Index[5]; got == want {
		t.Fatalf("fixture broken: both dumps carry index %d at (1,1)", got)
	}
}

// TestDiffFirstMismatchWins pins the scan order: with two deliberate mismatches
// the report names the row-major-first one.
func TestDiffFirstMismatchWins(t *testing.T) {
	// given
	a := mustLoadDump(t, "base", 4, 2)
	b := mustLoadDump(t, "base", 4, 2)
	b.Index[2] ^= 0xFF
	b.Index[5] ^= 0xFF

	// when
	report := parity.Diff(a, b, parity.Options{})

	// then
	expectFirstDiff(t, report, 2, 0)
}

// TestDiffTileAt32pxGrid pins the tile attribution on a mismatch far from the
// origin: pixel (33,35) sits in tile (1,1) of the 32px grid.
func TestDiffTileAt32pxGrid(t *testing.T) {
	// given
	a, b := loadTwoFamilies(t, "wide-base", "wide-flipped", 40, 40)

	// when
	report := parity.Diff(a, b, parity.Options{})

	// then
	expectFirstDiff(t, report, 33, 35)
}

// TestDiffPaletteMismatchOnly pins the colour gate: identical index bytes with
// a diverging palette leave the index gate green and fail the palette-expanded
// colour compare at the pixels that entry paints.
func TestDiffPaletteMismatchOnly(t *testing.T) {
	// given
	a, b := loadTwoFamilies(t, "base", "palshift", 4, 2)

	// when
	report := parity.Diff(a, b, parity.Options{})

	// then
	expectExact(t, report)
	expectColourDiff(t, report, 2, 0)
}

// TestDiffAllowListMasksColourOnly pins the allow-list semantics: a masked
// region drops out of ColourDiffs and is reported as masked, while the index
// gate stays untouched — the index-exact law is never relaxable.
func TestDiffAllowListMasksColourOnly(t *testing.T) {
	// given
	a, b := loadTwoFamilies(t, "base", "palshift", 4, 2)
	allow := parity.Region{X: 2, Y: 0, Width: 1, Height: 1}

	// when
	report := parity.Diff(a, b, parity.Options{AllowList: []parity.Region{allow}})

	// then
	expectExact(t, report)

	if len(report.ColourDiffs) != 0 {
		t.Fatalf("masked pixel still reported: %+v", report.ColourDiffs)
	}

	if len(report.Masked) != 1 || report.Masked[0] != allow {
		t.Fatalf("masked regions %+v, want [%+v]", report.Masked, allow)
	}
}

// TestDiffAllowListDoesNotMaskUnlistedPixels pins region scoping: a mismatch
// outside every allow-list entry stays in ColourDiffs, and an entry that
// covered no mismatch is not echoed as masked.
func TestDiffAllowListDoesNotMaskUnlistedPixels(t *testing.T) {
	// given
	a, b := loadTwoFamilies(t, "wide-base", "wide-flipped", 40, 40)
	left := parity.Region{X: 0, Y: 0, Width: 32, Height: 40}

	// when
	report := parity.Diff(a, b, parity.Options{AllowList: []parity.Region{left}})

	// then
	expectColourDiff(t, report, 33, 35)

	if len(report.Masked) != 0 {
		t.Fatalf("masked regions %+v, want none: the left band covers no mismatch", report.Masked)
	}
}

// TestDiffOverlappingAllowListRegions pins the multi-region input: a pixel
// covered by two entries (one listed twice) is masked once, and a region
// covering no mismatch is not echoed as masked.
func TestDiffOverlappingAllowListRegions(t *testing.T) {
	// given
	a, b := loadTwoFamilies(t, "base", "palshift", 4, 2)
	covers := parity.Region{X: 2, Y: 0, Width: 1, Height: 1}
	dormant := parity.Region{X: 0, Y: 1, Width: 1, Height: 1}

	// when
	report := parity.Diff(a, b, parity.Options{AllowList: []parity.Region{
		covers, covers, dormant,
	}})

	// then
	expectExact(t, report)

	if len(report.ColourDiffs) != 0 {
		t.Fatalf("masked pixel still reported: %+v", report.ColourDiffs)
	}

	if len(report.Masked) != 1 || report.Masked[0] != covers {
		t.Fatalf("masked regions %+v, want the single covering region %+v", report.Masked, covers)
	}
}

// loadTwoFamilies loads two committed fixture families at the same size.
func loadTwoFamilies(t *testing.T, left, right string, width, height int) (a, b parity.FrameDump) {
	t.Helper()

	a = mustLoadDump(t, left, width, height)
	b = mustLoadDump(t, right, width, height)

	return a, b
}

// expectExact asserts the report records an exact match.
func expectExact(t *testing.T, report parity.Report) {
	t.Helper()

	if !report.IndexExact {
		t.Fatalf("index compare not exact: %+v", report)
	}

	if report.IndexDiff != nil {
		t.Fatalf("exact match carries a first-diff record: %+v", *report.IndexDiff)
	}
}

// expectFirstDiff asserts the first differing pixel and its owning tile.
func expectFirstDiff(t *testing.T, report parity.Report, x, y int) {
	t.Helper()

	if report.IndexExact {
		t.Fatalf("index compare reported exact, want a diff: %+v", report)
	}

	if report.IndexDiff == nil {
		t.Fatal("index diff carries no first-diff record")
	}

	if report.IndexDiff.X != x || report.IndexDiff.Y != y {
		t.Fatalf("first diff at (%d,%d), want (%d,%d)", report.IndexDiff.X, report.IndexDiff.Y, x, y)
	}

	if got := report.IndexDiff.TileX; got != x/parity.TileSize {
		t.Fatalf("owning tile column %d, want %d", got, x/parity.TileSize)
	}

	if got := report.IndexDiff.TileY; got != y/parity.TileSize {
		t.Fatalf("owning tile row %d, want %d", got, y/parity.TileSize)
	}
}

// expectColourDiff asserts the colour compare found the palette-expanded
// mismatch and no others.
func expectColourDiff(t *testing.T, report parity.Report, x, y int) {
	t.Helper()

	if len(report.ColourDiffs) != 1 {
		t.Fatalf("colour diffs %+v, want exactly one entry", report.ColourDiffs)
	}

	if report.ColourDiffs[0].X != x || report.ColourDiffs[0].Y != y {
		t.Fatalf("colour diff at (%d,%d), want (%d,%d)",
			report.ColourDiffs[0].X, report.ColourDiffs[0].Y, x, y)
	}
}
