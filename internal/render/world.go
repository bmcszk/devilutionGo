package render

import "fmt"

// This file ports the static world pass of the dungeon rasteriser:
// DrawFloor, DrawTileContent and DrawOOB (Source/engine/render/scrollrt.cpp:
// 955-1116) plus the isometric projection (Source/engine/displacement.hpp:
// 151-164) and the per-tile layer order of DrawDungeon (scrollrt.cpp:782-953).
//
// The pass is written against a Sink so the same walk drives the index-buffer
// rasteriser and an order-recording test; the tile grids stay behind World so
// the render layer never imports the Diablo level packages.

const (
	// TileWidth is TILE_WIDTH (levels/dun_tile.hpp:7).
	TileWidth = 64
	// TileHeight is TILE_HEIGHT (levels/dun_tile.hpp:8).
	TileHeight = 32
	// LightsMax is the first light index at which corpses and objects are
	// skipped: `lightTableIndex < LightsMax` (engine/lighting_defs.hpp:11).
	LightsMax = 15
	// MaxDunX is MAXDUNX, 16 + DMAXX*2 + 16 with DMAXX = 40
	// (levels/gendung_defs.hpp:6-9).
	MaxDunX = 112
	// MaxDunY is MAXDUNY of the same definition.
	MaxDunY = 112
	// animationFraction is AnimationInfo::baseValueFraction, the fixed-point
	// scale of a walk progress (engine/animationinfo.h:114).
	animationFraction = 128
	// oobExtraRows is DrawOOB's "+5 extra rows ... bottom half of the screen"
	// (scrollrt.cpp:1090).
	oobExtraRows = 5
)

// The eight walking directions plus NoDirection, and the player modes the tile
// order reacts to. Both blocks are typed with the enums declared below; Go does
// not require a type to be declared before its constants.
const (
	South Direction = iota
	SouthWest
	West
	NorthWest
	North
	NorthEast
	East
	SouthEast
	NoDirection
)

// The player modes the tile order reacts to; every other mode draws in place.
const (
	ModeStand PlayerMode = iota
	ModeMoveNorthwards
	ModeMoveSouthwards
	ModeMoveSideways
)

// movingOffsets is MovingOffset, indexed South..SouthEast
// (scrollrt.cpp:1578-1580).
var movingOffsets = [8]PixelPos{
	{X: 0, Y: 32},
	{X: -32, Y: 16},
	{X: -64, Y: 0},
	{X: -32, Y: -16},
	{X: 0, Y: -32},
	{X: 32, Y: -16},
	{X: 64, Y: 0},
	{X: 32, Y: 16},
}

// Direction is `enum class Direction`, ordinal order preserved
// (engine/direction.hpp:12-22).
type Direction uint8

// PlayerMode is the walking part of PLR_MODE (player.h:110-123).
type PlayerMode uint8

// TilePos is one dPiece coordinate.
type TilePos struct{ X, Y int }

// PixelPos is one target-buffer coordinate.
type PixelPos struct{ X, Y int }

// Occupant is an entity standing on a tile. Encoded is the signed dPlayer /
// dMonster value: negative means the entity is mid-walk and is drawn from the
// tile it is leaving (player.cpp:2012-2017).
type Occupant struct {
	ID      int
	Encoded int8
	Mode    PlayerMode
	Dir     Direction
}

// ObjectRef is the dObject cell's render-relevant half (objects.h).
type ObjectRef struct {
	Pre bool
}

