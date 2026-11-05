package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	manifestpkg "github.com/bmcszk/devilutionGo/internal/assets/manifest"
)

// runMigrate runs one invocation through the real dispatch boundary and returns
// its exit code with the two message streams.
func runMigrate(t *testing.T, args ...string) (code int, stdout, stderr string) {
	t.Helper()

	var out, errOut bytes.Buffer

	code = run(args, &out, &errOut)

	return code, out.String(), errOut.String()
}

// writeFixtureFile writes a one-byte placeholder for one inventory path under
// root, creating the parent directory.
func writeFixtureFile(t *testing.T, root, path string) {
	t.Helper()

	full := filepath.Join(root, path)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(full), err)
	}

	if err := os.WriteFile(full, []byte("x"), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// writeInventory writes m as assets/inventory.yaml under dir and returns its path.
func writeInventory(t *testing.T, dir string, m manifestpkg.Manifest) string {
	t.Helper()

	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir inventory dir: %v", err)
	}

	path := filepath.Join(dir, "inventory.yaml")

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create inventory: %v", err)
	}

	if err := manifestpkg.Write(file, m); err != nil {
		t.Fatalf("write inventory: %v", err)
	}

	if err := file.Close(); err != nil {
		t.Fatalf("close inventory: %v", err)
	}

	return path
}

// fixtureInventory is the baseline two-entry inventory the nested subcommands
// verify and list.
func fixtureInventory() manifestpkg.Manifest {
	m := manifestpkg.Manifest{
		Tool:    manifestpkg.Tool,
		Version: manifestpkg.Version,
		Entries: []manifestpkg.Entry{
			{
				Family: "ui.data", Path: "data/health.clx", Format: "clx", Frames: 1, Width: 264,
				WidthSource: "Source/items.cpp:2424", Loader: "Source/diablo.cpp:1359",
			},
			{
				Family: "fonts", Path: "fonts/smallfont.trn", Format: "trn",
				Loader: "Source/engine/assets.cpp:410",
			},
		},
	}
	m.Finalize()

	return m
}
