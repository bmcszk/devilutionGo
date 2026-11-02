package persist

import (
	"fmt"
	"os"
	"strconv"
)

// The reference-not-found / same / difference outcomes of a demo comparison
// (Source/pfile.h:89-93).
const (
	// DemoReferenceNotFound means the reference save for the demo is absent.
	DemoReferenceNotFound HeroCompareStatus = iota
	// DemoSame means every compared archive entry matched.
	DemoSame
	// DemoDifference means at least one entry differed.
	DemoDifference
)

// HeroCompareStatus is HeroCompareResult::Status (Source/pfile.h:88-93).
type HeroCompareStatus uint8

// HeroCompareResult is HeroCompareResult (Source/pfile.h:88-95): the status
// plus the human-readable diff report.
type HeroCompareResult struct {
	// Status is the comparison outcome.
	Status HeroCompareStatus
	// Message is the newline-joined diff report; empty when Status is Same or
	// ReferenceNotFound.
	Message string
}

// String names the status for test failure messages and demo logging
// (Source/engine/demomode.cpp:970-980).
func (s HeroCompareStatus) String() string {
	switch s {
	case DemoReferenceNotFound:
		return "ReferenceNotFound"
	case DemoSame:
		return "Same"
	case DemoDifference:
		return "Difference"
	default:
		return "HeroCompareStatus(" + strconv.Itoa(int(s)) + ")"
	}
}

// DemoReferencePrefix is the save prefix of the demo reference save
// (pfile.cpp:667, StrCat("demo_", demo, "_reference_")).
func DemoReferencePrefix(demo int) string {
	return "demo_" + strconv.Itoa(demo) + "_reference_"
}

// DemoActualPrefix is the save prefix of the demo actual save (pfile.cpp:672).
func DemoActualPrefix(demo int) string {
	return "demo_" + strconv.Itoa(demo) + "_actual_"
}

// CompareHeroDemo is pfile_compare_hero_demo (Source/pfile.cpp:665-679): it
// checks for the demo reference archive, asks the caller to write the actual
// archive (CopySaveFile + pfile_write_hero at pfile.cpp:673-677) and compares
// the two.
//
// writeActual receives the actual archive path and must produce that archive;
// it is the injection point for the game-side save write, which this package
// does not own.
func CompareHeroDemo(ns SaveNamespace, dir string, slot, demo int, opts CompareOptions,
	writeActual func(path string) error,
) (HeroCompareResult, error) {
	reference, actual, err := demoPaths(ns, dir, slot, demo)
	if err != nil {
		return HeroCompareResult{}, err
	}

	// pfile.cpp:669-670: a missing reference is not an error, it is a status.
	if !demoReferenceExists(reference) {
		return HeroCompareResult{Status: DemoReferenceNotFound}, nil
	}

	if err := writeActual(actual); err != nil {
		return HeroCompareResult{}, fmt.Errorf("persist: write actual demo save %s: %w", actual, err)
	}

	mode := ns.mode()

	return CompareSaves(LegacySVStore{Path: actual, Mode: mode},
		LegacySVStore{Path: reference, Mode: mode}, opts)
}

// demoReferenceExists is the FileExists guard at pfile.cpp:669.
func demoReferenceExists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}

// demoPaths resolves the reference and actual save paths of a demo run
// (pfile.cpp:667, :672).
func demoPaths(ns SaveNamespace, dir string, slot, demo int) (reference, actual string, err error) {
	reference, err = ns.SavePath(dir, slot, DemoReferencePrefix(demo))
	if err != nil {
		return "", "", fmt.Errorf("persist: demo %d reference path: %w", demo, err)
	}

	actual, err = ns.SavePath(dir, slot, DemoActualPrefix(demo))
	if err != nil {
		return "", "", fmt.Errorf("persist: demo %d actual path: %w", demo, err)
	}

	return reference, actual, nil
}

// mode is the GameMode the namespace's password selection implies
// (pfile_get_password, pfile.cpp:638-643).
func (n SaveNamespace) mode() GameMode {
	return GameMode{Spawn: n.Spawn, Multiplayer: n.Multiplayer}
}

// compareTargets is the CompareTargets list CompareSaves walks
// (pfile.cpp:473-484): the hero and game streams, the additional-missile
// stream and one entry per perm level, the first of which is the town level.
func compareTargets(levelCount int) ([]compareTarget, error) {
	targets := make([]compareTarget, 0, 3+2*levelCount)
	targets = append(targets,
		compareTarget{fileName: HeroEntryName},
		compareTarget{fileName: GameEntryName},
		compareTarget{fileName: AdditionalMissilesEntryName},
	)

	for i := range levelCount * 2 {
		kind, levelNum := LevelRegular, i

		if i >= levelCount {
			kind, levelNum = LevelSet, i-levelCount
		}

		name, err := PermLevelEntryName(kind, levelNum)
		if err != nil {
			return nil, fmt.Errorf("persist: compare target %d: %w", i, err)
		}

		targets = append(targets, compareTarget{fileName: name, isTownLevel: i == 0})
	}

	return targets, nil
}
