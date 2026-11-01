package stores_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/entities"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
	"github.com/bmcszk/devilutionGo/internal/diablo/inventory"
	"github.com/bmcszk/devilutionGo/internal/diablo/stores"
)

// holdCases is the pinned hold-acceptance table (stores.cpp:610-626,
// 836-849, 1084-1093).
var holdCases = []holdCase{
	{"empty", inventory.EmptyItem(), false, false, false},
	{
		name:       "damaged sword",
		item:       entities.Item{Itype: int8(inventory.ItemTypeSword), MaxDur: 100, Durability: 40},
		wantRepair: true,
	},
	{"full sword", entities.Item{Itype: int8(inventory.ItemTypeSword), MaxDur: 100, Durability: 100}, false, false, false},
	{
		name:       "indestructible sword",
		item:       entities.Item{Itype: int8(inventory.ItemTypeSword), MaxDur: itemgen.DurIndestructible},
		wantRepair: false,
	},
	{
		name: "misc is never repaired",
		item: entities.Item{Itype: int8(inventory.ItemTypeMisc), MaxDur: 100, Durability: 1},
	},
	{
		"gold is never repaired",
		entities.Item{Itype: int8(inventory.ItemTypeGold), MaxDur: 100, Durability: 1},
		false,
		false,
		false,
	},
	{
		name:         "spent staff recharges",
		item:         entities.Item{Itype: int8(inventory.ItemTypeStaff), Charges: 3, MaxCharges: 20},
		wantRecharge: true,
	},
	{
		name:         "full staff does not recharge",
		item:         entities.Item{Itype: int8(inventory.ItemTypeStaff), Charges: 20, MaxCharges: 20},
		wantRecharge: false,
	},
	{
		name:         "unique charged relic recharges",
		item:         entities.Item{Itype: int8(inventory.ItemTypeMisc), MiscId: 27, Charges: 1, MaxCharges: 5},
		wantRecharge: true,
	},
	{
		name:         "unidentified magic item is identifiable",
		item:         entities.Item{Itype: int8(inventory.ItemTypeSword), Magical: 1},
		wantIdentify: true,
	},
	{
		name:         "identified magic item is not",
		item:         entities.Item{Itype: int8(inventory.ItemTypeSword), Magical: 1, Identified: true},
		wantIdentify: false,
	},
	{
		name: "normal item is not identifiable",
		item: entities.Item{Itype: int8(inventory.ItemTypeSword)},
	},
}

// repairCase is one row of the repair-price table.
type repairCase struct {
	name      string
	item      entities.Item
	want      int
	wantOffer bool
}

// holdCase is one row of the hold-acceptance table.
type holdCase struct {
	name         string
	item         entities.Item
	wantRepair   bool
	wantRecharge bool
	wantIdentify bool
}

// buyCase is one row of the vendor acceptance table.
type buyCase struct {
	name  string
	item  entities.Item
	env   stores.Env
	smith bool
	witch bool
}

// TestSellValue pins the quarter-value sell price, including the identified
// magic item's snap to Ivalue and the one-gold floor (stores.cpp:558-562).
func TestSellValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		item entities.Item
		want int
	}{
		{"normal item uses Value", entities.Item{Value: 1000, Ivalue: 4000}, 250},
		{"identified magic uses Ivalue", entities.Item{Magical: 1, Identified: true, Value: 500, Ivalue: 900}, 225},
		{"unidentified magic keeps Value", entities.Item{Magical: 1, Value: 500, Ivalue: 900}, 125},
		{"two gold floors at one", entities.Item{Value: 2}, 1},
		{"eight gold quarters to two", entities.Item{Value: 8}, 2},
		{"zero value floors at one", entities.Item{}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when the item is sold
			got := stores.SellValue(tt.item)

			// then the quarter-value rules hold
			assertValue(t, got, tt.want)
		})
	}
}

