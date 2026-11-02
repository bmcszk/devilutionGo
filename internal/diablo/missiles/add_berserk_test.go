// Package missiles_test pins the Hellfire Berserk add body (missiles.cpp:
// 1295-1344).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
)

var berserkSkipCases = []berserkSkipCase{
	{"a minion is skipped", func(cb *missiles.ProcessCallbacks) {
		cb.MonsterIsPlayerMinion = func(int) bool { return true }
	}},
	{"an already berserked monster is skipped", func(cb *missiles.ProcessCallbacks) {
		cb.MonsterFlags = func(int) int { return 1 << 11 }
	}},
	{"a unique is skipped", func(cb *missiles.ProcessCallbacks) {
		cb.MonsterIsUnique = func(int) bool { return true }
	}},
	{"Diablo is skipped", func(cb *missiles.ProcessCallbacks) {
		cb.MonsterAI = func(int) int { return 27 }
	}},
	{"a fading monster is skipped", func(cb *missiles.ProcessCallbacks) {
		cb.MonsterMode = func(int) int { return 8 }
	}},
	{"a charging monster is skipped", func(cb *missiles.ProcessCallbacks) {
		cb.MonsterMode = func(int) int { return 14 }
	}},
	{"a magic-immune monster is skipped", func(cb *missiles.ProcessCallbacks) {
		cb.MonsterResistance = func(int) int { return 1 << 3 }
	}},
	{"a lost resist-magic coin flip is skipped", func(cb *missiles.ProcessCallbacks) {
		cb.MonsterResistance = func(int) int { return 1 << 0 }
	}},
	{"an empty cell is skipped", func(_ *missiles.ProcessCallbacks) {}},
}

// berserkSkipCase pairs an ineligible monster state with its rejection.
type berserkSkipCase struct {
	name      string
	configure func(*missiles.ProcessCallbacks)
}

// berserkSpy records the monster-store writes AddBerserk reaches.
type berserkSpy struct {
	flagsR    int
	flagsSet  bool
	flags     int
	damage    missiles.MonsterDamage
	damageSet bool
	lightTile missiles.TilePosition
	radius    uint8
	monster   int
}

// callbacks wires the spy into a HitWorld with an eligible monster by default.
func (s *berserkSpy) callbacks() missiles.ProcessCallbacks {
	cb := missiles.NoopProcessCallbacks()
	cb.PlayerSpellLevel = func(_, _ int) int { return 3 }
	cb.MonsterFlags = func(int) int { return s.flagsR }
	cb.SetMonsterFlags = func(_, flags int) {
		s.flagsSet = true
		s.flags = flags
	}
	cb.MonsterDamage = func(int) missiles.MonsterDamage {
		return missiles.MonsterDamage{Min: 10, Max: 20, MinSpecial: 30, MaxSpecial: 40}
	}
	cb.SetMonsterDamage = func(_ int, damage missiles.MonsterDamage) {
		s.damageSet = true
		s.damage = damage
	}
	cb.SetMonsterLight = func(monster, _ int) { s.monster = monster }
	cb.AddLight = func(x, y int, radius uint8) int {
		s.lightTile = missiles.TilePosition{X: x, Y: y}
		s.radius = radius

		return 77
	}

	return cb
}

