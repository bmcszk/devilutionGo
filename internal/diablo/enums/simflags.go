package enums

// SimFlags are the game-mode globals that gate Diablo content:
// gbIsSpawn (shareware data), gbIsHellfire and gbIsMultiplayer
// (game_mode.hpp:41-45), plus the multiplayerFullQuests option that
// participates in the quest-table choice (multi.cpp:614).
type SimFlags struct {
	// Spawn is gbIsSpawn: only the shareware data set is available.
	Spawn bool
	// Hellfire is gbIsHellfire: the Hellfire expansion data is loaded.
	Hellfire bool
	// Multiplayer is gbIsMultiplayer.
	Multiplayer bool
	// FullQuests is Options.Gameplay.multiplayerFullQuests.
	FullQuests bool
}

// UseMultiplayerQuests mirrors UseMultiplayerQuests: sgGameInitInfo.fullQuests
// is (!gbIsMultiplayer || multiplayerFullQuests) ? 1 : 0, and the multiplayer
// quest table is used when it is 0 (questdat.cpp:59-62, multi.cpp:614).
func (f SimFlags) UseMultiplayerQuests() bool {
	return f.Multiplayer && !f.FullQuests
}

// QuestsEnabled reports whether quest progression runs at all: CheckQuests
// and CheckQuestKill both return immediately under gbIsSpawn
// (quests.cpp:135-136, 227-229).
func (f SimFlags) QuestsEnabled() bool {
	return !f.Spawn
}
