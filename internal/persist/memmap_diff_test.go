package persist_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

// memMapContents are the synthetic descriptors the memory-map tests install;
// each test writes the file its case names plus any sub-descriptor it uses.
//
// The perm-level entry names are real CompareSaves targets, so the tests drive
// the whole compare path instead of the unexported diff helper.
//
//nolint:gochecknoglobals // test fixture table
var memMapContents = map[string]string{
	"perml00.txt": "R_DL 8 a\nR 8 b\n",
	"perml01.txt": "R_DL 8 a\nR 8 b\n",
	"perml02.txt": "R_HF 8 a\nR_DA 8 b\nR 8 c\n",
	"perml03.txt": "LC 32 n\nM n 8 v\n",
	"perml04.txt": "C 2 sub nested\n",
	"perml05.txt": "R 8 a\n",
	"perml06.txt": "R 8 a\nR 8 b\n",
	"sub.txt":     "R 8 x\n",
}

// memMapCase is one row of the descriptor-driven compare tests.
type memMapCase struct {
	name     string
	level    int
	original []byte
	mutated  []byte
	hellfire bool
	want     string
}

func TestDiffEntryFields(t *testing.T) {
	t.Parallel()

	cases := []memMapCase{
		{
			"second field", 6,
			[]byte{1, 2},
			[]byte{1, 9},
			false,
			`file "perml06" has different content.` + "\nDiff found in perml06.b count: 1",
		},
		{
			"first field", 6,
			[]byte{1, 2},
			[]byte{9, 2},
			false,
			`file "perml06" has different content.` + "\nDiff found in perml06.a count: 1",
		},
		{
			"every field", 6,
			[]byte{1, 2},
			[]byte{8, 9},
			false,
			`file "perml06" has different content.` + "\nDiff found in perml06.a count: 1" +
				"\nDiff found in perml06.b count: 1",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// when the pair is compared with detail logging on
			result, err := memMapCompare(t, tc)

			// then the report matches the descriptor walk
			requireDetailMessage(t, result, err, tc.want)
		})
	}
}

func TestDiffEntryRepeatedCounter(t *testing.T) {
	t.Parallel()

	// Descriptor perml03.txt reads a count (LC 32 n) and then that many 1-byte
	// values (M n 8 v). The entries' counts differ, so only the shared prefix
	// is compared.
	caseData := memMapCase{
		level:    3,
		original: []byte{0, 0, 0, 2, 7, 8},
		mutated:  []byte{0, 0, 0, 3, 7, 9, 10},
		// the count field and the one differing shared value are both
		// reported (pfile.cpp:424-448)
		want: `file "perml03" is different size. Expected: 6 Actual: 7` +
			"\nDiff found in perml03.n count: 1\nDiff found in perml03.v count: 1",
	}

	// when compared
	result, err := memMapCompare(t, caseData)

	// then the count field and the differing shared value are both reported
	requireDetailMessage(t, result, err, caseData.want)
}

func TestDiffEntryCompositeNesting(t *testing.T) {
	t.Parallel()

	// Descriptor perml04.txt nests two sub-descriptor passes (C 2 sub nested)
	// into sub.txt (R 8 x).
	caseData := memMapCase{
		level:    4,
		original: []byte{1, 2},
		mutated:  []byte{1, 9},
		// the diff key carries the dotted nesting path
		want: `file "perml04" has different content.` + "\nDiff found in perml04.nested.x count: 1",
	}

	// when compared
	result, err := memMapCompare(t, caseData)

	// then the nested diff key is reported
	requireDetailMessage(t, result, err, caseData.want)
}

func TestDiffEntryHellfireSuffixes(t *testing.T) {
	t.Parallel()

	// Descriptor perml02.txt has one Hellfire-only and one Diablo-only line
	// (R_HF 8 a / R_DA 8 b / R 8 c).
	cases := []memMapCase{
		{
			"diablo reads the _DA line", 2,
			[]byte{1, 2},
			[]byte{9, 9},
			false,
			`file "perml02" has different content.` +
				"\nDiff found in perml02.b count: 1\nDiff found in perml02.c count: 1",
		},
		{
			"hellfire reads the _HF line", 2,
			[]byte{1, 2},
			[]byte{9, 9},
			true,
			`file "perml02" has different content.` +
				"\nDiff found in perml02.a count: 1\nDiff found in perml02.c count: 1",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// when the pair is compared with detail logging on
			result, err := memMapCompare(t, tc)

			// then the report matches the descriptor walk
			requireDetailMessage(t, result, err, tc.want)
		})
	}
}

