package quests

import "github.com/bmcszk/devilutionGo/internal/diablo/enums"

// ResyncContext is one ResyncMPQuests call.
type ResyncContext struct {
	// Level is currlevel.
	Level int
	// SetLevel is setlevel, which Quest::IsAvailable consults.
	SetLevel bool
	// Flags are the game-mode gates.
	Flags enums.SimFlags
}

// ResyncEffects are the ResyncMPQuests side effects: the quest packet per
// activation and the AltBoy object spawn.
type ResyncEffects struct {
	// SendQuest is NetSendCmdQuest(true, quest).
	SendQuest func(q Quest)
	// SpawnAltBoy is AddObject(OBJ_ALTBOY, ...).
	SpawnAltBoy func()
}

// ResyncMPQuests ports ResyncMPQuests (quests.cpp:349-397): multiplayer quests
// become ACTIVE as the player comes within reach of their level.
func ResyncMPQuests(qs []Quest, rows []QuestRow, ctx ResyncContext, fx ResyncEffects) {
	if !ctx.Flags.QuestsEnabled() {
		return
	}

	activateWithinOneLevel(qs, []questID{QSkelking, QButcher}, ctx, fx)
	activateOneLevelAbove(qs, []questID{QBetrayer, QGrave, QDefiler, QNakrul}, ctx, fx)

	if isBetrayerAvailable(qs, rows, ctx) && fx.SpawnAltBoy != nil {
		fx.SpawnAltBoy()
	}
}

// activateWithinOneLevel is the sentinel block (quests.cpp:353-365).
func activateWithinOneLevel(qs []Quest, ids []questID, ctx ResyncContext, fx ResyncEffects) {
	for _, id := range ids {
		if withinOneLevel(qs[id], ctx.Level) {
			activateQuest(qs, id, fx)
		}
	}
}

// activateOneLevelAbove is the level 14+ block (quests.cpp:367-397).
func activateOneLevelAbove(qs []Quest, ids []questID, ctx ResyncContext, fx ResyncEffects) {
	for _, id := range ids {
		if wakesOneLevelAbove(qs[id], ctx.Level) {
			activateQuest(qs, id, fx)
		}
	}
}

// isBetrayerAvailable gates the AltBoy spawn on Quest::IsAvailable
// (quests.cpp:376-377).
func isBetrayerAvailable(qs []Quest, rows []QuestRow, ctx ResyncContext) bool {
	return qs[QBetrayer].IsAvailable(
		rows[QBetrayer],
		LevelContext{Level: ctx.Level, SetLevel: ctx.SetLevel},
		ctx.Flags,
	)
}

// withinOneLevel is the sentinel range check: INIT and currlevel within one
// level either side (quests.cpp:353-365).
func withinOneLevel(q Quest, level int) bool {
	if q.Active != QuestInit {
		return false
	}

	return level >= q.Level-1 && level <= q.Level+1
}

// wakesOneLevelAbove is the exact check for the level 15+ quests:
// currlevel == _qlevel - 1 (quests.cpp:367-397).
func wakesOneLevelAbove(q Quest, level int) bool {
	return q.Active == QuestInit && level == q.Level-1
}

// activateQuest flips a quest to ACTIVE and broadcasts it (quests.cpp:355-359).
func activateQuest(qs []Quest, id questID, fx ResyncEffects) {
	qs[id].Active = QuestActive

	if fx.SendQuest != nil {
		fx.SendQuest(qs[id])
	}
}
