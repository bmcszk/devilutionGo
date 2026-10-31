package itemgen_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
)

func TestUniqueLevelUsesIblvl(t *testing.T) {
	tests := []struct {
		name     string
		uper     int
		wantName string
	}{
		// iblvl = lvl+4 = 14 under uper15, so MinLevel 12 is reachable...
		{name: "inside the uper15 window", uper: 15, wantName: "Inside"},
		// ...while uper1 keeps iblvl at 10, below the unique's MinLevel.
		{name: "above the plain level", uper: 1, wantName: ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			items, uniques := uniqueLevelTables()
			req := itemgen.SetupRequest{ItemIndex: 0, Seed: 10, Level: 10, Uper: tc.uper}
			rnd := &tapeRnd{t: t, values: []int32{0, 0, 0, 0, 0, 0, 0, 0}}

			// when
			plan, err := itemgen.SetupAllItems(items, uniques, nil, nil, req, rnd.rnd, func(int) {})
			// then
			if err != nil {
				t.Fatalf("SetupAllItems: %v", err)
			}
			assertUniqueWindow(t, plan, tc.wantName)
		})
	}
}

func assertUniqueWindow(t *testing.T, plan itemgen.SetupPlan, wantName string) {
	t.Helper()

	if wantName == "" {
		if plan.Quality == itemgen.QualityUnique {
			t.Fatalf("plan = %+v, want no unique above the iblvl window", plan)
		}

		return
	}
	if plan.Name != wantName || plan.Quality != itemgen.QualityUnique {
		t.Fatalf("plan = %+v, want unique %q inside the iblvl window", plan, wantName)
	}
}

// uniqueLevelTables builds a unique whose MinLevel sits inside the uper15
// window (lvl, lvl+4] and one above it, both on the same base item.
func uniqueLevelTables() ([]itemdat.ItemData, []itemdat.UniqueItemData) {
	items := []itemdat.ItemData{
		{ID: "CLEAVER", Name: "Cleaver", ItemType: "Sword", Durability: 100},
	}
	uniques := []itemdat.UniqueItemData{
		{Name: "Inside", UniqueBaseItem: "CLEAVER", MinLevel: 12, Value: 100},
	}

	return items, uniques
}
