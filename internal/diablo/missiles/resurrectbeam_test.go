// Package missiles_test pins AddResurrectBeam (missiles.cpp:2749-2754) and
// ProcessResurrectBeam (missiles.cpp:4164-4170).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// resurrectBeamCases pin the ticks: the beam counts down to 0 and deletes on
// the tick that reaches it, always republishing to the missile layer.
var resurrectBeamCases = []resurrectBeamCase{
	{name: "first tick counts down", duration: 12, wantDuration: 11, wantDel: false},
	{name: "last tick deletes", duration: 1, wantDuration: 0, wantDel: true},
}

// resurrectBeamCase pins one ProcessResurrectBeam tick.
type resurrectBeamCase struct {
	name         string
	duration     int
	wantDuration int
	wantDel      bool
}

// TestAddResurrectBeam pins the beam on its destination tile and the duration
// taken from the row's frame-0 animation length (cpp:2751-2753).
func TestAddResurrectBeam(t *testing.T) {
	// given
	w := resurrectBeamWorld()
	m := missiles.Missile{Type: missiles.ResurrectBeam, AnimLengths: []int{12, 11}}
	p := &missiles.AddParameter{Dst: missiles.TilePosition{X: 41, Y: 27}}

	// when
	missiles.AddResurrectBeam(w, &m, p)

	// then
	if m.Position.Tile != p.Dst || m.Position.Start != p.Dst {
		t.Fatalf("tile/start = %v/%v, want %v", m.Position.Tile, m.Position.Start, p.Dst)
	}

	if m.Duration != 12 { // frameLength[0] of the Resurrect graphic (cpp:2753)
		t.Fatalf("Duration = %d, want 12", m.Duration)
	}
}

// TestProcessResurrectBeam pins the countdown and the delete tail.
func TestProcessResurrectBeam(t *testing.T) {
	// given
	cases := resurrectBeamCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			w := resurrectBeamWorld()
			m := missiles.Missile{Type: missiles.ResurrectBeam, Duration: tc.duration}
			m.SetTraveledFromTile(4, 4)

			// when
			missiles.ProcessResurrectBeam(w, &m)

			// then
			if m.Duration != tc.wantDuration || m.DelFlag != tc.wantDel {
				t.Fatalf("duration/del = %d/%v, want %d/%v",
					m.Duration, m.DelFlag, tc.wantDuration, tc.wantDel)
			}
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// resurrectBeamWorld builds the no-callback world the beam needs.
func resurrectBeamWorld() *missiles.HitWorld {
	return &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: missiles.NoopProcessCallbacks(),
	}
}
