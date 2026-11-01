package inventory_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/entities"
	"github.com/bmcszk/devilutionGo/internal/diablo/inventory"
)

// dropped records the stacks the drop sink spawns, in order.
var dropped []int

// goldResult is one gold-placement observation plus its expectation.
type goldResult struct {
	name      string
	got       int
	want      int
	gotCount  int
	wantCount int
}

// dropCase is one DropGold row: input state, variant and expectations.
type dropCase struct {
	name      string
	piles     []int
	amount    int
	drop      func(inventory.Inv, int) int
	wantLeft  int
	wantDead  []int
	wantNumIn int
}

// dropResult is one DropGold observation plus its expectation.
type dropResult struct {
	name     string
	left     int
	wantLeft int
	dropped  []int
	wantDead []int
	numIn    int
	wantNum  int
}

// splitResult is one SplitGold observation plus its expectation.
type splitResult struct {
	p         *playerFixture
	wantPile  int
	wantHeld  int
	wantEmpty bool
}

func TestGoldCursor(t *testing.T) {
	tests := []struct {
		name  string
		value int
		want  uint8
	}{
		{"small pile", 400, inventory.CursorGoldSmall},
		{"just above small", 1001, inventory.CursorGoldMedium},
		{"medium pile", 2499, inventory.CursorGoldMedium},
		{"large pile", 2500, inventory.CursorGoldLarge},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// when
			got := inventory.GoldCursor(tt.value)
			// then
			if got != tt.want {
				t.Fatalf("GoldCursor(%d) = %d, want %d", tt.value, got, tt.want)
			}
		})
	}
}

func TestCalculateGold(t *testing.T) {
	// given four gold piles among other items
	p := newPlayer(t)
	inv := newView(t, p)
	stack(p, 0, 100)
	stack(p, 2, 5000)
	p.InvList[1] = newItem(cursConsumable)
	stack(p, 3, 750)
	p.NumInv = 4
	// when
	got := inv.CalculateGold()
	// then
	if got != 5850 {
		t.Fatalf("CalculateGold = %d, want 5850", got)
	}
}