// World is the tile-grid view the world pass reads. Every callback is a table
// lookup upstream (scrollrt.cpp:784-953 reads the dPiece/dLight/dSpecial/...
// globals directly); injecting them keeps the render layer free of the level
// packages and lets tests script a scene.
type World struct {
	// InBounds is InDungeonBounds (levels/dun_tile_data.hpp:143-146).
	InBounds func(TilePos) bool
	// IsFloor is `!TileHasAny(tile, Solid|BlockMissile)`
	// (scrollrt.cpp:114-117).
	IsFloor func(TilePos) bool
	// IsWall is `!IsFloor(tile) || dSpecial[tile] != 0` (scrollrt.cpp:119-122).
	IsWall func(TilePos) bool
	// NotSolid is IsTileNotSolid (levels/tile_properties.cpp:11-19).
	NotSolid func(TilePos) bool
	// Light is dLight (levels/dun_tile_data.cpp:54).
	Light func(TilePos) uint8
	// Piece is dPiece, the MICROS entry of the tile's cel.
	Piece func(TilePos) uint16
	// Special is dSpecial, the arch/special cel index.
	Special func(TilePos) int8
	// Transparent is TransList[dTransVal[tile]] (scrollrt.cpp:565).
	Transparent func(TilePos) bool
	// Corpse is dCorpse (1-based; 0 = none).
	Corpse func(TilePos) int8
	// Item is dItem (1-based; 0 = none).
	Item func(TilePos) int8
	// ItemPostDraw is Items[id]._iPostDraw (scrollrt.cpp:820, :906).
	ItemPostDraw func(int8) bool
	// DeadPlayer is TileContainsDeadPlayer (levels/dun_tile_data.hpp:163).
	DeadPlayer func(TilePos) bool
	// Object is FindObjectAtPosition (objects.cpp).
	Object func(TilePos) (ObjectRef, bool)
	// Player is PlayerAtPosition (player.cpp:2040-2050).
	Player func(TilePos) (Occupant, bool)
	// Monster is FindMonsterAtPosition (monster.cpp:4637).
	Monster func(TilePos) (Occupant, bool)
	// MissilePre is MissilePreFlag.
	MissilePre bool
	// Town is `leveltype == DTYPE_TOWN`, which selects the tree-leaf deferral
	// and the black OOB tile (scrollrt.cpp:910, :1093).
	Town bool
}

// DrawTarget is where one layer draw lands: the tile being drawn, its dPiece
// position (which for a walking player is the tile it is leaving), the
// target-buffer coordinate and the light table the draw keys on.
type DrawTarget struct {
	Tile  TilePos
	At    PixelPos
	Light uint8
}

// Sink receives the layer draws in the upstream order. The index rasteriser and
// the order-recording test are the two implementations.
type Sink interface {
	DrawFloorTile(DrawTarget)
	DrawCell(DrawTarget)
	DrawMissilePre(DrawTarget)
	DrawCorpse(DrawTarget)
	DrawObjectPre(DrawTarget)
	DrawItemPre(DrawTarget)
	DrawDeadPlayer(DrawTarget)
	DrawPlayer(DrawTarget, int)
	DrawMonster(DrawTarget, int)
	DrawItemPost(DrawTarget)
	DrawObjectPost(DrawTarget)
	DrawMissilePost(DrawTarget)
	DrawArch(DrawTarget, int8)
	DrawArchBlended(DrawTarget, int8)
	DrawArchDirect(DrawTarget, int8)
	DrawBlackTile(PixelPos)
}

// View is one frame's worth of tile-walk state: the first tile of the view, its
// target-buffer origin, the row and column counts and the micro-tile row count
// DrawGame computes (scrollrt.cpp:1237-1283).
type View struct {
	Start   TilePos
	Origin  PixelPos
	Rows    int
	Columns int
	// ScreenWidth is gnScreenWidth, the clipping bound of the wall pre-pass
	// (scrollrt.cpp:1009).
	ScreenWidth int
	// MicroTiles is MicroTileLen: 10, 12 for hell levels, 16 for town
	// (levels/dun_tile_data.cpp:166-174).
	MicroTiles int
}

// DrawWorld runs the three world passes in DrawGame's order: floor, tile
// content, then the out-of-bounds fill (scrollrt.cpp:1281-1283).
func DrawWorld(sink Sink, world World, view View) {
	DrawFloor(sink, world, view)
	DrawTileContent(sink, world, view)
	DrawOOB(sink, world, view)
}

