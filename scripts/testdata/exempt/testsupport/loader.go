// Package testsupport is the exemption control: test scaffolding may open the
// committed goldens, so the gate must stay silent here.
package testsupport

import (
	"fmt"
	"os"
)

// Load reads the fixture named by path.
func Load(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load %s: %w", path, err)
	}

	return data, nil
}
