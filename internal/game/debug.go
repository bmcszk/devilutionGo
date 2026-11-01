package game

import (
	"errors"
	"fmt"
	"strings"
)

// MaxMonsters is monster.h:39 MaxMonsters, the fixed monster array size.
const MaxMonsters = 200

// DebugMaxCrawlRadius ports MaxCrawlRadius (lighting.h:74), the search radius
// the spawn cheats crawl from the player tile.
const DebugMaxCrawlRadius = 18

// ErrDebugDisabled reports that no debug router exists, or that the command's
// game hook was never wired: cheats fail closed.
var ErrDebugDisabled = errors.New("debug: commands disabled")

// DebugTile is a dungeon tile coordinate.
type DebugTile struct{ X, Y int }

// DebugEnv is the game surface the cheats drive, injected so the router stays
// free of level, item, monster and renderer state. A nil hook makes its
// command fail closed with ErrDebugDisabled.
type DebugEnv struct {
	// LevelType reports the current leveltype: LevelTypeTown rejects monster
	// spawns (monsters.cpp:26).
	LevelType func() LevelType
	// PlayerTile is MyPlayer.position.tile, the crawl origin (monsters.cpp:153).
	PlayerTile func() DebugTile
	// PlayerDir is MyPlayer._pdir, the spawned monster's facing
	// (monsters.cpp:162).
	PlayerDir func() int
	// Occupied reports dPlayer/dMonster occupancy (monsters.cpp:154).
	Occupied func(DebugTile) bool
	// Walkable reports IsTileWalkable (monsters.cpp:156).
	Walkable func(DebugTile) bool
	// ActiveMonsters is ActiveMonsterCount (monsters.cpp:149).
	ActiveMonsters func() int
	// MonsterTypeID resolves a lower-case name substring to a monster type;
	// -1 means not found (monsters.cpp:100-120).
	MonsterTypeID func(name string) int
	// SpawnMonster places one monster of the given type facing dir
	// (monsters.cpp:158); this hook owns upstream's monster RNG draws.
	SpawnMonster func(tile DebugTile, dir, monsterTypeID int)
	// SpawnItem generates an item matching the name and reports the iteration
	// count (items.cpp:85-119); upstream seeds the LCG from BetterRng, so the
	// hook owns that non-deterministic seeding.
	SpawnItem func(name string) (iterations int, err error)
	// CurrentLevel is currlevel (warp.cpp:24).
	CurrentLevel func() int
	// MaxLevel is the highest dungeon level: 16, or 24 for Hellfire
	// (warp.cpp:21). Zero falls back to 16.
	MaxLevel int
	// JumpLevel performs StartNewLvl (warp.cpp:29).
	JumpLevel func(level int) error
}

// DebugRouter dispatches the chat-style cheat commands that upstream keeps
// behind #ifdef _DEBUG (Source/lua/modules/dev/*.cpp): spawn an item, spawn
// monsters, jump to a dungeon level.
type DebugRouter struct{ env DebugEnv }

// NewDebugRouter returns an enabled router over env. A nil *DebugRouter is the
// shipped default and makes every command a no-op; upstream gates the same
// surface on the _DEBUG build flag, Go builds opt in explicitly (see
// NewDebugRouterFromFlag) instead of duplicating the tree behind a build tag.
func NewDebugRouter(env DebugEnv) *DebugRouter {
	return &DebugRouter{env: env}
}

// Execute runs one cheat command line and returns upstream's status message
// (for example "Spawned 2 monsters."). Malformed commands return an error and
// never reach a hook.
func (r *DebugRouter) Execute(line string) (string, error) {
	if r == nil {
		return "", ErrDebugDisabled
	}

	fields := strings.Fields(line)
	if len(fields) == 0 {
		return "", errors.New("debug: empty command")
	}

	return r.dispatch(fields[0], fields[1:])
}

// dispatch routes one parsed command to its cheat.
func (r *DebugRouter) dispatch(verb string, args []string) (string, error) {
	switch verb {
	case "spawn-monster":
		name, count, err := parseSpawnMonsterArgs(args)
		if err != nil {
			return "", err
		}

		return r.spawnMonster(name, count)
	case "spawn-item":
		name := strings.Join(args, " ")
		if name == "" {
			return "", errors.New("debug: spawn-item needs an item name")
		}

		return r.spawnItem(name)
	case "jump-level":
		level, err := parseLevelArg(args)
		if err != nil {
			return "", err
		}

		return r.jumpLevel(level)
	default:
		return "", fmt.Errorf("debug: unknown command %q", verb)
	}
}

// ExecuteDebug runs one cheat command line through the game's router and
// returns upstream's status message. Without a router (the shipped default) it
// fails closed with ErrDebugDisabled and touches no game state.
func (g *Game) ExecuteDebug(line string) (string, error) {
	return g.Debug.Execute(line)
}
