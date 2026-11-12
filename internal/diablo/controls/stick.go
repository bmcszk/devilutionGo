// Package controls ports the fork gamepad stick decision core
// (Source/controls/plrctrls.cpp, controller_motion.{h,cpp} @ ca91d778f):
// StickPush classification, 8-way sectoring, the FaceDir table, and the
// Movement/TurnToDir/WalkInDir/IsPathBlocked pipeline over a narrow
// PlayerView + func hooks. Ebiten wiring (axis reading, deadzone scaling)
// lands in the wiring lane; this package takes raw stick floats and emits
// commands through hooks only (no ebiten import, no game-kernel coupling).
package controls

// StickDirectionThreshold (controller_motion.h:29): minimum scaled stick
// magnitude to walk; below it, the character only turns in place.
const StickDirectionThreshold = 0.6

// StickDirectionThresholdSquared (controller_motion.h:30) avoids the sqrt
// in the push classification.
const StickDirectionThresholdSquared = StickDirectionThreshold * StickDirectionThreshold

// DiagonalCutoff (controller_motion.cpp:229): tan(22.5°) — the 8-way
// sectoring boundary between cardinals and diagonals.
const DiagonalCutoff = 0.41421356

// WalkNone (player.h:58 WALK_NONE): the idle walkpath[0] sentinel.
const WalkNone int8 = -1

// DestActionNone (player.h:128 ACTION_NONE): the idle destAction sentinel.
const DestActionNone int8 = -1

// StickPush enum values (controller_motion.h:32-37).
const (
	// PushNone (StickPush::None): inside the deadzone. The deadzone itself is
	// applied upstream by ScaleJoystickAxes before the values reach this
	// port — a zero stick here means "no push"; this package takes RAW
	// floats and deliberately does NOT re-apply any deadzone (the wiring
	// lane owns scaling).
	PushNone StickPush = iota
	// PushWeak (StickPush::Weak): deflected, but below the walk threshold.
	PushWeak
	// PushFirm (StickPush::Firm): at or above the walk threshold.
	PushFirm
)

// AxisDirectionX ordinals (axis_direction.h:7-11) — the FaceDir table
// indexes on these.
const (
	AxisXNone  AxisDirectionX = 0 // AxisDirectionX_NONE
	AxisXLeft  AxisDirectionX = 1 // AxisDirectionX_LEFT
	AxisXRight AxisDirectionX = 2 // AxisDirectionX_RIGHT
)

// AxisDirectionY ordinals (axis_direction.h:12-16). Up = screen-north
// (decreasing tile Y).
const (
	AxisYNone AxisDirectionY = 0 // AxisDirectionY_NONE
	AxisYUp   AxisDirectionY = 1 // AxisDirectionY_UP
	AxisYDown AxisDirectionY = 2 // AxisDirectionY_DOWN
)

// Direction ordinals (engine/direction.hpp:10-20) — Left/Right rotation
// arithmetic depends on this order.
const (
	South Direction = iota // direction.hpp:11
	SouthWest
	West
	NorthWest
	North
	NorthEast
	East
	SouthEast
	// NoDirection (direction.hpp:19): "no facing"; TurnToDir's no-op guard.
	NoDirection
)

// PlrMode values the movement pipeline reads (player.h:109 PLR_MODE). The
// Go entity enum keeps the upstream names on entities' Player; this local
// mirror keeps the controls package import-free. ModeWalk collapses the
// upstream PM_WALK_NORTHWARDS/SOUTHWARDS/SIDEWAYS trio (isWalking,
// player.cpp:1709-1713); ModeAttack collapses PM_ATTACK/PM_RATTACK (both
// gate on _pAFNum identically).
const (
	ModeStand  PlrMode = iota // player.h:111 PM_STAND
	ModeWalk                  // player.h:112-114 PM_WALK_*
	ModeAttack                // player.h PM_ATTACK / PM_RATTACK
	ModeSpell                 // player.h PM_SPELL (PM_CAST gates on _pSFNum too)
)

// leftRotation (direction.hpp:26 Left): a left turn from the facing.
var leftRotation = [8]Direction{SouthEast, South, SouthWest, West, NorthWest, North, NorthEast, East}

// rightRotation (direction.hpp:33 Right): a right turn from the facing.
var rightRotation = [8]Direction{SouthWest, West, NorthWest, North, NorthEast, East, SouthEast, South}

// deltaByDirection ports DisplacementOf::fromDirection
// (engine/displacement.hpp:247-275): the tile step per Direction.
var deltaByDirection = [9][2]int{
	{1, 1},   // South
	{0, 1},   // SouthWest
	{-1, 1},  // West
	{-1, 0},  // NorthWest
	{-1, -1}, // North
	{0, -1},  // NorthEast
	{1, -1},  // East
	{1, 0},   // SouthEast
	{0, 0},   // NoDirection
}

