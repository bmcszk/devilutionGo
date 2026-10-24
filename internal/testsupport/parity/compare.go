package parity

import "fmt"

// Exported for the repo-wide parity harness (consumed by port slices; external-test pattern).
// CompareTable reports every difference between the parsed rows and the golden
// fixture as one human-readable line; no lines means the tables agree.
func CompareTable(gotRows [][]string, fixture Fixture) []string {
	diffs := make([]string, 0)
	if len(gotRows) != len(fixture.Rows) {
		diffs = append(diffs, fmt.Sprintf("rows: got %d, want %d", len(gotRows), len(fixture.Rows)))
	}

	comparableRows := min(len(gotRows), len(fixture.Rows))
	for rowIndex := range comparableRows {
		for columnIndex, column := range fixture.Columns {
			got := cellValue(gotRows[rowIndex], columnIndex)
			want := cellValue(fixture.Rows[rowIndex], columnIndex)
			if got != want {
				diffs = append(diffs, fmt.Sprintf("row %d col %s: got %s, want %s", rowIndex, column, got, want))
			}
		}
	}
	return diffs
}

// A row shorter than the column is treated as an empty cell.
func cellValue(row []string, columnIndex int) string {
	if columnIndex >= len(row) {
		return ""
	}
	return row[columnIndex]
}
