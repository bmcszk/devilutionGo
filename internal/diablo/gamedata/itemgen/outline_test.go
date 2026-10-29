package itemgen_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
)

func TestGetOutlineColor(t *testing.T) {
	tests := []struct {
		name     string
		itemType string
		quality  int
		statFlag bool
		checkReq bool
		want     int
	}{
		{name: "unusable is red", itemType: "Sword", checkReq: true, want: itemgen.IColRed},
		{name: "gold is yellow", itemType: "Gold", statFlag: true, checkReq: true, want: itemgen.IColYellow},
		{
			name: "magic is blue", itemType: "Sword", quality: itemgen.QualityMagic,
			statFlag: true, checkReq: true, want: itemgen.IColBlue,
		},
		{
			name: "unique is yellow", itemType: "Sword", quality: itemgen.QualityUnique,
			statFlag: true, checkReq: true, want: itemgen.IColYellow,
		},
		{
			name: "plain is white", itemType: "Sword", quality: itemgen.QualityNormal,
			statFlag: true, checkReq: true, want: itemgen.IColWhite,
		},
		{name: "unchecked stat flag ignored", itemType: "Sword", want: itemgen.IColWhite},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// when
			got := itemgen.GetOutlineColor(tc.itemType, tc.quality, tc.statFlag, tc.checkReq)

			// then
			if got != tc.want {
				t.Fatalf("colour = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestOutlineColorPaletteBases(t *testing.T) {
	// when
	colours := [4]int{itemgen.IColYellow, itemgen.IColWhite, itemgen.IColBlue, itemgen.IColRed}

	// then
	want := [4]int{197, 245, 181, 229} // PAL16 + 5 (inv.h:78-83, palette.h:33-40)
	for i, got := range colours {
		if got != want[i] {
			t.Fatalf("colour %d = %d, want %d", i, got, want[i])
		}
	}
}
