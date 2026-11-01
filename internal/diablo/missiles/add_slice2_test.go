// Package missiles_test pins the slice-2 wizard/elemental Add bodies:
// AddChargedBolt (missiles.cpp:2697-2713) and AddHolyBolt (:2715-2736).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
)

// chargedBoltCases pins the AddChargedBolt rows (missiles.cpp:2697-2713).
var chargedBoltCases = []chargedHolyCase{
	{
		name: "player caster rolls radius, damage and start frame",
		script: addScript{
			dst: [2]int{10, 10}, midir: missiles.South, caster: 0, source: 0,
			magic: 40, light: 42, rnd: []int{3, 5, 2},
		},
		wantVel: [2]int32{0, 262144}, wantDam: 6, wantLid: 42, wantMirnd: 4, wantFrame: 3, wantLightR: 5,
	},
	{
		name: "monster caster keeps the fixed 15 damage",
		script: addScript{
			dst: [2]int{10, 10}, midir: missiles.South,
			caster: int(missiles.TargetPlayers), source: 3,
			magic: 40, light: 42, rnd: []int{3, 2},
		},
		wantVel: [2]int32{0, 262144}, wantDam: 15, wantLid: 42, wantMirnd: 4, wantFrame: 3, wantLightR: 5,
	},
	{
		name: "same tile shot shifts by midir",
		script: addScript{
			start: [2]int{5, 5}, dst: [2]int{5, 5}, midir: missiles.South,
			caster: 0, source: 0, magic: 8, light: 7, rnd: []int{0, 1, 0},
		},
		wantVel: [2]int32{0, 262144}, wantDam: 2, wantLid: 7, wantMirnd: 1, wantFrame: 1, wantLightR: 5,
	},
	{
		name: "the seed frame group survives since the body never steers",
		script: addScript{
			dst: [2]int{10, 10}, midir: missiles.South, caster: 0, source: 0,
			magic: 40, light: 42, mfnum: 7, rnd: []int{3, 5, 2},
		},
		wantVel: [2]int32{0, 262144}, wantDam: 6, wantLid: 42, wantMirnd: 4, wantFrame: 3,
		wantMF: 7, wantLightR: 5,
	},
	{
		name: "trap casters still take the magic-quarter branch",
		script: addScript{
			dst: [2]int{10, 10}, midir: missiles.South, caster: 0, source: -1,
			magic: 8, light: 5, rnd: []int{3, 1, 2},
		},
		wantVel: [2]int32{0, 262144}, wantDam: 2, wantLid: 5, wantMirnd: 4, wantFrame: 3,
		wantLightR: 5,
	},
	{
		name: "magic below four collapses the damage range to one",
		script: addScript{
			dst: [2]int{10, 10}, midir: missiles.South, caster: 0, source: 0,
			magic: 3, light: 42, rnd: []int{3, 0, 2},
		},
		wantVel: [2]int32{0, 262144}, wantDam: 1, wantLid: 42, wantMirnd: 4, wantFrame: 3,
		wantLightR: 5,
	},
	{
		name: "pre-set damage is overwritten unlike the other adds",
		script: addScript{
			dst: [2]int{10, 10}, midir: missiles.South, caster: 0, source: 0,
			dam: 99, magic: 40, light: 42, rnd: []int{3, 5, 2},
		},
		wantVel: [2]int32{0, 262144}, wantDam: 6, wantLid: 42, wantMirnd: 4, wantFrame: 3, wantLightR: 5,
	},
}

