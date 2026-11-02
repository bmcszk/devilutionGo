// Package missiles_test: the registry gate for the Hellfire rows — each wired
// id must resolve to the ported Add* body of its hf-misdat.tsv row, never the
// unported placeholder or an absent entry (registry.go).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
)

// hellfireWiredRows is the slice-3c wiring ledger: every row whose Add* body
// is ported.
var hellfireWiredRows = []hellfireWiredRow{
	{
		id:   missiles.OpenNest,
		name: "OpenNest",
		build: func(t *testing.T) *addFixture {
			t.Helper()

			return newAddFixture(t, addScript{start: [2]int{40, 40}, dst: [2]int{43, 40}, source: 2, dam: 9})
		},
		check: func(t *testing.T, m *missiles.Missile) {
			t.Helper()

			if !m.DelFlag {
				t.Fatal("nest row not deleted")
			}
		},
	},
	{
		id:   missiles.HorkSpawn,
		name: "HorkSpawn",
		build: func(t *testing.T) *addFixture {
			t.Helper()

			return newAddFixture(t, addScript{
				start: [2]int{40, 40}, dst: [2]int{43, 40}, source: 2, midir: missiles.South,
			})
		},
		check: func(t *testing.T, m *missiles.Missile) {
			t.Helper()

			if m.Duration != 9 || m.Var1 != int(missiles.South) {
				t.Fatalf("duration/var1 = %d/%d, want 9/%d", m.Duration, m.Var1, int(missiles.South))
			}
		},
	},
	{
		id:   missiles.Jester,
		name: "Jester",
		build: func(t *testing.T) *addFixture {
			t.Helper()

			return newAddFixture(t, addScript{
				start: [2]int{40, 40}, dst: [2]int{43, 40}, source: 2, rnd: []int{4},
			})
		},
		check: func(t *testing.T, m *missiles.Missile) {
			t.Helper()

			if !m.DelFlag {
				t.Fatal("jester row not deleted")
			}
		},
	},
	{
		id:   missiles.LightningWall,
		name: "LightningWall",
		build: func(t *testing.T) *addFixture {
			t.Helper()

			return newAddFixture(t, addScript{
				start: [2]int{40, 40}, dst: [2]int{43, 40}, source: -1, spllvl: 1, rnd: []int{0},
			})
		},
		check: func(t *testing.T, m *missiles.Missile) {
			t.Helper()

			if m.Duration != 510 || m.Var1 != 40 || m.Var2 != 40 {
				t.Fatalf("duration/var1/var2 = %d/%d/%d, want 510/40/40", m.Duration, m.Var1, m.Var2)
			}
		},
	},
	{
		id:   missiles.Search,
		name: "Search",
		build: func(t *testing.T) *addFixture {
			t.Helper()

			return newAddFixture(t, addScript{
				start: [2]int{40, 40}, dst: [2]int{43, 40}, source: 0, level: 3,
			})
		},
		check: func(t *testing.T, m *missiles.Missile) {
			t.Helper()

			if m.Duration != 251 {
				t.Fatalf("duration = %d, want 6 + 245", m.Duration)
			}
		},
	},
	{
		id:   missiles.Mana,
		name: "Mana",
		build: func(t *testing.T) *addFixture {
			t.Helper()

			return newAddFixture(t, addScript{
				start: [2]int{40, 40}, dst: [2]int{43, 40}, source: 0,
				level: 1, spllvl: 1, rnd: []int{0, 0, 0},
			})
		},
		check: func(t *testing.T, m *missiles.Missile) {
			t.Helper()

			if !m.DelFlag {
				t.Fatal("manapotion row not deleted")
			}
		},
	},
	{
		id:   missiles.Magi,
		name: "Magi",
		build: func(t *testing.T) *addFixture {
			t.Helper()

			return newAddFixture(t, addScript{start: [2]int{40, 40}, dst: [2]int{43, 40}, source: 0})
		},
		check: func(t *testing.T, m *missiles.Missile) {
			t.Helper()

			if !m.DelFlag {
				t.Fatal("fullmanapotion row not deleted")
			}
		},
	},
	{
		id:   missiles.Reflect,
		name: "Reflect",
		build: func(t *testing.T) *addFixture {
			t.Helper()

			return newAddFixture(t, addScript{
				start: [2]int{40, 40}, dst: [2]int{43, 40}, source: 0, spllvl: 2, level: 3,
			})
		},
		check: func(t *testing.T, m *missiles.Missile) {
			t.Helper()

			if !m.DelFlag {
				t.Fatal("reflect row not deleted")
			}
		},
	},
	{
		id:   missiles.StealMana,
		name: "StealMana",
		build: func(t *testing.T) *addFixture {
			t.Helper()

			return newAddFixture(t, addScript{start: [2]int{40, 40}, dst: [2]int{43, 40}, source: -1})
		},
		check: func(t *testing.T, m *missiles.Missile) {
			t.Helper()

			if !m.DelFlag {
				t.Fatal("stealmana row not deleted")
			}
		},
	},
	{
		id:   missiles.SpectralArrow,
		name: "SpectralArrow",
		build: func(t *testing.T) *addFixture {
			t.Helper()

			return newAddFixture(t, addScript{
				start: [2]int{40, 40}, dst: [2]int{42, 41}, source: 0,
				class: missiles.HeroClassRogue, level: 9,
			})
		},
		check: func(t *testing.T, m *missiles.Missile) {
			t.Helper()

			if m.Var3 != 2 || m.Duration != 1 {
				t.Fatalf("var3/duration = %d/%d, want 2/1", m.Var3, m.Duration)
			}
		},
	},
	{
		id:   missiles.FireballBow,
		name: "FireballBow",
		build: func(t *testing.T) *addFixture {
			t.Helper()

			return newAddFixture(t, addScript{
				start: [2]int{40, 40}, dst: [2]int{43, 40}, source: 0, light: 9,
			})
		},
		check: func(t *testing.T, m *missiles.Missile) {
			t.Helper()

			if m.Duration != 256 || m.Lid != 9 {
				t.Fatalf("duration/lid = %d/%d, want 256/9", m.Duration, m.Lid)
			}
		},
	},
	{
		id:   missiles.LightningBow,
		name: "LightningBow",
		build: func(t *testing.T) *addFixture {
			t.Helper()

			return newAddFixture(t, addScript{
				start: [2]int{40, 40}, dst: [2]int{43, 40}, source: -1, dam: 2, rnd: []int{0},
			})
		},
		check: func(t *testing.T, m *missiles.Missile) {
			t.Helper()

			if m.Duration != 255 || m.Dam != 2<<6 {
				t.Fatalf("duration/dam = %d/%d, want 255/%d", m.Duration, m.Dam, 2<<6)
			}
		},
	},
	{
		id:   missiles.ChargedBoltBow,
		name: "ChargedBoltBow",
		build: func(t *testing.T) *addFixture {
			t.Helper()

			return newAddFixture(t, addScript{
				start: [2]int{40, 40}, dst: [2]int{43, 40}, source: 0,
				caster: int(missiles.TargetPlayers), dam: 7, rnd: []int{4, 2}, light: 9,
			})
		},
		check: func(t *testing.T, m *missiles.Missile) {
			t.Helper()

			if m.Dam != 15 || m.Duration != 256 || m.Mirnd != 5 {
				t.Fatalf("dam/duration/mirnd = %d/%d/%d, want 15/256/5", m.Dam, m.Duration, m.Mirnd)
			}
		},
	},
	{
		id:   missiles.RuneOfImmolation,
		name: "RuneOfImmolation",
		build: func(t *testing.T) *addFixture {
			t.Helper()

			return newRuneFixture(t, runeCase{
				script: addScript{start: [2]int{30, 30}, dst: [2]int{33, 30}, source: 0, light: 7},
			})
		},
		check: func(t *testing.T, m *missiles.Missile) {
			t.Helper()

			if m.Var1 != int(missiles.Immolation) {
				t.Fatalf("var1 = %d, want %d (Immolation)", m.Var1, int(missiles.Immolation))
			}
		},
	},
	{
		id:   missiles.RuneOfStone,
		name: "RuneOfStone",
		build: func(t *testing.T) *addFixture {
			t.Helper()

			return newRuneFixture(t, runeCase{
				script: addScript{start: [2]int{30, 30}, dst: [2]int{33, 30}, source: 0, light: 7},
			})
		},
		check: func(t *testing.T, m *missiles.Missile) {
			t.Helper()

			if m.Var1 != int(missiles.StoneCurse) {
				t.Fatalf("var1 = %d, want %d (StoneCurse)", m.Var1, int(missiles.StoneCurse))
			}
		},
	},
	{
		id:   missiles.BigExplosion,
		name: "BigExplosion",
		build: func(t *testing.T) *addFixture {
			t.Helper()

			return newAddFixture(t, addScript{
				start: [2]int{40, 40}, dst: [2]int{41, 40}, source: 0,
				rnd: []int{1, 1}, animLen: 12, light: 9,
			})
		},
		check: func(t *testing.T, m *missiles.Missile) {
			t.Helper()

			if m.Dam != 8 || m.Duration != 11 {
				t.Fatalf("dam/duration = %d/%d, want 8/11", m.Dam, m.Duration)
			}
		},
	},
	{
		id:   missiles.RingOfFire,
		name: "RingOfFire",
		build: func(t *testing.T) *addFixture {
			t.Helper()

			return newAddFixture(t, addScript{start: [2]int{42, 44}, dst: [2]int{43, 44}})
		},
		check: func(t *testing.T, m *missiles.Missile) {
			t.Helper()

			if m.Duration != 7 || m.Var1 != 42 {
				t.Fatalf("duration/var1 = %d/%d, want 7/42", m.Duration, m.Var1)
			}
		},
	},
	{
		id:   missiles.RuneOfFire,
		name: "RuneOfFire",
		build: func(t *testing.T) *addFixture {
			t.Helper()

			return newRuneFixture(t, runeCase{
				script: addScript{start: [2]int{30, 30}, dst: [2]int{33, 30}, source: 0, light: 7},
			})
		},
		check: func(t *testing.T, m *missiles.Missile) {
			t.Helper()

			if m.Var1 != int(missiles.BigExplosion) {
				t.Fatalf("var1 = %d, want %d (BigExplosion)", m.Var1, int(missiles.BigExplosion))
			}
		},
	},
	{
		id:   missiles.RuneOfLight,
		name: "RuneOfLight",
		build: func(t *testing.T) *addFixture {
			t.Helper()

			return newRuneFixture(t, runeCase{
				script: addScript{
					start: [2]int{30, 30}, dst: [2]int{33, 30}, source: -1, light: 7,
					rnd: []int{1, 1},
				},
			})
		},
		check: func(t *testing.T, m *missiles.Missile) {
			t.Helper()

			if m.Var1 != int(missiles.LightningWall) || m.Dam != 64 {
				t.Fatalf("var1/dam = %d/%d, want %d/64", m.Var1, m.Dam, int(missiles.LightningWall))
			}
		},
	},
	{
		id:   missiles.RuneOfNova,
		name: "RuneOfNova",
		build: func(t *testing.T) *addFixture {
			t.Helper()

			return newRuneFixture(t, runeCase{
				script: addScript{start: [2]int{30, 30}, dst: [2]int{33, 30}, source: 0, light: 7},
			})
		},
		check: func(t *testing.T, m *missiles.Missile) {
			t.Helper()

			if m.Var1 != int(missiles.Nova) {
				t.Fatalf("var1 = %d, want %d (Nova)", m.Var1, int(missiles.Nova))
			}
		},
	},
}

// hellfireWiredRow is one wired Hellfire row and the observable its add body
// leaves behind.
type hellfireWiredRow struct {
	id    missiles.MissileID
	name  string
	build func(t *testing.T) *addFixture
	// check asserts the value the row's add body must leave behind.
	check func(t *testing.T, m *missiles.Missile)
}

func TestHellfireRegistryWiresPortedAddBodies(t *testing.T) {
	// given
	rows := hellfireWiredRows

	// when
	subtests := make([]subtest, 0, len(rows))
	for _, row := range rows {
		subtests = append(subtests, subtest{row.name, func(t *testing.T) {
			t.Helper()
			runHellfireWiredRow(t, row)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// runHellfireWiredRow dispatches one wired row through the registry and checks
// the observable its add body leaves.
func runHellfireWiredRow(t *testing.T, row hellfireWiredRow) {
	t.Helper()

	// given
	behavior, ok := missiles.BehaviorFor(row.id)
	if !ok {
		t.Fatalf("no behavior registered for %s (%d)", row.name, row.id)
	}

	fx := row.build(t)

	// when
	behavior.Add(fx.world, fx.missile, fx.param)
	fx.tape.assertConsumed()

	// then
	row.check(t, fx.missile)
}
