// Spawn (shareware) gating per V12.9: all content gating is gbIsSpawn
// branches — restrict.cpp is only the read-only test (restrict.cpp:19-28).
package enums

// SpawnGate names a content system the shareware build gates on the runtime
// gbIsSpawn flag. The flag itself lives in game state; these members name the
// gated systems the port must honor.
type SpawnGate uint8

const (
	// SpawnGateNone is the zero value: no gated system.
	SpawnGateNone SpawnGate = iota
	// SpawnGateQuests: InitQuests wipes all quests to QUEST_NOTAVAIL
	// (quests.cpp:94-99); CheckQuests/CheckQuestKill early-return
	// (quests.cpp:139-141, :229); ForceQuests (quests.cpp:203-205).
	SpawnGateQuests
	// SpawnGateMonsters: IsMonsterAvailable rejects availability==Retail
	// (monster.cpp:3159-3167).
	SpawnGateMonsters
)
