package inventory_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/entities"
	"github.com/bmcszk/devilutionGo/internal/diablo/inventory"
)

// Cursor graphics the fixture items use. Upstream derives an item's inventory
// footprint from its cursor sprite size (inv.cpp:2292-2297), so tests script
// the sprite table instead of loading CLX assets.
const (
	cursConsumable uint8 = 1 // 32x32 -> 1x1
	cursSword      uint8 = 2 // 32x96 -> 1x3
	cursBow        uint8 = 3 // 64x96 -> 2x3
	cursArmor      uint8 = 4 // 64x64 -> 2x2
	cursRagged     uint8 = 5 // 63x96 -> 1x3 (integer division, inv.cpp:2296)
)

// testSpriteSizes is the scripted pCursCels stand-in (cursor.cpp:476).
var testSpriteSizes = map[uint8]inventory.Size{
	cursConsumable: {Width: 32, Height: 32},
	cursSword:      {Width: 32, Height: 96},
	cursBow:        {Width: 64, Height: 96},
	cursArmor:      {Width: 64, Height: 64},
	cursRagged:     {Width: 63, Height: 96},
}

// slotResult is one FindSlot observation plus its expectation, so a failure
// names the size and both outcomes.
type slotResult struct {
	size      inventory.Size
	got       int
	found     bool
	want      int
	wantFound bool
}

// placeResult is one Place observation plus its expectation.
type placeResult struct {
	placed   bool
	player   *entities.Player
	wantGrid map[int]int8
}

func TestItemSize(t *testing.T) {
	tests := []struct {
		name string
		curs uint8
		want inventory.Size
	}{
		{"consumable", cursConsumable, inventory.Size{Width: 1, Height: 1}},
		{"sword", cursSword, inventory.Size{Width: 1, Height: 3}},
		{"bow", cursBow, inventory.Size{Width: 2, Height: 3}},
		{"armor", cursArmor, inventory.Size{Width: 2, Height: 2}},
		{"ragged width rounds down", cursRagged, inventory.Size{Width: 1, Height: 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// when
			got := inventory.ItemSize(newItem(tt.curs), testSprites)
			// then
			if got != tt.want {
				t.Fatalf("ItemSize(curs %d) = %+v, want %+v", tt.curs, got, tt.want)
			}
		})
	}
}

func TestFitsInSlot(t *testing.T) {
	tests := []struct {
		name   string
		occupi map[int]int8 // grid cell -> InvGrid value
		slot   int
		size   inventory.Size
		ignore int
		want   bool
	}{
		{name: "empty grid", slot: 0, size: inventory.Size{Width: 2, Height: 3}, want: true},
		{
			name:   "occupied by another item",
			occupi: map[int]int8{11: -2},
			slot:   0,
			size:   inventory.Size{Width: 2, Height: 3},
			want:   false,
		},
		{
			name:   "occupied by the ignored item",
			occupi: map[int]int8{10: 2, 11: -2},
			slot:   0,
			size:   inventory.Size{Width: 2, Height: 3},
			ignore: 1,
			want:   true,
		},
		{
			name: "too wide for the column",
			slot: 9,
			size: inventory.Size{Width: 2, Height: 1},
			want: false,
		},
		{
			name: "runs past the last row",
			slot: 30,
			size: inventory.Size{Width: 1, Height: 2},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// when
			got := occupiedInv(t, tt.occupi).FitsInSlot(tt.slot, tt.size, tt.ignore)
			// then
			if got != tt.want {
				t.Fatalf("FitsInSlot(%d, %+v, %d) = %v, want %v", tt.slot, tt.size, tt.ignore, got, tt.want)
			}
		})
	}
}

func TestFindSlotOrder(t *testing.T) {
	tests := []struct {
		name   string
		occupi map[int]int8
		size   inventory.Size
		want   int
		wantOK bool
	}{
		{name: "1x1 prefers the bottom row left to right", size: inventory.Size{Width: 1, Height: 1}, want: 30, wantOK: true},
		{
			name:   "1x1 falls back to columns right to left, bottom to top",
			occupi: bottomRowOccupied(),
			size:   inventory.Size{Width: 1, Height: 1},
			want:   29,
			wantOK: true,
		},
		{name: "2x2 starts at the right edge", size: inventory.Size{Width: 2, Height: 2}, want: 8, wantOK: true},
		{name: "1x3 walks contiguous indices", size: inventory.Size{Width: 1, Height: 3}, want: 0, wantOK: true},
		{name: "2x3 skips the ninth column", size: inventory.Size{Width: 2, Height: 3}, want: 0, wantOK: true},
		{
			name:   "2x3 second block starts at 10",
			occupi: firstBlockOccupied(),
			size:   inventory.Size{Width: 2, Height: 3},
			want:   10,
			wantOK: true,
		},
		{
			name:   "2x3 gives up when both blocks are full",
			occupi: bothBlocksOccupied(),
			size:   inventory.Size{Width: 2, Height: 3},
			wantOK: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			p, inv := newInv(t)
			for cell, val := range tt.occupi {
				p.InvGrid[cell] = val
			}
			// when
			got, ok := inv.FindSlot(tt.size, -1)
			// then
			assertSlot(t, slotResult{size: tt.size, got: got, found: ok, want: tt.want, wantFound: tt.wantOK})
		})
	}
}

func TestCanFit(t *testing.T) {
	tests := []struct {
		name string
		curs uint8
		fill int
		want bool
	}{
		{name: "empty inventory fits a bow", curs: cursBow, want: true},
		{name: "inventory choked by 1x1s", curs: cursBow, fill: 40, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// when
			got := filledInv(t, tt.fill).CanFit(newItem(tt.curs))
			// then
			if got != tt.want {
				t.Fatalf("CanFit(curs %d) = %v, want %v", tt.curs, got, tt.want)
			}
		})
	}
}

