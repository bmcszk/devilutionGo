// Package missiles: the missile position and 16.16 fixed point motion.
package missiles

// fracBits is the 16.16 fixed point shift: a pixel is stored as a numerator of 2^16.
const fracBits = 16

// tilePosition is a world tile coordinate.
type tilePosition struct {
	X, Y int
}

// positionDisplacement is a signed two-axis delta. Offsets and velocities are
// in pixels; traveled vectors are in 16.16 fixed point.
type positionDisplacement struct {
	DeltaX int32
	DeltaY int32
}

// Position is the missile position: pixel offset and velocity, a 16.16
// traveled vector, and the tile coordinates.
type Position struct {
	// Offset is the sprite's pixel offset from the tile.
	Offset positionDisplacement
	// Velocity is the pixel velocity per tick.
	Velocity positionDisplacement
	// Traveled is the 16.16 traveled vector: pixels as a numerator of 2^16.
	Traveled positionDisplacement
	Tile     tilePosition
	Start    tilePosition
	// TileForRendering is written by the renderer and read when the missile stops.
	TileForRendering tilePosition
	// OffsetForRendering is the renderer-written pixel offset.
	OffsetForRendering positionDisplacement
}

func traveledForTile(x, y int) positionDisplacement {
	return positionDisplacement{
		DeltaX: int32(x) << fracBits,
		DeltaY: int32(y) << fracBits,
	}
}

// SetTraveledFromTile seeds the traveled vector from a tile and sets the
// current and start tile.
func (m *Missile) SetTraveledFromTile(x, y int) {
	m.Position.Traveled = traveledForTile(x, y)
	m.Position.Tile = tilePosition{X: x, Y: y}
	m.Position.Start = m.Position.Tile
}

// MidPosition returns the traveled position in whole pixels; the arithmetic
// shift floors toward negative infinity.
func (m *Missile) MidPosition() positionDisplacement {
	return positionDisplacement{
		DeltaX: m.Position.Traveled.DeltaX >> fracBits,
		DeltaY: m.Position.Traveled.DeltaY >> fracBits,
	}
}

// Step moves the missile by velocity pixels and re-derives the current tile
// from the traveled vector.
func Step(m *Missile, velocity positionDisplacement) {
	m.Position.Traveled.DeltaX += velocity.DeltaX << fracBits
	m.Position.Traveled.DeltaY += velocity.DeltaY << fracBits
	m.Position.Tile.X = int(m.Position.Traveled.DeltaX >> fracBits)
	m.Position.Tile.Y = int(m.Position.Traveled.DeltaY >> fracBits)
}

// pathDistance returns the tile distance counting the diagonal as one step.
func pathDistance(from, to tilePosition) int {
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
