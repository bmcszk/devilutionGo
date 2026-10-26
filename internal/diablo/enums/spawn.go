package enums

const (
	SpawnGateNone SpawnGate = iota
	SpawnGateQuests
	SpawnGateMonsters
)

// SpawnGate names a content system the shareware build gates on the runtime
// gbIsSpawn flag; the flag lives in game state, not here.
type SpawnGate uint8
