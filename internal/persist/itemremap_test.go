package persist_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

// remapCase is one row of the loadsave.cpp remap tables.
type remapCase struct {
	name string
	from int16
	want int16
}

// TestRemapItemIdxFromDiablo pins loadsave.cpp:2204-2222.
func TestRemapItemIdxFromDiablo(t *testing.T) {
	t.Parallel()

	for _, c := range diabloToCurrentCases() {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			// when a Diablo-only row is read into the current table
			got := persist.RemapItemIdxFromDiablo(c.from)

			// then it lands on the current row
			requireRemap(t, got, c.want)
		})
	}
}

// TestRemapItemIdxToDiablo pins loadsave.cpp:2226-2246, including the
// Hellfire-exclusive rows a Diablo build cannot represent.
func TestRemapItemIdxToDiablo(t *testing.T) {
	t.Parallel()

	for _, c := range currentToDiabloCases() {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			// when a current row is written for a Diablo build
			got := persist.RemapItemIdxToDiablo(c.from)

			// then it lands on the Diablo row or is dropped
			requireRemap(t, got, c.want)
		})
	}
}

// TestRemapItemIdxFromSpawn pins loadsave.cpp:2248-2277.
func TestRemapItemIdxFromSpawn(t *testing.T) {
	t.Parallel()

	for _, c := range spawnToCurrentCases() {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			// when a spawn row is read into the current table
			got := persist.RemapItemIdxFromSpawn(c.from)

			// then it lands on the current row
			requireRemap(t, got, c.want)
		})
	}
}

// TestRemapItemIdxToSpawn pins loadsave.cpp:2279-2302.
func TestRemapItemIdxToSpawn(t *testing.T) {
	t.Parallel()

	for _, c := range spawnToCurrentCases() {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			// when the current row is written back for a spawn save
			got := persist.RemapItemIdxToSpawn(c.want)

			// then it lands on the spawn row again
			requireRemap(t, got, c.from)
		})
	}
}

// TestRemapItemIdxSpawnRoundTrip checks the spawn pair is a true inverse over
// the whole row range, so a spawn save cannot drift a row.
func TestRemapItemIdxSpawnRoundTrip(t *testing.T) {
	t.Parallel()

	for i := range int16(200) {
		// when every row is read back and written again
		spawn := persist.RemapItemIdxFromSpawn(i)

		// then it returns to its original row
		requireRemap(t, persist.RemapItemIdxToSpawn(spawn), i)
	}
}

// TestRemapItemIdxDiabloRoundTrip pins the range where the Diablo pair is a
// true inverse, plus the upstream quirk above it: RemapItemIdxFromDiablo pushes
// the Hellfire tail (156..165) into the 161+ rows that RemapItemIdxToDiablo
// then drops, and 156 collides with IDI_SORCERER_DIABLO, so a Diablo save
// holding a Hellfire tail row does not survive a read/write cycle. Ported
// as-is (loadsave.cpp:2204-2246).
func TestRemapItemIdxDiabloRoundTrip(t *testing.T) {
	t.Parallel()

	for i := range int16(156) {
		// when a Diablo row below the Hellfire tail is read back and written again
		diablo := persist.RemapItemIdxFromDiablo(i)

		// then it returns to its original row
		requireRemap(t, persist.RemapItemIdxToDiablo(diablo), i)
	}

	// 156 lands exactly on IDI_SORCERER_DIABLO, which the reverse remap reads
	// as the sorcerer row; every other tail row round-trips to a drop.
	tail := persist.RemapItemIdxFromDiablo(156)
	requireRemap(t, persist.RemapItemIdxToDiablo(tail), 5)

	for i := int16(157); i < 200; i++ {
		// when a Hellfire tail row is read back and written again
		diablo := persist.RemapItemIdxFromDiablo(i)

		// then upstream drops it instead of restoring it
		requireRemap(t, persist.RemapItemIdxToDiablo(diablo), -1)
	}
}

// diabloToCurrentCases covers each branch of RemapItemIdxFromDiablo: the
// sorcerer swap, the Hellfire tail, the Scroll of Search and the oils.
func diabloToCurrentCases() []remapCase {
	return []remapCase{
		{"sorcerer", persist.ItemSorcererRow, persist.ItemSorcererDiabloRow},
		{"veil of steel unchanged", 14, 14},
		{"first oil", 83, 87},
		{"last oil", 86, 90},
		{"first scroll", 88, 93},
		{"scroll of search", 92, 97},
		{"last scroll", 93, 98},
		{"before hellfire tail", 155, 160},
		{"first hellfire row", 156, 166},
		{"last hellfire row", 160, 170},
		{"sorcerer diablo collides", persist.ItemSorcererDiabloRow, 176},
	}
}

// currentToDiabloCases covers the drop branches: oils, the Scroll of Search and
// everything from the Hellfire tail on.
func currentToDiabloCases() []remapCase {
	return []remapCase{
		{"sorcerer diablo", persist.ItemSorcererDiabloRow, persist.ItemSorcererRow},
		{"first oil dropped", 83, -1},
		{"last oil dropped", 86, -1},
		{"before oil", 82, 82},
		{"after oil first", 87, 83},
		{"scroll of search dropped", 92, -1},
		{"first scroll", 93, 88},
		{"before hellfire tail", 160, 155},
		{"hellfire tail dropped", 161, -1},
		{"far hellfire row dropped", 200, -1},
	}
}

// spawnToCurrentCases covers the armor block and the six scroll insertions.
func spawnToCurrentCases() []remapCase {
	return []remapCase{
		{"light armor unchanged", 61, 61},
		{"first armor", 62, 71},
		{"last armor", 70, 79},
		{"before stone curse", 95, 110},
		{"stone curse", 96, 111},
		{"guardian", 98, 113},
		{"third scroll", 99, 114},
		{"golem", 101, 116},
		{"none", 102, 117},
		{"apocalypse", 104, 119},
		{"tail item", 130, 145},
	}
}

// requireRemap fails unless got is want.
func requireRemap(t *testing.T, got, want int16) {
	t.Helper()

	if got != want {
		t.Fatalf("remap = %d, want %d", got, want)
	}
}
