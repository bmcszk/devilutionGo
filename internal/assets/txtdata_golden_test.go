package assets_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/assets"
)

const (
	// txtdataGoldens holds the hand-built schema fixtures (see its README).
	txtdataGoldens = "../testdata/assets/txtdata/golden"
	// txtdataUpstream holds the committed devilutionX corpus, one file per
	// upstream table, byte-identical to the pin (see manifest.tsv).
	txtdataUpstream = "../testdata/assets/txtdata/upstream"
	// txtdataManifest holds one row per committed corpus table: its relative
	// path, its data-row count and the sha256 of its bytes. The counts were
	// measured once offline against the upstream checkout; no test recomputes
	// them, so the oracle shares no code with the parser.
	txtdataManifest = "../testdata/assets/txtdata/manifest.tsv"
)

// upstreamTableCount is the number of tables in devilutionX's assets/txtdata
// at the pin; the manifest must hold exactly one row per table.
const upstreamTableCount = 49

// manifestEntry is one row of the committed corpus manifest.
type manifestEntry struct {
	table  string
	rows   int
	sha256 string
}

// TestSyntheticSchemaGoldens pins the row-count contract on the two hand-built
// tables. They exercise the loader's shapes (id strings, enum-name strings,
// flags, numbers) but carry invented values: they are schema fixtures, not
// corpus parities.
func TestSyntheticSchemaGoldens(t *testing.T) {
	for _, tc := range []struct {
		file string
		rows int
	}{
		{"synthetic-monstdat-schema.tsv", 113},
		{"synthetic-experience-schema.tsv", 50},
	} {
		t.Run(tc.file, func(t *testing.T) {
			// given the committed synthetic schema table
			path := filepath.Join(txtdataGoldens, tc.file)

			// when it is loaded
			table := loadTxtdataPath(t, path)

			// then every row parses with the declared column count
			expectRowCount(t, tc.file, table, tc.rows)
		})
	}
}

// TestSyntheticMonstdatColumns pins the accessors on the synthetic monstdat
// schema fixture.
func TestSyntheticMonstdatColumns(t *testing.T) {
	// given the 113-row synthetic monster table
	table := loadTxtdataPath(t, filepath.Join(txtdataGoldens, "synthetic-monstdat-schema.tsv"))

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

// TestUpstreamTablesParseFully drains every committed upstream table through the
// loader and compares the parsed row count with the count measured offline. It
// reads committed bytes only: no C++ checkout, no MPQ, no network, and a
// missing fixture fails rather than skips.
func TestUpstreamTablesParseFully(t *testing.T) {
	// given the committed corpus manifest covering the whole corpus
	entries := readTxtdataManifest(t)

	// when every committed table is parsed and drained
	for _, entry := range entries {
		t.Run(entry.table, func(t *testing.T) {
			// when the fixture bytes load
			raw := readUpstreamTable(t, entry.table)
			table := loadCommittedTable(t, entry.table, raw)

			// then the loader consumed exactly the measured rows
			expectRowCount(t, entry.table, table, entry.rows)
		})
	}

	// then the manifest named every corpus table exactly once
	if len(entries) != upstreamTableCount {
		t.Fatalf("manifest holds %d tables, want %d", len(entries), upstreamTableCount)
	}
}

// TestUpstreamManifestCoversCorpus fails when a committed table has no manifest
// row or a manifest row has no committed table, so dropping either side of the
// golden cannot pass silently.
func TestUpstreamManifestCoversCorpus(t *testing.T) {
	// given the manifest and the committed corpus tree
	manifest := make([]string, 0, upstreamTableCount)
	for _, entry := range readTxtdataManifest(t) {
		manifest = append(manifest, entry.table)
	}
	corpus := listUpstreamTables(t)

	// when both sides are sorted for set comparison
	slices.Sort(manifest)
	slices.Sort(corpus)

	// then they name the same tables
	if !slices.Equal(manifest, corpus) {
		t.Fatalf("manifest tables = %q; committed corpus = %q", manifest, corpus)
	}
}

// parseManifestLine converts one manifest row into an entry.
func parseManifestLine(t *testing.T, lineNo int, line string) manifestEntry {
	t.Helper()

	fields := strings.Split(line, "	")
	if len(fields) != 3 {
		t.Fatalf("manifest line %d has %d fields, want 3: %q", lineNo, len(fields), line)
	}

	var rows int
	if _, err := fmt.Sscanf(fields[1], "%d", &rows); err != nil {
		t.Fatalf("manifest line %d rows %q: %v", lineNo, fields[1], err)
	}

	return manifestEntry{table: fields[0], rows: rows, sha256: fields[2]}
}

// readTxtdataManifest parses the committed manifest. It reads our own LF-only
// TSV with a plain split, deliberately sharing no normalization code with the
// loader under test.
func readTxtdataManifest(t *testing.T) []manifestEntry {
	t.Helper()

	raw, err := os.ReadFile(txtdataManifest)
	if err != nil {
		t.Fatalf("committed manifest %s: %v", txtdataManifest, err)
	}

	lines := strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n")
	if len(lines) < 2 || lines[0] != "table	rows	sha256" {
		t.Fatalf("manifest header = %q, want table/rows/sha256", lines[0])
	}

	entries := make([]manifestEntry, 0, len(lines)-1)
	for i, line := range lines[1:] {
		entries = append(entries, parseManifestLine(t, i+2, line))
	}

	return entries
}

// readUpstreamTable reads one committed corpus table or fails.
func readUpstreamTable(t *testing.T, table string) []byte {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join(txtdataUpstream, filepath.FromSlash(table)))
	if err != nil {
		t.Fatalf("committed fixture for %s: %v", table, err)
	}

	return raw
}

