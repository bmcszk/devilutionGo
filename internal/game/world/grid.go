// Package world ports the DRLG world-grid state and tile properties:
// dun_tile_data.hpp:26-120 (arrays, DungeonFlag, MegaTile), gendung.h
// Miniset, tile_properties.cpp:11-41 + CanStep (:66-86).
package world

import "github.com/bmcszk/devilutionGo/internal/engine"

// Grid dimensions (gendung.h MAXDUNX/MAXDUNY).
const (
	MAXDUNX = 112
	MAXDUNY = 112
)

const (
	TileNone         TileProperties = 0
	TileSolid        TileProperties = 1 << 0
	TileBlockMissile TileProperties = 1 << 1
	TileOpaque       TileProperties = 1 << 2
	TileTrap         TileProperties = 1 << 3
)

const (
	FlagNone         DungeonFlag = 0
	FlagMissile      DungeonFlag = 1 << 0
	FlagVisible      DungeonFlag = 1 << 1
	FlagDeadPlayer   DungeonFlag = 1 << 2
	FlagPopulated    DungeonFlag = 1 << 3
	FlagMissileFire  DungeonFlag = 1 << 4
	FlagMissileLight DungeonFlag = 1 << 5
	FlagLit          DungeonFlag = 1 << 6
	FlagExplored     DungeonFlag = 1 << 7
)

// SavedFlags = ~(Missile | Visible | DeadPlayer) (dun_tile_data.hpp:50).
// LoadedFlags (dun_tile_data.hpp:51).
const (
	SavedFlags  = FlagPopulated | FlagLit | FlagExplored
	LoadedFlags = FlagMissile | FlagVisible | FlagDeadPlayer | FlagPopulated | FlagLit | FlagExplored
)

// TileProperties mirrors tile_properties.hpp flag bits (SOL entry).
type TileProperties uint16

// DungeonFlag mirrors dun_tile_data.hpp:32-49 (saved/loaded flag bits).
type DungeonFlag uint8

// WorldObject is the IsTileWalkable-visible object slice
// (objects.h _oSolidFlag / isDoor) plus the two flags CheckMissileCol /
// IsMissileBlockedByTile read (devilGo-mwlm slice 2):
// _oMissFlag (objects.h:54 — true if missiles pass through) and
// _oBreak (objects.h:51 — IsBreakable is _oBreak == 1, objects.h:154-157).
type WorldObject struct {
	SolidFlag bool
	IsDoor    bool
	// MissFlag ports _oMissFlag (objects.h:54): missiles pass through.
	MissFlag bool
	// Breakable ports IsBreakable() (objects.h:154-157, _oBreak == 1).
	Breakable bool
}

// Grid holds the world arrays the tile-property predicates read. Upstream
// these are globals (gendung.h); here they are one struct so tests can
// script them.
type Grid struct {
	// DFlags: the DungeonFlag lookup table (dFlags, gendung.h),
	// indexed [y][x] like dPiece; PutMissile registers missiles here.
	DFlags [MAXDUNY][MAXDUNX]DungeonFlag
	// DPiece holds the SOL/micro piece index per tile. Upstream dPiece is
	// [MAXDUNX][MAXDUNY] indexed [x][y] (dun_tile_data.hpp:99; the same
	// [x][y] layout as dMonster at :118). This Go array transposes it to
	// [y][x] on purpose — do not transpose again when porting.
	DPiece [MAXDUNY][MAXDUNX]int
	// dObject: 0 = no object, else 1-based Objects index.
	DObject [MAXDUNY][MAXDUNX]int
	// SOLData: per-piece tile properties (LoadLevelSOLData fills it).
	SOLData []TileProperties
	// Objects: 0-based object storage; dObject stores index+1.
	Objects []WorldObject
}

// NewGrid builds an empty grid with a SOL table covering the base piece ids.
func NewGrid() *Grid {
	return &Grid{SOLData: make([]TileProperties, 1379)} // MAXTILES
}

// InDungeonBounds mirrors InDungeonBounds (point.hpp).
func InDungeonBounds(x, y int) bool {
	return x >= 0 && x < MAXDUNX && y >= 0 && y < MAXDUNY
}

