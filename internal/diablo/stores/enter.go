package stores

import (
	"github.com/bmcszk/devilutionGo/internal/diablo/entities"
	"github.com/bmcszk/devilutionGo/internal/diablo/inventory"
)

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

// enterHandlers routes StoreEnter to the active page's handler
// (stores.cpp:2715-2790). Pages that do not handle Enter are absent and leave
// the store unchanged.
var enterHandlers = map[TalkID]enterHandler{
	TalkIDSmith:                   (*Session).enterSmith,
	TalkIDWitch:                   (*Session).enterWitch,
	TalkIDHealer:                  (*Session).enterHealer,
	TalkIDStoryteller:             (*Session).enterStoryteller,
	TalkIDBoy:                     (*Session).enterBoy,
	TalkIDTavern:                  townerHandler(TalkIDTavern),
	TalkIDDrunk:                   townerHandler(TalkIDDrunk),
	TalkIDBarmaid:                 (*Session).enterBarmaid,
	TalkIDGossip:                  (*Session).enterTalk,
	TalkIDSmithBuy:                vendorHandler(TalkIDSmith, 12, ActionBuyBasic),
	TalkIDSmithPremiumBuy:         vendorHandler(TalkIDSmith, 14, ActionBuyPremium),
	TalkIDHealerBuy:               vendorHandler(TalkIDHealer, 14, ActionBuyHealer),
	TalkIDWitchBuy:                vendorHandler(TalkIDWitch, 14, ActionBuyWitch),
	TalkIDSmithSell:               (*Session).enterSellList,
	TalkIDWitchSell:               (*Session).enterSellList,
	TalkIDSmithRepair:             holdHandler(TalkIDSmith, 18, ActionRepair),
	TalkIDWitchRecharge:           (*Session).enterWitchRecharge,
	TalkIDStorytellerIdentify:     holdHandler(TalkIDStoryteller, 14, ActionIdentify),
	TalkIDStorytellerIdentifyShow: (*Session).enterShowIdentify,
	TalkIDBoyBuy:                  (*Session).enterBoyBuy,
	TalkIDConfirm:                 (*Session).enterConfirm,
	TalkIDNoMoney:                 (*Session).enterRestore,
	TalkIDNoRoom:                  (*Session).enterRestore,
}

// Action is the trade ConfirmEnter executed, so the caller can apply the
// side effects the store does not own (moving the item, spending gold,
// recalculating the character).
type Action uint8

// enterHandler runs one page's Enter.
type enterHandler func(*Session, GameView) (Outcome, error)

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
	handler, ok := enterHandlers[s.Active]
	if !ok {
		return Outcome{}, nil
	}

	return handler(s, in)
}

// vendorHandler builds the Enter handler of a vendor list page, whose back line
// and action differ per page (stores.cpp:1349-1420).
func vendorHandler(parent TalkID, parentLine int, action Action) enterHandler {
	return func(s *Session, in GameView) (Outcome, error) {
		return s.enterVendorList(parent, parentLine, action, in)
	}
}

// holdHandler builds the Enter handler of a repair, recharge or identify list
// page (stores.cpp:1509-1530, 1628-1660, 1947-1968).
func holdHandler(parent TalkID, parentLine int, action Action) enterHandler {
	return func(s *Session, in GameView) (Outcome, error) {
		return s.enterHoldList(parent, parentLine, action, in)
	}
}

// townerHandler builds the Enter handler of a towner menu page
// (stores.cpp:2013-2058).
func townerHandler(page TalkID) enterHandler {
	return func(s *Session, in GameView) (Outcome, error) {
		return s.enterTownerLine(page, in)
	}
}

// enterWitchRecharge ports WitchRechargeEnter's back line, which depends on the
// store UI (stores.cpp:2533-2536).
func (s *Session) enterWitchRecharge(in GameView) (Outcome, error) {
	return s.enterHoldList(TalkIDWitch, witchRechargeBackLine(in), ActionRecharge, in)
}

// enterShowIdentify ports StorytellerIdentifyShow's Enter: the identify result
// page reopens the list (stores.cpp:1947-1968).
func (s *Session) enterShowIdentify(in GameView) (Outcome, error) {
	return Outcome{}, s.Start(TalkIDStorytellerIdentify, in)
}

