package quests_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/enums"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/quests"
)

// Quest::IsAvailable needs the player on the quest level, outside a set level,
// with the quest in the game and allowed by the single player only rule
// (questdat.cpp:64-74).
func TestQuestIsAvailable(t *testing.T) {
	rows := mustRows(t)
	sp := mustInit(t, enums.SimFlags{})

	tests := []struct {
		name    string
		slot    int
		level   quests.LevelContext
		flags   enums.SimFlags
		want    bool
		because string
	}{
		{
			name:    "onQuestLevel",
			slot:    6, // Q_BUTCHER, qdlvl 2
			level:   quests.LevelContext{Level: rows[6].QDLvl},
			want:    true,
			because: "level, state and single player rule all allow it",
		},
		{
			name:    "otherLevel",
			slot:    6,
			level:   quests.LevelContext{Level: rows[6].QDLvl + 1},
			want:    false,
			because: "currlevel must equal _qlevel",
		},
		{
			name:    "setLevel",
			slot:    6,
			level:   quests.LevelContext{Level: rows[6].QDLvl, SetLevel: true},
			want:    false,
			because: "set levels short-circuit IsAvailable",
		},
		{
			name:    "singlePlayerOnlyQuestInMultiplayer",
			slot:    0, // Q_ROCK is single player only
			level:   quests.LevelContext{Level: rows[0].QDLvl},
			flags:   enums.SimFlags{Multiplayer: true},
			want:    false,
			because: "isSinglePlayerOnly is hidden under UseMultiplayerQuests",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// when
			got := sp[tt.slot].IsAvailable(rows[tt.slot], tt.level, tt.flags)
			// then
			if got != tt.want {
				t.Fatalf("IsAvailable() = %v, want %v: %s", got, tt.want, tt.because)
			}
		})
	}
}

// A quest rolled out of the pools never triggers, whatever the level
// (questdat.cpp:69-70).
func TestQuestIsAvailableRejectsNotAvail(t *testing.T) {
	// given
	rows := mustRows(t)
	qs := mustInit(t, enums.SimFlags{})
	qs[6].Active = quests.QuestNotAvail
	// when
	got := qs[6].IsAvailable(rows[6], quests.LevelContext{Level: rows[6].QDLvl}, enums.SimFlags{})
	// then
	if got {
		t.Fatal("QUEST_NOTAVAIL quest must not be available")
	}
}
