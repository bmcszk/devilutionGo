package itemgen_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
)

type windowCase struct {
	name       string
	createInfo int
	miscID     string
	values     []int32
	want       itemgen.LevelWindow
}

type windowCases []windowCase

func TestItemLevelWindowVendorBranches(t *testing.T) {
	tests := windowCases{
		{
			name:       "griswold premium",
			createInfo: 10 | itemgen.CreateInfoSmithPremium,
			want:       itemgen.LevelWindow{MinLevel: 5, MaxLevel: 10, Iblvl: 10},
		},
		{
			name:       "wirt doubles",
			createInfo: 10 | itemgen.CreateInfoBoy,
			want:       itemgen.LevelWindow{MinLevel: 10, MaxLevel: 20, Iblvl: 20},
		},
		{
			name:       "adria hits five percent",
			createInfo: 10 | itemgen.CreateInfoWitch,
			values:     []int32{5},
			want:       itemgen.LevelWindow{MinLevel: 10, MaxLevel: 20, Iblvl: 20},
		},
		{
			name:       "adria staff upgrade",
			createInfo: 10 | itemgen.CreateInfoWitch,
			miscID:     "STAFF",
			values:     []int32{6},
			want:       itemgen.LevelWindow{MinLevel: 10, MaxLevel: 20, Iblvl: 20},
		},
		{
			name:       "adria miss",
			createInfo: 10 | itemgen.CreateInfoWitch,
			values:     []int32{6},
			want:       itemgen.LevelWindow{MinLevel: 0, MaxLevel: -1, Iblvl: -1},
		},
		{
			name:       "default onlygood path",
			createInfo: 40 | itemgen.CreateInfoOnlyGood,
			values:     []int32{11, 41},
			want:       itemgen.LevelWindow{MinLevel: 20, MaxLevel: 40, Iblvl: 40},
		},
		{
			name:       "default collapsed",
			createInfo: 10,
			values:     []int32{11, 11},
			// -1/2 truncates to 0 in C++, so the minimum is 0, not -1.
			want: itemgen.LevelWindow{MinLevel: 0, MaxLevel: -1, Iblvl: -1},
		},
		{
			name:       "cap to twenty five",
			createInfo: 60,
			values:     []int32{1},
			want:       itemgen.LevelWindow{MinLevel: 25, MaxLevel: 60, Iblvl: 60},
		},
		{
			name:       "uper15 bumps four",
			createInfo: 40 | itemgen.CreateInfoUper15,
			values:     []int32{11, 41},
			want:       itemgen.LevelWindow{MinLevel: 22, MaxLevel: 44, Iblvl: 44},
		},
	}
	// when
	failures := tests.run(t)

	// then
	if failures != 0 {
		t.Fatalf("%d window cases failed", failures)
	}
}

func (cases windowCases) run(t *testing.T) (failures int) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			rnd := &tapeRnd{t: t, values: tc.values}

			// when
			got := itemgen.ItemLevelWindow(tc.createInfo, tc.miscID, rnd.rnd)

			// then
			if got != tc.want {
				t.Errorf("window = %+v, want %+v", got, tc.want)
				failures++
			}
		})
	}

	return failures
}

func TestAffixTypeForItemType(t *testing.T) {
	tests := []struct {
		itemType string
		want     itemgen.AffixItemType
		wantOK   bool
	}{
		{"Sword", itemgen.AffixWeapon, true},
		{"Axe", itemgen.AffixWeapon, true},
		{"Mace", itemgen.AffixWeapon, true},
		{"Bow", itemgen.AffixBow, true},
		{"Shield", itemgen.AffixShield, true},
		{"LightArmor", itemgen.AffixArmor, true},
		{"Helm", itemgen.AffixArmor, true},
		{"MediumArmor", itemgen.AffixArmor, true},
		{"HeavyArmor", itemgen.AffixArmor, true},
		{"Staff", itemgen.AffixStaff, true},
		{"Ring", itemgen.AffixMisc, true},
		{"Amulet", itemgen.AffixMisc, true},
		{"Misc", 0, false},
		{"Gold", 0, false},
		{"None", 0, false},
		{"Bogus", 0, false},
	}
	for _, tc := range tests {
		t.Run(tc.itemType, func(t *testing.T) {
			// when
			got, ok := itemgen.AffixTypeForItemType(tc.itemType)

			// then
			if got != tc.want || ok != tc.wantOK {
				t.Fatalf("type = %v ok = %v, want %v/%v", got, ok, tc.want, tc.wantOK)
			}
		})
	}
}
