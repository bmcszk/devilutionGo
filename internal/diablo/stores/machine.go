package stores

import (
	"errors"
	"fmt"

	"github.com/bmcszk/devilutionGo/internal/diablo/entities"
)

// Session is the store dialog state: ActiveStore, the page it was opened from,
// the panel it draws and the offered item list (stores.cpp:40-62).
type Session struct {
	// Active is ActiveStore, or TalkIDNone when no store is open.
	Active TalkID
	// OldActive is OldActiveStore: the page the current dialog returns to.
	OldActive TalkID
	// Dialog is the text panel.
	Dialog Dialog
	// ItemCount is CurrentItemIndex, the number of entries PlayerItemIndexes
	// holds after the current page built its list.
	ItemCount int
	// OldTextLine and OldScrollPos remember the list selection while a
	// confirmation or error page is shown (stores.cpp:55-60).
	OldTextLine  int
	OldScrollPos int
	// Holds is PlayerItems: the items the current list page offers.
	Holds []Hold
	// TempItem is the item a confirmation page asks about.
	TempItem entities.Item
	// Towner is TownerNames[TownerId], the name the talk page greets with.
	Towner string
}

// GameView is the world state the pages read while they build.
type GameView struct {
	// Env is the game-mode branches.
	Env Env
	// Menus carries the store-UI mode and the talk page inputs.
	Menus MenuInputs
	// Player is the acting player; its body, inventory and belt feed the
	// sell, repair, recharge and identify lists.
	Player *entities.Player
	// Gold is the player's carried gold, StashGold the shared stash.
	Gold      int
	StashGold int
	// SmithItems, PremiumItems, WitchItems and HealerItems are the vendor
	// stock lists the pages print (SmithItems and friends, stores.cpp:48-52).
	SmithItems   []entities.Item
	PremiumItems []entities.Item
	WitchItems   []entities.Item
	HealerItems  []entities.Item
	// BoyItem is Wirt's single offer.
	BoyItem entities.Item
	// TempItem is the item under negotiation.
	TempItem entities.Item
	// StaffCost resolves the spell staff cost RechargePrice adds
	// (GetSpellData(spell).staffCost()).
	StaffCost func(spell int32) int
	// DetailLines returns PrintStoreItem's stat block for a list entry; the
	// item-text layer owns the affix and durability formatting.
	DetailLines func(entities.Item) []string
	// CanFitItem is StoreAutoPlace(item, persistItem=false): whether the item
	// fits the player's inventory (stores.cpp:2173-2181).
	CanFitItem func(entities.Item) bool
	// GoldFits is StoreGoldFit for the item being sold (stores.cpp:1414-1426).
	GoldFits func(cost int, item entities.Item) bool
}

// ErrUnknownStore is returned for a TalkID outside the 24 dialog states.
var ErrUnknownStore = errors.New("unknown store")

// Start ports StartStore (stores.cpp:2287-2430): clear the panel, rebuild the
// page for id, then select its first option and make it active.
func (s *Session) Start(id TalkID, in GameView) error {
	s.Dialog.reset()

	if int(id) >= TalkIDCount {
		return fmt.Errorf("start store %d: %w", id, ErrUnknownStore)
	}

	done, err := s.buildPage(id, in)
	if err != nil {
		return err
	}

	if done {
		return nil
	}

	s.Dialog.CurrentTextLine = s.Dialog.FirstSelectable()
	s.Active = id

	return nil
}

