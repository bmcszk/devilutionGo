package stores

import (
	"github.com/bmcszk/devilutionGo/internal/diablo/entities"
)

// boyLookFee is BoyEnter's 50 gold "just to take a look" charge
// (stores.cpp:1696-1706).
const boyLookFee = 50

// scrollPitch is PreviousScrollPos, the first item row line: rows are four
// lines apart, so (line-5)/4 indexes the offer list (stores.cpp:409).
const scrollPitch = 5

// entryIndex is the ScrollPos + (CurrentTextLine - PreviousScrollPos) / 4
// lookup every list Enter handler uses (stores.cpp:1355, 1396, 1485, 1576).
func (s *Session) entryIndex() int {
	return s.Dialog.ScrollPos + (s.Dialog.CurrentTextLine-scrollPitch)/4
}

// enterVendorList ports SmithBuyEnter, SmithPremiumBuyEnter, WitchBuyEnter and
// HealerBuyEnter (stores.cpp:1349-1420, 1564-1590, 1902-1926): the back line
// returns to the menu, otherwise affordability and inventory room are checked
// before the confirmation page opens.
func (s *Session) enterVendorList(parent TalkID, parentLine int, action Action, in GameView) (Outcome, error) {
	if s.Dialog.CurrentTextLine == BackButtonLine {
		return Outcome{}, s.escTo(parent, parentLine, in)
	}

	list := s.vendorList(in)
	idx := s.entryIndex()
	if idx < 0 || idx >= len(list) {
		return Outcome{}, nil
	}

	item := list[idx]

	s.OldTextLine = s.Dialog.CurrentTextLine
	s.OldScrollPos = s.Dialog.ScrollPos
	s.OldActive = s.Active

	if !CanAfford(TotalGold(in.Gold, in.StashGold), item.Ivalue) {
		return Outcome{}, s.Start(TalkIDNoMoney, in)
	}

	if in.CanFitItem != nil && !in.CanFitItem(item) {
		return Outcome{}, s.Start(TalkIDNoRoom, in)
	}

	s.TempItem = item

	return Outcome{Action: action, Item: item, Price: item.Ivalue}, s.Start(TalkIDConfirm, in)
}

// vendorList returns the stock slice the active vendor page is showing.
func (s *Session) vendorList(in GameView) []entities.Item {
	switch s.Active {
	case TalkIDSmithBuy:
		return in.SmithItems
	case TalkIDSmithPremiumBuy:
		return in.PremiumItems
	case TalkIDWitchBuy:
		return in.WitchItems
	case TalkIDHealerBuy:
		return in.HealerItems
	case TalkIDSmith, TalkIDSmithSell, TalkIDSmithRepair, TalkIDWitch,
		TalkIDWitchSell, TalkIDWitchRecharge, TalkIDHealer, TalkIDStoryteller,
		TalkIDStorytellerIdentify, TalkIDBoy, TalkIDBoyBuy, TalkIDNoMoney,
		TalkIDNoRoom, TalkIDConfirm, TalkIDGossip, TalkIDStorytellerIdentifyShow,
		TalkIDTavern, TalkIDDrunk, TalkIDBarmaid, TalkIDNone:
	}

	return nil
}

// enterSellList ports SmithSellEnter and WitchSellEnter (stores.cpp:1455-1478,
// 1592-1626): selling only has to fit the gold the sale adds.
func (s *Session) enterSellList(in GameView) (Outcome, error) {
	parent, parentLine := TalkIDSmith, 16
	if s.Active == TalkIDWitchSell {
		parent, parentLine = TalkIDWitch, 16
	}

	if s.Dialog.CurrentTextLine == BackButtonLine {
		return Outcome{}, s.escTo(parent, parentLine, in)
	}

	idx := s.entryIndex()
	if idx < 0 || idx >= len(s.Holds) {
		return Outcome{}, nil
	}

	s.OldTextLine = s.Dialog.CurrentTextLine
	s.OldActive = s.Active
	s.OldScrollPos = s.Dialog.ScrollPos

	hold := s.Holds[idx]
	if in.GoldFits != nil && !in.GoldFits(hold.Item.Ivalue, hold.Item) {
		return Outcome{}, s.Start(TalkIDNoRoom, in)
	}

	s.TempItem = hold.Item

	return Outcome{Action: ActionSell, Item: hold.Item, Price: hold.Item.Ivalue, Index: hold.Index}, s.Start(TalkIDConfirm, in)
}

// enterHoldList ports SmithRepairEnter, WitchRechargeEnter and
// StorytellerIdentifyEnter (stores.cpp:1509-1530, 1628-1660, 1947-1968).
func (s *Session) enterHoldList(parent TalkID, parentLine int, action Action, in GameView) (Outcome, error) {
	if s.Dialog.CurrentTextLine == BackButtonLine {
		return Outcome{}, s.escTo(parent, parentLine, in)
	}

	idx := s.entryIndex()
	if idx < 0 || idx >= len(s.Holds) {
		return Outcome{}, nil
	}

	s.OldActive = s.Active
	s.OldTextLine = s.Dialog.CurrentTextLine
	s.OldScrollPos = s.Dialog.ScrollPos

	hold := s.Holds[idx]
	if !CanAfford(TotalGold(in.Gold, in.StashGold), hold.Item.Ivalue) {
		return Outcome{}, s.Start(TalkIDNoMoney, in)
	}

	s.TempItem = hold.Item

	return Outcome{Action: action, Item: hold.Item, Price: hold.Item.Ivalue, Index: hold.Index}, s.Start(TalkIDConfirm, in)
}