func TestDiffEntryTownLevelSuffix(t *testing.T) {
	t.Parallel()

	// Descriptor perml00/perml01 start with a _DL gated line (R_DL 8 a, then
	// R 8 b). perml00 is the town level target (CompareTargets isTownLevel)
	// and perml01 is not.
	cases := []memMapCase{
		// the town entry carries no value for the gated field, so the
		// descriptor only covers the second byte
		{
			"town level skips the _DL line", 0,
			[]byte{2},
			[]byte{9},
			false,
			`file "perml00" has different content.` + "\nDiff found in perml00.b count: 1",
		},
		{
			"non-town level reads the _DL line", 1,
			[]byte{1, 2},
			[]byte{9, 9},
			false,
			`file "perml01" has different content.` +
				"\nDiff found in perml01.a count: 1\nDiff found in perml01.b count: 1",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// when the pair is compared with detail logging on
			result, err := memMapCompare(t, tc)

			// then the report matches the descriptor walk
			requireDetailMessage(t, result, err, tc.want)
		})
	}
}

func TestDiffEntryUncomparedBytes(t *testing.T) {
	t.Parallel()

	// Descriptor perml05.txt covers one byte (R 8 a); the entry is two bytes.
	caseData := memMapCase{level: 5, original: []byte{1, 2}, mutated: []byte{1, 3}}

	// when compared
	_, err := memMapCompare(t, caseData)

	// then the walk reports the uncompared tail (pfile.cpp:520-523)
	if err == nil {
		t.Fatal("CompareSaves succeeded with an incomplete memory map")
	}
}

// requireDetailMessage fails unless the comparison produced the wanted report.
func requireDetailMessage(t *testing.T, result persist.HeroCompareResult, err error, want string) {
	t.Helper()

	if err != nil {
		t.Fatalf("CompareSaves: %v", err)
	}

	if result.Message != want {
		t.Fatalf("message = %q, want %q", result.Message, want)
	}
}

// memMapCompare writes both archives holding the case's perm-level entry and
// compares them against the synthetic descriptors.
func memMapCompare(t *testing.T, tc memMapCase) (persist.HeroCompareResult, error) {
	t.Helper()

	dir := t.TempDir()
	mapDir := writeMemMaps(t, memMapContents)

	entry, err := persist.PermLevelEntryName(persist.LevelRegular, tc.level)
	if err != nil {
		t.Fatalf("PermLevelEntryName: %v", err)
	}

	referencePath := filepath.Join(dir, "demo_0_reference_single_0.sv")
	actualPath := filepath.Join(dir, "demo_0_actual_single_0.sv")
	writeExtraArchive(t, referencePath, entry, tc.original)
	writeExtraArchive(t, actualPath, entry, tc.mutated)

	result, err := persist.CompareSaves(
		persist.LegacySVStore{Path: actualPath},
		persist.LegacySVStore{Path: referencePath},
		persist.CompareOptions{
			LevelCount: persist.LevelsDiablo,
			LogDetails: true,
			MemMapDir:  mapDir,
			Hellfire:   tc.hellfire,
		},
	)
	if err != nil {
		return persist.HeroCompareResult{}, fmt.Errorf("compare archives: %w", err)
	}

	return result, nil
}

// writeMemMaps installs descriptors into a temp dir and returns that dir.
func writeMemMaps(t *testing.T, files map[string]string) string {
	t.Helper()

	dir := t.TempDir()

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatalf("write memory map %s: %v", name, err)
		}
	}

	return dir
}

// writeExtraArchive writes a legacy archive carrying one auxiliary entry.
func writeExtraArchive(t *testing.T, path, entry string, data []byte) {
	t.Helper()

	store := persist.LegacySVStore{Path: path}
	err := store.WriteSave(persist.HeroSave{
		Hero:  persist.HeroRecord{Slot: 0, Pack: goldenPlayerPack().Encode()},
		Extra: map[string][]byte{entry: data},
	})
	if err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