// TestRepairPrice pins AddStoreHoldRepair's two price branches (stores.cpp:2218-2233).
func TestRepairPrice(t *testing.T) {
	t.Parallel()

	tests := []repairCase{
		{
			name:      "normal item quarters the missing durability share",
			item:      entities.Item{Value: 1000, MaxDur: 100, Durability: 40},
			want:      300,
			wantOffer: true,
		},
		{
			name:      "identified magic item uses the thirty percent formula",
			item:      entities.Item{Magical: 1, Identified: true, Ivalue: 1000, MaxDur: 100, Durability: 40},
			want:      90,
			wantOffer: true,
		},
		{
			name:      "identified magic item with a rounding zero is not offered",
			item:      entities.Item{Magical: 1, Identified: true, Ivalue: 100, MaxDur: 100, Durability: 99},
			want:      0,
			wantOffer: false,
		},
		{
			name:      "normal item floors at one gold",
			item:      entities.Item{Value: 1, MaxDur: 100, Durability: 99},
			want:      1,
			wantOffer: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when the item is repaired
			price, offer := stores.RepairPrice(tt.item)

			// then the price and the offer flag follow the two branches
			assertValue(t, price, tt.want)
			assertValue(t, offer, tt.wantOffer)
		})
	}
}

// TestRechargePrice pins AddStoreHoldRecharge (stores.cpp:851-859).
func TestRechargePrice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		item      entities.Item
		staffCost int
		want      int
	}{
		{"half empty staff", entities.Item{Value: 1000, Charges: 5, MaxCharges: 20}, 600, 600},
		{"full staff costs nothing", entities.Item{Value: 1000, Charges: 20, MaxCharges: 20}, 600, 0},
		{"no staff cost", entities.Item{Value: 1000, Charges: 10, MaxCharges: 20}, 0, 250},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when the staff is recharged
			got := stores.RechargePrice(tt.item, tt.staffCost)

			// then the price is the staff cost or the missing-charge share
			assertValue(t, got, tt.want)
		})
	}
}

// TestBoyPrice pins Wirt's haggle: a quarter off in Hellfire, half again
// otherwise (stores.cpp:1821-1824).
func TestBoyPrice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		env  stores.Env
		want int
	}{
		{"vanilla adds half", stores.Env{}, 1500},
		{"hellfire takes a quarter", stores.Env{Hellfire: true}, 750},
	}

	item := entities.Item{Ivalue: 1000}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when he haggles
			got := stores.BoyPrice(item, tt.env)

			// then Hellfire takes a quarter and vanilla adds half
			assertValue(t, got, tt.want)
		})
	}
}

// TestGoldChecks pins PlayerCanAfford, TotalPlayerGold and StoreGoldFit
// (stores.cpp:459-462, 1414-1426, 2861-2864).
func TestGoldChecks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		gold  int
		price int
		want  bool
	}{
		{"exact gold affords", 1500, 1500, true},
		{"one short does not", 1499, 1500, false},
		{"free costs nothing", 0, 0, true},
		{"negative price is unaffordable", 5000, -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when the pair is checked
			got := stores.CanAfford(tt.gold, tt.price)

			// then the purse only covers a non-negative price it reaches
			assertValue(t, got, tt.want)
		})
	}

	fitTests := []struct {
		name            string
		cost            int
		itemRoomForGold int
		roomForGold     int
		want            bool
	}{
		{"fits in the item's own cells", 5000, 20000, 0, true},
		{"fits with free cells", 25000, 20000, 5000, true},
		{"does not fit", 25001, 20000, 5000, false},
	}

	for _, tt := range fitTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when the price is fitted into the pack and the stash
			got := stores.GoldFit(tt.cost, tt.itemRoomForGold, tt.roomForGold)

			// then only the covered share counts
			assertValue(t, got, tt.want)
		})
	}

	t.Run("total gold", func(t *testing.T) {
		t.Parallel()

		// when a character holding 100 gold and a stash holding 50 are summed
		got := stores.TotalGold(100, 50)

		// then the total is both purses
		assertValue(t, got, 150)
	})
}

