//go:build windowed

package main

import (
	"os"
	"path/filepath"
)

// readCommittedFixture reads one committed txtdata fixture from the repo's
// internal/testdata corpus: the producer-side bytes the typed loaders consume,
// so a missing file fails the test rather than skipping. class names the
// per-class dir, or "classes" for the classdat.tsv table itself.
func readCommittedFixture(class, name string) ([]byte, error) {
	return os.ReadFile(filepath.Join(
		"..", "..", "internal", "testdata", "assets", "txtdata", "upstream", "classes", class, name,
	))
}
