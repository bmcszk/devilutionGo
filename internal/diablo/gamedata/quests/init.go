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

// PoolOptions is the randomizeQuests branch of InitQuests: the pools are drawn
// from the level 15 dungeon seed (quests.cpp:97-99). A nil Rnd means the
// option is off, so no quest is rolled out of the game.
type PoolOptions struct {
	// Seed is DungeonSeeds[15].
	Seed uint32
	// SetRndSeed resets the generator to Seed.
	SetRndSeed func(uint32)
	// Rnd is GenerateRnd on that generator.
	Rnd func(int32) int32
}

// InitQuests ports InitQuests (quests.cpp:60-110) without the optional pools.
func InitQuests(rows []QuestRow, flags enums.SimFlags) ([]Quest, error) {
	return InitQuestsWithPools(rows, flags, PoolOptions{})
}

// InitQuestsWithPools ports InitQuests including the randomizeQuests pools.
// Order matters: the pools run before the fixed starting variables, so a
// pooled-out sentinel quest still gets its var2 = 2 mark (quests.cpp:97-101).
func InitQuestsWithPools(rows []QuestRow, flags enums.SimFlags, pools PoolOptions) ([]Quest, error) {
	if len(rows) != QuestIDCount {
		return nil, fmt.Errorf("init quests: %d rows: %w", len(rows), ErrRowCount)
	}

	out := make([]Quest, len(rows))
	for i, row := range rows {
		out[i] = newQuest(questID(i), row, flags)
	}

	if pools.Rnd != nil && !flags.UseMultiplayerQuests() {
		applyPools(out, pools)
	}

	applyInitialVars(out, flags)

	if !flags.QuestsEnabled() {
		markAllUnavailable(out)
	}

	return out, nil
}

// applyPools rolls the quest pools and marks the picked quests unavailable
// (quests.cpp:97-99).
func applyPools(qs []Quest, pools PoolOptions) {
	for i, state := range InitialiseQuestPools(pools.Seed, pools.SetRndSeed, pools.Rnd) {
		if state == QuestNotAvail {
			qs[i].Active = QuestNotAvail
		}
	}
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