func TestAddBerserk_BerserksTheAimedMonster(t *testing.T) {
	// given
	spy := &berserkSpy{}
	fx := newAddFixture(t, addScript{
		start: [2]int{40, 40}, dst: [2]int{43, 40}, source: 0, spllvl: 1,
		// one d10 per damage field
		rnd: []int{0, 9, 4, 5},
	})
	fx.world.ProcessCallbacks = spy.callbacks()
	fx.world.ProcessCallbacks.GenerateRnd = fx.tape.draw
	fx.world.DMonster[40][43] = 5 // monster 4 stands on the aim tile

	// when
	missiles.AddBerserk(fx.world, fx.missile, fx.param)
	fx.tape.assertConsumed()

	// then
	wantFlags := (1 << 11) | (1 << 5) // MFLAG_BERSERK | MFLAG_GOLEM
	if !spy.flagsSet || spy.flags != wantFlags {
		t.Fatalf("flags = %d (set %v), want %d", spy.flags, spy.flagsSet, wantFlags)
	}

	want := missiles.MonsterDamage{Min: 15, Max: 28, MinSpecial: 40, MaxSpecial: 53}
	if !spy.damageSet || spy.damage != want {
		t.Fatalf("damage = %+v, want %+v", spy.damage, want)
	}

	if spy.radius != 3 {
		t.Fatalf("light radius = %d, want 3 outside a nest", spy.radius)
	}

	if !fx.missile.DelFlag || fx.param.SpellFizzled {
		t.Fatalf("delFlag/fizzled = %v/%v, want true/false", fx.missile.DelFlag, fx.param.SpellFizzled)
	}
}

func TestAddBerserk_NestLevelWidensTheLight(t *testing.T) {
	// given
	spy := &berserkSpy{}
	fx := newAddFixture(t, addScript{start: [2]int{40, 40}, dst: [2]int{43, 40}, source: 0, rnd: []int{0, 0, 0, 0}})
	fx.world.ProcessCallbacks = spy.callbacks()
	fx.world.ProcessCallbacks.GenerateRnd = fx.tape.draw
	fx.world.ProcessCallbacks.IsNestLevel = func() bool { return true }
	fx.world.DMonster[40][43] = 1

	// when
	missiles.AddBerserk(fx.world, fx.missile, fx.param)
	fx.tape.assertConsumed()

	// then
	if spy.radius != 9 {
		t.Fatalf("light radius = %d, want the nest 9", spy.radius)
	}
}

func TestAddBerserk_SkipsIneligibleMonsters(t *testing.T) {
	// given
	cases := berserkSkipCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()
			runBerserkSkip(t, tc.name, tc.configure)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

func TestAddBerserk_TrapCastFizzlesWithoutATarget(t *testing.T) {
	// given
	spy := &berserkSpy{}
	fx := newAddFixture(t, addScript{start: [2]int{40, 40}, dst: [2]int{43, 40}, source: -1})
	fx.world.ProcessCallbacks = spy.callbacks()
	fx.world.DMonster[40][43] = 1

	// when
	missiles.AddBerserk(fx.world, fx.missile, fx.param)

	// then
	if spy.flagsSet || spy.damageSet {
		t.Fatal("a trap cast touched the monster store")
	}

	if !fx.missile.DelFlag || !fx.param.SpellFizzled {
		t.Fatalf("delFlag/fizzled = %v/%v, want true/true", fx.missile.DelFlag, fx.param.SpellFizzled)
	}
}

// runBerserkSkip checks one ineligible-target case leaves the spells fizzled.
func runBerserkSkip(t *testing.T, name string, configure func(*missiles.ProcessCallbacks)) {
	t.Helper()

	// given
	spy := &berserkSpy{}
	fx := newAddFixture(t, addScript{start: [2]int{40, 40}, dst: [2]int{43, 40}, source: 0})
	fx.world.ProcessCallbacks = spy.callbacks()

	// the empty-cell case leaves the grid bare; the coin-flip case scripts
	// a losing draw (odd = false).
	fx.world.ProcessCallbacks.GenerateRnd = func(int) int { return 1 }
	if name != "an empty cell is skipped" {
		fx.world.DMonster[40][43] = 1
	}

	configure(&fx.world.ProcessCallbacks)

	// when
	missiles.AddBerserk(fx.world, fx.missile, fx.param)

	// then
	if spy.flagsSet || spy.damageSet {
		t.Fatalf("%s: monster store touched", name)
	}

	if !fx.param.SpellFizzled {
		t.Fatalf("%s: spell did not fizzle", name)
	}
}
