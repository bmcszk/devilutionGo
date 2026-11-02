// Package missiles_test pins the Hellfire StealPotions add body
// (missiles.cpp:1392-1456).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
)

// beltSlot is one scripted belt entry.
type beltSlot struct {
	miscID int
	typ    int
}

// beltSpy records the belt writes AddStealPotions reaches.
type beltSpy struct {
	slots   []beltSlot
	removed []int
	reinit  map[int]int
	sounds  int
	redraws int
}

// callbacks wires the spy into a HitWorld with one player on the trap tile.
func (s *beltSpy) callbacks() missiles.ProcessCallbacks {
	cb := missiles.NoopProcessCallbacks()
	cb.PlayerIndexAtPosition = func(tile missiles.TilePosition) (int, bool) {
		return 0, tile == (missiles.TilePosition{X: 40, Y: 40})
	}
	cb.BeltItem = func(_, slot int) missiles.BeltItem {
		if slot >= len(s.slots) {
			return missiles.BeltItem{Type: 1}
		}

		return missiles.BeltItem{Type: s.slots[slot].typ, MiscID: s.slots[slot].miscID}
	}
	cb.RemoveBeltItem = func(_, slot int) { s.removed = append(s.removed, slot) }
	cb.ItemIndexForMisc = func(miscID int) int { return 100 + miscID }
	cb.ReinitBeltItem = func(_, slot, itemIndex int) {
		s.reinit[slot] = itemIndex
	}
	cb.PlaySfxLoc = func(string, int, int) { s.sounds++ }
	cb.RedrawEverything = func() { s.redraws++ }

	return cb
}

func TestAddStealPotions_DemotesFullPotionsAndRemovesNormalOnes(t *testing.T) {
	// given
	cases := []struct {
		name        string
		slot        beltSlot
		wantRemoved int
		wantIx      int
	}{
		{"a full heal becomes a heal", beltSlot{miscID: 2, typ: 0}, 0, 103},
		{"a heal is removed", beltSlot{miscID: 3, typ: 0}, 1, 0},
		{"a mana potion is removed", beltSlot{miscID: 6, typ: 0}, 1, 0},
		{"a full mana becomes a mana potion", beltSlot{miscID: 7, typ: 0}, 0, 106},
	}

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()
			runStealSlot(t, tc.slot, tc.wantRemoved, tc.wantIx)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

func TestAddStealPotions_RollsTheRejuvReplacements(t *testing.T) {
	// given
	cases := []struct {
		name   string
		miscID int
		rolls  []int
		wantIx int
	}{
		// the first draw is the slot's coin flip (1 = not skipped)
		{"a rejuv becomes the coin's pick", 18, []int{1, 0}, 103},
		{"a rejuv takes the second pick on a lost flip", 18, []int{1, 1}, 106},
		{"a full rejuv rolls full mana", 19, []int{1, 0}, 107},
		{"a full rejuv rolls full heal", 19, []int{1, 1}, 102},
		{"a full rejuv rolls a rejuv", 19, []int{1, 2}, 118},
	}

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()
			runStealRejuv(t, tc.miscID, tc.rolls, tc.wantIx)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

func TestAddStealPotions_PlaysOneSoundPerPlayer(t *testing.T) {
	// given
	spy := &beltSpy{slots: []beltSlot{{miscID: 2, typ: 0}, {miscID: 7, typ: 0}}, reinit: map[int]int{}}
	fx := newAddFixture(t, addScript{start: [2]int{40, 40}, dst: [2]int{43, 40}, source: -1, rnd: []int{1, 1}})
	fx.world.ProcessCallbacks = spy.callbacks()
	fx.world.ProcessCallbacks.GenerateRnd = fx.tape.draw

	// when
	missiles.AddStealPotions(fx.world, fx.missile, fx.param)
	fx.tape.assertConsumed()

	// then
	if spy.sounds != 1 || spy.redraws != 1 {
		t.Fatalf("sounds/redraws = %d/%d, want 1/1", spy.sounds, spy.redraws)
	}

	if !fx.missile.DelFlag {
		t.Fatal("stealpotions row not deleted")
	}
}

func TestAddStealPotions_IgnoresNonPotionsAndWonCoins(t *testing.T) {
	// given
	spy := &beltSpy{slots: []beltSlot{{miscID: 3, typ: 1}, {miscID: 3, typ: 0}}, reinit: map[int]int{}}
	fx := newAddFixture(t, addScript{start: [2]int{40, 40}, dst: [2]int{43, 40}, source: -1, rnd: []int{0}})
	fx.world.ProcessCallbacks = spy.callbacks()
	fx.world.ProcessCallbacks.GenerateRnd = fx.tape.draw

	// when
	missiles.AddStealPotions(fx.world, fx.missile, fx.param)
	fx.tape.assertConsumed()

	// then
	if len(spy.removed) != 0 {
		t.Fatalf("removed = %v, want none", spy.removed)
	}

	if spy.sounds != 1 { // the non-potion slot still counts as a spoiled row
		t.Fatalf("sounds = %d, want 1", spy.sounds)
	}

	if spy.redraws != 1 {
		t.Fatalf("redraws = %d, want 1", spy.redraws)
	}
}

// runStealSlot checks one belt slot's spoil with a lost coin flip (the flip
// that keeps the slot in play).
func runStealSlot(t *testing.T, slot beltSlot, wantRemoved, wantIx int) {
	t.Helper()

	// given
	spy := &beltSpy{slots: []beltSlot{slot}, reinit: map[int]int{}}
	fx := newAddFixture(t, addScript{start: [2]int{40, 40}, dst: [2]int{43, 40}, source: -1, rnd: []int{1}})
	fx.world.ProcessCallbacks = spy.callbacks()
	fx.world.ProcessCallbacks.GenerateRnd = fx.tape.draw

	// when
	missiles.AddStealPotions(fx.world, fx.missile, fx.param)
	fx.tape.assertConsumed()

	// then
	assertSpoiledSlot(t, spy, wantRemoved, wantIx)
}

// assertSpoiledSlot checks the removed flag and reinit index of one spoiled
// slot; removed is the case's expectation, not a branch input to the code under
// test.
func assertSpoiledSlot(t *testing.T, spy *beltSpy, wantRemoved, wantIx int) {
	t.Helper()

	if len(spy.removed) != wantRemoved {
		t.Fatalf("removed = %v, want %d slots", spy.removed, wantRemoved)
	}

	if spy.reinit[0] != wantIx {
		t.Fatalf("reinit index = %d, want %d", spy.reinit[0], wantIx)
	}
}

// runStealRejuv checks one rejuv spoil with its scripted replacement rolls.
func runStealRejuv(t *testing.T, miscID int, rolls []int, wantIx int) {
	t.Helper()

	// given
	spy := &beltSpy{slots: []beltSlot{{miscID: miscID, typ: 0}}, reinit: map[int]int{}}
	fx := newAddFixture(t, addScript{start: [2]int{40, 40}, dst: [2]int{43, 40}, source: -1, rnd: rolls})
	fx.world.ProcessCallbacks = spy.callbacks()
	fx.world.ProcessCallbacks.GenerateRnd = fx.tape.draw

	// when
	missiles.AddStealPotions(fx.world, fx.missile, fx.param)
	fx.tape.assertConsumed()

	// then
	if spy.reinit[0] != wantIx {
		t.Fatalf("reinit index = %d, want %d", spy.reinit[0], wantIx)
	}
}