func TestPlace(t *testing.T) {
	tests := []struct {
		name     string
		curs     uint8
		wantGrid map[int]int8
	}{
		{
			name:     "1x1 item takes the first bottom-row cell",
			curs:     cursConsumable,
			wantGrid: map[int]int8{30: 1},
		},
		{
			name:     "2x3 item fills six cells, only the bottom-left is positive",
			curs:     cursBow,
			wantGrid: map[int]int8{0: -1, 1: -1, 10: -1, 11: -1, 20: 1, 21: -1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			p, inv := newInv(t)
			// when
			ok := inv.Place(newItem(tt.curs))
			// then
			assertPlaced(t, placeResult{placed: ok, player: p, wantGrid: tt.wantGrid})
		})
	}
}

func TestPlaceFullGrid(t *testing.T) {
	// given a grid choked by 40 single-cell items
	p, inv := newInv(t)
	for i := range 40 {
		p.InvGrid[i] = 1
	}
	// when
	ok := inv.Place(newItem(cursConsumable))
	// then
	if ok {
		t.Fatal("Place on a full grid = true, want false")
	}
}

func TestCouldFit(t *testing.T) {
	// given forty 1x1 items filling every cell
	_, inv := newInv(t)
	for i := range 40 {
		if !inv.Place(newItem(cursConsumable)) {
			t.Fatalf("Place #%d = false, want true", i)
		}
	}
	// when
	withoutIgnore := inv.CouldFit(newItem(cursConsumable), -1)
	withIgnore := inv.CouldFit(newItem(cursConsumable), 4)
	// then the grid only takes an item if the ignored one is removed
	if withoutIgnore {
		t.Fatal("CouldFit without ignoring = true, want false")
	}
	if !withIgnore {
		t.Fatal("CouldFit ignoring item 4 = false, want true")
	}
}

// testSprites resolves the scripted sprite table above.
func testSprites(curs uint8) inventory.Size {
	return testSpriteSizes[curs]
}

// newInv returns a player and the inventory view over its arrays, with every
// slot empty and a permissive item table (every ID is usable).
func newInv(t *testing.T) (*entities.Player, inventory.Inv) {
	t.Helper()
	p := &entities.Player{}
	seedEmpty(p)
	data := inventory.Data{
		Sprite:  testSprites,
		Usable:  func(int16) bool { return true },
		MaxGold: inventory.MaxGoldLimit,
	}
	return p, inventory.New(p, data)
}

// seedEmpty fills the player's item slots with EmptyItem, matching C++'s
// default-constructed Item (items.h:194).
func seedEmpty(p *entities.Player) {
	for i := range p.InvBody {
		p.InvBody[i] = inventory.EmptyItem()
	}
	for i := range p.InvList {
		p.InvList[i] = inventory.EmptyItem()
	}
	for i := range p.SpdList {
		p.SpdList[i] = inventory.EmptyItem()
	}
	p.HoldItem = inventory.EmptyItem()
}

// newItem is the smallest item that carries what the grid logic reads.
func newItem(curs uint8) entities.Item {
	return entities.Item{Itype: int8(inventory.ItemTypeMisc), Curs: curs}
}

// bottomRowOccupied marks the last row as taken by item 1.
func bottomRowOccupied() map[int]int8 {
	m := make(map[int]int8, inventory.InventoryPitch)
	for i := 30; i <= 39; i++ {
		m[i] = 1
	}
	return m
}

// firstBlockOccupied marks the 2x3 first block of the search order as taken.
func firstBlockOccupied() map[int]int8 {
	m := make(map[int]int8, 9)
	for i := range 9 {
		m[i] = 1
	}
	return m
}

// bothBlocksOccupied marks both 2x3 search blocks as taken.
func bothBlocksOccupied() map[int]int8 {
	m := firstBlockOccupied()
	for i := range 9 {
		m[i+10] = 1
	}
	return m
}

// occupiedInv returns an inventory whose grid cells are pre-set.
func occupiedInv(t *testing.T, occupi map[int]int8) inventory.Inv {
	t.Helper()
	p, inv := newInv(t)
	for cell, val := range occupi {
		p.InvGrid[cell] = val
	}
	return inv
}

// filledInv returns an inventory with the first `fill` cells marked taken.
func filledInv(t *testing.T, fill int) inventory.Inv {
	t.Helper()
	p, inv := newInv(t)
	for i := range fill {
		p.InvGrid[i] = 1
	}
	return inv
}

// assertSlot checks a FindSlot result against the expectation.
func assertSlot(t *testing.T, r slotResult) {
	t.Helper()
	if r.found != r.wantFound || (r.found && r.got != r.want) {
		t.Fatalf("FindSlot(%+v) = (%d, %v), want (%d, %v)", r.size, r.got, r.found, r.want, r.wantFound)
	}
}

// assertPlaced checks the grid cells written by Place.
func assertPlaced(t *testing.T, r placeResult) {
	t.Helper()
	if !r.placed {
		t.Fatal("Place = false, want true")
	}
	if r.player.NumInv != 1 {
		t.Fatalf("NumInv = %d, want 1", r.player.NumInv)
	}
	if inventory.IsEmpty(r.player.InvList[0]) {
		t.Fatal("InvList[0] should hold the placed item")
	}
	if !inventory.IsEmpty(r.player.InvList[1]) {
		t.Fatal("InvList[1] should stay empty")
	}
	for cell, want := range r.wantGrid {
		if got := r.player.InvGrid[cell]; got != want {
			t.Fatalf("InvGrid[%d] = %d, want %d", cell, got, want)
		}
	}
}
