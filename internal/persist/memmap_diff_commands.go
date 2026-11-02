package persist

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// compareBytes is the compareBytes lambda (Source/pfile.cpp:344-358): it
// compares count bytes at both cursors, advances only the cursors that exist,
// and treats a missing side as "not compared, not different".
func compareBytes(count int, reference, actual *compareInfo) (bool, error) {
	if !reference.remaining(count) {
		return false, fmt.Errorf("comparison failed, not enough bytes in reference: %w", ErrShortRecord)
	}

	if !actual.remaining(count) {
		return false, fmt.Errorf("comparison failed, not enough bytes in actual: %w", ErrShortRecord)
	}

	same := true
	if reference.dataExists && actual.dataExists {
		same = bytes.Equal(reference.data[reference.pos:reference.pos+count],
			actual.data[actual.pos:actual.pos+count])
	}

	if reference.dataExists {
		reference.pos += count
	}

	if actual.dataExists {
		actual.pos += count
	}

	return same, nil
}

// readInt32 reads the read32BitInt lambda (Source/pfile.cpp:360-372) in its
// default big-endian mode: a 4-byte peek at the cursor, not an advance.
func readInt32(info compareInfo) (int32, error) {
	raw, err := peekInt32(info)
	if err != nil {
		return 0, err
	}

	if raw == nil {
		return 0, nil
	}

	return int32(binary.BigEndian.Uint32(raw)), nil //nolint:gosec // save-field bit pattern
}

// readInt32LittleEndian is read32BitInt(compareInfo, useLE=true), the LC_LE
// variant (pfile.cpp:425-426).
func readInt32LittleEndian(info compareInfo) (int32, error) {
	raw, err := peekInt32(info)
	if err != nil {
		return 0, err
	}

	if raw == nil {
		return 0, nil
	}

	return int32(binary.LittleEndian.Uint32(raw)), nil //nolint:gosec // save-field bit pattern
}

// peekInt32 returns the 4 bytes at the cursor, or nil when the entry is absent.
func peekInt32(info compareInfo) ([]byte, error) {
	if !info.dataExists {
		return nil, nil
	}

	if info.pos+4 > info.size {
		return nil, fmt.Errorf("readInt32 failed, fewer than 4 bytes to read: %w", ErrShortRecord)
	}

	return info.data[info.pos : info.pos+4], nil
}

// detailRead handles the R/LT/LC/LC_LE branch (pfile.cpp:409-434): a fixed bit
// width, optionally recording the town flag (LT) or a repetition counter (LC).
func (s diffState) detailRead(prefix string, tokens []string, command string, reference, actual *compareInfo,
	counters map[string]compareCounter,
) error {
	args, err := readTokens(tokens, 2)
	if err != nil {
		return err
	}

	size, err := parseBits(args[0])
	if err != nil {
		return err
	}

	comment := args[1]

	if err := applyReadModifier(command, size, comment, reference, actual, counters); err != nil {
		return err
	}

	same, err := compareBytes(size, reference, actual)
	if err != nil {
		return err
	}

	if !same {
		s.addDiff(prefix + "." + comment)
	}

	return nil
}

// applyReadModifier runs the LT/LC/LC_LE side effect of a read command; plain
// R has none (pfile.cpp:416-430). detailLine only routes those four commands
// here, so an unknown command is a no-op.
func applyReadModifier(command string, size int, comment string, reference, actual *compareInfo,
	counters map[string]compareCounter,
) error {
	switch command {
	case "LT":
		return markTownLevel(size, reference, actual)
	case "LC":
		return recordCounterBigEndian(size, comment, reference, actual, counters)
	case "LC_LE":
		return recordCounterLittleEndian(size, comment, reference, actual, counters)
	default:
		return nil
	}
}

// markTownLevel is the LT case (pfile.cpp:416-423): level type zero means town,
// which gates the _DL lines.
func markTownLevel(size int, reference, actual *compareInfo) error {
	if size != 4 {
		return fmt.Errorf("persist: LT expects 32 bits, got %d: %w", size*8, ErrShortRecord)
	}

	referenceLevel, err := readInt32(*reference)
	if err != nil {
		return err
	}

	actualLevel, err := readInt32(*actual)
	if err != nil {
		return err
	}

	reference.isTownLevel = referenceLevel == 0
	actual.isTownLevel = actualLevel == 0

	return nil
}

// recordCounterBigEndian is the LC case (pfile.cpp:424-430): a field whose
// value becomes the repetition count of a later M or C command.
func recordCounterBigEndian(size int, comment string, reference, actual *compareInfo,
	counters map[string]compareCounter,
) error {
	return recordCounter(size, comment, readInt32, reference, actual, counters)
}

// recordCounterLittleEndian is the LC_LE case (pfile.cpp:425-426).
func recordCounterLittleEndian(size int, comment string, reference, actual *compareInfo,
	counters map[string]compareCounter,
) error {
	return recordCounter(size, comment, readInt32LittleEndian, reference, actual, counters)
}

// recordCounter stores the LC counter pair once both sides are read.
func recordCounter(size int, comment string, read func(compareInfo) (int32, error),
	reference, actual *compareInfo, counters map[string]compareCounter,
) error {
	if size != 4 {
		return fmt.Errorf("persist: LC expects 32 bits, got %d: %w", size*8, ErrShortRecord)
	}

	referenceValue, err := read(*reference)
	if err != nil {
		return err
	}

	actualValue, err := read(*actual)
	if err != nil {
		return err
	}

	counters[comment] = compareCounter{reference: int(referenceValue), actual: int(actualValue)}

	return nil
}

// detailMulti handles the M branch (pfile.cpp:434-449): count repeated fields
// of a fixed width.
func (s diffState) detailMulti(prefix string, tokens []string, reference, actual *compareInfo,
	counters map[string]compareCounter,
) error {
	args, err := readTokens(tokens, 3)
	if err != nil {
		return err
	}

	count, err := parseCounter(args[0], counters)
	if err != nil {
		return err
	}

	size, err := parseBits(args[1])
	if err != nil {
		return err
	}

	return s.diffRepeated(count, size, prefix+"."+args[2], reference, actual)
}

// diffRepeated compares count fields of size bytes and records a diff per
// mismatch (pfile.cpp:444-448).
func (s diffState) diffRepeated(count compareCounter, size int, key string,
	reference, actual *compareInfo,
) error {
	for i := range count.max() {
		count.checkIfDataExists(i, reference, actual)

		same, err := compareBytes(size, reference, actual)
		if err != nil {
			return err
		}

		if !same {
			s.addDiff(key)
		}
	}

	return nil
}

// detailComposite handles the C branch (pfile.cpp:450-463): count repetitions
// of a sub-descriptor, each nesting one level deeper in the diff-key path.
func (s diffState) detailComposite(prefix string, tokens []string, reference, actual *compareInfo,
	counters map[string]compareCounter,
) error {
	args, err := readTokens(tokens, 3)
	if err != nil {
		return err
	}

	count, err := parseCounter(args[0], counters)
	if err != nil {
		return err
	}

	subMap, comment := args[1], args[2]
	subPrefix := prefix + "." + comment

	for i := range count.max() {
		count.checkIfDataExists(i, reference, actual)

		if err := s.detail(subPrefix, subMap, reference, actual); err != nil {
			return fmt.Errorf("persist: sub-map %q: %w", subMap, err)
		}
	}

	return nil
}
