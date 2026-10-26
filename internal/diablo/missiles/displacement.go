// Package missiles holds the missile state, geometry, and per-tick logic.
package missiles

import "math"

// WorldToScreen converts a world tile delta to isometric screen pixels.
func WorldToScreen(d PositionDisplacement) PositionDisplacement {
	return PositionDisplacement{
		DeltaX: (d.DeltaY - d.DeltaX) * 32,
		DeltaY: (d.DeltaY + d.DeltaX) * -16,
	}
}

// ScreenToWorld converts an isometric screen delta to world tiles.
func ScreenToWorld(d PositionDisplacement) PositionDisplacement {
	return PositionDisplacement{
		DeltaX: (2*d.DeltaY + d.DeltaX) / -64,
		DeltaY: (2*d.DeltaY - d.DeltaX) / -64,
	}
}

// ScreenToMissile converts an isometric screen delta to a missile position delta.
//
// The ±32 nudge on each numerator rounds the exact-halving case away from zero.
func ScreenToMissile(d PositionDisplacement) PositionDisplacement {
	xNum := 2*d.DeltaY + d.DeltaX
	yNum := 2*d.DeltaY - d.DeltaX

	xOff := int32(-32)
	if xNum >= 0 {
		xOff = 32
	}

	yOff := int32(-32)
	if yNum >= 0 {
		yOff = 32
	}

	return PositionDisplacement{
		DeltaX: (xNum + xOff) / 64,
		DeltaY: (yNum + yOff) / 64,
	}
}

// ScreenToLight converts an isometric screen delta to the coarser light grid.
func ScreenToLight(d PositionDisplacement) PositionDisplacement {
	return PositionDisplacement{
		DeltaX: (2*d.DeltaY + d.DeltaX) / 8,
		DeltaY: (2*d.DeltaY - d.DeltaX) / 8,
	}
}

// WorldToNormalScreen converts a world tile delta to a normalized screen
// vector with the vertical component halved.
func WorldToNormalScreen(d PositionDisplacement) PositionDisplacement {
	rotated := PositionDisplacement{
		DeltaX: d.DeltaY - d.DeltaX,
		DeltaY: -(d.DeltaY + d.DeltaX),
	}
	n := normalized(rotated)

	return PositionDisplacement{
		DeltaX: n.DeltaX,
		DeltaY: n.DeltaY / 2,
	}
}

// ManhattanDistance returns the tile distance |dx|+|dy|.
func ManhattanDistance(from, to TilePosition) int {
	dx := from.X - to.X
	dy := from.Y - to.Y

	if dx < 0 {
		dx = -dx
	}

	if dy < 0 {
		dy = -dy
	}

	return dx + dy
}

// normalized returns d scaled to unit length in 16.16 fixed point.
func normalized(d PositionDisplacement) PositionDisplacement {
	mag := float32(math.Sqrt(float64(d.DeltaX*d.DeltaX + d.DeltaY*d.DeltaY)))

	return PositionDisplacement{
		DeltaX: int32(float32(d.DeltaX<<16) / mag),
		DeltaY: int32(float32(d.DeltaY<<16) / mag),
	}
}
