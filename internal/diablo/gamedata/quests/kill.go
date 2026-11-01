package quests

import "github.com/bmcszk/devilutionGo/internal/diablo/enums"

// Quest monster ids (monstdat.tsv _monster_id) and unique types
// (monster.h:60-75 UniqueMonsterType) the kill hook reacts to.
const (
	// MonsterSkeletonKing is MT_SKING.
	MonsterSkeletonKing = "MT_SKING"
	// MonsterButcher is MT_CLEAVER.
	MonsterButcher = "MT_CLEAVER"
	// UniqueGarbud is UniqueMonsterType::Garbud, "Gharbad the Weak".
	UniqueGarbud = "Garbud"
	// UniqueZhar is UniqueMonsterType::Zhar.
	UniqueZhar = "Zhar"
	// UniqueLazarus is UniqueMonsterType::Lazarus, the Arch-Bishop.
	UniqueLazarus = "Lazarus"
	// UniqueWarlordOfBlood is UniqueMonsterType::WarlordOfBlood.
	UniqueWarlordOfBlood = "WarlordOfBlood"
)

// Hero speech ids the kill hook triggers (player.h HeroSpeech, 30 ticks).
const (
	// SpeechRestWellLeoric is RestWellLeoricIllFindYourSon.
	SpeechRestWellLeoric = "RestWellLeoricIllFindYourSon"
	// SpeechSpiritsAvenged is TheSpiritsOfTheDeadAreNowAvenged.
	SpeechSpiritsAvenged = "TheSpiritsOfTheDeadAreNowAvenged"
	// SpeechNotImpressed is ImNotImpressed.
	SpeechNotImpressed = "ImNotImpressed"
	// SpeechSorryConcentration is ImSorryDidIBreakYourConcentration.
	SpeechSorryConcentration = "ImSorryDidIBreakYourConcentration"
	// SpeechMadnessEndsHere is YourMadnessEndsHereBetrayer.
	SpeechMadnessEndsHere = "YourMadnessEndsHereBetrayer"
	// SpeechReignOfPainEnded is YourReignOfPainHasEnded.
	SpeechReignOfPainEnded = "YourReignOfPainHasEnded"
)

// lazarusVar1Done is the Q_BETRAYER var1 written when Lazarus dies
// (quests.cpp:255).
const lazarusVar1Done uint8 = 7

// lazarusVar2Portal is the Q_BETRAYER var2 that records the opened portal
// (quests.cpp:274).
const lazarusVar2Portal uint8 = 4

// KillVictim is the kill hook's view of a dead monster: its monster table id
// and its unique type, both empty for ordinary monsters.
type KillVictim struct {
	// MonsterID is MonsterData.ID, the _monster_id column (e.g. MT_SKING).
	MonsterID string
	// UniqueType is UniqueMonsterType; "" means none.
	UniqueType string
}

// KillContext is one CheckQuestKill call site.
type KillContext struct {
	// Victim is the monster that just died.
	Victim KillVictim
	// Flags are the game-mode gates.
	Flags enums.SimFlags
	// SendMsgs is upstream's sendmsg argument.
	SendMsgs bool
}

// KillEffects are the quest-kill side effects outside quest state.
type KillEffects struct {
	// Say is Player::Say(speech, 30).
	Say func(speech string)
	// SendQuest is NetSendCmdQuest(true, quest).
	SendQuest func(q Quest)
	// LazarusPortal is the single player branch: InitVPTriggers plus the red
	// portal missile at (35, 32) (quests.cpp:271-275).
	LazarusPortal func()
	// LazarusTriggers is the multiplayer branch: the trigger scan for tile 369
	// (quests.cpp:263-270).
	LazarusTriggers func()
}

// CheckQuestKill ports CheckQuestKill (quests.cpp:227-289): the quest state
// changes a killing blow causes. Effects outside quest state go through fx.
func CheckQuestKill(qs []Quest, ctx KillContext, fx KillEffects) {
	if !ctx.Flags.QuestsEnabled() {
		return
	}

	switch ctx.Victim.MonsterID {
	case MonsterSkeletonKing:
		finishQuest(qs, QSkelking, SpeechRestWellLeoric, fx)
		broadcast(ctx, qs[QSkelking], fx)
	case MonsterButcher:
		finishQuest(qs, QButcher, SpeechSpiritsAvenged, fx)
		broadcast(ctx, qs[QButcher], fx)
	default:
		checkUniqueKill(qs, ctx, fx)
	}
}

// checkUniqueKill is the UniqueMonsterType chain (quests.cpp:241-289); the
// monster type cases above take precedence over it.
func checkUniqueKill(qs []Quest, ctx KillContext, fx KillEffects) {
	switch ctx.Victim.UniqueType {
	case UniqueGarbud:
		finishQuest(qs, QGarbud, SpeechNotImpressed, fx)
		broadcastAlways(qs[QGarbud], fx)
	case UniqueZhar:
		finishQuest(qs, QZhar, SpeechSorryConcentration, fx)
		broadcastAlways(qs[QZhar], fx)
	case UniqueLazarus:
		killLazarus(qs, ctx, fx)
	case UniqueWarlordOfBlood:
		finishQuest(qs, QWarlord, SpeechReignOfPainEnded, fx)
		broadcastAlways(qs[QWarlord], fx)
	default:
		// Ordinary monsters finish no quest.
		return
	}
}

// killLazarus closes Q_BETRAYER, starts Q_DIABLO and opens the way down
// (quests.cpp:251-289).
func killLazarus(qs []Quest, ctx KillContext, fx KillEffects) {
	qs[QBetrayer].Active = QuestDone
	if fx.Say != nil {
		fx.Say(SpeechMadnessEndsHere)
	}

	qs[QBetrayer].Var1 = lazarusVar1Done
	qs[QDiablo].Active = QuestActive

	if ctx.Flags.UseMultiplayerQuests() {
		if fx.LazarusTriggers != nil {
			fx.LazarusTriggers()
		}
	} else {
		if fx.LazarusPortal != nil {
			fx.LazarusPortal()
		}

		qs[QBetrayer].Var2 = lazarusVar2Portal
	}

	broadcast(ctx, qs[QBetrayer], fx)
	broadcast(ctx, qs[QDiablo], fx)
}

// finishQuest sets a quest DONE and has the hero comment (quests.cpp:234-238).
func finishQuest(qs []Quest, id questID, speech string, fx KillEffects) {
	qs[id].Active = QuestDone

	if fx.Say != nil {
		fx.Say(speech)
	}
}

// broadcast sends a quest packet when the caller asked for messages.
func broadcast(ctx KillContext, q Quest, fx KillEffects) {
	if !ctx.SendMsgs || fx.SendQuest == nil {
		return
	}

	fx.SendQuest(q)
}

// broadcastAlways sends a quest packet: the unique kills do not consult
// sendmsg (quests.cpp:243-244, 247-248, 285-288).
func broadcastAlways(q Quest, fx KillEffects) {
	if fx.SendQuest != nil {
		fx.SendQuest(q)
	}
}
