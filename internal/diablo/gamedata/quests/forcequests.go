package quests

// questTriggerNames is QuestTriggerNames (quests.cpp:40-46), indexed by
// _qslvl - 1: skeleton king, bone chamber, maze, poison water, vile betrayer.
var questTriggerNames = [...]string{
	"King Leoric's Tomb",
	"The Chamber of Bone",
	"Maze",
	"A Dark Passage",
	"Unholy Altar",
}

// Position is a world tile (Point).
type Position struct {
	// X is tile X.
	X int
	// Y is tile Y.
	Y int
}

// ForceQuestTarget is the quest portal the cursor stepped into.
type ForceQuestTarget struct {
	// Cursor is the position to move to (quest.position).
	Cursor Position
	// MapName is the QuestTriggerNames entry; the "To {:s}" wrapper is text
	// layer state (quests.cpp:217).
	MapName string
}

// ForceQuests ports ForceQuests (quests.cpp:203-225): in single player, moving
// the cursor into a quest portal on the owning level snaps it to the portal and
// names the map. contains is EntranceBoundaryContains.
func ForceQuests(
	qs []Quest,
	positions []Position,
	cursor Position,
	ctx ResyncContext,
	contains func(questPos, cursor Position) bool,
) (ForceQuestTarget, bool) {
	if !ctx.Flags.QuestsEnabled() || ctx.Flags.UseMultiplayerQuests() {
		return ForceQuestTarget{}, false
	}

	for i := range qs {
		if !isQuestPortal(qs[i], ctx.Level) {
			continue
		}

		if contains(positions[i], cursor) {
			return ForceQuestTarget{
				Cursor:  positions[i],
				MapName: questTriggerNames[setLevelIndex(qs[i].QSLvl)-1],
			}, true
		}
	}

	return ForceQuestTarget{}, false
}

// setLevelIndex maps the _setlevels names to their enum value
// (gendung_defs.hpp:47-61). SL_NONE is 0.
func setLevelIndex(name string) int {
	switch name {
	case SLSkelking:
		return 1
	case SLBonechamb:
		return 2
	case "SL_MAZE":
		return 3
	case SLPoisonwater:
		return 4
	case SLVilebetrayer:
		return 5
	default:
		return 0
	}
}

// isQuestPortal is the loop guard: not the betrayer, on its own level, and a
// special level (quests.cpp:211).
func isQuestPortal(q Quest, level int) bool {
	return q.ID != QBetrayer && level == q.Level && setLevelIndex(q.QSLvl) != 0
}
