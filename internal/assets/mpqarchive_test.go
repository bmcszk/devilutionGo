package assets_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/assets"
)

func TestOpenMPQArchiveRejectsForeignFile(t *testing.T) {
	// given a file that is not an MPQ archive
	path := filepath.Join(t.TempDir(), "not-an-archive.mpq")
	if err := os.WriteFile(path, []byte("plain text, no magic"), 0o644); err != nil {
		t.Fatal(err)
	}

	// when it is opened as an archive
	_, err := assets.OpenMPQArchive(path)

	// then the magic check refuses it
	if err == nil {
		t.Fatal("OpenMPQArchive accepted a non-MPQ file")
	}

	if !strings.Contains(err.Error(), "magic") {
		t.Errorf("error = %v, want a magic rejection", err)
	}
}

func TestMPQArchiveRecordCountIncludesListfile(t *testing.T) {
	// given an archive written with two members
	path := writeArchive(t, map[string]string{
		`txtdata\items.tsv`:    "id\tname\n",
		`txtdata\monsters.tsv`: "id\tname\n",
	})

	// when the archive is opened
	archive := openArchive(t, path)

	// then it holds both members plus the (listfile) and (attributes)
	// bookkeeping records the writer emits
	if got, want := archive.RecordCount(), 4; got != want {
		t.Errorf("RecordCount = %d, want %d", got, want)
	}
}

func TestMPQArchiveReadFileRoundTripsStoredRecord(t *testing.T) {
	// given an archive holding a payload too small to be worth compressing
	content := "stored-member"
	path := writeArchive(t, map[string]string{`txtdata\items.tsv`: content})

	// when that member is read back
	archive := openArchive(t, path)

	got, err := archive.ReadFile(`txtdata\items.tsv`)
	// then the bytes match the written payload
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if string(got) != content {
		t.Errorf("ReadFile = %q, want %q", got, content)
	}
}

func TestMPQArchiveReadFileRoundTripsSectoredRecord(t *testing.T) {
	// given an archive holding a payload large enough to span sectors
	content := strings.Repeat("monster-row\tzombie\t0\n", 200)
	path := writeArchive(t, map[string]string{`txtdata\monsters.tsv`: content})

	// when that member is read back
	archive := openArchive(t, path)

	got, err := archive.ReadFile(`txtdata\monsters.tsv`)
	// then the decoded bytes match the payload across every sector
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if string(got) != content {
		t.Errorf("ReadFile returned %d bytes, want %d", len(got), len(content))
	}
}

func TestMPQArchiveReadFileAcceptsCaseInsensitiveNames(t *testing.T) {
	// given an archive whose member name is lower case
	path := writeArchive(t, map[string]string{`txtdata\items.tsv`: "id\n"})

	// when it is read through an upper-case spelling
	archive := openArchive(t, path)

	got, err := archive.ReadFile(`TXTData\ITEMS.TSV`)
	// then the hash lookup still resolves it
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if string(got) != "id\n" {
		t.Errorf("ReadFile = %q, want %q", got, "id\n")
	}
}

func TestMPQArchiveReadFileRejectsUnknownName(t *testing.T) {
	// given an archive without the requested member
	path := writeArchive(t, map[string]string{`txtdata\items.tsv`: "id\n"})

	// when an absent name is read
	archive := openArchive(t, path)

	_, err := archive.ReadFile(`txtdata\absent.tsv`)
	// then the lookup reports the missing record
	if err == nil {
		t.Fatal("ReadFile accepted an absent member")
	}
}

func TestMPQArchiveNamesListsWrittenMembers(t *testing.T) {
	// given an archive written with two members
	path := writeArchive(t, map[string]string{
		`txtdata\items.tsv`:    "id\n",
		`txtdata\monsters.tsv`: "id\n",
	})

	// when its own (listfile) record is read
	archive := openArchive(t, path)

	names, err := archive.Names()
	// then both inner names come back in archive spelling
	if err != nil {
		t.Fatalf("Names: %v", err)
	}

	if len(names) != 2 {
		t.Fatalf("Names = %v, want 2 entries", names)
	}

	for _, want := range []string{`txtdata\items.tsv`, `txtdata\monsters.tsv`} {
		if !containsName(names, want) {
			t.Errorf("Names = %v, missing %q", names, want)
		}
	}
}

// writeArchive writes map members into a fresh archive and returns its path.
func writeArchive(t *testing.T, files map[string]string) string {
	t.Helper()

	root := t.TempDir()
	writeMPQ(t, root, files)

	return filepath.Join(root, "spawn.mpq")
}

// openArchive opens path and closes it when the test ends.
func openArchive(t *testing.T, path string) *assets.MPQArchive {
	t.Helper()

	archive, err := assets.OpenMPQArchive(path)
	if err != nil {
		t.Fatalf("open archive: %v", err)
	}

	t.Cleanup(func() {
		if err := archive.Close(); err != nil {
			t.Errorf("close archive: %v", err)
		}
	})

	return archive
}

// containsName reports whether names holds want.
func containsName(names []string, want string) bool {
	for _, name := range names {
		if name == want {
			return true
		}
	}

	return false
}
