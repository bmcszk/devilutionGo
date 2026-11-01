package persist

import (
	"fmt"
	"strings"
)

// compareTarget is CompareTargets (Source/pfile.cpp:473-478): an archive entry
// plus the memory-map descriptor used for the detail diff.
type compareTarget struct {
	// fileName is the archive entry name and its memory-map descriptor name.
	fileName string
	// isTownLevel marks the level whose memory map records the town flag.
	isTownLevel bool
}

// CompareOptions carries the inputs CompareSaves takes from upstream globals:
// giNumberOfLevels (docs/research/dg-requirements.md, level counts), the
// logDetails flag and the memory-map directory that replaces
// paths::BasePath()+"/test/fixtures/memory_map" (pfile.cpp:312).
type CompareOptions struct {
	// LevelCount is giNumberOfLevels: 16 for Diablo, 21 with Hellfire.
	LevelCount int
	// LogDetails enables the memory-map detail diff (pfile.cpp:509).
	LogDetails bool
	// MemMapDir holds the "<entry>.txt" memory-map descriptors. It is only
	// read when LogDetails is set and an entry differs.
	MemMapDir string
	// Hellfire is gbIsHellfire, which gates the _HF/_DA descriptor lines
	// (pfile.cpp:386-397).
	Hellfire bool
}

// CompareSaves is CompareSaves (Source/pfile.cpp:476-524): it walks the hero,
// game, additionalMissiles and perm-level entries of both archives and reports
// the size/content differences. Entries absent from both archives are skipped
// (pfile.cpp:497-499).
func CompareSaves(actual, reference LegacySVStore, opts CompareOptions) (HeroCompareResult, error) {
	targets, err := compareTargets(opts.LevelCount)
	if err != nil {
		return HeroCompareResult{}, err
	}

	actualEntries := readCompareEntries(actual, targets)
	referenceEntries := readCompareEntries(reference, targets)

	message, same := contentDiffs(targets, referenceEntries, actualEntries)

	status := DemoDifference
	if same {
		status = DemoSame
	}

	if !opts.LogDetails || same {
		return HeroCompareResult{Status: status, Message: strings.Join(message, "\n")}, nil
	}

	details, err := detailDiffs(targets, referenceEntries, actualEntries, opts)
	if err != nil {
		return HeroCompareResult{}, err
	}

	return HeroCompareResult{Status: status, Message: strings.Join(append(message, details...), "\n")}, nil
}

// contentDiffs collects the per-entry size/content messages (pfile.cpp:495-508)
// and reports whether every compared entry matched.
func contentDiffs(targets []compareTarget, referenceEntries, actualEntries [][]byte) ([]string, bool) {
	var message []string

	same := true

	for i, target := range targets {
		actualData, referenceData := actualEntries[i], referenceEntries[i]

		if len(actualData) == 0 && len(referenceData) == 0 {
			continue
		}

		if string(actualData) == string(referenceData) {
			continue
		}

		same = false
		message = append(message, sizeOrContentDiff(target.fileName, len(referenceData), len(actualData)))
	}

	return message, same
}

// sizeOrContentDiff formats the two diff wordings of pfile.cpp:505-508.
func sizeOrContentDiff(fileName string, referenceLen, actualLen int) string {
	if referenceLen != actualLen {
		return fmt.Sprintf("file %q is different size. Expected: %d Actual: %d",
			fileName, referenceLen, actualLen)
	}

	return fmt.Sprintf("file %q has different content.", fileName)
}

// readCompareEntries decodes the entries of one archive in target order; a
// missing entry yields nil, which CompareSaves treats as "absent"
// (ReadArchive returning nullptr, pfile.cpp:495-499).
func readCompareEntries(store LegacySVStore, targets []compareTarget) [][]byte {
	entries := make([][]byte, len(targets))

	for i, target := range targets {
		data, err := readCompareEntry(store, target.fileName)
		if err != nil {
			continue
		}

		entries[i] = data
	}

	return entries
}

// readCompareEntry is ReadArchive (pfile.cpp:613-628) for one entry: the hero
// entry is codec-decoded with the mode password. Unlike upstream, which runs
// codec_decode over every entry, this store keeps the non-hero streams in
// plaintext (LegacySVStore.WriteSave), so those pass through unchanged --
// feeding them to the codec would be rejected as a bad signature.
func readCompareEntry(store LegacySVStore, name string) ([]byte, error) {
	data, err := store.ReadEntry(name)
	if err != nil {
		return nil, err
	}

	if name != HeroEntryName {
		return data, nil
	}

	password := HeroPassword(SaveNamespace{Spawn: store.Mode.Spawn, Multiplayer: store.Mode.Multiplayer})

	pack, err := DecodeHeroEntry(data, password)
	if err != nil {
		return nil, err
	}

	return pack, nil
}

// detailDiffs runs the memory-map diff for every differing entry
// (pfile.cpp:509-521) and returns the rendered "Diff found in" lines.
func detailDiffs(targets []compareTarget, referenceEntries, actualEntries [][]byte, opts CompareOptions,
) ([]string, error) {
	var lines []string

	for i, target := range targets {
		referenceData, actualData := referenceEntries[i], actualEntries[i]

		if !needsDetailDiff(referenceData, actualData) {
			continue
		}

		diffs, err := diffEntry(opts, target.fileName, target.isTownLevel, referenceData, actualData)
		if err != nil {
			return nil, fmt.Errorf("persist: detail diff %q: %w", target.fileName, err)
		}

		lines = append(lines, renderDiffs(diffs)...)
	}

	return lines, nil
}

// needsDetailDiff reports whether an entry pair is worth a detail walk: both
// absent, or byte-identical, means CompareSaves already skipped it
// (pfile.cpp:497-503).
func needsDetailDiff(referenceData, actualData []byte) bool {
	if len(referenceData) == 0 && len(actualData) == 0 {
		return false
	}

	return string(referenceData) != string(actualData)
}

// renderDiffs formats the foundDiffs map (pfile.cpp:518-520) in sorted key
// order.
func renderDiffs(diffs map[string]int) []string {
	lines := make([]string, 0, len(diffs))

	for _, key := range sortedDiffKeys(diffs) {
		lines = append(lines, fmt.Sprintf("Diff found in %s count: %d", key, diffs[key]))
	}

	return lines
}
