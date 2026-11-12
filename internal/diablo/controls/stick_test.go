// Package controls ports the fork gamepad stick semantics: StickPush
// classification, 8-way sectoring, the FaceDir table, and the
// Movement/TurnToDir/WalkInDir decision pipeline
// (Source/controls/plrctrls.cpp:100-105,1704-1754,1832-1893;
// Source/controls/controller_motion.cpp:210-242 @ fork ca91d778f).
package controls_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/controls"
)

var classifyPushCases = []struct {
	name string
	x    float64
	y    float64
	want controls.StickPush
}{
	{name: "zeros → None (deadzone applied upstream)", x: 0.0, y: 0.0, want: controls.PushNone},
	{name: "exact threshold boundary 0.36 is Firm (≥)", x: 0.6, y: 0.0, want: controls.PushFirm},
	{name: "0.6² on both axes is Firm", x: 0.6, y: 0.6, want: controls.PushFirm},
	{name: "half deflection 0.5² is Weak", x: 0.5, y: 0.0, want: controls.PushWeak},
	{name: "tiny sub-deadzone deflection is Weak, not None", x: 0.01, y: 0.0, want: controls.PushWeak},
	{name: "just below threshold stays Weak", x: 0.599999, y: 0.0, want: controls.PushWeak},
	{name: "full deflect diagonal 1/1 is Firm", x: 1.0, y: 1.0, want: controls.PushFirm},
	{name: "negative deflection classifies on magnitude", x: -0.7, y: 0.0, want: controls.PushFirm},
	{name: "corner compensation (0.7071,0.7071)²≈1 is Firm", x: 0.7071, y: 0.7071, want: controls.PushFirm},
}

var sectoringCases = []struct {
	name  string
	x     float64
	y     float64
	wantX controls.AxisDirectionX
	wantY controls.AxisDirectionY
}{
	// Upstream has no zero-vector special case: absX==0 falls through to
	// `stickY > 0 ? UP : DOWN`, so y==0 selects DOWN (unreachable from
	// Movement, which skips PushNone inputs first).
	{
		name:  "zero vector selects DOWN (no zero special case upstream)",
		x:     0,
		y:     0,
		wantX: controls.AxisXNone,
		wantY: controls.AxisYDown,
	},
	{name: "pure right", x: 1, y: 0, wantX: controls.AxisXRight, wantY: controls.AxisYNone},
	{name: "pure left", x: -1, y: 0, wantX: controls.AxisXLeft, wantY: controls.AxisYNone},
	{name: "absX==0, y>0 → UP special case", x: 0, y: 1, wantX: controls.AxisXNone, wantY: controls.AxisYUp},
	{name: "absX==0, y<0 → DOWN special case", x: 0, y: -1, wantX: controls.AxisXNone, wantY: controls.AxisYDown},
	{name: "absX==0, y=-0.3 → DOWN", x: 0, y: -0.3, wantX: controls.AxisXNone, wantY: controls.AxisYDown},
	{
		name:  "ratio == cutoff exactly → RIGHT cardinal",
		x:     1,
		y:     0.41421356,
		wantX: controls.AxisXRight,
		wantY: controls.AxisYNone,
	},
	{
		name:  "ratio == cutoff exactly → LEFT cardinal",
		x:     -1,
		y:     0.41421356,
		wantX: controls.AxisXLeft,
		wantY: controls.AxisYNone,
	},
	{
		name:  "ratio == cutoff exactly → UP cardinal",
		x:     0.41421356,
		y:     1,
		wantX: controls.AxisXNone,
		wantY: controls.AxisYUp,
	},
	{
		name:  "ratio == cutoff exactly → DOWN cardinal",
		x:     0.41421356,
		y:     -1,
		wantX: controls.AxisXNone,
		wantY: controls.AxisYDown,
	},
	{
		name:  "ratio just below cutoff → RIGHT cardinal",
		x:     1,
		y:     0.41421355,
		wantX: controls.AxisXRight,
		wantY: controls.AxisYNone,
	},
	{
		name:  "ratio just above cutoff → UP-RIGHT diagonal",
		x:     1,
		y:     0.41421357,
		wantX: controls.AxisXRight,
		wantY: controls.AxisYUp,
	},
	{
		name:  "ratio == 1/cutoff exactly → UP cardinal",
		x:     0.4142135,
		y:     1,
		wantX: controls.AxisXNone,
		wantY: controls.AxisYUp,
	},
	{
		name:  "just below inverse cutoff → UP-LEFT diagonal",
		x:     -0.4142136,
		y:     1,
		wantX: controls.AxisXLeft,
		wantY: controls.AxisYUp,
	},
	{
		name:  "just above inverse cutoff (x negated) → UP-LEFT diagonal",
		x:     -0.4142137,
		y:     1,
		wantX: controls.AxisXLeft,
		wantY: controls.AxisYUp,
	},
	{
		name:  "just above inverse cutoff (x positive) → UP-RIGHT diagonal",
		x:     0.4142137,
		y:     1,
		wantX: controls.AxisXRight,
		wantY: controls.AxisYUp,
	},
	{name: "up-right diagonal", x: 1, y: 1, wantX: controls.AxisXRight, wantY: controls.AxisYUp},
	{name: "up-left diagonal", x: -1, y: 1, wantX: controls.AxisXLeft, wantY: controls.AxisYUp},
	{name: "down-right diagonal", x: 1, y: -1, wantX: controls.AxisXRight, wantY: controls.AxisYDown},
	{name: "down-left diagonal", x: -1, y: -1, wantX: controls.AxisXLeft, wantY: controls.AxisYDown},
	{name: "shallow negative-x push → LEFT", x: -0.9, y: 0.2, wantX: controls.AxisXLeft, wantY: controls.AxisYNone},
	{name: "steep negative-y push → DOWN", x: 0.2, y: -0.9, wantX: controls.AxisXNone, wantY: controls.AxisYDown},
}

