package levels

import (
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/leveldat"
	"github.com/bmcszk/devilutionGo/internal/light"
)

// This file ports the level-state half of the lighting subsystem: the dLight
// grid and the DoLighting write pass (Source/lighting.cpp:64-82, 117-180) that
// the static level light passes drive. The dynamic light list (AddLight,
// ProcessLightList, Source/lighting.cpp:254-430) and vision are separate ports.

// Light is one entry of the static light-source list a level pass hands to
// [DoLighting].
type Light struct {
	// Position is the world tile the light is centred on.
	Position Point
	// Radius indexes LightFalloffs (Source/lighting.cpp:170).
	Radius uint8
}

// Offset is DisplacementOf<int8_t> (engine/displacement.hpp:14-30): the
// sub-tile light offset DoLighting uses to pick a light cone.
type Offset struct{ DeltaX, DeltaY int8 }

// lightCone is the walking state of one DoLighting call: the cone origin, the
// four rotating offsets and the per-quadrant bounds
// (Source/lighting.cpp:122-151).
type lightCone struct {
	position Point
	offset   Offset
	dist     Offset
	lit      Offset
	block    Offset
	minX     int
	maxX     int
	minY     int
	maxY     int
}

// InitLight resets `dLight` to InitGlobals' default (levels/gendung.cpp:301-307):
// 0 in town, LightsMax everywhere else.
func (d *Dungeon) InitLight() {
	defaultLight := uint8(light.LightsMax)
	if d.LevelType == Town {
		defaultLight = 0
	}

	d.Light = [MAXDUNX][MAXDUNY]uint8{}
	for x := range d.Light {
		for y := range d.Light[x] {
			d.Light[x][y] = defaultLight
		}
	}
}

// ApplyLights stamps a finished level's static light sources into dLight. It is
// the single build-time entry point for the three dPiece-driven families:
// CreateL5Dungeon calls the crypt pass (levels/drlg_l1.cpp:1305-1312) and
// CreateL3Dungeon PlaceLights (levels/drlg_l3.cpp:2198-2199), which dispatches to
// the caves or nest pass (:2180-2188), all after Pass3.
//
// The light tables are built here rather than read from a global, so a test
// scripts the level type without touching package state.
func ApplyLights(d *Dungeon, pieces PieceGrid) {
	tables := light.MakeLightTables(levelTypeDungeonType(d.LevelType))

	for _, source := range levelLightSources(d, pieces) {
		DoLighting(d, tables, source.Position, source.Radius, Offset{})
	}
}

// DoLighting lights every tile within 15 world tiles of position down to
// LightFalloffs[radius][distance], and only ever brightens a tile: a lower
// dLight value means a brighter tile (Source/lighting.cpp:117-180).
// position must be inside the dungeon grid, as upstream asserts (:120).
func DoLighting(d *Dungeon, tables light.LightTables, position Point, radius uint8, offset Offset) {
	cone := newLightCone(position, offset)
	cone.brightenCentre(d, tables, radius)

	for quadrant := range 4 {
		cone.stampQuadrant(d, tables, radius, quadrant)
		cone.rotate()
	}
}

// levelLightSources is the static light collection of one level type. Only the
// three dPiece-driven families this package owns are modelled; the town,
// catacombs and hell generators call DoLighting nowhere.
func levelLightSources(d *Dungeon, pieces PieceGrid) []Light {
	switch d.LevelType {
	case Crypt:
		return PlaceCryptLights(pieces)
	case Caves:
		return PlaceCaveLights(pieces)
	case Nest:
		return PlaceHiveLights(pieces)
	default:
		return nil
	}
}

// levelTypeDungeonType maps the generator's dungeon_type onto the light table
// package's copy. Both enums are the same upstream enum
// (levels/gendung_defs.hpp:15-26), so the ordinals agree.
func levelTypeDungeonType(levelType LevelType) leveldat.DungeonType {
	return leveldat.DungeonType(levelType)
}

// inDungeonBounds is InDungeonBounds (levels/dun_tile_data.hpp:158-161).
func inDungeonBounds(p Point) bool {
	return p.X >= 0 && p.X < MAXDUNX && p.Y >= 0 && p.Y < MAXDUNY
}

