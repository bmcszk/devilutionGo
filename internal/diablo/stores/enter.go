package stores

import (
	"github.com/bmcszk/devilutionGo/internal/diablo/entities"
	"github.com/bmcszk/devilutionGo/internal/diablo/inventory"
)

// Action is the trade ConfirmEnter executed, so the caller can apply the
// side effects the store does not own (moving the item, spending gold,
// recalculating the character).
type Action uint8

// The transaction outcomes ConfirmEnter can produce (stores.cpp:1795-1866).
const (
	// ActionNone means nothing was traded.
	ActionNone Action = iota
	// ActionBuyBasic is SmithBuyItem: pay Ivalue, take the smith list entry.
	ActionBuyBasic
	// ActionBuyPremium is SmithBuyPItem: pay Ivalue, take the premium entry.
	ActionBuyPremium
	// ActionSell is StoreSellItem: remove the item, add its sell value.
	ActionSell
	// ActionRepair is SmithRepairItem: pay the price, restore durability.
	ActionRepair
	// ActionBuyWitch is WitchBuyItem: pay Ivalue, take the witch list entry.
	ActionBuyWitch
	// ActionRecharge is WitchRechargeItem: pay the price, refill charges.
	ActionRecharge
	// ActionIdentify is StorytellerIdentifyItem: pay the fee, identify.
	ActionIdentify
	// ActionBuyBoy is BoyBuyItem: pay the price, take Wirt's item.
	ActionBuyBoy
	// ActionBuyHealer is HealerBuyItem: pay Ivalue, take the healer entry.
	ActionBuyHealer
	// ActionOpenVisualStore is the VisualGrid menu's OpenVisualStore call.
	ActionOpenVisualStore
	// ActionOpenStorage is the barmaid's stash open.
	ActionOpenStorage
	// ActionGossipText is TalkEnter selecting a gossip or quest line.
	ActionGossipText
)

// Outcome is what one Enter produced.
type Outcome struct {
	// Action is the executed trade, or ActionNone.
	Action Action
	// Item is the item traded, when the action involves one.
	Item entities.Item
	// Price is the gold the action moves.
	Price int
	// Index is the PlayerItemIndexes entry the action removes or edits.
	Index int8
}

// Enter ports StoreEnter (stores.cpp:2715-2790): route the selected line to its
// page handler, which either opens a sub-page, closes the store, or asks for
// confirmation.
func (s *Session) Enter(in GameView) (Outcome, error) {
	switch s.Active {
	case TalkIDSmith:
		return s.enterSmith(in)
	case TalkIDWitch:
		return s.enterWitch(in)
	case TalkIDHealer:
		return s.enterHealer(in)
	case TalkIDStoryteller:
		return s.enterStoryteller(in)
	case TalkIDBoy:
		return s.enterBoy(in)
	case TalkIDTavern:
		return s.enterTownerLine(TalkIDTavern, in)
	case TalkIDBarmaid:
		return s.enterBarmaid(in)
	case TalkIDDrunk:
		return s.enterTownerLine(TalkIDDrunk, in)
	case TalkIDGossip:
		return s.enterTalk(in)
	case TalkIDSmithBuy:
		return s.enterVendorList(TalkIDSmith, 12, ActionBuyBasic, in)
	case TalkIDSmithPremiumBuy:
		return s.enterVendorList(TalkIDSmith, 14, ActionBuyPremium, in)
	case TalkIDHealerBuy:
		return s.enterVendorList(TalkIDHealer, 14, ActionBuyHealer, in)
	case TalkIDWitchBuy:
		return s.enterVendorList(TalkIDWitch, 14, ActionBuyWitch, in)
	case TalkIDSmithSell, TalkIDWitchSell:
		return s.enterSellList(in)
	case TalkIDSmithRepair:
		return s.enterHoldList(TalkIDSmith, 18, ActionRepair, in)
	case TalkIDWitchRecharge:
		return s.enterHoldList(TalkIDWitch, s.witchRechargeBackLine(in), ActionRecharge, in)
	case TalkIDStorytellerIdentify:
		return s.enterHoldList(TalkIDStoryteller, 14, ActionIdentify, in)
	case TalkIDStorytellerIdentifyShow:
		return Outcome{}, s.Start(TalkIDStorytellerIdentify, in)
	case TalkIDBoyBuy:
		return s.enterBoyBuy(in)
	case TalkIDConfirm:
		return s.enterConfirm(in)
	case TalkIDNoMoney, TalkIDNoRoom:
		return s.enterRestore(in)
	case TalkIDNone:
	}

	return Outcome{}, nil
}

// witchRechargeBackLine is StoreESC's StoreUi-dependent return line
// (stores.cpp:2533-2536).
func (s *Session) witchRechargeBackLine(in GameView) int {
	if in.Menus.StoreUI == StoreUIVisualGrid {
		return 16
	}

	return 18
}

// enterMenuLine is the shared handler of a towner menu page: the greet line
// opens the talk page, a trailing line closes the store, and any other line is
// the sub-page the caller maps.
func (s *Session) enterMenuLine(greetLine, closeLine int, in GameView) (bool, Outcome, error) {
	switch s.Dialog.CurrentTextLine {
	case greetLine:
		s.OldTextLine = greetLine
		s.OldActive = s.Active
		s.Towner = townerDisplayName(s.Active)

		return true, Outcome{}, s.Start(TalkIDGossip, in)
	case closeLine:
		s.Active = TalkIDNone

		return true, Outcome{}, nil
	}

	return false, Outcome{}, nil
}

