// Package missiles_test pins AddInfernoControl (missiles.cpp:2685-2696) and
// ProcessInfernoControl (cpp:3961-3987).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// infernoControlCases pin the walk: each tick moves the control missile, and
// every tile it reaches that differs from the remembered one either forks an
// Inferno there (damage = var3, which then increments) or kills the walk on a
// blocking tile (cpp:3966-3985). Three forks end it (cpp:3986).
var infernoControlCases = []infernoControlCase{
	{
		name: "walk forks an inferno on the new tile", duration: 256, var1: 4, var2: 4,
		wantDuration: 255, wantVar1: 5, wantVar2: 5, wantVar3: 1, wantSpawns: 1,
	},
	{
		name: "same tile forks nothing", duration: 256, var1: 5, var2: 5,
		wantDuration: 255, wantVar1: 5, wantVar2: 5, wantSpawns: 0,
	},
	{
		name: "blocking tile ends the walk", duration: 256, var1: 4, var2: 4, blockTile: true,
		wantDuration: 0, wantVar1: 5, wantVar2: 5, wantVar3: 1, wantSpawns: 0, wantDel: true,
	},
	{
		name: "third fork deletes the control", duration: 256, var1: 4, var2: 4, var3: 2,
		wantDuration: 255, wantVar1: 5, wantVar2: 5, wantVar3: 3, wantSpawns: 1, wantDel: true,
	},
}

// infernoControlCase pins one ProcessInfernoControl tick.
type infernoControlCase struct {
	name         string
	duration     int
	var1         int
	var2         int
	var3         int
	blockTile    bool
	wantDuration int
	wantVar1     int
	wantVar2     int
	wantVar3     int
	wantSpawns   int
	wantDel      bool
}

// TestAddInfernoControl pins the remembered launch tile, the 32-pixel heading
// and the 256-tick life (cpp:2688-2695).
func TestAddInfernoControl(t *testing.T) {
	// given
	w := infernoControlWorld(&rndTape{t: t}, infernoControlCase{})
	m := missiles.Missile{Type: missiles.InfernoControl}
	m.SetTraveledFromTile(4, 4)
	m.Position.Start = missiles.TilePosition{X: 4, Y: 4}
	p := &missiles.AddParameter{Dst: missiles.TilePosition{X: 4, Y: 4}, Midir: missiles.South}

	// when
	missiles.AddInfernoControl(w, &m, p)

	// then
	if m.Var1 != 4 || m.Var2 != 4 { // cpp:2692-2693
		t.Fatalf("var1/var2 = %d/%d, want the launch tile 4/4", m.Var1, m.Var2)
	}

	if m.Duration != 256 { // cpp:2694
		t.Fatalf("Duration = %d, want 256", m.Duration)
	}

	if m.Position.Velocity == (missiles.PositionDisplacement{}) { // cpp:2691
		t.Fatal("velocity stays zero, want the 32-pixel heading")
	}
}

// TestProcessInfernoControl pins the walk, the fork and the delete tail.
func TestProcessInfernoControl(t *testing.T) {
	// given
	cases := infernoControlCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			calls, w := infernoControlFixture(t, tc)
			m := missiles.Missile{
				Type: missiles.InfernoControl, Duration: tc.duration, Var1: tc.var1, Var2: tc.var2,
				Var3: tc.var3, Source: 2, Caster: 1,
			}
			m.SetTraveledFromTile(4, 4)
			m.Position.Start = missiles.TilePosition{X: 4, Y: 4}
			m.Position.Velocity = missiles.PositionDisplacement{DeltaY: 32 << 16} // 32 pixels south

			// when
			missiles.ProcessInfernoControl(w, &m)

			// then
			assertInfernoControlTick(t, m, calls, tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// assertInfernoControlFork checks the fork count and, for a single fork, its
// type and damage (cpp:3968-3978).
func assertInfernoControlFork(t *testing.T, calls *chainSpawns, tc infernoControlCase) {
	t.Helper()

	if len(calls.calls) != tc.wantSpawns {
		t.Fatalf("spawns = %d, want %d", len(calls.calls), tc.wantSpawns)
	}

	if tc.wantSpawns != 1 {
		return
	}

	if spawn := calls.calls[0]; spawn.spawnID != missiles.Inferno || spawn.dam != tc.var3 {
		t.Fatalf("spawn = %v/%d, want Inferno/%d", spawn.spawnID, spawn.dam, tc.var3)
	}
}

// infernoControlWorld wires the grid and the tape.
func infernoControlWorld(tape *rndTape, tc infernoControlCase) *missiles.HitWorld {
	cb := missiles.NoopProcessCallbacks()
	cb.GenerateRnd = tape.draw

	w := &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: cb,
	}
	if tc.blockTile {
		w.Grid.DPiece[5][5] = 1378
		w.Grid.SOLData[1378] = world.TileBlockMissile
	}

	return w
}

// infernoControlFixture wires the world and the fork recorder.
func infernoControlFixture(t *testing.T, tc infernoControlCase) (*chainSpawns, *missiles.HitWorld) {
	t.Helper()

	calls := &chainSpawns{}
	w := infernoControlWorld(&rndTape{t: t}, tc)
	w.ProcessCallbacks.SpawnMissile = func(src, dst missiles.TilePosition, dir missiles.Direction,
		spawnID missiles.MissileID, caster, source, dam, spellLevel int, parent *missiles.Missile,
	) {
		calls.calls = append(calls.calls,
			spawnCall{src, dst, dir, spawnID, caster, source, dam, spellLevel, parent})
	}

	return calls, w
}

// assertInfernoControlTick checks the walk bookkeeping, the fork and the delete
// tail.
func assertInfernoControlTick(t *testing.T, m missiles.Missile, calls *chainSpawns, tc infernoControlCase) {
	t.Helper()

	if m.Duration != tc.wantDuration || m.DelFlag != tc.wantDel { // cpp:3963, 3986
		t.Fatalf("duration/del = %d/%v, want %d/%v", m.Duration, m.DelFlag, tc.wantDuration, tc.wantDel)
	}

	if m.Var1 != tc.wantVar1 || m.Var2 != tc.wantVar2 || m.Var3 != tc.wantVar3 { // cpp:3981-3983
		t.Fatalf("var1/var2/var3 = %d/%d/%d, want %d/%d/%d",
			m.Var1, m.Var2, m.Var3, tc.wantVar1, tc.wantVar2, tc.wantVar3)
	}

	assertInfernoControlFork(t, calls, tc)
}
