package itemgen

import (
	"errors"
	"fmt"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/txtdata"
)

// errNoBaseItem is the missing itemdat row for a unique's UIItemId, which
// upstream assumes can never happen (the idx++ walk would run off the table).
var errNoBaseItem = errors.New("no itemdat row for the unique base item")

// SpawnDeps is everything SpawnItem's generation tail reads from the game.
type SpawnDeps struct {
	Items       []itemdat.ItemData
	Uniques     []itemdat.UniqueItemData
	Prefixes    []Affix
	Suffixes    []Affix
	Spells      []txtdata.SpellData
	Env         SpellEnv
	Dropped     map[int]bool
	Multiplayer bool
	Rnd         func(int32) int32
	Discard     func(int)
	NextSeed    func() uint32
}

// MonsterItem is the SpawnItem outcome (items.cpp:3422-3481).
type MonsterItem struct {
	Dropped bool // false when the plan dropped nothing
	Plan    DropPlan
	Setup   SetupPlan
	Unique  RandomUniqueOutcome
}

// UniqueSpawn is the SpawnUnique outcome (items.cpp:3195-3228).
type UniqueSpawn struct {
	BaseIndex int   // the itemdat row whose ItemID matches the unique's
	Err       error // non-nil when no itemdat row carries the unique's ItemID
	// Direct is DIFF_NORMAL's GetItemAttrs + GetUniqueItem path; otherwise the
	// setup ran the pool pick at curlv*2 with uper 15 plus the unique retry.
	Direct bool
	Setup  SetupPlan
	Unique RandomUniqueOutcome
}

// SpawnUnique ports SpawnUnique (items.cpp:3195-3228) for item generation: the
// unique row picks the base itemdat row, a NORMAL-difficulty game applies the
// unique straight onto the base attributes, and every other difficulty reruns
// the droppable pool pick at curlv*2 with uper 15 and then the unique retry.
func SpawnUnique(deps SpawnDeps, uid, level, difficulty int, seed uint32) (UniqueSpawn, error) {
	if uid < 0 || uid >= len(deps.Uniques) {
		return UniqueSpawn{}, fmt.Errorf("unique %d out of range: %w", uid, errUniqueNotInValid)
	}
	base, ok := uniqueBaseIndex(deps.Items, deps.Uniques[uid])
	if !ok {
		return UniqueSpawn{}, fmt.Errorf("unique %d: %w", uid, errNoBaseItem)
	}

	out := UniqueSpawn{BaseIndex: base, Direct: difficulty == DifficultyNormal}
	if out.Direct {
		out.Setup = directUniqueSetup(deps, uid, base, level)

		return out, nil
	}

	// Higher difficulties pick a droppable item of the same itype and force
	// uper 15 on it (:3217-3222).
	drop, err := RndTypeItems(deps.Items, deps.Items[base].ItemType, "", level*2, DropEnv{}, deps.Rnd)
	if err != nil {
		return UniqueSpawn{}, fmt.Errorf("unique %d base pick: %w", uid, err)
	}
	setup, err := SetupAllItems(deps.Items, deps.Uniques, deps.Prefixes, deps.Suffixes, SetupRequest{
		ItemIndex: drop, Seed: seed, Level: level * 2, Uper: 15, OnlyGood: true,
		Spells: deps.Spells, Env: deps.Env,
	}, deps.Rnd, deps.Discard)
	if err != nil {
		return UniqueSpawn{}, fmt.Errorf("unique %d setup: %w", uid, err)
	}
	retry, err := RandomUnique(deps.randomUniqueEnv(), RandomUniqueRequest{
		CreateInfo: setup.CreateInfo, MiscID: deps.Items[drop].MiscID,
		BaseItemID: deps.Items[drop].ID, ItemIndex: drop, Seed: seed,
		Level: level * 2, Uper: 15, OnlyGood: true,
		Spells: deps.Spells, Env: deps.Env,
	})
	if err != nil {
		return UniqueSpawn{}, fmt.Errorf("unique %d retry: %w", uid, err)
	}
	out.Setup, out.Unique = setup, retry

	return out, nil
}

// SpawnItem ports SpawnItem (items.cpp:3422-3481) end to end: the T_MASK /
// T_NODROP / T_UNIQ decode and mushroom override land in the DropPlan, the
// T_UNIQ treasure takes the SpawnUnique short-circuit, and every other drop
// feeds its level/uper/onlygood into the SetupAllItems tail.
func SpawnItem(
	deps SpawnDeps,
	monster MonsterDrop,
	brain BrainQuest,
	env DropEnv,
	seed uint32,
) (MonsterItem, error) {
	plan, err := PlanDrop(deps.Items, monster, env, brain, deps.Rnd)
	if err != nil {
		return MonsterItem{}, fmt.Errorf("spawn plan: %w", err)
	}
	if plan.UniqueIndex >= 0 {
		// dropsSpecialTreasure && !UseMultiplayerQuests (:3429-3433).
		unique, uniqueErr := SpawnUnique(deps, plan.UniqueIndex, plan.Level, DifficultyNormal, seed)
		if uniqueErr != nil {
			return MonsterItem{Plan: plan}, uniqueErr
		}
		out := MonsterItem{Dropped: true, Plan: plan, Setup: unique.Setup, Unique: unique.Unique}
		if unique.Direct { // the direct path spawns the unique item itself
			out.Unique = RandomUniqueOutcome{UID: plan.UniqueIndex, Plan: unique.Setup}
		}

		return out, nil
	}
	item, err := SpawnMonsterItem(deps, plan, seed)
	item.Plan = plan

	return item, err
}

