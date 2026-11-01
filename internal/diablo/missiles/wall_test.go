// Package missiles_test pins AddWallControl (missiles.cpp:2350-2370),
// ProcessWallControl (cpp:3770-3845) and AddFlameWaveControl /
// ProcessFlameWaveControl (cpp:2559-2566, 3875-3905): the two control missiles
// that grow a Fire Wall / Flame Wave out of one tile per tick.
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// wallCases pin the ProcessWallControl ticks: the first tick seeds the centre
// wall and walks both arms out (cpp:3806-3816), later ticks grow both arms with
// TryGrowWall (cpp:3817-3828) and the arms stop independently (cpp:3830-3841).
var wallCases = []wallCase{
	{
		name: "the seed tick plants one wall and arms both sides", start: true,
		var1: 10, var2: 10, var5: 10, var6: 10, var3: 1, var4: 3,
		wantVar1: 10, wantVar2: 11, wantVar5: 9, wantVar6: 10, wantPlaced: 1,
	},
	{
		name: "grown arms plant one segment each", var1: 11, var2: 10, var5: 10, var6: 11, var3: 1, var4: 3,
		wantVar1: 11, wantVar2: 11, wantVar5: 9, wantVar6: 11, wantPlaced: 2,
	},
	{
		name: "a stopped right arm leaves its cursor alone", var1: 11, var2: 10, var5: 10, var6: 11, var3: 1, var4: 3,
		var7: 1, wantVar1: 11, wantVar2: 11, wantVar5: 10, wantVar6: 11, wantPlaced: 1,
	},
	{name: "the last tick deletes the control", duration: 1, wantDel: true},
}

// wallCase pins one ProcessWallControl tick.
type wallCase struct {
	name       string
	start      bool
	duration   int
	var1       int
	var2       int
	var3       int
	var4       int
	var5       int
	var6       int
	var7       int
	wantVar1   int
	wantVar2   int
	wantVar5   int
	wantVar6   int
	wantPlaced int
	wantDel    bool
}

// wallSpy records the wall spawn calls.
type wallSpy struct {
	placed []missiles.TilePosition
}

// TestAddWallControl pins the spread search, the two arm directions and the
// 7-tick life (cpp:2350-2370).
func TestAddWallControl(t *testing.T) {
	// given
	w, _ := wallWorld(&rndTape{t: t}, &wallSpy{})
	m := missiles.Missile{Type: missiles.FireWallControl, Source: 0}
	m.SetTraveledFromTile(10, 10)
	m.Position.Start = missiles.TilePosition{X: 10, Y: 10}

	// when
	missiles.AddWallControl(w, &m, &missiles.AddParameter{
		Dst: missiles.TilePosition{X: 10, Y: 10}, Midir: missiles.South,
	})

	// then
	// the aim tile is the launch tile, so the search takes the first free
	// neighbour it finds, one tile south
	if m.Var1 != 10 || m.Var2 != 11 || m.Var5 != 10 || m.Var6 != 11 {
		t.Fatalf("spread = %d/%d %d/%d, want 10/11 10/11", m.Var1, m.Var2, m.Var5, m.Var6)
	}

	// Left(Left(South)) = East (6), Right(Right(South)) = West (2)
	if m.Var3 != 6 || m.Var4 != 2 || m.Duration != 7 || m.DelFlag {
		t.Fatalf("var3/var4/duration/del = %d/%d/%d/%v, want 6/2/7/false", m.Var3, m.Var4, m.Duration, m.DelFlag)
	}
}

