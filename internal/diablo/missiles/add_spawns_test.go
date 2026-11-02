// Package missiles_test pins the Hellfire spawn add bodies (missiles.cpp:
// 1240-1247, 1346-1390, 1582-1601, 1693-1714).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
)

// rowSpawnCall records one SpawnMissileRow request.
type rowSpawnCall struct {
	args missiles.AddMissileArgs
}

// spawnSpy collects the child rows an add body casts.
type spawnSpy struct {
	calls  []rowSpawnCall
	fizzle bool
}

// callbacks wires the spy into a HitWorld.
func (s *spawnSpy) callbacks() missiles.ProcessCallbacks {
	cb := missiles.NoopProcessCallbacks()
	cb.SpawnMissileRow = func(args missiles.AddMissileArgs) *missiles.Missile {
		s.calls = append(s.calls, rowSpawnCall{args: args})
		if s.fizzle {
			return nil
		}

		return &missiles.Missile{}
	}
	cb.IsMyPlayer = func(player int) bool { return player == 0 }
	cb.ShowAutoMapItems = func() { s.calls = append(s.calls, rowSpawnCall{}) }

	return cb
}

func TestAddOpenNest_DetonatesFourBigExplosionsAtTheMouth(t *testing.T) {
	// given
	spy := &spawnSpy{}
	fx := newAddFixture(t, addScript{
		start: [2]int{40, 40}, dst: [2]int{41, 40}, source: 2, dam: 17, midir: missiles.South,
	})
	fx.world.ProcessCallbacks = spy.callbacks()

	// when
	missiles.AddOpenNest(fx.world, fx.missile, fx.param)

	// then
	assertNestSpawns(t, spy.calls)

	if !fx.missile.DelFlag {
		t.Fatal("nest row not deleted")
	}
}

func TestAddHorkSpawn_FliesAtEightPixelsForNineTicks(t *testing.T) {
	// given
	fx := newAddFixture(t, addScript{start: [2]int{40, 40}, dst: [2]int{43, 40}, midir: missiles.East})

	// when
	missiles.AddHorkSpawn(fx.world, fx.missile, fx.param)

	// then
	want := missiles.PositionDisplacement{DeltaX: 370720, DeltaY: 185360} // 8 px
	if fx.missile.Position.Velocity != want {
		t.Fatalf("velocity = %v, want %v", fx.missile.Position.Velocity, want)
	}

	if fx.missile.Duration != 9 || fx.missile.Var1 != int(missiles.East) {
		t.Fatalf("duration/var1 = %d/%d, want 9/%d", fx.missile.Duration, fx.missile.Var1, int(missiles.East))
	}
}

