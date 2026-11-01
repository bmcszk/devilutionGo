package quests

// Special level names as the quest table spells them (_qslvl / setlvlnum).
const (
	// SLNone is SL_NONE.
	SLNone = "SL_NONE"
	// SLSkelking is SL_SKELKING.
	SLSkelking = "SL_SKELKING"
	// SLBonechamb is SL_BONECHAMB.
	SLBonechamb = "SL_BONECHAMB"
	// SLPoisonwater is SL_POISONWATER.
	SLPoisonwater = "SL_POISONWATER"
	// SLVilebetrayer is SL_VILEBETRAYER.
	SLVilebetrayer = "SL_VILEBETRAYER"
)

// MapReturnLevel ports GetMapReturnLevel (quests.cpp:289-303): leaving a quest
// level returns the player to the level the quest lives on, or to town (0).
func MapReturnLevel(qs []Quest, setLevel string) int {
	switch setLevel {
	case SLSkelking:
		return qs[QSkelking].Level
	case SLBonechamb:
		return qs[QSchamb].Level
	case SLPoisonwater:
		return qs[QPwater].Level
	case SLVilebetrayer:
		return qs[QBetrayer].Level
	default:
		return 0
	}
}
