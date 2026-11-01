package quests_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/enums"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/quests"
)

// slotWant is the expected quest slot state after a SetMultiQuest call.
type slotWant struct {
	state quests.QuestState
	log   bool
}

// effectWant is the expected completion effect count.
type effectWant struct {
	purify int
	girl   int
	cow    int
}

// SetMultiQuest raises a quest, latches the log flag and only ever grows qvar1
// (quests.cpp:595-607).
func TestSetMultiQuestRaisesState(t *testing.T) {
	tests := []struct {
		name      string
		from      quests.QuestState
		to        quests.QuestState
		log       bool
		wantState quests.QuestState
		wantLog   bool
	}{
		{
			name:      "initToActive",
			from:      quests.QuestInit,
			to:        quests.QuestActive,
			wantState: quests.QuestActive,
		},
		{
			name:      "activeToDone",
			from:      quests.QuestActive,
			to:        quests.QuestDone,
			log:       true,
			wantState: quests.QuestDone,
			wantLog:   true,
		},
		{
			name:      "neverGoesBack",
			from:      quests.QuestActive,
			to:        quests.QuestInit,
			wantState: quests.QuestActive,
		},
		{
			name:      "doneIsFinal",
			from:      quests.QuestDone,
			to:        quests.QuestActive,
			wantState: quests.QuestDone,
		},
		{
			name:      "hiveTeaseYieldsToActive",
			from:      quests.QuestHiveTease2,
			to:        quests.QuestActive,
			wantState: quests.QuestActive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			qs := mustInit(t, enums.SimFlags{})
			qs[4].Active = tt.from
			// when
			quests.SetMultiQuest(qs, quests.MultiQuestUpdate{
				Slot: 4, State: tt.to, Log: tt.log, Var1: 2, Var2: 3, Msg: "MSG",
			}, quests.ResyncContext{}, quests.MultiQuestEffects{})
			// then
			assertSlot(t, qs[4], slotWant{state: tt.wantState, log: tt.wantLog})
		})
	}
}

// The shareware build ignores quest updates (quests.cpp:593-594), and the
// completed-only side effects need QUEST_DONE to arrive (quests.cpp:611-619).
func TestSetMultiQuestGatesAndCompletion(t *testing.T) {
	tests := []struct {
		name      string
		slot      int
		from      quests.QuestState
		qslvl     string
		onLevel   string
		wantPurif int
		wantGirl  int
		wantCow   int
	}{
		{name: "plainUpdateNoEffect"},
		{
			name: "waterCompletedOnItsLevel", slot: 13, from: quests.QuestActive, qslvl: "l3",
			onLevel: "l3", wantPurif: 1,
		},
		{
			name: "waterCompletedElsewhere", slot: 13, from: quests.QuestActive, qslvl: "l3",
			onLevel: "0",
		},
		{
			name: "girlCompletedInTown", slot: 18, from: quests.QuestActive,
			onLevel: "0", wantGirl: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			flags := enums.SimFlags{}
			qs := mustInit(t, flags)
			slot := tt.slot
			qs[slot].Active = tt.from
			qs[slot].QSLvl = tt.qslvl
			rec := &completionRecorder{onLevel: tt.onLevel}
			// when
			quests.SetMultiQuest(qs, quests.MultiQuestUpdate{Slot: slot, State: quests.QuestDone},
				quests.ResyncContext{Flags: flags}, rec.effects())
			// then
			assertEffects(t, rec, effectWant{purify: tt.wantPurif, girl: tt.wantGirl, cow: tt.wantCow})
		})
	}
}

// The shareware build ignores quest updates (quests.cpp:593-594).
func TestSetMultiQuestSpawnGate(t *testing.T) {
	// given
	flags := enums.SimFlags{Spawn: true}
	qs := mustInit(t, flags)
	rec := &completionRecorder{}
	// when
	quests.SetMultiQuest(qs, quests.MultiQuestUpdate{Slot: 4, State: quests.QuestDone},
		quests.ResyncContext{Flags: flags}, rec.effects())
	// then
	if qs[4].Active == quests.QuestDone || rec.synced != 0 {
		t.Fatalf("spawn build applied the update: state %d, resyncs %d", qs[4].Active, rec.synced)
	}
}

// assertSlot pins the state, the log flag and the variables of one slot.
func assertSlot(t *testing.T, q quests.Quest, want slotWant) {
	t.Helper()

	if q.Active != want.state || q.Log != want.log {
		t.Fatalf("state/log = %d/%v, want %d/%v", q.Active, q.Log, want.state, want.log)
	}

	if q.Var1 != 2 || q.Var2 != 3 || q.Msg != "MSG" {
		t.Fatalf("vars = %d/%d/%q, want 2/3/MSG", q.Var1, q.Var2, q.Msg)
	}
}

// assertEffects pins the completion hooks that fired.
func assertEffects(t *testing.T, rec *completionRecorder, want effectWant) {
	t.Helper()

	if rec.purify != want.purify || rec.girl != want.girl || rec.cow != want.cow {
		t.Fatalf("effects = %d/%d/%d, want %d/%d/%d",
			rec.purify, rec.girl, rec.cow, want.purify, want.girl, want.cow)
	}
}
