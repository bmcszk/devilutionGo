package persist

import (
	"fmt"
	"strings"
)

// diffState is the mutable state of one CreateDetailDiffs walk
// (Source/pfile.cpp:322-470): the descriptor directory, the Hellfire flag that
// gates _HF/_DA lines, and the accumulated diffs.
type diffState struct {
	// dir holds the "<name>.txt" memory-map descriptors (pfile.cpp:312).
	dir string
	// hellfire is gbIsHellfire (pfile.cpp:386-397).
	hellfire bool
	// found accumulates diff keys and their counts (pfile.cpp:334-341).
	found map[string]int
}

// addDiff is the addDiff lambda (pfile.cpp:334-341).
func (s diffState) addDiff(key string) {
	s.found[key]++
}

// detail is the recursive body of CreateDetailDiffs (pfile.cpp:307-470): it
// loads one descriptor and walks its lines against the two cursors. prefix is
// the dotted diff-key path of this nesting level.
func (s diffState) detail(prefix, memMapFile string, reference, actual *compareInfo) error {
	buffer, err := readMemoryMap(s.dir, memMapFile)
	if err != nil {
		return err
	}

	counters := make(map[string]compareCounter)
	dataExistsReference, dataExistsActual := reference.dataExists, actual.dataExists

	for _, line := range strings.Split(buffer, "\n") {
		line = strings.TrimSuffix(line, "\r")

		if line == "" {
			continue
		}

		if err := s.detailLine(prefix, line, reference, actual, counters); err != nil {
			return err
		}

		// pfile.cpp:467-468: the _DL toggles are per line.
		reference.dataExists = dataExistsReference
		actual.dataExists = dataExistsActual
	}

	return nil
}

// detailLine dispatches one descriptor line (pfile.cpp:374-466).
func (s diffState) detailLine(prefix, line string, reference, actual *compareInfo,
	counters map[string]compareCounter,
) error {
	tokens := strings.Split(line, " ")

	command, skip := s.applyModeSuffix(tokens[0], reference, actual)
	if skip {
		return nil
	}

	switch command {
	case "R", "LT", "LC", "LC_LE":
		return s.detailRead(prefix, tokens, command, reference, actual, counters)
	case "M":
		return s.detailMulti(prefix, tokens, reference, actual, counters)
	case "C":
		return s.detailComposite(prefix, tokens, reference, actual, counters)
	}

	return nil
}

// applyModeSuffix strips the _HF/_DA/_DL gating suffixes (pfile.cpp:386-408)
// and reports whether the line is skipped in this run.
func (s diffState) applyModeSuffix(command string, reference, actual *compareInfo) (string, bool) {
	switch {
	case endsWith(command, "_HF"):
		return strings.TrimSuffix(command, "_HF"), !s.hellfire
	case endsWith(command, "_DA"):
		return strings.TrimSuffix(command, "_DA"), s.hellfire
	case endsWith(command, "_DL"):
		return applyTownSuffix(command, reference, actual)
	}

	return command, false
}

// applyTownSuffix is the _DL case (pfile.cpp:398-407): the line only applies
// when at most one side is the town level, and the town side stops comparing.
func applyTownSuffix(command string, reference, actual *compareInfo) (string, bool) {
	if reference.isTownLevel && actual.isTownLevel {
		return command, true
	}

	if reference.isTownLevel {
		reference.dataExists = false
	}

	if actual.isTownLevel {
		actual.dataExists = false
	}

	return strings.TrimSuffix(command, "_DL"), false
}

// readTokens pulls n whitespace-separated tokens from a command line.
func readTokens(tokens []string, n int) ([]string, error) {
	if len(tokens) < n+1 {
		return nil, fmt.Errorf("persist: memory map command %q needs %d arguments, has %d: %w",
			tokens[0], n, len(tokens)-1, ErrShortRecord)
	}

	return tokens[1 : n+1], nil
}

// parseBits is the ParseInt<size_t>(bitsAsString)/8 conversion the R, M, LT
// and LC branches share (pfile.cpp:410-415, :440-443).
func parseBits(token string) (int, error) {
	bits, err := parseCounter(token, map[string]compareCounter{})
	if err != nil {
		return 0, err
	}

	if bits.reference%8 != 0 {
		return 0, fmt.Errorf("persist: bit length %d is not a multiple of 8: %w", bits.reference, ErrShortRecord)
	}

	return bits.reference / 8, nil
}
