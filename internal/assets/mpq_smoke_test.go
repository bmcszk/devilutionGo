package assets_test

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	mpq "github.com/suprsokr/go-mpq"
)

// assetArchiveNames are the retail archives the extraction smoke test must
// decode exhaustively, in the order the game stacks them.
var assetArchiveNames = []string{"DIABDAT.MPQ", "hellfire.mpq", "devilutionx.mpq"}

// TestRealArchivesDecodeEveryRecord walks every (listfile) record of every
// retail archive and asserts go-mpq decodes all of them. The archives are
// environment inputs, never committed fixtures, so an absent archive skips
// instead of failing — the same politeness DevilutionX uses for its
// real-asset tests (test/timedemo_test.cpp:47-49).
func TestRealArchivesDecodeEveryRecord(t *testing.T) {
	// given the retail archives reachable from DG_ASSET_ARCHIVES
	archives := findAssetArchives(t)

	// when every record of every archive is decoded
	failures := decodeAllRecords(t, archives)

	// then no record fails
	if len(failures) > 0 {
		t.Fatalf("%d record(s) failed to decode:\n%s", len(failures), strings.Join(failures, "\n"))
	}
}

// findAssetArchives locates the retail archives under the directories listed in
// DG_ASSET_ARCHIVES (os.PathListSeparator separated) and skips the test when it
// finds none.
func findAssetArchives(t *testing.T) []string {
	t.Helper()

	dirs := filepath.SplitList(os.Getenv("DG_ASSET_ARCHIVES"))
	found := make([]string, 0, len(assetArchiveNames))

	for _, dir := range dirs {
		found = append(found, archivesInDir(dir)...)
	}

	if len(found) == 0 {
		t.Skip("no retail MPQ found; set DG_ASSET_ARCHIVES to the Diablo install directories")
	}

	return found
}

// archivesInDir returns the known archive names that dir holds, spelling them
// with the case the file system reports so the retail uppercase naming
// resolves.
func archivesInDir(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	found := make([]string, 0, len(assetArchiveNames))

	for _, entry := range entries {
		if !entry.IsDir() && isAssetArchiveName(entry.Name()) {
			found = append(found, filepath.Join(dir, entry.Name()))
		}
	}

	slices.Sort(found)

	return found
}

// isAssetArchiveName reports whether name matches one of the retail archives,
// ignoring case: the retail archive is spelled uppercase, the add-ons lowercase.
func isAssetArchiveName(name string) bool {
	for _, archive := range assetArchiveNames {
		if strings.EqualFold(name, archive) {
			return true
		}
	}

	return false
}

// decodeAllRecords extracts every record of every archive into one scratch tree
// and returns the failures as "<archive> <record>: <error>" lines, sorted so the
// report is reproducible.
func decodeAllRecords(t *testing.T, archives []string) []string {
	t.Helper()

	scratch := t.TempDir()
	failures := make([]string, 0, len(archives))

	for _, path := range archives {
		failures = append(failures, decodeArchive(t, path, scratch)...)
	}

	slices.Sort(failures)

	return failures
}

// decodeArchive extracts every record of one archive and reports the failures
// and the record count.
func decodeArchive(t *testing.T, path, scratch string) []string {
	t.Helper()

	archive, err := mpq.Open(path)
	if err != nil {
		return []string{fmt.Sprintf("%s: open: %v", filepath.Base(path), err)}
	}
	defer func() {
		if closeErr := archive.Close(); closeErr != nil {
			t.Errorf("close %s: %v", path, closeErr)
		}
	}()

	records, err := archive.ListFiles()
	if err != nil {
		return []string{fmt.Sprintf("%s: listfile: %v", filepath.Base(path), err)}
	}

	dest := filepath.Join(scratch, strings.ReplaceAll(filepath.Base(path), ".", "-"))
	failures := make([]string, 0)

	for index, record := range records {
		target := filepath.Join(dest, fmt.Sprintf("%06d", index))

		if err := archive.ExtractFile(record, target); err != nil {
			failures = append(failures, fmt.Sprintf("%s %s: %v", filepath.Base(path), record, err))
		}
	}

	t.Logf("%s: %d record(s), %d failure(s)", filepath.Base(path), len(records), len(failures))

	return failures
}
