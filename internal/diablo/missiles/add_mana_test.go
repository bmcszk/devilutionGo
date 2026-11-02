// Package missiles_test pins the Hellfire mana and Reflect add bodies
// (missiles.cpp:1278-1297, 1458-1476, 1655-1684).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
)

// manaSpy records the mana sinks AddMana/AddMagi/AddStealMana reach.
type manaSpy struct {
	pool       missiles.PlayerMana
	setMana    int
	setBase    int
	redraws    int
	calcInv    int
	sfxName    string
	sfxTile    missiles.TilePosition
	reflectSet int
	sentTotal  int
}

// manaCallbacks wires the spy into a HitWorld.
func (s *manaSpy) manaCallbacks() missiles.ProcessCallbacks {
	cb := missiles.NoopProcessCallbacks()
	cb.PlayerMana = func(int) missiles.PlayerMana { return s.pool }
	cb.SetPlayerMana = func(_, mana, manaBase int) {
		s.setMana = mana
		s.setBase = manaBase
	}
	cb.ClassManaRestore = func(int) int { return 32 } // 0.5 in 26.6
	cb.RedrawMana = func() { s.redraws++ }
	cb.CalcPlayerInv = func(int) { s.calcInv++ }
	cb.PlaySfxLoc = func(name string, x, y int) {
		s.sfxName = name
		s.sfxTile = missiles.TilePosition{X: x, Y: y}
	}
	cb.PlayerReflections = func(int) int { return s.reflectSet }
	cb.SetPlayerReflections = func(_, reflections int) { s.reflectSet = reflections }
	cb.SendSetReflect = func(reflections int) { s.sentTotal = reflections }
	cb.IsMyPlayer = func(int) bool { return true }

	return cb
}

func TestAddMana_AddsTheScaledRollToBothPools(t *testing.T) {
	// given
	spy := &manaSpy{pool: missiles.PlayerMana{Mana: 100, MaxMana: 1000, ManaBase: 50, MaxManaBase: 400}}
	fx := newAddFixture(t, addScript{
		source: 0, level: 2, spllvl: 1,
		// d10 + 1 = 6; two d4 + 1 = 3, 4; d6 + 1 = 3 → 16 mana in 26.6,
		// scaled by splRestoreMana 0.5 → 8 mana.
		rnd: []int{5, 2, 3, 2},
	})
	fx.world.ProcessCallbacks = spy.manaCallbacks()
	fx.world.ProcessCallbacks.GenerateRnd = fx.tape.draw

	// when
	missiles.AddMana(fx.world, fx.missile, fx.param)
	fx.tape.assertConsumed()

	// then
	if spy.setMana != 612 { // 100 + 8<<6
		t.Fatalf("mana = %d, want 612", spy.setMana)
	}

	if spy.setBase != 400 {
		t.Fatalf("base = %d, want the 400 cap", spy.setBase)
	}

	if !fx.missile.DelFlag || spy.redraws != 1 {
		t.Fatalf("delFlag/redraws = %v/%d, want true/1", fx.missile.DelFlag, spy.redraws)
	}
}

func TestAddMagi_RefillsBothPools(t *testing.T) {
	// given
	spy := &manaSpy{pool: missiles.PlayerMana{Mana: 3, MaxMana: 800, ManaBase: 7, MaxManaBase: 900}}
	fx := newAddFixture(t, addScript{source: 0})
	fx.world.ProcessCallbacks = spy.manaCallbacks()

	// when
	missiles.AddMagi(fx.world, fx.missile, fx.param)

	// then
	if spy.setMana != 800 || spy.setBase != 900 {
		t.Fatalf("mana/base = %d/%d, want 800/900", spy.setMana, spy.setBase)
	}

	if !fx.missile.DelFlag || spy.redraws != 1 {
		t.Fatalf("delFlag/redraws = %v/%d, want true/1", fx.missile.DelFlag, spy.redraws)
	}
}

func TestAddStealMana_DrainsTheTrappedPlayer(t *testing.T) {
	// given
	spy := &manaSpy{pool: missiles.PlayerMana{Mana: 500, MaxMana: 300, ManaBase: 40, MaxManaBase: 400}}
	fx := newAddFixture(t, addScript{start: [2]int{40, 40}, source: -1})
	fx.world.ProcessCallbacks = spy.manaCallbacks()
	fx.world.DPlayer[41][40] = 3 // player 2 stands one tile south

	// when
	missiles.AddStealMana(fx.world, fx.missile, fx.param)

	// then
	if spy.setMana != 0 {
		t.Fatalf("mana = %d, want 0", spy.setMana)
	}

	if spy.setBase != 0+400-300 {
		t.Fatalf("base = %d, want %d (rebase keeps the equipment bonus)", spy.setBase, 0+400-300)
	}

	if spy.calcInv != 1 || spy.redraws != 1 || spy.sfxName != "Pig" {
		t.Fatalf("calc/redraw/sfx = %d/%d/%q, want 1/1/Pig", spy.calcInv, spy.redraws, spy.sfxName)
	}

	if spy.sfxTile != (missiles.TilePosition{X: 40, Y: 41}) {
		t.Fatalf("sfx tile = %v, want the trapped tile", spy.sfxTile)
	}

	if !fx.missile.DelFlag {
		t.Fatal("missile not deleted")
	}
}

func TestAddStealMana_NoPlayerMeansNoDrain(t *testing.T) {
	// given
	spy := &manaSpy{}
	fx := newAddFixture(t, addScript{start: [2]int{40, 40}, source: -1})
	fx.world.ProcessCallbacks = spy.manaCallbacks()

	// when
	missiles.AddStealMana(fx.world, fx.missile, fx.param)

	// then
	if spy.calcInv != 0 || spy.sfxName != "" {
		t.Fatalf("calc/sfx = %d/%q, want no sink call", spy.calcInv, spy.sfxName)
	}

	if !fx.missile.DelFlag {
		t.Fatal("missile not deleted")
	}
}

func TestAddReflect_AddsSpellLevelTimesCharacterLevel(t *testing.T) {
	// given
	cases := []struct {
		name       string
		source     int
		spllvl     int
		level      int
		reflection int
		want       int
	}{
		{"unlevelled reflect adds twice the level", 0, 0, 5, 100, 110},
		{"spell level replaces the flat 2", 0, 3, 5, 100, 115},
		{"the uint16 ceiling drops the gain", 0, 2, 5, 65530, 65530},
		{"a trap adds nothing", -1, 3, 5, 100, 100},
	}

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()
			runReflect(t, tc.source, tc.spllvl, tc.level, tc.reflection, tc.want)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// runReflect runs one AddReflect case.
func runReflect(t *testing.T, source, spllvl, level, reflection, want int) {
	t.Helper()

	// given
	spy := &manaSpy{reflectSet: reflection}
	fx := newAddFixture(t, addScript{source: source, spllvl: spllvl, level: level})
	fx.world.ProcessCallbacks = spy.manaCallbacks()
	wantSent := want
	if source < 0 {
		wantSent = 0 // a trap caster never sends
	}

	// when
	missiles.AddReflect(fx.world, fx.missile, fx.param)

	// then
	if spy.reflectSet != want {
		t.Fatalf("reflections = %d, want %d", spy.reflectSet, want)
	}

	if spy.sentTotal != wantSent {
		t.Fatalf("sent total = %d, want %d", spy.sentTotal, wantSent)
	}

	if !fx.missile.DelFlag {
		t.Fatal("missile not deleted")
	}
}
