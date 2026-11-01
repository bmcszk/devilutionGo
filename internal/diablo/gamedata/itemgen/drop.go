package itemgen

import (
	"errors"
	"fmt"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemdat"
)

// Treasure bits of a monstdat treasure value (tables/monstdat.h:86-88).
const (
	TreasureMask   = 0xFFF
	TreasureNoDrop = 0x4000
	TreasureUnique = 0x8000
)

// ItemNone is IDI_NONE (tables/itemdat.h:86), the "no item" index.
const ItemNone = -1

// itemGoldRow is IDI_GOLD, the first itemdat row (tables/itemdat.h:25).
const itemGoldRow = 0

// Pinned itemdat row indices baked into IsItemAvailable (itemdat.tsv row
// order equals the _item_indexes enum order).
const (
	rowBrain     = 18
	rowSorcerer  = 5
	rowMapOfDoom = 22
	rowLgtForge  = 32
	rowOil       = 35
	rowBardSword = 37
	rowBardDagg  = 38
	rowGreySuit  = 47
	rowScroll92  = 92
)

// spawnArmorFirst/spawnArmorLast are the gbIsSpawn armor block (items.cpp:2378).
const (
	spawnArmorFirst = 62
	spawnArmorLast  = 70
)

// errEmptyDropPool is the empty-pool case; upstream dereferences upper_bound's
// end iterator there (items.cpp:1374), which is UB.
var errEmptyDropPool = errors.New("no droppable item")

// DropEnv is the game state the item pools and SpawnItem read.
type DropEnv struct {
	Hellfire    bool
	Spawn       bool
	Multiplayer bool
	UseMPQuests bool
	TestBard    bool
}

// BrainQuest is the Q_MUSHROOM state SpawnItem checks (items.cpp:3427).
type BrainQuest struct {
	Active  bool // _qactive == QUEST_ACTIVE
	Given   bool // _qvar1 == QS_MUSHGIVEN
	SendMsg bool
}

// MonsterDrop is the monster state SpawnItem reads (items.cpp:3425-3464).
type MonsterDrop struct {
	Treasure int
	Unique   bool
	// Level is monster.level(nDifficulty), the difficulty-adjusted level the
	// pool picks index on (items.cpp:3435, 3462).
	Level int
	// DataLevel is the raw monstdat level SpawnItem hands SetupAllItems as
	// mLevel (items.cpp:3467).
	DataLevel int
}

// DropPlan is the SpawnItem decision (items.cpp:3422-3471).
type DropPlan struct {
	UniqueIndex int  // >= 0: SpawnUnique(treasure & T_MASK) short-circuit (:3430)
	ItemIndex   int  // chosen itemdat row, ItemNone for no drop
	Brain       bool // quest brain replaces the drop (:3438-3446)
	BrainExtra  bool // multiplayer: spawn the brain as an extra item (:3448-3454)
	NoDrop      bool
	OnlyGood    bool
	Uper        int
	// Level is the raw monster level SetupAllItems receives as mLevel
	// (items.cpp:3467).
	Level int
}

// IsItemAvailable ports items.cpp:2373-2404. i is the itemdat row.
func IsItemAvailable(items []itemdat.ItemData, i int, env DropEnv) bool {
	if i < 0 || i >= len(items) {
		return false
	}
	if env.Spawn && unavailableOnSpawn(i) {
		return false
	}
	return env.Hellfire || vanillaAvailable(i) || isBardItem(i, env.TestBard)
}

// RndItemForMonsterLevel ports items.cpp:3256-3266.
func RndItemForMonsterLevel(
	items []itemdat.ItemData,
	monsterLevel int,
	env DropEnv,
	rnd func(int32) int32,
) (int, error) {
	if rnd(100) > 40 {
		return ItemNone, nil
	}
	if rnd(100) > 25 {
		return itemGoldRow, nil
	}

	idx, err := GetItemIndexForDroppableItem(items, env, true, func(item itemdat.ItemData) bool {
		return item.MinMonsterLevel <= monsterLevel
	}, rnd)
	if err != nil {
		return ItemNone, fmt.Errorf("monster level %d drop: %w", monsterLevel, err)
	}

	return idx, nil
}

// RndUItem ports items.cpp:1376-1390: a unique/treasure-killer base item pick.
func RndUItem(items []itemdat.ItemData, monsterLevel int, env DropEnv, rnd func(int32) int32) (int, error) {
	idx, err := GetItemIndexForDroppableItem(items, env, false, func(item itemdat.ItemData) bool {
		if item.ItemType == "Misc" && item.MiscID == "BOOK" {
			return true
		}
		if monsterLevel < item.MinMonsterLevel {
			return false
		}

		return item.ItemType != "Gold" && item.ItemType != "Misc"
	}, rnd)
	if err != nil {
		return ItemNone, fmt.Errorf("unique base item at level %d: %w", monsterLevel, err)
	}

	return idx, nil
}

