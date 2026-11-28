package audit_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/audit"
)

const (
	manifestFile = "audit_manifest.json"
	repoRoot     = "../../../.."
)

// TestEveryCommittedTSVIsClassified is the gate: the committed corpora are a
// closed world, every registered table agrees with its file, and the manifest
// expectation is still the measured one.
func TestEveryCommittedTSVIsClassified(t *testing.T) {
	// given
	manifest := mustLoadManifest(t, manifestFile)

	// when
	report, err := audit.Audit(repoRoot, manifestFile)
	// then
	if err != nil {
		t.Fatalf("audit: %v", err)
	}

	if !report.OK() {
		t.Fatalf("data audit is red:\n%s", report.Text())
	}

	assertRegisteredRowsComplete(t, report)
	assertManifestCoversEveryParser(t, manifest, report)
}

// TestUnregisteredTSVIsANamedFailure is the closed-world half of the gate: a
// .tsv the registry does not declare must fail the audit by name, never be
// silently ignored.
func TestUnregisteredTSVIsANamedFailure(t *testing.T) {
	// given
	root := copyCorpora(t)
	writeRepoFile(t, root, audit.CorpusGamedata+"/mystery.tsv", "id\tname\n1\talpha\n")

	// when
	report, err := audit.Audit(root, manifestFile)
	// then
	if err != nil {
		t.Fatalf("audit: %v", err)
	}

	if report.OK() {
		t.Fatalf("an unregistered .tsv passed the audit:\n%s", report.Text())
	}

	assertProblemNames(t, report, "mystery.tsv")
}

// TestParsedRowDriftIsANamedFailure is the other half: the fingerprint must be
// read off the parsed Go row, so a parser whose rows drifted from the file
// fails even though the row count still matches and the file is pristine.
func TestParsedRowDriftIsANamedFailure(t *testing.T) {
	// given
	families := driftMonstdat(t, audit.Families())

	// when
	report, err := audit.Run(repoRoot, manifestFile, families)
	// then
	if err != nil {
		t.Fatalf("audit: %v", err)
	}

	if report.OK() {
		t.Fatalf("a row-shifted parser passed the audit:\n%s", report.Text())
	}

	assertProblemNames(t, report, "monstdat/monstdat.tsv")
}

// TestCorruptedManifestFailsWithTableAndVariant keeps the manifest an
// expectation, not a mirror: a drifted wantRows fails and names its table.
func TestCorruptedManifestFailsWithTableAndVariant(t *testing.T) {
	// given
	manifest := mustLoadManifest(t, manifestFile)
	if len(manifest.Entries) == 0 {
		t.Fatal("manifest has no entries")
	}

	target := manifest.Entries[0]
	manifest.Entries[0].WantRows++

	// when
	report, err := audit.Audit(repoRoot, writeManifest(t, manifest))
	// then
	if err != nil {
		t.Fatalf("audit: %v", err)
	}

	if report.OK() {
		t.Fatal("expected the corrupted manifest to fail the row count check")
	}

	assertProblemNames(t, report, target.Table, target.Variant)
}

// TestReportIsDiffableText pins the shape the `goldens audit` subcommand
// prints: a header, one line per committed file, and the problems at the end.
func TestReportIsDiffableText(t *testing.T) {
	// given
	report, err := audit.Audit(repoRoot, manifestFile)
	if err != nil {
		t.Fatalf("audit: %v", err)
	}

	// when
	text := report.Text()
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")

	// then
	if !strings.HasPrefix(lines[0], "corpus") {
		t.Fatalf("first report line = %q, want the column header", lines[0])
	}

	if len(lines) != 1+len(report.Rows) {
		t.Fatalf("report lines = %d, want %d (header + one per file)", len(lines), 1+len(report.Rows))
	}

	for _, want := range []string{audit.CorpusGamedata, audit.CorpusUpstream} {
		if !strings.Contains(text, want) {
			t.Fatalf("report does not mention corpus %q", want)
		}
	}
}

// mustLoadManifest loads the audit manifest or fails the test.
func mustLoadManifest(t *testing.T, path string) audit.Manifest {
	t.Helper()

	m, err := audit.LoadManifest(path)
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}

	return m
}

// assertRegisteredRowsComplete checks every registered row carries a parsed
// count that matches the file and a manifest expectation.
func assertRegisteredRowsComplete(t *testing.T, report audit.Report) {
	t.Helper()

	for _, row := range report.Rows {
		if row.Status != audit.StatusRegistered {
			continue
		}

		if row.RowsGo != row.RowsTSV {
			t.Errorf("%s/%s: Go rows = %d, file rows = %d", row.Corpus, row.File, row.RowsGo, row.RowsTSV)
		}

		if row.WantRows != row.RowsTSV {
			t.Errorf("%s/%s: manifest wantRows = %d, file rows = %d", row.Corpus, row.File, row.WantRows, row.RowsTSV)
		}
	}
}

// assertManifestCoversEveryParser checks no registered table is left without a
// measured expectation.
func assertManifestCoversEveryParser(t *testing.T, m audit.Manifest, report audit.Report) {
	t.Helper()

	registered := 0
	for _, row := range report.Rows {
		if row.Status == audit.StatusRegistered {
			registered++
		}
	}

	if registered != len(m.Entries) {
		t.Fatalf("registered files = %d, manifest entries = %d", registered, len(m.Entries))
	}
}

// driftMonstdat returns the registry with the monstdat/base parser shifted one
// row behind the file: same count, wrong first and last identity.
func driftMonstdat(t *testing.T, families []audit.Family) []audit.Family {
	t.Helper()

	found := false

	for i := range families {
		if families[i].Table == "monstdat" && families[i].Variant == "base" {
			families[i].Parse = shiftOneRowBehind(families[i].Parse)
			found = true
		}
	}

	if !found {
		t.Fatal("registry has no monstdat/base family to shift")
	}

	return families
}

// shiftOneRowBehind rewrites a parser so it reports the file's rows one behind
// while keeping the row count: the drift only the parsed-row check can catch.
func shiftOneRowBehind(parse audit.Parser) audit.Parser {
	return func(data []byte) ([]string, error) {
		keys, err := parse(data)
		if err != nil || len(keys) < 2 {
			return keys, err
		}

		return append(keys[1:], keys[len(keys)-1]), nil
	}
}

// assertProblemNames fails unless one problem names every want fragment.
func assertProblemNames(t *testing.T, report audit.Report, wants ...string) {
	t.Helper()

	joined := strings.Join(report.Problems, "\n")
	for _, want := range wants {
		if !strings.Contains(joined, want) {
			t.Fatalf("no problem names %q; problems:\n%s", want, joined)
		}
	}
}

// copyCorpora copies both committed corpora under a temp root so a test can
// mutate the world without touching the tree.
func copyCorpora(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	for _, corpus := range []string{audit.CorpusGamedata, audit.CorpusUpstream} {
		dst := filepath.Join(root, corpus)
		if err := os.CopyFS(dst, os.DirFS(filepath.Join(repoRoot, corpus))); err != nil {
			t.Fatalf("copy %s: %v", corpus, err)
		}
	}

	return root
}

// writeRepoFile writes content at a root-relative slash path.
func writeRepoFile(t *testing.T, root, rel, content string) {
	t.Helper()

	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// writeManifest marshals m to a temp file and returns its path.
func writeManifest(t *testing.T, m audit.Manifest) string {
	t.Helper()

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}

	path := filepath.Join(t.TempDir(), "audit_manifest.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	return path
}
