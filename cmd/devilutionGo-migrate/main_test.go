package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWithoutSubcommandPrintsUsage(t *testing.T) {
	// given no arguments
	var stdout, stderr bytes.Buffer

	// when the command runs
	code := run(nil, &stdout, &stderr)

	// then it reports usage on stderr and exits 2
	if code != exitUsage {
		t.Errorf("exit code = %d, want %d", code, exitUsage)
	}

	if !strings.Contains(stderr.String(), "usage: devilutionGo-migrate") {
		t.Errorf("stderr = %q, want the usage text", stderr.String())
	}
}

func TestRunRejectsUnknownSubcommand(t *testing.T) {
	// given a subcommand that does not exist
	var stdout, stderr bytes.Buffer

	// when the command runs
	code := run([]string{"extract"}, &stdout, &stderr)

	// then it prints usage and exits 2
	if code != exitUsage {
		t.Errorf("exit code = %d, want %d", code, exitUsage)
	}

	if !strings.Contains(stderr.String(), "usage: devilutionGo-migrate") {
		t.Errorf("stderr = %q, want the usage text", stderr.String())
	}
}

func TestRunReportsSavesLaneNotImplemented(t *testing.T) {
	// when the saves subcommand runs
	code, stderr := runPendingLane(t, "saves")

	// then it says so and exits 2 instead of pretending to migrate
	assertPendingLane(t, code, stderr, "saves")
}

func TestRunReportsSettingsLaneNotImplemented(t *testing.T) {
	// when the settings subcommand runs
	code, stderr := runPendingLane(t, "settings")

	// then it says so and exits 2 instead of pretending to migrate
	assertPendingLane(t, code, stderr, "settings")
}

func TestRunAssetsWithoutArchivesFails(t *testing.T) {
	// given no --mpq flag and no DG_ASSET_ARCHIVES
	t.Setenv(assetArchivesEnv, "")

	var stdout, stderr bytes.Buffer

	// when a dry run is asked for
	code := run([]string{"assets", "--dry-run", "--out", t.TempDir()}, &stdout, &stderr)

	// then the run explains what is missing and exits 1
	if code != exitFailure {
		t.Errorf("exit code = %d, want %d (stderr: %s)", code, exitFailure, stderr.String())
	}

	if !strings.Contains(stderr.String(), assetArchivesEnv) {
		t.Errorf("stderr = %q, want it to name %s", stderr.String(), assetArchivesEnv)
	}
}

func TestRunAssetsRejectsRemovedArchive(t *testing.T) {
	// given a --mpq path that does not exist
	missing := filepath.Join(t.TempDir(), "diabdat.mpq")
	var stdout, stderr bytes.Buffer

	// when the run starts
	code := run([]string{"assets", "--mpq", missing, "--out", t.TempDir()}, &stdout, &stderr)

	// then it fails before writing anything
	if code != exitFailure {
		t.Errorf("exit code = %d, want %d", code, exitFailure)
	}

	if !strings.Contains(stderr.String(), "diabdat.mpq") {
		t.Errorf("stderr = %q, want it to name the missing archive", stderr.String())
	}
}

func TestRunAssetsDiscoversArchivesFromEnv(t *testing.T) {
	// given a configured dir holding one archive
	archivesDir := t.TempDir()
	archive := writeArchive(t, archivesDir, map[string][]byte{`txtdata\items.tsv`: tsvFixture()})

	t.Setenv(assetArchivesEnv, archivesDir)

	out := filepath.Join(t.TempDir(), "assets-out")

	var stdout, stderr bytes.Buffer

	// when the run needs no --mpq flag
	code := run([]string{"assets", "--out", out}, &stdout, &stderr)

	// then the discovered archive was migrated
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d (stderr: %s)", code, exitOK, stderr.String())
	}

	if listed := readManifest(t, out); len(listed.Archives) != 1 {
		t.Fatalf("manifest archives = %v, want the configured archive", listed.Archives)
	}

	if listed := readManifest(t, out); listed.Archives[0].Path != archive {
		t.Errorf("manifest archive = %q, want %q", listed.Archives[0].Path, archive)
	}
}

func TestRunAssetsRejectsDuplicateArchiveBaseNames(t *testing.T) {
	// given two dirs that each hold spawn.mpq
	first := filepath.Join(t.TempDir(), "first")
	second := filepath.Join(t.TempDir(), "second")
	writeArchive(t, first, map[string][]byte{})
	writeArchive(t, second, map[string][]byte{})

	t.Setenv(assetArchivesEnv, first+string(os.PathListSeparator)+second)

	var stdout, stderr bytes.Buffer

	// when both are migrated in one run
	code := run([]string{"assets", "--out", t.TempDir()}, &stdout, &stderr)

	// then the clash is refused instead of one tree overwriting the other
	if code != exitFailure {
		t.Fatalf("exit code = %d, want %d", code, exitFailure)
	}

	if !strings.Contains(stderr.String(), "share the base name") {
		t.Errorf("stderr = %q, want the base-name clash message", stderr.String())
	}
}

// runPendingLane runs a subcommand whose lane has not landed yet.
func runPendingLane(t *testing.T, name string) (int, string) {
	t.Helper()

	var stdout, stderr bytes.Buffer

	return run([]string{name}, &stdout, &stderr), stderr.String()
}

// assertPendingLane checks the not-implemented report and the exit code.
func assertPendingLane(t *testing.T, code int, stderr, name string) {
	t.Helper()

	if code != exitUsage {
		t.Errorf("exit code = %d, want %d", code, exitUsage)
	}

	if want := name + ": not implemented yet"; !strings.Contains(stderr, want) {
		t.Errorf("stderr = %q, want it to contain %q", stderr, want)
	}
}
