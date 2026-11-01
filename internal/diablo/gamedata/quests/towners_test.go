package quests_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/enums"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/quests"
)

// The wounded townsman stands in town while the butcher quest runs
// (towners.cpp:736-737).
func TestIsTownerPresentDeadGuy(t *testing.T) {
	tests := []struct {
		name   string
		states map[int8]quests.QuestState
		want   bool
	}{
		{name: "butcherPending", want: true},
		{name: "butcherRolledOut", states: map[int8]quests.QuestState{6: quests.QuestNotAvail}},
		{name: "butcherDone", states: map[int8]quests.QuestState{6: quests.QuestDone}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			qs := mustInitStates(t, enums.SimFlags{}, tt.states)
			// when
			got := quests.IsTownerPresent(quests.TownerDeadGuy, qs, quests.TownerContext{})
			// then
			if got != tt.want {
				t.Fatalf("IsTownerPresent(deadguy) = %v, want %v", got, tt.want)
			}
		})
	}
}

// Lester lives in Hellfire only, and leaves with bCowQuest or the finished hive
// (towners.cpp:738-739).
func TestIsTownerPresentFarmer(t *testing.T) {
	hellfire := enums.SimFlags{Hellfire: true}

	tests := []struct {
		name   string
		flags  enums.SimFlags
		states map[int8]quests.QuestState
		want   bool
	}{
		{name: "hellfire", flags: hellfire, want: true},
		{name: "noHellfire"},
		{name: "cowQuest", flags: enums.SimFlags{Hellfire: true, CowQuest: true}},
		{name: "hiveDone", flags: hellfire, states: map[int8]quests.QuestState{17: quests.QuestHiveDone}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			qs := mustInitStates(t, tt.flags, tt.states)
			// when
			got := quests.IsTownerPresent(quests.TownerFarmer, qs, quests.TownerContext{Flags: tt.flags})
			// then
			if got != tt.want {
				t.Fatalf("IsTownerPresent(farmer) = %v, want %v", got, tt.want)
			}
		})
	}
}

// The cow, Celia and the untouched towners (towners.cpp:740-748).
func TestIsTownerPresentHellfireSwaps(t *testing.T) {
	tests := []struct {
		name   string
		towner string
		flags  enums.SimFlags
		ctx    quests.TownerContext
		want   bool
	}{
		{
			name:   "cowWithCowQuest",
			towner: quests.TownerCowFarm,
			flags:  enums.SimFlags{Hellfire: true, CowQuest: true},
			want:   true,
		},
		{
			name:   "cowWithoutCowQuest",
			towner: quests.TownerCowFarm,
			flags:  enums.SimFlags{Hellfire: true},
		},
		{
			name:   "girlInHellfireAfterLevel17",
			towner: quests.TownerGirl,
			flags:  enums.SimFlags{Hellfire: true, TheoQuest: true},
			ctx:    quests.TownerContext{Visited17: true},
			want:   true,
		},
		{
			name:   "girlBeforeLevel17",
			towner: quests.TownerGirl,
			flags:  enums.SimFlags{Hellfire: true, TheoQuest: true},
		},
		{
			name:   "girlWithoutTheoQuest",
			towner: quests.TownerGirl,
			flags:  enums.SimFlags{Hellfire: true},
			ctx:    quests.TownerContext{Visited17: true},
		},
		{name: "smithAlwaysPresent", towner: "TOWN_SMITH", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			qs := mustInitStates(t, tt.flags, nil)
			ctx := tt.ctx
			ctx.Flags = tt.flags
			// when
			got := quests.IsTownerPresent(tt.towner, qs, ctx)
			// then
			if got != tt.want {
				t.Fatalf("IsTownerPresent(%s) = %v, want %v", tt.towner, got, tt.want)
			}
		})
	}
}
