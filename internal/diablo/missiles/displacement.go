// Package missiles: the Displacement coordinate conversions and vector
// helpers of Source/engine/displacement.hpp plus the companion point
// metrics of Source/engine/point.hpp — 1:1 ports for the devilGo-mwlm
// slice (displacement.go).
package missiles

import "math"

// WorldToScreen ports Displacement::worldToScreen
// (Source/engine/displacement.hpp:151): {(dy - dx) * 32, (dy + dx) * -16} —
// world tiles to isometric screen pixels.
func WorldToScreen(d PositionDisplacement) PositionDisplacement {
	return PositionDisplacement{
		DeltaX: (d.DeltaY - d.DeltaX) * 32,
		DeltaY: (d.DeltaY + d.DeltaX) * -16,
	}
}

// ScreenToWorld ports Displacement::screenToWorld
// (Source/engine/displacement.hpp:160): {(2*dy + dx) / -64, (2*dy - dx) / -64}.
// C++ int division truncates toward zero; Go's signed `/` matches exactly.
func ScreenToWorld(d PositionDisplacement) PositionDisplacement {
	return PositionDisplacement{
		DeltaX: (2*d.DeltaY + d.DeltaX) / -64,
		DeltaY: (2*d.DeltaY - d.DeltaX) / -64,
	}
}

// ScreenToMissile ports Displacement::screenToMissile
// (Source/engine/displacement.hpp:174-183): xNum = 2*dy + dx;
// yNum = 2*dy - dx; each numerator is nudged ±32 toward +infinity on the
// offset side before dividing by 64 (the exact-halving case rounds away
// from zero — upstream's asymmetric rounding, replicated verbatim).
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

// ScreenToLight ports Displacement::screenToLight
// (Source/engine/displacement.hpp:184): {(2*dy + dx) / 8, (2*dy - dx) / 8} —
// the coarser light grid (positive divisor, unlike screenToWorld's -64).
func ScreenToLight(d PositionDisplacement) PositionDisplacement {
	return PositionDisplacement{
		DeltaX: (2*d.DeltaY + d.DeltaX) / 8,
		DeltaY: (2*d.DeltaY - d.DeltaX) / 8,
	}
}

// WorldToNormalScreen ports Displacement::worldToNormalScreen
// (Source/engine/displacement.hpp:195-206): rotated = {dy - dx, -(dy + dx)};
// n = normalized(rotated); return {n.dx, n.dy / 2}.
func WorldToNormalScreen(d PositionDisplacement) PositionDisplacement {
	rotated := PositionDisplacement{
		DeltaX: d.DeltaY - d.DeltaX,
		DeltaY: -(d.DeltaY + d.DeltaX),
	}
	n := Normalized(rotated)
	return PositionDisplacement{
		DeltaX: n.DeltaX,
		DeltaY: n.DeltaY / 2,
	}
}

// Normalized ports Displacement::normalized
// (Source/engine/displacement.hpp:363-370): mag = float32(sqrt(float64(
// dx*dx + dy*dy))); out = {int(float32(dx << 16) / mag),
// int(float32(dy << 16) / mag)} — shift by 16 first, divide in float32,
// int() truncates toward zero.
func Normalized(d PositionDisplacement) PositionDisplacement {
	mag := float32(math.Sqrt(float64(d.DeltaX*d.DeltaX + d.DeltaY*d.DeltaY)))
	return PositionDisplacement{
		DeltaX: int32(float32(d.DeltaX<<16) / mag),
		DeltaY: int32(float32(d.DeltaY<<16) / mag),
	}
}

// ManhattanDistance ports PointOf::ManhattanDistance
// (Source/engine/point.hpp:147-152): |dx| + |dy| — the tile pathing
// companion of PathDistance (WalkingDistance, point.hpp:154-159).
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
