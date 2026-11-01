package gameplay

import (
	"github.com/bmcszk/devilutionGo/internal/diablo/enums"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/experience"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/quests"
)

// Quest and experience entry points the bean contract names for this tree. The
// ports live in internal/diablo/gamedata/quests and
// internal/diablo/gamedata/experience (AGENTS.md layer 3: Diablo data and the
// rules derived from it), so this file re-exports them the same way
// inventory.go and itemgen.go do — type aliases and function values only.

// Quest state and gating constants, in upstream value order.
const (
	// QuestNotAvail is QUEST_NOTAVAIL.
	QuestNotAvail = quests.QuestNotAvail
	// QuestInit is QUEST_INIT.
	QuestInit = quests.QuestInit
	// QuestActive is QUEST_ACTIVE.
	QuestActive = quests.QuestActive
	// QuestDone is QUEST_DONE.
	QuestDone = quests.QuestDone
	// QuestIDCount is the quest count (Q_INVALID excluded).
	QuestIDCount = quests.QuestIDCount
	// UniqueTypeNone is UniqueMonsterType::None (monster.h:74).
	UniqueTypeNone = quests.UniqueTypeNone
)

// Quest entry points, re-exported from internal/diablo/gamedata/quests.
var (
	// ParseQuestdat is LoadQuestDatFromFile (tables/questdat.cpp:35-53).
	ParseQuestdat = quests.ParseQuestdat
	// InitQuests is InitQuests (quests.cpp:60-110).
	InitQuests = quests.InitQuests
	// InitQuestsWithPools is InitQuests including the randomizeQuests pools.
	InitQuestsWithPools = quests.InitQuestsWithPools
	// InitialiseQuestPools is InitialiseQuestPools (quests.cpp:113-133).
	InitialiseQuestPools = quests.InitialiseQuestPools
	// CheckQuestKill is CheckQuestKill (quests.cpp:227-289).
	CheckQuestKill = quests.CheckQuestKill
	// QuestLogOrder is StartQuestlog's list build (quest_log.cpp:84-120).
	QuestLogOrder = quests.QuestLogOrder
	// VictimFromMonster adapts a monster (table id + UniqueType index) to the
	// kill hook's KillVictim (monster.h:261).
	VictimFromMonster = quests.VictimFromMonster
	// UniqueMonsterName maps a UniqueMonsterType index to its enum name.
	UniqueMonsterName = quests.UniqueMonsterName
)

// Experience entry points, re-exported from
// internal/diablo/gamedata/experience.
var (
	// AddExperience is Player::_addExperience (player.cpp:2407-2445).
	AddExperience = experience.AddExperience
	// SplitMonsterExperience is AddPlrMonstExper (player.cpp:2446-2460).
	SplitMonsterExperience = experience.SplitMonsterExperience
)

// Quest types, re-exported from internal/diablo/gamedata/quests.
type (
	// Quest is struct Quest (questdat.hpp:113-129).
	Quest = quests.Quest
	// QuestState is quest_state (questdat.hpp:79-90).
	QuestState = quests.QuestState
	// QuestRow is one questdat.tsv row.
	QuestRow = quests.QuestRow
	// QuestLog is the quest log layout.
	QuestLog = quests.QuestLog
	// LevelContext is currlevel plus the setlevel flag.
	LevelContext = quests.LevelContext
	// PoolOptions is the randomizeQuests branch of InitQuests.
	PoolOptions = quests.PoolOptions
	// KillVictim is a dead monster as the kill hook sees it.
	KillVictim = quests.KillVictim
	// KillContext is one CheckQuestKill call site.
	KillContext = quests.KillContext
	// KillEffects are the kill side effects outside quest state.
	KillEffects = quests.KillEffects
)

// Experience types, re-exported from internal/diablo/gamedata/experience.
type (
	// Thresholds is the Experience.tsv level ladder.
	Thresholds = experience.Thresholds
	// Progress is the local player's experience state.
	Progress = experience.Progress
	// Award is one addExperience call.
	Award = experience.Award
)

// SimFlags is the game-mode gate set (gbIsSpawn/gbIsHellfire/gbIsMultiplayer).
type SimFlags = enums.SimFlags
