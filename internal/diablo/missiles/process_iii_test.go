// Package missiles_test: devilGo-mwlm slice 3a-beta acceptance-test helpers,
// split from process_test.go to stay inside the revive file-length limit
// (Source/missiles.cpp:552-663 stop-missile and ProcessArrow tick tapes).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

func moveMissileStopCase(t *testing.T, tt moveMissileStopRow) {
	t.Helper()
	// given
	w := procWorld()
	m := freshMissile(64, 64)
	m.Position.Velocity = tt.vel
	mt := newMoveTape(w)
	mt.probeIn = func(missiles.TilePosition) bool { return false }

	// when
	changed := missiles.MoveMissile(w, &m, mt.checkTile, missiles.DontMoveOnFailedProbe(tt.dm))

	// then
	if !changed {
		t.Fatal("moveMissile changed = false, want true")
	}

	if len(mt.probes) != len(tt.wantProbes) || mt.probes[0] != tt.wantProbes[0] {
		t.Fatalf("probes = %v, want %v", mt.probes, tt.wantProbes)
	}

	if m.Position.Tile != tt.wantTile {
		t.Errorf("tile = %v, want %v", m.Position.Tile, tt.wantTile)
	}

	if m.Position.Traveled != tt.wantTrav {
		t.Errorf("traveled = %v, want %v", m.Position.Traveled, tt.wantTrav)
	}

	if m.Position.Offset != tt.wantOff {
		t.Errorf("offset = %v, want %v", m.Position.Offset, tt.wantOff)
	}

	if m.Position.Velocity != tt.wantVel {
		t.Errorf("velocity = %v, want %v", m.Position.Velocity, tt.wantVel)
	}
}

func arrowTick1(t *testing.T, spies *colSpies, w *missiles.HitWorld, m missiles.Missile) {
	t.Helper()
	// then
	if m.Duration != 2 || m.Dist != 1 || m.DelFlag {
		t.Fatalf("tick 1: duration = %d dist = %d delFlag = %v, want 2 1 false",
			m.Duration, m.Dist, m.DelFlag)
	}

	roll := spies.monsterMHitSeen[0]

	wantRoll := rollCall{
		minDamage: arrowMinDam, maxDamage: arrowMaxDam, dist: 1,
		missileType: missiles.Arrow, damageType: missiles.DamagePhysical,
	}
	if roll != wantRoll {
		t.Fatalf("tick 1: roll = %+v, want %+v", roll, wantRoll)
	}

	if w.Grid.DFlags[63][65]&world.FlagMissile == 0 {
		t.Fatal("tick 1: FlagMissile missing at (65, 63)")
	}
}

func arrowTick2(t *testing.T, spies *colSpies, m missiles.Missile) {
	t.Helper()

	if m.Duration != 1 || m.Dist != 2 || m.DelFlag {
		t.Fatalf("tick 2: duration = %d dist = %d delFlag = %v, want 1 2 false",
			m.Duration, m.Dist, m.DelFlag)
	}

	if rolls := len(spies.monsterMHitSeen); rolls != 1 {
		t.Fatalf("tick 2: MonsterMHit rolls = %d, want 1 (hash guard)", rolls)
	}
}

func arrowTick3(t *testing.T, spies *colSpies, w *missiles.HitWorld, m missiles.Missile) {
	t.Helper()

	if m.Duration != 0 || m.Dist != 3 || !m.DelFlag {
		t.Fatalf("tick 3: duration = %d dist = %d delFlag = %v, want 0 3 true",
			m.Duration, m.Dist, m.DelFlag)
	}

	roll := spies.monsterMHitSeen[1]

	wantRoll := rollCall{
		minDamage: arrowMinDam, maxDamage: arrowMaxDam, dist: 3,
		missileType: missiles.Arrow, damageType: missiles.DamagePhysical,
	}
	if roll != wantRoll {
		t.Fatalf("tick 3: roll = %+v, want %+v", roll, wantRoll)
	}

	if w.Grid.DFlags[62][66]&world.FlagMissile != 0 {
		t.Fatal("tick 3: deleted missile registered in dFlags")
	}
}

func projTapeTick1(t *testing.T, light *lightSpies, w *missiles.HitWorld, m missiles.Missile) {
	t.Helper()

	if m.Duration != 1 {
		t.Fatalf("tick 1: duration = %d, want 1", m.Duration)
	}

	if m.Var1 != 66 || m.Var2 != 64 {
		t.Fatalf("tick 1: var = (%d, %d), want (66, 64)", m.Var1, m.Var2)
	}

	expectChangeLight(t, light.changes, changeLightCall{lid: 7, x: 66, y: 64, radius: 8}, "lid 7 at (66, 64) radius 8")
	expectFlagMissileAt(t, w.Grid, 66, 64)

	if len(light.unlights) != 0 || len(light.spawnCalls) != 0 {
		t.Fatalf("tick 1: unlights = %v spawns = %v, want none", light.unlights, light.spawnCalls)
	}
}

func projTapeTick2(t *testing.T, light *lightSpies, w *missiles.HitWorld, m *missiles.Missile) {
	t.Helper()

	if m.Duration != 0 || !m.DelFlag {
		t.Fatalf("tick 2: duration = %d delFlag = %v, want 0 true", m.Duration, m.DelFlag)
	}

	if len(light.spawnCalls) != 1 {
		t.Fatalf("tick 2: spawns = %d, want 1", len(light.spawnCalls))
	}

	expectExplosionSpawnCall(t, light, m)

	if len(light.unlights) != 1 || light.unlights[0] != 7 {
		t.Fatalf("tick 2: unlights = %v, want [7]", light.unlights)
	}

	if w.Grid.DFlags[64][68]&world.FlagMissile != 0 {
		t.Fatal("tick 2: deleted missile registered in dFlags")
	}

	if len(light.changes) != 1 {
		t.Fatalf("tick 2: ChangeLight calls = %d, want 1 (alive branch only)", len(light.changes))
	}
}

