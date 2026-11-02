// Package missiles_test pins the Hellfire bow add bodies (missiles.cpp:
// 1479-1505, 1620-1653, 1716-1731).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
)

// spectralArrowCases is the class/item bonus table (missiles.cpp:1483-1499).
var spectralArrowCases = []bowCase{
	{
		"rogue gains a point per four levels",
		addScript{class: missiles.HeroClassRogue, level: 9},
		func(t *testing.T, fx *addFixture) {
			t.Helper()

			if fx.missile.Var3 != 2 {
				t.Fatalf("var3 = %d, want 2", fx.missile.Var3)
			}
		},
	},
	{
		"warrior gains a point per eight levels",
		addScript{class: missiles.HeroClassWarrior, level: 17},
		func(t *testing.T, fx *addFixture) {
			t.Helper()

			if fx.missile.Var3 != 2 {
				t.Fatalf("var3 = %d, want 2", fx.missile.Var3)
			}
		},
	},
	{
		"bard uses the warrior step",
		addScript{class: missiles.HeroClassBard, level: 9},
		func(t *testing.T, fx *addFixture) {
			t.Helper()

			if fx.missile.Var3 != 1 {
				t.Fatalf("var3 = %d, want 1", fx.missile.Var3)
			}
		},
	},
	{
		"sorcerer gains no class step",
		addScript{class: missiles.HeroClassSorcerer, level: 30},
		func(t *testing.T, fx *addFixture) {
			t.Helper()

			if fx.missile.Var3 != 0 {
				t.Fatalf("var3 = %d, want 0", fx.missile.Var3)
			}
		},
	},
	{
		"the four fast-attack bits stack",
		addScript{
			class: missiles.HeroClassRogue, level: 5,
			itemFlags: []int{
				missiles.ItemSpecialEffectQuickAttack,
				missiles.ItemSpecialEffectFastAttack,
				missiles.ItemSpecialEffectFasterAttack,
				missiles.ItemSpecialEffectFastestAttack,
			},
		},
		func(t *testing.T, fx *addFixture) {
			t.Helper()

			if fx.missile.Var3 != 16 {
				t.Fatalf("var3 = %d, want 16", fx.missile.Var3)
			}
		},
	},
}

// immolationCases is the speed table of AddImmolation (missiles.cpp:1626-1629).
var immolationCases = []bowCase{
	{
		"a player caster adds one speed per spell level",
		addScript{caster: int(missiles.TargetMonsters), spllvl: 10},
		func(t *testing.T, fx *addFixture) {
			t.Helper()

			want := missiles.PositionDisplacement{DeltaX: 1204840, DeltaY: 602420} // 26 px
			if fx.missile.Position.Velocity != want {
				t.Fatalf("velocity = %v, want %v", fx.missile.Position.Velocity, want)
			}
		},
	},
	{
		"a monster caster keeps the flat 16",
		addScript{caster: int(missiles.TargetPlayers), spllvl: 5},
		func(t *testing.T, fx *addFixture) {
			t.Helper()

			want := missiles.PositionDisplacement{DeltaX: 741440, DeltaY: 370720} // 16 px
			if fx.missile.Position.Velocity != want {
				t.Fatalf("velocity = %v, want %v", fx.missile.Position.Velocity, want)
			}
		},
	},
	{
		"the spell-level bonus caps at 34",
		addScript{caster: int(missiles.TargetMonsters), spllvl: 40},
		func(t *testing.T, fx *addFixture) {
			t.Helper()

			want := missiles.PositionDisplacement{DeltaX: 2317000, DeltaY: 1158500} // 50 px
			if fx.missile.Position.Velocity != want {
				t.Fatalf("velocity = %v, want %v", fx.missile.Position.Velocity, want)
			}
		},
	},
}

// lightningBowCases is the caster-tile table of AddLightningBow
// (missiles.cpp:1644-1651).
var lightningBowCases = []bowCase{
	{
		"a trap stores its launch tile",
		addScript{source: -1, rnd: []int{4}},
		func(t *testing.T, fx *addFixture) {
			t.Helper()

			if fx.missile.Var1 != 40 || fx.missile.Var2 != 40 {
				t.Fatalf("var1/var2 = %d/%d, want the launch tile 40/40", fx.missile.Var1, fx.missile.Var2)
			}
		},
	},
	{
		"a player stores its own tile",
		addScript{source: 3, tile: [2]int{12, 34}, rnd: []int{4}},
		func(t *testing.T, fx *addFixture) {
			t.Helper()

			if fx.missile.Var1 != 12 || fx.missile.Var2 != 34 {
				t.Fatalf("var1/var2 = %d/%d, want the player tile 12/34", fx.missile.Var1, fx.missile.Var2)
			}
		},
	},
}

