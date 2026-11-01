package quests_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/enums"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/quests"
)

// poolSeeds pins the pool outcome for ten level-15 seeds.
//
// Method: the pools are drawn from DiabloGenerator seeded with DungeonSeeds[15]
// (quests.cpp:113-133), so the outcome depends only on the seed. The expected
// NOTAVAIL slots were derived by hand from the LCG arithmetic, e.g. seed 1:
//
//	draw 1: state 0x015A4E36, abs 22695478, hi16 346 -> 346 % 2 = 0 -> Q_SKELKING
//	draw 2: state 0x8082A52F, abs 2138921681, hi16 32637 -> 32637 % 3 = 0 -> Q_BUTCHER
//	draw 3: state 0xAAE684BC, abs 1427733316, hi16 21785 -> 21785 % 3 = 2 -> Q_BLOOD
//	draw 4: state 0x0442C2ED, abs 71484141, hi16 1090 -> 1090 % 3 = 1 -> Q_ZHAR
//	draw 5: state 0xAD889112, abs 1383558894, hi16 21111 -> 21111 % 2 = 1 -> Q_WARLORD
//
// Seed 988045466 is vanilla's broken QuestGroup1[-2] path: it discards one draw
// and leaves The Butcher available (quests.cpp:117-127), so it has four picks.
var poolSeeds = []struct {
	seed  uint32
	slots []int8
}{
	{seed: 0, slots: []int8{4, 7, 8, 10, 12}},           // VEIL LTBANNER BLIND ANVIL SKELKING
	{seed: 1, slots: []int8{3, 6, 9, 11, 12}},           // ZHAR BUTCHER BLOOD WARLORD SKELKING
	{seed: 7, slots: []int8{0, 2, 3, 4, 12}},            // ROCK GARBUD ZHAR VEIL SKELKING
	{seed: 42, slots: []int8{2, 8, 10, 11, 12}},         // GARBUD BLIND ANVIL WARLORD SKELKING
	{seed: 123456789, slots: []int8{6, 9, 10, 11, 12}},  // BUTCHER BLOOD ANVIL WARLORD SKELKING
	{seed: 988045466, slots: []int8{0, 1, 4, 12}},       // ROCK MUSHROOM VEIL SKELKING (bug path)
	{seed: 999999937, slots: []int8{2, 8, 10, 11, 12}},  // GARBUD BLIND ANVIL WARLORD SKELKING
	{seed: 2147483648, slots: []int8{0, 6, 10, 11, 13}}, // ROCK BUTCHER ANVIL WARLORD PWATER
	{seed: 4000000000, slots: []int8{1, 7, 9, 11, 13}},  // MUSHROOM LTBANNER BLOOD WARLORD PWATER
	{seed: 4294967295, slots: []int8{1, 4, 7, 9, 12}},   // MUSHROOM VEIL LTBANNER BLOOD SKELKING
}

// testLCG is an independent Borland LCG (engine/random.hpp:17-83) used as the
// pool harness: it re-derives the draws from the algorithm rather than from
// engine/random.go, so a wrong port cannot agree with itself.
type testLCG struct {
	state uint32
}

// seed sets the engine state (DiabloGenerator::DiabloGenerator).
func (g *testLCG) seed(s uint32) { g.state = s }

// advance is AdvanceRndSeed: state = state*0x015A4E35 + 1, then abs(signed).
func (g *testLCG) advance() int32 {
	g.state = g.state*0x015A4E35 + 1
	v := int32(g.state)
	if v == -1<<31 {
		return v
	}

	if v < 0 {
		return -v
	}

	return v
}

// rnd is GenerateRnd (random.hpp:40-57).
func (g *testLCG) rnd(v int32) int32 {
	if v <= 0 {
		return 0
	}

	if v <= 0x7FFF {
		return (g.advance() >> 16) % v
	}

	return g.advance() % v
}

// The six level-15 seed pools mark exactly the hand-derived quests NOTAVAIL
// (quests.cpp:113-133; InitialiseQuestPools is called from InitQuests:97-99).
func TestInitialiseQuestPoolsSeeds(t *testing.T) {
	for _, tc := range poolSeeds {
		t.Run(seedName(tc.seed), func(t *testing.T) {
			// given
			rng := &testLCG{}
			// when
			got := quests.InitialiseQuestPools(tc.seed, rng.seed, rng.rnd)
			// then
			assertPoolSlots(t, tc.seed, got, tc.slots)
		})
	}
}

// The pools feed InitQuests before the fixed starting variables, so a
// pooled-out sentinel quest still gets var2 = 2 (quests.cpp:97-101).
func TestInitQuestsWithPoolsMarksSentinels(t *testing.T) {
	for _, tc := range poolSeeds {
		t.Run(seedName(tc.seed), func(t *testing.T) {
			// given
			rng := &testLCG{}
			pools := quests.PoolOptions{Seed: tc.seed, SetRndSeed: rng.seed, Rnd: rng.rnd}
			// when
			qs, err := quests.InitQuestsWithPools(mustRows(t), enums.SimFlags{}, pools)
			// then
			if err != nil {
				t.Fatalf("quests.InitQuestsWithPools: %v", err)
			}

			assertPoolSlots(t, tc.seed, poolStates(qs), tc.slots)
			assertSentinelVars(t, qs)
		})
	}
}

// PoolOptions without Rnd leaves every quest in the game (option off).
func TestInitQuestsWithoutPoolsKeepsAllQuests(t *testing.T) {
	// when
	qs, err := quests.InitQuestsWithPools(mustRows(t), enums.SimFlags{}, quests.PoolOptions{})
	// then
	if err != nil {
		t.Fatalf("quests.InitQuestsWithPools: %v", err)
	}

	assertPoolSlots(t, 0, poolStates(qs), nil)
}
