// Package audit loads the row-count manifest that pins the committed TSV goldens.
package audit

import (
	"encoding/json"
	"fmt"
	"os"
)

// Entry pins one table variant: its expected row count and its first and last row.
type Entry struct {
	Table    string   `json:"table"`
	Variant  string   `json:"variant"`
	WantRows int      `json:"wantRows"`
	FirstRow []string `json:"firstRow"`
	LastRow  []string `json:"lastRow"`
}

// Manifest is the set of expected table variants.
type Manifest struct {
	Entries []Entry `json:"entries"`
}

// LoadManifest decodes the manifest at path; read and decode failures carry the path.
func LoadManifest(path string) (Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("read manifest %q: %w", path, err)
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return Manifest{}, fmt.Errorf("decode manifest %q: %w", path, err)
	}

	return manifest, nil
}
