package gameplay_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/gameplay"
)

// shimLadder is a two-level ladder for the shim smoke test.
type shimLadder []uint32

func (l shimLadder) MaxLevel() int { return len(l) }

func (l shimLadder) ThresholdForLevel(level int) uint32 {
	if level <= 0 {
		return 0
	}

	if level > len(l) {
		return l[len(l)-1]
	}

	return l[level-1]
}

// The gameplay shim resolves the quest API the contract names: a single player
// init through gameplay.InitQuests marks every slot QUEST_INIT
// (quests.cpp:79-82).
func TestQuestsShimInitQuests(t *testing.T) {
	// given
	rows := []gameplay.QuestRow{
		{ID: "0", QDLvl: 5, QDMultLvl: -1},
		{ID: "1", QDLvl: 9, QDMultLvl: -1},
	}
	// when
	_, err := gameplay.InitQuests(rows, gameplay.SimFlags{})
	// then
	if err == nil {
		t.Fatal("two-row table = nil error, want the row-count error")
	}
}

// Thresholds and the award rules resolve through the shim too.
func TestQuestsShimAddExperience(t *testing.T) {
	// given
	var ladder gameplay.Thresholds = shimLadder{1000, 3000}
	// when
	got := gameplay.AddExperience(gameplay.Progress{Level: 1},
		gameplay.Award{Experience: 500, Alive: true}, ladder, gameplay.SimFlags{}, nil)
	// then
	if got.Experience != 500 || got.Level != 1 {
		t.Fatalf("progress = %d/%d, want 500/1", got.Experience, got.Level)
	}
}
