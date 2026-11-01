package stores_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/stores"
)

// TestTalkIDOrdinals pins every TalkID value to its upstream declaration
// position (stores.h:44-69).
func TestTalkIDOrdinals(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		id   stores.TalkID
		want uint8
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

	if len(tests) != stores.TalkIDCount {
		t.Fatalf("pinned %d states, TalkIDCount is %d", len(tests), stores.TalkIDCount)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when the ordinal and its name are read
			ordinal := uint8(tt.id)
			name := tt.id.String()

			// then both match the upstream declaration position
			assertValue(t, ordinal, tt.want)
			assertValue(t, name, tt.name)
		})
	}
}

// TestTownerNameForTalkID pins the eight top-level pages that carry a towner
// name and the sub-pages that do not (stores.cpp:78-92).
func TestTownerNameForTalkID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		id   stores.TalkID
		want string
		ok   bool
	}{
		{stores.TalkIDSmith, "griswold", true},
		{stores.TalkIDWitch, "adria", true},
		{stores.TalkIDBoy, "wirt", true},
		{stores.TalkIDHealer, "pepin", true},
		{stores.TalkIDStoryteller, "cain", true},
		{stores.TalkIDTavern, "ogden", true},
		{stores.TalkIDDrunk, "farnham", true},
		{stores.TalkIDBarmaid, "gillian", true},
		{stores.TalkIDSmithBuy, "", false},
		{stores.TalkIDSmithSell, "", false},
		{stores.TalkIDWitchSell, "", false},
		{stores.TalkIDHealerBuy, "", false},
		{stores.TalkIDStorytellerIdentify, "", false},
		{stores.TalkIDGossip, "", false},
		{stores.TalkIDConfirm, "", false},
		{stores.TalkIDNone, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.id.String(), func(t *testing.T) {
			t.Parallel()

			// when the page's towner name is resolved
			got, ok := stores.TownerNameForTalkID(tt.id)

			// then only the eight top-level pages carry one
			assertValue(t, got, tt.want)
			assertValue(t, ok, tt.ok)
		})
	}
}
