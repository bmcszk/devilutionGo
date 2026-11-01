package txtdata_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/txtdata"
)

// TestLoadExperienceThresholdsCommittedFixture pins the real ladder: the
// committed assets/txtdata/Experience.tsv copy at pin c20b2f2456 holds levels
// 1-50 (see internal/testdata/gamedata/experience/README.md).
func TestLoadExperienceThresholdsCommittedFixture(t *testing.T) {
	// given
	data, err := os.ReadFile(filepath.Join(
		"..", "..", "..", "internal", "testdata", "gamedata", "experience", "Experience.tsv",
	))
	if err != nil {
		t.Fatalf("read committed Experience.tsv: %v", err)
	}

	// when
	e, err := txtdata.LoadExperienceThresholds(mustParse(t, data))
	// then
	if err != nil {
		t.Fatalf("txtdata.LoadExperienceThresholds: %v", err)
	}

	if e.MaxLevel() != 50 {
		t.Fatalf("MaxLevel = %d, want 50", e.MaxLevel())
	}

	want := []struct {
		level int
		xp    uint32
	}{
		{level: 1, xp: 2000},
		{level: 2, xp: 4620},
		{level: 15, xp: 313656},
		{level: 30, xp: 20066900},
		{level: 50, xp: 1583495809},
	}
	for _, w := range want {
		if got := e.ThresholdForLevel(w.level); got != w.xp {
			t.Fatalf("level %d threshold = %d, want %d", w.level, got, w.xp)
		}
	}
}

// The ladder is strictly increasing, which the level-up loop relies on.
func TestCommittedExperienceLadderIncreases(t *testing.T) {
	// given
	data, err := os.ReadFile(filepath.Join(
		"..", "..", "..", "internal", "testdata", "gamedata", "experience", "Experience.tsv",
	))
	if err != nil {
		t.Fatalf("read committed Experience.tsv: %v", err)
	}

	// when
	e, err := txtdata.LoadExperienceThresholds(mustParse(t, data))
	// then
	if err != nil {
		t.Fatalf("txtdata.LoadExperienceThresholds: %v", err)
	}

	for level := 2; level <= e.MaxLevel(); level++ {
		prev, cur := e.ThresholdForLevel(level-1), e.ThresholdForLevel(level)
		if cur <= prev {
			t.Fatalf("level %d threshold %d <= level %d threshold %d", level, cur, level-1, prev)
		}
	}
}
