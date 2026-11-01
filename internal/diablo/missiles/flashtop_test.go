// Package missiles_test pins AddFlashTop (missiles.cpp:2138-2152) and
// ProcessFlashTop (missiles.cpp:3454-3471).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// flashTopAddCases hold the hand-computed adds: a monster-cast, non-trap flash
// rolls level + 1 as the dice base and then GenerateRndSum(20, base) draws
// (cpp:2140-2145); a trap uses currlevel / 2 (cpp:2147); a non-monster cast
// keeps the caller's damage. All rows set _miPreFlag and duration 19
// (cpp:2150-2151).
// flashTopCases hold the hand-computed ticks: the caster is re-shielded every
// tick before the duration decrements (cpp:3456-3459) and unshielded only by
// the deleting tick (cpp:3466-3470), duration
// decrements, three offsets are probed (North, NorthEast, East) and 0 deletes.
var flashTopCases = []flashTopCase{
	{
		name:     "monster cast probes three tiles and shields the caster",
		duration: 3, caster: int(missiles.TargetMonsters), source: 0,
		wantDuration: 2, wantShield: 1, wantProbes: 3,
	},
	{
		name:     "player cast never touches the shield",
		duration: 3, caster: int(missiles.TargetPlayers), source: 0,
		wantDuration: 2, wantShield: 0, wantProbes: 3,
	},
	{
		name:     "traps leave the shield alone",
		duration: 3, caster: int(missiles.TargetMonsters), source: -1,
		wantDuration: 2, wantShield: 0, wantProbes: 3,
	},
	{
		name:     "last tick drops the shield and deletes the flash",
		duration: 1, caster: int(missiles.TargetMonsters), source: 0,
		wantShield: 1, wantProbes: 3, wantDel: true, wantUnshield: true,
	},
}

// flashTopCase pins one ProcessFlashTop tick.
var flashTopAddCases = []flashTopAddCase{
	{
		name:   "level 2 monster cast rolls two dice",
		caster: int(missiles.TargetMonsters), source: 0, level: 2, rnd: []int{3, 4, 5},
		wantDam: 22,
	},
	{
		name:   "trap cast takes half the level",
		caster: int(missiles.TargetMonsters), source: -1, currlevel: 9,
		wantDam: 4,
	},
	{
		name:   "player cast keeps the caller damage",
		caster: int(missiles.TargetPlayers), source: 0, level: 7,
		wantDam: 55,
	},
}

// flashTopAddCase pins one AddFlashTop row.
type flashTopAddCase struct {
	name      string
	caster    int
	source    int
	level     int
	currlevel int
	rnd       []int
	wantDam   int
}

type flashTopCase struct {
	name         string
	duration     int
	caster       int
	source       int
	wantDuration int
	wantShield   int
	wantProbes   int
	wantUnshield bool
	wantDel      bool
}

// flashTopShield records the invincibility writes of one tick.
type flashTopShield struct {
	sets   int
	unsets int
}

