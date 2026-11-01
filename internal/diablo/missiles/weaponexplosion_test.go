// Package missiles_test pins AddWeaponExplosion (missiles.cpp:2892-2907) and
// ProcessWeaponExplosion (cpp:3573-3600).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// weaponExplosionAddCases pin the two flavours: dst.x == 1 is the fire
// explosion (Fire damage, MagmaBallExplosion sheet, cpp:2894-2904), anything
// else is the lightning one (ChargedBolt sheet), and a zero damage skips the
// immediate hit (cpp:2896).
// weaponExplosionCases pin the light ramp of the ten-frame curve
// (missiles.cpp:3575): the first tick raises the light, the following ticks
// dim it along ExpLight[var1], and the last tick deletes and unlights.
var weaponExplosionCases = []weaponExplosionCase{
	{
		name: "first tick raises the light", duration: 9, var1: 0,
		wantDuration: 8, wantVar1: 1, wantLight: 9, wantAddLight: 1, wantPut: true,
	},
	{
		name: "second tick dims along the curve", duration: 9, var1: 1,
		wantDuration: 8, wantVar1: 2, wantLight: 10, wantChanges: 1, wantPut: true,
	},
	{
		name: "last tick deletes and unlights", duration: 1, var1: 4,
		wantDuration: 0, wantVar1: 5, wantDel: true, wantUnlight: true,
	},
}

var weaponExplosionAddCases = []weaponExplosionAddCase{
	{
		name: "fire explosion hits as fire", dstX: 1, dam: 12,
		wantType: missiles.DamageFire, wantGraphic: missiles.GraphicMagmaBallExplosion, wantProbes: 1,
	},
	{
		name: "lightning explosion hits as lightning", dstX: 0, dam: 7,
		wantType: missiles.DamageLightning, wantGraphic: missiles.GraphicChargedBolt, wantProbes: 1,
	},
	{
		name: "zero damage skips the hit", dstX: 1, dam: 0,
		wantType: missiles.DamageFire, wantGraphic: missiles.GraphicMagmaBallExplosion,
	},
}

// weaponExplosionAddCase pins one AddWeaponExplosion row.
type weaponExplosionAddCase struct {
	name        string
	dstX        int
	dam         int
	wantType    missiles.DamageType
	wantGraphic missiles.GraphicID
	wantProbes  int
}

// weaponExplosionCase pins one ProcessWeaponExplosion tick.
type weaponExplosionCase struct {
	name         string
	duration     int
	var1         int
	wantDuration int
	wantVar1     int
	wantLight    uint8
	wantAddLight int
	wantChanges  int
	wantUnlight  bool
	wantDel      bool
	wantPut      bool
}

// weaponExplosionSpies records the light sink of one tick.
type weaponExplosionSpies struct {
	radius   uint8
	adds     int
	changes  int
	unlights int
}

