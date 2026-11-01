package quests_test

import (
	"fmt"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/enums"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/quests"
)

// killRecorder captures the CheckQuestKill side effects.
type killRecorder struct {
	speech   []string
	sent     []int8
	portal   int
	triggers int
	altBoy   int
	elixir   int
}

// spawnElixir records one SpawnQuestItem(IDI_GLDNELIX, ...) call.
func (r *killRecorder) spawnElixir() { r.elixir++ }

// sendQuest records one NetSendCmdQuest(true, quest) call.
func (r *killRecorder) sendQuest(q quests.Quest) { r.sent = append(r.sent, int8(q.ID)) }

// hooks returns the injected effects that write into the recorder.
func (r *killRecorder) hooks() quests.KillEffects {
	return quests.KillEffects{
		Say:             func(speech string) { r.speech = append(r.speech, speech) },
		SendQuest:       func(q quests.Quest) { r.sent = append(r.sent, int8(q.ID)) },
		LazarusPortal:   func() { r.portal++ },
		LazarusTriggers: func() { r.triggers++ },
	}
}

// mustRows parses the committed questdat fixture or fails the test.
func mustRows(t *testing.T) []quests.QuestRow {
	t.Helper()

	rows, err := quests.ParseQuestdat(questdatFixture(t))
	if err != nil {
		t.Fatalf("quests.ParseQuestdat: %v", err)
	}

	return rows
}

// mustInit initializes the quest slots from the committed table or fails.
func mustInit(t *testing.T, flags enums.SimFlags) []quests.Quest {
	t.Helper()

	got, err := quests.InitQuests(mustRows(t), flags)
	if err != nil {
		t.Fatalf("quests.InitQuests: %v", err)
	}

	return got
}

// assertCleanVars fails when a slot other than Q_LTBANNER carries a variable.
func assertCleanVars(t *testing.T, i int, q quests.Quest) {
	t.Helper()

	if i != 7 && (q.Var1 != 0 || q.Var2 != 0) {
		t.Fatalf("quest %d vars = %d/%d, want 0/0", i, q.Var1, q.Var2)
	}
}

// assertMPQuest pins one multiplayer slot against its table row.
func assertMPQuest(t *testing.T, i int, row quests.QuestRow, q quests.Quest) {
	t.Helper()

	if row.IsSinglePlayerOnly {
		if q.Active != quests.QuestNotAvail || q.Level != 0 {
			t.Fatalf("single player quest %d = state %d level %d, want NOTAVAIL/0",
				i, q.Active, q.Level)
		}

		return
	}

	if q.Active != quests.QuestInit || q.Level != row.QDMultLvl {
		t.Fatalf("multiplayer quest %d = state %d level %d, want INIT/%d",
			i, q.Active, q.Level, row.QDMultLvl)
	}
}

// assertKillEffects pins the speech and the single quest packet of a kill.
func assertKillEffects(t *testing.T, rec *killRecorder, speech string, slot int) {
	t.Helper()

	if len(rec.speech) != 1 || rec.speech[0] != speech {
		t.Fatalf("speech = %v, want %s", rec.speech, speech)
	}

	if len(rec.sent) != 1 || rec.sent[0] != int8(slot) {
		t.Fatalf("sent = %v, want only quest %d", rec.sent, slot)
	}
}

// assertLazarus pins the Q_BETRAYER/Q_DIABLO state after Lazarus dies.
func assertLazarus(t *testing.T, qs []quests.Quest, wantVar2 uint8) {
	t.Helper()

	if qs[15].Active != quests.QuestDone || qs[15].Var1 != 7 {
		t.Fatalf("Q_BETRAYER = state %d var1 %d, want DONE/7", qs[15].Active, qs[15].Var1)
	}

	if qs[15].Var2 != wantVar2 {
		t.Fatalf("Q_BETRAYER var2 = %d, want %d", qs[15].Var2, wantVar2)
	}

	if qs[5].Active != quests.QuestActive {
		t.Fatalf("Q_DIABLO = %d, want QUEST_ACTIVE", qs[5].Active)
	}
}

