package inventory_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/entities"
	"github.com/bmcszk/devilutionGo/internal/diablo/inventory"
)

func TestIsScrollAndRune(t *testing.T) {
	tests := []struct {
		name   string
		miscID int8
		scroll bool
		rune   bool
	}{
		{"potion is neither", inventory.MiscNone, false, false},
		{"identify scroll", inventory.MiscScroll, true, false},
		{"town portal scroll", inventory.MiscScrollT, true, false},
		{"rune of fire", inventory.MiscRuneFirst + 1, false, true},
		{"rune of the rift", inventory.MiscRuneLast - 1, false, true},
		{"map of the doom", inventory.MiscMapOfDoom, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			item := newScroll(tt.miscID)
			// when
			isScroll := inventory.IsScroll(item)
			isRune := inventory.IsRune(item)
			// then
			if isScroll != tt.scroll || isRune != tt.rune {
				t.Fatalf("IsScroll/IsRune = %v/%v, want %v/%v", isScroll, isRune, tt.scroll, tt.rune)
			}
		})
	}
}

func TestRemoveInvItem(t *testing.T) {
	// given three items in the grid, removing the middle one
	p, inv := newInv(t)
	for range 3 {
		inv.Place(newItem(cursConsumable))
	}
	second := p.InvList[1].Curs
	third := p.InvList[2].Curs
	// when
	inv.Remove(1)
	// then
	if p.NumInv != 2 {
		t.Fatalf("NumInv = %d, want 2", p.NumInv)
	}
	if p.InvList[1].Curs != third {
		t.Fatalf("InvList[1].Curs = %d, want the shifted %d", p.InvList[1].Curs, third)
	}
	if !inventory.IsEmpty(p.InvList[2]) {
		t.Fatal("the tail slot should be empty after the shift")
	}
	if p.InvList[0].Curs != second-0 {
		t.Fatalf("InvList[0].Curs = %d, want %d", p.InvList[0].Curs, second)
	}
}

func TestRemoveClearsPlaceholders(t *testing.T) {
	// given a 1x3 sword taking three cells
	p, inv := newInv(t)
	p.HoldItem = newItem(cursSword)
	inv.PutInGrid(0, inventory.Size{Width: 1, Height: 3})
	// when
	inv.Remove(0)
	// then
	for _, cell := range []int{0, 10, 20} {
		if p.InvGrid[cell] != 0 {
			t.Fatalf("InvGrid[%d] = %d, want 0", cell, p.InvGrid[cell])
		}
	}
}

func TestConsumeBeltSlot(t *testing.T) {
	tests := []struct {
		name      string
		item      entities.Item
		wantEmpty bool
	}{
		{name: "potion leaves the belt", item: newItem(cursConsumable), wantEmpty: true},
		{name: "scroll stays until it is consumed", item: newScroll(inventory.MiscScroll), wantEmpty: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			p, inv := newInv(t)
			p.SpdList[0] = tt.item
			// when
			inv.ConsumeBeltSlot(0)
			// then
			if empty := inventory.IsEmpty(p.SpdList[0]); empty != tt.wantEmpty {
				t.Fatalf("belt slot empty = %v, want %v", empty, tt.wantEmpty)
			}
		})
	}
}

func TestConsumeScrollFromSelectedSlot(t *testing.T) {
	tests := []struct {
		name      string
		spellFrom int8
		place     func(p *playerFixture)
		wantSlot  func(p *playerFixture) bool
	}{
		{
			name:      "inventory slot",
			spellFrom: inventory.InvItemFirst,
			place: func(p *playerFixture) {
				p.InvList[0] = newScrollOf(7)
				p.InvGrid[30] = 1
				p.NumInv = 1
			},
			wantSlot: func(p *playerFixture) bool { return inventory.IsEmpty(p.InvList[0]) },
		},
		{
			name:      "belt slot",
			spellFrom: inventory.BeltItemFirst,
			place: func(p *playerFixture) {
				p.SpdList[0] = newScrollOf(7)
			},
			wantSlot: func(p *playerFixture) bool { return inventory.IsEmpty(p.SpdList[0]) },
		},
		{
			name:      "falls back to the first match",
			spellFrom: 0,
			place: func(p *playerFixture) {
				p.SpdList[2] = newScrollOf(7)
			},
			wantSlot: func(p *playerFixture) bool { return inventory.IsEmpty(p.SpdList[2]) },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			p, inv := newInv(t)
			tt.place(p)
			// when
			inv.ConsumeScroll(tt.spellFrom, 7)
			// then
			if !tt.wantSlot(p) {
				t.Fatal("the scroll was not consumed from the expected slot")
			}
		})
	}
}

func TestConsumeScrollKeepsOtherScroll(t *testing.T) {
	// given a scroll of a different spell in the selected slot
	p, inv := newInv(t)
	p.InvList[0] = newScrollOf(3)
	p.InvGrid[30] = 1
	p.NumInv = 1
	// when the executed spell expects spell 7
	inv.ConsumeScroll(inventory.InvItemFirst, 7)
	// then
	if inventory.IsEmpty(p.InvList[0]) {
		t.Fatal("a scroll of another spell must not be consumed")
	}
}

// newScroll is a scroll item with the given misc id and no spell.
func newScroll(miscID int8) entities.Item {
	item := newItem(cursConsumable)
	item.MiscId = miscID
	return item
}

// newScrollOf is an identify-shaped scroll carrying the given spell id.
func newScrollOf(spell int32) entities.Item {
	item := newScroll(inventory.MiscScroll)
	item.Spell = spell
	return item
}
