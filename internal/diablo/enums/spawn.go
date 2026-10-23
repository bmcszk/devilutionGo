package enums

// spawnGate names a content system the shareware build gates on the runtime
// gbIsSpawn flag; the flag lives in game state, not here.
type spawnGate uint8

const (
	SpawnGateNone spawnGate = iota
	SpawnGateQuests
	SpawnGateMonsters
)
