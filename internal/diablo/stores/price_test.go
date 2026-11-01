package stores_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/entities"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
	"github.com/bmcszk/devilutionGo/internal/diablo/inventory"
	"github.com/bmcszk/devilutionGo/internal/diablo/stores"
)

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

			if got := stores.SellValue(tt.item); got != tt.want {
				t.Fatalf("SellValue = %d, want %d", got, tt.want)
			}
		})
	}
}

// TestRepairPrice pins AddStoreHoldRepair's two price branches (stores.cpp:2218-2233).
func TestRepairPrice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		item      entities.Item
		want      int
		wantOffer bool
	}{
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

			got, offered := stores.RepairPrice(tt.item)
			if got != tt.want || offered != tt.wantOffer {
				t.Fatalf("RepairPrice = %d, %v; want %d, %v", got, offered, tt.want, tt.wantOffer)
			}
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

			if got := stores.RechargePrice(tt.item, tt.staffCost); got != tt.want {
				t.Fatalf("RechargePrice = %d, want %d", got, tt.want)
			}
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

			if got := stores.BoyPrice(item, tt.env); got != tt.want {
				t.Fatalf("BoyPrice = %d, want %d", got, tt.want)
			}
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

			if got := stores.CanAfford(tt.gold, tt.price); got != tt.want {
				t.Fatalf("CanAfford(%d, %d) = %v, want %v", tt.gold, tt.price, got, tt.want)
			}
		})
	}

	if got := stores.TotalGold(100, 50); got != 150 {
		t.Fatalf("TotalGold = %d, want 150", got)
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

			if got := stores.GoldFit(tt.cost, tt.itemRoomForGold, tt.roomForGold); got != tt.want {
				t.Fatalf("GoldFit = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestBuyRules pins SmithWillBuy and WitchWillBuy (stores.cpp:2866-2908).
func TestBuyRules(t *testing.T) {
	t.Parallel()

	hellfire := stores.Env{Hellfire: true, SpellCount: 64}

	tests := []struct {
		name  string
		item  entities.Item
		env   stores.Env
		smith bool
		witch bool
	}{
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

			if got := stores.SmithWillBuy(tt.item, tt.env); got != tt.smith {
				t.Fatalf("SmithWillBuy = %v, want %v", got, tt.smith)
			}

			if got := stores.WitchWillBuy(tt.item, tt.env); got != tt.witch {
				t.Fatalf("WitchWillBuy = %v, want %v", got, tt.witch)
			}
		})
	}
}

// TestHoldRules pins SmithRepairOk, WitchRechargeOk and IdItemOk
// (stores.cpp:610-626, 836-849, 1084-1093).
func TestHoldRules(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		item         entities.Item
		wantRepair   bool
		wantRecharge bool
		wantIdentify bool
	}{
		{"empty", inventory.EmptyItem(), false, false, false},
		{"damaged sword", entities.Item{Itype: int8(inventory.ItemTypeSword), MaxDur: 100, Durability: 40}, true, false, false},
		{"full sword", entities.Item{Itype: int8(inventory.ItemTypeSword), MaxDur: 100, Durability: 100}, false, false, false},
		{
			name:       "indestructible sword",
			item:       entities.Item{Itype: int8(inventory.ItemTypeSword), MaxDur: itemgen.DurIndestructible},
			wantRepair: false,
		},
		{"misc is never repaired", entities.Item{Itype: int8(inventory.ItemTypeMisc), MaxDur: 100, Durability: 1}, false, false, false},
		{"gold is never repaired", entities.Item{Itype: int8(inventory.ItemTypeGold), MaxDur: 100, Durability: 1}, false, false, false},
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := stores.Repairable(tt.item); got != tt.wantRepair {
				t.Fatalf("Repairable = %v, want %v", got, tt.wantRepair)
			}

			if got := stores.Rechargeable(tt.item); got != tt.wantRecharge {
				t.Fatalf("Rechargeable = %v, want %v", got, tt.wantRecharge)
			}

			if got := stores.Identifiable(tt.item); got != tt.wantIdentify {
				t.Fatalf("Identifiable = %v, want %v", got, tt.wantIdentify)
			}
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

			if got := stores.TextColor(tt.item); got != tt.want {
				t.Fatalf("TextColor = %v, want %v", got, tt.want)
			}
		})
	}
}