// buildPage is the StartStore switch (stores.cpp:2305-2400). The re-entry
// cases — an empty vendor list, an empty premium list — restart the parent
// page and land the cursor on the entry that sent them there; they report done
// because StartStore returns before its ActiveStore tail there.
func (s *Session) buildPage(id TalkID, in GameView) (bool, error) {
	switch id {
	case TalkIDSmith:
		s.startSmith(in.Menus)
	case TalkIDSmithBuy:
		if len(in.SmithItems) == 0 {
			return true, s.startFailing(TalkIDSmith, 12, in)
		}

		s.startVendorList(in.SmithItems, in)
	case TalkIDSmithSell:
		s.startSellList(SmithSellHolds(in.Player, in.Env), "You have nothing I want.", in)
	case TalkIDSmithRepair:
		s.startRepairList(in)
	case TalkIDWitch:
		s.startWitch(in.Menus)
	case TalkIDWitchBuy:
		if s.ItemCount > 0 {
			s.startVendorList(in.WitchItems, in)
		}
	case TalkIDWitchSell:
		s.startSellList(WitchSellHolds(in.Player, in.Env), "You have nothing I want.", in)
	case TalkIDWitchRecharge:
		s.startRechargeList(in)
	case TalkIDNoMoney:
		return false, s.startNoMoney(in)
	case TalkIDNoRoom:
		return false, s.startNoRoom(in)
	case TalkIDConfirm:
		return false, s.startConfirm(in)
	case TalkIDBoy:
		s.startBoy(in)
	case TalkIDBoyBuy:
		s.startBoyBuy(in)
	case TalkIDHealer:
		s.startHealer()
	case TalkIDStoryteller:
		s.startStoryteller()
	case TalkIDHealerBuy:
		if s.ItemCount > 0 {
			s.startVendorList(in.HealerItems, in)
		}
	case TalkIDStorytellerIdentify:
		s.startIdentifyList(in)
	case TalkIDSmithPremiumBuy:
		if len(in.PremiumItems) == 0 {
			return true, s.startFailing(TalkIDSmith, 14, in)
		}

		s.startPremiumList(in.PremiumItems, in)
	case TalkIDGossip:
		s.startTalk(in.Menus)
	case TalkIDStorytellerIdentifyShow:
		return false, s.startIdentifyShow(in.TempItem, in)
	case TalkIDTavern:
		s.startTavern()
	case TalkIDDrunk:
		s.startDrunk()
	case TalkIDBarmaid:
		s.startBarmaid()
	case TalkIDNone:
	}

	return false, nil
}

// startFailing rebuilds the parent page and puts the cursor on want, which is
// how StoreNoMoney, StoreNoRoom, StoreConfirm and StoreESC re-enter a page
// (stores.cpp:911-929, 1192, 2497-2553).
func (s *Session) startFailing(parent TalkID, want int, in GameView) error {
	if parent == TalkIDNone {
		s.Active = TalkIDNone

		return nil
	}

	if err := s.Start(parent, in); err != nil {
		return err
	}

	if want >= 0 {
		s.Dialog.CurrentTextLine = want
	}

	return nil
}

// startVendorList builds a basic-goods vendor page: header, divider, the
// visible rows and the back button (stores.cpp:464-483, 734-755, 1050-1071).
func (s *Session) startVendorList(items []entities.Item, in GameView) {
	s.startStockList(items, in, "I have these items for sale:")
}

// startPremiumList builds Griswold's premium page (stores.cpp:496-523).
func (s *Session) startPremiumList(items []entities.Item, in GameView) {
	s.startStockList(items, in, "I have these premium items for sale:")
}

// startStockList is the shared vendor list layout.
func (s *Session) startStockList(items []entities.Item, in GameView, header string) {
	d := &s.Dialog
	d.IsTextFullSize = true
	d.HasScrollbar = true
	d.ScrollPos = 0
	d.RenderGold = true

	d.AddSText(20, 1, header, ColorWhitegold, AlignDefault, false)
	d.AddSLine(3)

	s.scrollVendorStore(items, d.ScrollPos, in)
	d.AddItemListBackButton()
	s.ItemCount = len(items)
	d.NumTextLines = max(len(items)-4, 0)
}

// scrollVendorStore is ScrollVendorStore (stores.cpp:410-430): four-line rows
// from line 5 to 19, name plus price plus the stat block.
func (s *Session) scrollVendorStore(items []entities.Item, idx int, in GameView) {
	d := &s.Dialog
	d.ClearSText(5, 21)

	for l := 5; l < 20 && idx < len(items); l += 4 {
		item := items[idx]
		color := TextColor(item)
		d.AddSText(20, l, itemName(item), color, AlignDefault, true)
		if item.Identified {
			d.AddSTextVal(l, item.Ivalue)
		} else {
			d.AddSTextVal(l, item.Value)
		}

		if in.DetailLines != nil {
			detail := in.DetailLines(item)
			for i, line := range detail[:min(len(detail), 2)] {
				d.AddSText(40, l+1+i, line, color, AlignDefault, false)
			}
		}

		idx++
	}
}

// startSellList builds a sell page from the offered holds (stores.cpp:555-606,
// 780-836).
func (s *Session) startSellList(holds []Hold, emptyText string, in GameView) {
	s.Holds = holds
	s.ItemCount = len(holds)

	if len(holds) == 0 {
		d := &s.Dialog
		d.HasScrollbar = false
		d.RenderGold = true
		d.AddSText(20, 1, emptyText, ColorWhitegold, AlignDefault, false)
		d.AddSLine(3)
		d.AddSelectableItemListBackButton()

		return
	}

	s.startHoldList()
}

