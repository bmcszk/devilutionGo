package missiles_test

import "github.com/bmcszk/devilutionGo/internal/diablo/missiles"

import (
	"fmt"
	"testing"
)

// TestWorldToScreen ports Displacement::missiles.WorldToScreen
// (Source/engine/displacement.hpp:151): {(dy - dx) * 32, (dy + dx) * -16}.
func TestWorldToScreen(t *testing.T) {
	cases := []struct {
		name string
		in   missiles.PositionDisplacement
		want missiles.PositionDisplacement
	}{
		{name: "zero", in: disp(0, 0), want: disp(0, 0)},
		{name: "axis x", in: disp(10, 0), want: disp(-320, -160)},
		{name: "axis y", in: disp(0, 10), want: disp(320, -160)},
		{name: "negative mixed", in: disp(-3, 5), want: disp(256, -32)},
		{name: "diagonal down-right", in: disp(7, 7), want: disp(0, -224)},
		{name: "diagonal up-left", in: disp(-4, -4), want: disp(0, 128)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := missiles.WorldToScreen(tc.in)
			if got != tc.want {
				t.Fatalf("missiles.WorldToScreen(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

// TestScreenToWorld ports Displacement::missiles.ScreenToWorld
// (Source/engine/displacement.hpp:160): {(2*dy + dx) / -64, (2*dy - dx) / -64}.
// C++ int division truncates toward zero, matching Go's `/` on signed ints;
// the "trunc toward zero" row pins that: flooring would give (-5, -2).
func TestScreenToWorld(t *testing.T) {
	cases := []struct {
		name string
		in   missiles.PositionDisplacement
		want missiles.PositionDisplacement
	}{
		{name: "zero", in: disp(0, 0), want: disp(0, 0)},
		{name: "axis y", in: disp(320, -160), want: disp(0, 10)},
		{name: "negative mixed", in: disp(256, -32), want: disp(-3, 5)},
		{name: "trunc toward zero", in: disp(96, 96), want: disp(-4, -1)},
		{name: "negative numerator", in: disp(-65, -33), want: disp(2, 0)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := missiles.ScreenToWorld(tc.in)
			if got != tc.want {
				t.Fatalf("missiles.ScreenToWorld(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

// TestWorldScreenRoundTrip pins the exact inverse pair: whole-tile world
// coordinates survive missiles.ScreenToWorld(missiles.WorldToScreen(w)) unchanged
// (displacement.hpp:151,:160).
func TestWorldScreenRoundTrip(t *testing.T) {
	worlds := []missiles.PositionDisplacement{
		disp(0, 0),
		disp(10, 0),
		disp(0, 10),
		disp(-3, 5),
		disp(7, 7),
		disp(-4, -4),
	}
	for _, w := range worlds {
		t.Run(fmt.Sprintf("%v", w), func(t *testing.T) {
			got := missiles.ScreenToWorld(missiles.WorldToScreen(w))
			if got != w {
				t.Fatalf("missiles.ScreenToWorld(missiles.WorldToScreen(%v)) = %v, want %v", w, got, w)
			}
		})
	}
}

// TestScreenToMissile ports Displacement::missiles.ScreenToMissile
// (Source/engine/displacement.hpp:174-183): xNum = 2*dy + dx;
// yNum = 2*dy - dx; xOff/yOff = +32 when the numerator is >= 0 else -32;
// result {(xNum + xOff) / 64, (yNum + yOff) / 64} — rounds toward the
// offset side, then int division truncates toward zero.
//
// Note (320, -160) — the exact screen image of tile (0, 10) — lands on
// (0, -10): upstream's asymmetric rounding, replicated verbatim.
func TestScreenToMissile(t *testing.T) {
	cases := []struct {
		name string
		in   missiles.PositionDisplacement
		want missiles.PositionDisplacement
	}{
		{name: "zero nudges positive", in: disp(0, 0), want: disp(0, 0)},
		{name: "axis y asymmetry", in: disp(320, -160), want: disp(0, -10)},
		{name: "xNum -1 takes negative nudge", in: disp(65, -33), want: disp(0, -2)},
		{name: "yNum 3.5 truncates down", in: disp(-96, 48), want: disp(0, 3)},
		{name: "exact halving", in: disp(-64, 32), want: disp(0, 2)},
		{name: "diagonal mirror", in: disp(0, 128), want: disp(4, 4)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := missiles.ScreenToMissile(tc.in)
			if got != tc.want {
				t.Fatalf("missiles.ScreenToMissile(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

// TestScreenToLight ports Displacement::missiles.ScreenToLight
// (Source/engine/displacement.hpp:184): {(2*dy + dx) / 8, (2*dy - dx) / 8}

// — the coarser light grid, note the positive divisor vs missiles.ScreenToWorld's
// negative one.
func TestScreenToLight(t *testing.T) {
	cases := []struct {
		name string
		in   missiles.PositionDisplacement
		want missiles.PositionDisplacement
	}{
		{name: "zero", in: disp(0, 0), want: disp(0, 0)},
		{name: "axis y", in: disp(320, -160), want: disp(0, -80)},
		{name: "mixed", in: disp(96, 96), want: disp(36, 12)},
		{name: "negative", in: disp(-96, 48), want: disp(0, 24)},
		{name: "trunc toward zero", in: disp(23, -7), want: disp(1, -4)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := missiles.ScreenToLight(tc.in)
			if got != tc.want {
				t.Fatalf("missiles.ScreenToLight(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

// TestScreenToLightVsScreenToWorldScale pins the divisor relation between
// the two ports (displacement.hpp:160,:184): dividing by +8 instead of
// -64 means missiles.ScreenToLight(s) == -8 * missiles.ScreenToWorld(s) whenever both
// numerators divide exactly.
func TestScreenToLightVsScreenToWorldScale(t *testing.T) {
	screens := []missiles.PositionDisplacement{
		disp(320, -160),
		disp(-96, 48),
		disp(256, -256),
	}
	for _, s := range screens {
		t.Run(fmt.Sprintf("%v", s), func(t *testing.T) {
			light := missiles.ScreenToLight(s)
			world := missiles.ScreenToWorld(s)
			want := disp(-8*world.DeltaX, -8*world.DeltaY)
			if light != want {
				t.Fatalf("missiles.ScreenToLight(%v) = %v, want -8*missiles.ScreenToWorld = %v", s, light, want)
			}
		})
	}
}

// TestWorldToNormalScreen ports Displacement::missiles.WorldToNormalScreen
// (Source/engine/displacement.hpp:195-206): rotated = {dy - dx, -(dy + dx)};
// n = normalized(rotated); return {n.dx, n.dy / 2}.
//
// normalized (displacement.hpp:363-370):
// mag = float32(sqrt(float64(dx*dx + dy*dy)));
// out = {int(float32(dx << 16) / mag), int(float32(dy << 16) / mag)}

// — shift first, float32 divide, int() truncates toward zero.
//
// Hand computations (sqrt(2) ≈ 1.4142135, sqrt(50) ≈ 7.0710678):
//
//	(1, 0): rotated = (0-1, -(0+1)) = (-1, -1);
//	        -65536 / 1.4142135 ≈ -46340.95 → -46340 (both axes);
//	        dy/2 = -46340/2 = -23170 (exact).
//	        → (-46340, -23170)
//	(0, 1): rotated = (1-0, -(1+0)) = (1, -1);
//	        65536 / 1.4142135 ≈ 46340.95 → 46340; -46340.95 → -46340;
//	        dy/2 = -23170.
//	        → (46340, -23170)
//	(3, 4): rotated = (4-3, -(4+3)) = (1, -7);
//	        65536 / 7.0710678 ≈ 9268.19 → 9268;
//	        -7 << 16 = -458752; -458752 / 7.0710678 ≈ -64877.4 → -64877;
//	        dy/2 = -64877/2 = -32438 (trunc toward zero past -32438.5).
//	        → (9268, -32438)
func TestWorldToNormalScreen(t *testing.T) {
	cases := []struct {
		name string
		in   missiles.PositionDisplacement
		want missiles.PositionDisplacement
	}{
		{name: "east tile", in: disp(1, 0), want: disp(-46340, -23170)},
		{name: "south tile", in: disp(0, 1), want: disp(46340, -23170)},
		{name: "3-4 diagonal", in: disp(3, 4), want: disp(9268, -32438)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := missiles.WorldToNormalScreen(tc.in)
			if got != tc.want {
				t.Fatalf("missiles.WorldToNormalScreen(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

// TestManhattanDistance ports PointOf::missiles.ManhattanDistance
// (Source/engine/point.hpp:147-152): |dx| + |dy| — companion to
// missiles.PathDistance (WalkingDistance, point.hpp:154-159, motion.go).
func TestManhattanDistance(t *testing.T) {
	cases := []struct {
		name string
		from missiles.TilePosition
		to   missiles.TilePosition
		want int
	}{
		{name: "diagonal", from: missiles.TilePosition{X: 0, Y: 0}, to: missiles.TilePosition{X: 3, Y: 4}, want: 7},
		{name: "both negative deltas", from: missiles.TilePosition{X: -2, Y: 5},
			to: missiles.TilePosition{X: 4, Y: -1}, want: 12},
		{name: "same tile", from: missiles.TilePosition{X: 2, Y: 2}, to: missiles.TilePosition{X: 2, Y: 2}, want: 0},
		{name: "axis x", from: missiles.TilePosition{X: -1, Y: 0}, to: missiles.TilePosition{X: -5, Y: 0}, want: 4},
		{name: "axis y", from: missiles.TilePosition{X: 0, Y: -3}, to: missiles.TilePosition{X: 0, Y: 2}, want: 5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := missiles.ManhattanDistance(tc.from, tc.to)
			if got != tc.want {
				t.Fatalf("missiles.ManhattanDistance(%v, %v) = %d, want %d", tc.from, tc.to, got, tc.want)
			}
		})
	}
}

// disp builds a missiles.PositionDisplacement from plain ints (test shorthand).
func disp(x, y int32) missiles.PositionDisplacement {
	return missiles.PositionDisplacement{DeltaX: x, DeltaY: y}
}
