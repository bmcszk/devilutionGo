package quests

import "github.com/bmcszk/devilutionGo/internal/diablo/enums"

// Towner ids the quest state gates (towners.h _talker_id). Only the gated ones
// are named here; every other towner is always present.
const (
	// TownerDeadGuy is TOWN_DEADGUY.
	TownerDeadGuy = "TOWN_DEADGUY"
	// TownerFarmer is TOWN_FARMER, Lester.
	TownerFarmer = "TOWN_FARMER"
	// TownerCowFarm is TOWN_COWFARM, the cow.
	TownerCowFarm = "TOWN_COWFARM"
	// TownerGirl is TOWN_GIRL, Celia.
	TownerGirl = "TOWN_GIRL"
)

// TownerContext is the state IsTownerPresent reads besides the quest slots.
type TownerContext struct {
	// Flags are the game-mode gates, including the Hellfire quest options.
	Flags enums.SimFlags
	// Visited17 is MyPlayer._pLvlVisited[17].
	Visited17 bool
}

// IsTownerPresent ports IsTownerPresent (towners.cpp:735-749): the wounded
// townsman, Lester, the cow and Celia come and go with the quest state.
func IsTownerPresent(towner string, qs []Quest, ctx TownerContext) bool {
	switch towner {
	case TownerDeadGuy:
		return deadGuyPresent(qs)
	case TownerFarmer:
		return farmerPresent(qs, ctx)
	case TownerCowFarm:
		return cowPresent(ctx)
	case TownerGirl:
		return girlPresent(qs, ctx)
	default:
		return true
	}
}

// deadGuyPresent is towners.cpp:736-737: the butcher is neither rolled out nor
// done.
func deadGuyPresent(qs []Quest) bool {
	return qs[QButcher].Active != QuestNotAvail && qs[QButcher].Active != QuestDone
}

// farmerPresent is towners.cpp:738-739: Hellfire, no cow quest, hive not done.
func farmerPresent(qs []Quest, ctx TownerContext) bool {
	return ctx.Flags.Hellfire && !ctx.Flags.CowQuest && qs[QFarmer].Active != QuestHiveDone
}

// cowPresent is towners.cpp:740-741: the cow replaces Lester with bCowQuest.
func cowPresent(ctx TownerContext) bool {
	return ctx.Flags.Hellfire && ctx.Flags.CowQuest
}

// girlPresent is towners.cpp:742-743: Celia needs bTheoQuest, a level 17 visit
// and an unfinished quest.
func girlPresent(qs []Quest, ctx TownerContext) bool {
	return ctx.Flags.Hellfire && ctx.Flags.TheoQuest && ctx.Visited17 && qs[QGirl].Active != QuestDone
}