// newLightCone resolves DoLighting's cone origin and bounds. A negative offset
// belongs to the previous tile's cone quadrant (Source/lighting.cpp:125-132).
func newLightCone(position Point, offset Offset) *lightCone {
	if offset.DeltaX < 0 {
		offset.DeltaX += 8
		position.X--
	}
	if offset.DeltaY < 0 {
		offset.DeltaY += 8
		position.Y--
	}

	cone := &lightCone{position: position, offset: offset, dist: offset}
	cone.minX, cone.maxX = coneExtent(position.X, MAXDUNX)
	cone.minY, cone.maxY = coneExtent(position.Y, MAXDUNY)

	return cone
}

// coneExtent is DoLighting's per-axis cone extent
// (Source/lighting.cpp:136-151): 15 tiles, narrowed at the north and west
// dungeon edges.
func coneExtent(coord, limit int) (low, high int) {
	low, high = 15, 15
	if coord-15 < 0 {
		low = coord + 1
	}
	if coord+15 > limit {
		high = limit - coord
	}

	return low, high
}

// brightenCentre writes the cone origin tile. Crypt and nest allow a dim base
// light instead of full brightness (Source/lighting.cpp:153-159).
func (c *lightCone) brightenCentre(d *Dungeon, tables light.LightTables, radius uint8) {
	if d.LevelType == Nest || d.LevelType == Crypt {
		c.dimCentre(d, tables, radius)
		return
	}

	d.Light[c.position.X][c.position.Y] = 0
}

// dimCentre lowers the origin only as far as the radius-0 falloff.
func (c *lightCone) dimCentre(d *Dungeon, tables light.LightTables, radius uint8) {
	if shade := tables.Falloffs[radius][0]; shade < d.Light[c.position.X][c.position.Y] {
		d.Light[c.position.X][c.position.Y] = shade
	}
}

// stampQuadrant writes one quadrant of the light cone
// (Source/lighting.cpp:161-177). The middle two quadrants reach further along y
// than the outer ones (upstream's yBound switch on i).
func (c *lightCone) stampQuadrant(d *Dungeon, tables light.LightTables, radius uint8, quadrant int) {
	for y := range c.yBound(quadrant) {
		for x := 1; x < c.xBound(quadrant); x++ {
			c.stampTile(d, tables, radius, quadrant, Point{X: x, Y: y})
		}
	}
}

// yBound is upstream's `i > 0 && i < 3 ? maxY : minY`.
func (c *lightCone) yBound(quadrant int) int {
	if quadrant > 0 && quadrant < 3 {
		return c.maxY
	}

	return c.minY
}

// xBound is upstream's `i < 2 ? maxX : minX`.
func (c *lightCone) xBound(quadrant int) int {
	if quadrant < 2 {
		return c.maxX
	}

	return c.minX
}

// stampTile writes one cone cell if the cone reaches it and it is inside the
// dungeon (Source/lighting.cpp:164-176).
func (c *lightCone) stampTile(
	d *Dungeon, tables light.LightTables, radius uint8, quadrant int, at Point,
) {
	distance := tables.Cone[c.offset.DeltaX][c.offset.DeltaY][at.X+int(c.block.DeltaX)][at.Y+int(c.block.DeltaY)]
	if distance >= 128 {
		return
	}

	target := c.position.Add(at.rotate(-quadrant))
	if !inDungeonBounds(target) {
		return
	}

	if shade := tables.Falloffs[radius][distance]; shade < d.Light[target.X][target.Y] {
		d.Light[target.X][target.Y] = shade
	}
}

// rotate is RotateRadius (Source/lighting.cpp:52-66): it turns the cone a
// quadrant and re-derives the wrapped offset plus its block correction.
func (c *lightCone) rotate() {
	c.dist.DeltaX, c.dist.DeltaY = 7-c.dist.DeltaY, c.dist.DeltaX
	c.lit.DeltaX, c.lit.DeltaY = 7-c.lit.DeltaY, c.lit.DeltaX
	c.offset.DeltaX, c.offset.DeltaY = c.dist.DeltaX-c.lit.DeltaX, c.dist.DeltaY-c.lit.DeltaY

	c.block.DeltaX = 0
	if c.offset.DeltaX < 0 {
		c.offset.DeltaX += 8
		c.block.DeltaX = 1
	}

	c.block.DeltaY = 0
	if c.offset.DeltaY < 0 {
		c.offset.DeltaY += 8
		c.block.DeltaY = 1
	}
}
