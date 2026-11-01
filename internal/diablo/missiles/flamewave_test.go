// Package missiles_test pins AddFlameWave (missiles.cpp:2170-2179) and
// ProcessFlameWave (missiles.cpp:3475-3507).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// flameWaveAddCases hold the hand-computed adds: GenerateRnd(10) + the caster's
// character level + 1 (cpp:2172), speed 16, the 255-tick duration and the
// render adjustment that moves the tile one step south and the sprite 32
// pixels up (cpp:2177-2178).
var flameWaveAddCases = []flameWaveAddCase{
	{
		name:  "level 10 sorcerer rolls 7 for 18 damage",
		level: 10, draw: 7, wantDam: 18,
	},
	{
		name:  "level 0 rolls the minimum",
		level: 0, draw: 0, wantDam: 1,
	},
}

// flameWaveProcessCases hold the hand-computed ticks: the var1 counter flips the
// idle group when it reaches AnimLen (cpp:3484-3488), the light starts at
// radius 2 and walks ExpLight clamping at its last entry (cpp:3497-3501), and a
// zero duration deletes the wave and drops the light (cpp:3493-3495).
var flameWaveProcessCases = []flameWaveProcessCase{
	{
		name:     "first tick adds the light at radius 2",
		duration: 255, var2: 0, animLen: 5,
		wantDuration: 255, wantVar1: 1, wantVar2: 1, wantLid: 51, wantLight: true, wantChange: 2,
	},
	{
		name:     "var1 reaching AnimLen flips the idle group",
		duration: 255, var1: 4, var2: 5, animLen: 5, rnd: []int{3},
		wantDuration: 255, wantVar1: 5, wantVar2: 6, wantGroup: 1, wantFrame: 4, wantChange: 6,
	},
	{
		name:     "sweep clamps on the last ExpLight entry",
		duration: 255, var2: 11, animLen: 9,
		wantDuration: 255, wantVar1: 1, wantVar2: 11, wantChange: 12,
	},
	{
		name:     "zero duration deletes the wave and unlights it",
		duration: 0, var2: 3, animLen: 1, rnd: []int{0},
		wantVar1: 1, wantVar2: 3, wantGroup: 1, wantFrame: 1, wantUnlight: true,
	},
}

// flameWaveAddCase pins one AddFlameWave row.
type flameWaveAddCase struct {
	name    string
	level   int
	draw    int
	wantDam int
}

// flameWaveProcessCase pins one ProcessFlameWave tick.
type flameWaveProcessCase struct {
	name         string
	duration     int
	var1         int
	var2         int
	animLen      int
	rnd          []int
	wantDuration int
	wantVar1     int
	wantVar2     int
	wantGroup    int
	wantFrame    int
	wantLid      int
	wantChange   uint8
	wantLight    bool
	wantUnlight  bool
}

