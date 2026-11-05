package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestAssetsConversionRerunIsIdempotent(t *testing.T) {
	// given one converting pass over a synthetic archive
	archive := writeArchive(t, t.TempDir(), map[string][]byte{
		`data\talkbutton.cel`: celFixture(t),
		`data\health.cl2`:     cl2Fixture(t),
	})
	inventoryPath := writeInventory(t, t.TempDir(), conversionInventory())
	out := filepath.Join(t.TempDir(), "assets-out")

	var first, stderr bytes.Buffer

	firstCode := run([]string{"assets", "--mpq", archive, "--inventory", inventoryPath, "--out", out},
		&first, &stderr)
	if firstCode != exitOK {
		t.Fatalf("first run: exit %d (stderr: %s)", firstCode, stderr.String())
	}

	firstSummary := readManifest(t, out).Summary.SHA256
	firstFiles := writtenDigests(t, out)

	// when the same command runs again on the same assets dir
	var second bytes.Buffer

	code := run([]string{"assets", "--mpq", archive, "--inventory", inventoryPath, "--out", out},
		&second, &stderr)

	// then the journal lets it skip the work and the output is byte-identical
	if code != exitOK {
		t.Fatalf("second run: exit %d (stderr: %s)", code, stderr.String())
	}

	if summary := readManifest(t, out).Summary.SHA256; summary != firstSummary {
		t.Errorf("summary = %s, want the first run's %s", summary, firstSummary)
	}

	if files := writtenDigests(t, out); !slices.Equal(files, firstFiles) {
		t.Errorf("second run dir = %v, want %v", files, firstFiles)
	}

	if !strings.Contains(second.String(), "skipped 2 already-journaled file(s)") {
		t.Errorf("stdout = %q, want the resumed run to skip both records", second.String())
	}
}

func TestAssetsConvertedDirHoldsExactlyTheManifestFiles(t *testing.T) {
	// given a converting pass, where the build embeds the assets dir (ADR-005)
	archive := writeArchive(t, t.TempDir(), map[string][]byte{
		`data\talkbutton.cel`: celFixture(t),
		`data\health.cl2`:     cl2Fixture(t),
	})
	inventoryPath := writeInventory(t, t.TempDir(), conversionInventory())
	out := filepath.Join(t.TempDir(), "assets-out")

	code, _, stderr := runMigrate(t, "assets", "--mpq", archive, "--inventory", inventoryPath, "--out", out)
	if code != exitOK {
		t.Fatalf("run: exit %d (stderr: %s)", code, stderr)
	}

	// when the embedded dir is walked
	want := append(manifestPaths(t, out), manifestName, journalName)
	slices.Sort(want)

	// then it holds the manifest's files plus the two bookkeeping files and no
	// legacy leftovers, so what the build embeds is what the tool wrote
	if got := relativeFiles(t, out); !slices.Equal(got, want) {
		t.Errorf("assets dir = %v, want %v", got, want)
	}
}

// assertOutputBytes checks one written output against its golden bytes.
func assertOutputBytes(t *testing.T, out, name string, want []byte) {
	t.Helper()

	if got := readOutput(t, out, name); !bytes.Equal(got, want) {
		t.Errorf("%s: %d bytes, want %d", name, len(got), len(want))
	}
}

// assertConvertedEntry checks one manifest row records the conversion that made
// it, hashing the inputs the converter was fed.
func assertConvertedEntry(t *testing.T, out, path string, want conversionRecord) {
	t.Helper()

	entry := manifestEntry(t, out, path)
	if entry.Format != "clx" || entry.Conversion == nil {
		t.Fatalf("%s: %s/%+v, want a converted clx record", path, entry.Format, entry.Conversion)

		return
	}

	got := *entry.Conversion
	digestOfWant := want
	digestOfWant.SHA256 = got.SHA256

	if got.Format != want.Format || got.Frames != want.Frames || !slices.Equal(got.Widths, want.Widths) {
		t.Errorf("%s: conversion = %+v, want %+v", path, got, want)
	}

	if got.SHA256 == "" || got.SHA256 != (spriteGeometry{widths: want.Widths, frames: want.Frames}).
		digest(want.Format) {
		t.Errorf("%s: conversion sha256 = %q, want the input digest", path, got.SHA256)
	}
}

// manifestEntry returns the manifest row for path, failing when there is none.
func manifestEntry(t *testing.T, out, path string) fileEntry {
	t.Helper()

	for _, entry := range readManifest(t, out).Files {
		if entry.Path == path {
			return entry
		}
	}

	t.Fatalf("%s is not in the manifest", path)

	return fileEntry{}
}

// manifestPaths returns the manifest's file paths, sorted.
func manifestPaths(t *testing.T, out string) []string {
	t.Helper()

	paths := make([]string, 0, len(readManifest(t, out).Files))
	for _, entry := range readManifest(t, out).Files {
		paths = append(paths, entry.Path)
	}

	slices.Sort(paths)

	return paths
}

// writtenDigests renders the whole assets dir as sorted path/hash lines, so two
// runs can be compared without keeping the bytes around.
func writtenDigests(t *testing.T, out string) []string {
	t.Helper()

	paths := relativeFiles(t, out)
	lines := make([]string, len(paths))

	for i, path := range paths {
		content, err := os.ReadFile(filepath.Join(out, filepath.FromSlash(path)))
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}

		lines[i] = path + " " + digest(content)
	}

	return lines
}

// relativeFiles lists every file below out as a sorted forward-slash path.
func relativeFiles(t *testing.T, out string) []string {
	t.Helper()

	paths := make([]string, 0)

	err := filepath.WalkDir(out, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}

		relative, relErr := filepath.Rel(out, path)
		if relErr != nil {
			return fmt.Errorf("relative path of %s: %w", path, relErr)
		}

		paths = append(paths, filepath.ToSlash(relative))

		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", out, err)
	}

	slices.Sort(paths)

	return paths
}