func expectExplosionSpawnCall(t *testing.T, light *lightSpies, parent *missiles.Missile) {
	t.Helper()

	call := light.spawnCalls[0]
	if call.src != (missiles.TilePosition{X: 68, Y: 64}) ||
		call.dst != (missiles.TilePosition{}) ||
		call.dir != missiles.South ||
		call.spawnID != missiles.MagmaBallExplosion ||
		call.caster != int(missiles.TargetMonsters) || call.source != 0 ||
		call.parent != parent {
		t.Fatalf("tick 2: spawn = %+v, want MagmaBallExplosion at (68, 64) "+
			"dir South parent self", call)
	}
}

func projSpawnSwitchCase(t *testing.T, tt projSpawnSwitchRow) {
	t.Helper()
	// given
	spies := newColSpies() // monster-target gate open (collision_test.go)
	w := &missiles.HitWorld{Grid: world.NewGrid(), HitCallbacks: spies.callbacks()}
	light := &lightSpies{}

	restore := light.wire(w)
	defer restore()

	m := freshMissile(64, 64)
	m.Type = tt.typ
	m.Caster = int(missiles.TargetMonsters)
	m.Source = 2
	m.Duration = 1
	m.Lid = missiles.NoLight // no light plumbing on this path
	m.Position.Velocity = missiles.PositionDisplacement{DeltaX: axisVX, DeltaY: axisVY}

	// when
	missiles.ProcessGenericProjectile(w, &m)

	// then
	if !m.DelFlag {
		t.Fatal("DelFlag not set")
	}

	expectExplosionSpawn(t, light, tt.wantExplosion)
}

func expectExplosionSpawn(t *testing.T, light *lightSpies, wantExplosion missiles.MissileID) {
	t.Helper()

	if len(light.spawnCalls) != 1 {
		t.Fatalf("spawns = %d, want 1", len(light.spawnCalls))
	}

	if got := light.spawnCalls[0].spawnID; got != wantExplosion {
		t.Fatalf("spawnID = %v, want %v", got, wantExplosion)
	}

	if len(light.unlights) != 0 || len(light.changes) != 0 {
		t.Fatalf("unlights = %v changes = %v, want none with NoLight",
			light.unlights, light.changes)
	}
	// ChangeLightOffset's own NO_LIGHT guard (lighting.cpp:332)
	// makes the missiles.UpdateMissilePos feed harmless — the hook fires,
	// but never with a real light id.
	for _, off := range light.offsets {
		if off.lid != missiles.NoLight {
			t.Fatalf("ChangeLightOffset lid = %d, want NoLight", off.lid)
		}
	}
}

// freshMissile seeds the upstream AddMissile state (cpp:2823-2827):
// tile = start = src, traveled/velocity/offset zero, no light, hash 0.
// (Slice-1's SetTraveledFromTile is the loadsave reload shape, not the
// spawn shape — AddMissile leaves traveled zeroed.)
func freshMissile(x, y int) missiles.Missile {
	var m missiles.Missile

	m.Position.Tile = missiles.TilePosition{X: x, Y: y}
	m.Position.Start = m.Position.Tile

	return m
}

func procWorld() *missiles.HitWorld {
	return &missiles.HitWorld{
		Grid:             world.NewGrid(),
		ProcessCallbacks: missiles.NoopProcessCallbacks(),
	}
}

// wire installs the recording hooks on the world's process sink; the
// returned func restores the previous callbacks (defer in each test).
func (s *lightSpies) wire(w *missiles.HitWorld) func() {
	prev := w.ProcessCallbacks
	cb := missiles.NoopProcessCallbacks()
	cb.ChangeLightOffset = func(lid int, offsetX, offsetY int8) {
		s.offsets = append(s.offsets, lightOffsetCall{lid: lid, offsetX: int(offsetX), offsetY: int(offsetY)})
	}
	cb.ChangeLight = func(lid, x, y int, radius uint8) {
		s.changes = append(s.changes, changeLightCall{lid: lid, x: x, y: y, radius: uint8(radius)})
	}
	cb.AddUnLight = func(lid int) {
		s.unlights = append(s.unlights, lid)
	}
	cb.SpawnMissile = func(src, dst missiles.TilePosition, dir missiles.Direction,
		spawnID missiles.MissileID, caster, source int, parent *missiles.Missile,
	) {
		s.spawnCalls = append(s.spawnCalls, spawnCall{src, dst, dir, spawnID, caster, source, parent})
	}
	cb.SourceDamage = func(_ *missiles.Missile) (int, int) {
		s.sourceDam = append(s.sourceDam, true)
		return arrowMinDam, arrowMaxDam
	}
	w.ProcessCallbacks = cb

	return func() { w.ProcessCallbacks = prev }
}

func expectFlagMissileAt(t *testing.T, g *world.Grid, x, y int) {
	t.Helper()

	if g.DFlags[y][x]&world.FlagMissile == 0 {
		t.Fatalf("FlagMissile missing at (%d, %d)", x, y)
	}
}
