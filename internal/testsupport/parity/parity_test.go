package parity_test

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/bmcszk/devilutionGo/internal/testsupport/parity"
	"testing"
)

func TestCompareTable_MatchingFirstSpellRow(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "testdata", "gamedata", "goldens", "spelldat-base.json")
	fixture, err := parity.LoadFixture(fixturePath)
	if err != nil {
		t.Fatal(err)
	}

	if diffs := parity.CompareTable(fixture.Rows, fixture); len(diffs) != 0 {
		t.Fatalf("got diffs: %v", diffs)
	}
}

func TestAssertTable_CorruptedFixtureReportsField(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "spelldat-base.json")
	contents := `{"table":"spelldat","variant":"base","probeCommit":"test","columns":["id"],` +
		`"rows":[["wrong"]],"sourceTSV":"spelldat.tsv"}`
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}

	fixture, err := parity.LoadFixture(path)
	if err != nil {
		t.Fatal(err)
	}

	diffs := parity.CompareTable([][]string{{"right"}}, fixture)
	if len(diffs) != 1 || diffs[0] != "row 0 col id: got right, want wrong" {
		t.Fatalf("got diffs %v", diffs)
	}
	if strings.Contains(strings.Join(diffs, "\n"), "rows:") {
		t.Fatalf("got count mismatch in field mismatch: %v", diffs)
	}
}
