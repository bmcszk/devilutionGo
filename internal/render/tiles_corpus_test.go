package render_test

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

const (
	// tilesCorpusDir is the tiles fixture tree the render tests consume.
	tilesCorpusDir = "../testdata/render/tiles"
	// tilesManifestName is the corpus' sha256 manifest, the provenance-check
	// pattern of internal/testdata/drlg (README.md:131).
	tilesManifestName = "manifest.txt"
)

// tilesManifestRow is one manifest line: the digest and the corpus-relative
// path it pins.
type tilesManifestRow struct {
	name   string
	digest string
}

// TestTilesFixtureManifest pins every committed tiles fixture: the manifest
// names exactly the corpus files and each row's sha256 matches the bytes on
// disk, so a hand edit or silent regeneration fails the suite.
func TestTilesFixtureManifest(t *testing.T) {
	// given the manifest rows and the committed corpus tree
	rows := readTilesManifest(t)
	corpus := tilesCorpusFiles(t)

	// when both sides are sorted for set comparison
	names := make([]string, 0, len(rows))
	for _, row := range rows {
		names = append(names, row.name)
	}
	slices.Sort(names)
	slices.Sort(corpus)

	// then they name the same files and every digest matches the bytes
	if !slices.Equal(names, corpus) {
		t.Fatalf("manifest files = %q; committed corpus = %q", names, corpus)
	}

	for _, row := range rows {
		expectTilesDigest(t, row)
	}
}

// readTilesManifest parses the committed manifest, skipping comment lines.
func readTilesManifest(t *testing.T) []tilesManifestRow {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join(tilesCorpusDir, tilesManifestName))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}

	rows := make([]tilesManifestRow, 0)
	for i, line := range strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		rows = append(rows, parseTilesManifestLine(t, i+1, line))
	}

	if len(rows) == 0 {
		t.Fatalf("manifest %s holds no rows", tilesManifestName)
	}

	return rows
}

// parseTilesManifestLine splits one `<sha256>  <relpath>` row.
func parseTilesManifestLine(t *testing.T, lineNo int, line string) tilesManifestRow {
	t.Helper()

	digest, name, found := strings.Cut(line, "  ")
	if !found || strings.ContainsAny(name, " \t") {
		t.Fatalf("manifest line %d is not <sha256>  <relpath>: %q", lineNo, line)
	}

	return tilesManifestRow{name: filepath.ToSlash(name), digest: digest}
}

// tilesCorpusFiles lists the committed fixture paths, manifest excluded,
// slash-separated for the set comparison; every fixture lives one level down.
func tilesCorpusFiles(t *testing.T) []string {
	t.Helper()

	entries, err := os.ReadDir(tilesCorpusDir)
	if err != nil {
		t.Fatalf("read corpus: %v", err)
	}

	names := make([]string, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		names = append(names, corpusFilesOf(t, entry.Name())...)
	}

	return names
}

// corpusFilesOf lists one corpus subdirectory's files as <dir>/<name>.
func corpusFilesOf(t *testing.T, dir string) []string {
	t.Helper()

	entries, err := os.ReadDir(filepath.Join(tilesCorpusDir, dir))
	if err != nil {
		t.Fatalf("read corpus %s: %v", dir, err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, dir+"/"+entry.Name())
	}

	return names
}

// expectTilesDigest recomputes one fixture's sha256 and compares it with the
// manifest row.
func expectTilesDigest(t *testing.T, row tilesManifestRow) {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join(tilesCorpusDir, filepath.FromSlash(row.name)))
	if err != nil {
		t.Fatalf("read fixture %s: %v", row.name, err)
	}

	sum := sha256.Sum256(raw)
	if got := hex.EncodeToString(sum[:]); got != row.digest {
		t.Fatalf("%s sha256 = %s, manifest pins %s", row.name, got, row.digest)
	}
}
