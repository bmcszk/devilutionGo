package stores_test

import (
	"errors"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/entities"
	"github.com/bmcszk/devilutionGo/internal/diablo/stores"
)

// smithMenuLabels is Griswold's static greeting text (stores.cpp:432-439).
var smithMenuLabels = map[int]string{
	1: "Welcome to the",
	3: "Blacksmith's shop",
	7: "Would you like to:",
}

// smithMenuOptions is Griswold's selectable menu text (stores.cpp:440-453).
var smithMenuOptions = map[int]string{
	10: "Talk to Griswold",
	12: "Buy basic items",
	14: "Buy premium items",
	16: "Sell items",
	18: "Repair items",
	20: "Leave the shop",
}

// startPageCase is one row of the page-flag table (stores.cpp:2404-2430).
type startPageCase struct {
	name          string
	id            stores.TalkID
	fullSize      bool
	hasScrollbar  bool
	renderGold    bool
	wantFirstLine int
}

// visualMenuCase is one row of the visual-grid menu table
// (stores.cpp:439-444, 695-700).
type visualMenuCase struct {
	name  string
	id    stores.TalkID
	lines map[int]string
}

// reentryCase is one row of the re-entry table (stores.cpp:2305-2320, 496-505).
type reentryCase struct {
	name     string
	id       stores.TalkID
	mutate   func(*stores.GameView)
	wantPage stores.TalkID
	wantLine int
}

// escCase is one row of the escape table (stores.cpp:2481-2553).
type escCase struct {
	name     string
	id       stores.TalkID
	oldLine  int
	wantPage stores.TalkID
	wantLine int
}

// TestStartPages pins each page's panel flags and the option the cursor lands
// on (stores.cpp:2404-2430).
func TestStartPages(t *testing.T) {
	t.Parallel()

	tests := []startPageCase{
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

			// when the page opens
			session := startSession(t, tt.id, view(fixture()))

			// then its flags and its first selectable line match upstream
			assertStartPage(t, session, tt)
		})
	}
}

// TestStartSmithMenuLines pins Griswold's menu text and lines
// (stores.cpp:432-453).
func TestStartSmithMenuLines(t *testing.T) {
	t.Parallel()

	// given Griswold's menu page
	session := startSession(t, stores.TalkIDSmith, view(fixture()))

	// when the panel is drawn
	dialog := &session.Dialog

	// then the greeting is static text, the options are selectable, a divider
	// separates them and the option count spans the last option line
	assertMenuText(t, dialog, smithMenuLabels, false)
	assertMenuText(t, dialog, smithMenuOptions, true)
	assertValue(t, dialog.Lines[5].IsDivider(), true)
	assertValue(t, session.ItemCount, 20)
}

// TestStartVisualGridMenus pins the StoreUi::VisualGrid variants
// (stores.cpp:439-444, 695-700).
func TestStartVisualGridMenus(t *testing.T) {
	t.Parallel()

	tests := []visualMenuCase{
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

			// given the visual-grid store UI
			in := view(fixture())
			in.Menus.StoreUI = stores.StoreUIVisualGrid

			// when the page opens
			session := startSession(t, tt.id, in)

			// then its lines are the visual grid's and line 20 stays empty
			assertMenuText(t, &session.Dialog, tt.lines, true)
			assertValue(t, session.Dialog.Lines[20].Text, "")
		})
	}
}

// TestStartReentry pins StartStore's re-entry semantics: an empty vendor list
// restarts the menu and lands the cursor on the option that opened it
// (stores.cpp:2305-2320, 496-505).
func TestStartReentry(t *testing.T) {
	t.Parallel()

	tests := []reentryCase{
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

			// given a page whose stock the case sets
			in := view(fixture())
			tt.mutate(&in)

			// when the page opens
			session := startSession(t, tt.id, in)

			// then it lands on the case's page and line
			assertValue(t, session.Active, tt.wantPage)
			assertValue(t, session.Dialog.CurrentTextLine, tt.wantLine)
		})
	}
}

// TestStartUnknownPage pins the guard on an out-of-range page.
func TestStartUnknownPage(t *testing.T) {
	t.Parallel()

	// given an empty session
	session := &stores.Session{}

	// when a page outside the 24 dialog states opens
	err := session.Start(stores.TalkID(200), view(fixture()))

	// then it reports the unknown store
	assertValue(t, errors.Is(err, stores.ErrUnknownStore), true)
}

// TestEscTable pins StoreESC's per-page destination (stores.cpp:2481-2553).
func TestEscTable(t *testing.T) {
	t.Parallel()

	tests := []escCase{
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

			// given a stocked store opened on the case's page
			in := stockedView()
			session := startSession(t, tt.id, in)
			session.OldTextLine = tt.oldLine

			// when escape is pressed
			err := session.Esc(in)

			// then it walks back to the case's page and line
			assertValue(t, err, nil)
			assertValue(t, session.Active, tt.wantPage)
			assertEscLine(t, session, tt)
		})
	}
}

// view builds a GameView over a fresh player: 3000 gold, no stash, list store
// UI, and the vendor stock the test passes in.
func view(fx *playerFixture) stores.GameView {
	return stores.GameView{
		Player:     fx.player,
		Gold:       3000,
		StashGold:  0,
		StaffCost:  func(int32) int { return 600 },
		CanFitItem: func(entities.Item) bool { return true },
		GoldFits:   func(int, entities.Item) bool { return true },
	}
}

// stockedView is a view with one item on every vendor's shelf.
func stockedView() stores.GameView {
	in := view(fixture())
	in.SmithItems = []entities.Item{sword()}
	in.WitchItems = []entities.Item{sword()}
	in.HealerItems = []entities.Item{sword()}
	in.PremiumItems = []entities.Item{sword()}

	return in
}

// startSession opens id and fails the test when the store refuses it.
func startSession(t *testing.T, id stores.TalkID, in stores.GameView) *stores.Session {
	t.Helper()

	session := &stores.Session{}
	if err := session.Start(id, in); err != nil {
		t.Fatalf("Start(%s): %v", id, err)
	}

	return session
}

// assertStartPage checks one page's panel flags, cursor landing and identity.
func assertStartPage(t *testing.T, session *stores.Session, tc startPageCase) {
	t.Helper()

	assertValue(t, session.Active, tc.id)
	assertValue(t, session.Dialog.IsTextFullSize, tc.fullSize)
	assertValue(t, session.Dialog.HasScrollbar, tc.hasScrollbar)
	assertValue(t, session.Dialog.RenderGold, tc.renderGold)
	assertValue(t, session.Dialog.CurrentTextLine, tc.wantFirstLine)
}

// assertMenuText checks the text and selectability of the given lines.
func assertMenuText(t *testing.T, dialog *stores.Dialog, lines map[int]string, selectable bool) {
	t.Helper()

	for line, text := range lines {
		assertValue(t, dialog.Lines[line].Text, text)
		assertValue(t, dialog.Lines[line].IsSelectable(), selectable)
	}
}

// assertEscLine checks the cursor line of an escape that stayed in the store.
func assertEscLine(t *testing.T, session *stores.Session, tc escCase) {
	t.Helper()

	if tc.wantPage == stores.TalkIDNone {
		return
	}

	assertValue(t, session.Dialog.CurrentTextLine, tc.wantLine)
}