// TestAddFlameWave pins the damage roll, velocity, duration and the two render
// adjustments.
func TestAddFlameWave(t *testing.T) {
	// given
	cases := flameWaveAddCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			fx := newAddFixture(t, addScript{
				start: [2]int{0, 0}, dst: [2]int{10, 10}, caster: int(missiles.TargetMonsters),
				source: 0, level: tc.level, rnd: []int{tc.draw},
			})

			// when
			missiles.AddFlameWave(fx.world, fx.missile, fx.param)
			fx.tape.assertConsumed()

			// then
			assertFlameWaveAdd(t, fx.missile, tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// TestProcessFlameWave pins the var1 counter, the idle flip, the light sweep
// and the delete tail.
func TestProcessFlameWave(t *testing.T) {
	// given
	cases := flameWaveProcessCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			w, rec := flameWaveWorld(t, tc)

			// when
			m := flameWaveMissile(tc)
			missiles.ProcessFlameWave(w, &m)

			// then
			assertFlameWaveTick(t, m, rec, tc)
			assertFlameWaveTile(t, m)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// flameWaveWorld wires the light sinks and the scripted GenerateRnd draws.
func flameWaveWorld(t *testing.T, tc flameWaveProcessCase) (*missiles.HitWorld, *fireWallTick) {
	t.Helper()

	tape := &rndTape{t: t, tape: tc.rnd}
	rec := &fireWallTick{added: &lightCall{}}
	w := &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: missiles.NoopProcessCallbacks(),
	}
	w.ProcessCallbacks.GenerateRnd = tape.draw
	w.ProcessCallbacks.AddLight = func(x, y int, radius uint8) int {
		*rec.added = lightCall{x: x, y: y, radius: radius}

		return tc.wantLid
	}
	w.ProcessCallbacks.ChangeLight = func(_, _, _ int, radius uint8) { rec.changeR = radius }
	w.ProcessCallbacks.AddUnLight = func(int) { rec.unlighted = true }

	return w, rec
}

// flameWaveMissile seeds the wave exactly as AddFlameWave leaves it: the
// traveled vector of the launch tile (4, 4), the tile shifted one step south
// and the sprite 32 pixels up (missiles.cpp:3475-3482 undo both).
func flameWaveMissile(tc flameWaveProcessCase) missiles.Missile {
	m := missiles.Missile{
		Type: missiles.FlameWave, Duration: tc.duration, Var1: tc.var1, Var2: tc.var2,
		AnimLen: tc.animLen, Dam: 18, Lid: 7,
	}
	m.SetTraveledFromTile(4, 4)
	m.Position.Tile = missiles.TilePosition{X: 5, Y: 5}
	m.Position.Offset.DeltaY = -32

	return m
}

// assertFlameWaveAdd checks one post-AddFlameWave state.
func assertFlameWaveAdd(t *testing.T, m *missiles.Missile, tc flameWaveAddCase) {
	t.Helper()

	if m.Dam != tc.wantDam { // cpp:2172
		t.Fatalf("Dam = %d, want %d", m.Dam, tc.wantDam)
	}

	velocity := m.Position.Velocity
	if velocity.DeltaX != 0 || velocity.DeltaY != 524288 { // cpp:2173
		t.Fatalf("velocity = (%d, %d), want (0, 524288)", velocity.DeltaX, velocity.DeltaY)
	}

	if m.Duration != 255 { // cpp:2174
		t.Fatalf("Duration = %d, want 255", m.Duration)
	}

	wantTile := missiles.TilePosition{X: 1, Y: 1}
	if m.Position.Tile != wantTile || m.Position.Offset.DeltaY != -32 { // cpp:2177-2178
		t.Fatalf("tile/offsetY = %v/%d, want %v/-32", m.Position.Tile, m.Position.Offset.DeltaY, wantTile)
	}
}

// assertFlameWaveTick checks one post-tick state.
func assertFlameWaveTick(t *testing.T, m missiles.Missile, rec *fireWallTick, tc flameWaveProcessCase) {
	t.Helper()

	if m.Duration != tc.wantDuration {
		t.Fatalf("Duration = %d, want %d", m.Duration, tc.wantDuration)
	}

	if m.Var1 != tc.wantVar1 || m.Var2 != tc.wantVar2 {
		t.Fatalf("var1/var2 = %d/%d, want %d/%d", m.Var1, m.Var2, tc.wantVar1, tc.wantVar2)
	}

	if m.MFNum != tc.wantGroup || m.AnimFrame != tc.wantFrame {
		t.Fatalf("group/frame = %d/%d, want %d/%d", m.MFNum, m.AnimFrame, tc.wantGroup, tc.wantFrame)
	}

	assertFlameWaveLights(t, m, rec, tc)
}

// assertFlameWaveLights checks the light sink calls of one tick.
func assertFlameWaveLights(t *testing.T, m missiles.Missile, rec *fireWallTick,
	tc flameWaveProcessCase,
) {
	t.Helper()

	if tc.wantLid != 0 && m.Lid != tc.wantLid {
		t.Fatalf("Lid = %d, want %d", m.Lid, tc.wantLid)
	}

	if tc.wantLight && rec.added.radius != 2 {
		t.Fatalf("AddLight radius = %d, want 2", rec.added.radius)
	}

	if tc.wantChange != 0 && rec.changeR != tc.wantChange {
		t.Fatalf("ChangeLight radius = %d, want %d", rec.changeR, tc.wantChange)
	}

	if rec.unlighted != tc.wantUnlight {
		t.Fatalf("unlit = %v, want %v", rec.unlighted, tc.wantUnlight)
	}
}

// assertFlameWaveTile checks that the head and tail tile shifts cancel, so the
// tick hands back the tile AddFlameWave left (missiles.cpp:3481, 3504).
func assertFlameWaveTile(t *testing.T, m missiles.Missile) {
	t.Helper()

	want := missiles.TilePosition{X: 5, Y: 5}
	if m.Position.Tile != want {
		t.Fatalf("tile = %v, want %v", m.Position.Tile, want)
	}
}