// DrawFloor renders the floor pass: every floor tile in the diamond row walk,
// in visit order (scrollrt.cpp:955-981).
func DrawFloor(sink Sink, world World, view View) {
	forEachTile(view, 0, func(tile TilePos, at PixelPos) {
		if !world.InBounds(tile) {
			return
		}

		if world.IsFloor(tile) {
			sink.DrawFloorTile(target(world, tile, at))
		}
	})
}

// DrawTileContent renders the wall and entity pass. The walk runs MicroTileLen
// rows past the view because tall wall cels span micro rows
// (scrollrt.cpp:992-1045).
func DrawTileContent(sink Sink, world World, view View) {
	skip := false

	forEachTile(view, view.MicroTiles, func(tile TilePos, at PixelPos) {
		if !world.InBounds(tile) {
			return
		}

		skipNext := drawWallBehindFirst(sink, world, tile, at, view.ScreenWidth)
		if !skip {
			DrawDungeon(sink, world, tile, at)
		}

		skip = skipNext
	})
}

// DrawOOB fills every out-of-bounds tile of the walk with the dirt-pattern
// fallback, or a black tile in town (scrollrt.cpp:1088-1116).
func DrawOOB(sink Sink, world World, view View) {
	forEachTile(view, oobExtraRows, func(tile TilePos, at PixelPos) {
		if world.InBounds(tile) {
			return
		}

		if world.Town {
			sink.DrawBlackTile(at)

			return
		}

		drawDirtTile(sink, world, tile, at)
	})
}

// DrawDungeon renders one tile's contents in the upstream layer order: cell,
// missiles(pre), corpse, object(pre), item(pre), dead players, players,
// monsters, missiles(post), object(post), item(post), arch
// (scrollrt.cpp:782-953). The bean cites :781; the function opens at :782.
func DrawDungeon(sink Sink, world World, tile TilePos, at PixelPos) {
	light := world.Light(tile)
	sink.DrawCell(target(world, tile, at))

	if world.MissilePre {
		sink.DrawMissilePre(target(world, tile, at))
	}

	drawCorpse(sink, world, tile, at, light)
	drawObjectsBehindEntities(sink, world, tile, at, light)
	drawItemsBehindEntities(sink, world, tile, at)

	if world.DeadPlayer(tile) {
		sink.DrawDeadPlayer(target(world, tile, at))
	}

	drawPlayer(sink, world, tile, at)
	drawMonster(sink, world, tile, at)

	sink.DrawMissilePost(target(world, tile, at))
	drawObjectsBeforeEntities(sink, world, tile, at, light)
	drawItemsBeforeArch(sink, world, tile, at)
	drawArch(sink, world, tile, at)
}

// WorldToScreen is Displacement::worldToScreen: the -135° rotation scaled to a
// 64×32 tile, `{ (dy - dx) * 32, (dy + dx) * -16 }`
// (engine/displacement.hpp:151-155).
func WorldToScreen(dx, dy int) PixelPos {
	return PixelPos{X: (dy - dx) * (TileWidth / 2), Y: (dy + dx) * -(TileHeight / 2)}
}

// WalkOffset is GetOffsetForWalking in camera mode: the per-direction moving
// offset scaled by the fixed-point animation progress, negated because the
// camera moves against the walk (scrollrt.cpp:1575-1597).
func WalkOffset(dir Direction, progress int) PixelPos {
	step := movingOffsets[dir]
	return PixelPos{
		X: -(step.X * progress / animationFraction),
		Y: -(step.Y * progress / animationFraction),
	}
}

// Add returns the tile displaced by one step of dir, the Direction table of
// engine/displacement.hpp:250-272.
func (t TilePos) Add(dir Direction) TilePos {
	return TilePos{X: t.X + dir.Delta().X, Y: t.Y + dir.Delta().Y}
}

// String names the tile in test expectations.
func (t TilePos) String() string { return posString(t.X, t.Y) }

// Add returns the pixel displaced by d.
func (p PixelPos) Add(d PixelPos) PixelPos { return PixelPos{X: p.X + d.X, Y: p.Y + d.Y} }

