package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestAssetsVerifyReportsMissingEntriesWithTheirLoader(t *testing.T) {
	// given one inventory entry present on disk and one absent
	root := t.TempDir()
	writeFixtureFile(t, root, "data/health.clx")
	inventoryPath := writeInventory(t, root, fixtureInventory())

	// when verify runs under that root
	code, stdout, stderr := runMigrate(t, "assets", "verify", "--inventory", inventoryPath, "--root", root)

	// then it names the miss with its loader and exits 1
	if code != exitFailure {
		t.Fatalf("exit code = %d, want %d (stderr %q)", code, exitFailure, stderr)
	}

	want := "missing fonts/smallfont.trn (requested by Source/engine/assets.cpp:410)"
	if !strings.Contains(stdout, want) {
		t.Fatalf("stdout = %q, want it to contain %q", stdout, want)
	}

	if !strings.Contains(stderr, "1 of 2 entries") {
		t.Fatalf("stderr = %q, want the missing count", stderr)
	}
}

func TestAssetsVerifyPassesWhenEveryEntryResolves(t *testing.T) {
	// given an inventory whose entries are all on disk and no archives anywhere
	t.Setenv(assetArchivesEnv, "")

	root := t.TempDir()
	for _, path := range []string{"data/health.clx", "fonts/smallfont.trn"} {
		writeFixtureFile(t, root, path)
	}

	inventoryPath := writeInventory(t, root, fixtureInventory())

	// when verify runs without an archive in sight
	code, stdout, stderr := runMigrate(t, "assets", "verify", "--inventory", inventoryPath, "--root", root)

	// then full coverage passes, proving verify needs no MPQ
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d (stderr %q)", code, exitOK, stderr)
	}

	if !strings.Contains(stdout, "2 entries in 2 families") {
		t.Fatalf("stdout = %q, want the coverage summary", stdout)
	}
}

func TestAssetsVerifyRejectsInvalidInventory(t *testing.T) {
	// given an inventory entry with a format no loader accepts
	broken := fixtureInventory()
	broken.Entries[0].Format = "gif"
	root := t.TempDir()
	inventoryPath := writeInventory(t, root, broken)

	// when verify runs
	code, _, stderr := runMigrate(t, "assets", "verify", "--inventory", inventoryPath, "--root", root)

	// then the format violation fails the run instead of counting as a miss
	if code != exitFailure {
		t.Fatalf("exit code = %d, want %d", code, exitFailure)
	}

	if !strings.Contains(stderr, `unknown format "gif"`) {
		t.Fatalf("stderr = %q, want the format violation", stderr)
	}
}

func TestAssetsVerifyDefaultsToTheGraphicsInventory(t *testing.T) {
	// given a working directory whose assets/inventory.yaml is the graphics list
	dir := t.TempDir()
	t.Chdir(dir)

	for _, path := range []string{"data/health.clx", "fonts/smallfont.trn"} {
		writeFixtureFile(t, filepath.Join(dir, "assets/loose"), path)
	}

	writeInventory(t, filepath.Join(dir, "assets"), fixtureInventory())

	// when verify runs without --inventory or --root
	code, stdout, stderr := runMigrate(t, "assets", "verify")

	// then the assets/inventory.yaml default resolves and the check passes
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d (stderr %q)", code, exitOK, stderr)
	}

	if !strings.Contains(stdout, "2 entries in 2 families") {
		t.Fatalf("stdout = %q, want the coverage summary", stdout)
	}
}

func TestAssetsVerifyAcceptsTheRootAsAPositionalArgument(t *testing.T) {
	// given an inventory whose entries all sit under a root given positionally
	root := t.TempDir()
	for _, path := range []string{"data/health.clx", "fonts/smallfont.trn"} {
		writeFixtureFile(t, root, path)
	}

	inventoryPath := writeInventory(t, t.TempDir(), fixtureInventory())

	// when verify takes the root as its argument instead of --root
	code, stdout, stderr := runMigrate(t, "assets", "verify", "--inventory", inventoryPath, root)

	// then the documented `assets verify <root>` form covers the inventory
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d (stderr %q)", code, exitOK, stderr)
	}

	if !strings.Contains(stdout, "2 entries in 2 families") {
		t.Fatalf("stdout = %q, want the coverage summary", stdout)
	}
}
