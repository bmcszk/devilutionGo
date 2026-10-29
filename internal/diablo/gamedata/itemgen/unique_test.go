package itemgen_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
)

type levelCase struct {
	name     string
	lvl      int
	miscID   string
	onlygood bool
	uper15   bool
	values   []int32
	want     int
	draws    int
}

type uniqueCase struct {
	name     string
	uper     int
	values   []int32
	offset   int
	wantUID  int
	wantBurn int
}

func TestGetItemBLevelRollOrder(t *testing.T) {
	tests := []levelCase{
		{name: "ten percent passes", lvl: 20, values: []int32{10}, want: 20, draws: 1},
		{name: "turns over to lvl", lvl: 20, values: []int32{11, 20}, want: 20, draws: 2},
		{name: "both fail", lvl: 5, values: []int32{11, 6}, want: -1, draws: 2},
		{name: "onlygood forces lvl", lvl: 20, onlygood: true, values: []int32{11, 21}, want: 20, draws: 2},
		{name: "staff forces lvl", lvl: 20, miscID: "STAFF", values: []int32{11, 21}, want: 20, draws: 2},
		{name: "uper15 overrides", lvl: 20, uper15: true, values: []int32{11, 21}, want: 24, draws: 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// when
			got, draws := runBLevel(t, tc)

			// then
			if got != tc.want || draws != tc.draws {
				t.Fatalf("iblvl/draws = %d/%d, want %d/%d (items.cpp:1483-1489 || chain)", got, draws, tc.want, tc.draws)
			}
		})
	}
}

func TestCheckUniqueGates(t *testing.T) {
	tests := []uniqueCase{
		{name: "uper15 hits then burns one", uper: 15, values: []int32{15}, offset: 0, wantUID: 3, wantBurn: 1},
		{name: "uper1 misses", uper: 1, values: []int32{2}, offset: 0, wantUID: itemgen.UniqueNone, wantBurn: 0},
		{name: "offset walks down", uper: 15, values: []int32{1}, offset: 2, wantUID: 1, wantBurn: 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// when
			uid, burns := runCheckUnique(t, tc)

			// then
			if uid != tc.wantUID || burns != tc.wantBurn {
				t.Fatalf("uid = %d burns = %d, want %d / %d", uid, burns, tc.wantUID, tc.wantBurn)
			}
		})
	}
}

func TestCheckUniqueEmptyAndOffsetBounds(t *testing.T) {
	// given
	empty := &tapeRnd{t: t, values: []int32{1}}
	bounded := &tapeRnd{t: t, values: []int32{1}}

	// when
	uidEmpty, errEmpty := itemgen.CheckUnique(nil, 15, 0, empty.rnd, func(int) {})
	uidBounds, errBounds := itemgen.CheckUnique([]int{7}, 15, 3, bounded.rnd, func(int) {})

	// then
	if errEmpty != nil || uidEmpty != itemgen.UniqueNone {
		t.Fatalf("empty = (%d, %v), want (UniqueNone, nil)", uidEmpty, errEmpty)
	}
	if errBounds == nil || uidBounds != itemgen.UniqueNone {
		t.Fatalf("bounds = (%d, %v), want the offset error", uidBounds, errBounds)
	}
}

func TestValidUniquesFilter(t *testing.T) {
	// given
	uniques := []itemdat.UniqueItemData{
		{Name: "Griswold's Edge", UniqueBaseItem: "Sword", MinLevel: 1},
		{Name: "The Butcher's Cleaver", UniqueBaseItem: "Cleaver", MinLevel: 1},
		{Name: "King's Sword of Haste", UniqueBaseItem: "Sword", MinLevel: 15},
	}

	// when
	low := itemgen.ValidUniques(uniques, "Sword", 1)
	high := itemgen.ValidUniques(uniques, "Sword", 15)

	// then
	if len(low) != 1 || low[0] != 0 {
		t.Fatalf("lvl 1 = %v, want only row 0", low)
	}
	if len(high) != 2 || high[0] != 0 || high[1] != 2 {
		t.Fatalf("lvl 15 = %v, want rows 0 and 2", high)
	}
}

func TestItemRndDurJitter(t *testing.T) {
	// given
	rnd := &tapeRnd{t: t, values: []int32{0, 4}}

	// when
	low := itemgen.ItemRndDur(10, 100, rnd.rnd)
	high := itemgen.ItemRndDur(10, 100, rnd.rnd)
	kept := itemgen.ItemRndDur(0, 100, seededRnd(1))
	indestructible := itemgen.ItemRndDur(itemgen.DurIndestructible, 100, seededRnd(1))

	// then
	if low != 26 || high != 30 {
		t.Fatalf("jitter = %d/%d, want 26 and 30 (maxDur/4+1 + rnd(maxDur/2))", low, high)
	}
	if kept != 0 || indestructible != itemgen.DurIndestructible {
		t.Fatalf("guards = %d/%d, want 0 and %d", kept, indestructible, itemgen.DurIndestructible)
	}
}

func runBLevel(t *testing.T, tc levelCase) (iblvl, draws int) {
	t.Helper()
	rnd := &tapeRnd{t: t, values: tc.values}

	return itemgen.GetItemBLevel(tc.lvl, tc.miscID, tc.onlygood, tc.uper15, rnd.rnd), len(rnd.args)
}

func runCheckUnique(t *testing.T, tc uniqueCase) (uid, burns int) {
	t.Helper()
	rnd := &tapeRnd{t: t, values: tc.values}
	burns = 0
	var err error
	uid, err = itemgen.CheckUnique([]int{0, 1, 2, 3}, tc.uper, tc.offset, rnd.rnd, func(n int) { burns += n })
	if err != nil {
		t.Fatalf("CheckUnique: %v", err)
	}

	return uid, burns
}