// TestProcessWallControl pins the wall growth.
func TestProcessWallControl(t *testing.T) {
	// when
	subtests := make([]subtest, 0, len(wallCases))
	for _, tc := range wallCases {
		subtests = append(subtests, subtest{tc.name, runWallTick(tc)})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// TestAddFlameWaveControl pins the aim tile, the 1-tick life and frame 4
// (cpp:2559-2566).
func TestAddFlameWaveControl(t *testing.T) {
	// given
	w, _ := wallWorld(&rndTape{t: t}, &wallSpy{})
	m := missiles.Missile{Type: missiles.FlameWaveControl, Source: 0}

	// when
	missiles.AddFlameWaveControl(w, &m, &missiles.AddParameter{Dst: missiles.TilePosition{X: 14, Y: 9}})

	// then
	if m.Var1 != 14 || m.Var2 != 9 || m.Duration != 1 || m.AnimFrame != 4 {
		t.Fatalf("vars = %d/%d duration %d frame %d, want 14/9 1 4",
			m.Var1, m.Var2, m.Duration, m.AnimFrame)
	}
}

// TestProcessFlameWaveControl pins the centre wall plus the
// (spellLevel / 2) + 2 segments grown to each side (cpp:3875-3900).
func TestProcessFlameWaveControl(t *testing.T) {
	// given
	spy := &wallSpy{}
	w, _ := wallWorld(&rndTape{t: t}, spy)
	m := missiles.Missile{
		Type: missiles.FlameWaveControl, Source: 0, SpellLevel: 2, Duration: 1,
		Var1: 10, Var2: 12,
	}
	m.SetTraveledFromTile(10, 10)
	m.Position.Tile = missiles.TilePosition{X: 10, Y: 10}

	// when
	missiles.ProcessFlameWaveControl(w, &m)

	// then
	// segmentsToAdd = 2/2 + 2 = 3 to each side of the tile 2 south of the caster
	if len(spy.placed) != 7 || spy.placed[0] != (missiles.TilePosition{X: 11, Y: 12}) {
		t.Fatalf("segments = %v, want the centre wall {11 12} plus 3 + 3", spy.placed)
	}

	if !m.DelFlag {
		t.Fatal("the 1-tick control must delete itself")
	}
}

// wallWorld wires the tape and a PlaceWall sink standing in for AddMissile.
func wallWorld(tape *rndTape, spy *wallSpy) (*missiles.HitWorld, *wallSpy) {
	blocked := map[missiles.TilePosition]bool{}
	cb := missiles.NoopProcessCallbacks()
	cb.GenerateRnd = tape.draw
	cb.PlaceWallMissile = func(src, dst missiles.TilePosition, _ missiles.Direction, _ missiles.MissileID,
		_, _, _, _ int,
	) *missiles.Missile {
		if blocked[dst] {
			return nil
		}

		spy.placed = append(spy.placed, dst)

		return &missiles.Missile{Position: missiles.Position{Tile: dst, Start: src}}
	}

	return &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: cb,
	}, spy
}

// runWallTick drives one ProcessWallControl tick.
func runWallTick(tc wallCase) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		// given
		w, spy := wallWorld(&rndTape{t: t}, &wallSpy{})
		m := wallControlMissile(tc)

		// when
		missiles.ProcessWallControl(w, &m)

		// then
		assertWallTick(t, m, spy, tc)
	}
}

// wallControlMissile is the control one tick starts from.
func wallControlMissile(tc wallCase) missiles.Missile {
	m := missiles.Missile{
		Type: missiles.FireWallControl, Source: 0, SpellLevel: 1, Duration: 6,
		Var1: tc.var1, Var2: tc.var2, Var3: tc.var3, Var4: tc.var4,
		Var5: tc.var5, Var6: tc.var6, Var7: tc.var7,
	}
	if tc.start {
		m.Var1, m.Var2, m.Var5, m.Var6 = 10, 10, 10, 10
	}

	if tc.duration != 0 {
		m.Duration = tc.duration
	}

	return m
}

// assertWallTick checks the delete, the segment count and the arm cursors.
func assertWallTick(t *testing.T, m missiles.Missile, spy *wallSpy, tc wallCase) {
	t.Helper()

	if m.DelFlag != tc.wantDel || len(spy.placed) != tc.wantPlaced {
		t.Fatalf("del/placed = %v/%d, want %v/%d", m.DelFlag, len(spy.placed), tc.wantDel, tc.wantPlaced)
	}

	if tc.wantDel {
		return
	}

	if m.Var1 != tc.wantVar1 || m.Var2 != tc.wantVar2 || m.Var5 != tc.wantVar5 || m.Var6 != tc.wantVar6 {
		t.Fatalf("arms = %d/%d %d/%d, want %d/%d %d/%d",
			m.Var1, m.Var2, m.Var5, m.Var6, tc.wantVar1, tc.wantVar2, tc.wantVar5, tc.wantVar6)
	}
}
