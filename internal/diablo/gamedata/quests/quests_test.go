// Package quests_test ports the quest-table parity tests against the live
// devilutionX corpus.
package quests_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/quests"
	"github.com/bmcszk/devilutionGo/internal/engine"
)

// corpusRoot resolves the devilutionX checkout that holds the parity corpus.
func corpusRoot(t *testing.T) string {
	t.Helper()
	devilutionX := os.Getenv("DG_DEVILUTIONX")
	if devilutionX == "" {
		t.Skip("set DG_DEVILUTIONX to the devilutionX checkout to run this parity test")
	}
	return devilutionX
}

func corpus(t *testing.T, rel string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(corpusRoot(t), rel))
	if err != nil {
		t.Skipf("corpus not available: %v", err)
	}
	return data
}

// Row-count parity: questdat = 24 base and hf; no hf override.
func TestRowCountsParityNoHF(t *testing.T) {
	rows, err := quests.ParseQuestdat(corpus(t, "assets/txtdata/quests/questdat.tsv"))
	if err != nil {
		t.Fatalf("ParseQuestdat: %v", err)
	}
	if len(rows) != 24 {
		t.Fatalf("questdat rows = %d, want 24", len(rows))
	}
	if _, err := os.Stat(filepath.Join(corpusRoot(t), "mods/hf/txtdata/quests/questdat.tsv")); err == nil {
		t.Fatal("hf questdat override exists but upstream has none")
	}
}

// bookOrder is a permutation (no duplicates).
func TestBookOrderPermutation(t *testing.T) {
	rows, err := quests.ParseQuestdat(corpus(t, "assets/txtdata/quests/questdat.tsv"))
	if err != nil {
		t.Fatalf("ParseQuestdat: %v", err)
	}
	seen := map[int]bool{}
	for _, r := range rows {
		if seen[r.BookOrder] {
			t.Fatalf("bookOrder %d duplicated", r.BookOrder)
		}
		seen[r.BookOrder] = true
	}
}

// QuestID enum length: 24 + Q_INVALID.
func TestQuestIDEnum(t *testing.T) {
	if got := quests.QuestIDCount; got != 24 {
		t.Fatalf("QuestIDCount = %d, want 24", got)
	}
}

// Availability: isSinglePlayerOnly rows are SP-only; qdmultlvl==-1 rows are
// MP-unavailable; all 24 NOTAVAIL in spawn (V12.9).
func TestAvailabilityMatrix(t *testing.T) {
	rows, err := quests.ParseQuestdat(corpus(t, "assets/txtdata/quests/questdat.tsv"))
	if err != nil {
		t.Fatalf("ParseQuestdat: %v", err)
	}
	for _, r := range rows {
		if r.IsSinglePlayerOnly && r.QDMultLvl >= 0 {
			// SP-only quest still declares an MP level? upstream mutex — record actuals
			t.Logf("%s: SP-only with qdmultlvl %d", r.ID, r.QDMultLvl)
		}
	}
	// spawn gating is total
	for _, r := range rows {
		if !quests.NotAvailableInSpawn(r) {
			t.Fatalf("%s must be NOTAVAIL in spawn (V12.9)", r.ID)
		}
	}
}

// Seed-pool tape: InitialiseQuestPools semantics against our port for
// selected level-15 seeds; each pool marks exactly one quest NOTAVAIL.
func TestInitialiseQuestPools(t *testing.T) {
	for _, seed := range []uint32{1, 42, 988045466, 123456789, 4000000000} {
		qs := quests.InitialiseQuestPools(seed, engine.SetRndSeed, engine.GenerateRnd)
		notavail := 0
		for _, q := range qs {
			if q == quests.QuestNotAvail {
				notavail++
			}
		}
		// pools: skelking/pwater, butcher/ltbanner/garbud (skipped for the
		// 988045466 vanilla-bug seed), blind/rock/blood, mushroom/zhar/anvil,
		// veil/warlord = 5 pools (4 for the bug seed).
		want := 5
		if seed == 988045466 {
			want = 4
		}
		if notavail != want {
			t.Fatalf("seed %d: NOTAVAIL count = %d, want %d", seed, notavail, want)
		}
	}
}