// String names the pixel in test expectations.
func (p PixelPos) String() string { return posString(p.X, p.Y) }

// Delta is the world displacement of one step in dir
// (Displacement::fromDirection, engine/displacement.hpp:250-272).
func (d Direction) Delta() PixelPos {
	switch d {
	case South:
		return PixelPos{X: 1, Y: 1}
	case SouthWest:
		return PixelPos{X: 0, Y: 1}
	case West:
		return PixelPos{X: -1, Y: 1}
	case NorthWest:
		return PixelPos{X: -1, Y: 0}
	case North:
		return PixelPos{X: -1, Y: -1}
	case NorthEast:
		return PixelPos{X: 0, Y: -1}
	case East:
		return PixelPos{X: 1, Y: -1}
	case SouthEast:
		return PixelPos{X: 1, Y: 0}
	case NoDirection:
		return PixelPos{}
	}

	return PixelPos{}
}

// Opposite is the direction facing back (engine/direction.hpp:46-50).
func (d Direction) Opposite() Direction { return (d + 4) % 8 }

// forEachTile walks the isometric diamond: one row runs east (x+1, y-1) per
// column, then the row cursor moves half a tile down and, on odd rows, one tile
// east with one column fewer, on even rows one tile south with one column more
// (scrollrt.cpp:957-980, :1001-1042, :1090-1114). extraRows appends rows to the
// walk (DrawTileContent's micro rows, DrawOOB's five).
func forEachTile(view View, extraRows int, draw func(TilePos, PixelPos)) {
	tile, at := view.Start, view.Origin
	columns := view.Columns
	rows := view.Rows + extraRows

	for row := range rows {
		for range columns {
			draw(tile, at)

			tile = tile.Add(East)
			at.X += TileWidth
		}

		tile.X += West.Delta().X * columns
		tile.Y += West.Delta().Y * columns
		at.X -= columns * TileWidth
		at.Y += TileHeight / 2

		if row&1 != 0 {
			tile.X++
			columns--
			at.X += TileWidth / 2
		} else {
			tile.Y++
			columns++
			at.X -= TileWidth / 2
		}
	}
}

// drawWallBehindFirst draws the objects behind a wall before the wall itself so
// a sprite walking between the two tiles cannot poke through, and reports
// whether the next tile of the row is already drawn (scrollrt.cpp:1009-1024).
func drawWallBehindFirst(sink Sink, world World, tile TilePos, at PixelPos, screenWidth int) bool {
	if tile.X+1 >= MaxDunX || tile.Y-1 < 0 || at.X+TileWidth > screenWidth {
		return false
	}

	if !wallAlignedOnXAxis(world, tile) {
		return false
	}

	behind, behindAt := tile.Add(East), PixelPos{X: at.X + TileWidth, Y: at.Y}
	if !world.NotSolid(TilePos{X: tile.X + 1, Y: tile.Y - 1}) ||
		!world.NotSolid(TilePos{X: tile.X, Y: tile.Y - 1}) {
		return false
	}

	DrawDungeon(sink, world, behind, behindAt)

	return true
}

// drawCorpse draws the corpse on the tile, which unlit tiles skip
// (scrollrt.cpp:788-811).
func drawCorpse(sink Sink, world World, tile TilePos, at PixelPos, light uint8) {
	if light >= LightsMax || world.Corpse(tile) == 0 {
		return
	}

	sink.DrawCorpse(target(world, tile, at))
}

// drawObjectsBehindEntities draws the objects flagged `_oPreFlag`
// (scrollrt.cpp:813-819).
func drawObjectsBehindEntities(sink Sink, world World, tile TilePos, at PixelPos, light uint8) {
	object, found := findObject(world, tile, light)
	if !found || !object.Pre {
		return
	}

	sink.DrawObjectPre(target(world, tile, at))
}

