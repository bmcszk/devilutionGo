// Package missiles_test pins AddGuardian (missiles.cpp:2181-2227) and
// ProcessGuardian (cpp:3509-3578) with GuardianTryFireAt (cpp:709-733).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// guardianFixture is the world one guardian body runs against.
type guardianFixture struct {
	level      int
	spellLevel int
	blocked    bool
	monster    bool
}

// guardianTickCase pins one ProcessGuardian tick.
type guardianTickCase struct {
	name         string
	duration     int
	var1         int
	monster      bool
	wantSpawns   int
	wantGroup    int
	wantDel      bool
	wantUnlights int
}

// guardianLights records the light traffic of one guardian.
type guardianLights struct {
	added    int
	unlights int
	changes  int
}

// TestAddGuardian pins the placement search, the light and the clamped
// duration (cpp:2219-2226): duration = min(spell level + level/2, 30) << 4,
// floored at 30, var1 = duration - AnimLen, var3 = 1.
func TestAddGuardian(t *testing.T) {
	// given
	w, lights := guardianWorld(&rndTape{t: t}, guardianFixture{level: 10, spellLevel: 3})
	m := missiles.Missile{Type: missiles.Guardian, Source: 0, AnimLen: 8, SpellLevel: 3}
	m.SetTraveledFromTile(5, 5)
	m.Position.Start = missiles.TilePosition{X: 5, Y: 5}
	p := &missiles.AddParameter{Dst: missiles.TilePosition{X: 5, Y: 5}}

	// when
	missiles.AddGuardian(w, &m, p)

	// then
	if m.DelFlag || p.SpellFizzled {
		t.Fatalf("del/fizzled = %v/%v, want the guardian placed at the aim tile",
			m.DelFlag, p.SpellFizzled)
	}

	if m.Duration != 128 || m.Var1 != 120 || m.Var3 != 1 { // cpp:2219-2225
		t.Fatalf("duration/var1/var3 = %d/%d/%d, want 128/120/1", m.Duration, m.Var1, m.Var3)
	}

	if m.Position.Tile != (missiles.TilePosition{X: 5, Y: 5}) {
		t.Fatalf("tile = %v, want the aim tile", m.Position.Tile)
	}

	if lights.added != 1 { // cpp:2218
		t.Fatalf("lights added = %d, want 1", lights.added)
	}
}

// TestAddGuardianFizzlesWithoutATile: the search fails on a fully blocked
// neighbourhood and the spell reports itself fizzled (cpp:2208-2212).
func TestAddGuardianFizzlesWithoutATile(t *testing.T) {
	// given
	w, _ := guardianWorld(&rndTape{t: t}, guardianFixture{level: 1, spellLevel: 0, blocked: true})
	m := missiles.Missile{Type: missiles.Guardian, Source: 0, AnimLen: 8}
	m.SetTraveledFromTile(5, 5)
	m.Position.Start = missiles.TilePosition{X: 5, Y: 5}
	p := &missiles.AddParameter{Dst: missiles.TilePosition{X: 5, Y: 5}}

	// when
	missiles.AddGuardian(w, &m, p)

	// then
	if !m.DelFlag || !p.SpellFizzled {
		t.Fatalf("del/fizzled = %v/%v, want true/true", m.DelFlag, p.SpellFizzled)
	}
}

