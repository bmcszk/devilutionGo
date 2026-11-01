package itemgen_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
	"github.com/bmcszk/devilutionGo/internal/diablo/testsupport"
)

func TestIsItemAvailableVanilla(t *testing.T) {
	// given
	items := realItems(t)
	vanilla := itemgen.DropEnv{}

	// when
	results := map[string]bool{
		"gold":      itemgen.IsItemAvailable(items, 0, vanilla),
		"sorcerer":  itemgen.IsItemAvailable(items, 5, vanilla),
		"mapofdoom": itemgen.IsItemAvailable(items, 22, vanilla),
		"lgtforge":  itemgen.IsItemAvailable(items, 32, vanilla),
		"bardsword": itemgen.IsItemAvailable(items, 37, vanilla),
		"bards":     itemgen.IsItemAvailable(items, 37, itemgen.DropEnv{TestBard: true}),
		"hellfire":  itemgen.IsItemAvailable(items, 22, itemgen.DropEnv{Hellfire: true}),
		"spawnarm":  itemgen.IsItemAvailable(items, 62, itemgen.DropEnv{Spawn: true}),
		"negative":  itemgen.IsItemAvailable(items, -1, vanilla),
		"pastend":   itemgen.IsItemAvailable(items, len(items), vanilla),
	}

	// then
	want := map[string]bool{
		"gold": true, "sorcerer": false, "mapofdoom": false, "lgtforge": false,
		"bardsword": false, "bards": true, "hellfire": true, "spawnarm": false,
		"negative": false, "pastend": false,
	}
	for name, got := range results {
		if got != want[name] {
			t.Fatalf("%s available = %v, want %v", name, got, want[name])
		}
	}
}

func TestGetItemIndexForDroppableItemWeights(t *testing.T) {
	// given
	items := testItems(3)
	items[0].DropRate = 1
	items[1].DropRate = 2
	items[2].DropRate = 3
	all := func(itemdat.ItemData) bool { return true }
	targets := []int32{0, 1, 2, 3, 4, 5}

	// when
	got, weightDraws := weightedPicks(t, items, all, targets)

	// then
	assertPicks(t, targets, got, weightDraws, []int{0, 1, 1, 2, 2, 2})
}

func TestGetItemIndexForDroppableItemSkips(t *testing.T) {
	// given
	items := testItems(3)
	items[0].DropRate = 0
	items[1].DropRate = 5
	items[1].Spell = "RESURRECT"
	items[2].DropRate = 5
	all := func(itemdat.ItemData) bool { return true }
	single := func() *tapeRnd { return &tapeRnd{t: t, values: []int32{0}} }

	// when
	singleIdx, err := itemgen.GetItemIndexForDroppableItem(items, itemgen.DropEnv{}, true, all, single().rnd)
	if err != nil {
		t.Fatalf("single-player: %v", err)
	}
	multiIdx, err := itemgen.GetItemIndexForDroppableItem(
		items,
		itemgen.DropEnv{Multiplayer: true},
		true,
		all,
		single().rnd,
	)
	if err != nil {
		t.Fatalf("multiplayer: %v", err)
	}

	// then
	if singleIdx != 2 {
		t.Fatalf("single-player row = %d, want 2 (zero drop rate and Resurrect skipped)", singleIdx)
	}
	if multiIdx != 1 {
		t.Fatalf("multiplayer row = %d, want 1 (Resurrect kept)", multiIdx)
	}
}

func TestGetItemIndexForDroppableItemEmptyPool(t *testing.T) {
	// given
	items := testItems(2)
	none := func(itemdat.ItemData) bool { return false }

	// when
	_, err := itemgen.GetItemIndexForDroppableItem(items, itemgen.DropEnv{}, true, none, seededRnd(1))

	// then
	if err == nil {
		t.Fatal("err = nil, want the empty-pool error")
	}
}

