package itemgen

import (
	"fmt"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/txtdata"
)

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

// MonsterItem is the SpawnItem tail outcome (items.cpp:3469-3474).
type MonsterItem struct {
	Dropped bool // false when the plan dropped nothing
	Setup   SetupPlan
	Unique  RandomUniqueOutcome
}

// SpawnMonsterItem ports the SetupAllItems / TryRandomUniqueItem pair SpawnItem
// runs once it has an item index: the plan's raw monster level, uper and
// onlygood seed the setup, then the unique retry runs on the fresh item
// (items.cpp:3469-3471).
func SpawnMonsterItem(deps SpawnDeps, plan DropPlan, seed uint32) (MonsterItem, error) {
	if plan.ItemIndex == ItemNone || plan.NoDrop {
		return MonsterItem{}, nil
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

	return MonsterItem{Dropped: true, Setup: setup, Unique: unique}, nil
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
