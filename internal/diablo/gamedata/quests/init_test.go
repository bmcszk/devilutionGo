package quests_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/enums"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/quests"
)

// InitQuests in single player copies each row's qdlvl and marks every quest
// QUEST_INIT (quests.cpp:79-82).
func TestInitQuestsSinglePlayer(t *testing.T) {
	// given
	rows := mustRows(t)
	// when
	got, err := quests.InitQuests(rows, enums.SimFlags{})
	// then
	if err != nil {
		t.Fatalf("quests.InitQuests: %v", err)
	}

	if len(got) != len(rows) {
		t.Fatalf("slots = %d, want %d", len(got), len(rows))
	}

	for i, row := range rows {
		if got[i].Active != quests.QuestInit {
			t.Fatalf("quest %d state = %d, want QUEST_INIT", i, got[i].Active)
		}

		if got[i].Level != row.QDLvl {
			t.Fatalf("quest %d level = %d, want qdlvl %d", i, got[i].Level, row.QDLvl)
		}
	}
}

// Quest IDs are the row positions: QuestsData is emplaced in file order and
// InitQuests walks both in lockstep (quests.cpp:67-71).
func TestInitQuestsIDsAreRowOrder(t *testing.T) {
	// when
	got, err := quests.InitQuests(mustRows(t), enums.SimFlags{})
	// then
	if err != nil {
		t.Fatalf("quests.InitQuests: %v", err)
	}

	want := []struct {
		slot int
		id   int8
	}{
		{slot: 0, id: 0},   // Q_ROCK
		{slot: 6, id: 6},   // Q_BUTCHER
		{slot: 15, id: 15}, // Q_BETRAYER
		{slot: 23, id: 23}, // Q_JERSEY
	}
	for _, w := range want {
		if int(int8(got[w.slot].ID)) != int(w.id) {
			t.Fatalf("slot %d id = %d, want %d", w.slot, got[w.slot].ID, w.id)
		}
	}
}

// InitQuests copies the static table columns into every slot
// (quests.cpp:72-77).
func TestInitQuestsCopiesTableColumns(t *testing.T) {
	// given
	rows := mustRows(t)
	// when
	got, err := quests.InitQuests(rows, enums.SimFlags{})
	// then
	if err != nil {
		t.Fatalf("quests.InitQuests: %v", err)
	}

	for i, row := range rows {
		if got[i].LvlType != row.QLvlT || got[i].QSLvl != row.QSLvl || got[i].Msg != row.QDMsg {
			t.Fatalf("quest %d columns = %q/%q/%q, want %q/%q/%q", i,
				got[i].LvlType, got[i].QSLvl, got[i].Msg, row.QLvlT, row.QSLvl, row.QDMsg)
		}

		if got[i].Log {
			t.Fatalf("quest %d starts with log %v, want false", i, got[i].Log)
		}
	}
}

// The block after the slot loop seeds the quest-specific starting variables
// (quests.cpp:97-105): only Q_LTBANNER starts one step in.
func TestInitQuestsInitialVars(t *testing.T) {
	// when
	sp := mustInit(t, enums.SimFlags{})
	// then
	if sp[7].Var1 != 1 || sp[7].Var2 != 0 {
		t.Fatalf("Q_LTBANNER vars = %d/%d, want 1/0", sp[7].Var1, sp[7].Var2)
	}

	for i, q := range sp {
		assertCleanVars(t, i, q)
	}
}

// An unavailable sentinel quest carries var2 = 2; a quest that is still in the
// game keeps var2 = 0 (quests.cpp:97-101).
func TestInitQuestsSentinelVars(t *testing.T) {
	// when the single player only quests drop out of a multiplayer game
	mp := mustInit(t, enums.SimFlags{Multiplayer: true})
	// then

	if mp[0].Active != quests.QuestNotAvail || mp[0].Var2 != 2 { // Q_ROCK
		t.Fatalf("Q_ROCK = state %d var2 %d, want QUEST_NOTAVAIL/2", mp[0].Active, mp[0].Var2)
	}

	if mp[12].Active != quests.QuestInit || mp[12].Var2 != 0 { // Q_SKELKING
		t.Fatalf("Q_SKELKING = state %d var2 %d, want QUEST_INIT/0", mp[12].Active, mp[12].Var2)
	}
}

// Multiplayer quests take qdmultlvl and skip the single player only rows,
// which stay NOTAVAIL at level 0 (quests.cpp:79-87).
func TestInitQuestsMultiplayer(t *testing.T) {
	// given
	rows := mustRows(t)
	// when
	got := mustInit(t, enums.SimFlags{Multiplayer: true})
	// then
	for i, row := range rows {
		assertMPQuest(t, i, row, got[i])
	}
}

// The shareware build starts with every quest unavailable, after the pools and
// the fixed variables were applied (quests.cpp:107-110).
func TestInitQuestsSpawnGatesEverything(t *testing.T) {
	// when
	got := mustInit(t, enums.SimFlags{Spawn: true})
	// then

	for i, q := range got {
		if q.Active != quests.QuestNotAvail {
			t.Fatalf("spawn quest %d = %d, want QUEST_NOTAVAIL", i, q.Active)
		}
	}
}

// A table that is not one row per quest is rejected instead of silently
// shifting ids (QuestsData is indexed by quest id).
func TestInitQuestsRejectsShortTable(t *testing.T) {
	// given
	rows := mustRows(t)[:len(mustRows(t))-1]
	// when
	_, err := quests.InitQuests(rows, enums.SimFlags{})
	// then
	if err == nil {
		t.Fatal("InitQuests with a short table = nil error, want error")
	}
}
