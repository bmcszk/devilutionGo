package stores_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/gameplay/stores"
)

// TestTalkIDValues pins the ordinal of every TalkID state against
// Source/stores.h:44-70 (upstream enum order is ABI, not decoration).
func TestTalkIDValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		id   stores.TalkID
		want stores.TalkID
	}{
		{"None", stores.TalkIDNone, 0},
		{"Smith", stores.TalkIDSmith, 1},
		{"SmithBuy", stores.TalkIDSmithBuy, 2},
		{"SmithSell", stores.TalkIDSmithSell, 3},
		{"SmithRepair", stores.TalkIDSmithRepair, 4},
		{"Witch", stores.TalkIDWitch, 5},
		{"WitchBuy", stores.TalkIDWitchBuy, 6},
		{"WitchSell", stores.TalkIDWitchSell, 7},
		{"WitchRecharge", stores.TalkIDWitchRecharge, 8},
		{"NoMoney", stores.TalkIDNoMoney, 9},
		{"NoRoom", stores.TalkIDNoRoom, 10},
		{"Confirm", stores.TalkIDConfirm, 11},
		{"Boy", stores.TalkIDBoy, 12},
		{"BoyBuy", stores.TalkIDBoyBuy, 13},
		{"Healer", stores.TalkIDHealer, 14},
		{"Storyteller", stores.TalkIDStoryteller, 15},
		{"HealerBuy", stores.TalkIDHealerBuy, 16},
		{"StorytellerIdentify", stores.TalkIDStorytellerIdentify, 17},
		{"SmithPremiumBuy", stores.TalkIDSmithPremiumBuy, 18},
		{"Gossip", stores.TalkIDGossip, 19},
		{"StorytellerIdentifyShow", stores.TalkIDStorytellerIdentifyShow, 20},
		{"Tavern", stores.TalkIDTavern, 21},
		{"Drunk", stores.TalkIDDrunk, 22},
		{"Barmaid", stores.TalkIDBarmaid, 23},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := uint8(tt.id); got != uint8(tt.want) {
				t.Fatalf("TalkID %s = %d, want %d", tt.name, got, uint8(tt.want))
			}
		})
	}
}
