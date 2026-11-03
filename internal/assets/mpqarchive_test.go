package assets_test

import (
	"encoding/binary"
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

func TestOpenMPQArchiveRejectsNonsensicalTableCounts(t *testing.T) {
	// given headers whose table counts no real archive carries
	cases := map[string]struct {
		hashCount  uint32
		blockCount uint32
		want       string
	}{
		"zero hashes":      {hashCount: 0, blockCount: 1, want: "not a power of two"},
		"odd hash count":   {hashCount: 3, blockCount: 1, want: "not a power of two"},
		"huge hash count":  {hashCount: 1 << 30, blockCount: 1, want: "past the"},
		"huge block count": {hashCount: 16, blockCount: 0xFFFFFFFF, want: "past the"},
	}

	// when each header is opened
	failures := openAll(t, cases)

	// then the layout guard refuses it with a named reason
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) { assertHeaderRejected(t, failures[name], name, testCase.want) })
	}
}

func TestOpenMPQArchiveAcceptsEmptyArchive(t *testing.T) {
	// given a legal archive that holds one hash slot and no blocks
	path := writeEmptyArchive(t)

	// when the archive is opened
	archive, err := assets.OpenMPQArchive(path)
	// then an empty archive is not a corrupt one: the migration reports it as
	// unaddressable instead of failing to open it
	if err != nil {
		t.Fatalf("OpenMPQArchive: %v", err)
	}

	if got := archive.RecordCount(); got != 0 {
		t.Errorf("RecordCount = %d, want 0", got)
	}
}

// assertHeaderRejected checks that a corrupt header was refused with the reason.
// openAll opens an archive per case and returns the resulting errors by case name.
func openAll(t *testing.T, cases map[string]struct {
	hashCount  uint32
	blockCount uint32
	want       string
},
) map[string]error {
	t.Helper()

	failures := make(map[string]error, len(cases))

	for name, testCase := range cases {
		_, failures[name] = assets.OpenMPQArchive(writeHeaderOnlyArchive(t, testCase.hashCount, testCase.blockCount))
	}

	return failures
}

func assertHeaderRejected(t *testing.T, err error, name, want string) {
	t.Helper()

	if err == nil {
		t.Fatalf("%s: OpenMPQArchive accepted a corrupt header", name)
	}

	if !strings.Contains(err.Error(), want) {
		t.Errorf("error = %v, want it to mention %q", err, want)
	}
}

// writeHeaderOnlyArchive writes a 32-byte header that points its tables just past
// the header, with the counts under test.
func writeHeaderOnlyArchive(t *testing.T, hashCount, blockCount uint32) string {
	t.Helper()

	header := make([]byte, 32)
	binary.LittleEndian.PutUint32(header[0:], 0x1A51504D)
	binary.LittleEndian.PutUint32(header[4:], 32)
	binary.LittleEndian.PutUint32(header[16:], 32)
	binary.LittleEndian.PutUint32(header[20:], 32)
	binary.LittleEndian.PutUint32(header[24:], hashCount)
	binary.LittleEndian.PutUint32(header[28:], blockCount)

	path := filepath.Join(t.TempDir(), "corrupt-header.mpq")
	if err := os.WriteFile(path, header, 0o644); err != nil {
		t.Fatalf("write header-only archive: %v", err)
	}

	return path
}

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

// writeEmptyArchive writes a legal archive with one hash slot and no blocks.
func writeEmptyArchive(t *testing.T) string {
	t.Helper()

	header := make([]byte, 32+16)
	binary.LittleEndian.PutUint32(header[0:], 0x1A51504D)
	binary.LittleEndian.PutUint32(header[4:], 32)
	binary.LittleEndian.PutUint32(header[8:], uint32(len(header)))
	binary.LittleEndian.PutUint32(header[16:], 32)
	binary.LittleEndian.PutUint32(header[20:], 48)
	binary.LittleEndian.PutUint32(header[24:], 1)

	path := filepath.Join(t.TempDir(), "empty.mpq")
	if err := os.WriteFile(path, header, 0o644); err != nil {
		t.Fatalf("write empty archive: %v", err)
	}

	return path
}
