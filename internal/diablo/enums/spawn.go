// Spawn (shareware) gating per V12.9: all content gating is gbIsSpawn
// branches — restrict.cpp is only the read-only test (restrict.cpp:19-28).
package enums

// SpawnGate lists the systems gated on gbIsSpawn, with the upstream branch
// sites. The runtime flag itself lives in the game-state bean; these are the
// documented gate sites the port must honor.
const (
	// SpawnGateQuests: InitQuests wipes all quests to QUEST_NOTAVAIL
	// (quests.cpp:94-99); CheckQuests/CheckQuestKill early-return
	// (quests.cpp:139-141, :229); ForceQuests (quests.cpp:203-205).
	SpawnGateQuests = "quests.cpp:94-99,139-141,203-205,229"
	// SpawnGateMonsters: IsMonsterAvailable rejects availability==Retail
	// (monster.cpp:3159-3167).
	SpawnGateMonsters = "monster.cpp:3159-3167"
)