// visitCorpusPath appends one .tsv corpus path to the accumulator.
func visitCorpusTables(t *testing.T, tables *[]string, path string) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if info.IsDir() || !strings.HasSuffix(path, ".tsv") {
		return
	}

	relative, relErr := filepath.Rel(txtdataUpstream, path)
	if relErr != nil {
		t.Fatalf("relative path of %s: %v", path, relErr)
	}

	*tables = append(*tables, filepath.ToSlash(relative))
}

// collectCorpusTables walks the corpus root and yields every .tsv path
// as a root-relative slash path.
func collectCorpusTables(t *testing.T) []string {
	t.Helper()

	var tables []string
	walkErr := filepath.WalkDir(txtdataUpstream, func(path string, _ os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		visitCorpusTables(t, &tables, path)
		return nil
	})
	if walkErr != nil {
		t.Fatalf("walk %s: %v", txtdataUpstream, walkErr)
	}
	return tables
}

// listUpstreamTables guards that the corpus is non-empty and returns it.
func listUpstreamTables(t *testing.T) []string {
	t.Helper()

	// when the corpus tree is walked
	tables := collectCorpusTables(t)

	// then at least one committed table exists
	if len(tables) == 0 {
		t.Fatalf("%s holds no committed tables", txtdataUpstream)
	}

	return tables
}

// sha256Hex is the hex sha256 of raw.
func sha256Hex(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

// expectRowCount asserts the parsed row count and the per-row field count.
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

// loadCommittedTable asserts the pinned digest then loads the table.
func loadCommittedTable(t *testing.T, name string, raw []byte) *assets.Table {
	t.Helper()

	// given the committed bytes
	// when their digest is checked against the manifest pin
	if got := sha256Hex(raw); got != readDigestFor(t, name) {
		t.Fatalf("%s sha256 = %s, want %s", name, got, readDigestFor(t, name))
	}

	// then the loader accepts the bytes
	table, err := assets.Load(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("Load(%s): %v", name, err)
	}

	return table
}

// manifestDigest maps table path to its pinned sha256.
func manifestDigest(t *testing.T) map[string]string {
	t.Helper()

	m := make(map[string]string, upstreamTableCount)
	for _, entry := range readTxtdataManifest(t) {
		m[entry.table] = entry.sha256
	}

	return m
}

// readDigestFor looks up one table's pinned digest.
func readDigestFor(t *testing.T, table string) string {
	t.Helper()

	digest, ok := manifestDigest(t)[table]
	if !ok {
		t.Fatalf("no manifest digest for %s", table)
	}

	return digest
}
