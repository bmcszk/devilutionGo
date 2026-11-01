package game

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/bmcszk/devilutionGo/internal/engine"
)

// parseSpawnMonsterArgs splits "<name...> [count]": the trailing token is the
// count only when it parses as a number (monsters.cpp:117 optional count).
func parseSpawnMonsterArgs(args []string) (string, int, error) {
	if len(args) == 0 {
		return "", 0, errors.New("debug: spawn-monster needs a monster name")
	}

	count := 1

	name := args
	if len(args) > 1 {
		if parsed, err := strconv.Atoi(args[len(args)-1]); err == nil {
			count = parsed
			name = args[:len(args)-1]
		}
	}

	joined := strings.Join(name, " ")
	if joined == "" {
		return "", 0, errors.New("debug: spawn-monster needs a monster name")
	}

	return joined, count, nil
}

// parseLevelArg reads the single dungeon level argument of jump-level.
func parseLevelArg(args []string) (int, error) {
	if len(args) != 1 {
		return 0, errors.New("debug: jump-level needs exactly one level number")
	}

	level, err := strconv.Atoi(args[0])
	if err != nil || level < 1 {
		return 0, fmt.Errorf("debug: jump-level %q is not a level number", args[0])
	}

	return level, nil
}

// spawnMonster ports DebugCmdSpawnMonster
// (Source/lua/modules/dev/monsters.cpp:98-168): resolve the name, take the
// first walkable and unoccupied crawl position from the player, and stop once
// the requested count or the free monster slots run out.
func (r *DebugRouter) spawnMonster(name string, count int) (string, error) {
	env := r.env

	err := env.spawnHooksReady()
	if err != nil {
		return "", err
	}

	if env.LevelType() == LevelTypeTown {
		return "Can't spawn monsters in town", nil
	}

	if count < 1 {
		return "count must be positive", nil
	}

	monsterTypeID := env.MonsterTypeID(strings.ToLower(name))
	if monsterTypeID < 0 {
		return "Monster not found", nil
	}

	free := MaxMonsters - env.ActiveMonsters()
	if free <= 0 {
		return "Can't spawn any monsters", nil
	}

	toSpawn := min(count, free)
	spawned := env.crawlSpawn(monsterTypeID, toSpawn)

	// monsters.cpp:165 compares the clamped count with the requested one, so a
	// full monster array reports the "(Unable to spawn more)" suffix even when
	// every free slot was filled; the crawl running dry stays a plain report.
	if toSpawn != count {
		return fmt.Sprintf("Spawned %d monsters. (Unable to spawn more)", spawned), nil
	}

	return fmt.Sprintf("Spawned %d monsters.", spawned), nil
}

// spawnHooksReady reports whether every hook the monster spawn needs is wired.
func (env DebugEnv) spawnHooksReady() error {
	if env.LevelType == nil || env.PlayerTile == nil || env.PlayerDir == nil ||
		env.Occupied == nil || env.Walkable == nil || env.ActiveMonsters == nil ||
		env.MonsterTypeID == nil || env.SpawnMonster == nil {
		return ErrDebugDisabled
	}

	return nil
}

// crawlSpawn walks the ring search from the player tile (monsters.cpp:152-164)
// and spawns until it has placed toSpawn monsters; it returns the placed count.
// The crawl itself draws no random values; SpawnMonster owns the monster RNG.
func (env DebugEnv) crawlSpawn(monsterTypeID, toSpawn int) int {
	origin := env.PlayerTile()
	dir := env.PlayerDir()
	spawned := 0

	engine.DoCrawl(0, DebugMaxCrawlRadius, func(displacement engine.Displacement) bool {
		tile := DebugTile{X: origin.X + displacement.DeltaX, Y: origin.Y + displacement.DeltaY}
		if env.Occupied(tile) || !env.Walkable(tile) {
			return true
		}

		env.SpawnMonster(tile, dir, monsterTypeID)
		spawned++

		return spawned < toSpawn
	})

	return spawned
}

// spawnItem ports DebugCmdSpawnItem (Source/lua/modules/dev/items.cpp:85-119).
// Parity exclusion: upstream seeds the LCG from BetterRng (items.cpp:87-89) so
// the post-cheat RNG state is not reproducible; only the message is pinned.
func (r *DebugRouter) spawnItem(name string) (string, error) {
	if r.env.SpawnItem == nil {
		return "", ErrDebugDisabled
	}

	iterations, err := r.env.SpawnItem(name)
	if err != nil {
		return "", fmt.Errorf("debug: spawn item %q: %w", name, err)
	}

	return fmt.Sprintf("Item generated successfully - iterations: %d", iterations), nil
}

// jumpLevel ports DebugCmdWarpToDungeonLevel
// (Source/lua/modules/dev/level/warp.cpp:18-30). Quest-level warps
// (warp.cpp:32-50) and the tristram return (level 21) need the quest state and
// are out of scope; JumpLevel carries upstream's StartNewLvl mode.
func (r *DebugRouter) jumpLevel(level int) (string, error) {
	if r.env.CurrentLevel == nil || r.env.JumpLevel == nil {
		return "", ErrDebugDisabled
	}

	maxLevel := r.env.MaxLevel
	if maxLevel <= 0 {
		maxLevel = 16 // warp.cpp:21 gbIsHellfire ? 24 : 16
	}

	if level > maxLevel {
		return fmt.Sprintf("Level %d does not exist!", level), nil
	}

	if level == r.env.CurrentLevel() {
		return fmt.Sprintf("You are already on level %d!", level), nil
	}

	err := r.env.JumpLevel(level)
	if err != nil {
		return "", fmt.Errorf("debug: jump to level %d: %w", level, err)
	}

	return fmt.Sprintf("Moved you to level %d.", level), nil
}