// enterSmith ports SmithEnter's line map (stores.cpp:1293-1333).
func (s *Session) enterSmith(in GameView) (Outcome, error) {
	if in.Menus.StoreUI == StoreUIVisualGrid {
		switch s.Dialog.CurrentTextLine {
		case 10:
			_, out, err := s.enterMenuLine(10, 14, in)

			return out, err
		case 12:
			s.Active = TalkIDNone

			return Outcome{Action: ActionOpenVisualStore}, nil
		case 14:
			s.Active = TalkIDNone
		}

		return Outcome{}, nil
	}

	if done, out, err := s.enterMenuLine(10, 20, in); done {
		return out, err
	}

	next := map[int]TalkID{12: TalkIDSmithBuy, 14: TalkIDSmithPremiumBuy, 16: TalkIDSmithSell, 18: TalkIDSmithRepair}

	if page, ok := next[s.Dialog.CurrentTextLine]; ok {
		return Outcome{}, s.Start(page, in)
	}

	return Outcome{}, nil
}

// enterWitch ports WitchEnter's line map (stores.cpp:1531-1580).
func (s *Session) enterWitch(in GameView) (Outcome, error) {
	if in.Menus.StoreUI == StoreUIVisualGrid {
		switch s.Dialog.CurrentTextLine {
		case 12:
			_, out, err := s.enterMenuLine(12, 18, in)

			return out, err
		case 14:
			s.Active = TalkIDNone

			return Outcome{Action: ActionOpenVisualStore}, nil
		case 16:
			return Outcome{}, s.Start(TalkIDWitchRecharge, in)
		case 18:
			s.Active = TalkIDNone
		}

		return Outcome{}, nil
	}

	if done, out, err := s.enterMenuLine(12, 20, in); done {
		return out, err
	}

	next := map[int]TalkID{14: TalkIDWitchBuy, 16: TalkIDWitchSell, 18: TalkIDWitchRecharge}

	if page, ok := next[s.Dialog.CurrentTextLine]; ok {
		return Outcome{}, s.Start(page, in)
	}

	return Outcome{}, nil
}

// enterHealer ports HealerEnter (stores.cpp:1876-1900).
func (s *Session) enterHealer(in GameView) (Outcome, error) {
	if done, out, err := s.enterMenuLine(12, 18, in); done {
		return out, err
	}

	if s.Dialog.CurrentTextLine == 14 {
		if in.Menus.StoreUI == StoreUIVisualGrid {
			s.Active = TalkIDNone

			return Outcome{Action: ActionOpenVisualStore}, nil
		}

		return Outcome{}, s.Start(TalkIDHealerBuy, in)
	}

	return Outcome{}, nil
}

// enterStoryteller ports StorytellerEnter (stores.cpp:1927-1945).
func (s *Session) enterStoryteller(in GameView) (Outcome, error) {
	if done, out, err := s.enterMenuLine(12, 18, in); done {
		return out, err
	}

	if s.Dialog.CurrentTextLine == 14 {
		return Outcome{}, s.Start(TalkIDStorytellerIdentify, in)
	}

	return Outcome{}, nil
}

// enterBoy ports BoyEnter (stores.cpp:1694-1724): the 50 gold look fee gates
// the offer page.
func (s *Session) enterBoy(in GameView) (Outcome, error) {
	boyHasItem := !inventory.IsEmpty(in.BoyItem)

	if boyHasItem && s.Dialog.CurrentTextLine == 18 {
		if !CanAfford(TotalGold(in.Gold, in.StashGold), boyLookFee) {
			s.OldActive = TalkIDBoy
			s.OldTextLine = 18
			s.OldScrollPos = s.Dialog.ScrollPos

			return Outcome{}, s.Start(TalkIDNoMoney, in)
		}

		if in.Menus.StoreUI == StoreUIVisualGrid {
			s.Active = TalkIDNone

			return Outcome{Action: ActionOpenVisualStore, Price: boyLookFee}, nil
		}

		return Outcome{Action: ActionNone, Price: boyLookFee}, s.Start(TalkIDBoyBuy, in)
	}

	greet := 8
	if !boyHasItem {
		greet = 12
	}

	if done, out, err := s.enterMenuLine(greet, 20, in); done {
		return out, err
	}

	s.Active = TalkIDNone

	return Outcome{}, nil
}

// enterTownerLine ports the tavern and drunk menus, which only greet and leave
// (stores.cpp:2013-2058).
func (s *Session) enterTownerLine(page TalkID, in GameView) (Outcome, error) {
	if done, out, err := s.enterMenuLine(12, 18, in); done {
		if page == TalkIDDrunk || page == TalkIDTavern {
			s.OldActive = page
		}

		return out, err
	}

	s.Active = TalkIDNone

	return Outcome{}, nil
}

// enterBarmaid ports BarmaidEnter: the stash line opens storage
// (stores.cpp:2028-2058).
func (s *Session) enterBarmaid(in GameView) (Outcome, error) {
	if done, out, err := s.enterMenuLine(12, 18, in); done {
		s.OldActive = TalkIDBarmaid

		return out, err
	}

	s.Active = TalkIDNone
	if s.Dialog.CurrentTextLine == 14 {
		return Outcome{Action: ActionOpenStorage}, nil
	}

	return Outcome{}, nil
}

// enterTalk ports TalkEnter (stores.cpp:1969-2010): the back line returns to
// the menu that opened the talk page, any other line opens its text.
func (s *Session) enterTalk(in GameView) (Outcome, error) {
	if s.Dialog.CurrentTextLine == BackButtonLine {
		return Outcome{}, s.escTo(s.OldActive, s.OldTextLine, in)
	}

	return Outcome{Action: ActionGossipText}, nil
}
