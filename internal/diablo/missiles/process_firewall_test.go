// Package missiles_test pins ProcessFireWall (missiles.cpp:3060-3096).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// fireWallProcessCases holds the hand-computed ticks: the burn-out group flip
// lands on var1 with a GenerateRnd(11) + 1 frame (cpp:3066-3069), the start
// group flips one tick before the animation ends (cpp:3070-3073), and the
// expansion light sweeps ExpLight[2..12] outward and back through
// 11 - |var2 - 11|, adding the light on var2 == 0 (cpp:3085-3093).
var fireWallProcessCases = []fireWallProcessCase{
	{
		name:     "idle group flip on the var1 tick draws the burn-out frame",
		duration: 4, var1: 3, animLen: 20, rnd: []int{6},
		wantDuration: 3, wantGroup: 1, wantFrame: 7, wantAnimAdd: 1,
	},
	{
		name:     "start group flips one tick before the animation ends and lights tile",
		duration: 3, var1: 99, animLen: 3, light: 42,
		wantDuration: 2, wantGroup: 0, wantFrame: 13, wantAnimAdd: -1,
		wantLight: 42, wantVar2: 1, wantChange: 2,
	},
	{
		name:     "mid-sweep tick only moves the light",
		duration: 3, var1: 99, animLen: 3, var2: 5,
		wantDuration: 2, wantGroup: 0, wantFrame: 13, wantAnimAdd: -1,
		wantChange: 6, wantVar2: 6,
	},
	{
		name:     "sweep stops past twice the last ExpLight index",
		duration: 3, var1: 99, animLen: 9, var2: 23,
		wantDuration: 2, wantGroup: 0, wantAnimAdd: 1,
		wantVar2: 23,
	},
	{
		name:     "reaching zero deletes the wall and unlights it",
		duration: 1, var1: 5, animLen: 9,
		wantDuration: 0, wantAnimAdd: 1, wantDel: true, wantUnlight: true,
	},
}

// fireWallProcessCase pins one ProcessFireWall tick.
type fireWallProcessCase struct {
	name         string
	duration     int
	var1         int
	var2         int
	light        int
	animLen      int
	rnd          []int
	wantDuration int
	wantGroup    int
	wantFrame    int
	wantAnimAdd  int
	wantVar2     int
	wantChange   uint8
	wantLight    int
	wantDel      bool
	wantUnlight  bool
}

// fireWallTick records the light sink calls of one tick.
type fireWallTick struct {
	added     *lightCall
	changeR   uint8
	changes   int
	unlighted bool
}

// TestProcessFireWall pins the two frame-group flips, the expansion light
// sweep and the delete tail.
func TestProcessFireWall(t *testing.T) {
	// given
	cases := fireWallProcessCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			w, rec := fireWallWorld(t, tc)
			m := fireWallMissile(tc)

			// when
			missiles.ProcessFireWall(w, &m)

			// then
			assertFireWallTick(t, m, rec, tc)
			assertFireWallLights(t, m, rec, tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// fireWallWorld wires the tick's light sinks and the scripted GenerateRnd.
func fireWallWorld(t *testing.T, tc fireWallProcessCase) (*missiles.HitWorld, *fireWallTick) {
	t.Helper()

	rec := &fireWallTick{}
	tape := &rndTape{t: t, tape: tc.rnd}
	w := &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: missiles.NoopProcessCallbacks(),
	}
	w.ProcessCallbacks.GenerateRnd = tape.draw
	w.ProcessCallbacks.AddLight = func(x, y int, radius uint8) int {
		*rec.added = lightCall{x: x, y: y, radius: radius}

		return tc.light
	}
	w.ProcessCallbacks.ChangeLight = func(_, _, _ int, radius uint8) {
		rec.changes++
		rec.changeR = radius
	}
	w.ProcessCallbacks.AddUnLight = func(int) { rec.unlighted = true }
	rec.added = &lightCall{}

	return w, rec
}

// fireWallMissile seeds the wall at tile (0, 0) with the case's animation state.
func fireWallMissile(tc fireWallProcessCase) missiles.Missile {
	m := missiles.Missile{
		Type: missiles.FireWall, Duration: tc.duration, Var1: tc.var1, Var2: tc.var2,
		Lid: 7, Dam: 40, AnimLen: tc.animLen, AnimAdd: 1,
	}
	m.SetTraveledFromTile(0, 0)

	return m
}

// assertFireWallTick checks the animation and delete state of one tick.
func assertFireWallTick(t *testing.T, m missiles.Missile, rec *fireWallTick,
	tc fireWallProcessCase,
) {
	t.Helper()

	if m.Duration != tc.wantDuration {
		t.Fatalf("Duration = %d, want %d", m.Duration, tc.wantDuration)
	}

	if m.MFNum != tc.wantGroup {
		t.Fatalf("frame group = %d, want %d", m.MFNum, tc.wantGroup)
	}

	if tc.wantFrame != 0 && m.AnimFrame != tc.wantFrame {
		t.Fatalf("AnimFrame = %d, want %d", m.AnimFrame, tc.wantFrame)
	}

	if m.AnimAdd != tc.wantAnimAdd {
		t.Fatalf("AnimAdd = %d, want %d", m.AnimAdd, tc.wantAnimAdd)
	}

	if m.DelFlag != tc.wantDel || rec.unlighted != tc.wantUnlight {
		t.Fatalf("del/unlight = %v/%v, want %v/%v",
			m.DelFlag, rec.unlighted, tc.wantDel, tc.wantUnlight)
	}
}

// assertFireWallLights checks the expansion light sweep of one tick.
func assertFireWallLights(t *testing.T, m missiles.Missile, rec *fireWallTick,
	tc fireWallProcessCase,
) {
	t.Helper()

	if m.Var2 != tc.wantVar2 {
		t.Fatalf("Var2 = %d, want %d", m.Var2, tc.wantVar2)
	}

	if tc.wantLight != 0 && m.Lid != tc.wantLight {
		t.Fatalf("Lid = %d, want %d", m.Lid, tc.wantLight)
	}

	if tc.wantChange != 0 && rec.changeR != tc.wantChange {
		t.Fatalf("ChangeLight radius = %d (%d calls), want %d", rec.changeR, rec.changes, tc.wantChange)
	}
}