// chargedBoltBowCases is the damage table of AddChargedBoltBow
// (missiles.cpp:1719-1721).
var chargedBoltBowCases = []bowCase{
	{
		"a player keeps its rolled damage",
		addScript{caster: int(missiles.TargetMonsters), dam: 7, rnd: []int{9, 3}},
		func(t *testing.T, fx *addFixture) {
			t.Helper()

			if fx.missile.Dam != 7 {
				t.Fatalf("dam = %d, want 7", fx.missile.Dam)
			}
		},
	},
	{
		"a monster gets the flat 15",
		addScript{caster: int(missiles.TargetPlayers), dam: 7, rnd: []int{9, 3}},
		func(t *testing.T, fx *addFixture) {
			t.Helper()

			if fx.missile.Dam != 15 {
				t.Fatalf("dam = %d, want 15", fx.missile.Dam)
			}
		},
	},
}

// bowCase is one scripted bow-row run: the add body's inputs plus the
// assertion on what it left behind.
type bowCase struct {
	name   string
	script addScript
	check  func(t *testing.T, fx *addFixture)
}

func TestAddSpectralArrow_ClassAndItemBonus(t *testing.T) {
	// given
	cases := spectralArrowCases

	// when
	subtests := bowSubtests(missiles.AddSpectralArrow, cases)

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

func TestAddImmolation_PlayerAndMonsterSpeeds(t *testing.T) {
	// given
	cases := immolationCases

	// when
	subtests := bowSubtests(missiles.AddImmolation, cases)

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

func TestAddLightningBow_SeedsTheCasterTile(t *testing.T) {
	// given
	cases := lightningBowCases

	// when
	subtests := bowSubtests(missiles.AddLightningBow, cases)

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

func TestAddChargedBoltBow_FlatDamageForNonPlayerCasters(t *testing.T) {
	// given
	cases := chargedBoltBowCases

	// when
	subtests := bowSubtests(missiles.AddChargedBoltBow, cases)

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

func TestAddChargedBoltBow_SeedsItsBoltCountAndLight(t *testing.T) {
	// given
	script := addScript{
		start: [2]int{40, 40}, dst: [2]int{43, 40}, source: 0,
		caster: int(missiles.TargetPlayers), dam: 7, rnd: []int{9, 3},
		midir: missiles.North, light: 9,
	}
	fx := newAddFixture(t, script)

	// when
	missiles.AddChargedBoltBow(fx.world, fx.missile, fx.param)
	fx.tape.assertConsumed()

	// then
	if fx.missile.Mirnd != 10 {
		t.Fatalf("mirnd = %d, want 10 (GenerateRnd(15)+1)", fx.missile.Mirnd)
	}

	if fx.missile.Var1 != 5 || fx.missile.Var2 != int(missiles.North) || fx.missile.Duration != 256 {
		t.Fatalf("var1/var2/duration = %d/%d/%d, want 5/%d/256",
			fx.missile.Var1, fx.missile.Var2, fx.missile.Duration, int(missiles.North))
	}

	want := missiles.PositionDisplacement{DeltaX: 370720, DeltaY: 185360} // 8 px
	if fx.missile.Position.Velocity != want {
		t.Fatalf("velocity = %v, want %v", fx.missile.Position.Velocity, want)
	}

	if fx.missile.Lid != 9 || fx.light.radius != 5 {
		t.Fatalf("lid/radius = %d/%d, want 9/5", fx.missile.Lid, fx.light.radius)
	}
}

func TestAddLightningBow_KeepsItsOwnDurationAndScale(t *testing.T) {
	// given
	cases := lightningBowCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for i, tc := range cases {
		script := tc.script
		script.start = [2]int{40, 40}
		script.dst = [2]int{43, 40}
		script.dam = 2
		script.rnd = []int{i}
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			fx := newAddFixture(t, script)

			// when
			missiles.AddLightningBow(fx.world, fx.missile, fx.param)
			fx.tape.assertConsumed()

			// then
			if fx.missile.Duration != 255 || fx.missile.Dam != 2<<6 || fx.missile.AnimFrame != i+1 {
				t.Fatalf("duration/dam/frame = %d/%d/%d, want 255/%d/%d",
					fx.missile.Duration, fx.missile.Dam, fx.missile.AnimFrame, 2<<6, i+1)
			}
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// bowSubtests scripts one bow row per case and dispatches it.
func bowSubtests(add func(*missiles.HitWorld, *missiles.Missile, *missiles.AddParameter), cases []bowCase) []subtest {
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()
			runBowCase(t, add, tc)
		}})
	}

	return subtests
}

// runBowCase runs one scripted bow row and applies the case's assertion.
func runBowCase(t *testing.T, add func(*missiles.HitWorld, *missiles.Missile, *missiles.AddParameter), tc bowCase) {
	t.Helper()

	// given
	script := tc.script
	script.start = [2]int{40, 40}
	script.dst = [2]int{43, 40}
	if script.light == 0 {
		script.light = 9
	}

	fx := newAddFixture(t, script)

	// when
	add(fx.world, fx.missile, fx.param)
	fx.tape.assertConsumed()

	// then
	tc.check(t, fx)
}
