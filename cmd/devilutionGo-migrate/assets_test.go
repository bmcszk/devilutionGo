package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// assetExpectation is one produced file as the manifest must describe it.
type assetExpectation struct {
	family string
	format string
	rows   int
	width  int
	height int
	count  int
}

func TestAssetsRunWritesTreeManifestAndTables(t *testing.T) {
	// given an archive holding one record per asset family
	archive := writeArchive(t, t.TempDir(), migrationArchive(t))
	out := filepath.Join(t.TempDir(), "assets-out")

	var stdout, stderr bytes.Buffer

	// when the assets subcommand migrates it
	code := run([]string{"assets", "--mpq", archive, "--out", out}, &stdout, &stderr)

	// then it succeeds and writes the tree, the tables, the list and the journal
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d (stderr: %s)", code, exitOK, stderr.String())
	}

	t.Run("table becomes csv", func(t *testing.T) { assertTableCSV(t, out) })
	t.Run("records mirror the archive paths", func(t *testing.T) { assertRecordsMirrorPaths(t, out) })
	t.Run("manifest lists each file", func(t *testing.T) { assertManifestEntries(t, out) })
	t.Run("journal records the run", func(t *testing.T) { assertJournalWritten(t, out) })
}

func TestAssetsRerunIsIdempotent(t *testing.T) {
	// given a completed migration
	archive := writeArchive(t, t.TempDir(), migrationArchive(t))
	out := filepath.Join(t.TempDir(), "assets-out")

	var first, stderr bytes.Buffer

	if code := run([]string{"assets", "--mpq", archive, "--out", out}, &first, &stderr); code != exitOK {
		t.Fatalf("first run: exit %d (stderr: %s)", code, stderr.String())
	}

	firstManifest, err := os.ReadFile(filepath.Join(out, manifestName))
	if err != nil {
		t.Fatalf("read first manifest: %v", err)
	}

	firstSummary := readManifest(t, out).Summary

	// when the same command runs again on the same assets dir
	var second bytes.Buffer

	code := run([]string{"assets", "--mpq", archive, "--out", out}, &second, &stderr)

	// then it resumes from the journal and writes nothing
	if code != exitOK {
		t.Fatalf("second run: exit %d (stderr: %s)", code, stderr.String())
	}

	if !strings.Contains(second.String(), "wrote 0 file(s), skipped 7") {
		t.Errorf("second run report = %q, want every file skipped", second.String())
	}

	// then the summary, and therefore the whole list, is identical
	secondManifest, err := os.ReadFile(filepath.Join(out, manifestName))
	if err != nil {
		t.Fatalf("read second manifest: %v", err)
	}

	if secondSummary := readManifest(t, out).Summary; secondSummary != firstSummary {
		t.Errorf("summary = %+v, want %+v", secondSummary, firstSummary)
	}

	if !bytes.Equal(firstManifest, secondManifest) {
		t.Error("manifest changed on a rerun of the same archive")
	}
}

func TestAssetsRerunReExtractsChangedArchive(t *testing.T) {
	// given a completed migration
	dir := t.TempDir()
	archive := writeArchive(t, dir, migrationArchive(t))
	out := filepath.Join(t.TempDir(), "assets-out")

	var stdout, stderr bytes.Buffer

	if code := run([]string{"assets", "--mpq", archive, "--out", out}, &stdout, &stderr); code != exitOK {
		t.Fatalf("first run: exit %d (stderr: %s)", code, stderr.String())
	}

	// when one output file is deleted and the archive is replaced by one with a
	// different record
	if err := os.Remove(outputPath(out, "spawn/data/health.clx")); err != nil {
		t.Fatalf("remove output: %v", err)
	}

	replaced := writeArchive(t, dir, map[string][]byte{`data\health.clx`: clxFixture(t)})

	// then the journal no longer describes the archive and it is extracted again
	if code := run([]string{"assets", "--mpq", replaced, "--out", out}, &stdout, &stderr); code != exitOK {
		t.Fatalf("second run: exit %d (stderr: %s)", code, stderr.String())
	}

	if _, err := os.Stat(outputPath(out, "spawn/data/health.clx")); err != nil {
		t.Errorf("missing re-extracted output: %v", err)
	}
}

