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
