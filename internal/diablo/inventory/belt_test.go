package inventory_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/entities"
	"github.com/bmcszk/devilutionGo/internal/diablo/inventory"
)

// beltResult is one PlaceInBelt observation plus its expectation.
type beltResult struct {
	placed   bool
	player   *entities.Player
	want     bool
	wantSlot int
}

func TestIsBeltItem(t *testing.T) {
	tests := []struct {
		name string
		curs uint8
		want bool
	}{
		{"potion", cursConsumable, true},
		{"sword", cursSword, false},
		{"armor", cursArmor, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// when
			got := inventory.IsBeltItem(newItem(tt.curs), testSprites)
			// then
			if got != tt.want {
				t.Fatalf("IsBeltItem(curs %d) = %v, want %v", tt.curs, got, tt.want)
			}
		})
	}
}

func TestCanUse(t *testing.T) {
	tests := []struct {
		name           string
		str, mag, dex  int
		minStr, minMag int
		minDex         int
		want           bool
	}{
		{name: "meets every requirement", str: 20, mag: 10, dex: 15, minStr: 20, minMag: 10, minDex: 15, want: true},
		{name: "short on strength", str: 19, mag: 10, dex: 15, minStr: 20, minMag: 10, minDex: 15, want: false},
		{name: "short on magic", str: 20, mag: 9, dex: 15, minStr: 20, minMag: 10, minDex: 15, want: false},
		{name: "short on dexterity", str: 20, mag: 10, dex: 14, minStr: 20, minMag: 10, minDex: 15, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			p, inv := newInv(t)
			p.Strength, p.Magic, p.Dexterity = tt.str, tt.mag, tt.dex
			item := newItem(cursConsumable)
			item.MinStr = int8(tt.minStr)
			item.MinMag = uint8(tt.minMag)
			item.MinDex = int8(tt.minDex)
			// when
			got := inv.CanUse(item)
			// then
			if got != tt.want {
				t.Fatalf("CanUse = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCanPlaceOnBelt(t *testing.T) {
	tests := []struct {
		name   string
		item   func() entities.Item
		usable bool
		want   bool
	}{
		{
			name:   "usable potion",
			item:   func() entities.Item { return newItem(cursConsumable) },
			usable: true,
			want:   true,
		},
		{name: "gold never goes on the belt", item: newGold, usable: true, want: false},
		{
			name:   "two-cell item does not fit",
			item:   func() entities.Item { return newItem(cursSword) },
			usable: true,
			want:   false,
		},
		{
			name:   "unusable item",
			item:   func() entities.Item { return newItem(cursConsumable) },
			usable: false,
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			p, _ := newInv(t)
			inv := inventory.New(p, inventory.Data{
				Sprite: testSprites,
				Usable: func(int16) bool { return tt.usable },
			})
			// when
			got := inv.CanPlaceOnBelt(tt.item())
			// then
			if got != tt.want {
				t.Fatalf("CanPlaceOnBelt = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlaceInBelt(t *testing.T) {
	tests := []struct {
		name     string
		fill     int
		want     bool
		wantSlot int
	}{
		{name: "first free slot", want: true, wantSlot: 0},
		{name: "first two slots taken", fill: 2, want: true, wantSlot: 2},
		{name: "belt full", fill: inventory.BeltSlots, want: false, wantSlot: -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			p, inv := newInv(t)
			for i := range tt.fill {
				p.SpdList[i] = newItem(cursConsumable)
			}
			// when
			ok := inv.PlaceInBelt(newItem(cursConsumable))
			// then
			assertBeltSlot(t, beltResult{placed: ok, player: p, want: tt.want, wantSlot: tt.wantSlot})
		})
	}
}

func TestRemoveBeltSlot(t *testing.T) {
	// given a belt with a potion in the third slot
	p, inv := newInv(t)
	p.SpdList[2] = newItem(cursConsumable)
	// when
	inv.RemoveBeltSlot(2)
	// then
	if !inventory.IsEmpty(p.SpdList[2]) {
		t.Fatal("SpdList[2] should be empty after RemoveBeltSlot")
	}
	if !inventory.IsEmpty(p.SpdList[1]) {
		t.Fatal("RemoveBeltSlot must not touch other slots")
	}
}

// assertBeltSlot checks the placement outcome and which slot took the item.
func assertBeltSlot(t *testing.T, r beltResult) {
	t.Helper()
	if r.placed != r.want {
		t.Fatalf("PlaceInBelt = %v, want %v", r.placed, r.want)
	}
	if r.wantSlot < 0 {
		return
	}
	if inventory.IsEmpty(r.player.SpdList[r.wantSlot]) {
		t.Fatalf("SpdList[%d] should hold the placed item", r.wantSlot)
	}
}

// newGold is a gold stack with a value but no cursor set yet.
func newGold() entities.Item {
	return entities.Item{Itype: int8(inventory.ItemTypeGold), Ivalue: 100}
}
