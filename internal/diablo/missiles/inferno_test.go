// Package missiles_test pins AddInferno (missiles.cpp:2464-2486) and
// ProcessInferno (cpp:3335-3371).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// infernoCases pin the ticks: the duration and var2 both count down
// (cpp:3337-3338), a killing hit refunds the duration (cpp:3342-3343), var2
// reaching 0 parks the frame at 20 (cpp:3344-3345), the light follows the
// frame with the 24-frame mirror (cpp:3346-3351) and only the spent inferno is
// republished (cpp:3359-3360).
var infernoCases = []infernoCase{
	{
		name: "burning tick stays invisible", duration: 30, var2: 6, frame: 4,
		wantDuration: 29, wantVar2: 5, wantFrame: 4,
	},
	{
		name: "var2 hitting zero parks the frame at 20", duration: 21, var2: 1, frame: 4,
		// The parked frame 20 mirrors to 24 - 20 = 4 (cpp:3347-3349).
		wantDuration: 20, wantFrame: 20, wantLight: 4, wantPut: false,
	},
	{
		name: "mirrored frame dims the light", duration: 20, var2: -1, frame: 16,
		wantDuration: 19, wantVar2: -2, wantFrame: 16, wantLight: 8, wantPut: true,
	},
	{
		name: "spent inferno deletes and unlights", duration: 1, var2: -3, frame: 2,
		wantDuration: 0, wantVar2: -4, wantFrame: 2, wantLight: 2, wantDel: true,
		wantUnlight: true, wantPut: false,
	},
}

// infernoCase pins one ProcessInferno tick.
type infernoCase struct {
	name         string
	duration     int
	var2         int
	frame        int
	wantDuration int
	wantVar2     int
	wantFrame    int
	wantLight    uint8
	wantUnlight  bool
	wantDel      bool
	wantPut      bool
}

// infernoSpies records the light sink of one tick.
type infernoSpies struct {
	radius   uint8
	changed  int
	unlights int
}

// TestAddInferno pins the monster-cast damage roll and the duration/lifetime
// wiring of the flame jet (cpp:2464-2486).
func TestAddInferno(t *testing.T) {
	// given
	tape := &rndTape{t: t, tape: []int{2, 1}}
	w := infernoWorld(&infernoSpies{}, tape, 10)
	m := missiles.Missile{
		Type: missiles.Inferno, Dam: 4, Caster: int(missiles.TargetMonsters), Source: 0,
	}
	p := &missiles.AddParameter{Dst: missiles.TilePosition{X: 6, Y: 6}}

	// when
	missiles.AddInferno(w, &m, p)
	tape.assertConsumed()

	// then
	if m.Var2 != 20 || m.Duration != 40 { // cpp:2466, 2471
		t.Fatalf("var2/duration = %d/%d, want 20/40", m.Var2, m.Duration)
	}

	if m.Position.Start != p.Dst { // cpp:2467
		t.Fatalf("start = %v, want the aim tile %v", m.Position.Start, p.Dst)
	}

	if m.Dam != 60 { // (8*3+16) + (8*3+16)/2 = 40 + 20, cpp:2475-2477
		t.Fatalf("Dam = %d, want 60", m.Dam)
	}
}

// TestProcessInferno pins the countdown, the frame park, the mirrored light
// ramp and the delete tail.
func TestProcessInferno(t *testing.T) {
	// given
	cases := infernoCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			spies := &infernoSpies{}
			w := infernoWorld(spies, &rndTape{t: t}, 10)

			// when
			m := missiles.Missile{
				Type: missiles.Inferno, Duration: tc.duration, Var2: tc.var2,
				AnimFrame: tc.frame, Lid: 9, Dam: 5, Source: 0,
			}
			m.SetTraveledFromTile(4, 4)
			missiles.ProcessInferno(w, &m)

			// then
			assertInfernoTick(t, m, spies, tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// infernoWorld wires the light sink, the draw tape and the stats reads.
func infernoWorld(spies *infernoSpies, tape *rndTape, level int) *missiles.HitWorld {
	cb := missiles.NoopProcessCallbacks()
	cb.GenerateRnd = tape.draw
	cb.ChangeLight = func(_, _, _ int, radius uint8) {
		spies.changed++
		spies.radius = radius
	}
	cb.AddUnLight = func(int) { spies.unlights++ }
	cb.AddLight = func(int, int, uint8) int { return 9 }

	return &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: cb,
		AddSource: missiles.AddSource{
			PlayerStats: func(int) missiles.AddPlayerStats { return missiles.AddPlayerStats{Level: level} },
			MonsterStats: func(int) missiles.AddMonsterStats {
				return missiles.AddMonsterStats{MinDamage: 3, MaxDamage: 5}
			},
		},
	}
}

// assertInfernoTick checks the state and the light calls of one tick.
func assertInfernoTick(t *testing.T, m missiles.Missile, spies *infernoSpies, tc infernoCase) {
	t.Helper()

	if m.Duration != tc.wantDuration || m.Var2 != tc.wantVar2 || m.AnimFrame != tc.wantFrame {
		t.Fatalf("duration/var2/frame = %d/%d/%d, want %d/%d/%d",
			m.Duration, m.Var2, m.AnimFrame, tc.wantDuration, tc.wantVar2, tc.wantFrame)
	}

	if m.DelFlag != tc.wantDel {
		t.Fatalf("DelFlag = %v, want %v", m.DelFlag, tc.wantDel)
	}

	if (spies.unlights != 0) != tc.wantUnlight {
		t.Fatalf("AddUnLight calls = %d, want unlight %v", spies.unlights, tc.wantUnlight)
	}

	if tc.wantLight != 0 && spies.radius != tc.wantLight {
		t.Fatalf("ChangeLight radius = %d, want %d", spies.radius, tc.wantLight)
	}
}
