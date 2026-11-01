package inventory_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/inventory"
)

// goldResult is one gold-placement observation plus its expectation.
type goldResult struct {
	name      string
	got       int
	want      int
	gotCount  int
	wantCount int
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
