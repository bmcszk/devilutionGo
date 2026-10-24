package testsupport

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Corpus loads a committed gamedata fixture from internal/testdata/gamedata.
// rel is the source-relative txtdata path (e.g.
// "assets/txtdata/monsters/monstdat.tsv" or
// "mods/hf/txtdata/monsters/monstdat.tsv"); Hellfire paths under mods/hf/
// map to the "hf-" prefixed fixture file. Fails the test when the fixture is
// missing (zero-dependency parity contract: never skip).
func Corpus(t *testing.T, pkg, rel string) []byte {
	t.Helper()
	name := filepath.Base(rel)
	if strings.Contains(rel, "mods/hf/") {
		name = "hf-" + name
	}
	data, err := os.ReadFile(filepath.Join(
		"..", "..", "..", "..", "internal", "testdata", "gamedata", pkg, name,
	))
	if err != nil {
		t.Fatalf("read committed %s fixture %s: %v", pkg, name, err)
	}
	return data
}