func TestAddJester_CastsOneOfTenSpells(t *testing.T) {
	// given
	cases := []struct {
		roll int
		want missiles.MissileID
	}{
		{0, missiles.Firebolt},
		{1, missiles.Firebolt},
		{2, missiles.Fireball},
		{3, missiles.FireWallControl},
		{4, missiles.Guardian},
		{5, missiles.ChainLightning},
		{6, missiles.TownPortal},
		{7, missiles.Teleport},
		{8, missiles.Apocalypse},
		{9, missiles.StoneCurse},
	}

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{jesterName(tc.roll), func(t *testing.T) {
			t.Helper()
			assertJesterSpell(t, tc.roll, tc.want)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

func TestAddJester_ReportsAFizzledChild(t *testing.T) {
	// given
	spy := &spawnSpy{fizzle: true}
	fx := newAddFixture(t, addScript{start: [2]int{40, 40}, dst: [2]int{43, 40}, rnd: []int{5}})
	fx.world.ProcessCallbacks = spy.callbacks()
	fx.world.ProcessCallbacks.GenerateRnd = fx.tape.draw

	// when
	missiles.AddJester(fx.world, fx.missile, fx.param)
	fx.tape.assertConsumed()

	// then
	if !fx.param.SpellFizzled || !fx.missile.DelFlag {
		t.Fatalf("fizzled/delFlag = %v/%v, want true/true", fx.param.SpellFizzled, fx.missile.DelFlag)
	}
}

func TestAddLightningWall_RunsFor255TicksPerSpellLevel(t *testing.T) {
	// given
	cases := []struct {
		name   string
		source int
		spllvl int
		tile   [2]int
		wantX  int
		wantY  int
	}{
		{"a trap keeps the launch tile", -1, 0, [2]int{0, 0}, 40, 40},
		{"a player uses its own tile", 0, 2, [2]int{12, 34}, 12, 34},
	}

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()
			runLightningWall(t, tc.source, tc.spllvl, tc.tile, tc.wantX, tc.wantY)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

func TestAddSearch_MergesIntoARunningSearchRow(t *testing.T) {
	// given
	other := &missiles.Missile{Type: missiles.Search, Source: 0, Duration: 300}
	fx := newAddFixture(t, addScript{start: [2]int{40, 40}, dst: [2]int{43, 40}, source: 0, level: 5, spllvl: 1})
	fx.world.ProcessCallbacks = searchCallbacks([]*missiles.Missile{other})

	// when
	missiles.AddSearch(fx.world, fx.missile, fx.param)

	// then
	if fx.missile.Duration != 265 {
		t.Fatalf("duration = %d, want 10 + 10 + 245", fx.missile.Duration)
	}

	if other.Duration != 565 {
		t.Fatalf("merged duration = %d, want 565", other.Duration)
	}

	if !fx.missile.DelFlag {
		t.Fatal("merged Search row not deleted")
	}
}

func TestAddSearch_KeepsItsOwnTimerWithoutASibling(t *testing.T) {
	// given
	cases := []struct {
		name   string
		source int
		level  int
		want   int
	}{
		{"a trap starts at level 2", -1, 0, 2 + 245},
		{"a player doubles its level", 0, 5, 10 + 245},
	}

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()
			runSearch(t, tc.source, tc.level, tc.want)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// assertNestSpawns checks the four BigExplosion rows AddOpenNest casts.
func assertNestSpawns(t *testing.T, calls []rowSpawnCall) {
	t.Helper()

	wantTiles := []missiles.TilePosition{{X: 80, Y: 62}, {X: 81, Y: 62}, {X: 80, Y: 63}, {X: 81, Y: 63}}
	if len(calls) != len(wantTiles) {
		t.Fatalf("spawns = %d, want %d", len(calls), len(wantTiles))
	}

	for i, call := range calls {
		assertNestSpawn(t, i, call, wantTiles[i])
	}
}

// assertNestSpawn checks one nest explosion row.
func assertNestSpawn(t *testing.T, i int, call rowSpawnCall, wantTile missiles.TilePosition) {
	t.Helper()

	if call.args.Src != wantTile {
		t.Fatalf("spawn %d src = %v, want %v", i, call.args.Src, wantTile)
	}

	if call.args.Type != missiles.BigExplosion || call.args.Dst != (missiles.TilePosition{X: 80, Y: 62}) {
		t.Fatalf("spawn %d type/dst = %d/%v", i, call.args.Type, call.args.Dst)
	}

	if call.args.Source != 2 || call.args.Dam != 17 || call.args.Midir != missiles.South {
		t.Fatalf("spawn %d did not inherit source/dam/dir", i)
	}
}

// jesterName labels a jester roll case.
func jesterName(roll int) string {
	return "roll " + string(rune('0'+roll)) + " casts its spell"
}

// assertJesterSpell runs one jester roll and checks the cast spell.
func assertJesterSpell(t *testing.T, roll int, want missiles.MissileID) {
	t.Helper()

	// given
	spy := &spawnSpy{}
	fx := newAddFixture(t, addScript{start: [2]int{40, 40}, dst: [2]int{43, 40}, rnd: []int{roll}})
	fx.world.ProcessCallbacks = spy.callbacks()
	fx.world.ProcessCallbacks.GenerateRnd = fx.tape.draw

	// when
	missiles.AddJester(fx.world, fx.missile, fx.param)
	fx.tape.assertConsumed()

	// then
	if len(spy.calls) != 1 || spy.calls[0].args.Type != want {
		t.Fatalf("cast = %v, want %d", spy.calls, want)
	}
}

// runLightningWall runs one AddLightningWall case.
func runLightningWall(t *testing.T, source, spllvl int, tile [2]int, wantX, wantY int) {
	t.Helper()

	// given
	fx := newAddFixture(t, addScript{
		start: [2]int{40, 40}, dst: [2]int{43, 40}, source: source,
		spllvl: spllvl, tile: tile, rnd: []int{3},
	})

	// when
	missiles.AddLightningWall(fx.world, fx.missile, fx.param)
	fx.tape.assertConsumed()

	// then
	if fx.missile.Duration != 255*(spllvl+1) || fx.missile.AnimFrame != 4 {
		t.Fatalf("duration/frame = %d/%d, want %d/4", fx.missile.Duration, fx.missile.AnimFrame, 255*(spllvl+1))
	}

	if fx.missile.Var1 != wantX || fx.missile.Var2 != wantY {
		t.Fatalf("var1/var2 = %d/%d, want %d/%d", fx.missile.Var1, fx.missile.Var2, wantX, wantY)
	}

	want := missiles.PositionDisplacement{DeltaX: 741440, DeltaY: 370720} // 16 px
	if fx.missile.Position.Velocity != want {
		t.Fatalf("velocity = %v, want %v", fx.missile.Position.Velocity, want)
	}
}

// searchCallbacks wires a missile list plus the auto-map sink.
func searchCallbacks(rows []*missiles.Missile) missiles.ProcessCallbacks {
	cb := missiles.NoopProcessCallbacks()
	cb.OtherMissiles = func() []*missiles.Missile { return rows }
	cb.IsMyPlayer = func(player int) bool { return player == 0 }

	return cb
}

// runSearch runs one AddSearch case with no sibling row.
func runSearch(t *testing.T, source, level, want int) {
	t.Helper()

	// given
	fx := newAddFixture(t, addScript{start: [2]int{40, 40}, dst: [2]int{43, 40}, source: source, level: level})
	fx.world.ProcessCallbacks = searchCallbacks(nil)

	// when
	missiles.AddSearch(fx.world, fx.missile, fx.param)

	// then
	if fx.missile.Duration != want {
		t.Fatalf("duration = %d, want %d", fx.missile.Duration, want)
	}

	if fx.missile.DelFlag {
		t.Fatal("unmerged Search row deleted itself")
	}
}
