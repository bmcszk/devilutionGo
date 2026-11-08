package light

// This file ports the vision half of the dungeon state the level layer owns:
// the dFlags bits the reveal passes write (levels/dun_tile_data.hpp:34-47) and
// the automap exploration kinds they pass on (automap.h:19-29).

const (
	// MaxVision is MAXVISION (engine/lighting_defs.hpp:8): the vision pool is
	// one slot per player.
	MaxVision = 4

	// FlagNone is DungeonFlag::None (levels/dun_tile_data.hpp:36).
	FlagNone DungeonFlag = 0
	// FlagMissile marks a tile holding a missile (dun_tile_data.hpp:37).
	FlagMissile DungeonFlag = 1 << 0
	// FlagVisible is DungeonFlag::Visible, set by DoVisionFlags
	// (dun_tile_data.hpp:38) and cleared by DoUnVision (lighting.cpp:190).
	FlagVisible DungeonFlag = 1 << 1
	// FlagDeadPlayer marks a tile holding a player corpse (:39).
	FlagDeadPlayer DungeonFlag = 1 << 2
	// FlagPopulated marks a tile holding an object (:40).
	FlagPopulated DungeonFlag = 1 << 3
	// FlagMissileFireWall marks a fire-wall missile (:41).
	FlagMissileFireWall DungeonFlag = 1 << 4
	// FlagMissileLightningWall marks a lightning-wall missile (:42).
	FlagMissileLightningWall DungeonFlag = 1 << 5
	// FlagLit is DungeonFlag::Lit: the tile is in a lit area (:43).
	FlagLit DungeonFlag = 1 << 6
	// FlagExplored is DungeonFlag::Explored: the automap already shows it (:44).
	FlagExplored DungeonFlag = 1 << 7
)

// The automap exploration kinds, in their own block so iota starts at
// MAP_EXP_NONE: upstream lists them in automap.h:19-29 and the ordinal is what
// the automap stores.
const (
	// ExpNone is MAP_EXP_NONE (automap.h:20): the tile stays unexplored.
	ExpNone Exploration = iota
	// ExpOld is MAP_EXP_OLD (:22), vanilla compatibility.
	ExpOld
	// ExpShrine is MAP_EXP_SHRINE (:24), explored by a shrine.
	ExpShrine
	// ExpOthers is MAP_EXP_OTHERS (:26), explored by another player in
	// multiplayer.
	ExpOthers
	// ExpSelf is MAP_EXP_SELF (:28), explored by the current player.
	ExpSelf
)

// DungeonFlag is DungeonFlag (levels/dun_tile_data.hpp:34-47): the per-tile
// bit set the vision, automap and missile passes share.
type DungeonFlag uint8

// Exploration is MapExplorationType (automap.h:19-29): who explored a tile.
type Exploration uint8
