package entities_test

import (
	"go/build"
	"strings"
	"testing"
)

// TestEntitiesImportLaw makes the layer law machine-checked (AGENTS.md): the
// entity data port may not reach into internal/game or the renderer, which are
// allowed to import entities, never the other way round.
func TestEntitiesImportLaw(t *testing.T) {
	// given the entities package as the compiler sees it
	pkg, err := build.ImportDir(".", 0)
	if err != nil {
		t.Fatalf("import dir: %v", err)
	}

	// when its production imports are inspected
	forbidden := []string{"internal/game", "internal/engine/render", "/render"}

	// then none of them belongs to a forbidden layer
	for _, path := range pkg.Imports {
		for _, bad := range forbidden {
			if strings.Contains(path, bad) {
				t.Fatalf("entities imports forbidden layer %q (import %q)", bad, path)
			}
		}
	}
}
