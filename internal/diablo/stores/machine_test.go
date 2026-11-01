package stores_test

import (
	"errors"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/entities"
	"github.com/bmcszk/devilutionGo/internal/diablo/stores"
)

// view builds a GameView over a fresh player: 3000 gold, no stash, list store
// UI, and the vendor stock the test passes in.
func view(fx *playerFixture) stores.GameView {
	return stores.GameView{
		Player:     fx.player,
		Gold:       3000,
		StashGold:  0,
		StaffCost:  func(spell int32) int { return 600 },
		CanFitItem: func(entities.Item) bool { return true },
		GoldFits:   func(int, entities.Item) bool { return true },
	}
}

// TestStartPages pins each page's panel flags and the option the cursor lands
// on (stores.cpp:2404-2430).
func TestStartPages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		id            stores.TalkID
		fullSize      bool
		hasScrollbar  bool
		renderGold    bool
		wantFirstLine int
	}{
		{"Smith", stores.TalkIDSmith, false, false, false, 10},
		{"Witch", stores.TalkIDWitch, false, false, false, 12},
		{"Healer", stores.TalkIDHealer, false, false, false, 12},
		{"Storyteller", stores.TalkIDStoryteller, false, false, false, 12},
		{"Boy", stores.TalkIDBoy, false, false, false, 8},
		{"Tavern", stores.TalkIDTavern, false, false, false, 12},
		{"Barmaid", stores.TalkIDBarmaid, false, false, false, 12},
		{"Drunk", stores.TalkIDDrunk, false, false, false, 12},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			session := &stores.Session{}
			in := view(fixture())
			in.BoyItem = sword()

			if err := session.Start(tt.id, in); err != nil {
				t.Fatalf("Start(%s): %v", tt.id, err)
			}

			if session.Active != tt.id {
				t.Fatalf("Active = %s, want %s", session.Active, tt.id)
			}

			if session.Dialog.IsTextFullSize != tt.fullSize || session.Dialog.HasScrollbar != tt.hasScrollbar {
				t.Fatalf("panel flags = full:%v scroll:%v, want full:%v scroll:%v",
					session.Dialog.IsTextFullSize, session.Dialog.HasScrollbar, tt.fullSize, tt.hasScrollbar)
			}

			if session.Dialog.RenderGold != tt.renderGold {
				t.Fatalf("RenderGold = %v, want %v", session.Dialog.RenderGold, tt.renderGold)
			}

			if session.Dialog.CurrentTextLine != tt.wantFirstLine {
				t.Fatalf("CurrentTextLine = %d, want %d", session.Dialog.CurrentTextLine, tt.wantFirstLine)
			}
		})
	}
}

// TestStartSmithMenuLines pins Griswold's menu text and lines
// (stores.cpp:432-453).
func TestStartSmithMenuLines(t *testing.T) {
	t.Parallel()

	session := &stores.Session{}
	if err := session.Start(stores.TalkIDSmith, view(fixture())); err != nil {
		t.Fatalf("Start(Smith): %v", err)
	}

	labels := map[int]string{
		1: "Welcome to the",
		3: "Blacksmith's shop",
		7: "Would you like to:",
	}

	for line, text := range labels {
		if got := session.Dialog.Lines[line].Text; got != text {
			t.Fatalf("line %d = %q, want %q", line, got, text)
		}

		if session.Dialog.Lines[line].IsSelectable() {
			t.Fatalf("line %d is a label, want not selectable", line)
		}
	}

	options := map[int]string{
		10: "Talk to Griswold",
		12: "Buy basic items",
		14: "Buy premium items",
		16: "Sell items",
		18: "Repair items",
		20: "Leave the shop",
	}

	for line, text := range options {
		if got := session.Dialog.Lines[line].Text; got != text {
			t.Fatalf("line %d = %q, want %q", line, got, text)
		}

		if !session.Dialog.Lines[line].IsSelectable() {
			t.Fatalf("line %d is not selectable", line)
		}
	}

	if !session.Dialog.Lines[5].IsDivider() {
		t.Fatalf("line 5 is not the menu divider")
	}

	if session.ItemCount != 20 {
		t.Fatalf("ItemCount = %d, want 20", session.ItemCount)
	}
}

// TestStartVisualGridMenus pins the StoreUi::VisualGrid variants
// (stores.cpp:439-444, 695-700).
func TestStartVisualGridMenus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		id    stores.TalkID
		lines map[int]string
	}{
		{
			name: "smith",
			id:   stores.TalkIDSmith,
			lines: map[int]string{
				12: "Trade / Repair",
				14: "Leave the shop",
			},
		},
		{
			name: "witch",
			id:   stores.TalkIDWitch,
			lines: map[int]string{
				14: "Buy / Sell",
				16: "Recharge staves",
				18: "Leave the shack",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			session := &stores.Session{}
			in := view(fixture())
			in.Menus.StoreUI = stores.StoreUIVisualGrid

			if err := session.Start(tt.id, in); err != nil {
				t.Fatalf("Start(%s): %v", tt.id, err)
			}

			for line, text := range tt.lines {
				if got := session.Dialog.Lines[line].Text; got != text {
					t.Fatalf("line %d = %q, want %q", line, got, text)
				}
			}

			if session.Dialog.Lines[20].Text != "" {
				t.Fatalf("line 20 = %q, want empty in the visual-grid menu", session.Dialog.Lines[20].Text)
			}
		})
	}
}

