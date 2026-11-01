package itemgen_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
	"github.com/bmcszk/devilutionGo/internal/diablo/testsupport"
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

// TestSpawnItemUniqueTreasureShortCircuit pins the T_UNIQ branch: the treasure
// low bits are the unique index and SpawnUnique takes over entirely
// (items.cpp:3429-3433).
func TestSpawnItemUniqueTreasureShortCircuit(t *testing.T) {
	// given
	deps := realSpawnDeps(t)
	monster := itemgen.MonsterDrop{Treasure: itemgen.TreasureUnique, DataLevel: 5}

	// when
	out, err := itemgen.SpawnItem(deps, monster, itemgen.BrainQuest{}, itemgen.DropEnv{}, 42)
	// then
	if err != nil {
		t.Fatalf("SpawnItem: %v", err)
	}
	if out.Plan.UniqueIndex != 0 || out.Setup.Quality != itemgen.QualityUnique || out.Setup.UID != 0 {
		t.Fatalf("out = %+v, want the CLEAVER unique from uid 0", out)
	}
}

// TestSpawnItemNormalDrop pins the normal branch: no T_UNIQ, so the plan's raw
// level reaches SetupAllItems (items.cpp:3455-3471).
func TestSpawnItemNormalDrop(t *testing.T) {
	// given
	deps := realSpawnDeps(t)
	monster := itemgen.MonsterDrop{Level: 5, DataLevel: 7}

	// when
	out, err := itemgen.SpawnItem(deps, monster, itemgen.BrainQuest{}, itemgen.DropEnv{}, 42)
	// then
	if err != nil {
		t.Fatalf("SpawnItem: %v", err)
	}
	if !out.Dropped || out.Plan.UniqueIndex != itemgen.UniqueNone || out.Plan.Level != 7 {
		t.Fatalf("plan = %+v, want a normal drop at data level 7", out.Plan)
	}
}

// TestSpawnUniqueHigherDifficulty pins the non-NORMAL branch: the base row is
// re-picked from the droppable pool of the same itype at curlv*2 with uper 15
// (items.cpp:3215-3222).
func TestSpawnUniqueHigherDifficulty(t *testing.T) {
	// given
	deps := syntheticSpawnDeps(t)

	// when
	out, err := itemgen.SpawnUnique(deps, 0, 5, 1, 42)
	// then
	if err != nil {
		t.Fatalf("SpawnUnique: %v", err)
	}
	if out.Direct {
		t.Fatalf("out = %+v, want the pool-pick path on a non-normal difficulty", out)
	}
	if out.Setup.CreateInfo&itemgen.CreateInfoUper15 == 0 {
		t.Fatalf("createInfo = %d, want the uper15 flag", out.Setup.CreateInfo)
	}
}

// realSpawnDeps loads the committed itemdat / uniqueitemdat fixtures so the
// droppable pools are non-empty (the synthetic tables have no DropRate rows).
// syntheticSpawnDeps builds droppable rows plus a unique that matches them by
// ID, so the non-NORMAL pool pick has a pool to pick from.

// TestSpawnUniqueLevelFallback pins the nullopt/optional level split: DIFF_NORMAL
// always uses ItemsGetCurrlevel() (items.cpp:3203-3214), and a non-normal
// difficulty overrides curlv only when a level is actually given (:3204-3205).
func TestSpawnUniqueLevelFallback(t *testing.T) {
	// given
	deps := syntheticSpawnDeps(t)
	deps.Currlevel = 3

	// when
	normal, errNormal := itemgen.SpawnUnique(deps, 0, itemgen.LevelNotGiven, itemgen.DifficultyNormal, 7)
	hard, errHard := itemgen.SpawnUnique(deps, 0, itemgen.LevelNotGiven, itemgen.DifficultyNightmare, 7)
	// then
	if errNormal != nil || errHard != nil {
		t.Fatalf("SpawnUnique: %v / %v", errNormal, errHard)
	}
	if normal.Setup.CreateInfo&0x0F != 3 {
		t.Fatalf("normal createInfo = %#x, want the curlv 3 from ItemsGetCurrlevel", normal.Setup.CreateInfo)
	}
	if hard.Setup.CreateInfo&0x0F != (3*2)&0x0F {
		t.Fatalf("nightmare createInfo = %#x, want the curlv*2 level", hard.Setup.CreateInfo)
	}
}

func spawnDeps(t *testing.T) itemgen.SpawnDeps {
	t.Helper()

	items, uniques := setupTables()
	prefixes, suffixes := lockAffixes()
	rnd := &tapeRnd{t: t, values: []int32{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}}

	return itemgen.SpawnDeps{
		Items: items, Uniques: uniques, Prefixes: prefixes, Suffixes: suffixes,
		Currlevel: 1, Rnd: rnd.rnd, Discard: func(int) {}, NextSeed: func() uint32 { return 99 },
	}
}

func syntheticSpawnDeps(t *testing.T) itemgen.SpawnDeps {
	t.Helper()

	rnd := &tapeRnd{t: t, values: make([]int32, 32)}

	return itemgen.SpawnDeps{
		Items: testItems(3),
		Uniques: []itemdat.UniqueItemData{{
			Name: "Test Unique", UniqueBaseItem: "IDI_TEST", MinLevel: 1, Value: 100,
		}},
		Currlevel: 1, Rnd: rnd.rnd, Discard: func(int) {}, NextSeed: func() uint32 { return 99 },
	}
}

func realSpawnDeps(t *testing.T) itemgen.SpawnDeps {
	t.Helper()

	uniques, err := itemdat.LoadUniqueItemDat(testsupport.Corpus(t, "items", "assets/txtdata/items/unique_itemdat.tsv"))
	if err != nil {
		t.Fatalf("LoadUniqueItemDat: %v", err)
	}
	rnd := &tapeRnd{t: t, values: make([]int32, 256)}

	return itemgen.SpawnDeps{
		Items: realItems(t), Uniques: uniques,
		Prefixes: realPrefixes(t), Suffixes: realSuffixes(t),
		Currlevel: 1, Rnd: rnd.rnd, Discard: func(int) {}, NextSeed: func() uint32 { return 99 },
	}
}
