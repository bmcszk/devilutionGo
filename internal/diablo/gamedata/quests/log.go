package quests

import "slices"

// QuestLog is the quest log layout: the open quests flagged for the log, then
// the finished ones, each block ordered by questBookOrder
// (quest_log.cpp:85-120).
type QuestLog struct {
	// Order holds quest ids in drawing order.
	Order []questID
	// FirstFinished is the first finished entry; the blocks split when it is
	// neither 0 nor len(Order).
	FirstFinished int
}

// QuestLogOrder ports StartQuestlog's list build (quest_log.cpp:84-120).
func QuestLogOrder(qs []Quest, rows []QuestRow) QuestLog {
	open := collectOpenQuests(qs)
	finished := collectFinishedQuests(qs)

	sortByBookOrder(open, rows)
	sortByBookOrder(finished, rows)

	return QuestLog{Order: append(open, finished...), FirstFinished: len(open)}
}

// collectOpenQuests is the active quests the player has been told about
// (quest_log.cpp:95-101).
func collectOpenQuests(qs []Quest) []questID {
	var ids []questID

	for _, q := range qs {
		if q.Active == QuestActive && q.Log {
			ids = append(ids, q.ID)
		}
	}

	return ids
}

// collectFinishedQuests is the done block: state, not the log flag, decides
// (quest_log.cpp:103-109).
func collectFinishedQuests(qs []Quest) []questID {
	var ids []questID

	for _, q := range qs {
		if q.Active == QuestDone || q.Active == QuestHiveDone {
			ids = append(ids, q.ID)
		}
	}

	return ids
}

// sortByBookOrder is the questBookOrder comparator (quest_log.cpp:93).
// Book order values are unique per quest, so the sort is stable in effect.
func sortByBookOrder(ids []questID, rows []QuestRow) {
	slices.SortStableFunc(ids, func(a, b questID) int {
		return rows[a].BookOrder - rows[b].BookOrder
	})
}