// GetPathDirection delegates to the engine port (path.cpp:178-182).
func GetPathDirection(sx, sy, dx, dy int) int8 {
	return engine.GetPathDirection(engine.Point{X: sx, Y: sy}, engine.Point{X: dx, Y: dy})
}

// directionDelta mirrors DisplacementOf::fromDirection
// (engine/displacement.hpp:247-275): {dx, dy} per Direction enum order.
func directionDelta(d int8) (deltaX, deltaY int) {
	switch d {
	case 0: // South
		return 1, 1
	case 1: // SouthWest
		return 0, 1
	case 2: // West
		return -1, 1
	case 3: // NorthWest
		return -1, 0
	case 4: // North
		return -1, -1
	case 5: // NorthEast
		return 0, -1
	case 6: // East
		return 1, -1
	case 7: // SouthEast
		return 1, 0
	default: // NoDirection
		return 0, 0
	}
}

// TileHasAny reports whether the tile's SOL entry has any of the flags.
func (g *Grid) TileHasAny(x, y int, flags TileProperties) bool {
	piece := g.DPiece[y][x]
	if piece < 0 || piece >= len(g.SOLData) {
		return false
	}

	return g.SOLData[piece]&flags != 0
}

// IsTileNotSolid ports tile_properties.cpp:11-19.
func (g *Grid) IsTileNotSolid(x, y int) bool {
	if !InDungeonBounds(x, y) {
		return false
	}

	return !g.TileHasAny(x, y, TileSolid)
}

// IsTileSolid ports tile_properties.cpp:21-29.
func (g *Grid) IsTileSolid(x, y int) bool {
	if !InDungeonBounds(x, y) {
		return false
	}

	return g.TileHasAny(x, y, TileSolid)
}

// IsTileWalkable ports tile_properties.cpp:31-45 (object occupancy,
// doors not ignored).
func (g *Grid) IsTileWalkable(x, y int) bool {
	if obj := g.FindObjectAtPosition(x, y); obj != nil {
		if obj.SolidFlag {
			return false
		}
	}

	return g.IsTileNotSolid(x, y)
}

// IsTileWalkableIgnoreDoors is the ignoreDoors=true variant of
// IsTileWalkable (tile_properties.cpp:31-45): doors do not block.
func (g *Grid) IsTileWalkableIgnoreDoors(x, y int) bool {
	if obj := g.FindObjectAtPosition(x, y); obj != nil {
		if obj.IsDoor {
			return true
		}

		if obj.SolidFlag {
			return false
		}
	}

	return g.IsTileNotSolid(x, y)
}

// FindObjectAtPosition resolves dObject's 1-based index.
func (g *Grid) FindObjectAtPosition(x, y int) *WorldObject {
	idx := g.DObject[y][x]
	if idx == 0 || idx > len(g.Objects) {
		return nil
	}

	return &g.Objects[idx-1]
}

// CanStep ports tile_properties.cpp:66-86: diagonal steps check the two
// tiles flanking the corner; cardinal steps are always allowed.
func (g *Grid) CanStep(sx, sy, dx, dy int) bool {
	rv := true

	switch GetPathDirection(sx, sy, dx, dy) {
	case 5: // Stepping north
		a, b := directionDelta(1) // SouthWest
		c, d := directionDelta(7) // SouthEast
		rv = g.IsTileNotSolid(dx+a, dy+b) && g.IsTileNotSolid(dx+c, dy+d)
	case 6: // Stepping east
		a, b := directionDelta(1) // SouthWest
		c, d := directionDelta(3) // NorthWest
		rv = g.IsTileNotSolid(dx+a, dy+b) && g.IsTileNotSolid(dx+c, dy+d)
	case 7: // Stepping south
		a, b := directionDelta(5) // NorthEast
		c, d := directionDelta(3) // NorthWest
		rv = g.IsTileNotSolid(dx+a, dy+b) && g.IsTileNotSolid(dx+c, dy+d)
	case 8: // Stepping west
		a, b := directionDelta(7) // SouthEast
		c, d := directionDelta(5) // NorthEast
		rv = g.IsTileNotSolid(dx+a, dy+b) && g.IsTileNotSolid(dx+c, dy+d)
	default:
		// Non-adjacent / same-tile steps: upstream switch covers 5-8 only.
	}

	return rv
}