// SpawnMonsterItem ports the SetupAllItems / TryRandomUniqueItem pair SpawnItem
// runs once it has an item index: the plan's raw monster level, uper and
// onlygood seed the setup, then the unique retry runs on the fresh item
// (items.cpp:3469-3471).
func SpawnMonsterItem(deps SpawnDeps, plan DropPlan, seed uint32) (MonsterItem, error) {
	if plan.ItemIndex == ItemNone || plan.NoDrop {
		return MonsterItem{Plan: plan}, nil
	}
	item := deps.Items[plan.ItemIndex]

	req := SetupRequest{
		ItemIndex: plan.ItemIndex,
		Seed:      seed,
		Level:     plan.Level,
		Uper:      plan.Uper,
		OnlyGood:  plan.OnlyGood,
		Spells:    deps.Spells,
		Env:       deps.Env,
	}
	setup, err := SetupAllItems(deps.Items, deps.Uniques, deps.Prefixes, deps.Suffixes, req, deps.Rnd, deps.Discard)
	if err != nil {
		return MonsterItem{}, fmt.Errorf("setup %q: %w", item.ID, err)
	}

	unique, err := RandomUnique(deps.randomUniqueEnv(), RandomUniqueRequest{
		CreateInfo: setup.CreateInfo,
		MiscID:     item.MiscID,
		BaseItemID: item.ID,
		ItemIndex:  plan.ItemIndex,
		Seed:       seed,
		Level:      plan.Level,
		Uper:       plan.Uper,
		OnlyGood:   plan.OnlyGood,
		Spells:     deps.Spells,
		Env:        deps.Env,
	})
	if err != nil {
		return MonsterItem{}, fmt.Errorf("unique retry for %q: %w", item.ID, err)
	}

	return MonsterItem{Dropped: true, Plan: plan, Setup: setup, Unique: unique}, nil
}

func (d SpawnDeps) randomUniqueEnv() RandomUniqueEnv {
	return RandomUniqueEnv{
		Uniques:     d.Uniques,
		Dropped:     d.Dropped,
		Multiplayer: d.Multiplayer,
		Rnd:         d.Rnd,
		Discard:     d.Discard,
		NextSeed:    d.NextSeed,
		Setup: func(req SetupRequest) (SetupPlan, error) {
			return SetupAllItems(d.Items, d.Uniques, d.Prefixes, d.Suffixes, req, d.Rnd, d.Discard)
		},
	}
}

// uniqueBaseIndex is SpawnUnique's `while (AllItemsList[idx].iItemId !=
// UniqueItems[uid].UIItemId) idx++` walk (:3207-3208).
// uniqueBaseIndex is SpawnUnique's `while (AllItemsList[idx].iItemId !=
// UniqueItems[uid].UIItemId) idx++` walk (items.cpp:3207-3208), keyed on the
// itemdat uniqueBaseItem column.
func uniqueBaseIndex(items []itemdat.ItemData, unique itemdat.UniqueItemData) (int, bool) {
	for i := range items {
		if uniqueKey(items[i]) == unique.UniqueBaseItem {
			return i, true
		}
	}

	return 0, false
}

// directUniqueSetup is the DIFF_NORMAL branch (:3210-3214): GetItemAttrs on the
// base row, then GetUniqueItem over it, with the unique durability skip left to
// the caller's SetupItem.
// directUniqueSetup is the DIFF_NORMAL branch (items.cpp:3210-3214):
// GetItemAttrs on the base row, then GetUniqueItem over it, with the unique
// durability skip left to the caller's SetupItem.
func directUniqueSetup(deps SpawnDeps, uid, base, level int) SetupPlan {
	item := deps.Items[base]
	plan := SetupPlan{
		CreateInfo: level | CreateInfoUnique, UID: uid, Quality: QualityUnique,
		Cursor: item.CursorGraphic, Value: item.Value, Ivalue: item.Value,
	}
	if deps.Env.Hellfire {
		plan.Buff |= BuffHellfire
	}
	applied, err := ApplyUniqueItem(deps.Uniques[uid], uid, item.MiscID, plan.CreateInfo, deps.Rnd)
	if err != nil {
		return plan
	}
	plan.Name, plan.UID, plan.Powers = applied.Name, applied.UID, applied.Powers
	plan.CreateInfo, plan.Ivalue = applied.CreateInfo, applied.Value
	if applied.Cursor != "" {
		plan.Cursor = applied.Cursor
	}

	return plan
}