func TestAssetsDryRunWritesNothing(t *testing.T) {
	// given an archive and an out dir that does not exist yet
	archive := writeArchive(t, t.TempDir(), migrationArchive(t))
	out := filepath.Join(t.TempDir(), "assets-out")

	var stdout, stderr bytes.Buffer

	// when a dry run is asked for
	code := run([]string{"assets", "--mpq", archive, "--out", out, "--dry-run"}, &stdout, &stderr)

	// then it reports the plan without touching the disk
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d (stderr: %s)", code, exitOK, stderr.String())
	}

	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Errorf("out dir exists after a dry run: %v", err)
	}

	if !strings.Contains(stdout.String(), "planned 7 file(s)") {
		t.Errorf("report = %q, want the planned count", stdout.String())
	}

	if !strings.Contains(stdout.String(), "summary sha256 ") {
		t.Errorf("report = %q, want the checksum summary", stdout.String())
	}
}

func TestAssetsRunReportsDecodeFailuresAndExitsOne(t *testing.T) {
	// given an archive whose palette record is truncated
	archive := writeArchive(t, t.TempDir(), map[string][]byte{
		`ui_art\diablo.pal`: palFixture(t)[:100],
		`ui_art\title.pcx`:  pcxFixture(t),
	})
	out := filepath.Join(t.TempDir(), "assets-out")

	var stdout, stderr bytes.Buffer

	// when the migration runs
	code := run([]string{"assets", "--mpq", archive, "--out", out}, &stdout, &stderr)

	// then the failure is named on stderr and the run exits 1
	if code != exitFailure {
		t.Fatalf("exit code = %d, want %d", code, exitFailure)
	}

	if !strings.Contains(stderr.String(), `ui_art\diablo.pal`) {
		t.Errorf("stderr = %q, want it to name the failing record", stderr.String())
	}

	// then the healthy record was still migrated and is the only one listed
	if _, err := os.Stat(outputPath(out, "spawn/ui_art/title.pcx")); err != nil {
		t.Errorf("healthy record was not written: %v", err)
	}

	if listed := readManifest(t, out); len(listed.Files) != 1 {
		t.Errorf("manifest files = %d, want only the healthy record", len(listed.Files))
	}
}

func TestAssetsRunSkipsUnaddressableArchive(t *testing.T) {
	// given an archive that ships no (listfile) and has no committed inventory,
	// so none of its records can be named (MPQ decrypts by name)
	archive := writeArchive(t, t.TempDir(), map[string][]byte{})
	out := filepath.Join(t.TempDir(), "assets-out")

	var stdout, stderr bytes.Buffer

	// when the migration runs
	code := run([]string{"assets", "--mpq", archive, "--out", out}, &stdout, &stderr)

	// then the archive is reported and left alone instead of failing the run
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d (stderr: %s)", code, exitOK, stderr.String())
	}

	if !strings.Contains(stderr.String(), "skipped archive spawn.mpq") {
		t.Errorf("stderr = %q, want the skipped-archive report", stderr.String())
	}

	if listed := readManifest(t, out); len(listed.Files) != 0 {
		t.Errorf("manifest files = %d, want none", len(listed.Files))
	}
}

func TestAssetsRunRejectsTruncatedTable(t *testing.T) {
	// given a .tsv whose data row has fewer fields than the header
	archive := writeArchive(t, t.TempDir(), map[string][]byte{
		`txtdata\items.tsv`: []byte("id\tname\n0\n"),
	})

	var stdout, stderr bytes.Buffer

	// when the migration runs
	code := run([]string{"assets", "--mpq", archive, "--out", filepath.Join(t.TempDir(), "out")},
		&stdout, &stderr)

	// then the ragged table is a failure, not a silently truncated CSV
	if code != exitFailure {
		t.Fatalf("exit code = %d, want %d", code, exitFailure)
	}

	if !strings.Contains(stderr.String(), "fields") {
		t.Errorf("stderr = %q, want the field-count failure", stderr.String())
	}
}

// migrationArchive is the synthetic archive the run tests migrate: one record per
// family, all payloads from the committed gfx goldens.
func migrationArchive(t *testing.T) map[string][]byte {
	t.Helper()

	return map[string][]byte{
		`txtdata\items.tsv`:      tsvFixture(),
		`ui_art\diablo.pal`:      palFixture(t),
		`gendata\pause.trn`:      trnFixture(t),
		`data\health.clx`:        clxFixture(t),
		`ui_art\title.pcx`:       pcxFixture(t),
		`fonts\12-05.clx`:        fontSheetFixture(t),
		`sfx\items\titlemov.wav`: []byte("RIFF----WAVEfmt "),
	}
}

// outputPath returns the mirrored output path inside out.
func outputPath(out, name string) string { return filepath.Join(out, filepath.FromSlash(name)) }

