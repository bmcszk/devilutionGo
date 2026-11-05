// Command gamedatagoldens regenerates full-table golden JSON fixtures from
// the committed raw TSV inputs under internal/testdata/gamedata.
//
// Zero-dependency policy: this generator reads only committed TSVs — it never
// builds devilutionX, reads MPQs, or touches the network. Output files use the
// parity.Fixture JSON schema consumed by internal/testsupport/parity.
package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// probeCommit identifies the generator recipe version that produced the golden
// (the refresh path reads committed TSVs, not a C++ probe commit).
const probeCommit = "refresh-recipe-v1"

const goldenDir = "internal/testdata/gamedata/goldens"

// targets lists every table with a committed full-table golden. A new table
// enters coverage by appending here and committing the regenerated file.
var targets = []goldenTarget{
	{"spelldat", "internal/testdata/gamedata/spells/spelldat.tsv", "base"},
	{"misdat", "internal/testdata/gamedata/misdat/misdat.tsv", "base"},
	{"monstdat", "internal/testdata/gamedata/monstdat/monstdat.tsv", "base"},
	{"objdat", "internal/testdata/gamedata/objdat/objdat.tsv", "base"},
	{"textdat", "internal/testdata/gamedata/textdat/textdat.tsv", "base"},
	{"questdat", "internal/testdata/gamedata/quests/questdat.tsv", "base"},
	{"towners", "internal/testdata/gamedata/towners/towners.tsv", "base"},
}

// goldenTarget is one committed TSV table and the golden file it regenerates.
type goldenTarget struct {
	table   string
	tsvPath string // repo-root relative
	variant string
}

// goldenTable pairs the golden JSON payload with the parsed rows/columns the
// run summary reports.
type goldenTable struct {
	fixture map[string]any
	columns []string
	rows    [][]string
}

func main() {
	if err := run(); err != nil {
		warn(err)
		os.Exit(1)
	}
}

// warn reports a fatal run error on stderr; a failed write has no further
// reporting channel.
func warn(err error) {
	_, _ = fmt.Fprintln(os.Stderr, "gamedatagoldens:", err)
}

func run() error {
	for _, target := range targets {
		table, err := buildFixture(target)
		if err != nil {
			return err
		}

		if err := writeFixture(target, table); err != nil {
			return err
		}

		slog.Info("fixture written", "tsv", target.tsvPath,
			"rows", len(table.rows), "columns", len(table.columns))
	}

	return nil
}

// buildFixture reads one committed TSV into the golden JSON schema.
func buildFixture(target goldenTarget) (*goldenTable, error) {
	columns, rows, err := readTSV(target.tsvPath)
	if err != nil {
		return nil, err
	}

	return &goldenTable{
		fixture: map[string]any{
			"table":       target.table,
			"variant":     target.variant,
			"probeCommit": probeCommit,
			"columns":     columns,
			"rows":        rows,
			"sourceTSV":   target.tsvPath,
		},
		columns: columns,
		rows:    rows,
	}, nil
}

func writeFixture(target goldenTarget, table *goldenTable) error {
	data, err := json.MarshalIndent(table.fixture, "", "  ")
	if err != nil {
		return fmt.Errorf("encode %s golden: %w", target.table, err)
	}

	data = append(data, '\n')

	path := filepath.Join(goldenDir, target.table+"-"+target.variant+".json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %q: %w", path, err)
	}

	return nil
}

// readTSV splits a tab-separated table into its header and data rows,
// trimming the CR of CRLF line endings. Blank trailing lines are dropped.
func readTSV(path string) (columns []string, rows [][]string, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("read %q: %w", path, err)
	}

	for line := range strings.SplitSeq(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n") {
		line = strings.TrimSuffix(line, "\r")
		if line == "" {
			continue
		}

		fields := strings.Split(line, "\t")
		if columns == nil {
			columns = fields
			continue
		}

		rows = append(rows, fields)
	}

	if columns == nil {
		return nil, nil, fmt.Errorf("%q has no header row", path)
	}

	return columns, rows, nil
}