// FaceDir ports FaceDir[3][3] (plrctrls.cpp:100-105), indexed [x][y] with
// the AxisDirectionX/Y ordinals: rows NONE:{S,N,S} LEFT:{W,NW,SW}
// RIGHT:{E,NE,SE}. NOTE the NONE/NONE cell is South — upstream makes the
// x-axis-none-only stick face south, not "no direction".
var FaceDir = [3][3]Direction{
	// NONE             UP                DOWN
	{South, North, South},        // NONE
	{West, NorthWest, SouthWest}, // LEFT
	{East, NorthEast, SouthEast}, // RIGHT
}

// StickPush mirrors enum class StickPush : uint8_t
// (controller_motion.h:32-37): how hard the left stick is pushed.
type StickPush uint8

// AxisDirectionX mirrors enum AxisDirectionX (axis_direction.h:7-11).
type AxisDirectionX uint8

// AxisDirectionY mirrors enum AxisDirectionY (axis_direction.h:12-16).
type AxisDirectionY uint8

// AxisDirection mirrors struct AxisDirection (axis_direction.h:19-22): the
// 8-way direction of a thumb stick (or D-Pad at the wiring layer).
type AxisDirection struct {
	X AxisDirectionX
	Y AxisDirectionY
}

// Direction mirrors enum class Direction : uint8_t
// (engine/direction.hpp:10-20), ordinal order preserved.
type Direction uint8

// PlrMode mirrors enum PLR_MODE (player.h:109) collapsed to the values the
// movement pipeline reads.
type PlrMode uint8

// Left (direction.hpp:24-29) maps a facing to its left turn.
func Left(facing Direction) Direction {
	return leftRotation[facing]
}

// Right (direction.hpp:31-36) maps a facing to its right turn.
func Right(facing Direction) Direction {
	return rightRotation[facing]
}

// DirectionDelta (displacement.hpp:247-275) returns the tile step (dX,dY)
// a facing corresponds to; NoDirection steps nowhere.
func DirectionDelta(facing Direction) (dX, dY int) {
	return deltaByDirection[facing][0], deltaByDirection[facing][1]
}

// ClassifyStickPush ports GetAnalogStickPush (controller_motion.cpp:210-218):
// compare squared magnitudes to avoid the sqrt; 0.0 → None (the deadzone is
// applied upstream by ScaleJoystickAxes — raw input here never re-deadzones),
// ≥ StickDirectionThresholdSquared → Firm, else Weak.
func ClassifyStickPush(stickX, stickY float64) StickPush {
	magnitudeSquared := (stickX * stickX) + (stickY * stickY)
	if magnitudeSquared == 0.0 {
		return PushNone
	}
	if magnitudeSquared >= StickDirectionThresholdSquared {
		return PushFirm
	}
	return PushWeak
}

// StickDirection ports GetAnalogStickDirection (controller_motion.cpp:220-242):
// 8-way sectoring with 22.5° cutoffs. absX == 0 short-circuits to a pure
// vertical by sign (the ratio division is undefined there); ratio ≤ cutoff →
// pure horizontal; ratio ≥ 1/cutoff → pure vertical; else diagonal by both
// signs. At ratio == DiagonalCutoff exactly the `<=` selects the CARDINAL —
// the float64 branch mirrors the upstream float comparison.
func StickDirection(stickX, stickY float64) AxisDirection {
	absX := abs(stickX)
	absY := abs(stickY)
	result := AxisDirection{X: AxisXNone, Y: AxisYNone}

	if absX == 0.0 {
		return verticalBySign(stickY)
	}

	ratio := absY / absX
	if ratio <= DiagonalCutoff {
		return AxisDirection{X: horizontalBySign(stickX), Y: AxisYNone}
	}
	if ratio >= 1.0/DiagonalCutoff {
		return verticalBySign(stickY)
	}

	result.X = horizontalBySign(stickX)
	result.Y = verticalBySign(stickY).Y
	return result
}

// verticalBySign is the UP/DOWN-by-sign slice shared by the absX==0 and
// steep-ratio branches (controller_motion.cpp:225-227,234-237).
func verticalBySign(stickY float64) AxisDirection {
	if stickY > 0 {
		return AxisDirection{X: AxisXNone, Y: AxisYUp}
	}
	return AxisDirection{X: AxisXNone, Y: AxisYDown}
}

// horizontalBySign is the LEFT/RIGHT-by-sign slice
// (controller_motion.cpp:230-233,238-240).
func horizontalBySign(stickX float64) AxisDirectionX {
	if stickX > 0 {
		return AxisXRight
	}
	return AxisXLeft
}

// abs is std::fabs for float64 (controller_motion.cpp:221-222 std::fabs).
func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
