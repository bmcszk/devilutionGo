package inventory_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/entities"
	"github.com/bmcszk/devilutionGo/internal/diablo/inventory"
)

// canUseScrollCase is one CanUseScroll row.
type canUseScrollCase struct {
	name  string
	setup func(p *playerFixture)
	spell int32
	town  bool
	allow bool
	want  bool
}

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

func TestHasInventoryOrBeltItem(t *testing.T) {
	// given a scroll lying in the belt and a potion in the grid
	p, _ := newInv(t)
	inv := newView(t, p)
	p.InvList[0] = newItem(cursConsumable)
	p.NumInv = 1
	p.InvGrid[30] = 1
	p.SpdList[1] = newScrollOf(0)
	isScroll := func(item entities.Item) bool { return inventory.IsScroll(item) }
	// when each search runs over that state
	got := map[string]bool{
		"inventory":   inv.HasInventoryItem(isScroll),
		"belt":        inv.HasBeltItem(isScroll),
		"either":      inv.HasInventoryOrBeltItem(isScroll),
		"unknown id":  inv.HasBeltItem(func(entities.Item) bool { return false }),
		"missing inv": inv.HasInventoryItem(func(entities.Item) bool { return false }),
	}
	// then only the belt search finds the scroll, and the union does too
	want := map[string]bool{
		"inventory":   false,
		"belt":        true,
		"either":      true,
		"unknown id":  false,
		"missing inv": false,
	}
	for name, wantValue := range want {
		if got[name] != wantValue {
			t.Fatalf("%s = %v, want %v", name, got[name], wantValue)
		}
	}
}

func TestHasInventoryItemIgnoresTail(t *testing.T) {
	// given a scroll stored past NumInv, i.e. not live
	p, _ := newInv(t)
	inv := newView(t, p)
	p.InvList[3] = newScrollOf(0)
	p.NumInv = 2
	// when
	found := inv.HasInventoryItem(inventory.IsScroll)
	// then the search stops at NumInv
	if found {
		t.Fatal("HasInventoryItem found a slot past NumInv, want false")
	}
}

func TestRemoveInvItemMirrorsUpstream(t *testing.T) {
	// given a lone two-slot misc item: | item | item |
	p, _ := newInv(t)
	inv := newView(t, p)
	p.NumInv = 1
	p.InvGrid[0] = 1
	p.InvGrid[1] = -1
	p.InvList[0] = withType(newItem(cursConsumable), inventory.ItemTypeMisc)
	// when the item is removed
	inv.Remove(0)
	// then all cells and the count are cleared
	if p.InvGrid[0] != 0 || p.InvGrid[1] != 0 {
		t.Fatalf("grid = %d,%d, want 0,0", p.InvGrid[0], p.InvGrid[1])
	}
	if p.NumInv != 0 {
		t.Fatalf("NumInv = %d, want 0", p.NumInv)
	}
}

// TestRemoveInvItemShiftsFromMiddle ports test/inv_test.cpp:195-226
// (RemoveInvItem_shiftsListFromMiddle).
func TestRemoveInvItemShiftsFromMiddle(t *testing.T) {
	// given | item | item | ring | item | item |
	p, _ := newInv(t)
	inv := newView(t, p)
	p.NumInv = 3
	p.InvGrid[0] = 1
	p.InvGrid[1] = -1
	p.InvList[0] = withType(newItem(cursConsumable), inventory.ItemTypeMisc)
	p.InvGrid[2] = 2
	p.InvList[1] = withType(newItem(cursConsumable), inventory.ItemTypeRing)
	p.InvGrid[3] = 3
	p.InvGrid[4] = -3
	p.InvList[2] = withType(newItem(cursConsumable), inventory.ItemTypeMisc)
	// when the middle item is removed
	inv.Remove(1)
	// then the grid drops the ring's cell and renumbers the tail
	wantGrid := map[int]int8{0: 1, 1: -1, 2: 0, 3: 2, 4: -2}
	for i, want := range wantGrid {
		if p.InvGrid[i] != want {
			t.Fatalf("InvGrid[%d] = %d, want %d", i, p.InvGrid[i], want)
		}
	}
	if inventory.ItemType(p.InvList[0].Itype) != inventory.ItemTypeMisc ||
		inventory.ItemType(p.InvList[1].Itype) != inventory.ItemTypeMisc {
		t.Fatalf("types = %d,%d, want misc,misc", p.InvList[0].Itype, p.InvList[1].Itype)
	}
	if p.NumInv != 2 {
		t.Fatalf("NumInv = %d, want 2", p.NumInv)
	}
}

// TestRemoveInvItemShiftsFromFront ports test/inv_test.cpp:228-259
// (RemoveInvItem_shiftsListFromFront).
func TestRemoveInvItemShiftsFromFront(t *testing.T) {
	// given the same layout as the middle case
	p, _ := newInv(t)
	inv := newView(t, p)
	p.NumInv = 3
	p.InvGrid[0] = 1
	p.InvGrid[1] = -1
	p.InvList[0] = withType(newItem(cursConsumable), inventory.ItemTypeMisc)
	p.InvGrid[2] = 2
	p.InvList[1] = withType(newItem(cursConsumable), inventory.ItemTypeRing)
	p.InvGrid[3] = 3
	p.InvGrid[4] = -3
	p.InvList[2] = withType(newItem(cursConsumable), inventory.ItemTypeMisc)
	// when the first item is removed
	inv.Remove(0)
	// then both of its cells empty and every later slot shifts down
	wantGrid := map[int]int8{0: 0, 1: 0, 2: 1, 3: 2, 4: -2}
	for i, want := range wantGrid {
		if p.InvGrid[i] != want {
			t.Fatalf("InvGrid[%d] = %d, want %d", i, p.InvGrid[i], want)
		}
	}
	if inventory.ItemType(p.InvList[0].Itype) != inventory.ItemTypeRing ||
		inventory.ItemType(p.InvList[1].Itype) != inventory.ItemTypeMisc {
		t.Fatalf("types = %d,%d, want ring,misc", p.InvList[0].Itype, p.InvList[1].Itype)
	}
	if p.NumInv != 2 {
		t.Fatalf("NumInv = %d, want 2", p.NumInv)
	}
}

