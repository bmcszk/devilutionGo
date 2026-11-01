package quests_test

import (
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
}

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
