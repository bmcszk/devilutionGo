// Package missiles_test pins AddChainLightning (missiles.cpp:3565-3570) and
// ProcessChainLightning (cpp:3572-3600).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// chainCases pin the ticks: the bolt always forks one control missile at the
// aim tile (cpp:3579) plus one per monster inside the crawl radius
// min(spell level + 3, MaxCrawlRadius) (cpp:3587-3595), and the single-tick
// life deletes it (cpp:3597-3599).
var chainCases = []chainCase{
	{
		name: "aim tile only", spellLevel: 0, monsters: nil,
		wantSpawns: 1, wantDel: true,
	},
	{
		name: "monsters inside the radius fork", spellLevel: 2,
		monsters: [][2]int{{5, 4}, {6, 6}}, wantSpawns: 3, wantDel: true,
	},
	{
		name: "monsters outside the radius are ignored", spellLevel: 0,
		monsters: [][2]int{{9, 4}}, wantSpawns: 1, wantDel: true,
	},
}

// chainCase pins one ProcessChainLightning tick.
type chainCase struct {
	name       string
	spellLevel int
	monsters   [][2]int
	wantSpawns int
	wantDel    bool
}

// chainSpawns records the control missiles one tick forks.
type chainSpawns struct {
	calls []spawnCall
}

// TestAddChainLightning pins the remembered aim tile and the single-tick life
// (cpp:3567-3569).
func TestAddChainLightning(t *testing.T) {
	// given
	w := chainWorld(&rndTape{t: t})
	m := missiles.Missile{Type: missiles.ChainLightning}

	// when
	missiles.AddChainLightning(w, &m, &missiles.AddParameter{Dst: missiles.TilePosition{X: 7, Y: 8}})

	// then
	if m.Var1 != 7 || m.Var2 != 8 { // cpp:3567-3568
		t.Fatalf("var1/var2 = %d/%d, want 7/8", m.Var1, m.Var2)
	}

	if m.Duration != 1 { // cpp:3569
		t.Fatalf("Duration = %d, want 1", m.Duration)
	}
}

// TestProcessChainLightning pins the fork targets, their directions and the
// delete tail.
func TestProcessChainLightning(t *testing.T) {
	// given
	cases := chainCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			calls, w := chainFixture(t, tc)

			// when
			m := missiles.Missile{
				Type: missiles.ChainLightning, SpellLevel: tc.spellLevel, Var1: 7, Var2: 4,
				Duration: 1, Source: 2, Caster: 1,
			}
			m.SetTraveledFromTile(4, 4)
			missiles.ProcessChainLightning(w, &m)

			// then
			assertChainTick(t, m, calls, tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// chainFixture wires the fork recorder, the tape and the monster tiles.
func chainFixture(t *testing.T, tc chainCase) (*chainSpawns, *missiles.HitWorld) {
	t.Helper()

	calls := &chainSpawns{}
	w := chainWorld(&rndTape{t: t})
	w.ProcessCallbacks.SpawnMissile = func(src, dst missiles.TilePosition, dir missiles.Direction,
		spawnID missiles.MissileID, caster, source, dam, spellLevel int, parent *missiles.Missile,
	) {
		calls.calls = append(calls.calls,
			spawnCall{src, dst, dir, spawnID, caster, source, dam, spellLevel, parent})
	}

	for _, tile := range tc.monsters {
		w.DMonster[tile[1]][tile[0]] = 1
	}

	return calls, w
}

// chainWorld wires the fork sink and the grid.
func chainWorld(tape *rndTape) *missiles.HitWorld {
	cb := missiles.NoopProcessCallbacks()
	cb.GenerateRnd = tape.draw

	return &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: cb,
	}
}

// assertChainTick checks the fork count, the fork targets and the delete tail.
func assertChainTick(t *testing.T, m missiles.Missile, calls *chainSpawns, tc chainCase) {
	t.Helper()

	if len(calls.calls) != tc.wantSpawns {
		t.Fatalf("spawns = %d, want %d", len(calls.calls), tc.wantSpawns)
	}

	// The first fork always targets the aim tile (cpp:3579).
	if first := calls.calls[0]; first.dst != (missiles.TilePosition{X: 7, Y: 4}) ||
		first.spawnID != missiles.LightningControl {
		t.Fatalf("first spawn = %v/%v, want the aim tile as LightningControl", first.dst, first.spawnID)
	}

	if m.DelFlag != tc.wantDel { // cpp:3597-3599
		t.Fatalf("DelFlag = %v, want %v", m.DelFlag, tc.wantDel)
	}
}