// TestAddWeaponExplosion pins the flavour switch, the immediate hit and the
// AnimLen - 1 life.
func TestAddWeaponExplosion(t *testing.T) {
	// given
	cases := weaponExplosionAddCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			w, probes := weaponExplosionAddFixture(t)
			m := missiles.Missile{Type: missiles.WeaponExplosion, Dam: tc.dam, AnimLen: 10}
			m.SetTraveledFromTile(4, 4)

			// when
			missiles.AddWeaponExplosion(w, &m, &missiles.AddParameter{
				Dst: missiles.TilePosition{X: tc.dstX, Y: 0},
			})

			// then
			assertWeaponExplosionAdd(t, m, probes(), tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// TestProcessWeaponExplosion pins the light curve and the delete tail.
func TestProcessWeaponExplosion(t *testing.T) {
	// given
	cases := weaponExplosionCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			spies := &weaponExplosionSpies{}
			w := weaponExplosionWorld(spies, &rndTape{t: t})
			m := missiles.Missile{
				Type: missiles.WeaponExplosion, Duration: tc.duration, Var1: tc.var1, Lid: 7,
			}
			m.SetTraveledFromTile(4, 4)

			// when
			missiles.ProcessWeaponExplosion(w, &m)

			// then
			assertWeaponExplosionTick(t, m, spies, tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// weaponExplosionAddFixture wires the world and counts the collision probes.
func weaponExplosionAddFixture(t *testing.T) (*missiles.HitWorld, func() int) {
	t.Helper()

	probes := 0
	w := weaponExplosionWorld(&weaponExplosionSpies{}, &rndTape{t: t})
	w.HitCallbacks.MonsterTargeted = func(int16, *missiles.Direction) bool {
		probes++

		return false
	}
	w.DMonster[4][4] = 1

	return w, func() int { return probes }
}

// assertWeaponExplosionAdd checks the sheet, the AnimLen - 1 life and the
// immediate hit.
func assertWeaponExplosionAdd(t *testing.T, m missiles.Missile, probes int, tc weaponExplosionAddCase) {
	t.Helper()

	if m.AnimType != int(tc.wantGraphic) { // cpp:2900-2903
		t.Fatalf("AnimType = %d, want %d", m.AnimType, tc.wantGraphic)
	}

	if m.Duration != 9 { // cpp:2904: AnimLen - 1
		t.Fatalf("Duration = %d, want 9", m.Duration)
	}

	if probes != tc.wantProbes {
		t.Fatalf("collision probes = %d, want %d", probes, tc.wantProbes)
	}
}

// assertWeaponExplosionCounters checks the countdown and the light calls.
func assertWeaponExplosionCounters(t *testing.T, m missiles.Missile, spies *weaponExplosionSpies,
	tc weaponExplosionCase,
) {
	t.Helper()

	if m.Duration != tc.wantDuration || m.Var1 != tc.wantVar1 { // cpp:3577-3588
		t.Fatalf("duration/var1 = %d/%d, want %d/%d", m.Duration, m.Var1, tc.wantDuration, tc.wantVar1)
	}

	assertWeaponExplosionLight(t, spies, tc)

	if m.DelFlag != tc.wantDel || (spies.unlights != 0) != tc.wantUnlight { // cpp:3589-3595
		t.Fatalf("del/unlight = %v/%d, want %v/%v", m.DelFlag, spies.unlights, tc.wantDel, tc.wantUnlight)
	}
}

// assertWeaponExplosionLight checks the light call counts and the radius of the
// last one.
func assertWeaponExplosionLight(t *testing.T, spies *weaponExplosionSpies, tc weaponExplosionCase) {
	t.Helper()

	if spies.adds != tc.wantAddLight || spies.changes != tc.wantChanges {
		t.Fatalf("AddLight/ChangeLight = %d/%d, want %d/%d",
			spies.adds, spies.changes, tc.wantAddLight, tc.wantChanges)
	}

	if tc.wantLight != 0 && spies.radius != tc.wantLight {
		t.Fatalf("light radius = %d, want %d", spies.radius, tc.wantLight)
	}
}

// weaponExplosionWorld wires the light sink and the tape.
func weaponExplosionWorld(spies *weaponExplosionSpies, tape *rndTape) *missiles.HitWorld {
	cb := missiles.NoopProcessCallbacks()
	cb.GenerateRnd = tape.draw
	cb.AddLight = func(_, _ int, radius uint8) int {
		spies.adds++
		spies.radius = radius

		return 7
	}
	cb.ChangeLight = func(_, _, _ int, radius uint8) {
		spies.changes++
		spies.radius = radius
	}
	cb.AddUnLight = func(int) { spies.unlights++ }

	return &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: cb,
	}
}

// assertWeaponExplosionTick checks the light curve bookkeeping of one tick.
func assertWeaponExplosionTick(t *testing.T, m missiles.Missile, spies *weaponExplosionSpies,
	tc weaponExplosionCase,
) {
	t.Helper()

	assertWeaponExplosionCounters(t, m, spies, tc)
}
