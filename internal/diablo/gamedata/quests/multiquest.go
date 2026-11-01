package quests

// townLevel is "0", the level MyPlayer->isOnLevel(0) checks.
const townLevel = "0"

// MultiQuestUpdate is one SetMultiQuest call (quests.cpp:591).
type MultiQuestUpdate struct {
	// Slot is the quest index q.
	Slot int
	// State is the new quest_state s.
	State QuestState
	// Log is the log flag: the quest appears in the quest log.
	Log bool
	// Var1 is the new qvar1; it only ever grows.
	Var1 uint8
	// Var2 is the new qvar2.
	Var2 uint8
	// Msg is the new qmsg.
	Msg string
}

// MultiQuestEffects are the SetMultiQuest side effects outside quest state.
type MultiQuestEffects struct {
	// ResyncQuests is the single-player resync (quests.cpp:609).
	ResyncQuests func()
	// IsOnLevel is MyPlayer->isOnLevel(level): a quest level name for Q_PWATER,
	// "0" for the Hellfire towner quests.
	IsOnLevel func(qslvl string) bool
	// StartPWaterPurify is StartPWaterPurify (quests.cpp:614-615).
	StartPWaterPurify func()
	// UpdateGirlAnim is UpdateGirlAnimAfterQuestComplete (quests.cpp:616-617).
	UpdateGirlAnim func()
	// UpdateCowFarmerAnim is UpdateCowFarmerAnimAfterQuestComplete
	// (quests.cpp:618-619).
	UpdateCowFarmerAnim func()
}

// SetMultiQuest ports SetMultiQuest (quests.cpp:591-620): the shared quest
// update the network layer replays, with monotonic state and variables.
func SetMultiQuest(qs []Quest, u MultiQuestUpdate, ctx ResyncContext, fx MultiQuestEffects) {
	if ctx.Flags.Spawn {
		return
	}

	q := &qs[u.Slot]
	before := q.Active

	raiseQuestState(q, u)
	if u.Var1 > q.Var1 {
		q.Var1 = u.Var1
	}

	q.Var2 = u.Var2
	q.Msg = u.Msg

	if ctx.Flags.UseMultiplayerQuests() {
		return
	}

	if fx.ResyncQuests != nil {
		fx.ResyncQuests()
	}

	applyQuestCompletion(*q, before, fx)
}

// raiseQuestState is the guarded state bump: never below the current state, and
// the Hellfire hive teases give way to ACTIVE/DONE (quests.cpp:598-605).
func raiseQuestState(q *Quest, u MultiQuestUpdate) {
	if q.Active == QuestDone {
		return
	}

	if u.State > q.Active || (isActiveOrDone(u.State) && isHiveTease(q.Active)) {
		q.Active = u.State
	}

	if u.Log {
		q.Log = true
	}
}

// isActiveOrDone is IsAnyOf(s, QUEST_ACTIVE, QUEST_DONE).
func isActiveOrDone(s QuestState) bool {
	return s == QuestActive || s == QuestDone
}

// isHiveTease is IsAnyOf(s, QUEST_HIVE_TEASE1, QUEST_HIVE_TEASE2,
// QUEST_HIVE_ACTIVE).
func isHiveTease(s QuestState) bool {
	return s == QuestHiveTease1 || s == QuestHiveTease2 || s == QuestHiveActive
}

// applyQuestCompletion runs the completion animations for the quests that have
// them (quests.cpp:611-619).
func applyQuestCompletion(q Quest, before QuestState, fx MultiQuestEffects) {
	if before == QuestDone || q.Active != QuestDone {
		return
	}

	switch q.ID {
	case QPwater:
		completeWater(q, fx)
	case QGirl:
		completeGirl(fx)
	case QJersey:
		completeCow(fx)
	default:
	}
}

// completeWater purifies the water when the player stands on the poison water
// level (quests.cpp:614-615).
func completeWater(q Quest, fx MultiQuestEffects) {
	if onQuestLevel(q, fx) && fx.StartPWaterPurify != nil {
		fx.StartPWaterPurify()
	}
}

// completeGirl updates Celia when the player is in town (quests.cpp:616-617).
func completeGirl(fx MultiQuestEffects) {
	if onTownLevel(fx) && fx.UpdateGirlAnim != nil {
		fx.UpdateGirlAnim()
	}
}

// completeCow updates the cow farmer when the player is in town
// (quests.cpp:618-619).
func completeCow(fx MultiQuestEffects) {
	if onTownLevel(fx) && fx.UpdateCowFarmerAnim != nil {
		fx.UpdateCowFarmerAnim()
	}
}

// onQuestLevel is MyPlayer->isOnLevel(quest._qslvl).
func onQuestLevel(q Quest, fx MultiQuestEffects) bool {
	return fx.IsOnLevel != nil && fx.IsOnLevel(q.QSLvl)
}

// onTownLevel is MyPlayer->isOnLevel(0).
func onTownLevel(fx MultiQuestEffects) bool {
	return fx.IsOnLevel != nil && fx.IsOnLevel(townLevel)
}