func TestCanUseStaff(t *testing.T) {
	tests := []struct {
		name  string
		item  func() entities.Item
		spell int32
		want  bool
	}{
		{name: "charged staff of the spell", item: chargedStaff, spell: 12, want: true},
		{name: "wrong spell", item: chargedStaff, spell: 13, want: false},
		{name: "spent staff", item: spentStaff, spell: 12, want: false},
		{name: "scroll is not a staff", item: func() entities.Item { return newScrollOf(12) }, spell: 12, want: false},
		{name: "empty hand", item: inventory.EmptyItem, spell: 12, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// when
			got := inventory.CanUseStaff(tt.item(), tt.spell)
			// then inv.cpp:1079-1085
			if got != tt.want {
				t.Fatalf("CanUseStaff = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsScrollOfAndRuneOf(t *testing.T) {
	// given scrolls and runes of different spells
	fireboltScroll := newScrollOf(1)
	healingScroll := newScrollOf(2)
	fireRune := newRuneOf(inventory.MiscRuneF)
	stoneRune := newRuneOf(inventory.MiscRuneS)
	tests := []struct {
		name  string
		item  entities.Item
		spell int32
		want  bool
	}{
		{name: "scroll of the spell", item: fireboltScroll, spell: 1, want: true},
		{name: "scroll of another spell", item: healingScroll, spell: 1, want: false},
		{name: "rune of fire", item: fireRune, spell: inventory.SpellRuneOfFire, want: true},
		{name: "rune of stone", item: stoneRune, spell: inventory.SpellRuneOfStone, want: true},
		{
			name:  "rune asked for another spell",
			item:  fireRune,
			spell: inventory.SpellRuneOfLight,
			want:  false,
		},
		{name: "plain potion", item: newItem(cursConsumable), spell: 1, want: false},
	}
	// when each item is checked for the spell it should carry
	mismatches := make([]string, 0, len(tests))
	for _, tt := range tests {
		got := inventory.IsScrollOf(tt.item, tt.spell) || inventory.IsRuneOf(tt.item, tt.spell)
		if got != tt.want {
			mismatches = append(mismatches, tt.name)
		}
	}
	// then only the matching pairs report true
	if len(mismatches) > 0 {
		t.Fatalf("predicate mismatches: %v", mismatches)
	}
}

func TestCanUseScroll(t *testing.T) {
	for _, tt := range canUseScrollCases() {
		t.Run(tt.name, func(t *testing.T) {
			// given a player holding the listed scrolls
			p, _ := newInv(t)
			// when
			got := canUseScroll(p, tt)
			// then inv.cpp:2054-2062 decides it
			if got != tt.want {
				t.Fatalf("CanUseScroll = %v, want %v", got, tt.want)
			}
		})
	}
}

// canUseScroll wires the case into a view and runs the check.
func canUseScroll(p *playerFixture, c canUseScrollCase) bool {
	c.setup(p)
	inv := inventory.New(p, inventory.Data{
		Sprite:             testSprites,
		InTown:             func() bool { return c.town },
		SpellAllowedInTown: func(int32) bool { return c.allow },
	})
	return inv.CanUseScroll(c.spell)
}

// canUseScrollCases are the UseScroll rows ported from test/inv_test.cpp:68-95
// plus the town rule of inv.cpp:2056.
func canUseScrollCases() []canUseScrollCase {
	return []canUseScrollCase{
		{
			name:  "scroll in the grid outside town",
			setup: func(p *playerFixture) { p.InvList[2] = newScrollOf(1); p.NumInv = 5 },
			spell: 1,
			want:  true,
		},
		{
			name:  "scroll in the belt outside town",
			setup: func(p *playerFixture) { p.SpdList[2] = newScrollOf(1) },
			spell: 1,
			want:  true,
		},
		{name: "no scroll anywhere", setup: func(*playerFixture) {}, spell: 1, want: false},
		{
			name:  "in town with a town-forbidden spell",
			setup: func(p *playerFixture) { p.InvList[2] = newScrollOf(1); p.NumInv = 5 },
			spell: 1,
			town:  true,
			want:  false,
		},
		{
			name:  "in town with a town spell",
			setup: func(p *playerFixture) { p.InvList[2] = newScrollOf(2); p.NumInv = 5 },
			spell: 2,
			town:  true,
			allow: true,
			want:  true,
		},
	}
}

// chargedStaff is a staff of spell 12 with charges left.
func chargedStaff() entities.Item {
	staff := newItem(cursSword)
	staff.MiscId = inventory.MiscStaff
	staff.Spell = 12
	staff.Charges = 3
	return staff
}

// spentStaff is the same staff with no charges.
func spentStaff() entities.Item {
	staff := chargedStaff()
	staff.Charges = 0
	return staff
}

// newRuneOf is a rune carrying the given misc id.
func newRuneOf(miscID int8) entities.Item {
	runeItem := newItem(cursConsumable)
	runeItem.MiscId = miscID
	return runeItem
}

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

// TestRemoveInvItemMirrorsUpstream ports test/inv_test.cpp:176-193
// (RemoveInvItem): a two-slot item leaves both cells empty.