var faceDirCases = []struct {
	name string
	x    controls.AxisDirectionX
	y    controls.AxisDirectionY
	want controls.Direction
}{
	{name: "NONE/NONE → South", x: controls.AxisXNone, y: controls.AxisYNone, want: controls.South},
	{name: "NONE/UP → North", x: controls.AxisXNone, y: controls.AxisYUp, want: controls.North},
	{name: "NONE/DOWN → South", x: controls.AxisXNone, y: controls.AxisYDown, want: controls.South},
	{name: "LEFT/NONE → West", x: controls.AxisXLeft, y: controls.AxisYNone, want: controls.West},
	{name: "LEFT/UP → NorthWest", x: controls.AxisXLeft, y: controls.AxisYUp, want: controls.NorthWest},
	{name: "LEFT/DOWN → SouthWest", x: controls.AxisXLeft, y: controls.AxisYDown, want: controls.SouthWest},
	{name: "RIGHT/NONE → East", x: controls.AxisXRight, y: controls.AxisYNone, want: controls.East},
	{name: "RIGHT/UP → NorthEast", x: controls.AxisXRight, y: controls.AxisYUp, want: controls.NorthEast},
	{name: "RIGHT/DOWN → SouthEast", x: controls.AxisXRight, y: controls.AxisYDown, want: controls.SouthEast},
}

var rotationCases = []struct {
	name      string
	facing    controls.Direction
	wantLeft  controls.Direction
	wantRight controls.Direction
}{
	{name: "South", facing: controls.South, wantLeft: controls.SouthEast, wantRight: controls.SouthWest},
	{name: "SouthWest", facing: controls.SouthWest, wantLeft: controls.South, wantRight: controls.West},
	{name: "West", facing: controls.West, wantLeft: controls.SouthWest, wantRight: controls.NorthWest},
	{name: "NorthWest", facing: controls.NorthWest, wantLeft: controls.West, wantRight: controls.North},
	{name: "North", facing: controls.North, wantLeft: controls.NorthWest, wantRight: controls.NorthEast},
	{name: "NorthEast", facing: controls.NorthEast, wantLeft: controls.North, wantRight: controls.East},
	{name: "East", facing: controls.East, wantLeft: controls.NorthEast, wantRight: controls.SouthEast},
	{name: "SouthEast", facing: controls.SouthEast, wantLeft: controls.East, wantRight: controls.South},
}

var deltaCases = []struct {
	name   string
	facing controls.Direction
	wantX  int
	wantY  int
}{
	{name: "South → {1,1}", facing: controls.South, wantX: 1, wantY: 1},
	{name: "SouthWest → {0,1}", facing: controls.SouthWest, wantX: 0, wantY: 1},
	{name: "West → {-1,1}", facing: controls.West, wantX: -1, wantY: 1},
	{name: "NorthWest → {-1,0}", facing: controls.NorthWest, wantX: -1, wantY: 0},
	{name: "North → {-1,-1}", facing: controls.North, wantX: -1, wantY: -1},
	{name: "NorthEast → {0,-1}", facing: controls.NorthEast, wantX: 0, wantY: -1},
	{name: "East → {1,-1}", facing: controls.East, wantX: 1, wantY: -1},
	{name: "SouthEast → {1,0}", facing: controls.SouthEast, wantX: 1, wantY: 0},
	{name: "NoDirection → {0,0}", facing: controls.NoDirection, wantX: 0, wantY: 0},
}