// enterBoyBuy ports BoyBuyEnter (stores.cpp:1813-1830): the haggle price is
// checked before the deal page.
func (s *Session) enterBoyBuy(in GameView) (Outcome, error) {
	if s.Dialog.CurrentTextLine != 10 {
		s.Active = TalkIDNone

		return Outcome{}, nil
	}

	price := BoyPrice(in.BoyItem, in.Env)

	s.OldActive = TalkIDBoyBuy
	s.OldScrollPos = s.Dialog.ScrollPos
	s.OldTextLine = 10

	if !CanAfford(TotalGold(in.Gold, in.StashGold), price) {
		return Outcome{}, s.Start(TalkIDNoMoney, in)
	}

	s.TempItem = in.BoyItem
	s.TempItem.Ivalue = price

	return Outcome{Action: ActionBuyBoy, Item: in.BoyItem, Price: price}, s.Start(TalkIDConfirm, in)
}

// enterConfirm ports ConfirmEnter (stores.cpp:1783-1866): "Yes" executes the
// trade the asking page recorded, then the asking page is rebuilt.
func (s *Session) enterConfirm(in GameView) (Outcome, error) {
	out := Outcome{}

	if s.Dialog.CurrentTextLine == 18 {
		out = Outcome{Action: confirmAction(s.OldActive), Item: s.TempItem, Price: s.TempItem.Ivalue, Index: holdIndexOf(s.Holds, s.OldTextLine, s.OldScrollPos)}
	}

	if out.Action == ActionIdentify {
		return out, s.Start(TalkIDStorytellerIdentifyShow, in)
	}

	if err := s.Start(s.OldActive, in); err != nil {
		return out, err
	}

	s.restoreSelection(in)

	return out, nil
}

// restoreSelection is ConfirmEnter's tail: skip the back line, restore the old
// line clamped to the new scroll range, then walk up to a selectable line
// (stores.cpp:1852-1866).
func (s *Session) restoreSelection(in GameView) {
	d := &s.Dialog
	if d.CurrentTextLine == BackButtonLine {
		return
	}

	d.CurrentTextLine = s.OldTextLine
	d.ScrollPos = min(s.OldScrollPos, d.NumTextLines)

	for d.CurrentTextLine != -1 && !d.Lines[d.CurrentTextLine].IsSelectable() {
		d.CurrentTextLine--
	}

	_ = in
}

// enterRestore ports the NoMoney and NoRoom Enter branch: both pages just
// return to the page that failed (stores.cpp:2749-2754).
func (s *Session) enterRestore(in GameView) (Outcome, error) {
	if err := s.Start(s.OldActive, in); err != nil {
		return Outcome{}, err
	}

	s.Dialog.CurrentTextLine = s.OldTextLine
	s.Dialog.ScrollPos = s.OldScrollPos

	return Outcome{}, nil
}

// confirmAction maps the asking page to the trade ConfirmEnter runs
// (stores.cpp:1786-1818).
func confirmAction(old TalkID) Action {
	switch old {
	case TalkIDSmithBuy:
		return ActionBuyBasic
	case TalkIDSmithSell, TalkIDWitchSell:
		return ActionSell
	case TalkIDSmithRepair:
		return ActionRepair
	case TalkIDWitchBuy:
		return ActionBuyWitch
	case TalkIDWitchRecharge:
		return ActionRecharge
	case TalkIDBoyBuy:
		return ActionBuyBoy
	case TalkIDHealerBuy:
		return ActionBuyHealer
	case TalkIDStorytellerIdentify:
		return ActionIdentify
	case TalkIDSmithPremiumBuy:
		return ActionBuyPremium
	case TalkIDSmith, TalkIDWitch, TalkIDHealer, TalkIDStoryteller, TalkIDBoy,
		TalkIDNoMoney, TalkIDNoRoom, TalkIDConfirm, TalkIDGossip,
		TalkIDStorytellerIdentifyShow, TalkIDTavern, TalkIDDrunk, TalkIDBarmaid,
		TalkIDNone:
	}

	return ActionNone
}

// holdIndexOf resolves the offered item the confirmation page was opened for.
func holdIndexOf(holds []Hold, line, scrollPos int) int8 {
	idx := scrollPos + (line-scrollPitch)/4
	if idx < 0 || idx >= len(holds) {
		return 0
	}

	return holds[idx].Index
}

// escTo rebuilds page and puts the cursor on line (StoreESC's pattern,
// stores.cpp:2497-2553).
func (s *Session) escTo(page TalkID, line int, in GameView) error {
	if page == TalkIDNone {
		s.Active = TalkIDNone

		return nil
	}

	if err := s.Start(page, in); err != nil {
		return err
	}

	if line >= 0 {
		s.Dialog.CurrentTextLine = line
	}

	return nil
}

// townerDisplayName is TownerNames[] (stores.cpp:194-204): the name the talk
// page greets with.
func townerDisplayName(page TalkID) string {
	switch page {
	case TalkIDSmith:
		return "Griswold"
	case TalkIDHealer:
		return "Pepin"
	case TalkIDTavern:
		return "Ogden"
	case TalkIDStoryteller:
		return "Cain"
	case TalkIDDrunk:
		return "Farnham"
	case TalkIDWitch:
		return "Adria"
	case TalkIDBarmaid:
		return "Gillian"
	case TalkIDBoy:
		return "Wirt"
	case TalkIDNone, TalkIDSmithBuy, TalkIDSmithSell, TalkIDSmithRepair,
		TalkIDWitchBuy, TalkIDWitchSell, TalkIDWitchRecharge, TalkIDNoMoney,
		TalkIDNoRoom, TalkIDConfirm, TalkIDBoyBuy, TalkIDHealerBuy,
		TalkIDStorytellerIdentify, TalkIDSmithPremiumBuy, TalkIDGossip,
		TalkIDStorytellerIdentifyShow:
	}

	return ""
}
