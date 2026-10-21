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
		if len(got[y]) != len(want[y]) {
			diffs = append(diffs, fmt.Sprintf("[%d]: cols: got %d, want %d", y, len(got[y]), len(want[y])))
			continue
		}
		for x, g := range got[y] {
			if w := want[y][x]; g != w {
				diffs = append(diffs, fmt.Sprintf("[%d][%d]: got %d, want %d", y, x, g, w))
			}
		}
	}
	return diffs
}