// drawItemsBehindEntities draws the items without `_iPostDraw`, over the
// objects but under the entities (scrollrt.cpp:813-822).
func drawItemsBehindEntities(sink Sink, world World, tile TilePos, at PixelPos) {
	item := world.Item(tile)
	if item <= 0 || world.ItemPostDraw(item) {
		return
	}

	sink.DrawItemPre(target(world, tile, at))
}

// drawPlayer draws the player on the tile. A southwards walk and an eastwards
// sideways walk register on the destination tile with a negative dPlayer id so
// the tile order paints them from the tile they are leaving
// (scrollrt.cpp:827-862).
func drawPlayer(sink Sink, world World, tile TilePos, at PixelPos) {
	occupant, found := world.Player(tile)
	if !found {
		return
	}

	if occupant.Encoded != playerGridID(occupant) {
		return
	}

	drawn, drawnAt := playerOrigin(occupant, tile, at)
	sink.DrawPlayer(target(world, drawn, drawnAt), occupant.ID)
}

// drawMonster draws the monster standing on the tile, offset back to the tile
// it is leaving while it walks (scrollrt.cpp:864-899, monster.h:470-482).
func drawMonster(sink Sink, world World, tile TilePos, at PixelPos) {
	occupant, found := world.Monster(tile)
	if !found {
		return
	}

	drawn, drawnAt := monsterOrigin(occupant, tile, at)
	sink.DrawMonster(target(world, drawn, drawnAt), occupant.ID)
}

// drawObjectsBeforeEntities draws the objects without `_oPreFlag`, after the
// entities on the tile (scrollrt.cpp:903-905).
func drawObjectsBeforeEntities(sink Sink, world World, tile TilePos, at PixelPos, light uint8) {
	object, found := findObject(world, tile, light)
	if !found || object.Pre {
		return
	}

	sink.DrawObjectPost(target(world, tile, at))
}

// drawItemsBeforeArch draws the items whose flag puts them over the entities
// (scrollrt.cpp:906-908).
func drawItemsBeforeArch(sink Sink, world World, tile TilePos, at PixelPos) {
	item := world.Item(tile)
	if item <= 0 || !world.ItemPostDraw(item) {
		return
	}

	sink.DrawItemPost(target(world, tile, at))
}

// drawArch draws the tile's arch cel, blended for a transparent region and, in
// town, deferred to the row below from the north-west neighbour
// (scrollrt.cpp:910-952).
func drawArch(sink Sink, world World, tile TilePos, at PixelPos) {
	if !world.Town {
		drawDungeonArch(sink, world, tile, at)

		return
	}

	drawTownLeaves(sink, world, tile, at)
}

// drawDungeonArch draws the arch cel of the tile itself, blended when the tile
// is a transparency window (scrollrt.cpp:910-943).
func drawDungeonArch(sink Sink, world World, tile TilePos, at PixelPos) {
	arch := world.Special(tile) - 1
	if arch < 0 {
		return
	}

	if world.Transparent(tile) {
		sink.DrawArchBlended(target(world, tile, at), arch)

		return
	}

	sink.DrawArch(target(world, tile, at), arch)
}

// drawTownLeaves draws the town tree leaves of the north-west neighbour one row
// up, so they cover a player entering or leaving the tile
// (scrollrt.cpp:944-952).
func drawTownLeaves(sink Sink, world World, tile TilePos, at PixelPos) {
	if tile.X <= 0 || tile.Y <= 0 || at.Y <= TileHeight {
		return
	}

	neighbour := TilePos{X: tile.X - 1, Y: tile.Y - 1}
	arch := world.Special(neighbour) - 1
	if arch < 0 {
		return
	}

	raised := PixelPos{X: at.X, Y: at.Y - TileHeight}
	sink.DrawArchDirect(target(world, neighbour, raised), arch)
}

// drawDirtTile draws the dirt fallback for an out-of-bounds tile: the 2×2 dirt
// pattern cell matching the tile's grid parity, or a black tile when the cell is
// empty (scrollrt.cpp:1047-1078).
func drawDirtTile(sink Sink, world World, tile TilePos, at PixelPos) {
	sample := dirtSample(tile)
	if !world.InBounds(sample) || !world.IsFloor(sample) || world.Piece(sample) == 0 {
		sink.DrawBlackTile(at)

		return
	}

	sink.DrawCell(target(world, sample, at))
}

