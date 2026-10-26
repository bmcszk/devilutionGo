package parity_test

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/testsupport/parity"
)

func TestCompareTable_SpelldatTSVAgreesWithGoldenSlice(t *testing.T) {
	// given
	fixturePath := filepath.Join("..", "..", "testdata", "gamedata", "goldens", "spelldat-base.json")

	fixture, err := parity.LoadFixture(fixturePath)
	if err != nil {
		t.Fatal(err)
	}

	tsvPath := filepath.Join("..", "..", "testdata", "gamedata", "spells", "spelldat.tsv")

	header, gotRows := readTSV(t, tsvPath)
	if !slices.Equal(header, fixture.Columns) {
		t.Fatalf("tsv header = %v, want fixture columns %v", header, fixture.Columns)
	}

	// when
	diffs := parity.CompareTable(gotRows, fixture)

	// then
	if len(gotRows) == len(fixture.Rows) {
		if len(diffs) != 0 {
			t.Fatalf("diffs = %v, want none", diffs)
		}

		return
	}

	want := fmt.Sprintf("rows: got %d, want %d", len(gotRows), len(fixture.Rows))
	if len(diffs) != 1 || diffs[0] != want {
		t.Fatalf("diffs = %v, want exactly [%q]", diffs, want)
	}
}

// TestCompareTable_CorruptedFixtureReportsField: a row-matching-length fixture
// must yield only field diffs, never a row-count diff.
func TestCompareTable_CorruptedFixtureReportsField(t *testing.T) {
	// given
	dir := t.TempDir()
	path := filepath.Join(dir, "spelldat-base.json")

	contents := `{"table":"spelldat","variant":"base","probeCommit":"test","columns":["id"],` +
		`"rows":[["wrong"]],"sourceTSV":"spelldat.tsv"}`
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}

	fixture, err := parity.LoadFixture(path)
	if err != nil {
		t.Fatal(err)
	}

	// when
	diffs := parity.CompareTable([][]string{{"right"}}, fixture)

	// then
	if len(diffs) != 1 || diffs[0] != "row 0 col id: got right, want wrong" {
		t.Fatalf("got diffs %v", diffs)
	}

	if strings.Contains(strings.Join(diffs, "\n"), "rows:") {
		t.Fatalf("got count mismatch in field mismatch: %v", diffs)
	}
}

// readTSV splits a tab-separated table into its header and data rows,
// trimming the CR of CRLF line endings.
func readTSV(t *testing.T, path string) (header []string, rows [][]string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %q: %v", path, err)
	}

	for line := range strings.SplitSeq(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n") {
		line = strings.TrimSuffix(line, "\r")
		if line == "" {
			continue
		}

		fields := strings.Split(line, "\t")
		if header == nil {
			header = fields
			continue
		}

		rows = append(rows, fields)
	}

	if header == nil {
		t.Fatalf("%q has no header row", path)
	}

	return header, rows
}
