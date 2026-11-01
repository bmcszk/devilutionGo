// Package experience ports the player experience rules: the award scaling, the
// multiplayer power-levelling cap and the level-up loop.
package experience

import (
	"math"

	"github.com/bmcszk/devilutionGo/internal/diablo/enums"
)

// levelFactor is the 1 + levelDelta/10.0 multiplier (player.cpp:2415).
const levelFactorDivisor = 10.0

// multiplayerLevelFactor is the 200 * level power-levelling cap
// (player.cpp:2423).
const multiplayerLevelFactor = 200

// lowLevelCapDivisor is the threshold/20 cap for characters below level 6
// (player.cpp:2423).
const lowLevelCapDivisor = 20

// Thresholds is the level → experience ladder loaded from Experience.tsv
// (playerdat.cpp:35-70 ExperienceData; txtdata.ExperienceThresholds).
type Thresholds interface {
	// MaxLevel is getMaxCharacterLevel.
	MaxLevel() int
	// ThresholdForLevel is GetNextExperienceThresholdForLevel: the experience
	// needed to reach level, 0 for level 0.
	ThresholdForLevel(level int) uint32
}

// Progress is the local player's experience state.
type Progress struct {
	// Experience is _pExperience.
	Experience uint32
	// Level is _pLevel (getCharacterLevel).
	Level int
}

// Award is one addExperience call (player.h:826-846).
type Award struct {
	// Experience is the base value, usually monster.exp().
	Experience uint32
	// LevelDelta is monsterLevel - getCharacterLevel().
	LevelDelta int
	// Alive is !hasNoLife(): a dead player earns nothing.
	Alive bool
}

// AddExperience ports Player::_addExperience (player.cpp:2407-2445) and
// returns the new progress. levelUp is NextPlrLevel run once per level gained;
// it may be nil when the caller only tracks level and experience.
func AddExperience(p Progress, a Award, t Thresholds, flags enums.SimFlags, levelUp func(level int)) Progress {
	if !a.Alive {
		return p
	}

	if p.Level >= t.MaxLevel() {
		return p
	}

	clamped := scaledAward(a.Experience, a.LevelDelta)
	if flags.Multiplayer {
		clamped = min(clamped, t.ThresholdForLevel(p.Level)/lowLevelCapDivisor)
		clamped = min(clamped, uint32(multiplayerLevelFactor*p.Level))
	}

	// The remaining-cap difference wraps like upstream's uint32 subtraction;
	// ValidatePlayer clamps the stored value afterwards.
	remaining := t.ThresholdForLevel(t.MaxLevel()) - p.Experience
	p.Experience += min(clamped, remaining)

	return levelUpLoop(p, t, levelUp)
}

// SplitMonsterExperience ports AddPlrMonstExper (player.cpp:2446-2460): the
// monster's experience is shared evenly between the players in the kill mask,
// and a partial share is awarded only when the local player is in it.
func SplitMonsterExperience(experience uint32, players int, mine bool) (share uint32, awarded bool) {
	if players == 0 {
		return 0, false
	}

	return experience / uint32(players), mine
}

// scaledAward is the level-difference scaling with the clamp to the uint32
// range (player.cpp:2415).
func scaledAward(experience uint32, levelDelta int) uint32 {
	scaled := float64(experience) * (1 + float64(levelDelta)/levelFactorDivisor)
	if scaled <= 0 {
		return 0
	}

	if scaled >= math.MaxUint32 {
		return math.MaxUint32
	}

	return uint32(int64(scaled))
}

// levelUpLoop is the while loop that raises the character level
// (player.cpp:2438-2442).
func levelUpLoop(p Progress, t Thresholds, levelUp func(level int)) Progress {
	for p.Level < t.MaxLevel() && p.Experience >= t.ThresholdForLevel(p.Level) {
		p.Level++

		if levelUp != nil {
			levelUp(p.Level)
		}
	}

	return p
}