// startRepairList builds the repair page (stores.cpp:629-688).
func (s *Session) startRepairList(in GameView) {
	s.Holds = RepairHolds(in.Player)
	s.ItemCount = len(s.Holds)

	if len(s.Holds) == 0 {
		d := &s.Dialog
		d.HasScrollbar = false
		d.RenderGold = true
		d.AddSText(20, 1, "You have nothing to repair.", ColorWhitegold, AlignDefault, false)
		d.AddSLine(3)
		d.AddSelectableItemListBackButton()

		return
	}

	s.startHoldList()
	s.Dialog.NumTextLines = in.Player.NumInv
}

// startRechargeList builds the recharge page (stores.cpp:861-907).
func (s *Session) startRechargeList(in GameView) {
	s.Holds = RechargeHolds(in.Player, in.StaffCost)
	s.ItemCount = len(s.Holds)

	if len(s.Holds) == 0 {
		d := &s.Dialog
		d.HasScrollbar = false
		d.RenderGold = true
		d.AddSText(20, 1, "You have nothing to recharge.", ColorWhitegold, AlignDefault, false)
		d.AddSLine(3)
		d.AddSelectableItemListBackButton()

		return
	}

	s.startHoldList()
	s.Dialog.NumTextLines = in.Player.NumInv
}

// startIdentifyList builds Cain's identify page (stores.cpp:1104-1160).
func (s *Session) startIdentifyList(in GameView) {
	s.Holds = IdentifyHolds(in.Player)
	s.ItemCount = len(s.Holds)
	s.Dialog.IsTextFullSize = true

	if len(s.Holds) == 0 {
		d := &s.Dialog
		d.HasScrollbar = false
		d.RenderGold = true
		d.AddSText(20, 1, "You have nothing to identify.", ColorWhitegold, AlignDefault, false)
		d.AddSLine(3)
		d.AddSelectableItemListBackButton()

		return
	}

	s.startHoldList()
}

// startHoldList is the shared header of the player-item list pages: scrollable
// full-size panel with the item rows and the back button (stores.cpp:602-606,
// 829-834, 683-687, 1040-1044).
func (s *Session) startHoldList() {
	d := &s.Dialog
	d.IsTextFullSize = true
	d.HasScrollbar = true
	d.ScrollPos = 0
	d.RenderGold = true
	d.AddSText(20, 1, "Which item?", ColorWhitegold, AlignDefault, false)
	d.AddSLine(3)

	d.ClearSText(5, 21)
	for l, i := 5, 0; l < 20 && i < len(s.Holds); l, i = l+4, i+1 {
		item := s.Holds[i].Item
		color := TextColor(item)
		d.AddSText(20, l, itemName(item), color, AlignDefault, true)
		d.AddSTextVal(l, item.Ivalue)
	}

	d.AddItemListBackButton()
	d.NumTextLines = max(len(s.Holds)-4, 0)
}

// startBoyBuy builds Wirt's one-item page (stores.cpp:990-1016).
func (s *Session) startBoyBuy(in GameView) {
	item := in.BoyItem

	d := &s.Dialog
	d.IsTextFullSize = true
	d.HasScrollbar = true
	d.RenderGold = true
	d.AddSText(20, 1, "Would you like to buy this item?", ColorWhitegold, AlignDefault, false)
	d.AddSLine(3)
	d.AddSText(20, 5, itemName(item), TextColor(item), AlignDefault, true)
	d.AddSTextVal(5, item.Ivalue)
	d.AddItemListBackButton()

	s.ItemCount = 1
}

// TextColor ports Item::getTextColorWithStatCheck (items.h:435-451): red when
// the player cannot use it, blue for magic, whitegold for unique.
func TextColor(item entities.Item) Color {
	if !item.StatFlag {
		return ColorRed
	}

	switch item.Magical {
	case 1:
		return ColorBlue
	case 2:
		return ColorWhitegold
	default:
		return ColorWhite
	}
}

// vendorItemUnused is kept out: the vendor lists are the caller's stock slices
// (GameView.SmithItems and friends), so no per-kind lookup is needed.

// pick returns items[idx], or the zero item when the list is shorter.
func pick(items []entities.Item, idx int) entities.Item {
	if idx < 0 || idx >= len(items) {
		return entities.Item{}
	}

	return items[idx]
}