// TestStartReentry pins StartStore's re-entry semantics: an empty vendor list
// restarts the menu and lands the cursor on the option that opened it
// (stores.cpp:2305-2320, 496-505).
func TestStartReentry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		id       stores.TalkID
		mutate   func(*stores.GameView)
		wantPage stores.TalkID
		wantLine int
	}{
		{
			name:     "empty smith list falls back to the buy option",
			id:       stores.TalkIDSmithBuy,
			mutate:   func(*stores.GameView) {},
			wantPage: stores.TalkIDSmith,
			wantLine: 12,
		},
		{
			name:     "empty premium shelf falls back to the premium option",
			id:       stores.TalkIDSmithPremiumBuy,
			mutate:   func(*stores.GameView) {},
			wantPage: stores.TalkIDSmith,
			wantLine: 14,
		},
		{
			name: "stocked smith list stays on the list page",
			id:   stores.TalkIDSmithBuy,
			mutate: func(in *stores.GameView) {
				in.SmithItems = []entities.Item{sword()}
			},
			wantPage: stores.TalkIDSmithBuy,
			wantLine: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			session := &stores.Session{}
			in := view(fixture())
			tt.mutate(&in)

			if err := session.Start(tt.id, in); err != nil {
				t.Fatalf("Start(%s): %v", tt.id, err)
			}

			if session.Active != tt.wantPage {
				t.Fatalf("Active = %s, want %s", session.Active, tt.wantPage)
			}

			if session.Dialog.CurrentTextLine != tt.wantLine {
				t.Fatalf("CurrentTextLine = %d, want %d", session.Dialog.CurrentTextLine, tt.wantLine)
			}
		})
	}
}

// TestStartUnknownPage pins the guard on an out-of-range page.
func TestStartUnknownPage(t *testing.T) {
	t.Parallel()

	session := &stores.Session{}

	err := session.Start(stores.TalkID(200), view(fixture()))
	if !errors.Is(err, stores.ErrUnknownStore) {
		t.Fatalf("Start(200) error = %v, want ErrUnknownStore", err)
	}
}

// TestEscTable pins StoreESC's per-page destination (stores.cpp:2481-2553).
func TestEscTable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		id       stores.TalkID
		oldLine  int
		wantPage stores.TalkID
		wantLine int
	}{
		{"Smith closes", stores.TalkIDSmith, 10, stores.TalkIDNone, 0},
		{"Witch closes", stores.TalkIDWitch, 12, stores.TalkIDNone, 0},
		{"Boy closes", stores.TalkIDBoy, 8, stores.TalkIDNone, 0},
		{"BoyBuy closes", stores.TalkIDBoyBuy, 10, stores.TalkIDNone, 0},
		{"Healer closes", stores.TalkIDHealer, 12, stores.TalkIDNone, 0},
		{"Storyteller closes", stores.TalkIDStoryteller, 12, stores.TalkIDNone, 0},
		{"Tavern closes", stores.TalkIDTavern, 12, stores.TalkIDNone, 0},
		{"Drunk closes", stores.TalkIDDrunk, 12, stores.TalkIDNone, 0},
		{"Barmaid closes", stores.TalkIDBarmaid, 12, stores.TalkIDNone, 0},
		{"SmithBuy returns to buy", stores.TalkIDSmithBuy, 5, stores.TalkIDSmith, 12},
		{"SmithSell returns to sell", stores.TalkIDSmithSell, 5, stores.TalkIDSmith, 16},
		{"SmithRepair returns to repair", stores.TalkIDSmithRepair, 5, stores.TalkIDSmith, 18},
		{"SmithPremiumBuy returns to premium", stores.TalkIDSmithPremiumBuy, 5, stores.TalkIDSmith, 14},
		{"WitchBuy returns to buy", stores.TalkIDWitchBuy, 5, stores.TalkIDWitch, 14},
		{"WitchSell returns to sell", stores.TalkIDWitchSell, 5, stores.TalkIDWitch, 16},
		{"WitchRecharge returns to recharge", stores.TalkIDWitchRecharge, 5, stores.TalkIDWitch, 18},
		{"HealerBuy returns to buy", stores.TalkIDHealerBuy, 5, stores.TalkIDHealer, 14},
		{"StorytellerIdentify returns to identify", stores.TalkIDStorytellerIdentify, 5, stores.TalkIDStoryteller, 14},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			session := &stores.Session{}
			in := view(fixture())
			in.SmithItems = []entities.Item{sword()}
			in.WitchItems = []entities.Item{sword()}
			in.HealerItems = []entities.Item{sword()}
			in.PremiumItems = []entities.Item{sword()}

			if err := session.Start(tt.id, in); err != nil {
				t.Fatalf("Start(%s): %v", tt.id, err)
			}

			session.OldTextLine = tt.oldLine

			if err := session.Esc(in); err != nil {
				t.Fatalf("Esc: %v", err)
			}

			if session.Active != tt.wantPage {
				t.Fatalf("Active = %s, want %s", session.Active, tt.wantPage)
			}

			if tt.wantPage != stores.TalkIDNone && session.Dialog.CurrentTextLine != tt.wantLine {
				t.Fatalf("CurrentTextLine = %d, want %d", session.Dialog.CurrentTextLine, tt.wantLine)
			}
		})
	}
}
