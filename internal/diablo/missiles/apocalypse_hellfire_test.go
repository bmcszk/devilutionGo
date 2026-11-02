// Package missiles_test pins the Hellfire line gate of ProcessApocalypse
// (missiles.cpp:3861-3862).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// apocalypseSpawns records the boom calls of one box tick.
type apocalypseSpawns struct {
	dst []missiles.TilePosition
}

// apocalypseWorldConfig selects the gate flags of the line fixture.
type apocalypseWorldConfig struct {
	hellfire  bool
	blockLine bool
}

// TestProcessApocalypseHellfireBlocksLine pins the gbIsHellfire gate: in
// Hellfire a monster behind a missile-blocking tile is skipped, the box walk
// exhausts and the missile dies (missiles.cpp:3861-3862, 3872).
func TestProcessApocalypseHellfireBlocksLine(t *testing.T) {
	// given
	spawns := &apocalypseSpawns{}
	w := apocalypseLineWorld(apocalypseWorldConfig{hellfire: true, blockLine: true}, spawns)
	m := apocalypseLineMissile()

	// when
	missiles.ProcessApocalypse(w, &m)

	// then
	if len(spawns.dst) != 0 {
		t.Fatalf("spawns = %v, want none", spawns.dst)
	}

	if m.Var2 != 12 || m.Var4 != 12 {
		t.Fatalf("cursor = %d/%d, want 12/12", m.Var2, m.Var4)
	}

	if !m.DelFlag {
		t.Fatalf("del = %v, want true", m.DelFlag)
	}
}

// TestProcessApocalypseBaseGameIgnoresLine pins the missing-Hellfire arm: the
// base game has no line gate, so the same monster takes the boom
// (missiles.cpp:3861).
func TestProcessApocalypseBaseGameIgnoresLine(t *testing.T) {
	// given
	spawns := &apocalypseSpawns{}
	w := apocalypseLineWorld(apocalypseWorldConfig{blockLine: true}, spawns)
	m := apocalypseLineMissile()

	// when
	missiles.ProcessApocalypse(w, &m)

	// then
	if len(spawns.dst) != 1 {
		t.Fatalf("spawns = %v, want one", spawns.dst)
	}

	if spawns.dst[0] != (missiles.TilePosition{X: 13, Y: 12}) {
		t.Fatalf("spawn dst = %v, want 13/12", spawns.dst[0])
	}

	if m.Var2 != 12 || m.Var4 != 14 {
		t.Fatalf("cursor = %d/%d, want 12/14", m.Var2, m.Var4)
	}

	if m.DelFlag {
		t.Fatalf("del = %v, want false", m.DelFlag)
	}
}

// TestProcessApocalypseHellfireFiresOnClearLine pins that the Hellfire gate
// only skips obstructed tiles: a clear line still fires (missiles.cpp:3861).
func TestProcessApocalypseHellfireFiresOnClearLine(t *testing.T) {
	// given
	spawns := &apocalypseSpawns{}
	w := apocalypseLineWorld(apocalypseWorldConfig{hellfire: true}, spawns)
	m := apocalypseLineMissile()

	// when
	missiles.ProcessApocalypse(w, &m)

	// then
	if len(spawns.dst) != 1 {
		t.Fatalf("spawns = %v, want one", spawns.dst)
	}

	if spawns.dst[0] != (missiles.TilePosition{X: 13, Y: 12}) {
		t.Fatalf("spawn dst = %v, want 13/12", spawns.dst[0])
	}

	if m.Var2 != 12 || m.Var4 != 14 {
		t.Fatalf("cursor = %d/%d, want 12/14", m.Var2, m.Var4)
	}
}

// apocalypseLineMissile is a first-tick Apocalypse box around (10,12).
func apocalypseLineMissile() missiles.Missile {
	m := missiles.Missile{
		Type: missiles.Apocalypse, Source: 0, Dam: 9, Duration: 255,
		Var1: 8, Var2: 12, Var3: 28, Var4: 12, Var5: 28, Var6: 12,
	}
	m.Position.Tile = missiles.TilePosition{X: 10, Y: 12}

	return m
}

// apocalypseLineWorld wires a box holding one monster at (13,12) and, when the
// config asks, a missile-blocking tile on the line from (10,12) to it.
func apocalypseLineWorld(cfg apocalypseWorldConfig, spawns *apocalypseSpawns) *missiles.HitWorld {
	cb := missiles.NoopProcessCallbacks()
	cb.IsHellfire = func() bool { return cfg.hellfire }
	cb.PlayerDirection = func(int) missiles.Direction { return missiles.East }
	cb.SpawnMissile = func(_, dst missiles.TilePosition, _ missiles.Direction,
		_ missiles.MissileID, _, _, _, _ int, _ *missiles.Missile,
	) {
		spawns.dst = append(spawns.dst, dst)
	}

	w := &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: cb,
	}
	w.DMonster[12][13] = 1
	if cfg.blockLine {
		w.Grid.DPiece[12][11] = 1378
		w.Grid.SOLData[1378] = world.TileBlockMissile
	}

	return w
}
