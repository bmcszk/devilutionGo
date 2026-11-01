// Package missiles_test pins AddNova (missiles.cpp:2567-2581) and ProcessNova
// (cpp:3309-3312 through ProcessNovaCommon, cpp:3280-3304).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// novaAddCases pin the damage roll: a player cast draws GenerateRndSum(6, 5) —
// five dice of six — adds the character level plus 5, then halves and scales
// (cpp:2574-2575); a trap draws GenerateRndSum(3, 3) plus currlevel / 2
// (cpp:2577).
var novaProcessCases = []novaProcessCase{
	{name: "wave then countdown", duration: 2, wantDuration: 1},
	{name: "last tick deletes", duration: 1, wantDel: true},
}

var novaAddCases = []novaAddCase{
	{
		name: "player cast rolls five dice of six", source: 0, level: 3, spellLevel: 1,
		rnd: []int{1, 2, 3, 4, 5}, wantDam: 12,
	},
	{
		name: "trap cast rolls three dice", source: -1, currlevel: 4,
		rnd: []int{2, 2, 2}, wantDam: 8,
	},
}

// novaAddCase pins one AddNova row.
type novaAddCase struct {
	name       string
	source     int // _misource: -1 marks a trap (collision.go:277)
	level      int
	currlevel  int
	spellLevel int
	rnd        []int
	wantDam    int
}

// novaProcessCases pin the ticks: the spawn wave fires on every tick and the
// duration counts down to the delete.
// novaProcessCase pins one ProcessNova tick.
type novaProcessCase struct {
	name         string
	duration     int
	wantDuration int
	wantDel      bool
}

// novaSpies records the spawn wave of one tick.
type novaSpies struct {
	spawns []spawnCall
}

// TestAddNova pins the damage roll of both caster branches and the one-tick
// duration.
func TestAddNova(t *testing.T) {
	// given
	cases := novaAddCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			tape := &rndTape{t: t, tape: tc.rnd}
			w := novaWorld(&novaSpies{}, tc.currlevel, tc.level, tape)
			m := missiles.Missile{
				Type: missiles.Nova, SpellLevel: tc.spellLevel, Source: tc.source,
				Caster: int(missiles.TargetMonsters),
			}
			p := &missiles.AddParameter{Dst: missiles.TilePosition{X: 7, Y: 2}}

			// when
			missiles.AddNova(w, &m, p)
			tape.assertConsumed()

			// then
			assertNovaAdd(t, m, p.Dst, tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// TestProcessNova pins the 36-missile wave, its direction/damage pack and the
// countdown.
func TestProcessNova(t *testing.T) {
	// given
	cases := novaProcessCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			spies := &novaSpies{}
			w := novaWorld(spies, 0, 0, &rndTape{t: t})

			// when
			m := missiles.Missile{
				Type: missiles.Nova, Duration: tc.duration, Dam: 9, SpellLevel: 2,
				Caster: int(missiles.TargetMonsters), Source: 1,
			}
			m.SetTraveledFromTile(4, 4)
			missiles.ProcessNova(w, &m)

			// then
			assertNovaTick(t, m, tc)
			assertNovaWave(t, spies, m)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// novaWorld wires the spawn sink and the player-direction read.
func novaWorld(spies *novaSpies, currlevel, level int, tape *rndTape) *missiles.HitWorld {
	cb := missiles.NoopProcessCallbacks()
	cb.SpawnMissile = func(src, dst missiles.TilePosition, dir missiles.Direction,
		spawnID missiles.MissileID, caster, source, dam, spellLevel int,
		parent *missiles.Missile,
	) {
		spies.spawns = append(spies.spawns,
			spawnCall{src, dst, dir, spawnID, caster, source, dam, spellLevel, parent})
	}
	cb.Currlevel = currlevel
	cb.GenerateRnd = tape.draw
	cb.PlayerDirection = func(int) missiles.Direction { return missiles.NorthWest }

	return &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: cb,
		AddSource: missiles.AddSource{
			PlayerStats: func(int) missiles.AddPlayerStats { return missiles.AddPlayerStats{Level: level} },
		},
	}
}

// assertNovaAdd checks the damage roll, the stashed aim tile and the one-tick
// duration of one AddNova run.
func assertNovaAdd(t *testing.T, m missiles.Missile, dst missiles.TilePosition, tc novaAddCase) {
	t.Helper()

	if m.Dam != tc.wantDam { // cpp:2574-2577
		t.Fatalf("Dam = %d, want %d", m.Dam, tc.wantDam)
	}

	if m.Var1 != dst.X || m.Var2 != dst.Y { // cpp:2569-2570
		t.Fatalf("var1/var2 = %d/%d, want the aim tile %v", m.Var1, m.Var2, dst)
	}

	if m.Duration != 1 { // cpp:2580
		t.Fatalf("Duration = %d, want 1", m.Duration)
	}
}

// assertNovaTick checks the countdown and the delete flag of one tick.
func assertNovaTick(t *testing.T, m missiles.Missile, tc novaProcessCase) {
	t.Helper()

	if m.Duration != tc.wantDuration || m.DelFlag != tc.wantDel {
		t.Fatalf("duration/del = %d/%v, want %d/%v",
			m.Duration, m.DelFlag, tc.wantDuration, tc.wantDel)
	}
}

// assertNovaWave checks the 36 spawns: 9 quarter offsets x 4 flips, all
// NovaBall, aimed from the nova tile with the parent's direction, damage and
// spell level (cpp:3295-3301).
func assertNovaWave(t *testing.T, spies *novaSpies, m missiles.Missile) {
	t.Helper()

	if len(spies.spawns) != 36 {
		t.Fatalf("spawns = %d, want 36 (9 quarter offsets x 4 flips)", len(spies.spawns))
	}

	for _, spawn := range spies.spawns {
		if spawn.spawnID != missiles.NovaBall || spawn.src != m.Position.Tile ||
			spawn.dir != missiles.NorthWest || spawn.dam != 9 || spawn.spellLevel != 2 ||
			spawn.caster != int(missiles.TargetMonsters) || spawn.parent != nil {
			t.Fatalf("spawn = %+v, want a NovaBall from tile %v with the parent's pack",
				spawn, m.Position.Tile)
		}
	}
}
