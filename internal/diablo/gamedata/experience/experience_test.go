package experience_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/enums"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/experience"
)

// testLadder mirrors the committed table's first levels plus a small cap.
var testLadder = ladder{100, 300, 600, 1000, 1500}

// longLadder keeps the first levels cheap and the rest linear so the
// multiplayer caps can be checked by hand.
var longLadder = linearLadder(50)

// ladder is the committed Experience.tsv ladder used by most tests; xpAt
// returns the threshold for a 1-based level.
type ladder []uint32

func (l ladder) MaxLevel() int { return len(l) }

func (l ladder) ThresholdForLevel(level int) uint32 {
	if level <= 0 {
		return 0
	}

	if level > len(l) {
		return l[len(l)-1]
	}

	return l[level-1]
}

// A plain award on an alive player adds the raw experience (player.cpp:2415,
// 2431).
func TestAddExperiencePlainAward(t *testing.T) {
	// given
	p := experience.Progress{Level: 1}
	// when
	got := experience.AddExperience(p, experience.Award{Experience: 50, Alive: true}, testLadder,
		enums.SimFlags{}, nil)
	// then
	if got.Experience != 50 || got.Level != 1 {
		t.Fatalf("progress = %d/%d, want 50/1", got.Experience, got.Level)
	}
}

// The award scales with the level difference: experience * (1 + levelDelta/10)
// truncated toward zero and clamped at 0 (player.cpp:2415).
func TestAddExperienceLevelDeltaScaling(t *testing.T) {
	tests := []struct {
		name       string
		levelDelta int
		want       uint32
	}{
		{name: "even", levelDelta: 0, want: 100},
		{name: "higherMonster", levelDelta: 3, want: 130},
		{name: "lowerMonster", levelDelta: -5, want: 50},
		{name: "zeroFactor", levelDelta: -10, want: 0},
		{name: "negativeClamp", levelDelta: -20, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			a := experience.Award{Experience: 100, LevelDelta: tt.levelDelta, Alive: true}
			// when
			got := experience.AddExperience(experience.Progress{Level: 1}, a, testLadder, enums.SimFlags{}, nil)
			// then
			if got.Experience != tt.want {
				t.Fatalf("experience = %d, want %d", got.Experience, tt.want)
			}
		})
	}
}

// A dead player gains nothing (player.cpp:2409-2410).
func TestAddExperienceDeadPlayer(t *testing.T) {
	// when
	got := experience.AddExperience(experience.Progress{Experience: 10, Level: 1},
		experience.Award{Experience: 100}, testLadder, enums.SimFlags{}, nil)
	// then
	if got.Experience != 10 {
		t.Fatalf("experience = %d, want unchanged 10", got.Experience)
	}
}

// Multiplayer caps the award at threshold/20 and at 200*level, whichever is
// smaller, to stop power levelling (player.cpp:2417-2425).
func TestAddExperienceMultiplayerCap(t *testing.T) {
	tests := []struct {
		name  string
		level int
		want  uint32
	}{
		// level 3: threshold 3000/20 = 150 vs 200*3 = 600 -> 150
		{name: "lowLevelThresholdCap", level: 3, want: 150},
		// level 20: 20000/20 = 1000 vs 200*20 = 4000 -> 1000
		{name: "highLevelThresholdCap", level: 20, want: 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			a := experience.Award{Experience: 5000, Alive: true}
			// when
			got := experience.AddExperience(experience.Progress{Level: tt.level}, a, longLadder,
				enums.SimFlags{Multiplayer: true}, nil)
			// then
			if got.Experience != tt.want {
				t.Fatalf("experience = %d, want the cap %d", got.Experience, tt.want)
			}
		})
	}
}

// The level-up loop runs once per threshold crossed and reports each level
// (player.cpp:2438-2442).
func TestAddExperienceLevelsUpRepeatedly(t *testing.T) {
	// given
	var seen []int
	levelUp := func(level int) { seen = append(seen, level) }
	// when
	got := experience.AddExperience(experience.Progress{Level: 1},
		experience.Award{Experience: 1200, Alive: true}, testLadder, enums.SimFlags{}, levelUp)
	// then
	if got.Level != 5 {
		t.Fatalf("level = %d, want 5", got.Level)
	}

	want := []int{2, 3, 4, 5}
	if len(seen) != len(want) {
		t.Fatalf("levelUp calls = %v, want %v", seen, want)
	}

	for i := range want {
		if seen[i] != want[i] {
			t.Fatalf("levelUp calls = %v, want %v", seen, want)
		}
	}
}

// A character at the maximum level earns nothing more (player.cpp:2412-2414).
func TestAddExperienceAtMaxLevel(t *testing.T) {
	// when
	got := experience.AddExperience(experience.Progress{Experience: 1500, Level: 5},
		experience.Award{Experience: 1000, Alive: true}, testLadder, enums.SimFlags{}, nil)
	// then
	if got.Experience != 1500 || got.Level != 5 {
		t.Fatalf("progress = %d/%d, want unchanged 1500/5", got.Experience, got.Level)
	}
}

// The stored experience never passes the top threshold (player.cpp:2428-2431).
func TestAddExperienceStopsAtTopThreshold(t *testing.T) {
	// when
	got := experience.AddExperience(experience.Progress{Experience: 900, Level: 4},
		experience.Award{Experience: 10000, Alive: true}, testLadder, enums.SimFlags{}, nil)
	// then
	if got.Experience != 1500 || got.Level != 5 {
		t.Fatalf("progress = %d/%d, want 1500/5", got.Experience, got.Level)
	}
}

// AddPlrMonstExper divides the monster's experience by the number of players
// in the kill mask and awards the local player only when it is in the mask
// (player.cpp:2446-2460).
func TestSplitMonsterExperience(t *testing.T) {
	tests := []struct {
		name       string
		experience uint32
		players    int
		mine       bool
		wantShare  uint32
		wantAward  bool
	}{
		{name: "solo", experience: 1000, players: 1, mine: true, wantShare: 1000, wantAward: true},
		{name: "shared", experience: 1000, players: 4, mine: true, wantShare: 250, wantAward: true},
		{name: "truncated", experience: 999, players: 2, mine: true, wantShare: 499, wantAward: true},
		{name: "otherPlayerOnly", experience: 1000, players: 4, mine: false, wantShare: 250, wantAward: false},
		{name: "noPlayers", experience: 1000, players: 0, mine: true, wantShare: 0, wantAward: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// when
			share, awarded := experience.SplitMonsterExperience(tt.experience, tt.players, tt.mine)
			// then
			if share != tt.wantShare || awarded != tt.wantAward {
				t.Fatalf("share/awarded = %d/%v, want %d/%v", share, awarded, tt.wantShare, tt.wantAward)
			}
		})
	}
}

// linearLadder builds a ladder whose level n threshold is n*1000.
func linearLadder(levels int) ladder {
	out := make(ladder, levels)
	for i := range out {
		out[i] = uint32(i+1) * 1000
	}

	return out
}
