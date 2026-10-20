package parity

import "fmt"

// CompareTable returns count and field differences between parsed rows and a golden fixture.
func CompareTable(gotRows [][]string, fixture Fixture) []string {
	diffs := make([]string, 0)
	if len(gotRows) != len(fixture.Rows) {
		diffs = append(diffs, fmt.Sprintf("rows: got %d, want %d", len(gotRows), len(fixture.Rows)))
	}

	rowCount := min(len(gotRows), len(fixture.Rows))
	for rowIndex := range rowCount {
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

func cellValue(row []string, column int) string {
	if column >= len(row) {
		return ""
	}
	return row[column]
}