// holyBoltCases pins the AddHolyBolt rows (missiles.cpp:2715-2736).
var holyBoltCases = []chargedHolyCase{
	{
		name: "level 1 wizard fires at speed 16",
		script: addScript{
			dst: [2]int{10, 10}, midir: missiles.South, caster: 0, source: 0,
			level: 3, light: 42, rnd: []int{4},
		},
		wantVel: [2]int32{0, 524288}, wantDam: 16, wantMF: 0, wantLid: 42, wantLightR: 8,
	},
	{
		name: "spell level 5 adds 10 to the speed",
		script: addScript{
			dst: [2]int{10, 10}, midir: missiles.South, caster: 0, source: 0,
			spllvl: 5, level: 3, light: 42, rnd: []int{4},
		},
		wantVel: [2]int32{0, 851968}, wantDam: 16, wantMF: 0, wantLid: 42, wantLightR: 8,
	},
	{
		name: "spell level 30 clamps the bonus at 47",
		script: addScript{
			dst: [2]int{10, 10}, midir: missiles.South, caster: 0, source: 0,
			spllvl: 30, level: 3, light: 42, rnd: []int{4},
		},
		wantVel: [2]int32{0, 2064384}, wantDam: 16, wantMF: 0, wantLid: 42, wantLightR: 8,
	},
	{
		name: "the body overwrites a seeded frame group with the 16-step one",
		script: addScript{
			dst: [2]int{5, 0}, midir: missiles.South, caster: 0, source: 0,
			level: 3, light: 42, mfnum: 7, rnd: []int{4},
		},
		wantVel: [2]int32{741440, 370720}, wantDam: 16, wantMF: 14, wantLid: 42, wantLightR: 8,
	},
	{
		name: "trap keeps the base speed despite the spell level",
		script: addScript{
			dst: [2]int{10, 10}, midir: missiles.South, caster: 0, source: -1,
			spllvl: 10, level: 0, light: 5, rnd: []int{0},
		},
		wantVel: [2]int32{0, 524288}, wantDam: 9, wantMF: 0, wantLid: 5, wantLightR: 8,
	},
	{
		name: "monster source reads the sink player slot for the level",
		script: addScript{
			dst: [2]int{10, 10}, midir: missiles.South,
			caster: int(missiles.TargetMonsters), source: 3,
			level: 2, light: 42, rnd: []int{4},
		},
		wantVel: [2]int32{0, 524288}, wantDam: 15, wantMF: 0, wantLid: 42, wantLightR: 8,
	},
	{
		name: "diagonal destination sets the 16-step frame group",
		script: addScript{
			dst: [2]int{5, 0}, midir: missiles.South, caster: 0, source: 0,
			level: 3, light: 42, rnd: []int{4},
		},
		wantVel: [2]int32{741440, 370720}, wantDam: 16, wantMF: 14, wantLid: 42, wantLightR: 8,
	},
}

// chargedHolyCase pins one wizard-spell initiation row.
type chargedHolyCase struct {
	name    string
	script  addScript
	wantVel [2]int32
	wantDam int
	wantMF  int
	wantLid int
	// wantLightR is the AddLight radius the body requests: 5 for ChargedBolt
	// (cpp:2707), 8 for HolyBolt (cpp:2733).
	wantLightR uint8
	// wantMirnd/wantFrame are the AddChargedBolt rolls (0 for AddHolyBolt rows).
	wantMirnd uint8
	wantFrame int
}

