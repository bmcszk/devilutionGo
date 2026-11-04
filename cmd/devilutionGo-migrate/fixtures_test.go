package main

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	mpq "github.com/suprsokr/go-mpq"
	"gopkg.in/yaml.v3"
)

// gfxFixtures holds the committed, text-only gfx goldens the migrate tests feed
// into synthetic archives (internal/testdata/gfx/README.md). No test reads a
// retail MPQ or a devilutionX checkout.
const gfxFixtures = "../../internal/testdata/gfx"

// readHex decodes a committed 16-bytes-per-line hexdump.
func readHex(t *testing.T, path string) []byte {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}

	decoded, err := hex.DecodeString(strings.Join(strings.Fields(string(content)), ""))
	if err != nil {
		t.Fatalf("decode fixture %s: %v", path, err)
	}

	return decoded
}

// fixtureBytes returns one committed gfx golden by its repo-relative name.
func fixtureBytes(t *testing.T, name string) []byte {
	t.Helper()

	return readHex(t, filepath.Join(gfxFixtures, filepath.FromSlash(name)))
}

// palFixture is the 768-byte retail palette (ui_art/diablo.pal).
func palFixture(t *testing.T) []byte {
	t.Helper()

	return fixtureBytes(t, "pal/diablo.pal.hex")
}

// trnFixture is a 256-byte colour translation (gendata/pause.trn).
func trnFixture(t *testing.T) []byte {
	t.Helper()

	return fixtureBytes(t, "trn/gendata_pause.trn.hex")
}

// clxFixture is a real single-frame CLX file (data/health.clx).
func clxFixture(t *testing.T) []byte {
	t.Helper()

	return fixtureBytes(t, "clx/data_health.clx.hex")
}

// pcxFixture is a 4x3 8-bit PCX (file rows 0x10, 0x20, 0x30).
func pcxFixture(t *testing.T) []byte {
	t.Helper()

	return fixtureBytes(t, "pcx/four_by_three.pcx.hex")
}

// fontSheetFixture is a real 256-glyph CLX font sheet (fonts/12-05.clx), from the
// committed assets fixtures rather than the gfx corpus.
func fontSheetFixture(t *testing.T) []byte {
	t.Helper()

	return readHex(t, "../../internal/testdata/assets/fonts/12-05.clx.hex")
}

// tsvFixture is a small game table with a header and two data rows.
func tsvFixture() []byte {
	return []byte("id\tname\tlevel\n0\tZombie\t1\n1\tSkeleton\t2\n")
}

// writeArchive packs files into a synthetic MPQ under dir and returns its path.
// go-mpq writes the archive, the proven wrapper reads it.
func writeArchive(t *testing.T, dir string, files map[string][]byte) string {
	t.Helper()

	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create archive dir: %v", err)
	}

	path := filepath.Join(dir, "spawn.mpq")

	archive, err := mpq.Create(path, 1)
	if err != nil {
		t.Fatalf("create archive: %v", err)
	}

	addMembers(t, archive, files)

	if err := archive.Close(); err != nil {
		t.Fatalf("write archive: %v", err)
	}

	return path
}

// readManifest parses the YAML asset list the run wrote.
func readManifest(t *testing.T, out string) manifest {
	t.Helper()

	content, err := os.ReadFile(filepath.Join(out, manifestName))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}

	listed := manifest{}

	if err := yaml.Unmarshal(content, &listed); err != nil {
		t.Fatalf("parse manifest: %v", err)
	}

	return listed
}

// readOutput reads one written output file.
func readOutput(t *testing.T, out, name string) []byte {
	t.Helper()

	content, err := os.ReadFile(filepath.Join(out, filepath.FromSlash(name)))
	if err != nil {
		t.Fatalf("read output %s: %v", name, err)
	}

	return content
}

// addMembers packs every member into the archive.
func addMembers(t *testing.T, archive *mpq.Archive, files map[string][]byte) {
	t.Helper()

	for inner, content := range files {
		source := filepath.Join(t.TempDir(), "member.bin")

		if err := os.WriteFile(source, content, 0o644); err != nil {
			t.Fatalf("write member source: %v", err)
		}

		if err := archive.AddFile(source, inner); err != nil {
			t.Fatalf("add %s: %v", inner, err)
		}
	}
}
