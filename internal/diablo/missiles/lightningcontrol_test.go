// Package missiles_test pins AddLightningControl (missiles.cpp:2782-2790) and
// ProcessLightningControl (cpp:3238-3262) with its SpawnLightning tail
// (cpp:802-849).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// monsterTypeStorm is MT_STORM (tables/monstdat.h:262) — monstdat.tsv file
// line 78, so data row 76.
const monsterTypeStorm = 76

// lightningControlCases pin the ticks: every branch rolls its own damage and
// hands it to SpawnLightning, which walks the bolt forward and forks a
// Lightning (or ThinLightning for the storm monsters, cpp:825-828) back at the
// launch tile.
var lightningControlCases = []lightningControlCase{
	{
		name: "monster cast doubles its damage range", caster: 2, source: 1,
		minDamage: 2, maxDamage: 4, rnd: []int{2}, wantDam: 8, wantType: missiles.Lightning,
	},
	{
		name: "storm monster forks thin lightning", caster: 2, source: 1,
		monsterType: monsterTypeStorm, minDamage: 1, maxDamage: 1, rnd: []int{0}, wantDam: 2,
		wantType: missiles.ThinLightning,
	},
	{
		name: "player cast rolls level and a coin", caster: 0, source: 0,
		level: 3, rnd: []int{1, 2}, wantDam: 320, wantType: missiles.Lightning,
	},
	{
		name: "trap rolls currlevel", caster: 2, source: -1,
		currlevel: 5, rnd: []int{4}, wantDam: 14, wantType: missiles.Lightning,
	},
}

// lightningControlCase pins one ProcessLightningControl tick.
type lightningControlCase struct {
	name        string
	caster      int
	source      int
	level       int
	currlevel   int
	monsterType int
	minDamage   int
	maxDamage   int
	rnd         []int
	wantDam     int
	wantType    missiles.MissileID
}

// TestAddLightningControl pins the remembered launch tile, the 32-pixel
// velocity, the random start frame and the 256-tick life (cpp:2784-2789).
func TestAddLightningControl(t *testing.T) {
	// given
	tape := &rndTape{t: t, tape: []int{5}}
	w := lightningControlWorld(tape, lightningControlCase{})
	m := missiles.Missile{Type: missiles.LightningControl}
	m.SetTraveledFromTile(4, 4)
	m.Position.Start = missiles.TilePosition{X: 2, Y: 3}

	// when
	missiles.AddLightningControl(w, &m, &missiles.AddParameter{Dst: missiles.TilePosition{X: 2, Y: 9}})
	tape.assertConsumed()

	// then
	if m.Var1 != 2 || m.Var2 != 3 { // cpp:2784-2785
		t.Fatalf("var1/var2 = %d/%d, want the launch tile 2/3", m.Var1, m.Var2)
	}

	if m.AnimFrame != 6 { // cpp:2787: RandomIntBetween(1, 8), tape 5 -> 6
		t.Fatalf("AnimFrame = %d, want 6", m.AnimFrame)
	}

	if m.Duration != 256 { // cpp:2788
		t.Fatalf("Duration = %d, want 256", m.Duration)
	}

	if m.Position.Velocity == (missiles.PositionDisplacement{}) { // cpp:2786
		t.Fatal("velocity stays zero, want the 32-pixel heading")
	}
}

// TestProcessLightningControl pins the damage branches and the fork their
// SpawnLightning tail issues.
func TestProcessLightningControl(t *testing.T) {
	// given
	cases := lightningControlCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			calls, w := lightningControlFixture(t, tc)
			m := missiles.Missile{
				Type: missiles.LightningControl, Duration: 3, Source: tc.source, Caster: tc.caster,
				Var1: 3, Var2: 9,
			}
			m.SetTraveledFromTile(4, 4)
			m.Position.Start = missiles.TilePosition{X: 4, Y: 4}

			// when
			missiles.ProcessLightningControl(w, &m)

			// then
			assertLightningControlTick(t, &m, calls, tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// lightningControlWorld wires the tape and the stat reads.
func lightningControlWorld(tape *rndTape, tc lightningControlCase) *missiles.HitWorld {
	cb := missiles.NoopProcessCallbacks()
	cb.GenerateRnd = tape.draw
	cb.Currlevel = tc.currlevel

	return &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: cb,
		AddSource: missiles.AddSource{
			PlayerStats: func(int) missiles.AddPlayerStats {
				return missiles.AddPlayerStats{Level: tc.level}
			},
			MonsterStats: func(int) missiles.AddMonsterStats {
				return missiles.AddMonsterStats{
					Type: tc.monsterType, MinDamage: tc.minDamage, MaxDamage: tc.maxDamage,
				}
			},
		},
	}
}

// lightningControlFixture wires the world, the tape and the fork recorder.
func lightningControlFixture(t *testing.T, tc lightningControlCase) (*chainSpawns, *missiles.HitWorld) {
	t.Helper()

	calls := &chainSpawns{}
	w := lightningControlWorld(&rndTape{t: t, tape: tc.rnd}, tc)
	w.ProcessCallbacks.SpawnMissile = func(src, dst missiles.TilePosition, dir missiles.Direction,
		spawnID missiles.MissileID, caster, source, dam, spellLevel int, parent *missiles.Missile,
	) {
		calls.calls = append(calls.calls,
			spawnCall{src, dst, dir, spawnID, caster, source, dam, spellLevel, parent})
	}

	return calls, w
}

// assertLightningControlTick checks the fork's damage, type and start tile plus
// the bolt's own state.
func assertLightningControlTick(t *testing.T, m *missiles.Missile, calls *chainSpawns,
	tc lightningControlCase,
) {
	t.Helper()

	if len(calls.calls) != 1 {
		t.Fatalf("spawns = %d, want 1", len(calls.calls))
	}

	spawn := calls.calls[0]
	if spawn.spawnID != tc.wantType || spawn.dam != tc.wantDam { // cpp:825-838
		t.Fatalf("spawn = %v/%d, want %v/%d", spawn.spawnID, spawn.dam, tc.wantType, tc.wantDam)
	}

	if spawn.dst != (missiles.TilePosition{X: 4, Y: 4}) { // cpp:830
		t.Fatalf("spawn dst = %v, want the launch tile", spawn.dst)
	}

	if spawn.parent != m { // cpp:838
		t.Fatalf("spawn parent = %v, want the control missile", spawn.parent)
	}

	if m.Var1 != m.Position.Tile.X || m.Var2 != m.Position.Tile.Y { // cpp:840-841
		t.Fatalf("var1/var2 = %d/%d, want the walked tile %v", m.Var1, m.Var2, m.Position.Tile)
	}
}
