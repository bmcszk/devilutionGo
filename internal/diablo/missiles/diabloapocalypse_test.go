// Package missiles_test pins AddDiabloApocalypse (missiles.cpp:2788-2799), the
// body that drops one boom on every active player with a clear missile line.
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// diabloApocalypseCase pins one AddDiabloApocalypse call.
type diabloApocalypseCase struct {
	name       string
	futures    []missiles.TilePosition
	blocked    missiles.TilePosition
	wantSpawns int
}

// TestAddDiabloApocalypse pins the per-player drop and the 2x2-free
// requirement (cpp:2790-2798): active players only, a clear missile line from
// the launch tile, source tile {0,0}, direction South and the stored damage.
func TestAddDiabloApocalypse(t *testing.T) {
	// given
	cases := []diabloApocalypseCase{
		{
			name: "two players in sight", futures: []missiles.TilePosition{{X: 6, Y: 6}, {X: 7, Y: 7}},
			wantSpawns: 2,
		},
		{
			name: "a wall hides one player", futures: []missiles.TilePosition{{X: 6, Y: 6}, {X: 9, Y: 5}},
			blocked: missiles.TilePosition{X: 8, Y: 5}, wantSpawns: 1,
		},
		{name: "nobody active is nothing to do", wantSpawns: 0},
	}

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			spawns := &chainSpawns{}
			w := diabloApocalypseWorld(t, spawns, tc)
			m := missiles.Missile{
				Type: missiles.DiabloApocalypse, Source: 0, Dam: 40,
				Caster: int(missiles.TargetPlayers),
			}
			m.Position.Start = missiles.TilePosition{X: 5, Y: 5}

			// when
			missiles.AddDiabloApocalypse(w, &m, &missiles.AddParameter{})

			// then
			if len(spawns.calls) != tc.wantSpawns || !m.DelFlag { // cpp:2798-2799
				t.Fatalf("spawns/del = %d/%v, want %d/true", len(spawns.calls), m.DelFlag, tc.wantSpawns)
			}

			assertDiabloBooms(t, spawns.calls, tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// assertDiabloBooms checks one boom per surviving player (cpp:2797).
func assertDiabloBooms(t *testing.T, calls []spawnCall, tc diabloApocalypseCase) {
	t.Helper()

	for i, spawn := range calls {
		if spawn.dst != tc.futures[i] || spawn.spawnID != missiles.DiabloApocalypseBoom ||
			spawn.src != (missiles.TilePosition{}) || spawn.dir != missiles.South || spawn.dam != 40 {
			t.Fatalf("spawn %d = %+v, want a South DiabloApocalypseBoom at the player tile", i, spawn)
		}
	}
}

// diabloApocalypseWorld wires the player store, the missile recorder and the
// grid.
func diabloApocalypseWorld(t *testing.T, spawns *chainSpawns, tc diabloApocalypseCase) *missiles.HitWorld {
	t.Helper()

	cb := missiles.NoopProcessCallbacks()
	cb.ActivePlayers = func() []int {
		ids := make([]int, len(tc.futures))
		for i := range ids {
			ids[i] = i
		}

		return ids
	}
	cb.PlayerFutureTile = func(player int) missiles.TilePosition { return tc.futures[player] }
	cb.SpawnMissile = func(src, dst missiles.TilePosition, dir missiles.Direction,
		spawnID missiles.MissileID, caster, source, dam, spellLevel int, parent *missiles.Missile,
	) {
		spawns.calls = append(spawns.calls,
			spawnCall{src, dst, dir, spawnID, caster, source, dam, spellLevel, parent})
	}

	w := &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: cb,
	}
	if tc.blocked != (missiles.TilePosition{}) {
		w.Grid.DPiece[tc.blocked.Y][tc.blocked.X] = 1378
		w.Grid.SOLData[1378] = world.TileBlockMissile
	}

	return w
}
