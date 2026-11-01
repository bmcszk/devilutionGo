package quests

import (
	"errors"
	"fmt"

	"github.com/bmcszk/devilutionGo/internal/diablo/enums"
)

// Sub-states of the multi-step quests (questdat.hpp:17-52), kept in upstream
// order because saves and net packets carry the raw number.
const (
	// Mushroom quest (questdat.hpp:17-26).
	QSMushInit     uint8 = 0 // QS_INIT
	QSTomeSpawned  uint8 = 1 // QS_TOMESPAWNED
	QSTomeGiven    uint8 = 2 // QS_TOMEGIVEN
	QSMushSpawned  uint8 = 3 // QS_MUSHSPAWNED
	QSMushPicked   uint8 = 4 // QS_MUSHPICKED
	QSMushGiven    uint8 = 5 // QS_MUSHGIVEN
	QSBrainSpawned uint8 = 6 // QS_BRAINSPAWNED
	QSBrainGiven   uint8 = 7 // QS_BRAINGIVEN
)

// ErrRowCount is returned when the quest table does not hold one row per quest.
var ErrRowCount = errors.New("quests: unexpected questdat row count")

// Quest ports struct Quest (questdat.hpp:113-129): one per-game quest slot.
// quest.position is world-map state and stays with the level layer.
type Quest struct {
	ID      questID    // _qidx
	Active  QuestState // _qactive
	Level   int        // _qlevel
	LvlType string     // _qlvltype
	QSLvl   string     // _qslvl
	Var1    uint8      // _qvar1
	Var2    uint8      // _qvar2
	Log     bool       // _qlog
	Msg     string     // _qmsg
}

// InitQuests ports InitQuests (quests.cpp:60-110): reset every slot from the
// quest table, pick the level/state from the game mode, apply the fixed
// starting variables, then gate everything off in the shareware build.
// The randomizeQuests pools are applied separately by ApplyQuestPools.
func InitQuests(rows []QuestRow, flags enums.SimFlags) ([]Quest, error) {
	if len(rows) != QuestIDCount {
		return nil, fmt.Errorf("init quests: %d rows: %w", len(rows), ErrRowCount)
	}

	out := make([]Quest, len(rows))
	for i, row := range rows {
		out[i] = newQuest(questID(i), row, flags)
	}

	applyInitialVars(out, flags)

	if !flags.QuestsEnabled() {
		markAllUnavailable(out)
	}

	return out, nil
}

// newQuest is one loop body of InitQuests (quests.cpp:67-87).
func newQuest(id questID, row QuestRow, flags enums.SimFlags) Quest {
	q := Quest{
		ID:      id,
		Active:  QuestNotAvail,
		LvlType: row.QLvlT,
		QSLvl:   row.QSLvl,
		Msg:     row.QDMsg,
	}

	// Single player takes the retail level; multiplayer skips the single
	// player only quests entirely, leaving them NOTAVAIL at level 0.
	if !flags.UseMultiplayerQuests() {
		q.Level = row.QDLvl
		q.Active = QuestInit
	} else if !row.IsSinglePlayerOnly {
		q.Level = row.QDMultLvl
		q.Active = QuestInit
	}

	return q
}

// applyInitialVars is the fixed starting-state block after the slot loop
// (quests.cpp:97-105).
func applyInitialVars(qs []Quest, flags enums.SimFlags) {
	if qs[QSkelking].Active == QuestNotAvail {
		qs[QSkelking].Var2 = 2
	}

	if qs[QRock].Active == QuestNotAvail {
		qs[QRock].Var2 = 2
	}

	qs[QLtbanner].Var1 = 1

	if flags.UseMultiplayerQuests() {
		qs[QBetrayer].Var1 = 2
	}

	// In multiplayer items spawn during level generation to avoid desyncs.
	if flags.Multiplayer && qs[QMushroom].Active == QuestInit {
		qs[QMushroom].Var1 = QSTomeSpawned
	}
}

// markAllUnavailable is the shareware gate (quests.cpp:107-110).
func markAllUnavailable(qs []Quest) {
	for i := range qs {
		qs[i].Active = QuestNotAvail
	}
}
