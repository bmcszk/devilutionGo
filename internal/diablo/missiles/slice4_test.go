// Package missiles_test aggregates the slice-4 rows: each one must be
// reachable through the registry dispatch (BehaviorFor) and its wired Add body
// must leave the row's own signature on the missile. This is the slice-level
// companion to the per-row tables.
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
)

// sliceFourRows lists the rows ported in slice 4 with the post-condition their
// Add body guarantees on a zeroed missile.
var sliceFourRows = []sliceFourRow{
	{
		name: "Lightning", id: missiles.Lightning,
		prep: func(m *missiles.Missile) { m.Caster, m.Source = 0, 0 },
		want: func(t *testing.T, m missiles.Missile) {
			t.Helper()

			if m.AnimFrame < 1 || m.AnimFrame > 8 { // missiles.cpp:2794-2796
				t.Fatalf("AnimFrame = %d, want the RandomIntBetween(1, 8) frame", m.AnimFrame)
			}
		},
	},
	{
		name: "FireWall", id: missiles.FireWall,
		want: func(t *testing.T, m missiles.Missile) {
			t.Helper()

			if m.Dam <= 0 || m.Duration <= 0 { // missiles.cpp:1956-1968
				t.Fatalf("dam/duration = %d/%d, want a rolled wall", m.Dam, m.Duration)
			}
		},
	},
	{
		name: "FlameWave", id: missiles.FlameWave,
		want: func(t *testing.T, m missiles.Missile) {
			t.Helper()

			if m.Duration != 255 { // missiles.cpp:2174
				t.Fatalf("Duration = %d, want 255", m.Duration)
			}
		},
	},
	{
		name: "FlashTop", id: missiles.FlashTop,
		prep: func(m *missiles.Missile) { m.Caster = 0 },
		want: func(t *testing.T, m missiles.Missile) {
			t.Helper()

			if m.Duration != 19 || !m.PreFlag { // missiles.cpp:2150-2151
				t.Fatalf("duration/preflag = %d/%v, want 19/true", m.Duration, m.PreFlag)
			}
		},
	},
	{
		name: "ApocalypseBoom", id: missiles.ApocalypseBoom,
		prep: func(m *missiles.Missile) { m.AnimLen = 12 },
		want: func(t *testing.T, m missiles.Missile) {
			t.Helper()

			if m.Duration != 12 { // missiles.cpp:2459
				t.Fatalf("Duration = %d, want the AnimLen 12", m.Duration)
			}
		},
	},
	{
		name: "Infravision", id: missiles.Infravision,
		want: func(t *testing.T, m missiles.Missile) {
			t.Helper()

			if m.Duration != 1584 { // missiles.cpp:2556
				t.Fatalf("Duration = %d, want 1584", m.Duration)
			}
		},
	},
	{
		name: "ResurrectBeam", id: missiles.ResurrectBeam,
		prep: func(m *missiles.Missile) { m.AnimLengths = []int{12} },
		want: func(t *testing.T, m missiles.Missile) {
			t.Helper()

			if m.Duration != 12 || m.Position.Tile != (missiles.TilePosition{X: 9, Y: 3}) {
				t.Fatalf("duration/tile = %d/%v, want 12/{9 3} (missiles.cpp:2751-2753)",
					m.Duration, m.Position.Tile)
			}
		},
	},
	{
		name: "AcidSplat", id: missiles.AcidSplat,
		prep: func(m *missiles.Missile) { m.AnimLen = 12 },
		param: func() *missiles.AddParameter {
			// The splat copies its parent exploder's position (cpp:2041-2044).
			return &missiles.AddParameter{Parent: &missiles.Missile{AnimLen: 12}}
		},
		want: func(t *testing.T, m missiles.Missile) {
			t.Helper()

			if m.Duration != 12 { // AddMissileExplosion, missiles.cpp:1611-1613
				t.Fatalf("Duration = %d, want the AnimLen 12", m.Duration)
			}
		},
	},
}

// sliceFourRow pins one slice-4 row's registry reachability and Add post-state.
type sliceFourRow struct {
	name  string
	id    missiles.MissileID
	prep  func(*missiles.Missile)
	param func() *missiles.AddParameter
	want  func(*testing.T, missiles.Missile)
}

// add runs the row's wired Add body on a seeded missile and returns the result.
func (row sliceFourRow) add(behavior missiles.Behavior) missiles.Missile {
	m := missiles.Missile{Type: row.id}
	if row.prep != nil {
		row.prep(&m)
	}

	param := &missiles.AddParameter{Dst: missiles.TilePosition{X: 9, Y: 3}}
	if row.param != nil {
		param = row.param()
	}

	behavior.Add(&m, param)

	return m
}

// TestSliceFourRowsKeepTheirAddSignature: every slice-4 row resolves through
// the registry and its Add body leaves its documented post-state, so a wiring
// regression (row falling back to the no-op stub) fails here first.
func TestSliceFourRowsKeepTheirAddSignature(t *testing.T) {
	// given
	cases := sliceFourRows

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, row := range cases {
		subtests = append(subtests, subtest{row.name, func(t *testing.T) {
			t.Helper()

			// given
			behavior := registryBehavior(t, row)

			// when
			m := row.add(behavior)

			// then
			row.want(t, m)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// registryBehavior resolves one row's dispatch or fails the test — a row
// missing from the registry is the wiring regression this test exists for.
func registryBehavior(t *testing.T, row sliceFourRow) missiles.Behavior {
	t.Helper()

	behavior, ok := missiles.BehaviorFor(row.id)
	if !ok {
		t.Fatalf("%s has no registry behavior", row.name)
	}

	return behavior
}