// GetItemIndexForDroppableItem ports items.cpp:1353-1375: the cumulative-weight
// pick over the rows passing IsItemAvailable and the caller's filter.
//
//nolint:revive // considerDropRate is upstream's parameter, not control coupling.
func GetItemIndexForDroppableItem(
	items []itemdat.ItemData,
	env DropEnv,
	considerDropRate bool,
	okay func(itemdat.ItemData) bool,
	rnd func(int32) int32,
) (int, error) {
	indexes := make([]int, 0, len(items))
	weights := make([]int, 0, len(items))
	cumulative := 0
	for i, item := range items {
		if !IsItemAvailable(items, i, env) || item.DropRate == 0 || !okay(item) {
			continue
		}
		if !env.Multiplayer && (item.Spell == "RESURRECT" || item.Spell == "HEALOTHER") {
			continue
		}
		if considerDropRate {
			cumulative += item.DropRate
		} else {
			cumulative++
		}
		indexes = append(indexes, i)
		weights = append(weights, cumulative)
	}
	if len(indexes) == 0 {
		return ItemNone, errEmptyDropPool
	}

	target := max(rnd(int32(cumulative)), 0) // RandomIntLessThan (:1372)

	return indexes[upperBound(weights, int(target))], nil
}

// PlanDrop ports SpawnItem's decision tree (items.cpp:3422-3466).
func PlanDrop(
	items []itemdat.ItemData,
	monster MonsterDrop,
	env DropEnv,
	brain BrainQuest,
	rnd func(int32) int32,
) (DropPlan, error) {
	plan := DropPlan{ItemIndex: ItemNone, UniqueIndex: UniqueNone, OnlyGood: true, Uper: 1}
	dropsTreasure := monster.Treasure&TreasureUnique != 0

	if dropsTreasure && !env.UseMPQuests {
		plan.UniqueIndex = monster.Treasure & TreasureMask

		return plan, nil
	}
	if monster.Unique || dropsTreasure {
		return planUniqueDrop(items, monster, env, rnd)
	}
	if brain.Active && brain.Given && !env.Multiplayer {
		plan.Brain = true
		plan.ItemIndex = rowBrain

		return plan, nil
	}

	return planNormalDrop(items, monster, env, brain, rnd)
}

// planUniqueDrop is the unique / special-treasure base item pick (:3434-3437).
func planUniqueDrop(
	items []itemdat.ItemData,
	monster MonsterDrop,
	env DropEnv,
	rnd func(int32) int32,
) (DropPlan, error) {
	idx, err := RndUItem(items, monster.Level, env, rnd)
	if err != nil {
		return DropPlan{}, fmt.Errorf("unique drop: %w", err)
	}

	return DropPlan{ItemIndex: idx, OnlyGood: true, Uper: planUper(monster), Level: monster.DataLevel}, nil
}

// planNormalDrop is the no-drop / monster-level branch (:3448-3466).
func planNormalDrop(
	items []itemdat.ItemData,
	monster MonsterDrop,
	env DropEnv,
	brain BrainQuest,
	rnd func(int32) int32,
) (DropPlan, error) {
	plan := DropPlan{ItemIndex: ItemNone, UniqueIndex: UniqueNone, OnlyGood: true, Uper: 1}
	plan.BrainExtra = brain.Active && brain.Given && env.Multiplayer && brain.SendMsg
	if monster.Treasure&TreasureNoDrop != 0 {
		plan.NoDrop = true

		return plan, nil
	}

	plan.OnlyGood = false
	plan.Uper = planUper(monster)
	plan.Level = monster.DataLevel
	idx, err := RndItemForMonsterLevel(items, monster.Level, env, rnd)
	if err != nil {
		return DropPlan{}, fmt.Errorf("normal drop: %w", err)
	}
	plan.ItemIndex = idx

	return plan, nil
}

// planUper is `uper = monster.isUnique() ? 15 : 1` (items.cpp:3464).
func planUper(monster MonsterDrop) int {
	if monster.Unique {
		return 15
	}

	return 1
}

// upperBound returns the first weight strictly above target (std::upper_bound).
func upperBound(weights []int, target int) int {
	for i, w := range weights {
		if w > target {
			return i
		}
	}

	return len(weights) - 1
}

// unavailableOnSpawn is the gbIsSpawn block of IsItemAvailable (items.cpp:2378-2383).
func unavailableOnSpawn(i int) bool {
	if i >= spawnArmorFirst && i <= spawnArmorLast {
		return true // Medium and heavy armors
	}
	switch i {
	case 105, 107, 108, 110, 111, 113:
		return true // Unavailable scrolls
	}

	return false
}

// isBardItem is the testBard carve-out of IsItemAvailable (:2402-2403).
func isBardItem(i int, testBard bool) bool {
	return testBard && (i == rowBardSword || i == rowBardDagg)
}

// vanillaAvailable is the Diablo-only filter of IsItemAvailable (:2389-2397).
func vanillaAvailable(i int) bool {
	if i == rowMapOfDoom || i == rowLgtForge || i == rowSorcerer || i == rowScroll92 {
		return false
	}

	return !exclusiveRange(i)
}

// exclusiveRange covers the Hellfire-only, oil and rune row blocks.
func exclusiveRange(i int) bool {
	return (i >= rowOil && i <= rowGreySuit) || (i >= 83 && i <= 86) || (i >= 161 && i <= 165)
}
