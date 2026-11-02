// Package missiles_test pins AddApocalypse (missiles.cpp:2652-2665) and
// ProcessApocalypse (cpp:3850-3872), the 16x16 box walk that drops one
// ApocalypseBoom per tick until the box is exhausted.
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// apocalypseTickCases pin the one-boom-per-tick walk (cpp:3850-3872).
var apocalypseTickCases = []apocalypseTickCase{
	{
		name: "first monster in the box takes the boom", monsterX: 13, monsterY: 12,
		wantSpawns: 1, wantVar2: 12, wantVar4: 14,
	},
	{
		name: "a player minion is skipped", monsterX: 13, monsterY: 12, minion: true,
		wantDel: true, wantVar2: 12, wantVar4: 12,
	},
	{name: "an empty box ends the row", wantDel: true, wantVar2: 12, wantVar4: 12},
}

// apocalypseTickCase pins one ProcessApocalypse tick.
type apocalypseTickCase struct {
	name       string
	monsterX   int
	monsterY   int
	minion     bool
	wantSpawns int
	wantDel    bool
	wantVar2   int
	wantVar4   int
}

// TestAddApocalypse pins the box, the damage roll and the 255-tick life
// (cpp:2656-2664).
func TestAddApocalypse(t *testing.T) {
	// given
	w := apocalypseWorld(&rndTape{t: t, tape: []int{1, 2, 3}}, &chainSpawns{})
	m := missiles.Missile{Type: missiles.Apocalypse, Source: 0}
	m.SetTraveledFromTile(20, 20)
	m.Position.Start = missiles.TilePosition{X: 20, Y: 20}

	// when
	missiles.AddApocalypse(w, &m, &missiles.AddParameter{})

	// then
	if m.Var1 != 8 || m.Var2 != 12 || m.Var3 != 28 || m.Var4 != 12 || m.Var5 != 28 || m.Var6 != 12 {
		t.Fatalf("box = %d/%d/%d/%d/%d/%d, want 8/12/28/12/28/12",
			m.Var1, m.Var2, m.Var3, m.Var4, m.Var5, m.Var6)
	}

	if m.Dam != 9 || m.Duration != 255 { // cpp:2663-2664: (1+2+3) + level 3
		t.Fatalf("dam/duration = %d/%d, want 9/255", m.Dam, m.Duration)
	}
}

// TestProcessApocalypse pins the walk over the box.
func TestProcessApocalypse(t *testing.T) {
	// when
	subtests := make([]subtest, 0, len(apocalypseTickCases))
	for _, tc := range apocalypseTickCases {
		subtests = append(subtests, subtest{tc.name, runApocalypseTick(tc)})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// runApocalypseTick drives one box-walk tick.
func runApocalypseTick(tc apocalypseTickCase) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		// given
		spawns := &chainSpawns{}
		w := apocalypseWorld(&rndTape{t: t}, spawns)
		w.ProcessCallbacks.MonsterIsPlayerMinion = func(int) bool { return tc.minion }
		if tc.monsterX != 0 {
			w.DMonster[tc.monsterY][tc.monsterX] = 1
		}

		m := missiles.Missile{
			Type: missiles.Apocalypse, Source: 0, Dam: 9, Duration: 255,
			Var1: 8, Var2: 12, Var3: 28, Var4: 12, Var5: 28, Var6: 12,
		}

		// when
		missiles.ProcessApocalypse(w, &m)

		// then
		assertApocalypseTick(t, m, spawns, tc)
	}
}

// apocalypseWorld wires the tape and the boom recorder.
func apocalypseWorld(tape *rndTape, spawns *chainSpawns) *missiles.HitWorld {
	cb := missiles.NoopProcessCallbacks()
	cb.GenerateRnd = tape.draw
	cb.PlayerDirection = func(int) missiles.Direction { return missiles.East }
	cb.SpawnMissile = func(src, dst missiles.TilePosition, dir missiles.Direction,
		spawnID missiles.MissileID, caster, source, dam, spellLevel int, parent *missiles.Missile,
	) {
		spawns.calls = append(spawns.calls,
			spawnCall{src, dst, dir, spawnID, caster, source, dam, spellLevel, parent})
	}

	return &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: cb,
		AddSource: missiles.AddSource{
			PlayerStats: func(int) missiles.AddPlayerStats { return missiles.AddPlayerStats{Level: 3} },
		},
	}
}

// assertApocalypseTick checks one box-walk tick: the cursor advances past the
// tile it fired at or the box is exhausted (cpp:3866-3872).
func assertApocalypseTick(t *testing.T, m missiles.Missile, spawns *chainSpawns, tc apocalypseTickCase) {
	t.Helper()

	if len(spawns.calls) != tc.wantSpawns || m.DelFlag != tc.wantDel {
		t.Fatalf("spawns/del = %d/%v, want %d/%v", len(spawns.calls), m.DelFlag, tc.wantSpawns, tc.wantDel)
	}

	if tc.wantSpawns == 1 {
		assertApocalypseBoom(t, spawns.calls[0], tc)
	}

	if m.Var2 != tc.wantVar2 || m.Var4 != tc.wantVar4 { // cpp:3866-3870
		t.Fatalf("cursor = %d/%d, want %d/%d", m.Var2, m.Var4, tc.wantVar2, tc.wantVar4)
	}
}

// assertApocalypseBoom checks the boom the walk dropped (cpp:3865).
func assertApocalypseBoom(t *testing.T, spawn spawnCall, tc apocalypseTickCase) {
	t.Helper()

	if spawn.spawnID != missiles.ApocalypseBoom || spawn.dam != 9 || spawn.dir != missiles.East ||
		spawn.dst != (missiles.TilePosition{X: tc.monsterX, Y: tc.monsterY}) {
		t.Fatalf("spawn = %v dam %d dir %v at %v, want ApocalypseBoom dam 9 East at the monster",
			spawn.spawnID, spawn.dam, spawn.dir, spawn.dst)
	}
}
