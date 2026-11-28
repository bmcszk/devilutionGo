package testsupport

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// DecodeXXD decodes a committed text hex fixture into the bytes it encodes.
//
// Two layouts are accepted, both used by the committed fixtures: an xxd dump
// ("OFFSET: 0011 2233 ...  ascii", the layout `xxd` writes) and bare hex lines
// of whitespace-separated bytes (internal/persist/testdata/
// sv_hero_entry_single.hex). Lines starting with '#' and blank lines are
// provenance comments. A malformed line fails rather than yielding silently
// short output: a fixture that does not decode is a defect, not a skip.
func DecodeXXD(data []byte) ([]byte, error) {
	var raw []byte

	for i, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		hexBody := xxdHexColumn(trimmed)

		decoded, err := hex.DecodeString(strings.Join(strings.Fields(hexBody), ""))
		if err != nil {
			return nil, fmt.Errorf("hex dump line %d (%q): %w", i+1, trimmed, err)
		}

		raw = append(raw, decoded...)
	}

	return raw, nil
}

// xxdHexColumn returns the hex part of one dump line: everything after the
// "OFFSET:" prefix, cut before any "  " separator that opens xxd's ASCII
// column (which may itself contain hex-looking four-letter words).
func xxdHexColumn(line string) string {
	if _, after, ok := strings.Cut(line, ":"); ok {
		if hexPart, _, found := strings.Cut(after, "  "); found {
			return strings.TrimSpace(hexPart)
		}

		return strings.TrimSpace(after)
	}

	return line
}