func TestRndItemForMonsterLevelBranches(t *testing.T) {
	// given
	items := testItems(2)
	items[0].DropRate = 1
	items[1].DropRate = 1
	pass := &tapeRnd{t: t, values: []int32{41}}
	gold := &tapeRnd{t: t, values: []int32{40, 26}}
	pool := &tapeRnd{t: t, values: []int32{40, 25, 1}}

	// when
	passIdx, err := itemgen.RndItemForMonsterLevel(items, 10, itemgen.DropEnv{}, pass.rnd)
	if err != nil {
		t.Fatalf("RndItemForMonsterLevel pass: %v", err)
	}
	goldIdx, err := itemgen.RndItemForMonsterLevel(items, 10, itemgen.DropEnv{}, gold.rnd)
	if err != nil {
		t.Fatalf("RndItemForMonsterLevel gold: %v", err)
	}
	poolIdx, err := itemgen.RndItemForMonsterLevel(items, 10, itemgen.DropEnv{}, pool.rnd)
	if err != nil {
		t.Fatalf("RndItemForMonsterLevel pool: %v", err)
	}

	// then
	if passIdx != itemgen.ItemNone {
		t.Fatalf("roll 41 -> %d, want ItemNone (items.cpp:3258)", passIdx)
	}
	if goldIdx != 0 {
		t.Fatalf("roll 26 -> %d, want row 0 (gold, :3261)", goldIdx)
	}
	if poolIdx != 1 {
		t.Fatalf("pool pick -> %d, want row 1", poolIdx)
	}
}

func TestRndUItemFilter(t *testing.T) {
	// given
	items := testItems(3)
	items[0].ItemType = "Gold"
	items[1].ItemType = "Sword"
	items[1].MinMonsterLevel = 20
	items[2].ItemType = "Misc"
	items[2].MiscID = "BOOK"
	items[2].MinMonsterLevel = 50
	rnd := &tapeRnd{t: t, values: []int32{0}}

	// when
	idx, err := itemgen.RndUItem(items, 1, itemgen.DropEnv{}, rnd.rnd)
	// then
	if err != nil {
		t.Fatalf("RndUItem: %v", err)
	}
	if idx != 2 {
		t.Fatalf("row = %d, want 2 (book passes the level filter)", idx)
	}
	if rnd.args[0] != 1 {
		t.Fatalf("weight draw = %d, want 1 (single unweighted candidate)", rnd.args[0])
	}
}

func TestPlanDropSpecialTreasure(t *testing.T) {
	// given
	items := testItems(2)
	monster := itemgen.MonsterDrop{Treasure: itemgen.TreasureUnique | 7}
	rnd := &tapeRnd{t: t}

	// when
	plan, err := itemgen.PlanDrop(items, monster, itemgen.DropEnv{}, itemgen.BrainQuest{}, rnd.rnd)
	// then
	if err != nil {
		t.Fatalf("PlanDrop: %v", err)
	}
	if plan.UniqueIndex != 7 {
		t.Fatalf("UniqueIndex = %d, want 7 (treasure & T_MASK)", plan.UniqueIndex)
	}
	if plan.ItemIndex != itemgen.ItemNone || len(rnd.args) != 0 {
		t.Fatalf("plan = %+v draws = %v, want the SpawnUnique short-circuit", plan, rnd.args)
	}
}

func TestPlanDropUniqueMonster(t *testing.T) {
	// given
	items := testItems(2)
	monster := itemgen.MonsterDrop{Unique: true, Level: 10}
	rnd := &tapeRnd{t: t, values: []int32{1}}

	// when
	plan, err := itemgen.PlanDrop(items, monster, itemgen.DropEnv{}, itemgen.BrainQuest{}, rnd.rnd)
	// then
	if err != nil {
		t.Fatalf("PlanDrop: %v", err)
	}
	if plan.ItemIndex != 1 || !plan.OnlyGood || plan.Uper != 15 {
		t.Fatalf("plan = %+v, want RndUItem row 1, onlygood, uper 15", plan)
	}
}

func TestPlanDropBrainSinglePlayer(t *testing.T) {
	// given
	items := testItems(2)
	monster := itemgen.MonsterDrop{Treasure: itemgen.TreasureNoDrop, Level: 10}
	brain := itemgen.BrainQuest{Active: true, Given: true}
	rnd := &tapeRnd{t: t}

	// when
	plan, err := itemgen.PlanDrop(items, monster, itemgen.DropEnv{}, brain, rnd.rnd)
	// then
	if err != nil {
		t.Fatalf("PlanDrop: %v", err)
	}
	if !plan.Brain || plan.ItemIndex != 18 || plan.NoDrop {
		t.Fatalf("plan = %+v, want the IDI_BRAIN row 18 replacing the drop", plan)
	}
}