// T1 — StickPush classification (controller_motion.cpp:210-218):
// squared magnitude 0.0 → None; ≥ StickDirectionThresholdSquared (0.6²=0.36) →
// Firm; else Weak. Raw floats in — no deadzone is applied here (upstream
// comment: the deadzone is applied by ScaleJoystickAxes; the k736 wiring lane
// owns scaling).
func TestClassifyStickPush(t *testing.T) {
	// when
	got := make([]controls.StickPush, 0, len(classifyPushCases))
	for _, tc := range classifyPushCases {
		got = append(got, controls.ClassifyStickPush(tc.x, tc.y))
	}
	// then
	for i, tc := range classifyPushCases {
		t.Run(tc.name, func(t *testing.T) {
			assertFirmFixture(t, tc.x, tc.y, tc.want)
			if got[i] != tc.want {
				t.Fatalf("ClassifyStickPush(%v,%v) = %v, want %v", tc.x, tc.y, got[i], tc.want)
			}
		})
	}
}

// T2 — 8-way sectoring (controller_motion.cpp:220-242): 22.5° cutoffs,
// DiagonalCutoff = 0.41421356 (tan 22.5°). absX==0 → pure UP/DOWN by sign;
// ratio ≤ cutoff → pure LEFT/RIGHT; ratio ≥ 1/cutoff → pure UP/DOWN; else
// diagonal by both signs. NOTE: at ratio == cutoff exactly the upstream
// comparison `ratio <= DiagonalCutoff` selects the CARDINAL (float64
// numerics identical to the C++ float branch).
func TestStickDirection(t *testing.T) {
	// when
	got := make([]controls.AxisDirection, 0, len(sectoringCases))
	for _, tc := range sectoringCases {
		got = append(got, controls.StickDirection(tc.x, tc.y))
	}
	// then
	for i, tc := range sectoringCases {
		t.Run(tc.name, func(t *testing.T) {
			if got[i].X != tc.wantX || got[i].Y != tc.wantY {
				t.Fatalf("StickDirection(%v,%v) = {%v,%v}, want {%v,%v}", tc.x, tc.y, got[i].X, got[i].Y, tc.wantX, tc.wantY)
			}
		})
	}
}

// T3 — FaceDir[3][3] vs the C++ literal (plrctrls.cpp:100-105), indexed
// [x][y] with X_NONE/LEFT/RIGHT = 0/1/2 and Y_NONE/UP/DOWN = 0/1/2:
// NONE row: S,N,S; LEFT row: W,NW,SW; RIGHT row: E,NE,SE.
func TestFaceDirTable(t *testing.T) {
	for _, tc := range faceDirCases {
		t.Run(tc.name, func(t *testing.T) {
			// when
			got := controls.FaceDir[tc.x][tc.y]
			// then
			if got != tc.want {
				t.Fatalf("FaceDir[%d][%d] = %v, want %v", tc.x, tc.y, got, tc.want)
			}
		})
	}
}

// Direction rotation parity (direction.hpp:24-36): Left = (facing+7)%8,
// Right = (facing+1)%8 over the ordinal S,SW,W,NW,N,NE,E,SE.
func TestDirectionRotation(t *testing.T) {
	for _, tc := range rotationCases {
		t.Run(tc.name, func(t *testing.T) {
			// when
			left, right := controls.Left(tc.facing), controls.Right(tc.facing)
			// then
			if left != tc.wantLeft || right != tc.wantRight {
				t.Fatalf("%v: Left = %v, want %v; Right = %v, want %v", tc.facing, left, tc.wantLeft, right, tc.wantRight)
			}
		})
	}
}

// DirectionDelta parity (displacement.hpp:247-275 fromDirection): the tile
// step each Direction maps to, {0,0} for NoDirection.
func TestDirectionDelta(t *testing.T) {
	for _, tc := range deltaCases {
		t.Run(tc.name, func(t *testing.T) {
			// when
			gotX, gotY := controls.DirectionDelta(tc.facing)
			// then
			if gotX != tc.wantX || gotY != tc.wantY {
				t.Fatalf("%v: delta = {%d,%d}, want {%d,%d}", tc.facing, gotX, gotY, tc.wantX, tc.wantY)
			}
		})
	}
}

// assertFirmFixture is T1's self-check: a fixture whose (x²+y²) reaches the
// walk threshold must expect Firm, else the row is lying about the physics.
func assertFirmFixture(t *testing.T, x, y float64, want controls.StickPush) {
	t.Helper()
	const thresholdSquared = 0.6 * 0.6
	if x*x+y*y >= thresholdSquared && want != controls.PushFirm {
		t.Fatalf("fixture error: (%v,%v) exceeds threshold but expects non-Firm", x, y)
	}
}