// TestAddFlashTop pins the damage branches, the pre-flag and the 19-tick
// duration with a strict draw tape (level + 1 draws).
func TestAddFlashTop(t *testing.T) {
	// given
	cases := flashTopAddCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			fx := newAddFixture(t, addScript{
				caster: tc.caster, source: tc.source, level: tc.level,
				currlevel: tc.currlevel, dam: 55, rnd: tc.rnd,
			})

			// when
			missiles.AddFlashTop(fx.world, fx.missile, fx.param)
			fx.tape.assertConsumed()

			// then
			assertFlashTopAdd(t, fx.missile, tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// TestProcessFlashTop pins the shield write bracket, the three-tile probe sweep
// and the delete tail.
func TestProcessFlashTop(t *testing.T) {
	// given
	cases := flashTopCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			w, shield, probes := flashTopWorld(tc)

			// when
			m := flashTopMissile(tc)
			missiles.ProcessFlashTop(w, &m)

			// then
			assertFlashTopTick(t, m, tc)
			if probes() != tc.wantProbes {
				t.Fatalf("collision probes = %d, want %d", probes(), tc.wantProbes)
			}

			assertFlashTopShield(t, shield, tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// TestProcessFlashTopDamageType pins the row's damage type: the corpus flags
// FlashTop as "Magic" (misdat.tsv:13), so the collision probe must carry
// DamageMagic rather than DamageTypeFor's physical default.
func TestProcessFlashTopDamageType(t *testing.T) {
	// given
	var got missiles.DamageType

	w := &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: missiles.NoopProcessCallbacks(),
	}
	w.DMonster[3][3] = 1 // the North probe tile of a flash at (4, 4)
	w.HitCallbacks.MonsterTargeted = func(int16, *missiles.Direction) bool { return true }
	w.HitCallbacks.MonsterMHit = func(_, _, _ int, _ missiles.MissileID,
		damageType missiles.DamageType, _ bool, _ missiles.KeepOnCollision,
	) bool {
		got = damageType

		return false
	}
	m := missiles.Missile{Type: missiles.FlashTop, Duration: 3, Dam: 12}
	m.SetTraveledFromTile(4, 4)

	// when
	missiles.ProcessFlashTop(w, &m)

	// then
	if got != missiles.DamageMagic { // misdat.tsv:13, missiles.cpp:3462-3466
		t.Fatalf("probe damage type = %v, want DamageMagic", got)
	}
}

// flashTopWorld wires the shield sink and counts the collision probes.
func flashTopWorld(_ flashTopCase) (*missiles.HitWorld, *flashTopShield, func() int) {
	shield := &flashTopShield{}
	probes := 0
	w := &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: missiles.NoopProcessCallbacks(),
	}
	w.ProcessCallbacks.SetInvincible = func(_ int, invincible bool) {
		if invincible {
			shield.sets++
		} else {
			shield.unsets++
		}
	}
	// The three probed tiles of a flash at (4, 4): North, NorthEast, East.
	for _, tile := range [][2]int{{3, 3}, {4, 3}, {5, 3}} {
		w.DMonster[tile[1]][tile[0]] = 1
	}

	w.HitCallbacks.MonsterTargeted = func(int16, *missiles.Direction) bool {
		probes++

		return false
	}
	return w, shield, func() int { return probes }
}

// flashTopMissile seeds the flash at tile (4, 4).
func flashTopMissile(tc flashTopCase) missiles.Missile {
	m := missiles.Missile{
		Type: missiles.FlashTop, Duration: tc.duration, Caster: tc.caster,
		Source: tc.source, Dam: 12,
	}
	m.SetTraveledFromTile(4, 4)

	return m
}

// assertFlashTopAdd checks the damage, pre-flag and duration of one add.
func assertFlashTopAdd(t *testing.T, m *missiles.Missile, tc flashTopAddCase) {
	t.Helper()

	if m.Dam != tc.wantDam { // cpp:2143-2148
		t.Fatalf("Dam = %d, want %d", m.Dam, tc.wantDam)
	}

	if !m.PreFlag { // cpp:2150
		t.Fatal("PreFlag = false, want true")
	}

	if m.Duration != 19 { // cpp:2151
		t.Fatalf("Duration = %d, want 19", m.Duration)
	}
}

// assertFlashTopTick checks the duration and delete state of one tick.
func assertFlashTopTick(t *testing.T, m missiles.Missile, tc flashTopCase) {
	t.Helper()

	if m.Duration == tc.wantDuration && m.DelFlag == tc.wantDel {
		return
	}

	t.Fatalf("duration/del = %d/%v, want %d/%v", m.Duration, m.DelFlag, tc.wantDuration, tc.wantDel)
}

// assertFlashTopShield checks the shield writes of one tick.
func assertFlashTopShield(t *testing.T, shield *flashTopShield, tc flashTopCase) {
	t.Helper()

	if shield.sets != tc.wantShield {
		t.Fatalf("invincible sets = %d, want %d", shield.sets, tc.wantShield)
	}

	if (shield.unsets != 0) != tc.wantUnshield {
		t.Fatalf("invincible unsets = %d, want unshield %v", shield.unsets, tc.wantUnshield)
	}
}
