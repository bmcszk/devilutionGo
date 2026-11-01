package stores_test

import (
	"slices"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/stores"
	"github.com/bmcszk/devilutionGo/internal/engine"
)

// premiumBoundsCases is the out-of-range slot table of both shelves
// (items.cpp:4437-4438, stores.h:29-30).
var premiumBoundsCases = []struct {
	name string
	env  stores.Env
	idx  int
}{
	{name: "vanilla slot 6", idx: stores.NumSmithItems},
	{name: "vanilla slot -1", idx: -1},
	{name: "hellfire slot 15", env: stores.Env{Hellfire: true}, idx: stores.NumSmithItemsHF},
}

// premiumItem is the shelf payload: a spawn serial so the slot shuffle is
// visible, the plvl SpawnOnePremium was handed, and the seed it drew.
type premiumItem struct {
	plvl   int
	serial int
	seed   int32
}

// tape records the spawn calls a rotation makes, in order.
type tape struct {
	plvls []int
	seeds []int32
}

// spawn is the SpawnOnePremium stand-in: it draws one LCG seed per item, so the
// tape pins both the call order and the RNG consumption
// (items.cpp:1937-1954).
func (tp *tape) spawn(plvl int) premiumItem {
	item := premiumItem{
		plvl:   plvl,
		serial: len(tp.plvls),
		seed:   engine.AdvanceRndSeed(),
	}
	tp.plvls = append(tp.plvls, plvl)
	tp.seeds = append(tp.seeds, item.seed)

	return item
}

// TestSpawnPremiumFillsVanillaShelf pins the six-slot vanilla fill at
// PremiumItemLevel 1: each slot spawns at 1+itemLevelAdd[i], so the plvls run
// 0,0,1,1,2,3 (items.cpp:4442-4465, itemLevelAdd at :362-370).
func TestSpawnPremiumFillsVanillaShelf(t *testing.T) {
	t.Parallel()

	// given a fresh shelf and a level-1 character, so only the fill runs
	tp := &tape{}
	stock := stores.PremiumStock[premiumItem]{Level: 1}

	// when the shelf is filled
	stores.SpawnPremium(&stock, 1, stores.Env{}, tp.spawn)

	// then it holds six slots spawned bottom-up
	if len(stock.Items) != stores.NumSmithItems {
		t.Fatalf("shelf size = %d, want %d", len(stock.Items), stores.NumSmithItems)
	}

	checkSerials(t, stock.Items, []int{0, 1, 2, 3, 4, 5})
	checkTape(t, tp.plvls, []int{0, 0, 1, 1, 2, 3})
}

// TestSpawnPremiumFillsHellfireShelf pins the fifteen-slot Hellfire fill: the
// per-slot deltas run -1..3, so the plvls climb in four steps
// (items.cpp:4448, itemLevelAddHf at :372-388).
func TestSpawnPremiumFillsHellfireShelf(t *testing.T) {
	t.Parallel()

	// given a fresh shelf and a level-1 Hellfire character, so only the fill runs
	tp := &tape{}
	stock := stores.PremiumStock[premiumItem]{Level: 1}

	// when the shelf is filled
	stores.SpawnPremium(&stock, 1, stores.Env{Hellfire: true}, tp.spawn)

	// then it holds fifteen slots at the Hellfire deltas
	if len(stock.Items) != stores.NumSmithItemsHF {
		t.Fatalf("shelf size = %d, want %d", len(stock.Items), stores.NumSmithItemsHF)
	}

	checkTape(t, tp.plvls, []int{0, 0, 0, 1, 1, 1, 1, 2, 2, 2, 2, 3, 3, 4, 4})
}

