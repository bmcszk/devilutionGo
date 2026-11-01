package quests_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/enums"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/quests"
)

// StartQuestlog lists open quests flagged for the log first, then the finished
// ones, each block ordered by questBookOrder (quest_log.cpp:85-120).
func TestQuestLogOrder(t *testing.T) {
	// given
	rows := mustRows(t)
	qs := mustInit(t, enums.SimFlags{})

	qs[6].Active, qs[6].Log = quests.QuestActive, true   // Q_BUTCHER, bookOrder 0
	qs[12].Active, qs[12].Log = quests.QuestActive, true // Q_SKELKING, bookOrder 2
	qs[0].Active = quests.QuestActive                    // Q_ROCK, not in the log
	qs[13].Active = quests.QuestDone                     // Q_PWATER, bookOrder 1
	qs[9].Active = quests.QuestDone                      // Q_BLOOD, bookOrder 6
	qs[16].Active = quests.QuestDone                     // Q_GRAVE, bookOrder 17

	// when
	got := quests.QuestLogOrder(qs, rows)
	// then
	wantOrder := []int8{6, 12, 13, 9, 16}
	if len(got.Order) != len(wantOrder) {
		t.Fatalf("log order = %v, want %v", got.Order, wantOrder)
	}

	for i, want := range wantOrder {
		if int(int8(got.Order[i])) != int(want) {
			t.Fatalf("log entry %d = %d, want %d (order %v)", i, got.Order[i], want, got.Order)
		}
	}

	if got.FirstFinished != 2 {
		t.Fatalf("FirstFinished = %d, want 2", got.FirstFinished)
	}
}
