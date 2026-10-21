package parity

import (
	"encoding/json"
	"fmt"
	"os"
)

// The JSON tags are the committed golden file format; renaming one breaks the
// committed fixtures.
type Fixture struct {
	Table       string     `json:"table"`
	Variant     string     `json:"variant"`
	ProbeCommit string     `json:"probeCommit"`
	Columns     []string   `json:"columns"`
	Rows        [][]string `json:"rows"`
	SourceTSV   string     `json:"sourceTSV"`
}

// Missing or malformed fixtures error out: parity tests must fail, not skip.
func LoadFixture(path string) (Fixture, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Fixture{}, fmt.Errorf("read fixture %q: %w", path, err)
	}

	var fixture Fixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		return Fixture{}, fmt.Errorf("decode fixture %q: %w", path, err)
	}
	return fixture, nil
}