// assertTableCSV checks that the archived .tsv became a CSV with the same values.
func assertTableCSV(t *testing.T, out string) {
	t.Helper()

	content := string(readOutput(t, out, "spawn/txtdata/items.csv"))
	if want := "id,name,level\n0,Zombie,1\n1,Skeleton,2\n"; content != want {
		t.Errorf("csv = %q, want %q", content, want)
	}
}

// assertRecordsMirrorPaths checks that every non-table record was written
// byte-identically below its archive's tree.
func assertRecordsMirrorPaths(t *testing.T, out string) {
	t.Helper()

	expected := map[string][]byte{
		"spawn/ui_art/diablo.pal":      palFixture(t),
		"spawn/gendata/pause.trn":      trnFixture(t),
		"spawn/data/health.clx":        clxFixture(t),
		"spawn/ui_art/title.pcx":       pcxFixture(t),
		"spawn/fonts/12-05.clx":        fontSheetFixture(t),
		"spawn/sfx/items/titlemov.wav": []byte("RIFF----WAVEfmt "),
	}

	for name, want := range expected {
		if got := readOutput(t, out, name); !bytes.Equal(got, want) {
			t.Errorf("%s: %d bytes, want the extracted record (%d bytes)", name, len(got), len(want))
		}
	}
}

// assertManifestEntries checks the family, format, frame layout, row count and
// hash the manifest records for every produced file.
func assertManifestEntries(t *testing.T, out string) {
	t.Helper()

	listed := readManifest(t, out)

	if listed.Tool != manifestTool || listed.Version != manifestVersion {
		t.Fatalf("manifest header = %s/%d, want %s/%d", listed.Tool, listed.Version, manifestTool, manifestVersion)
	}

	expected := map[string]assetExpectation{
		"spawn/txtdata/items.csv":      {family: familyTable, format: "csv", rows: 2},
		"spawn/ui_art/diablo.pal":      {family: familyPalette, format: "pal"},
		"spawn/gendata/pause.trn":      {family: familyTRN, format: "trn"},
		"spawn/data/health.clx":        {family: familySprite, format: "clx", width: 264, height: 19, count: 1},
		"spawn/ui_art/title.pcx":       {family: familySprite, format: "pcx", width: 4, height: 3, count: 1},
		"spawn/fonts/12-05.clx":        {family: familyFont, format: "clx", count: 256},
		"spawn/sfx/items/titlemov.wav": {family: familyRaw, format: "wav"},
	}

	if len(listed.Files) != len(expected) {
		t.Fatalf("manifest lists %d file(s), want %d", len(listed.Files), len(expected))
	}

	for _, entry := range listed.Files {
		assertManifestEntry(t, out, entry, expected[entry.Path])
	}

	if listed.Summary.Files != len(expected) || listed.Summary.SHA256 == "" {
		t.Errorf("summary = %+v, want every file and a checksum", listed.Summary)
	}
}

// assertManifestEntry checks one manifest row against its expectation.
func assertManifestEntry(t *testing.T, out string, entry fileEntry, want assetExpectation) {
	t.Helper()

	if want.family == "" {
		t.Errorf("%s: unexpected file in the manifest", entry.Path)

		return
	}

	if entry.Family != want.family || entry.Format != want.format {
		t.Errorf("%s: %s/%s, want %s/%s", entry.Path, entry.Family, entry.Format, want.family, want.format)
	}

	if entry.Rows != want.rows {
		t.Errorf("%s: rows = %d, want %d", entry.Path, entry.Rows, want.rows)
	}

	if entry.SHA256 != digest(readOutput(t, out, entry.Path)) {
		t.Errorf("%s: sha256 does not match the written file", entry.Path)
	}

	assertFrameLayout(t, entry, want)
}

// assertFrameLayout checks the sprite geometry the manifest records.
func assertFrameLayout(t *testing.T, entry fileEntry, want assetExpectation) {
	t.Helper()

	if want.count == 0 {
		return
	}

	if entry.Frames == nil {
		t.Fatalf("%s: no frame layout, want %d frame(s)", entry.Path, want.count)
	}

	if entry.Frames.Count != want.count {
		t.Errorf("%s: frames = %d, want %d", entry.Path, entry.Frames.Count, want.count)
	}

	if want.width > 0 && (entry.Frames.Width != want.width || entry.Frames.Height != want.height) {
		t.Errorf("%s: frame size = %dx%d, want %dx%d", entry.Path, entry.Frames.Width, entry.Frames.Height,
			want.width, want.height)
	}
}

// assertJournalWritten checks that the run checkpointed its work.
func assertJournalWritten(t *testing.T, out string) {
	t.Helper()

	if _, err := os.Stat(filepath.Join(out, journalName)); err != nil {
		t.Errorf("journal: %v", err)
	}
}
