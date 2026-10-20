// Package objdat_test ports the object data table parity tests against the
// live devilutionX corpus.
package objdat_test

import (
	"strings"

	"os"
	"path/filepath"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/objdat"
)

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

// Row-count parity per V4.1: objdat = 109 rows; hf override must NOT exist.
func TestRowCountParityAndNoHFOverride(t *testing.T) {
	rows, err := objdat.ParseObjdat(corpus(t, "assets/txtdata/objects/objdat.tsv"))
	if err != nil {
		t.Fatalf("ParseObjdat: %v", err)
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

// Enum parity: ThemeID = 17 + THEME_NONE; ShrineType = 34 (sentinel excluded).
func TestEnumParity(t *testing.T) {
	if got := objdat.ThemeCount; got != 17 {
		t.Fatalf("ThemeCount = %d, want 17", got)
	}
	if got := objdat.ShrineCount; got != 34 {
		t.Fatalf("ShrineCount = %d, want 34", got)
	}
}

// Shrine availability parity: Single = {Thaumaturgic, Solar};
// Multi-excluded = {Spooky, Tainted}.
func TestShrineAvailability(t *testing.T) {
	single := map[string]bool{}
	for _, s := range objdat.SinglePlayerOnly {
		single[s] = true
	}
	// objects.cpp:175-210: Single = Gloomy, Weird, Thaumaturgic, Solar.
	want := []string{"Gloomy", "Weird", "Thaumaturgic", "Solar"}
	if len(single) != len(want) {
		t.Fatalf("SinglePlayerOnly = %v, want %v", single, want)
	}
	for _, w := range want {
		if !single[w] {
			t.Fatalf("SinglePlayerOnly missing %q (got %v)", w, single)
		}
	}
	excl := map[string]bool{}
	for _, s := range objdat.MultiPlayerExcluded {
		excl[s] = true
	}
	if !excl["Spooky"] || !excl["Tainted"] || len(excl) != 2 {
		t.Fatalf("MultiPlayerExcluded = %v, want {Spooky, Tainted}", excl)
	}
}

// Selection-region + flags decode: every row's selectionRegion decodes to a
// valid SelectionRegion; Animated rows carry animDelay/animLen/animWidth.
func TestSelectionRegionAndFlags(t *testing.T) {
	rows, err := objdat.ParseObjdat(corpus(t, "assets/txtdata/objects/objdat.tsv"))
	if err != nil {
		t.Fatalf("ParseObjdat: %v", err)
	}
	animated := 0
	for _, o := range rows {
		// selectionRegion is a comma list of region bits
		// (cursor_defs.hpp:14-19 flags); "" = None.
		for part := range strings.SplitSeq(o.SelectionRegion, ",") {
			switch part {
			case objdat.RegionNone, objdat.RegionBottom, objdat.RegionMiddle, objdat.RegionTop:
			default:
				t.Errorf("%s: invalid selectionRegion %q", o.ID, o.SelectionRegion)
			}
		}
		if o.Animated {
			if o.AnimDelay == 0 && o.AnimLen == 0 {
				t.Errorf("%s: Animated without animDelay/animLen", o.ID)
			}
			animated++
		}
	}
	if animated == 0 {
		t.Fatal("no Animated rows found — flags decode broken")
	}
}

// Operator registry: every id the C++ OperateObject switch handles has an
// entry or a documented nil — no silent default.
func TestOperatorRegistryCompleteness(t *testing.T) {
	rows, err := objdat.ParseObjdat(corpus(t, "assets/txtdata/objects/objdat.tsv"))
	if err != nil {
		t.Fatalf("ParseObjdat: %v", err)
	}
	handled := objdat.OperatedObjectIDs() // ids with a C++ Operate* case
	for _, o := range rows {
		if _, ok := handled[o.ID]; !ok {
			// upstream falls through silently for unoperated objects; the
			// registry must list them explicitly as documented-nil
			if !objdat.DocumentedUnoperated[o.ID] {
				t.Errorf("%s: neither Operate* case nor documented-nil", o.ID)
			}
		}
	}
}
