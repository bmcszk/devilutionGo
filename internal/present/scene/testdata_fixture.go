package scene

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// committedFixtureRoot is the seam the screen layer resolves the repo's
// internal/testdata corpus through (review F4): the production default walks
// up from the executable's working directory to the repo root, tests
// override it. No caller may build a CWD-relative "../.." path itself.
var committedFixtureRoot = defaultFixtureRoot

// ReadCommittedFixture reads one committed txtdata fixture from the repo's
// internal/testdata corpus: the producer-side bytes the typed loaders consume,
// so a missing file fails the caller rather than skipping. class names the
// per-class dir, or "classes" for the classdat.tsv table itself.
func ReadCommittedFixture(class, name string) ([]byte, error) {
	root := committedFixtureRoot()
	if root == "" {
		return nil, errors.New("read committed fixture: no internal/testdata root found")
	}

	// "classes" names the classdat.tsv table itself, which sits beside the
	// per-class dirs under upstream/classes/ (the corpus layout).
	rel := []string{"assets", "txtdata", "upstream", "classes"}
	if class != "classes" {
		rel = append(rel, class)
	}

	rel = append(rel, name)

	data, err := os.ReadFile(filepath.Join(append([]string{root}, rel...)...))
	if err != nil {
		return nil, fmt.Errorf("read committed fixture %s/%s: %w", class, name, err)
	}

	return data, nil
}

// defaultFixtureRoot walks upward from the working directory until it finds
// the repo's internal/testdata corpus, so the wired boot resolves fixtures
// from any working directory (the review's F4 seam). Unresolved stays "".
func defaultFixtureRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for range 8 {
		if _, serr := os.Stat(filepath.Join(dir, "internal", "testdata")); serr == nil {
			return filepath.Join(dir, "internal", "testdata")
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}

		dir = parent
	}

	return ""
}
