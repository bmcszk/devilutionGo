package itemgen_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
)

func TestRndAllItemsGoldGate(t *testing.T) {
	// given
	items := testItems(2)
	rnd := &tapeRnd{t: t, values: []int32{26}}

	// when
	idx, err := itemgen.RndAllItems(items, 5, itemgen.DropEnv{}, rnd.rnd)
	// then
	if err != nil {
		t.Fatalf("RndAllItems: %v", err)
	}
	if idx != 0 || len(rnd.args) != 1 {
		t.Fatalf("idx = %d draws = %v, want gold after a single roll", idx, rnd.args)
	}
}

func TestRndAllItemsLevelFilter(t *testing.T) {
	// given
	items := testItems(2)
	items[0].MinMonsterLevel = 1
	items[1].MinMonsterLevel = 11
	rnd := &tapeRnd{t: t, values: []int32{25, 0}}

	// when
	idx, err := itemgen.RndAllItems(items, 5, itemgen.DropEnv{}, rnd.rnd)
	// then
	if err != nil {
		t.Fatalf("RndAllItems: %v", err)
	}
	if idx != 0 {
		t.Fatalf("idx = %d, want row 0 (row 1 needs level 11, pool is currLevel*2)", idx)
	}
}

func TestRndTypeItemsFilters(t *testing.T) {
	// given
	items := testItems(3)
	items[0].ItemType = "Misc"
	items[0].MiscID = "BOOK"
	items[1].ItemType = "Misc"
	items[1].MiscID = "POTION"
	items[2].ItemType = "Sword"
	rnd := &tapeRnd{t: t, values: []int32{1}}

	// when
	idx, err := itemgen.RndTypeItems(items, "Misc", "POTION", 5, itemgen.DropEnv{}, rnd.rnd)
	anyMisc := &tapeRnd{t: t, values: []int32{0}}
	anyIdx, errAny := itemgen.RndTypeItems(items, "Misc", "", 5, itemgen.DropEnv{}, anyMisc.rnd)

	// then
	if err != nil || errAny != nil {
		t.Fatalf("RndTypeItems: %v / %v", err, errAny)
	}
	if idx != 1 || anyIdx != 0 {
		t.Fatalf("misc-id pick = %d any pick = %d, want 1 and 0", idx, anyIdx)
	}
	if rnd.args[0] != 1 || anyMisc.args[0] != 2 {
		t.Fatalf("pool sizes = %d/%d, want 1 (Potion only) and 2 (both Misc)", rnd.args[0], anyMisc.args[0])
	}
}

func TestTryRandomUniqueItemPicksUndropped(t *testing.T) {
	// given
	dropped := []bool{true, false, false, true}
	rnd := &tapeRnd{t: t, values: []int32{1}}

	// when
	roll := itemgen.TryRandomUniqueItem([]int{0, 1, 2, 3}, dropped, false, 15, 20, rnd.rnd)

	// then
	if !roll.Rolled || roll.UID != 2 {
		t.Fatalf("roll = %+v, want uid 2 (undropped list [1 2], index 1)", roll)
	}
	if rnd.args[0] != 2 {
		t.Fatalf("pool draw = %d, want 2 undropped candidates", rnd.args[0])
	}
	if roll.Uper != 15 || roll.Level != 20 {
		t.Fatalf("uper/level = %d/%d, want the unchanged 15/20", roll.Uper, roll.Level)
	}
}

func TestTryRandomUniqueItemMultiplayerKeepsAll(t *testing.T) {
	// given
	dropped := []bool{true, true}
	rnd := &tapeRnd{t: t, values: []int32{1}}

	// when
	roll := itemgen.TryRandomUniqueItem([]int{0, 1}, dropped, true, 15, 20, rnd.rnd)

	// then
	if !roll.Rolled || roll.UID != 1 || len(rnd.args) != 1 {
		t.Fatalf("roll = %+v draws = %v, want uid 1 from the full list", roll, rnd.args)
	}
}

func TestTryRandomUniqueItemDowngrade(t *testing.T) {
	// given
	dropped := []bool{true, true}
	rnd := &tapeRnd{t: t}

	// when
	roll := itemgen.TryRandomUniqueItem([]int{0, 1}, dropped, false, 15, 20, rnd.rnd)

	// then
	if !roll.Downgraded || roll.Rolled || roll.UID != itemgen.UniqueNone || len(rnd.args) != 0 {
		t.Fatalf("roll = %+v draws = %v, want the silent downgrade", roll, rnd.args)
	}
	if roll.Uper != 1 || roll.Level != 24 {
		t.Fatalf("uper/level = %d/%d, want 1/24 (uper 15 adds 4)", roll.Uper, roll.Level)
	}
}

func TestTryRandomUniqueItemDowngradeUperOne(t *testing.T) {
	// given
	rnd := &tapeRnd{t: t}

	// when
	roll := itemgen.TryRandomUniqueItem(nil, nil, false, 1, 20, rnd.rnd)

	// then
	if roll.Level != 20 || roll.Uper != 1 {
		t.Fatalf("uper/level = %d/%d, want 1/20 (no bump for uper 1)", roll.Uper, roll.Level)
	}
}