// TestSpawnPremiumRotatesVanillaShelf pins the vanilla level-up rotation from
// PremiumItemLevel 1 to a level-5 character. Each step moves slots 2..4 down
// onto 0..2, copies 5 onto 4, then respawns slots 3 and 5, so from the fill
// layout [0 1 2 3 4 5] the shelf ends at [7 10 9 12 11 13]
// (items.cpp:4454-4469).
func TestSpawnPremiumRotatesVanillaShelf(t *testing.T) {
	t.Parallel()

	// given a filled vanilla shelf at PremiumItemLevel 1
	tp := &tape{}
	stock := stores.PremiumStock[premiumItem]{Level: 1}
	stores.SpawnPremium(&stock, 1, stores.Env{}, tp.spawn)

	// when the character level rises to 5
	stores.SpawnPremium(&stock, 5, stores.Env{}, tp.spawn)

	// then every step shifted the shelf and respawned the two tail slots
	if stock.Level != 5 {
		t.Fatalf("shelf level = %d, want 5", stock.Level)
	}

	checkSerials(t, stock.Items, []int{7, 10, 9, 12, 11, 13})
	checkTape(t, tp.plvls, []int{0, 0, 1, 1, 2, 3, 2, 4, 3, 5, 4, 6, 5, 7})
}

// TestSpawnPremiumRotatesHellfireShelf pins the Hellfire rotation for one
// level step: slots 3..12 move down onto 0..9, slot 13 lands on 11, slot 14 on
// 13, then slots 10, 12 and 14 respawn (items.cpp:4457-4464).
func TestSpawnPremiumRotatesHellfireShelf(t *testing.T) {
	t.Parallel()

	// given a filled Hellfire shelf at PremiumItemLevel 1
	tp := &tape{}
	stock := stores.PremiumStock[premiumItem]{Level: 1}
	stores.SpawnPremium(&stock, 1, stores.Env{Hellfire: true}, tp.spawn)

	// when the character level rises by one
	stores.SpawnPremium(&stock, 2, stores.Env{Hellfire: true}, tp.spawn)

	// then the shelf shifted twelve slots and respawned the three tail slots
	checkSerials(t, stock.Items,
		[]int{3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 15, 13, 16, 14, 17})
	checkTape(t, tp.plvls, []int{
		0, 0, 0, 1, 1, 1, 1, 2, 2, 2, 2, 3, 3, 4, 4, // fill
		3, 4, 5, // one rotation step
	})
}

// TestSpawnPremiumConsumesSeededTape pins the RNG consumption of a vanilla
// fill plus a rotation to level 5: fourteen spawns, one AdvanceRndSeed each,
// in spawn order. The seed literals are the LCG tape seeded with 1234
// (random.cpp:17 generator constants, :89-98 AdvanceRndSeed).
func TestSpawnPremiumConsumesSeededTape(t *testing.T) {
	// given the LCG seeded at 1234
	engine.SetRndSeed(1234)

	// when a vanilla shelf fills and rotates up to level 5
	tp := &tape{}
	stock := stores.PremiumStock[premiumItem]{Level: 1}
	stores.SpawnPremium(&stock, 5, stores.Env{}, tp.spawn)

	// then exactly one seed was drawn per spawn, in spawn order
	if len(tp.seeds) != len(tp.plvls) {
		t.Fatalf("drew %d seeds for %d spawns", len(tp.seeds), len(tp.plvls))
	}

	want := []int32{
		2058552453, 238187912, 705807655, 1770671342, 155291321, 904336820,
		423864763, 1064770998, 1766679891, 1410548784, 1873001969,
		1199587098, 703265889, 1638256148,
	}

	if !slices.Equal(tp.seeds, want) {
		t.Fatalf("seed tape = %v, want %v", tp.seeds, want)
	}

	// and each shelf slot carries the seed drawn for it
	for i, item := range stock.Items {
		if item.seed != tp.seeds[item.serial] {
			t.Fatalf("slot %d seed = %d, want %d", i, item.seed, tp.seeds[item.serial])
		}
	}
}

