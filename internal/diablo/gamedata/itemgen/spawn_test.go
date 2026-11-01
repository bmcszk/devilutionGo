package itemgen_test

import (
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
)

// offsetResult is the UniqueOffset outcome a case checks.
type offsetResult struct {
	offset  int
	reverse bool
	err     error
}

// TestAvailableUniquesNoneDropped pins the uids filter of TryRandomUniqueItem:
// single player skips uniques already obtained (items.cpp:3328-3336).
func TestAvailableUniquesNoneDropped(t *testing.T) {
	// given
	valid := []int{3, 7, 9}

	// when
	got := itemgen.AvailableUniques(valid, nil, false)
	// then
	assertSameInts(t, got, []int{3, 7, 9})
}

func TestAvailableUniquesOneDropped(t *testing.T) {
	// given
	valid := []int{3, 7, 9}
	dropped := map[int]bool{7: true}

	// when
	got := itemgen.AvailableUniques(valid, dropped, false)
	// then
	assertSameInts(t, got, []int{3, 9})
}

func TestAvailableUniquesAllDropped(t *testing.T) {
	// given
	valid := []int{3, 7, 9}
	dropped := map[int]bool{3: true, 7: true, 9: true}

	// when
	got := itemgen.AvailableUniques(valid, dropped, false)
	// then
	if got != nil {
		t.Fatalf("available = %v, want nil so TryRandomUniqueItem takes the fallback", got)
	}
}

func TestAvailableUniquesMultiplayerKeepsDropped(t *testing.T) {
	// given
	valid := []int{3, 7, 9}
	dropped := map[int]bool{3: true, 7: true, 9: true}

	// when
	got := itemgen.AvailableUniques(valid, dropped, true)
	// then
	assertSameInts(t, got, valid)
}

// TestPickUniqueLastRow pins the uids index roll (items.cpp:3347).
func TestPickUniqueLastRow(t *testing.T) {
	// given
	uids := []int{5, 6, 7}

	// when
	uid, arg := pickUniqueWithDraw(uids, 2)
	// then
	if uid != 7 || arg != 3 {
		t.Fatalf("uid/arg = %d/%d, want 7/3", uid, arg)
	}
}

func TestPickUniqueNegativeClampsToZero(t *testing.T) {
	// given
	uids := []int{5, 6, 7}

	// when
	uid, _ := pickUniqueWithDraw(uids, -1)
	// then
	if uid != 5 {
		t.Fatalf("uid = %d, want 5 (upstream would index out of bounds)", uid)
	}
}

// TestUniqueOffsetLastRow pins the reverse-compatibility walk: the last valid
// row needs no offset (items.cpp:3366-3385).
func TestUniqueOffsetLastRow(t *testing.T) {
	// given
	uniques := offsetUniques()

	// when
	offset, reverse, err := itemgen.UniqueOffset([]int{0, 1, 2}, 2, uniques)
	// then
	assertUniqueOffset(t, offsetResult{offset, reverse, err}, offsetResult{offset: 0, reverse: true})
}

func TestUniqueOffsetHigherLevelLaterRow(t *testing.T) {
	// given
	uniques := offsetUniques()

	// when
	offset, reverse, err := itemgen.UniqueOffset([]int{0, 1, 2}, 1, uniques)
	// then
	assertUniqueOffset(t, offsetResult{offset, reverse, err}, offsetResult{offset: 1, reverse: true})
}

func TestUniqueOffsetLowerLevelLaterRow(t *testing.T) {
	// given
	uniques := offsetUniques()

	// when
	offset, reverse, err := itemgen.UniqueOffset([]int{0, 1, 2}, 0, uniques)
	// then
	assertUniqueOffset(t, offsetResult{offset, reverse, err}, offsetResult{offset: 2})
}

func TestUniqueOffsetUnknownUID(t *testing.T) {
	// given
	uniques := offsetUniques()

	// when
	_, _, err := itemgen.UniqueOffset([]int{0, 1, 2}, 4, uniques)
	// then
	if err == nil || !strings.Contains(err.Error(), "not in the valid list") {
		t.Fatalf("err = %v, want the offset guard", err)
	}
}

func assertUniqueOffset(t *testing.T, got, want offsetResult) {
	t.Helper()

	if got.err != nil {
		t.Fatalf("UniqueOffset: %v", got.err)
	}
	if got.offset != want.offset || got.reverse != want.reverse {
		t.Fatalf("offset/reverse = %d/%v, want %d/%v", got.offset, got.reverse, want.offset, want.reverse)
	}
}

func pickUniqueWithDraw(uids []int, draw int32) (int, int32) {
	var gotArg int32

	uid := itemgen.PickUnique(uids, func(n int32) int32 {
		gotArg = n

		return draw
	})

	return uid, gotArg
}

func assertSameInts(t *testing.T, got, want []int) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func offsetUniques() []itemdat.UniqueItemData {
	return []itemdat.UniqueItemData{
		{Name: "a", UniqueBaseItem: "SWORD", MinLevel: 5},
		{Name: "b", UniqueBaseItem: "SWORD", MinLevel: 3},
		{Name: "c", UniqueBaseItem: "SWORD", MinLevel: 9},
	}
}
