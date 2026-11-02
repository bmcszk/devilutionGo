package assets_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/assets"
)

const txtdataGoldens = "../testdata/assets/txtdata/golden"

// upstreamTxtdataCandidates are the relative roots a developer checkout of
// devilutionX may have. The live tests skip when none exists: no committed test
// may require the C++ tree.
var upstreamTxtdataCandidates = []string{
	"../../../devilutionX/assets/txtdata",
	"../../../../devilutionX/assets/txtdata",
}

func TestGoldenRowCounts(t *testing.T) {
	for _, tc := range []struct {
		file string
		rows int
	}{
		{"monstdat.tsv", 113},
		{"Experience.tsv", 50},
	} {
		t.Run(tc.file, func(t *testing.T) {
			// given the committed golden table
			path := filepath.Join(txtdataGoldens, tc.file)

			// when it is loaded
			table := loadTxtdataPath(t, path)

			// then the row count and column count match the header contract
			expectRowCount(t, tc.file, table, tc.rows)
		})
	}
}

func TestGoldenMonstdatColumns(t *testing.T) {
	// given the 113-row monster golden
	table := loadTxtdataPath(t, filepath.Join(txtdataGoldens, "monstdat.tsv"))

	// when the first and last rows are read by column name
	first, firstErr := table.Rows[0].String("_monster_id")
	last, lastErr := table.Rows[table.Len()-1].String("_monster_id")
	level, levelErr := table.Rows[0].Int("level")

	// then the id column is addressable and the numeric accessor converts
	if firstErr != nil || lastErr != nil || levelErr != nil {
		t.Fatalf("column errors = (%v, %v, %v)", firstErr, lastErr, levelErr)
	}
	if first != "MT_N000" {
		t.Fatalf("first id = %q, want MT_N000", first)
	}
	if last != "MT_N112" {
		t.Fatalf("last id = %q, want MT_N112", last)
	}
	if level != 1 {
		t.Fatalf("first level = %d, want 1", level)
	}
}

func TestUpstreamTablesParseFully(t *testing.T) {
	// given an optional devilutionX checkout
	root := findUpstreamTxtdata(t)

	// when every table under it is loaded
	paths := findTsvFiles(t, root)

	// then the loader consumes every non-empty line exactly once
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			expectParseWithoutLoss(t, path)
		})
	}
}

func expectRowCount(t *testing.T, name string, table *assets.Table, want int) {
	t.Helper()

	if got := table.Len(); got != want {
		t.Fatalf("%s row count = %d, want %d", name, got, want)
	}
	for i, row := range table.Rows {
		if len(row.Fields) != len(table.Header) {
			t.Fatalf("row %d has %d fields, header declares %d", i, len(row.Fields), len(table.Header))
		}
	}
}

func expectParseWithoutLoss(t *testing.T, path string) {
	t.Helper()

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	table, err := assets.Load(strings.NewReader(string(raw)))
	if err != nil {
		t.Fatalf("Load(%s): %v", path, err)
	}

	want := len(nonEmptyLines(raw)) - 1
	if got := table.Len(); got != want {
		t.Fatalf("%s parsed %d rows, file holds %d data lines", path, got, want)
	}
}

// nonEmptyLines counts the lines that survive normalization.
func nonEmptyLines(raw []byte) []string {
	text := strings.ReplaceAll(string(raw), "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	var lines []string
	for _, line := range strings.Split(text, "\n") {
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

// findUpstreamTxtdata returns the first candidate root that exists, or skips.
func findUpstreamTxtdata(t *testing.T) string {
	t.Helper()

	for _, candidate := range upstreamTxtdataCandidates {
		info, err := os.Stat(candidate)
		if err == nil && info.IsDir() {
			return candidate
		}
	}
	t.Skip("no devilutionX checkout next to this worktree")
	return ""
}

// findTsvFiles collects every .tsv under root.
func findTsvFiles(t *testing.T, root string) []string {
	t.Helper()

	var paths []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".tsv") {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	if len(paths) == 0 {
		t.Fatalf("%s holds no .tsv files", root)
	}
	return paths
}
