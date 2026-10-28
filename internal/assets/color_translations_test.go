package assets_test

import (
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/assets"
)

// colorTranslationTable is text_render.cpp:88-110 in upstream order.
var colorTranslationTable = []colorTranslationSlot{
	{name: "goldui", path: `fonts\goldui.trn`},
	{name: "grayui", path: `fonts\grayui.trn`},
	{name: "golduis", path: `fonts\golduis.trn`},
	{name: "grayuis", path: `fonts\grayuis.trn`},
	{name: "ColorDialogWhite", path: ""},
	{name: "ColorDialogRed", path: ""},
	{name: "yellow", path: `fonts\yellow.trn`},
	{name: "unnamed gap", path: ""},
	{name: "black", path: `fonts\black.trn`},
	{name: "white", path: `fonts\white.trn`},
	{name: "whitegold", path: `fonts\whitegold.trn`},
	{name: "red", path: `fonts\red.trn`},
	{name: "blue", path: `fonts\blue.trn`},
	{name: "orange", path: `fonts\orange.trn`},
	{name: "buttonface", path: `fonts\buttonface.trn`},
	{name: "buttonpushed", path: `fonts\buttonpushed.trn`},
	{name: "gamedialogwhite", path: `fonts\gamedialogwhite.trn`},
	{name: "gamedialogyellow", path: `fonts\gamedialogyellow.trn`},
	{name: "gamedialogred", path: `fonts\gamedialogred.trn`},
}

// colorTranslationSlot is one ColorTranslations row: the upstream path or ""
// for the nullptr slots (text_render.cpp:87-111).
type colorTranslationSlot struct {
	name string
	path string
}

func TestColorTranslationPathsByName(t *testing.T) {
	// given the 19 upstream slots
	slots := colorTranslationTable
	// when each slot is resolved
	resolved := make([]string, len(slots))
	hasTRN := make([]bool, len(slots))
	for i := range slots {
		resolved[i], hasTRN[i] = assets.ColorTranslationPath(i)
	}

	// then every index names its upstream path, or nothing for a null slot
	for i, slot := range slots {
		t.Run(slot.name, func(t *testing.T) {
			if resolved[i] != slot.path || hasTRN[i] != (slot.path != "") {
				t.Fatalf("ColorTranslationPath(%d) = %q, %v, want %q, %v",
					i, resolved[i], hasTRN[i], slot.path, slot.path != "")
			}
		})
	}
}

func TestColorTranslationSlotCounts(t *testing.T) {
	// given the upstream table (text_render.cpp:88-110)
	slots := assets.ColorTranslations()
	// when the slots are counted
	nulls := make([]int, 0, 3)
	filled := 0
	for i, path := range slots {
		if path == "" {
			nulls = append(nulls, i)

			continue
		}

		filled++
	}

	// then the table has 19 slots: 16 paths and exactly 3 nulls at 4, 5, 7
	if len(slots) != 19 {
		t.Fatalf("slots = %d, want 19", len(slots))
	}

	if filled != 16 {
		t.Errorf("non-null slots = %d, want 16", filled)
	}

	if len(nulls) != 3 || nulls[0] != 4 || nulls[1] != 5 || nulls[2] != 7 {
		t.Fatalf("null slots = %v, want [4 5 7]", nulls)
	}
}

func TestColorTranslationNullSlotHasNoTRN(t *testing.T) {
	// given a null slot and an empty asset source
	reg := &assets.TRNRegistry{}
	src := fakeSource{}
	// when the slot is loaded
	trn, ok, err := reg.ColorTranslationTRN(src, 4)
	// then it reports "no TRN" without an error
	if err != nil {
		t.Fatalf("ColorTranslationTRN(4) error = %v, want none", err)
	}

	if ok || trn != (assets.TRN{}) {
		t.Fatalf("ColorTranslationTRN(4) = %v, %v, want zero TRN and false", trn, ok)
	}
}

func TestColorTranslationFixtureSlots(t *testing.T) {
	// given the two committed font TRNs, served under their upstream paths
	red := readHexFixture(t, filepath.Join(gfxFixtureDir, "trn/fonts_red.trn.hex"))
	whitegold := readHexFixture(t, filepath.Join(gfxFixtureDir, "trn/fonts_whitegold.trn.hex"))
	src := fakeSource{
		`fonts\red.trn`:       red,
		`fonts\whitegold.trn`: whitegold,
	}
	reg := &assets.TRNRegistry{}

	tests := []struct {
		name    string
		slot    int
		fixture []byte
	}{
		{name: "red", slot: 11, fixture: red},
		{name: "whitegold", slot: 10, fixture: whitegold},
	}

	// when
	got := make([]assets.TRN, len(tests))
	for i, tt := range tests {
		trn, ok, err := reg.ColorTranslationTRN(src, tt.slot)
		if err != nil || !ok {
			t.Fatalf("%s: ColorTranslationTRN(%d) = ok %v, err %v", tt.name, tt.slot, ok, err)
		}

		got[i] = trn
	}

	// then the loaded table is the committed fixture
	for i, tt := range tests {
		if want := digestTRN(tt.fixture); digestTRN(got[i][:]) != want {
			t.Errorf("%s: trn digest = %s, want %s", tt.name, digestTRN(got[i][:]), want)
		}
	}
}

// digestTRN hashes a translation table.
func digestTRN(trn []byte) string {
	sum := sha256.Sum256(trn)

	return hex.EncodeToString(sum[:])
}
