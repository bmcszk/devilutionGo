// Package missiles: the MissilePosition block and the 16.16 fixed point
// motion arithmetic (Source/missiles.h:24-49, missiles.cpp:213-215,
// missiles.cpp:2810-2830; devilGo-mwlm slice 1).
package missiles

// fixed point shift: upstream "Pixels traveled as a numerator of 65,536"
// (missiles.h:30) — pixel velocities are stored scaled by 2^16.
const fracBits = 16

// TilePosition is a world tile coordinate (PointOf<WorldTileCoord>,
// Source/engine/world_tile.hpp:11-12).
type TilePosition struct {
	X, Y int
}

// PositionDisplacement is a signed 16.16 fixed point delta
// (Displacement, Source/engine/displacement.hpp:20 — DeltaT int).
type PositionDisplacement struct {
	DeltaX int32
	DeltaY int32
}

// Position is MissilePosition (Source/missiles.h:24-49): sprite pixel
// offset, pixel velocity, 16.16 traveled vector, and the tile/start
// coordinates.
type Position struct {
	// Sprite's pixel offset from tile (missiles.h:27 offset).
	Offset PositionDisplacement
	// Pixel velocity while moving (missiles.h:29 velocity).
	Velocity PositionDisplacement
	// Pixels traveled as a numerator of 65,536 (missiles.h:31 traveled).
	Traveled PositionDisplacement
	// Current tile (missiles.h:33 tile).
	Tile TilePosition
	// Start position (missiles.h:35 start).
	Start TilePosition
	// Renderer tile (missiles.h:40 tileForRendering) — written by the
	// renderer (scrollrt.cpp:148-158), read by StopMissile.
	TileForRendering TilePosition
	// Renderer offset (missiles.h:44 offsetForRendering).
	OffsetForRendering PositionDisplacement
}

// traveledForTile is the AddMissile seed (missiles.cpp:2801 region):
// a missile spawning on tile (x, y) starts with traveled = tile << 16.
func traveledForTile(x, y int) PositionDisplacement {
	return PositionDisplacement{
		DeltaX: int32(x) << fracBits,
		DeltaY: int32(y) << fracBits,
	}
}

// SetTraveledFromTile seeds position.traveled = tile << 16 (AddMissile,
// Source/missiles.cpp:2810-2830 semantics).
func (m *Missile) SetTraveledFromTile(x, y int) {
	m.Position.Traveled = traveledForTile(x, y)
	m.Position.Tile = TilePosition{X: x, Y: y}
	m.Position.Start = m.Position.Tile
}

// MidPosition ports the pixelsTravelled extraction
// (UpdateMissilePos, Source/missiles.cpp:213-215): traveled >> 16.
// int32 arithmetic shift — floors toward negative infinity like C++.
func (m *Missile) MidPosition() PositionDisplacement {
	return PositionDisplacement{
		DeltaX: m.Position.Traveled.DeltaX >> fracBits,
		DeltaY: m.Position.Traveled.DeltaY >> fracBits,
	}
}

// Step ports the MoveMissile travel line (Source/missiles.cpp:555):
// position.traveled += position.velocity, with the velocity scaled by
// 2^16 (fixed16NormalVector*velocityInPixels is 16.16 scaled,
// displacement.hpp:190-206 normalized(); scaleVelocity,
// displacement.hpp:219 `<< 16u`). UpdateMissilePos then re-derives the
// tile: traveled >> 16 (missiles.cpp:215). Collision/tile-walking is
// out of scope for slice 1.
func Step(m *Missile, velocity PositionDisplacement) {
	m.Position.Traveled.DeltaX += velocity.DeltaX << fracBits
	m.Position.Traveled.DeltaY += velocity.DeltaY << fracBits
	m.Position.Tile.X = int(m.Position.Traveled.DeltaX >> fracBits)
	m.Position.Tile.Y = int(m.Position.Traveled.DeltaY >> fracBits)
}

// PathDistance ports PointOf::WalkingDistance
// (Source/engine/point.hpp:154-159): max(|dx|, |dy|) — the possible tile
// visit count MoveMissile uses for axis-aligned travel
// (missiles.cpp:560-563).
func PathDistance(from, to TilePosition) int {
	dx := from.X - to.X
	dy := from.Y - to.Y
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	if dx >= dy {
		return dx
	}
	return dy
}