// TestReplacePremiumRefillsOneSlot pins ReplacePremium: the bought slot respawns
// at the character level plus that slot's delta, and no other slot moves
// (items.cpp:4435-4440).
func TestReplacePremiumRefillsOneSlot(t *testing.T) {
	t.Parallel()

	// given a filled vanilla shelf
	tp := &tape{}
	stock := stores.PremiumStock[premiumItem]{Level: 1}
	stores.SpawnPremium(&stock, 1, stores.Env{}, tp.spawn)

	// when slot 2 is bought by a level-5 character
	if !stores.ReplacePremium(&stock, 2, 5, stores.Env{}, tp.spawn) {
		t.Fatal("ReplacePremium reported an out-of-range slot")
	}

	// then only slot 2 respawned, at level 5 + itemLevelAdd[2]
	checkSerials(t, stock.Items, []int{0, 1, 6, 3, 4, 5})
	checkTape(t, tp.plvls[6:], []int{5})
}

// TestReplacePremiumRejectsOutOfRange pins the bounds check on both shelves:
// the vanilla shelf rejects slot 6 and the Hellfire shelf rejects slot 15
// (items.cpp:4437-4438, stores.h:29-30).
func TestReplacePremiumRejectsOutOfRange(t *testing.T) {
	t.Parallel()

	for _, tc := range premiumBoundsCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// given a shelf and a spawn recorder of this case's own
			local := &tape{}
			stock := stores.PremiumStock[premiumItem]{Level: 1}

			// when the out-of-range slot is bought
			accepted := stores.ReplacePremium(&stock, tc.idx, 5, tc.env, local.spawn)

			// then the slot is refused and nothing was spawned
			assertValue(t, accepted, false)
			assertValue(t, len(local.plvls), 0)
		})
	}
}

// TestNumberPremiumItems pins the shelf-size rule of both game modes
// (items.cpp:4445, loadsave.cpp:2489-2494).
func TestNumberPremiumItems(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		env  stores.Env
		want int
	}{
		{name: "vanilla", want: stores.NumSmithItems},
		{name: "hellfire", env: stores.Env{Hellfire: true}, want: stores.NumSmithItemsHF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when the mode's shelf size is asked for
			got := stores.NumberPremiumItems(tt.env)

			// then it is that mode's slot count and the delta table matches it
			assertValue(t, got, tt.want)
			assertValue(t, len(stores.PremiumDeltas(tt.env)), tt.want)
		})
	}
}

// TestResetPremiumRestoresInitStores pins InitStores' premium lines: the shelf
// empties and PremiumItemLevel returns to one, so the next fill starts over at
// the bottom slot deltas (stores.cpp:2145-2161).
func TestResetPremiumRestoresInitStores(t *testing.T) {
	t.Parallel()

	// given a filled and rotated shelf
	tp := &tape{}
	stock := stores.PremiumStock[premiumItem]{Level: 1}
	stores.SpawnPremium(&stock, 5, stores.Env{}, tp.spawn)

	// when the stores are initialised again
	stores.ResetPremium(&stock)

	// then the shelf is empty at level 1
	if len(stock.Items) != 0 {
		t.Fatalf("shelf size = %d, want 0", len(stock.Items))
	}

	if stock.Level != 1 {
		t.Fatalf("shelf level = %d, want 1", stock.Level)
	}

	// and the next fill spawns from the bottom deltas again
	tp.plvls = nil
	stores.SpawnPremium(&stock, 1, stores.Env{}, tp.spawn)
	checkTape(t, tp.plvls, []int{0, 0, 1, 1, 2, 3})
}

// serials flattens a shelf to its spawn serials, the slot-order fingerprint.
func serials(items []premiumItem) []int {
	out := make([]int, len(items))
	for i, item := range items {
		out[i] = item.serial
	}

	return out
}

// checkSerials compares the shelf layout against the hand-computed expectation.
func checkSerials(t *testing.T, items []premiumItem, want []int) {
	t.Helper()

	if got := serials(items); !slices.Equal(got, want) {
		t.Fatalf("shelf layout = %v, want %v", got, want)
	}
}

// checkTape compares the recorded spawn arguments against the expectation.
func checkTape(t *testing.T, got, want []int) {
	t.Helper()

	if !slices.Equal(got, want) {
		t.Fatalf("plvl tape = %v, want %v", got, want)
	}
}
