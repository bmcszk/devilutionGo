package quests_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/enums"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/quests"
)

// Leaving a quest level returns to the level the quest lives on; any other
// special level returns to town (quests.cpp:289-303).
func TestMapReturnLevel(t *testing.T) {
	tests := []struct {
		name     string
		setLevel string
		want     int
	}{
		{name: "skeletonKing", setLevel: quests.SLSkelking, want: 3},
		{name: "boneChamber", setLevel: quests.SLBonechamb, want: 6},
		{name: "poisonWater", setLevel: quests.SLPoisonwater, want: 2},
		{name: "vileBetrayer", setLevel: quests.SLVilebetrayer, want: 15},
		{name: "none", setLevel: quests.SLNone, want: 0},
		{name: "unknown", setLevel: "SL_OTHER", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			qs := mustInit(t, enums.SimFlags{})
			// when
			got := quests.MapReturnLevel(qs, tt.setLevel)
			// then
			if got != tt.want {
				t.Fatalf("MapReturnLevel(%s) = %d, want %d", tt.setLevel, got, tt.want)
			}
		})
	}
}