// witchRechargeBackLine is StoreESC's StoreUi-dependent return line
// (stores.cpp:2533-2536).
func witchRechargeBackLine(in GameView) int {
	if in.Menus.StoreUI == StoreUIVisualGrid {
		return 16
	}

	return 18
}

// enterMenuLine is the shared handler of a towner menu page: the greet line
// opens the talk page, a trailing line closes the store, and any other line is
// the sub-page the caller maps.
func (s *Session) enterMenuLine(greetLine, closeLine int, in GameView) (bool, error) {
	switch s.Dialog.CurrentTextLine {
	case greetLine:
		s.OldTextLine = greetLine
		s.OldActive = s.Active
		s.Towner = townerDisplayName(s.Active)

		return true, s.Start(TalkIDGossip, in)
	case closeLine:
		s.Active = TalkIDNone

		return true, nil
	default:
	}

	return false, nil
}

// enterSmith ports SmithEnter's line map (stores.cpp:1293-1333).
func (s *Session) enterSmith(in GameView) (Outcome, error) {
	if in.Menus.StoreUI == StoreUIVisualGrid {
		switch s.Dialog.CurrentTextLine {
		case 10:
			_, err := s.enterMenuLine(10, 14, in)

			return Outcome{}, err
		case 12:
			s.Active = TalkIDNone

			return Outcome{Action: ActionOpenVisualStore}, nil
		case 14:
			s.Active = TalkIDNone
		default:
		}

		return Outcome{}, nil
	}

	if done, err := s.enterMenuLine(10, 20, in); done {
		return Outcome{}, err
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
			_, err := s.enterMenuLine(12, 18, in)

			return Outcome{}, err
		case 14:
			s.Active = TalkIDNone

			return Outcome{Action: ActionOpenVisualStore}, nil
		case 16:
			return Outcome{}, s.Start(TalkIDWitchRecharge, in)
		case 18:
			s.Active = TalkIDNone
		default:
		}

		return Outcome{}, nil
	}

	if done, err := s.enterMenuLine(12, 20, in); done {
		return Outcome{}, err
	}

	next := map[int]TalkID{14: TalkIDWitchBuy, 16: TalkIDWitchSell, 18: TalkIDWitchRecharge}

	if page, ok := next[s.Dialog.CurrentTextLine]; ok {
		return Outcome{}, s.Start(page, in)
	}

	return Outcome{}, nil
}

// enterHealer ports HealerEnter (stores.cpp:1876-1900).
func (s *Session) enterHealer(in GameView) (Outcome, error) {
	if done, err := s.enterMenuLine(12, 18, in); done {
		return Outcome{}, err
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
	if done, err := s.enterMenuLine(12, 18, in); done {
		return Outcome{}, err
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
		return s.enterBoyLook(in)
	}

	greet := 8
	if !boyHasItem {
		greet = 12
	}

	if done, err := s.enterMenuLine(greet, 20, in); done {
		return Outcome{}, err
	}

	s.Active = TalkIDNone

	return Outcome{}, nil
}

// enterBoyLook is BoyEnter's look branch: pay the 50 gold fee, then open Wirt's
// offer page or the visual store (stores.cpp:1700-1710).
func (s *Session) enterBoyLook(in GameView) (Outcome, error) {
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

// enterTownerLine ports the tavern and drunk menus, which only greet and leave
// (stores.cpp:2013-2058).
func (s *Session) enterTownerLine(page TalkID, in GameView) (Outcome, error) {
	if done, err := s.enterMenuLine(12, 18, in); done {
		if page == TalkIDDrunk || page == TalkIDTavern {
			s.OldActive = page
		}

		return Outcome{}, err
	}

	s.Active = TalkIDNone

	return Outcome{}, nil
}

// enterBarmaid ports BarmaidEnter: the stash line opens storage
// (stores.cpp:2028-2058).
func (s *Session) enterBarmaid(in GameView) (Outcome, error) {
	if done, err := s.enterMenuLine(12, 18, in); done {
		s.OldActive = TalkIDBarmaid

		return Outcome{}, err
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
