package objdat_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/objdat"
)

func TestRowCountParityAndNoHFOverride(t *testing.T) {
	// when
	rows, err := objdat.ParseObjdat(corpus(t, "assets/txtdata/objects/objdat.tsv"))
	// then
	if err != nil {
		t.Fatalf("objdat.ParseObjdat: %v", err)
	}

	if len(rows) != 109 {
		t.Fatalf("objdat rows = %d, want 109", len(rows))
	}

	hfPath := filepath.Join(
		"..", "..", "..", "..", "internal", "testdata", "gamedata", "objdat", "hf-objdat.tsv",
	)
	if _, err := os.Stat(hfPath); err == nil {
		t.Fatal("hf objdat override exists but upstream has none — silent drift guard fired")
	}
}

func TestEnumParity(t *testing.T) {
	if got := objdat.ThemeCount; got != 17 {
		t.Fatalf("objdat.ThemeCount = %d, want 17", got)
	}

	if got := objdat.ShrineCount; got != 34 {
		t.Fatalf("objdat.ShrineCount = %d, want 34", got)
	}
}

func TestShrineAvailability(t *testing.T) {
	// given
	want := []string{"Gloomy", "Weird", "Thaumaturgic", "Solar"}
	// when
	single := nameSet(objdat.SinglePlayerOnly)
	excl := nameSet(objdat.MultiPlayerExcluded)
	// then
	if len(single) != len(want) {
		t.Fatalf("objdat.SinglePlayerOnly = %v, want %v", single, want)
	}

	for _, w := range want {
		if !single[w] {
			t.Fatalf("objdat.SinglePlayerOnly missing %q (got %v)", w, single)
		}
	}

	if !excl["Spooky"] || !excl["Tainted"] || len(excl) != 2 {
		t.Fatalf("objdat.MultiPlayerExcluded = %v, want {Spooky, Tainted}", excl)
	}
}

func TestSelectionRegionAndFlags(t *testing.T) {
	// when
	rows := mustParseObjdat(t, corpus(t, "assets/txtdata/objects/objdat.tsv"))
	// then
	animated := assertSelectionRegions(t, rows)
	if animated == 0 {
		t.Fatal("no Animated rows found — flags decode broken")
	}
}

func TestOperatorRegistryCompleteness(t *testing.T) {
	// given
	rows := mustParseObjdat(t, corpus(t, "assets/txtdata/objects/objdat.tsv"))
	// when
	handled := objdat.OperatedObjectIDs()
	// then
	for _, o := range rows {
		if _, ok := handled[o.ID]; !ok {
			if !objdat.DocumentedUnoperated[o.ID] {
				t.Errorf("%s: neither Operate* case nor documented-nil", o.ID)
			}
		}
	}
}

// nameSet builds a membership set from a name slice.
func nameSet(names []string) map[string]bool {
	set := map[string]bool{}
	for _, s := range names {
		set[s] = true
	}

	return set
}

// mustParseObjdat parses the objdat corpus or fails the test.
func mustParseObjdat(t *testing.T, data []byte) []objdat.ObjectData {
	t.Helper()

	rows, err := objdat.ParseObjdat(data)
	if err != nil {
		t.Fatalf("objdat.ParseObjdat: %v", err)
	}

	return rows
}

// assertSelectionRegions validates each row's selectionRegion parts and the
// Animated/animDelay/animLen coupling; returns the animated row count.
func assertSelectionRegions(t *testing.T, rows []objdat.ObjectData) int {
	t.Helper()

	animated := 0

	for _, o := range rows {
		assertRegionParts(t, o)

		if o.Animated {
			if o.AnimDelay == 0 && o.AnimLen == 0 {
				t.Errorf("%s: Animated without animDelay/animLen", o.ID)
			}

			animated++
		}
	}

	return animated
}

// assertRegionParts checks each comma-separated region part is a known region.
func assertRegionParts(t *testing.T, o objdat.ObjectData) {
	t.Helper()

	for part := range strings.SplitSeq(o.SelectionRegion, ",") {
		switch part {
		case objdat.RegionNone, objdat.RegionBottom, objdat.RegionMiddle, objdat.RegionTop:
		default:
			t.Errorf("%s: invalid selectionRegion %q", o.ID, o.SelectionRegion)
		}
	}
}

func corpus(t *testing.T, rel string) []byte {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(
		"..", "..", "..", "..", "internal", "testdata", "gamedata", "objdat", filepath.Base(rel),
	))
	if err != nil {
		t.Fatalf("read committed objdat fixture: %v", err)
	}

	return data
}