func TestPlanDropBrainExtraMultiplayer(t *testing.T) {
	// given
	items := testItems(2)
	monster := itemgen.MonsterDrop{Treasure: 0, Level: 10}
	brain := itemgen.BrainQuest{Active: true, Given: true, SendMsg: true}
	rnd := &tapeRnd{t: t, values: []int32{41}}

	// when
	plan, err := itemgen.PlanDrop(items, monster, itemgen.DropEnv{Multiplayer: true}, brain, rnd.rnd)
	// then
	if err != nil {
		t.Fatalf("PlanDrop: %v", err)
	}
	if !plan.BrainExtra || plan.Brain {
		t.Fatalf("plan = %+v, want the extra brain spawn and a normal drop", plan)
	}
	if plan.ItemIndex != itemgen.ItemNone || plan.OnlyGood {
		t.Fatalf("plan = %+v, want the normal (non-onlygood) monster-level path", plan)
	}
}

func TestPlanDropNoDrop(t *testing.T) {
	// given
	items := testItems(2)
	monster := itemgen.MonsterDrop{Treasure: itemgen.TreasureNoDrop, Level: 10}
	rnd := &tapeRnd{t: t}

	// when
	plan, err := itemgen.PlanDrop(items, monster, itemgen.DropEnv{}, itemgen.BrainQuest{}, rnd.rnd)
	// then
	if err != nil {
		t.Fatalf("PlanDrop: %v", err)
	}
	if !plan.NoDrop || plan.ItemIndex != itemgen.ItemNone || len(rnd.args) != 0 {
		t.Fatalf("plan = %+v draws = %v, want a silent no-drop", plan, rnd.args)
	}
}

func TestPlanDropSpecialTreasureWithMPQuests(t *testing.T) {
	// given
	items := testItems(2)
	monster := itemgen.MonsterDrop{Treasure: itemgen.TreasureUnique | 7, Level: 10}
	rnd := &tapeRnd{t: t, values: []int32{0}}

	// when
	plan, err := itemgen.PlanDrop(items, monster, itemgen.DropEnv{UseMPQuests: true}, itemgen.BrainQuest{}, rnd.rnd)
	// then
	if err != nil {
		t.Fatalf("PlanDrop: %v", err)
	}
	if plan.UniqueIndex != 0 {
		t.Fatalf("UniqueIndex = %d, want no SpawnUnique short-circuit", plan.UniqueIndex)
	}
	if plan.Uper != 1 || !plan.OnlyGood {
		t.Fatalf("plan = %+v, want uper 1 (not unique) and onlygood kept", plan)
	}
}

func weightedPicks(
	t *testing.T,
	items []itemdat.ItemData,
	all func(itemdat.ItemData) bool,
	targets []int32,
) ([]int, []int32) {
	t.Helper()

	got := make([]int, len(targets))
	draws := make([]int32, len(targets))
	for i, target := range targets {
		rnd := &tapeRnd{t: t, values: []int32{target}}
		idx, err := itemgen.GetItemIndexForDroppableItem(items, itemgen.DropEnv{}, true, all, rnd.rnd)
		if err != nil {
			t.Fatalf("GetItemIndexForDroppableItem: %v", err)
		}
		got[i], draws[i] = idx, rnd.args[0]
	}

	return got, draws
}

func assertPicks(t *testing.T, targets []int32, got []int, draws []int32, want []int) {
	t.Helper()

	for i, target := range targets {
		if got[i] != want[i] {
			t.Fatalf("target %d -> row %d, want %d", target, got[i], want[i])
		}
		if draws[i] != 6 {
			t.Fatalf("weight draw = %d, want 6 (total drop rate)", draws[i])
		}
	}
}

func realItems(t *testing.T) []itemdat.ItemData {
	t.Helper()

	items, err := itemdat.LoadItemDat(testsupport.Corpus(t, "items", "assets/txtdata/items/itemdat.tsv"))
	if err != nil {
		t.Fatalf("LoadItemDat: %v", err)
	}

	return items
}

// testItems builds n droppable vanilla-free rows (row indices 0..n-1 carry no
// pinned exclusions from IsItemAvailable).
func testItems(n int) []itemdat.ItemData {
	items := make([]itemdat.ItemData, n)
	for i := range items {
		items[i] = itemdat.ItemData{ID: "IDI_TEST", DropRate: 1, ItemType: "Sword"}
	}

	return items
}
