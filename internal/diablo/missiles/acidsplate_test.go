// Package missiles_test pins ProcessAcidSplate (missiles.cpp:3639-3654) — the
// Process side of the AcidSplat row (misdat row 58, whose Add is the shared
// AddMissileExplosion).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// acidSplateCases pin the ticks: the first tick (duration == AnimLen) applies
// the +SouthEast/-32 render shift (cpp:3641-3644), the middle ticks only count
// down, and the tick reaching 0 spawns the puddle and deletes the splat
// (cpp:3647-3653).
var acidSplateCases = []acidSplateCase{
	{
		name: "first tick applies the render shift", duration: 5, animLen: 5,
		wantDuration: 4, wantTile: missiles.TilePosition{X: 5, Y: 5}, wantOffsetY: -32,
		wantPut: true,
	},
	{
		name: "middle tick only counts down", duration: 4, animLen: 5,
		wantDuration: 3, wantTile: missiles.TilePosition{X: 4, Y: 4}, wantPut: true,
	},
	{
		name: "last tick spawns a level-1 puddle", duration: 1, animLen: 5, monsterLevel: 1,
		wantTile: missiles.TilePosition{X: 4, Y: 4}, wantSpawn: true, wantPuddleDam: 1,
	},
	{
		name: "last tick spawns a level-2 puddle", duration: 1, animLen: 5, monsterLevel: 2,
		wantTile: missiles.TilePosition{X: 4, Y: 4}, wantSpawn: true, wantPuddleDam: 2,
	},
}

// acidSplateCase pins one ProcessAcidSplate tick.
type acidSplateCase struct {
	name          string
	duration      int
	animLen       int
	monsterLevel  int
	wantDuration  int
	wantTile      missiles.TilePosition
	wantOffsetY   int
	wantSpawn     bool
	wantPuddleDam int
	wantPut       bool
}

// acidSplateSpies records the sink calls of one tick.
type acidSplateSpies struct {
	spawns []spawnCall
	puts   int
}

// TestProcessAcidSplate pins the render shift, the countdown, and the
// puddle spawn whose damage follows the source monster's level
// (cpp:3646-3650).
func TestProcessAcidSplate(t *testing.T) {
	// given
	cases := acidSplateCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			spies := &acidSplateSpies{}
			w := acidSplateWorld(spies, tc.monsterLevel)

			// when
			m := acidSplateMissile(tc)
			missiles.ProcessAcidSplate(w, &m)

			// then
			assertAcidSplateState(t, m, tc)
			assertAcidSplateSinks(t, m, spies, tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// acidSplateWorld wires the spawn sink and counts PutMissile calls.
func acidSplateWorld(spies *acidSplateSpies, monsterLevel int) *missiles.HitWorld {
	cb := missiles.NoopProcessCallbacks()
	cb.SpawnMissile = func(src, dst missiles.TilePosition, dir missiles.Direction,
		spawnID missiles.MissileID, caster, source, dam, spellLevel int,
		parent *missiles.Missile,
	) {
		spies.spawns = append(spies.spawns,
			spawnCall{src, dst, dir, spawnID, caster, source, dam, spellLevel, parent})
		spies.puts++
	}

	w := &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: cb,
		AddSource: missiles.AddSource{
			MonsterStats: func(int) missiles.AddMonsterStats {
				return missiles.AddMonsterStats{Level: monsterLevel}
			},
		},
	}

	return w
}

// acidSplateMissile seeds the splat at tile (4, 4) with the case's timing.
func acidSplateMissile(tc acidSplateCase) missiles.Missile {
	m := missiles.Missile{
		Type: missiles.AcidSplat, Duration: tc.duration, AnimLen: tc.animLen,
		Source: 3, SpellLevel: 1,
	}
	m.SetTraveledFromTile(4, 4)

	return m
}

// assertAcidSplateState checks the duration, delete flag and render shift.
func assertAcidSplateState(t *testing.T, m missiles.Missile, tc acidSplateCase) {
	t.Helper()

	wantDel := tc.wantSpawn
	if m.Duration != tc.wantDuration || m.DelFlag != wantDel {
		t.Fatalf("duration/del = %d/%v, want %d/%v",
			m.Duration, m.DelFlag, tc.wantDuration, wantDel)
	}

	if m.Position.Tile != tc.wantTile {
		t.Fatalf("tile = %v, want %v", m.Position.Tile, tc.wantTile)
	}

	if tc.wantOffsetY != 0 && int(m.Position.Offset.DeltaY) != tc.wantOffsetY {
		t.Fatalf("offsetY = %d, want %d", m.Position.Offset.DeltaY, tc.wantOffsetY)
	}
}

// assertAcidSplateSinks checks the spawn pack and the PutMissile branch.
func assertAcidSplateSinks(t *testing.T, m missiles.Missile, spies *acidSplateSpies, tc acidSplateCase) {
	t.Helper()

	spawns := spies.spawns
	if tc.wantSpawn != (len(spawns) == 1) {
		t.Fatalf("spawns = %d, want spawn %v", len(spawns), tc.wantSpawn)
	}

	if !tc.wantSpawn {
		return
	}

	spawn := spawns[0]
	if spawn.spawnID != missiles.AcidPuddle || spawn.dam != tc.wantPuddleDam ||
		spawn.caster != int(missiles.TargetPlayers) || spawn.dir != missiles.South {
		t.Fatalf("spawn = %+v, want AcidPuddle dam %d from a player-targeting south spawn",
			spawn, tc.wantPuddleDam)
	}

	if spawn.src != m.Position.Tile {
		t.Fatalf("spawn src = %v, want the splat tile %v", spawn.src, m.Position.Tile)
	}

	if spawn.dst != (missiles.TilePosition{}) || spawn.parent != nil { // cpp:3649
		t.Fatalf("spawn dst/parent = %v/%v, want the zero aim tile and no parent",
			spawn.dst, spawn.parent)
	}
}