// TestAddChargedBolt pins AddChargedBolt against hand-computed C++ values:
// _mirnd = GenerateRnd(15)+1, the magic-quarter damage roll (fixed 15 for
// monster casters), the RandomIntBetween(1,8) start frame, light 5 and speed 8.
func TestAddChargedBolt(t *testing.T) {
	// given
	cases := chargedBoltCases

	// when

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			fx := newAddFixture(t, tc.script)

			// when
			missiles.AddChargedBolt(fx.world, fx.missile, fx.param)
			fx.tape.assertConsumed()

			// then
			assertChargedBolt(t, fx, tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// TestAddHolyBolt pins AddHolyBolt: speed 16 + min(spllvl*2, 47) (traps keep
// 16), the 16-step frame group, light 8 and the level damage roll.
func TestAddHolyBolt(t *testing.T) {
	// given
	cases := holyBoltCases

	// when

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			fx := newAddFixture(t, tc.script)

			// when
			missiles.AddHolyBolt(fx.world, fx.missile, fx.param)
			fx.tape.assertConsumed()

			// then
			assertHolyBolt(t, fx, tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// TestAddChargedBoltFeedsProcessSteering pins the add→process contract: the
// Mirnd the Add rolls (missiles.cpp:2700) and the midir it stores in var2
// (cpp:2711) are exactly what ProcessChargedBolt indexes into BPath and
// steers with (cpp:3995-4006). BPath[4] = 0 keeps the heading, BPath[5] = 1
// turns it right, so one tick with var3 == 0 must advance Mirnd 4 → 5 and
// aim the next tick to the right of South.
func TestAddChargedBoltFeedsProcessSteering(t *testing.T) {
	// given
	fx := newAddFixture(t, addScript{
		dst: [2]int{10, 10}, midir: missiles.South, caster: 0, source: 0,
		magic: 40, light: 42, rnd: []int{3, 5, 2},
	})
	missiles.AddChargedBolt(fx.world, fx.missile, fx.param)
	fx.tape.assertConsumed()

	// when
	missiles.ProcessChargedBolt(fx.world, fx.missile)

	// then
	m := fx.missile
	if m.Mirnd != 5 { // cpp:4003, (mirnd+1) & 0xF
		t.Fatalf("Mirnd = %d, want 5 after one steering tick", m.Mirnd)
	}

	if m.Var3 != 16 { // cpp:4005
		t.Fatalf("Var3 = %d, want 16", m.Var3)
	}

	if m.Var2 != int(missiles.South) { // BPath[4] = 0 keeps the heading
		t.Fatalf("heading = %d, want South (%d)", m.Var2, int(missiles.South))
	}
}

// assertChargedBolt checks the post-AddChargedBolt fields
// (missiles.cpp:2697-2713).
func assertChargedBolt(t *testing.T, fx *addFixture, tc chargedHolyCase) {
	t.Helper()

	assertChargedHolyMotion(t, fx, tc)

	m := fx.missile

	if m.Mirnd != tc.wantMirnd { // cpp:2700, GenerateRnd(15)+1
		t.Fatalf("Mirnd = %d, want %d", m.Mirnd, tc.wantMirnd)
	}

	if m.Var1 != 5 || m.Var2 != int(missiles.South) { // cpp:2712-2713
		t.Fatalf("var1/var2 = %d/%d, want 5/%d", m.Var1, m.Var2, int(missiles.South))
	}

	if m.AnimFrame != tc.wantFrame { // cpp:2706, RandomIntBetween(1, 8)
		t.Fatalf("AnimFrame = %d, want %d", m.AnimFrame, tc.wantFrame)
	}
}

// assertHolyBolt checks the post-AddHolyBolt fields (missiles.cpp:2715-2736).
func assertHolyBolt(t *testing.T, fx *addFixture, tc chargedHolyCase) {
	t.Helper()

	assertChargedHolyMotion(t, fx, tc)

	m := fx.missile

	if m.Var1 != 0 || m.Var2 != 0 { // cpp:2732-2733, start tile
		t.Fatalf("var1/var2 = %d/%d, want the start tile 0/0", m.Var1, m.Var2)
	}
}

// assertChargedHolyMotion checks the fields both wizard spells share.
func assertChargedHolyMotion(t *testing.T, fx *addFixture, tc chargedHolyCase) {
	t.Helper()

	m := fx.missile

	velocity := m.Position.Velocity
	if velocity.DeltaX != tc.wantVel[0] || velocity.DeltaY != tc.wantVel[1] {
		t.Fatalf("velocity = (%d, %d), want (%d, %d)",
			velocity.DeltaX, velocity.DeltaY, tc.wantVel[0], tc.wantVel[1])
	}

	if m.Dam != tc.wantDam {
		t.Fatalf("Dam = %d, want %d", m.Dam, tc.wantDam)
	}

	if m.Duration != 256 { // cpp:2712, 2731
		t.Fatalf("Duration = %d, want 256", m.Duration)
	}

	assertWizardLight(t, fx, tc)
}

// assertWizardLight checks the frame group and the AddLight request both
// wizard spells make on their start tile.
func assertWizardLight(t *testing.T, fx *addFixture, tc chargedHolyCase) {
	t.Helper()

	m := fx.missile

	if m.MFNum != tc.wantMF {
		t.Fatalf("frame group = %d, want %d", m.MFNum, tc.wantMF)
	}

	if m.Lid != tc.wantLid { // cpp:2707 (light 5), cpp:2733 (light 8)
		t.Fatalf("Lid = %d, want the scripted %d", m.Lid, tc.wantLid)
	}

	if fx.light.radius != tc.wantLightR {
		t.Fatalf("AddLight radius = %d, want %d", fx.light.radius, tc.wantLightR)
	}

	if fx.light.x != m.Position.Start.X || fx.light.y != m.Position.Start.Y {
		t.Fatalf("AddLight at (%d, %d), want the start tile (%d, %d)",
			fx.light.x, fx.light.y, m.Position.Start.X, m.Position.Start.Y)
	}
}
