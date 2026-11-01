// Package missiles: the wall control missiles (Fire Wall, Flame Wave).
package missiles

import "github.com/bmcszk/devilutionGo/internal/game/world"

// Left ports Left (engine/direction.hpp:24-28): a counter-clockwise turn.
func Left(dir Direction) Direction { return Direction((int(dir) + 7) % 8) }

// Right ports Right (engine/direction.hpp:31-35): a clockwise turn.
func Right(dir Direction) Direction { return Direction((int(dir) + 1) % 8) }

// CanPlaceWall ports CanPlaceWall (missiles.cpp:736-745): inside the dungeon
// and not flagged BlockMissile.
func CanPlaceWall(w *HitWorld, position TilePosition) bool {
	if !world.InDungeonBounds(position.X, position.Y) { // cpp:738-739
		return false
	}

	return !w.Grid.TileHasAny(position.X, position.Y, world.TileBlockMissile) // cpp:744
}

// PlaceWall ports PlaceWall (missiles.cpp:747-750): a fresh wall segment from
// position toward its facing, cast by both sides, handed back so the grower can
// nudge it.
func PlaceWall(w *HitWorld, id int, spawnID MissileID, position TilePosition, direction Direction,
	spellLevel, damage int,
) *Missile {
	dx, dy := directionDeltaOf(direction)

	dst := TilePosition{X: position.X + dx, Y: position.Y + dy}

	return w.ProcessCallbacks.PlaceWallMissile(position, dst, direction, spawnID, int(TargetBoth), id,
		damage, spellLevel)
}

// TryGrowWall ports TryGrowWall (missiles.cpp:752-782): a cardinal grow
// direction lays the segment in the gap beside the wall — pushed 30 pixels to
// the far edge of the tile, not drawn, facing sideways — and then plants the
// segment itself; a blocked tile ends the arm.
func TryGrowWall(w *HitWorld, id int, spawnID MissileID, position TilePosition, growDirection,
	direction Direction, spellLevel, damage int,
) bool {
	if isCardinal(growDirection) {
		gyx, gyy := directionDeltaOf(growDirection)
		rx, ry := directionDeltaOf(Right(growDirection))
		gapPos := TilePosition{X: position.X + rx - gyx, Y: position.Y + ry - gyy}

		if CanPlaceWall(w, gapPos) {
			wall := PlaceWall(w, id, spawnID, gapPos, direction, spellLevel, damage)
			nudgeWall(w, wall, growDirection)
		}
	}

	if !CanPlaceWall(w, position) { // cpp:779-780
		return false
	}

	PlaceWall(w, id, spawnID, position, direction, spellLevel, damage)

	return true
}

// MoveWallToNextPosition ports MoveWallToNextPosition (missiles.cpp:783-790):
// one tile along the grow direction, or nothing outside the dungeon.
func MoveWallToNextPosition(position TilePosition, growDirection Direction) (TilePosition, bool) {
	dx, dy := directionDeltaOf(growDirection)
	next := TilePosition{X: position.X + dx, Y: position.Y + dy}

	return next, world.InDungeonBounds(next.X, next.Y)
}

// nudgeWall ports the segment adjustment inside TryGrowWall (cpp:766-773).
func nudgeWall(w *HitWorld, wall *Missile, growDirection Direction) {
	if wall == nil { // PlaceWall can fail upstream too; the assert below is guarded
		return
	}

	dx, dy := directionDeltaOf(Right(Right(growDirection)))
	travelled := WorldToNormalScreen(PositionDisplacement{DeltaX: int32(dx), DeltaY: int32(dy)})
	wall.Position.Traveled.DeltaX += travelled.DeltaX * 30 // cpp:768: to the tile edge, not over
	wall.Position.Traveled.DeltaY += travelled.DeltaY * 30
	wall.DrawFlag = false
	wall.Var3 = int(Left(Left(growDirection))) + 1
	UpdateMissilePos(w, wall)
}

// isCardinal reports whether the grow direction is one of the four the wall
// grower handles (cpp:755-758).
func isCardinal(dir Direction) bool {
	return dir == South || dir == West || dir == North || dir == East
}