func TestRoomForGold(t *testing.T) {
	tests := []struct {
		name   string
		occupi func(p *playerFixture)
		want   int
	}{
		{name: "empty inventory", occupi: func(*playerFixture) {}, want: 40 * inventory.MaxGoldLimit},
		{
			name: "non-gold items leave no room",
			occupi: func(p *playerFixture) {
				p.InvGrid[0] = 1
				p.InvList[0] = newItem(cursArmor)
				p.NumInv = 1
			},
			want: 39 * inventory.MaxGoldLimit,
		},
		{
			name: "partial pile tops off",
			occupi: func(p *playerFixture) {
				p.InvGrid[0] = 1
				stack(p, 0, 1000)
				p.NumInv = 1
			},
			want: 39*inventory.MaxGoldLimit + (inventory.MaxGoldLimit - 1000),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			p := newPlayer(t)
			inv := newView(t, p)
			tt.occupi(p)
			// when
			got := inv.RoomForGold()
			// then
			if got != tt.want {
				t.Fatalf("RoomForGold = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestAddGold(t *testing.T) {
	tests := []struct {
		name      string
		occupi    func(p *playerFixture)
		value     int
		wantLeft  int
		wantCount int
	}{
		{
			name:      "fresh stack in the last row",
			occupi:    func(*playerFixture) {},
			value:     100,
			wantLeft:  0,
			wantCount: 1,
		},
		{
			name: "tops off an existing pile first",
			occupi: func(p *playerFixture) {
				p.InvGrid[0] = 1
				stack(p, 0, 4000)
				p.NumInv = 1
			},
			value:     1500,
			wantLeft:  0,
			wantCount: 2,
		},
		{
			name:      "spills into as many cells as it needs",
			occupi:    func(*playerFixture) {},
			value:     12000,
			wantLeft:  0,
			wantCount: 3,
		},
		{
			name: "leaves the overflow behind when only one cell is free",
			occupi: func(p *playerFixture) {
				for i := range 39 {
					p.InvGrid[i] = 1
					p.InvList[i] = newItem(cursConsumable)
				}
				p.NumInv = 39
			},
			value:     12000,
			wantLeft:  7000,
			wantCount: 40,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			p := newPlayer(t)
			inv := newView(t, p)
			tt.occupi(p)
			// when
			left := inv.AddGold(tt.value)
			// then
			r := goldResult{name: tt.name, got: left, want: tt.wantLeft, gotCount: p.NumInv, wantCount: tt.wantCount}
			assertGold(t, r)
		})
	}
}

func TestGoldAutoPlace(t *testing.T) {
	// given a gold stack held on the cursor
	p := newPlayer(t)
	inv := newView(t, p)
	held := newGold()
	held.Ivalue = 3000
	// when
	placed := inv.GoldAutoPlace(&held)
	// then
	if !placed {
		t.Fatal("GoldAutoPlace = false, want true")
	}
	if held.Ivalue != 0 {
		t.Fatalf("held stack = %d, want 0", held.Ivalue)
	}
	if inventory.IsEmpty(p.InvList[0]) || p.InvList[0].Ivalue != 3000 {
		t.Fatalf("InvList[0] = %+v, want a 3000 gold stack", p.InvList[0])
	}
	if p.Gold != 3000 {
		t.Fatalf("player gold = %d, want 3000", p.Gold)
	}
}

func TestGoldAutoPlaceOverflow(t *testing.T) {
	// given an inventory with no free cell
	p := newPlayer(t)
	inv := newView(t, p)
	for i := range 40 {
		p.InvGrid[i] = 1
		p.InvList[i] = newItem(cursConsumable)
	}
	p.NumInv = 40
	held := newGold()
	held.Ivalue = 700
	// when
	placed := inv.GoldAutoPlace(&held)
	// then
	if placed {
		t.Fatal("GoldAutoPlace without room = true, want false")
	}
	if held.Ivalue != 700 {
		t.Fatalf("held stack = %d, want the untouched 700", held.Ivalue)
	}
}

func TestSplitGold(t *testing.T) {
	tests := []struct {
		name      string
		pile      int
		amount    int
		wantPile  int
		wantHeld  int
		wantEmpty bool
	}{
		{name: "partial split keeps the pile", pile: 400, amount: 300, wantPile: 100, wantHeld: 300},
		{name: "whole pile moves to the cursor", pile: 400, amount: 400, wantEmpty: true, wantHeld: 400},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given a gold pile in the grid
			p, inv := newInv(t)
			stack(p, 0, tt.pile)
			p.InvGrid[30] = 1
			p.NumInv = 1
			// when the player splits the requested amount out
			inv.SplitGold(0, tt.amount)
			// then the cursor carries the split and the pile keeps the rest
			assertSplit(t, splitResult{p: p, wantPile: tt.wantPile, wantHeld: tt.wantHeld, wantEmpty: tt.wantEmpty})
		})
	}
}

func TestDropGold(t *testing.T) {
	for _, tt := range dropCases() {
		t.Run(tt.name, func(t *testing.T) {
			// given piles of gold in the grid and a sink recording the drops
			p, inv := dropInv(t)
			for i, value := range tt.piles {
				stack(p, i, value)
				p.InvGrid[30+i] = int8(i + 1)
			}
			p.NumInv = len(tt.piles)
			dropped = nil
			// when the player drops the requested amount
			left := tt.drop(inv, tt.amount)
			// then the leftovers and the spawned stacks match upstream
			assertDrop(t, dropResult{
				name:     tt.name,
				left:     left,
				wantLeft: tt.wantLeft,
				dropped:  dropped,
				wantDead: tt.wantDead,
				numIn:    p.NumInv,
				wantNum:  tt.wantNumIn,
			})
		})
	}
}

// dropCases are the DropGold rows: piles, requested amount, the variant under
// test, the expected leftover, the spawned stacks and the surviving count.
func dropCases() []dropCase {
	return []dropCase{
		{
			name:      "split one pile",
			piles:     []int{400},
			amount:    300,
			drop:      inventory.Inv.DropGold,
			wantDead:  []int{300},
			wantNumIn: 1,
		},
		{
			name:      "drain the first pile completely",
			piles:     []int{400, 500},
			amount:    700,
			drop:      inventory.Inv.DropGold,
			wantDead:  []int{400, 300},
			wantNumIn: 1,
		},
		{
			name:      "stopped by the total",
			piles:     []int{400, 500},
			amount:    1200,
			drop:      inventory.Inv.DropGold,
			wantLeft:  300,
			wantDead:  []int{400, 500},
			wantNumIn: 0,
		},
		{
			name:      "skipping full stacks",
			piles:     []int{inventory.MaxGoldLimit, 600},
			amount:    1000,
			drop:      inventory.Inv.DropGoldSkippingFullStacks,
			wantLeft:  400,
			wantDead:  []int{600},
			wantNumIn: 1,
		},
	}
}

// dropInv returns a player whose inventory drops through the recording sink.
func dropInv(t *testing.T) (*playerFixture, inventory.Inv) {
	t.Helper()
	p, _ := newInv(t)
	return p, inventory.New(p, inventory.Data{
		Sprite:   testSprites,
		Usable:   func(int16) bool { return true },
		MaxGold:  inventory.MaxGoldLimit,
		DeadItem: func(item entities.Item) { dropped = append(dropped, item.Ivalue) },
	})
}

// assertSplit checks the split-off stack, the remaining pile and player gold.
func assertSplit(t *testing.T, r splitResult) {
	t.Helper()
	if r.p.HoldItem.Ivalue != r.wantHeld {
		t.Fatalf("held = %d, want %d", r.p.HoldItem.Ivalue, r.wantHeld)
	}
	empty := inventory.IsEmpty(r.p.InvList[0])
	if empty != r.wantEmpty {
		t.Fatalf("pile empty = %v, want %v", empty, r.wantEmpty)
	}
	if !r.wantEmpty && r.p.InvList[0].Ivalue != r.wantPile {
		t.Fatalf("pile = %d, want %d", r.p.InvList[0].Ivalue, r.wantPile)
	}
	if r.p.Gold != r.wantPile {
		t.Fatalf("player gold = %d, want %d", r.p.Gold, r.wantPile)
	}
}

// assertDrop checks the leftover, the spawned stacks and the surviving items.
func assertDrop(t *testing.T, r dropResult) {
	t.Helper()
	if r.left != r.wantLeft {
		t.Fatalf("%s: leftover = %d, want %d", r.name, r.left, r.wantLeft)
	}
	if len(r.dropped) != len(r.wantDead) {
		t.Fatalf("%s: dropped %v, want %v", r.name, r.dropped, r.wantDead)
	}
	for i := range r.dropped {
		if r.dropped[i] != r.wantDead[i] {
			t.Fatalf("%s: dropped %v, want %v", r.name, r.dropped, r.wantDead)
		}
	}
	if r.numIn != r.wantNum {
		t.Fatalf("%s: NumInv = %d, want %d", r.name, r.numIn, r.wantNum)
	}
}

// assertGold checks a gold-placement observation.
func assertGold(t *testing.T, r goldResult) {
	t.Helper()
	if r.got != r.want {
		t.Fatalf("%s: got %d, want %d", r.name, r.got, r.want)
	}
	if r.wantCount != 0 && r.gotCount != r.wantCount {
		t.Fatalf("%s: NumInv = %d, want %d", r.name, r.gotCount, r.wantCount)
	}
}