// assertLazarusEffects pins the branch hook and the two quest packets.
func assertLazarusEffects(t *testing.T, rec *killRecorder, wantPortal, wantTrig int) {
	t.Helper()

	if rec.portal != wantPortal || rec.triggers != wantTrig {
		t.Fatalf("portal/trigger hooks = %d/%d, want %d/%d", rec.portal, rec.triggers, wantPortal, wantTrig)
	}

	if len(rec.sent) != 2 || rec.sent[0] != 15 || rec.sent[1] != 5 {
		t.Fatalf("sent = %v, want Q_BETRAYER then Q_DIABLO", rec.sent)
	}
}

// seedName is the sub-test name for a seed.
func seedName(seed uint32) string {
	return fmt.Sprintf("seed%d", seed)
}

// assertPoolSlots compares the NOTAVAIL slots of a pool against the
// hand-derived list.
func assertPoolSlots(t *testing.T, seed uint32, got []quests.QuestState, want []int8) {
	t.Helper()

	notAvail := make([]int8, 0, len(want))
	for i, state := range got {
		if state == quests.QuestNotAvail {
			notAvail = append(notAvail, int8(i))
		}
	}

	if len(notAvail) != len(want) {
		t.Fatalf("seed %d: NOTAVAIL slots = %v, want %v", seed, notAvail, want)
	}

	for i := range want {
		if notAvail[i] != want[i] {
			t.Fatalf("seed %d: NOTAVAIL slots = %v, want %v", seed, notAvail, want)
		}
	}
}

// poolStates is the quest-state view InitQuests produces after the pools run.
func poolStates(qs []quests.Quest) []quests.QuestState {
	out := make([]quests.QuestState, len(qs))
	for i, q := range qs {
		out[i] = q.Active
	}

	return out
}

// assertSentinelVars pins the var2 = 2 mark on the two sentinel quests when
// they are out of the game (quests.cpp:97-101).
func assertSentinelVars(t *testing.T, qs []quests.Quest) {
	t.Helper()

	for _, slot := range []int{0, 12} { // Q_ROCK, Q_SKELKING
		want := uint8(0)
		if qs[slot].Active == quests.QuestNotAvail {
			want = 2
		}

		if qs[slot].Var2 != want {
			t.Fatalf("slot %d var2 = %d, want %d", slot, qs[slot].Var2, want)
		}
	}
}

// resyncEffects routes the ResyncMPQuests side effects into the recorder.
func resyncEffects(rec *killRecorder) quests.ResyncEffects {
	return quests.ResyncEffects{
		SendQuest:   func(q quests.Quest) { rec.sent = append(rec.sent, int8(q.ID)) },
		SpawnAltBoy: func() { rec.altBoy++ },
	}
}

// assertResyncSent compares the packet order with the expected slot list.
func assertResyncSent(t *testing.T, got, want []int8) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("sent = %v, want %v", got, want)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("sent = %v, want %v", got, want)
		}
	}
}

// assertResyncStates pins that exactly the broadcast quests turned ACTIVE.
func assertResyncStates(t *testing.T, qs []quests.Quest, activated []int8) {
	t.Helper()

	active := make(map[int8]bool, len(activated))
	for _, id := range activated {
		active[id] = true
	}

	for i, q := range qs {
		if active[int8(i)] != (q.Active == quests.QuestActive) {
			t.Fatalf("slot %d = %d, want ACTIVE only for %v", i, q.Active, active[int8(i)])
		}
	}
}

// mustInitStates inits the slots and then forces the given slot states.
func mustInitStates(t *testing.T, flags enums.SimFlags, states map[int8]quests.QuestState) []quests.Quest {
	t.Helper()

	qs := mustInit(t, flags)
	for id, state := range states {
		qs[id].Active = state
	}

	return qs
}
