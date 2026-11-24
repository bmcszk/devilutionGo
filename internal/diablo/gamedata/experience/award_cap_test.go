package experience_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/enums"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/experience"
	"github.com/bmcszk/devilutionGo/internal/diablo/txtdata"
)

// capThresholdDivisor and capLevelFactor are the two literal terms of the
// multiplayer cap (player.cpp:2423: nextLevelThreshold / 20U and 200U * level).
const (
	capThresholdDivisor = 20
	capLevelFactor      = 200
)

// The cap holds at its boundary for every awardable level of the committed
// ladder: at the cap the award lands whole, one point over is clamped, one
// under is granted (player.cpp:2423).
func TestAddExperienceMultiplayerCapBoundary(t *testing.T) {
	// given
	ladder := committedLadder(t)
	// when
	ceilings := awardCeilings(ladder)
	// then every awardable level has a cap and the boundary holds around it
	for level := 1; level < len(ceilings); level++ {
		ceiling := ceilings[level]
		if ceiling == 0 {
			t.Fatalf("level %d has a zero cap", level)
		}

		assertMultiplayerAward(t, ladder, level, ceiling, ceiling)
		assertMultiplayerAward(t, ladder, level, ceiling+1, ceiling)
		assertMultiplayerAward(t, ladder, level, ceiling-1, ceiling-1)
	}
}

// The two cap terms meet at level 6: below it the next-threshold term is the
// smaller one, from level 6 up the 200*level term wins (player.cpp:2420-2423).
func TestAddExperienceMultiplayerCapWorkedExamples(t *testing.T) {
	tests := []struct {
		name   string
		level  int
		want   uint32
		reason string
	}{
		{name: "level1", level: 1, want: 100, reason: "threshold 2000/20 = 100 beats 200*1"},
		{name: "level3", level: 3, want: 402, reason: "threshold 8040/20 = 402 beats 200*3"},
		{name: "level5", level: 5, want: 912, reason: "threshold 18258/20 = 912 beats 200*5"},
		{name: "level6", level: 6, want: 1200, reason: "200*6 = 1200 beats threshold 25712/20 = 1285"},
		{name: "level20", level: 20, want: 4000, reason: "200*20 = 4000 beats threshold 1366227/20"},
		{name: "level50", level: 50, want: 10000, reason: "200*50 = 10000 beats threshold 1583495809/20"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			ladder := committedLadder(t)
			// when
			got := awardCap(ladder, tt.level)
			// then the derived cap matches the upstream worked example
			if got != tt.want {
				t.Fatalf("cap = %d, want %d (%s)", got, tt.want, tt.reason)
			}
		})
	}
}

// A huge award past level 6 lands on the 200*level term while a low level
// character is stopped by the threshold/20 term (player.cpp:2423).
func TestAddExperienceMultiplayerCapClampsHugeAward(t *testing.T) {
	tests := []struct {
		name  string
		level int
	}{
		{name: "level1ThresholdTerm", level: 1},
		{name: "level6LevelTerm", level: 6},
		{name: "level30LevelTerm", level: 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			ladder := committedLadder(t)
			// when
			got := experience.AddExperience(experience.Progress{Level: tt.level},
				experience.Award{Experience: 1_000_000_000, Alive: true}, ladder,
				enums.SimFlags{Multiplayer: true}, nil)
			// then
			if want := awardCap(ladder, tt.level); got.Experience != want {
				t.Fatalf("experience = %d, want the cap %d", got.Experience, want)
			}
		})
	}
}

// The cap sees the level-delta-scaled award, not the raw value: a monster one
// level above the player multiplies 1000 by 1.1 before the cap decides
// (player.cpp:2415, 2423).
func TestAddExperienceMultiplayerCapAppliesAfterScaling(t *testing.T) {
	// given
	ladder := committedLadder(t)
	ceiling := awardCap(ladder, 1)
	// when the scaled value 1100 is over the level 1 cap of 100
	got := experience.AddExperience(experience.Progress{Level: 1},
		experience.Award{Experience: ceiling, LevelDelta: 1, Alive: true}, ladder,
		enums.SimFlags{Multiplayer: true}, nil)
	// then
	if got.Experience != ceiling {
		t.Fatalf("experience = %d, want the cap %d", got.Experience, ceiling)
	}
}

// The cap is a multiplayer rule: the same award in single player runs to the
// level-up ceiling instead (player.cpp:2419, 2428-2431).
func TestAddExperienceSinglePlayerSkipsTheCap(t *testing.T) {
	// given
	ladder := committedLadder(t)
	award := experience.Award{Experience: 1_000_000_000, Alive: true}
	// when
	single := experience.AddExperience(experience.Progress{Level: 1}, award, ladder, enums.SimFlags{}, nil)
	multi := experience.AddExperience(experience.Progress{Level: 1}, award, ladder,
		enums.SimFlags{Multiplayer: true}, nil)
	// then
	if single.Level <= multi.Level {
		t.Fatalf("single-player level %d, want more than the capped %d", single.Level, multi.Level)
	}

	if single.Experience <= awardCap(ladder, 1) {
		t.Fatalf("single-player experience %d, want more than the multiplayer cap %d",
			single.Experience, awardCap(ladder, 1))
	}
}

// assertMultiplayerAward checks one level/raw-award pair against the expected
// stored experience.
func assertMultiplayerAward(
	t *testing.T, ladder txtdata.ExperienceThresholds, level int, award, want uint32,
) {
	t.Helper()

	got := experience.AddExperience(experience.Progress{Level: level},
		experience.Award{Experience: award, Alive: true}, ladder, enums.SimFlags{Multiplayer: true}, nil)
	if got.Experience != want {
		t.Fatalf("level %d award %d = %d, want %d", level, award, got.Experience, want)
	}
}

// awardCap is the multiplayer power-levelling cap exactly as upstream writes
// it: min(clampedExp, nextLevelThreshold/20, 200*level)
// (player.cpp:2423). Every want below is computed through this helper from the
// committed Experience.tsv ladder, never restated from the implementation.
func awardCap(ladder txtdata.ExperienceThresholds, level int) uint32 {
	return min(ladder.ThresholdForLevel(level)/capThresholdDivisor, capLevelFactor*uint32(level))
}

// awardCeilings maps every awardable level (1-based) to its multiplayer cap.
func awardCeilings(ladder txtdata.ExperienceThresholds) []uint32 {
	out := make([]uint32, ladder.MaxLevel())
	for level := 1; level < ladder.MaxLevel(); level++ {
		out[level] = awardCap(ladder, level)
	}

	return out
}
