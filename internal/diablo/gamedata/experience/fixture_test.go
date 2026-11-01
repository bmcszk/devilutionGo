package experience_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/enums"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/experience"
	"github.com/bmcszk/devilutionGo/internal/diablo/txtdata"
	enginetxtdata "github.com/bmcszk/devilutionGo/internal/engine/txtdata"
)

// The committed Experience.tsv ladder plugs straight into the award rules: a
// level 1 character that kills a level 2 monster for 2000 xp dings to level 2
// (player.cpp:2407-2443).
func TestAddExperienceOnCommittedLadder(t *testing.T) {
	// given
	ladder := committedLadder(t)
	p := experience.Progress{Level: 1}
	// when
	got := experience.AddExperience(p,
		experience.Award{Experience: 2000, LevelDelta: 1, Alive: true}, ladder, enums.SimFlags{}, nil)
	// then
	if got.Level != 2 {
		t.Fatalf("level = %d, want 2", got.Level)
	}

	// 2000 * 1.1 = 2200, below the level 2 threshold of 4620.
	if got.Experience != 2200 {
		t.Fatalf("experience = %d, want 2200", got.Experience)
	}
}

// committedLadder loads the pinned Experience.tsv copy.
func committedLadder(t *testing.T) txtdata.ExperienceThresholds {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(
		"..", "..", "..", "testdata", "gamedata", "experience", "Experience.tsv",
	))
	if err != nil {
		t.Fatalf("read committed Experience.tsv: %v", err)
	}

	tbl, err := enginetxtdata.Parse(data)
	if err != nil {
		t.Fatalf("Parse Experience.tsv: %v", err)
	}

	ladder, err := txtdata.LoadExperienceThresholds(tbl)
	if err != nil {
		t.Fatalf("txtdata.LoadExperienceThresholds: %v", err)
	}

	return ladder
}
