package itemgen_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
)

// TestSpawnMonsterItemNoDrop pins the early exit: a plan without an item index
// spawns nothing (items.cpp:3455-3456).
func TestSpawnMonsterItemNoDrop(t *testing.T) {
	// given
	deps := spawnDeps(t)

	// when
	out, err := itemgen.SpawnMonsterItem(deps, itemgen.DropPlan{ItemIndex: itemgen.ItemNone}, 7)
	// then
	if err != nil {
		t.Fatalf("SpawnMonsterItem: %v", err)
	}
	if out.Dropped {
		t.Fatalf("out = %+v, want no item", out)
	}
}

// TestSpawnMonsterItemWiresPlanLevel pins the mLevel hand-off: the plan's raw
// monster level and uper reach SetupAllItems, and a non-unique item leaves the
// unique retry skipped (items.cpp:3464-3471).
func TestSpawnMonsterItemWiresPlanLevel(t *testing.T) {
	// given
	deps := spawnDeps(t)
	plan := itemgen.DropPlan{ItemIndex: 0, Level: 7, Uper: 1, OnlyGood: false}

	// when
	out, err := itemgen.SpawnMonsterItem(deps, plan, 42)
	// then
	if err != nil {
		t.Fatalf("SpawnMonsterItem: %v", err)
	}
	if !out.Dropped || out.Setup.Iblvl != 7 {
		t.Fatalf("setup = %+v, want iblvl 7 from the plan level", out.Setup)
	}
	if !out.Unique.Skipped {
		t.Fatalf("unique = %+v, want a skipped retry on a plain item", out.Unique)
	}
}

// TestPlanDropCarriesDataLevel pins the second level SpawnItem reads: the raw
// monstdat level is carried on the plan next to the difficulty-adjusted level
// used for the pool pick (items.cpp:3435, 3462, 3467).
func TestPlanDropCarriesDataLevel(t *testing.T) {
	// given
	items := realItems(t)
	monster := itemgen.MonsterDrop{Level: 5, DataLevel: 9}
	rnd := &tapeRnd{t: t, values: []int32{0, 0, 0, 0}}

	// when
	plan, err := itemgen.PlanDrop(items, monster, itemgen.DropEnv{}, itemgen.BrainQuest{}, rnd.rnd)
	// then
	if err != nil {
		t.Fatalf("PlanDrop: %v", err)
	}
	if plan.Level != 9 {
		t.Fatalf("plan level = %d, want 9 (monster.data().level)", plan.Level)
	}
}

func spawnDeps(t *testing.T) itemgen.SpawnDeps {
	t.Helper()

	items, uniques := setupTables()
	prefixes, suffixes := lockAffixes()
	rnd := &tapeRnd{t: t, values: []int32{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}}

	return itemgen.SpawnDeps{
		Items: items, Uniques: uniques, Prefixes: prefixes, Suffixes: suffixes,
		Rnd: rnd.rnd, Discard: func(int) {}, NextSeed: func() uint32 { return 99 },
	}
}