// TestProcessGuardian pins the idle/attack frame handover (cpp:3513-3517), the
// 16-tick arc sweep that fires a Firebolt at the first monster it sees
// (cpp:3522-3555) and the light ramp down to the delete (cpp:3558-3575).
func TestProcessGuardian(t *testing.T) {
	// given
	cases := []guardianTickCase{
		{
			name: "arc sweep fires at the monster in reach", duration: 113, var1: 120, monster: true,
			wantSpawns: 1, wantGroup: missiles.GuardianFrameAttack,
		},
		{
			name: "no target leaves the group alone", duration: 113, var1: 120,
			wantGroup: missiles.GuardianFrameIdle,
		},
		{
			name: "expiry deletes and unlights", duration: 1, var1: 120,
			wantDel: true, wantUnlights: 1, wantGroup: missiles.GuardianFrameIdle,
		},
	}

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			spawns, lights, w := guardianFixtureWorld(t, guardianFixture{level: 6, spellLevel: 2, monster: tc.monster})
			m := missiles.Missile{
				Type: missiles.Guardian, Source: 0, Duration: tc.duration, Var1: tc.var1,
				Var3: 1, Lid: 9, AnimLen: 8, SpellLevel: 2, MFNum: missiles.GuardianFrameIdle,
			}
			m.SetTraveledFromTile(5, 5)
			m.Position.Start = missiles.TilePosition{X: 5, Y: 5}
			m.Position.Tile = missiles.TilePosition{X: 5, Y: 5}

			// when
			missiles.ProcessGuardian(w, &m)

			// then
			assertGuardianTick(t, m, spawns, lights, tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// guardianFixtureWorld wires the grid, the light sinks and the fire recorder.
func guardianFixtureWorld(t *testing.T, fx guardianFixture) (*chainSpawns, *guardianLights, *missiles.HitWorld) {
	t.Helper()

	spawns := &chainSpawns{}
	w, lights := guardianWorld(&rndTape{t: t, tape: []int{7}}, fx)
	w.ProcessCallbacks.SpawnMissile = func(src, dst missiles.TilePosition, dir missiles.Direction,
		spawnID missiles.MissileID, caster, source, dam, spellLevel int, parent *missiles.Missile,
	) {
		spawns.calls = append(spawns.calls,
			spawnCall{src, dst, dir, spawnID, caster, source, dam, spellLevel, parent})
	}

	return spawns, lights, w
}

// guardianWorld wires the add-side sinks and the 5x5 neighbourhood.
func guardianWorld(tape *rndTape, fx guardianFixture) (*missiles.HitWorld, *guardianLights) {
	lights := &guardianLights{}
	cb := missiles.NoopProcessCallbacks()
	cb.GenerateRnd = tape.draw
	cb.AddLight = func(_, _ int, _ uint8) int {
		lights.added++

		return 9
	}
	cb.AddUnLight = func(int) { lights.unlights++ }
	cb.ChangeLight = func(_, _, _ int, _ uint8) { lights.changes++ }
	cb.MonsterIsPlayerMinion = func(int) bool { return false }
	cb.MonsterHasNoLife = func(int) bool { return false }
	cb.TileContainsMissile = func(missiles.TilePosition) bool { return false }
	cb.PlayerSpellLevel = func(int, int) int { return fx.spellLevel }

	w := &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: cb,
		AddSource: missiles.AddSource{
			PlayerStats: func(int) missiles.AddPlayerStats {
				return missiles.AddPlayerStats{Level: fx.level, HitPoints: 100 << 6}
			},
		},
	}
	if fx.monster { // one monster at the first arc offset the sweep reaches
		w.DMonster[5][5+6-6] = 1
		w.DMonster[5][11] = 1
	}
	if fx.blocked { // every neighbour of the aim tile blocks the search
		for y := range 12 {
			for x := range 12 {
				w.Grid.DPiece[y][x] = 1378
				w.Grid.SOLData[1378] = world.TileSolid | world.TileBlockMissile
			}
		}
	}

	return w, lights
}

// assertGuardianTick checks one ProcessGuardian tick.
func assertGuardianTick(t *testing.T, m missiles.Missile, spawns *chainSpawns,
	lights *guardianLights, tc guardianTickCase,
) {
	t.Helper()

	if m.DelFlag != tc.wantDel || lights.unlights != tc.wantUnlights { // cpp:3572-3575
		t.Fatalf("del/unlights = %v/%d, want %v/%d", m.DelFlag, lights.unlights, tc.wantDel, tc.wantUnlights)
	}

	if len(spawns.calls) != tc.wantSpawns { // cpp:3522-3555
		t.Fatalf("spawns = %d, want %d", len(spawns.calls), tc.wantSpawns)
	}

	if tc.wantSpawns == 0 && m.MFNum != tc.wantGroup { // cpp:3513-3517
		t.Fatalf("frame group = %d, want %d", m.MFNum, tc.wantGroup)
	}

	if tc.wantSpawns == 1 {
		assertGuardianFire(t, m, spawns.calls[0])
	}
}

// assertGuardianFire checks the Firebolt GuardianTryFireAt launches
// (cpp:724-726): damage = ScaleSpellEffect(GenerateRnd(10) + level/2 + 1),
// aimed with GetDirection, and the attack frame armed for three ticks.
func assertGuardianFire(t *testing.T, m missiles.Missile, spawn spawnCall) {
	t.Helper()

	if spawn.spawnID != missiles.Firebolt || spawn.parent == nil { // cpp:729
		t.Fatalf("spawn = %v (parent %v), want Firebolt with the guardian as parent",
			spawn.spawnID, spawn.parent)
	}

	if spawn.dam != 13 { // 7 (tape) + 6/2 + 1 = 11, +1/8 twice at spell level 2
		t.Fatalf("damage = %d, want 13", spawn.dam)
	}

	if m.MFNum != missiles.GuardianFrameAttack || m.Var2 != 3 { // cpp:730-731
		t.Fatalf("frame group/var2 = %d/%d, want %d/3", m.MFNum, m.Var2, missiles.GuardianFrameAttack)
	}
}
