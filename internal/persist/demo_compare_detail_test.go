package persist_test

import (
	"path/filepath"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

func TestCompareSavesLogDetailsHeroGold(t *testing.T) {
	t.Parallel()

	// given two hero entries that differ only in pGold, the field hero.txt
	// declares as "R 32 pGold"
	dir := t.TempDir()
	game := demoGameBytes(t)

	referencePack := goldenPlayerPack()
	actualPack := goldenPlayerPack()
	actualPack.Gold = referencePack.Gold + 7

	referencePath := filepath.Join(dir, "demo_7_reference_single_0.sv")
	actualPath := filepath.Join(dir, "demo_7_actual_single_0.sv")
	writeDemoArchive(t, referencePath, referencePack.Encode(), game)
	writeDemoArchive(t, actualPath, actualPack.Encode(), game)

	// when compared with detail logging on
	result, err := persist.CompareSaves(
		persist.LegacySVStore{Path: actualPath},
		persist.LegacySVStore{Path: referencePath},
		demoOptions(true),
	)
	if err != nil {
		t.Fatalf("CompareSaves: %v", err)
	}

	// then the report names the archive entry and the differing map field
	if result.Status != persist.DemoDifference {
		t.Fatalf("status = %v, want Difference", result.Status)
	}

	const want = "file \"hero\" has different content.\nDiff found in hero.pGold count: 1"
	if result.Message != want {
		t.Fatalf("message = %q, want %q", result.Message, want)
	}
}

func TestCompareSavesLogDetailsOffStopsAtEntry(t *testing.T) {
	t.Parallel()

	// given the same gold mismatch as above
	dir := t.TempDir()
	game := demoGameBytes(t)

	referencePack := goldenPlayerPack()
	actualPack := goldenPlayerPack()
	actualPack.Gold = referencePack.Gold + 7

	referencePath := filepath.Join(dir, "demo_8_reference_single_0.sv")
	actualPath := filepath.Join(dir, "demo_8_actual_single_0.sv")
	writeDemoArchive(t, referencePath, referencePack.Encode(), game)
	writeDemoArchive(t, actualPath, actualPack.Encode(), game)

	// when logDetails is off
	result, err := persist.CompareSaves(
		persist.LegacySVStore{Path: actualPath},
		persist.LegacySVStore{Path: referencePath},
		demoOptions(false),
	)
	if err != nil {
		t.Fatalf("CompareSaves: %v", err)
	}

	// then no "Diff found in" line is emitted (pfile.cpp:509-510)
	const want = "file \"hero\" has different content."
	if result.Message != want {
		t.Fatalf("message = %q, want %q", result.Message, want)
	}
}
