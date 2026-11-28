package monsterspawn

import (
	"fmt"
	"strconv"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/monstdat"
)

const (
	// Retail is a full game: always-available and retail monsters both spawn.
	Retail GameMode = iota
	// Shareware is gbIsSpawn set: retail monsters never spawn.
	Shareware
)

const (
	// AvailabilityNever never spawns (MonsterAvailability::Never).
	AvailabilityNever Availability = iota
	// AvailabilityAlways always spawns (MonsterAvailability::Always).
	AvailabilityAlways
	// AvailabilityRetail spawns only in retail builds (MonsterAvailability::Retail).
	AvailabilityRetail
)

// monstdat.tsv column indexes this package reads (header order of the committed
// table; monstdat.ParseMonstdat reads the same columns by index).
const (
	columnAvailability = 5
	columnMinDunLvl    = 12
	columnMaxDunLvl    = 13
)

// availabilityNames maps the monstdat availability column spelling to its
// MonsterAvailability ordinal (tables/monstdat.h:92-96).
var availabilityNames = map[string]Availability{
	"Never":  AvailabilityNever,
	"Always": AvailabilityAlways,
	"Retail": AvailabilityRetail,
}

// GameMode selects the shareware gate of IsMonsterAvailable: upstream's
// gbIsSpawn (monster.cpp:3162-3163).
type GameMode uint8

// Availability is a monstdat row's spawn class (MonsterAvailability,
// tables/monstdat.h:92-96).
type Availability uint8

// SpawnRule is the monstdat gating triple IsMonsterAvailable reads
// (monster.cpp:3158-3167).
type SpawnRule struct {
	// Availability is the monster's spawn class.
	Availability Availability
	// MinDunLvl and MaxDunLvl bound the dungeons it spawns on, inclusive.
	MinDunLvl, MaxDunLvl int
}

// IsMonsterAvailable is IsMonsterAvailable (monster.cpp:3158-3167): never-class
// monsters are out, shareware excludes retail-class monsters, and everything
// else is gated by the level band.
func IsMonsterAvailable(rule SpawnRule, mode GameMode, currLevel int) bool {
	if rule.Availability == AvailabilityNever {
		return false
	}

	if mode == Shareware && rule.Availability == AvailabilityRetail {
		return false
	}

	return currLevel >= rule.MinDunLvl && currLevel <= rule.MaxDunLvl
}

// ParseSpawnRule extracts the gate from one monstdat row: its availability class
// and its dungeon-level band.
func ParseSpawnRule(monster monstdat.MonsterData) (SpawnRule, error) {
	columns := monster.Raw()
	if len(columns) <= columnMaxDunLvl {
		return SpawnRule{}, fmt.Errorf("monsterspawn: %s has %d columns, want > %d",
			monster.ID, len(columns), columnMaxDunLvl)
	}

	availability, ok := availabilityNames[columns[columnAvailability]]
	if !ok {
		return SpawnRule{}, fmt.Errorf("monsterspawn: %s: unknown availability %q",
			monster.ID, columns[columnAvailability])
	}

	minDunLvl, err := parseColumnInt(monster.ID, "minDunLvl", columns[columnMinDunLvl])
	if err != nil {
		return SpawnRule{}, err
	}

	maxDunLvl, err := parseColumnInt(monster.ID, "maxDunLvl", columns[columnMaxDunLvl])
	if err != nil {
		return SpawnRule{}, err
	}

	return SpawnRule{Availability: availability, MinDunLvl: minDunLvl, MaxDunLvl: maxDunLvl}, nil
}

// parseColumnInt parses one integer monstdat column.
func parseColumnInt(id, column, raw string) (int, error) {
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("monsterspawn: %s: %s %q: %w", id, column, raw, err)
	}

	return value, nil
}
