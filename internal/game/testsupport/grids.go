// Package testsupport holds the helpers shared by the parity tests: the
// golden-grid diff and the canonical state hash.
package testsupport

import "fmt"

// CompareGrids returns one entry per differing cell ("[row][col]: got <got>,
// want <want>") plus a dim-mismatch entry when the shapes differ; an empty
// result means the grids match.
func CompareGrids(got, want [][]uint8) []string {
	var diffs []string

	if len(got) != len(want) {
		return []string{fmt.Sprintf("rows: got %d, want %d", len(got), len(want))}
	}

	for y := range got {
		diffs = append(diffs, compareGridRow(y, got[y], want[y])...)
	}

	return diffs
}

// compareGridRow appends the diffs of one grid row: a col-mismatch entry
// when the row shapes differ, else one entry per differing cell.
// Mirrors the row loop in CompareGrids (game testsupport).
func compareGridRow(y int, got, want []uint8) []string {
	if len(got) != len(want) {
		return []string{fmt.Sprintf("[%d]: cols: got %d, want %d", y, len(got), len(want))}
	}

	var diffs []string

	for x, g := range got {
		if w := want[x]; g != w {
			diffs = append(diffs, fmt.Sprintf("[%d][%d]: got %d, want %d", y, x, g, w))
		}
	}

	return diffs
}
