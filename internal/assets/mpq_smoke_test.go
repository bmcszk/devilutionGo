package assets_test

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/assets"
)

// mpqListfileNameForTest is the record name every archive uses for its own name
// list; no (listfile) record lists itself.
const mpqListfileNameForTest = "(listfile)"

// assetArchiveNames are the retail archives the extraction smoke test must
// decode exhaustively, in the order the game stacks them. spawn.mpq is absent
// from this environment (recorded in the wave PoW as an amendment, not a gap).
var assetArchiveNames = []string{"DIABDAT.MPQ", "hellfire.mpq", "devilutionx.mpq", "Patch_rt.mpq"}

// archivesWithoutListfile are the retail archives that ship no (listfile)
// record. MPQ encryption keys on the base file name, so without a name a record
// cannot even be decrypted; the embedded inventories under internal/assets/names
// are hash-verified.
var archivesWithoutListfile = map[string]bool{
	"diabdat.mpq":  true,
	"hellfire.mpq": true,
}

// TestRealArchivesDecodeEveryRecord walks every record of every retail archive
// and asserts all of them decode to their recorded size. The archives are
// environment inputs, never committed fixtures, so an absence skips instead of
// failing — the same politeness DevilutionX uses for its real-asset tests
// (test/timedemo_test.cpp:47-49). A record that exists but does not decode is a
// failure, never a skip.
func TestRealArchivesDecodeEveryRecord(t *testing.T) {
	// given the retail archives reachable from DG_ASSET_ARCHIVES
	archives := findAssetArchives(t)

	// when every record of every archive is decoded
	failures := decodeAllRecords(t, archives)

	// then no record fails
	if len(failures) > 0 {
		t.Fatalf("%d record(s) failed to decode:\n%s", len(failures), strings.Join(failures, "\n"))
	}
}

// TestRealMainMenuPCXDecodes decodes the retail PCX that AC4 names, through the
// same path the game will use: record lookup, ARC decryption, PKWARE DCL
// explode, then the PCX decoder.
func TestRealMainMenuPCXDecodes(t *testing.T) {
	// given the retail main menu PCX from DIABDAT.MPQ
	archive := openRetailArchive(t, "DIABDAT.MPQ")

	data, err := archive.ReadFile(`ui_art\mainmenu.pcx`)
	if err != nil {
		t.Fatalf("read mainmenu.pcx: %v", err)
	}

	// when it is decoded as a PCX
	frame, err := assets.DecodePCX(data)
	// then the header byte and the 640x480 palette frame come out
	if err != nil {
		t.Fatalf("DecodePCX: %v", err)
	}

	if data[0] != 0x0A {
		t.Errorf("manufacturer byte %#02x, want 0x0A", data[0])
	}

	if frame.Width != 640 || frame.Height != 480 {
		t.Errorf("frame = %dx%d, want 640x480", frame.Width, frame.Height)
	}

	if !frame.HasPalette {
		t.Error("frame carries no palette")
	}
}

// findAssetArchives locates the retail archives under the directories listed in
// DG_ASSET_ARCHIVES (os.PathListSeparator separated) and skips the test when it
// finds none.
func findAssetArchives(t *testing.T) []string {
	t.Helper()

	dirs := filepath.SplitList(os.Getenv("DG_ASSET_ARCHIVES"))
	found := make([]string, 0, len(assetArchiveNames))

	for _, dir := range dirs {
		found = append(found, archivesInDir(dir)...)
	}

	if len(found) == 0 {
		t.Skip("no retail MPQ found; set DG_ASSET_ARCHIVES to the Diablo install directories")
	}

	return found
}

// archivesInDir returns the known archive names that dir holds.
func archivesInDir(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	found := make([]string, 0, len(assetArchiveNames))

	for _, entry := range entries {
		if !entry.IsDir() && isAssetArchiveName(entry.Name()) {
			found = append(found, filepath.Join(dir, entry.Name()))
		}
	}

	slices.Sort(found)

	return found
}

// isAssetArchiveName reports whether name matches one of the retail archives,
// ignoring case: the retail archive is spelled uppercase, the add-ons lowercase.
func isAssetArchiveName(name string) bool {
	for _, archive := range assetArchiveNames {
		if strings.EqualFold(name, archive) {
			return true
		}
	}

	return false
}

