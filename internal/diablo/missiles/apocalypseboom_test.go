// Package missiles_test pins AddApocalypseBoom (missiles.cpp:2455-2460) and
// ProcessApocalypseBoom (missiles.cpp:3722-3732).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// apocalypseBoomAddCases pin the tile/start snap to the target tile and the
// duration taken from the row's AnimLen (cpp:2457-2459).
var apocalypseBoomAddCases = []apocalypseBoomAddCase{
	{name: "snap lands the boom on the target tile", dst: [2]int{7, 11}, animLen: 6},
	{name: "an empty AnimLen seeds a zero duration", dst: [2]int{1, 2}, animLen: 0},
}

// apocalypseBoomCases hold the hand-computed ticks: the collision only runs
// while var1 is still 0 (cpp:3725-3726), a hit latches var1 to 1 (cpp:3727-3728)
// and the last tick deletes the boom (cpp:3729-3730).
var apocalypseBoomCases = []apocalypseBoomCase{
	{name: "first tick probes the tile and survives", duration: 3, wantDuration: 2, wantProbes: 1},
	{
		name: "first tick latches var1 on a hit", duration: 3, hit: true,
		wantDuration: 2, wantVar1: 1, wantProbes: 1,
	},
	{
		name: "latched boom skips the collision", duration: 3, var1: 1, hit: true,
		wantDuration: 2, wantVar1: 1,
	},
	{
		name: "last tick deletes the boom", duration: 1, var1: 1,
		wantDuration: 0, wantVar1: 1, wantDel: true,
	},
}

// apocalypseBoomAddCase pins one AddApocalypseBoom row.
type apocalypseBoomAddCase struct {
	name    string
	dst     [2]int
	animLen int
}

// apocalypseBoomCase pins one ProcessApocalypseBoom tick.
type apocalypseBoomCase struct {
	name         string
	duration     int
	var1         int
	hit          bool
	wantDuration int
	wantVar1     int
	wantProbes   int
	wantDel      bool
}

// TestAddApocalypseBoom pins the tile/start snap and the AnimLen duration.
func TestAddApocalypseBoom(t *testing.T) {
	// given
	cases := apocalypseBoomAddCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			fx := newAddFixture(t, addScript{start: [2]int{3, 3}, dst: tc.dst, animLen: tc.animLen})

			// when
			missiles.AddApocalypseBoom(fx.world, fx.missile, fx.param)

			// then
			assertApocalypseBoomAdd(t, fx.missile, tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// TestProcessApocalypseBoom pins the var1 collision gate, the hit latch and the
// delete tail.
func TestProcessApocalypseBoom(t *testing.T) {
	// given
	cases := apocalypseBoomCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			w, probes := apocalypseBoomWorld(tc)

			// when
			m := apocalypseBoomMissile(tc)
			missiles.ProcessApocalypseBoom(w, &m)

			// then
			assertApocalypseBoomTick(t, m, tc)
			if probes() != tc.wantProbes {
				t.Fatalf("collision probes = %d, want %d", probes(), tc.wantProbes)
			}
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// assertApocalypseBoomAdd checks one post-AddApocalypseBoom state.
func assertApocalypseBoomAdd(t *testing.T, m *missiles.Missile, tc apocalypseBoomAddCase) {
	t.Helper()

	want := missiles.TilePosition{X: tc.dst[0], Y: tc.dst[1]}
	if m.Position.Tile != want || m.Position.Start != want { // cpp:2457-2458
		t.Fatalf("tile/start = %v/%v, want %v", m.Position.Tile, m.Position.Start, want)
	}

	if m.Duration != tc.animLen { // cpp:2459
		t.Fatalf("Duration = %d, want %d", m.Duration, tc.animLen)
	}
}

// apocalypseBoomMissile seeds one boom at tile (4, 4).
func apocalypseBoomMissile(tc apocalypseBoomCase) missiles.Missile {
	m := missiles.Missile{
		Type: missiles.ApocalypseBoom, Duration: tc.duration, Var1: tc.var1, Dam: 9,
	}
	m.SetTraveledFromTile(4, 4)

	return m
}

// assertApocalypseBoomTick checks one post-tick state.
func assertApocalypseBoomTick(t *testing.T, m missiles.Missile, tc apocalypseBoomCase) {
	t.Helper()

	if m.Duration != tc.wantDuration {
		t.Fatalf("Duration = %d, want %d", m.Duration, tc.wantDuration)
	}

	if m.Var1 != tc.wantVar1 {
		t.Fatalf("Var1 = %d, want %d", m.Var1, tc.wantVar1)
	}

	if m.DelFlag != tc.wantDel {
		t.Fatalf("DelFlag = %v, want %v", m.DelFlag, tc.wantDel)
	}
}

// apocalypseBoomWorld puts one monster on the boom's tile and counts the
// monster-target probes of the tick's collision check.
func apocalypseBoomWorld(tc apocalypseBoomCase) (*missiles.HitWorld, func() int) {
	probes := 0
	w := &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: missiles.NoopProcessCallbacks(),
	}
	w.DMonster[4][4] = 1
	w.HitCallbacks.MonsterTargeted = func(int16, *missiles.Direction) bool {
		probes++

		return tc.hit
	}
	w.HitCallbacks.MonsterMHit = func(int, int, int, missiles.MissileID,
		missiles.DamageType, bool, missiles.KeepOnCollision,
	) bool {
		return true
	}

	return w, func() int { return probes }
}
