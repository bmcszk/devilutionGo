// Package testsupport: the golden diff reporter shared by parity tests
// (devilGo-xcdz).
package testsupport

import "fmt"

// CompareGrids reports every cell where got differs from want, as
// "[row][col]: got <got>, want <want>" entries, plus a dim-mismatch entry
// when the shapes differ. Empty result = grids match.
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
