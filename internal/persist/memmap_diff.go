package persist

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// compareCounter is CompareCounter (pfile.cpp:284-296): the reference and
// actual counts of a repeated sub-map, so entries may differ in length.
type compareCounter struct {
	reference int
	actual    int
}

// compareInfo is CompareInfo (Source/pfile.cpp:276-282): a cursor over one
// archive entry plus the two flags the memory map toggles as it walks.
type compareInfo struct {
	// data is the decoded entry; nil when the entry is absent.
	data []byte
	// pos is currentPosition.
	pos int
	// size is size.
	size int
	// isTownLevel is toggled by the LT command (pfile.cpp:428-432).
	isTownLevel bool
	// dataExists is cleared by _DL handling and by exhausted counters.
	dataExists bool
}

// remaining reports whether read may advance by n more bytes.
func (c compareInfo) remaining(n int) bool {
	return !c.dataExists || c.pos+n <= c.size
}

// max is CompareCounter::max (pfile.cpp:289-292).
func (c compareCounter) max() int {
	if c.reference > c.actual {
		return c.reference
	}

	return c.actual
}

// checkIfDataExists is CompareCounter::checkIfDataExists (pfile.cpp:293-297).
func (c compareCounter) checkIfDataExists(count int, reference, actual *compareInfo) {
	if c.reference == count {
		reference.dataExists = false
	}

	if c.actual == count {
		actual.dataExists = false
	}
}

// sortedDiffKeys returns the found-diff keys in the order the segmented map
// iterates them (pfile.cpp:517-520).
func sortedDiffKeys(diffs map[string]int) []string {
	keys := make([]string, 0, len(diffs))
	for key := range diffs {
		keys = append(keys, key)
	}

	slices.Sort(keys)

	return keys
}

// diffEntry is CreateDetailDiffs (Source/pfile.cpp:307-470) for one archive
// entry: it walks the entry's memory-map descriptor and counts the fields that
// differ. An absent descriptor is an error, mirroring the app_fatal at
// pfile.cpp:316.
func diffEntry(opts CompareOptions, memMapFile string, isTownLevel bool, reference, actual []byte,
) (map[string]int, error) {
	state := diffState{
		dir:      opts.MemMapDir,
		hellfire: opts.Hellfire,
		found:    make(map[string]int),
	}
	ref := compareInfo{data: reference, size: len(reference), isTownLevel: isTownLevel, dataExists: true}
	act := compareInfo{data: actual, size: len(actual), isTownLevel: isTownLevel, dataExists: true}

	if err := state.detail(memMapFile, memMapFile, &ref, &act); err != nil {
		return nil, err
	}

	// pfile.cpp:520-523: uncompared trailing bytes are a hard failure.
	if ref.pos != ref.size {
		return nil, fmt.Errorf("comparison failed, uncompared bytes in reference (%d of %d). File: %s: %w",
			ref.pos, ref.size, memMapFile, ErrCompareUncompared)
	}

	if act.pos != act.size {
		return nil, fmt.Errorf("comparison failed, uncompared bytes in actual (%d of %d). File: %s: %w",
			act.pos, act.size, memMapFile, ErrCompareUncompared)
	}

	return state.found, nil
}

// readMemoryMap loads a "<name>.txt" descriptor from dir (pfile.cpp:310-321).
func readMemoryMap(dir, name string) (string, error) {
	path := filepath.Join(dir, name+".txt")

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("persist: memory map %s missing: %w", path, err)
	}

	return string(data), nil
}

// parseCounter resolves a map token to a counter, using the per-run overrides
// that the LC command installs (pfile.cpp:326-336).
func parseCounter(token string, counters map[string]compareCounter) (compareCounter, error) {
	if c, ok := counters[token]; ok {
		return c, nil
	}

	value, err := strconv.Atoi(token)
	if err != nil {
		return compareCounter{}, fmt.Errorf("persist: failed to parse %q as int: %w", token, err)
	}

	return compareCounter{reference: value, actual: value}, nil
}

// endsWith reports string_ends_with (pfile.cpp:298-303).
func endsWith(value, suffix string) bool {
	return strings.HasSuffix(value, suffix)
}