// openRetailArchive opens one named retail archive from DG_ASSET_ARCHIVES and
// skips when the environment does not hold it.
func openRetailArchive(t *testing.T, name string) *assets.MPQArchive {
	t.Helper()

	for _, path := range findAssetArchives(t) {
		if !strings.EqualFold(filepath.Base(path), name) {
			continue
		}

		return openArchive(t, path)
	}

	t.Skipf("%s not present under DG_ASSET_ARCHIVES", name)

	return nil
}

// decodeAllRecords decodes every record of every archive and returns the
// failures as "<archive> <record>: <error>" lines, sorted so the report is
// reproducible.
func decodeAllRecords(t *testing.T, archives []string) []string {
	t.Helper()

	failures := make([]string, 0, len(archives))

	for _, path := range archives {
		failures = append(failures, decodeArchive(t, path)...)
	}

	slices.Sort(failures)

	return failures
}

// decodeArchive decodes every record of one archive and reports the failures.
// Records are addressed by the archive's own (listfile) when it ships one, else
// by the committed fallback inventory.
func decodeArchive(t *testing.T, path string) []string {
	t.Helper()

	archive := openArchive(t, path)
	base := filepath.Base(path)

	names := archiveNames(t, archive, base)
	decoded := make(map[int]struct{}, len(names))
	failures := make([]string, 0)

	for _, name := range names {
		index, found := archive.RecordIndex(name)
		if !found {
			continue
		}

		if _, duplicate := decoded[index]; duplicate {
			continue
		}

		decoded[index] = struct{}{}

		if failure := decodeRecord(archive, base, index, name); failure != "" {
			failures = append(failures, failure)
		}
	}

	failures = append(failures, unreachableRecords(base, archive, decoded)...)
	t.Logf("%s: %d record(s) (%s), %d decoded, %d failure(s)",
		base, archive.RecordCount(), recordCensus(archive), len(decoded), len(failures))

	return failures
}

// archiveNames returns the names that address archive's records: its own
// (listfile) when present, else the committed fallback inventory.
func archiveNames(t *testing.T, archive *assets.MPQArchive, base string) []string {
	t.Helper()

	names, err := archive.Names()
	if err != nil {
		t.Fatalf("%s: listfile: %v", base, err)
	}

	if len(names) > 0 {
		// The listfile never lists itself, so add its own record name.
		return append(names, mpqListfileNameForTest)
	}

	return readListfile(t, base)
}

// decodeRecord reads one record and checks it decompressed to its recorded size.
func decodeRecord(archive *assets.MPQArchive, base string, index int, name string) string {
	content, err := archive.ReadFile(name)
	if err != nil {
		return fmt.Sprintf("%s record %d %s: %v", base, index, name, err)
	}

	if size := archive.Record(index).FileSize; len(content) != int(size) {
		return fmt.Sprintf("%s record %d %s: decoded %d bytes, want %d",
			base, index, name, len(content), size)
	}

	return ""
}

// unreachableRecords reports the used records no name addresses. An archive that
// ships neither a (listfile) nor a fallback inventory cannot be walked: MPQ
// keys encryption on the file name, so a nameless record cannot be decrypted.
func unreachableRecords(base string, archive *assets.MPQArchive, decoded map[int]struct{}) []string {
	failures := make([]string, 0)

	for index := range archive.RecordCount() {
		if !archive.Record(index).Used() {
			continue
		}

		if _, named := decoded[index]; named {
			continue
		}

		failures = append(failures, fmt.Sprintf("%s record %d: no name resolves it", base, index))
	}

	return failures
}

// recordCensus summarizes an archive's records by compression flag.
func recordCensus(archive *assets.MPQArchive) string {
	var stored, imploded, compressed, unused int

	for index := range archive.RecordCount() {
		record := archive.Record(index)

		switch {
		case !record.Used():
			unused++
		case record.Imploded():
			imploded++
		case record.Compressed():
			compressed++
		default:
			stored++
		}
	}

	return fmt.Sprintf("stored %d, imploded %d, compressed %d, unused %d",
		stored, imploded, compressed, unused)
}

// readListfile loads the committed name inventory for an archive that ships no
// (listfile), and fails when the archive has no inventory at all.
func readListfile(t *testing.T, base string) []string {
	t.Helper()

	if !archivesWithoutListfile[strings.ToLower(base)] {
		t.Fatalf("%s ships no (listfile) and has no committed name inventory", base)
	}

	names, ok := assets.NameInventory(base)
	if !ok {
		t.Fatalf("%s ships no (listfile) and has no committed name inventory", base)
	}

	return names
}