// dirtSample is the 2×2 dirt-pattern cell of an out-of-bounds tile: the pattern
// base plus the tile's parity (scrollrt.cpp:1050-1058).
func dirtSample(tile TilePos) TilePos {
	return TilePos{X: tileParity(tile.X), Y: tileParity(tile.Y)}
}

// tileParity is the 2×2 dirt-pattern index of one grid coordinate
// (scrollrt.cpp:1043-1045).
func tileParity(coordinate int) int {
	if coordinate&1 == 0 {
		return 0
	}

	return 1
}

// playerGridID is the dPlayer value the tile must hold for the player to be
// drawn here: negative while walking southwards or eastwards
// (scrollrt.cpp:829-836).
func playerGridID(occupant Occupant) int8 {
	if occupant.Mode == ModeMoveSouthwards ||
		(occupant.Mode == ModeMoveSideways && occupant.Dir == East) {
		return int8(-(occupant.ID + 1))
	}

	return int8(occupant.ID + 1)
}

// playerOrigin moves a walking player's draw back to the tile it is leaving
// (scrollrt.cpp:840-859).
func playerOrigin(occupant Occupant, tile TilePos, at PixelPos) (TilePos, PixelPos) {
	if occupant.Mode == ModeMoveSouthwards {
		return tile.Add(occupant.Dir.Opposite()), at.Add(southwardsOffset(occupant.Dir))
	}

	if occupant.Mode == ModeMoveSideways && occupant.Dir == East {
		return tile.Add(occupant.Dir.Opposite()), at.Add(PixelPos{X: -TileWidth})
	}

	return tile, at
}

// southwardsOffset is the half-tile back-step of a southwards walk
// (scrollrt.cpp:843-854).
func southwardsOffset(dir Direction) PixelPos {
	switch dir {
	case SouthWest:
		return PixelPos{X: TileWidth / 2, Y: -TileHeight / 2}
	case South:
		return PixelPos{X: 0, Y: -TileHeight}
	case SouthEast:
		return PixelPos{X: -TileWidth / 2, Y: -TileHeight / 2}
	default:
		return PixelPos{}
	}
}

// monsterOrigin moves a walking monster's draw back to the tile it is leaving
// (monster.h:470-482).
func monsterOrigin(occupant Occupant, tile TilePos, at PixelPos) (TilePos, PixelPos) {
	if occupant.Mode == ModeMoveSouthwards {
		return tile.Add(occupant.Dir.Opposite()), at.Add(southwardsOffset(occupant.Dir))
	}

	if occupant.Mode == ModeMoveSideways && occupant.Dir == East {
		return tile.Add(occupant.Dir.Opposite()), at.Add(PixelPos{X: -TileWidth})
	}

	return tile, at
}

// wallAlignedOnXAxis reports whether the tile is part of a wall running along
// the x-axis (scrollrt.cpp:1014).
func wallAlignedOnXAxis(world World, tile TilePos) bool {
	if !world.IsWall(tile) {
		return false
	}

	return world.IsWall(TilePos{X: tile.X + 1, Y: tile.Y}) ||
		(tile.X > 0 && world.IsWall(TilePos{X: tile.X - 1, Y: tile.Y}))
}

// findObject is FindObjectAtPosition gated on the light index
// (scrollrt.cpp:814-816).
func findObject(world World, tile TilePos, light uint8) (ObjectRef, bool) {
	if light >= LightsMax {
		return ObjectRef{}, false
	}

	return world.Object(tile)
}

// target builds the draw target of one layer draw.
func target(world World, tile TilePos, at PixelPos) DrawTarget {
	return DrawTarget{Tile: tile, At: at, Light: world.Light(tile)}
}

// posString formats a coordinate pair for test expectations.
func posString(x, y int) string { return fmt.Sprintf("(%d, %d)", x, y) }
