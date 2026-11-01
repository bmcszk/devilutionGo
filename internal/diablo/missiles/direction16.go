// Package missiles: the 16-step facing used by missile frame groups.
package missiles

// Direction16 mirrors enum class Direction16 (tables/misdat.h:147-164),
// ordinal order preserved: 22.8125° per step starting at South.
const (
	Direction16South Direction16 = iota
	Direction16SouthSouthWest
	Direction16SouthWest
	Direction16WestSouthWest
	Direction16West
	Direction16WestNorthWest
	Direction16NorthWest
	Direction16NorthNorthWest
	Direction16North
	Direction16NorthNorthEast
	Direction16NorthEast
	Direction16EastNorthEast
	Direction16East
	Direction16EastSouthEast
	Direction16SouthEast
	Direction16SouthSouthEast
)

// Direction16 is a 22.8125° facing (upstream enum class Direction16,
// tables/misdat.h:147).
type Direction16 uint8

// direction16Mirror carries the three upstream mirror flags
// (missiles.cpp:992-999).
type direction16Mirror struct {
	flipX      bool
	flipY      bool
	flipMedian bool
}

// GetDirection16 ports GetDirection16 (missiles.cpp:987-1020): the
// from→to facing at 16-step resolution. Upstream reduces the offset to the
// (absX <= absY) octant, picks South / South_SouthWest / SouthWest from the
// tan(11.25°) and tan(33.75°) approximations, then mirrors back with
// Direction16Flip.
func GetDirection16(from, to TilePosition) Direction16 {
	offsetX := to.X - from.X
	offsetY := to.Y - from.Y
	absX, absY := absoluteOffsets(offsetX, offsetY)
	mirror := offsetMirror(offsetX, offsetY, absX, absY)

	if mirror.flipMedian { // cpp:996-999
		absX, absY = absY, absX
	}

	return mirrorDirection16(direction16OfOctant(absX, absY), mirror)
}

// absoluteOffsets ports abs(offset) (missiles.cpp:990).
func absoluteOffsets(offsetX, offsetY int) (absX, absY int) {
	absX, absY = offsetX, offsetY

	if absX < 0 {
		absX = -absX
	}

	if absY < 0 {
		absY = -absY
	}

	return absX, absY
}

// offsetMirror derives the upstream mirror flags: the sign test of the
// abs() result (cpp:992-993) plus the x-dominant swap flag (cpp:996-999).
func offsetMirror(offsetX, offsetY, absX, absY int) direction16Mirror {
	return direction16Mirror{
		flipY:      offsetX != absX,
		flipX:      offsetY != absY,
		flipMedian: absX > absY,
	}
}

// direction16OfOctant picks the facing inside the (absX <= absY) octant
// (missiles.cpp:1001-1007).
func direction16OfOctant(absX, absY int) Direction16 {
	if 3*absX > absY*2 { // cpp:1002 tan(33.75°)
		return Direction16South
	}

	if 5*absX < absY { // cpp:1003 tan(11.25°)
		return Direction16SouthWest
	}

	return Direction16SouthSouthWest
}

// mirrorDirection16 mirrors the octant facing back through the flip chain
// (missiles.cpp:1009-1019).
func mirrorDirection16(ret Direction16, mirror direction16Mirror) Direction16 {
	medianPivot := Direction16South

	if mirror.flipY { // cpp:1010-1013
		ret = direction16Flip(ret, Direction16SouthWest)
		medianPivot = direction16Flip(medianPivot, Direction16SouthWest)
	}

	if mirror.flipX { // cpp:1014-1017
		ret = direction16Flip(ret, Direction16SouthEast)
		medianPivot = direction16Flip(medianPivot, Direction16SouthEast)
	}

	if mirror.flipMedian { // cpp:1018-1019
		ret = direction16Flip(ret, medianPivot)
	}

	return ret
}

// direction16Flip mirrors x around pivot on the 16-step ring
// (missiles.cpp:166-171).
func direction16Flip(x, pivot Direction16) Direction16 {
	return Direction16((2*int(pivot) + 16 - int(x)) % 16)
}
