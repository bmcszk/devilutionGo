package quests_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/enums"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/quests"
)

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
