// Package missiles_test pins ProcessNovaBall (missiles.cpp:3023-3042).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// novaBallProcessCases holds the hand-computed rows of the tick: the nova
// always decrements first (cpp:3026-3027) and records the remainder in j, a hit
// and a shrine on the anchor tile restore j after the collision zeroed the
// duration (cpp:3029-3036), and 0 deletes the missile (cpp:3038-3039).
var novaBallProcessCases = []novaBallProcessCase{
	{
		name: "plain tick decrements and survives", duration: 5,
		wantDuration: 4, wantDel: false,
	},
	{
		name: "hit restores the tick the collision zeroed", duration: 5, hit: true,
		wantDuration: 4, wantDel: false,
	},
	{
		name: "shrine on the anchor tile restores the tick", duration: 5,
		shrine: true, wantDuration: 4, wantDel: false,
	},
	{
		name: "reaching zero deletes the missile", duration: 1,
		wantDuration: 0, wantDel: true,
	},
}

// novaBallProcessCase pins one ProcessNovaBall tick.
type novaBallProcessCase struct {
	name         string
	duration     int
	hit          bool
	shrine       bool
	wantDuration int
	wantDel      bool
}

// TestProcessNovaBall pins the decrement, both refund branches and the delete.
func TestProcessNovaBall(t *testing.T) {
	// given
	cases := novaBallProcessCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			w := novaBallWorldFor(tc)
			if tc.shrine {
				putShrine(w, 3, 3)
			}
			m := novaBallMissile(tc.duration)

			// when
			missiles.ProcessNovaBall(w, &m)

			// then
			assertNovaBallTick(t, m, tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// novaBallMissile seeds a flying nova anchored on tile (3, 3) — the tile the
// shrine case occupies — with the given duration.
func novaBallMissile(duration int) missiles.Missile {
	m := missiles.Missile{Type: missiles.NovaBall, Duration: duration, Lid: missiles.NoLight}
	m.Var1 = 3
	m.Var2 = 3
	m.SetTraveledFromTile(3, 3)

	return m
}

// novaBallWorldFor wires the world one row needs: a plain grid, or one with
// the anchor tile occupied by a hittable monster.
func novaBallWorldFor(tc novaBallProcessCase) *missiles.HitWorld {
	w := novaBallWorld()
	if tc.hit {
		registerNovaHit(w)
	}

	if tc.shrine {
		putShrine(w, 3, 3)
	}

	return w
}

// novaBallWorld wires an empty grid for a plain tick.
func novaBallWorld() *missiles.HitWorld {
	return &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: missiles.NoopProcessCallbacks(),
	}
}

// registerNovaHit occupies the anchor tile with a monster and scripts a
// successful damage roll, so the tick takes the hit branch.
func registerNovaHit(w *missiles.HitWorld) {
	w.DMonster[3][3] = 1
	w.HitCallbacks.MonsterTargeted = func(int16, *missiles.Direction) bool { return true }
	w.HitCallbacks.MonsterMHit = func(int, int, int, missiles.MissileID,
		missiles.DamageType, bool, missiles.KeepOnCollision,
	) bool {
		return true
	}
}

// putShrine registers one shrine object on a tile.
func putShrine(w *missiles.HitWorld, x, y int) {
	w.Grid.Objects = append(w.Grid.Objects, world.WorldObject{Shrine: true})
	w.Grid.DObject[y][x] = len(w.Grid.Objects)
}

// assertNovaBallTick checks one post-tick state.
func assertNovaBallTick(t *testing.T, m missiles.Missile, tc novaBallProcessCase) {
	t.Helper()

	if m.Duration != tc.wantDuration {
		t.Fatalf("Duration = %d, want %d", m.Duration, tc.wantDuration)
	}

	if m.DelFlag != tc.wantDel {
		t.Fatalf("DelFlag = %v, want %v", m.DelFlag, tc.wantDel)
	}
}
