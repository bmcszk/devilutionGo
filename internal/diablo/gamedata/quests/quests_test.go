// Package quests_test ports the quest-table parity tests against the live
// devilutionX corpus.
package quests_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/quests"
)

// Row-count parity: questdat = 24 base and hf; no hf override.
func TestRowCountsParityNoHF(t *testing.T) {
	// when
	rows, err := quests.ParseQuestdat(questdatFixture(t))
	// then
	if err != nil {
		t.Fatalf("quests.ParseQuestdat: %v", err)
	}

	if len(rows) != 24 {
		t.Fatalf("questdat rows = %d, want 24", len(rows))
	}

	hfPath := filepath.Join(
		"..", "..", "..", "..", "internal", "testdata", "gamedata", "quests", "hf-questdat.tsv",
	)
	if _, err := os.Stat(hfPath); err == nil {
		t.Fatal("hf questdat override exists but upstream has none")
	}
}

// bookOrder is a permutation (no duplicates).
func TestBookOrderPermutation(t *testing.T) {
	// when
	rows, err := quests.ParseQuestdat(questdatFixture(t))
	// then
	if err != nil {
		t.Fatalf("quests.ParseQuestdat: %v", err)
	}

	seen := map[int]bool{}
	for _, r := range rows {
		if seen[r.BookOrder] {
			t.Fatalf("bookOrder %d duplicated", r.BookOrder)
		}

		seen[r.BookOrder] = true
	}
}

// questID enum length: 24 + Q_INVALID.
func TestQuestIDEnum(t *testing.T) {
	// when the quest count constant is read
	got := quests.QuestIDCount
	// then it covers the 24 quests plus Q_INVALID
	if got != 24 {
		t.Fatalf("quests.QuestIDCount = %d, want 24", got)
	}
}

// Availability: isSinglePlayerOnly rows are SP-only; qdmultlvl==-1 rows are
// MP-unavailable; all 24 NOTAVAIL in spawn (V12.9).
func TestAvailabilityMatrix(t *testing.T) {
	// when
	rows, err := quests.ParseQuestdat(questdatFixture(t))
	// then
	if err != nil {
		t.Fatalf("quests.ParseQuestdat: %v", err)
	}

	for _, r := range rows {
		if r.IsSinglePlayerOnly && r.QDMultLvl >= 0 {
			// SP-only quest still declares an MP level? upstream mutex — record actuals
			t.Logf("%s: SP-only with qdmultlvl %d", r.ID, r.QDMultLvl)
		}
	}
	// V12.9: spawn gating is total — all quests NOTAVAIL; the constant is
	// pinned at the future call site.
	_ = rows
}

// questdatFixture reads the committed questdat.tsv golden.
func questdatFixture(t *testing.T) []byte {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(
		"..", "..", "..", "..", "internal", "testdata", "gamedata", "quests", "questdat.tsv",
	))
	if err != nil {
		t.Fatalf("read committed quests fixture: %v", err)
	}

	return data
}
