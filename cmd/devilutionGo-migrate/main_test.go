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

func TestRunSavesWithoutDirIsAUsageError(t *testing.T) {
	// when the now-real saves subcommand runs without its directory
	var stdout, stderr bytes.Buffer

	code := run([]string{"saves"}, &stdout, &stderr)

	// then it is a usage error naming the requirement, not a fake lane
	if code != exitUsage {
		t.Errorf("exit code = %d, want %d", code, exitUsage)
	}

	if !strings.Contains(stderr.String(), "--dir is required") {
		t.Errorf("stderr = %q, want the --dir requirement", stderr.String())
	}
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

func TestTopLevelUsageNamesTheCoveredSaveFamilies(t *testing.T) {
	// given no arguments
	var stdout, stderr bytes.Buffer

	// when the command reports usage
	run(nil, &stdout, &stderr)

	// then the saves synopsis names the .sv families it imports and says
	// what happens to .sp files, instead of advertising an import that
	// classification would silently ignore
	want := "saves     import the legacy .sv saves into SQLite:"
	if !strings.Contains(stderr.String(), want) {
		t.Errorf("stderr = %q, want %q", stderr.String(), want)
	}

	if !strings.Contains(stderr.String(), ".sp") || !strings.Contains(stderr.String(), "untouched") {
		t.Errorf("stderr = %q, want it to say the .sp files stay untouched", stderr.String())
	}

	if strings.Contains(stderr.String(), ".sv/.sp") {
		t.Errorf("stderr = %q, want no .sv/.sp family pairing left", stderr.String())
	}
}

func TestUsageListsExactlyTheThreeTopLevelSubcommands(t *testing.T) {
	// given no arguments
	var stdout, stderr bytes.Buffer

	// when the command reports usage
	run(nil, &stdout, &stderr)

	// then the top level stays assets, saves and settings (ADR-014): the
	// inventory commands are nested under assets, never promoted here
	var listed []string

	for _, line := range strings.Split(stderr.String(), "\n") {
		if !strings.HasPrefix(line, "  ") || strings.HasPrefix(line, "    ") {
			continue
		}

		listed = append(listed, strings.Fields(line)[0])
	}

	want := []string{"assets", "saves", "settings"}
	if strings.Join(listed, " ") != strings.Join(want, " ") {
		t.Fatalf("top-level subcommands = %v, want %v", listed, want)
	}
}
