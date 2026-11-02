// Package missiles_test pins AddBigExplosion and AddRingOfFire
// (missiles.cpp:1603-1618, 1686-1691).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
)

// bigExplosionCases is the slice-3c damage-roll table (missiles.cpp:1605-1613).
var bigExplosionCases = []bigExplosionCase{
	{
		name: "2d10 plus level, scaled by the spell level",
		script: addScript{
			start: [2]int{40, 40}, dst: [2]int{41, 40}, source: 0, level: 3,
			rnd: []int{4, 5}, animLen: 12, light: 9, spllvl: 2,
		},
		occupied: true,
		wantDam:  34, // 2*(3+9)+4 = 28, then +8% twice (ScaleSpellEffect, cpp:119-125)
		wantHits: 9,
	},
	{
		name: "no monsters in the square means no damage rolls",
		script: addScript{
			start: [2]int{40, 40}, dst: [2]int{41, 40}, source: 0, level: 0,
			rnd: []int{1, 1}, animLen: 20, light: 9,
		},
		wantDam:  8, // 2*(0+2)+4
		wantHits: 0,
	},
}

// explosionHits scripts the monster side of the blast sweep and counts the
// damage rolls.
type explosionHits struct {
	hits int
}

// bigExplosionCase is one scripted AddBigExplosion run.
type bigExplosionCase struct {
	name     string
	script   addScript
	occupied bool
	wantDam  int
	wantHits int
}

func TestAddBigExplosion_PlayerCastSweepsItsSquare(t *testing.T) {
	// given
	cases := bigExplosionCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()
			runBigExplosion(t, tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

func TestAddBigExplosion_NonPlayerCastOnlySeedsTheBlast(t *testing.T) {
	// given
	fx := newAddFixture(t, addScript{
		start: [2]int{40, 40}, dst: [2]int{41, 40}, source: -1, animLen: 12, light: 9,
	})
	hits := &explosionHits{}
	fx.world.HitCallbacks = hits.hitCallbacks()
	occupySquare(fx.world, 40, 40)

	// when
	missiles.AddBigExplosion(fx.world, fx.missile, fx.param)

	// then
	if hits.hits != 0 {
		t.Fatalf("damage rolls = %d, want 0 for a trap cast (missiles.cpp:1605)", hits.hits)
	}

	if fx.missile.Duration != 11 {
		t.Fatalf("duration = %d, want 11", fx.missile.Duration)
	}
}

func TestAddRingOfFire_RemembersItsLaunchTile(t *testing.T) {
	// given
	fx := newAddFixture(t, addScript{start: [2]int{42, 44}, dst: [2]int{43, 44}})

	// when
	missiles.AddRingOfFire(fx.world, fx.missile, fx.param)

	// then
	if fx.missile.Var1 != 42 || fx.missile.Var2 != 44 {
		t.Fatalf("var1/var2 = %d/%d, want 42/44", fx.missile.Var1, fx.missile.Var2)
	}

	if fx.missile.Duration != 7 {
		t.Fatalf("duration = %d, want 7", fx.missile.Duration)
	}
}

// hitCallbacks wires the counting monster roll into the collision probe.
func (h *explosionHits) hitCallbacks() missiles.HitCallbacks {
	cb := missiles.NoopHitCallbacks()
	cb.MonsterTargeted = func(int16, *missiles.Direction) bool { return true }
	cb.MonsterMHit = func(int, int, int, missiles.MissileID, missiles.DamageType, bool,
		missiles.KeepOnCollision,
	) bool {
		h.hits++

		return true
	}

	return cb
}

// runBigExplosion runs one scripted blast.
func runBigExplosion(t *testing.T, tc bigExplosionCase) {
	t.Helper()

	// given
	fx := newAddFixture(t, tc.script)
	hits := &explosionHits{}
	fx.world.HitCallbacks = hits.hitCallbacks()
	if tc.occupied {
		occupySquare(fx.world, tc.script.start[0], tc.script.start[1])
	}

	// when
	missiles.AddBigExplosion(fx.world, fx.missile, fx.param)
	fx.tape.assertConsumed()

	// then
	if fx.missile.Dam != tc.wantDam {
		t.Fatalf("dam = %d, want %d", fx.missile.Dam, tc.wantDam)
	}

	if hits.hits != tc.wantHits {
		t.Fatalf("damage rolls = %d, want %d", hits.hits, tc.wantHits)
	}

	if fx.missile.Lid != tc.script.light || fx.light.radius != 8 {
		t.Fatalf("lid/radius = %d/%d, want %d/8", fx.missile.Lid, fx.light.radius, tc.script.light)
	}

	if fx.missile.Duration != fx.missile.AnimLen-1 || fx.missile.MFNum != 0 {
		t.Fatalf("duration/group = %d/%d, want %d/0",
			fx.missile.Duration, fx.missile.MFNum, fx.missile.AnimLen-1)
	}
}

// occupySquare fills the 3x3 blast square around a tile with monsters.
func occupySquare(w *missiles.HitWorld, x, y int) {
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			w.DMonster[y+dy][x+dx] = 1
		}
	}
}