// TestBuyRules pins SmithWillBuy and WitchWillBuy (stores.cpp:2866-2908).
func TestBuyRules(t *testing.T) {
	t.Parallel()

	hellfire := stores.Env{Hellfire: true, SpellCount: 64}

	tests := []buyCase{
		{"empty item", inventory.EmptyItem(), stores.Env{}, false, false},
		{"plain sword", entities.Item{Itype: int8(inventory.ItemTypeSword), IDidx: 1}, stores.Env{}, true, false},
		{"oil sells to both", entities.Item{Itype: int8(inventory.ItemTypeMisc), MiscId: 30}, stores.Env{}, true, false},
		{"potion is misc", entities.Item{Itype: int8(inventory.ItemTypeMisc)}, stores.Env{}, false, true},
		{"gold", entities.Item{Itype: int8(inventory.ItemTypeGold)}, stores.Env{}, false, false},
		{"staff never sells to the smith", entities.Item{Itype: int8(inventory.ItemTypeStaff)}, stores.Env{}, false, true},
		{
			name:  "hellfire staff with an invalid spell sells to the smith",
			item:  entities.Item{Itype: int8(inventory.ItemTypeStaff)},
			env:   hellfire,
			smith: true,
			witch: false,
		},
		{
			name:  "hellfire staff with a valid spell stays with the witch",
			item:  entities.Item{Itype: int8(inventory.ItemTypeStaff), Spell: 1},
			env:   hellfire,
			smith: false,
			witch: true,
		},
		{"quest class", entities.Item{Itype: int8(inventory.ItemTypeSword), Class: 5, IDidx: 1}, stores.Env{}, false, false},
		{"lazarus staff", entities.Item{Itype: int8(inventory.ItemTypeSword), IDidx: 33}, stores.Env{}, false, false},
		{"quest index", entities.Item{Itype: int8(inventory.ItemTypeMisc), IDidx: 6}, stores.Env{}, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when the item is offered to both vendors
			smith := stores.SmithWillBuy(tt.item, tt.env)
			witch := stores.WitchWillBuy(tt.item, tt.env)

			// then each takes only what upstream allows
			assertValue(t, smith, tt.smith)
			assertValue(t, witch, tt.witch)
		})
	}
}

// TestHoldRules pins SmithRepairOk, WitchRechargeOk and IdItemOk
// (stores.cpp:610-626, 836-849, 1084-1093).
func TestHoldRules(t *testing.T) {
	t.Parallel()

	for _, tt := range holdCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when the item is offered to the three lists
			repair := stores.Repairable(tt.item)
			recharge := stores.Rechargeable(tt.item)
			identify := stores.Identifiable(tt.item)

			// then every acceptance rule matches upstream
			assertValue(t, repair, tt.wantRepair)
			assertValue(t, recharge, tt.wantRecharge)
			assertValue(t, identify, tt.wantIdentify)
		})
	}
}

// TestTextColor pins getTextColorWithStatCheck (items.h:435-451).
func TestTextColor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		item entities.Item
		want stores.Color
	}{
		{"unusable is red", entities.Item{}, stores.ColorRed},
		{"plain is white", entities.Item{StatFlag: true}, stores.ColorWhite},
		{"magic is blue", entities.Item{StatFlag: true, Magical: 1}, stores.ColorBlue},
		{"unique is whitegold", entities.Item{StatFlag: true, Magical: 2}, stores.ColorWhitegold},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when the item is drawn in a store list
			got := stores.TextColor(tt.item)

			// then its color follows usability, quality and identification
			assertValue(t, got, tt.want)
		})
	}
}

// assertValue fails the test unless got equals want.
func assertValue[T comparable](t *testing.T, got, want T) {
	t.Helper()

	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}
