package quests

import "github.com/bmcszk/devilutionGo/internal/diablo/enums"

// LevelContext is the level the player currently occupies: currlevel plus the
// setlevel flag that marks a quest sub-level (quests.cpp:137).
type LevelContext struct {
	// Level is currlevel.
	Level int
	// SetLevel is setlevel: the player is inside a quest map.
	SetLevel bool
}

// IsAvailable ports Quest::IsAvailable (questdat.cpp:64-74): the quest must be
// on this level, outside a set level, still in the game, and allowed by the
// single-player-only rule.
func (q Quest) IsAvailable(row QuestRow, level LevelContext, flags enums.SimFlags) bool {
	if level.SetLevel {
		return false
	}

	if level.Level != q.Level {
		return false
	}

	if q.Active == QuestNotAvail {
		return false
	}

	if row.IsSinglePlayerOnly && flags.UseMultiplayerQuests() {
		return false
	}

	return true
}
