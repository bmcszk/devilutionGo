package assets

import (
	"embed"
	"path/filepath"
	"strings"
)

// nameInventories holds the committed name inventories for the retail archives
// that ship no (listfile) record. MPQ keys record encryption on the file name,
// so without a name a record cannot even be decrypted (names/README.md).
//
//go:embed names/*.txt
var nameInventories embed.FS

// NameInventory returns the committed name inventory for an archive that ships
// no (listfile) record, keyed by archive base name with or without extension.
// ok is false when no inventory was committed for that archive.
func NameInventory(archiveBase string) ([]string, bool) {
	stem := strings.ToLower(strings.TrimSuffix(filepath.Base(archiveBase), filepath.Ext(archiveBase)))

	content, err := nameInventories.ReadFile("names/" + stem + ".txt")
	if err != nil {
		return nil, false
	}

	return splitNames(string(content)), true
}

// splitNames parses an inventory dump: one name per line, CRLF or LF, blank
// lines dropped.
func splitNames(content string) []string {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")

	names := make([]string, 0, len(lines))

	for _, line := range lines {
		if line = strings.TrimSuffix(line, "\r"); line != "" {
			names = append(names, line)
		}
	}

	return names
}
