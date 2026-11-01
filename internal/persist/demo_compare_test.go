package persist_test

import (
	"path/filepath"
	"strconv"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

// demoMemMapDir is the committed copy of upstream's
// test/fixtures/memory_map (c20b2f2456), used for the detail-diff tests.
const demoMemMapDir = "testdata/memory_map"

// demoPrefixCase is one row of TestDemoPrefixes.
type demoPrefixCase struct {
	name   string
	prefix func(int) string
	demo   int
	want   string
}

func TestDemoPrefixes(t *testing.T) {
	t.Parallel()

	for _, tc := range demoPrefixCases() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// when the prefix is built for the row's demo number
			got := tc.prefix(tc.demo)

			// then the demo number is spelled out between the fixed tokens
			if got != tc.want {
				t.Fatalf("prefix = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDemoPrefixSavePath(t *testing.T) {
	t.Parallel()

	// given the single-player namespace
	ns := persist.SaveNamespace{Extension: "sv"}

	// when the demo reference path is resolved like pfile.cpp:667
	got, err := ns.SavePath("/saves", 2, persist.DemoReferencePrefix(3))
	if err != nil {
		t.Fatalf("SavePath: %v", err)
	}

	// then the demo prefix lands in front of the slot prefix
	const want = "/saves/demo_3_reference_single_2.sv"
	if got != want {
		t.Fatalf("path = %q, want %q", got, want)
	}
}

func TestCompareHeroDemoReferenceNotFound(t *testing.T) {
	t.Parallel()

	// given an empty save directory with no reference archive
	dir := t.TempDir()
	ns := persist.SaveNamespace{Extension: "sv"}

	called := false

	// when the demo comparison runs
	result, err := persist.CompareHeroDemo(ns, dir, 0, 1, demoOptions(false), func(string) error {
		called = true
		return nil
	})
	// then it reports the missing reference and writes nothing
	if err != nil {
		t.Fatalf("CompareHeroDemo: %v", err)
	}

	if result.Status != persist.DemoReferenceNotFound {
		t.Fatalf("status = %v, want ReferenceNotFound", result.Status)
	}

	if result.Message != "" {
		t.Fatalf("message = %q, want empty", result.Message)
	}

	if called {
		t.Fatal("writeActual ran despite a missing reference")
	}
}

func TestCompareHeroDemoSame(t *testing.T) {
	t.Parallel()

	// given a reference archive and an actual save with identical entries
	dir := t.TempDir()
	ns := persist.SaveNamespace{Extension: "sv"}
	pack := goldenPlayerPack().Encode()
	game := demoGameBytes(t)
	writeDemoArchive(t, filepath.Join(dir, "demo_4_reference_single_0.sv"), pack, game)

	// when the actual save is written with the same content
	result, err := persist.CompareHeroDemo(ns, dir, 0, 4, demoOptions(false), func(path string) error {
		writeDemoArchive(t, path, pack, game)
		return nil
	})
	if err != nil {
		t.Fatalf("CompareHeroDemo: %v", err)
	}

	// then the comparison is Same with no message
	if result.Status != persist.DemoSame {
		t.Fatalf("status = %v (%s), want Same: %s", result.Status, result.Status, result.Message)
	}

	if result.Message != "" {
		t.Fatalf("message = %q, want empty", result.Message)
	}
}

func TestCompareHeroDemoDifference(t *testing.T) {
	t.Parallel()

	// given a reference archive whose game stream differs from the actual one
	dir := t.TempDir()
	ns := persist.SaveNamespace{Extension: "sv"}
	pack := goldenPlayerPack().Encode()
	writeDemoArchive(t, filepath.Join(dir, "demo_5_reference_single_0.sv"), pack, demoGameBytes(t))

	shorter := []byte{0x00, 0x00, 0x00, 0x01}

	// when the actual save is written with a different game stream
	result, err := persist.CompareHeroDemo(ns, dir, 0, 5, demoOptions(false), func(path string) error {
		writeDemoArchive(t, path, pack, shorter)
		return nil
	})
	if err != nil {
		t.Fatalf("CompareHeroDemo: %v", err)
	}

	// then the comparison reports the size difference (pfile.cpp:505-506)
	if result.Status != persist.DemoDifference {
		t.Fatalf("status = %v, want Difference", result.Status)
	}

	want := `file "game" is different size. Expected: ` + strconv.Itoa(len(demoGameBytes(t))) + ` Actual: 4`
	if result.Message != want {
		t.Fatalf("message = %q, want %q", result.Message, want)
	}
}

func TestCompareSavesContentDifference(t *testing.T) {
	t.Parallel()

	// given two archives whose game streams have the same length
	dir := t.TempDir()
	pack := goldenPlayerPack().Encode()
	game := demoGameBytes(t)
	actualGame := append([]byte(nil), game...)
	actualGame[len(actualGame)-1] ^= 0xFF

	referencePath := filepath.Join(dir, "demo_6_reference_single_0.sv")
	actualPath := filepath.Join(dir, "demo_6_actual_single_0.sv")
	writeDemoArchive(t, referencePath, pack, game)
	writeDemoArchive(t, actualPath, pack, actualGame)

	// when they are compared without detail logging
	result, err := persist.CompareSaves(
		persist.LegacySVStore{Path: actualPath},
		persist.LegacySVStore{Path: referencePath},
		demoOptions(false),
	)
	if err != nil {
		t.Fatalf("CompareSaves: %v", err)
	}

	// then only the same-size wording is reported (pfile.cpp:507-508)
	if result.Status != persist.DemoDifference {
		t.Fatalf("status = %v, want Difference", result.Status)
	}

	if result.Message != `file "game" has different content.` {
		t.Fatalf("message = %q, want the different-content wording", result.Message)
	}
}

// demoOptions is the CompareOptions the demo tests use: no Hellfire, no detail
// diff unless logDetails asks for it.
func demoOptions(logDetails bool) persist.CompareOptions {
	return persist.CompareOptions{
		LevelCount: persist.LevelsDiablo,
		LogDetails: logDetails,
		MemMapDir:  demoMemMapDir,
	}
}

// demoGameBytes is a minimal SaveGameData stream: a valid header, no levels.
func demoGameBytes(t *testing.T) []byte {
	t.Helper()

	game, err := persist.GameHeader{CurrLevel: 1}.Encode()
	if err != nil {
		t.Fatalf("GameHeader.Encode: %v", err)
	}

	return game
}

// writeDemoArchive writes a legacy .sv with the given hero pack and game stream.
func writeDemoArchive(t *testing.T, path string, pack, game []byte) {
	t.Helper()

	store := persist.LegacySVStore{Path: path}
	if err := store.Write(persist.HeroRecord{Slot: 0, Pack: pack}, game); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// demoPrefixCases are upstream's demo save prefixes: StrCat("demo_", demo,
// "_reference_") (pfile.cpp:667) and StrCat("demo_", demo, "_actual_")
// (pfile.cpp:672).
func demoPrefixCases() []demoPrefixCase {
	return []demoPrefixCase{
		{"reference 0", persist.DemoReferencePrefix, 0, "demo_0_reference_"},
		{"actual 0", persist.DemoActualPrefix, 0, "demo_0_actual_"},
		{"reference 12", persist.DemoReferencePrefix, 12, "demo_12_reference_"},
		{"actual 12", persist.DemoActualPrefix, 12, "demo_12_actual_"},
	}
}
