// Package missiles holds the missile state, geometry, and per-tick logic.
package missiles

import "math"

// worldToScreen converts a world tile delta to isometric screen pixels.
func worldToScreen(d positionDisplacement) positionDisplacement {
	return positionDisplacement{
		DeltaX: (d.DeltaY - d.DeltaX) * 32,
		DeltaY: (d.DeltaY + d.DeltaX) * -16,
	}
}

// screenToWorld converts an isometric screen delta to world tiles.
func screenToWorld(d positionDisplacement) positionDisplacement {
	return positionDisplacement{
		DeltaX: (2*d.DeltaY + d.DeltaX) / -64,
		DeltaY: (2*d.DeltaY - d.DeltaX) / -64,
	}
}

// screenToMissile converts an isometric screen delta to a missile position delta.
//
// The ±32 nudge on each numerator rounds the exact-halving case away from zero.
func screenToMissile(d positionDisplacement) positionDisplacement {
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
	return positionDisplacement{
		DeltaX: (xNum + xOff) / 64,
		DeltaY: (yNum + yOff) / 64,
	}
}

// screenToLight converts an isometric screen delta to the coarser light grid.
func screenToLight(d positionDisplacement) positionDisplacement {
	return positionDisplacement{
		DeltaX: (2*d.DeltaY + d.DeltaX) / 8,
		DeltaY: (2*d.DeltaY - d.DeltaX) / 8,
	}
}

// worldToNormalScreen converts a world tile delta to a normalized screen
// vector with the vertical component halved.
func worldToNormalScreen(d positionDisplacement) positionDisplacement {
	rotated := positionDisplacement{
		DeltaX: d.DeltaY - d.DeltaX,
		DeltaY: -(d.DeltaY + d.DeltaX),
	}
	n := normalized(rotated)
	return positionDisplacement{
		DeltaX: n.DeltaX,
		DeltaY: n.DeltaY / 2,
	}
}

// normalized returns d scaled to unit length in 16.16 fixed point.
func normalized(d positionDisplacement) positionDisplacement {
	mag := float32(math.Sqrt(float64(d.DeltaX*d.DeltaX + d.DeltaY*d.DeltaY)))
	return positionDisplacement{
		DeltaX: int32(float32(d.DeltaX<<16) / mag),
		DeltaY: int32(float32(d.DeltaY<<16) / mag),
	}
}

// manhattanDistance returns the tile distance |dx|+|dy|.
func manhattanDistance(from, to tilePosition) int {
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
