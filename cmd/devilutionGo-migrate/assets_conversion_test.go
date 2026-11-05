package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	manifestpkg "github.com/bmcszk/devilutionGo/internal/assets/manifest"
)

func TestAssetsConvertsLegacyRecordsTheInventoryDescribes(t *testing.T) {
	// given an archive holding a CEL and a CL2 record the inventory has widths for
	archive := writeArchive(t, t.TempDir(), map[string][]byte{
		`data\talkbutton.cel`: celFixture(t),
		`data\health.cl2`:     cl2Fixture(t),
	})
	inventoryPath := writeInventory(t, t.TempDir(), conversionInventory())
	out := filepath.Join(t.TempDir(), "assets-out")

	var stdout, stderr bytes.Buffer

	// when the extraction pass runs with that inventory
	code := run([]string{"assets", "--mpq", archive, "--inventory", inventoryPath, "--out", out}, &stdout, &stderr)

	// then it converts both records instead of copying them
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d (stderr: %s)", code, exitOK, stderr.String())
	}

	t.Run("clx bytes match the proven converter goldens", func(t *testing.T) {
		assertOutputBytes(t, out, "spawn/data/talkbutton.clx", fixtureBytes(t, "cel/data_talkbutton.cel.clx.hex"))
		assertOutputBytes(t, out, "spawn/data/health.clx", fixtureBytes(t, "cl2/data_health.cl2.clx.hex"))
	})
	t.Run("the legacy container is not written", func(t *testing.T) {
		assertAbsent(t, out, "spawn/data/talkbutton.cel", "spawn/data/health.cl2")
	})
	t.Run("the manifest records the conversion inputs", func(t *testing.T) {
		assertConvertedEntry(t, out, "spawn/data/talkbutton.clx",
			conversionRecord{Format: "cel", Widths: []int{53, 53}, Frames: 2})
		assertConvertedEntry(t, out, "spawn/data/health.clx",
			conversionRecord{Format: "cl2", Widths: []int{264}, Frames: 1})
	})
	t.Run("the report counts the conversions", func(t *testing.T) {
		if !strings.Contains(stdout.String(), "converted 2 legacy record(s) to clx") {
			t.Errorf("stdout = %q, want the conversion count", stdout.String())
		}
	})
}

func TestAssetsCopiesLegacyRecordsTheInventoryDoesNotDescribe(t *testing.T) {
	// given a CEL record and an inventory that describes a different file
	archive := writeArchive(t, t.TempDir(), map[string][]byte{`data\talkbutton.cel`: celFixture(t)})
	inventoryPath := writeInventory(t, t.TempDir(), fixtureInventory())
	out := filepath.Join(t.TempDir(), "assets-out")

	var stdout, stderr bytes.Buffer

	// when the extraction pass runs
	code := run([]string{"assets", "--mpq", archive, "--inventory", inventoryPath, "--out", out}, &stdout, &stderr)

	// then the record keeps today's copy-as-is shape
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d (stderr: %s)", code, exitOK, stderr.String())
	}

	assertOutputBytes(t, out, "spawn/data/talkbutton.cel", celFixture(t))

	entry := manifestEntry(t, out, "spawn/data/talkbutton.cel")
	if entry.Format != "cel" || entry.Conversion != nil {
		t.Errorf("entry = %s/%+v, want an unconverted cel record", entry.Format, entry.Conversion)
	}

	if strings.Contains(stdout.String(), "converted ") {
		t.Errorf("stdout = %q, want no conversion count", stdout.String())
	}
}

func TestAssetsFailsOnAMalformedInventory(t *testing.T) {
	// given an inventory file that is not YAML
	archive := writeArchive(t, t.TempDir(), map[string][]byte{`data\health.clx`: clxFixture(t)})
	inventoryPath := filepath.Join(t.TempDir(), "inventory.yaml")
	if err := os.WriteFile(inventoryPath, []byte("entries: [\n"), 0o644); err != nil {
		t.Fatalf("write inventory: %v", err)
	}

	// when the extraction pass runs
	code, _, stderr := runMigrate(t, "assets", "--mpq", archive, "--inventory", inventoryPath,
		"--out", filepath.Join(t.TempDir(), "assets-out"))

	// then converting on a guess is refused, not silently skipped
	if code != exitFailure {
		t.Fatalf("exit code = %d, want %d (stderr %q)", code, exitFailure, stderr)
	}

	if !strings.Contains(stderr, "inventory") {
		t.Errorf("stderr = %q, want the unreadable inventory named", stderr)
	}
}

// assertAbsent checks none of names was written below out.
func assertAbsent(t *testing.T, out string, names ...string) {
	t.Helper()

	for _, name := range names {
		if _, err := os.Stat(outputPath(out, name)); !os.IsNotExist(err) {
			t.Errorf("%s was written, want only its clx conversion", name)
		}
	}
}

// celFixture is the committed data/talkbutton.cel input (two 53-wide frames).
func celFixture(t *testing.T) []byte {
	t.Helper()

	return fixtureBytes(t, "cel/data_talkbutton.cel.hex")
}

// cl2Fixture is the committed data/health.cl2 input (one 264-wide frame).
func cl2Fixture(t *testing.T) []byte {
	t.Helper()

	return fixtureBytes(t, "cl2/data_health.cl2.hex")
}

// conversionInventory describes the two legacy records the conversion tests feed
// in, with the widths the committed C++ goldens were produced from.
func conversionInventory() manifestpkg.Manifest {
	return manifestpkg.Manifest{
		Tool:    manifestpkg.Tool,
		Version: manifestpkg.Version,
		Entries: []manifestpkg.Entry{
			{
				Family: "ui.data", Path: "data/talkbutton.clx", Format: "clx", Frames: 2,
				Widths: []int{53, 53}, WidthSource: "data/talkbutton.widths",
				Loader: "Source/diablo.cpp:1359",
			},
			{
				Family: "ui.data", Path: "data/health.clx", Format: "clx", Frames: 1, Width: 264,
				WidthSource: "Source/items.cpp:2424", Loader: "Source/diablo.cpp:1359",
			},
		},
	}
}
