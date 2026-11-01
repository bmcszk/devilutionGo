package stores

import (
	"github.com/bmcszk/devilutionGo/internal/diablo/entities"
	"github.com/bmcszk/devilutionGo/internal/diablo/inventory"
)

// The page builders: one function per Start* in stores.cpp. Each writes the
// panel lines and the CurrentItemIndex the Enter handlers read afterwards.
// The VisualGrid store UI replaces the smith and witch menu with a single
// combined entry, so StoreUI is part of every menu that branches on it.

// The two StoreUi values the store code checks (options.h, StoreUi).
const (
	// StoreUIList is StoreUi::ListWithItemGraphics.
	StoreUIList StoreUI = iota
	// StoreUIVisualGrid is StoreUi::VisualGrid.
	StoreUIVisualGrid
)

// StoreUI is options.h's StoreUi setting values the store menus branch on.
type StoreUI uint8

// MenuInputs is the state the page builders read.
type MenuInputs struct {
	// StoreUI selects the list or the visual-grid menu variant.
	StoreUI StoreUI
	// BoyItemEmpty reports whether Wirt has an item to sell.
	BoyItemEmpty bool
	// TalkQuestLines are the active quest dialog lines StartTalk shows.
	TalkQuestLines []string
}

// startSmith builds Griswold's menu (stores.cpp:432-453).
func (s *Session) startSmith(in MenuInputs) {
	d := &s.Dialog
	d.IsTextFullSize = false
	d.HasScrollbar = false
	d.AddSText(0, 1, "Welcome to the", ColorWhitegold, AlignCenter)
	d.AddSText(0, 3, "Blacksmith's shop", ColorWhitegold, AlignCenter)
	d.AddSText(0, 7, "Would you like to:", ColorWhitegold, AlignCenter)
	d.AddSelectableSText(0, 10, "Talk to Griswold", ColorBlue, AlignCenter)

	if in.StoreUI == StoreUIVisualGrid {
		d.AddSelectableSText(0, 12, "Trade / Repair", ColorWhite, AlignCenter)
		d.AddSelectableSText(0, 14, "Leave the shop", ColorWhite, AlignCenter)
	} else {
		d.AddSelectableSText(0, 12, "Buy basic items", ColorWhite, AlignCenter)
		d.AddSelectableSText(0, 14, "Buy premium items", ColorWhite, AlignCenter)
		d.AddSelectableSText(0, 16, "Sell items", ColorWhite, AlignCenter)
		d.AddSelectableSText(0, 18, "Repair items", ColorWhite, AlignCenter)
		d.AddSelectableSText(0, 20, "Leave the shop", ColorWhite, AlignCenter)
	}

	d.AddSLine(5)
	s.ItemCount = 20
}

// startWitch builds Adria's menu (stores.cpp:690-711).
func (s *Session) startWitch(in MenuInputs) {
	d := &s.Dialog
	d.IsTextFullSize = false
	d.HasScrollbar = false
	d.AddSText(0, 2, "Witch's shack", ColorWhitegold, AlignCenter)
	d.AddSText(0, 9, "Would you like to:", ColorWhitegold, AlignCenter)
	d.AddSelectableSText(0, 12, "Talk to Adria", ColorBlue, AlignCenter)

	if in.StoreUI == StoreUIVisualGrid {
		d.AddSelectableSText(0, 14, "Buy / Sell", ColorWhite, AlignCenter)
		d.AddSelectableSText(0, 16, "Recharge staves", ColorWhite, AlignCenter)
		d.AddSelectableSText(0, 18, "Leave the shack", ColorWhite, AlignCenter)
		d.AddSLine(4)
	} else {
		d.AddSelectableSText(0, 14, "Buy items", ColorWhite, AlignCenter)
		d.AddSelectableSText(0, 16, "Sell items", ColorWhite, AlignCenter)
		d.AddSelectableSText(0, 18, "Recharge staves", ColorWhite, AlignCenter)
		d.AddSelectableSText(0, 20, "Leave the shack", ColorWhite, AlignCenter)
		d.AddSLine(5)
	}

	s.ItemCount = 20
}

// startHealer builds Pepin's menu (stores.cpp:1030-1043).
func (s *Session) startHealer() {
	d := &s.Dialog
	d.IsTextFullSize = false
	d.HasScrollbar = false
	d.AddSText(0, 1, "Welcome to the", ColorWhitegold, AlignCenter)
	d.AddSText(0, 3, "Healer's home", ColorWhitegold, AlignCenter)
	d.AddSText(0, 9, "Would you like to:", ColorWhitegold, AlignCenter)
	d.AddSelectableSText(0, 12, "Talk to Pepin", ColorBlue, AlignCenter)
	d.AddSelectableSText(0, 14, "Buy items", ColorWhite, AlignCenter)
	d.AddSelectableSText(0, 18, "Leave Healer's home", ColorWhite, AlignCenter)
	d.AddSLine(5)
	s.ItemCount = 20
}

// startStoryteller builds Cain's menu (stores.cpp:1072-1082).
func (s *Session) startStoryteller() {
	d := &s.Dialog
	d.IsTextFullSize = false
	d.HasScrollbar = false
	d.AddSText(0, 2, "The Town Elder", ColorWhitegold, AlignCenter)
	d.AddSText(0, 9, "Would you like to:", ColorWhitegold, AlignCenter)
	d.AddSelectableSText(0, 12, "Talk to Cain", ColorBlue, AlignCenter)
	d.AddSelectableSText(0, 14, "Identify an item", ColorWhite, AlignCenter)
	d.AddSelectableSText(0, 18, "Say goodbye", ColorWhite, AlignCenter)
	d.AddSLine(5)
}

// startBoy builds Wirt's menu; the offer lines depend on whether he has an item
// (stores.cpp:971-988).
func (s *Session) startBoy(in GameView) {
	d := &s.Dialog
	d.IsTextFullSize = false
	d.HasScrollbar = false
	d.AddSText(0, 2, "Wirt the Peg-legged boy", ColorWhitegold, AlignCenter)
	d.AddSLine(5)

	if !inventory.IsEmpty(in.BoyItem) {
		d.AddSelectableSText(0, 8, "Talk to Wirt", ColorBlue, AlignCenter)
		d.AddSText(0, 12, "I have something for sale,", ColorWhitegold, AlignCenter)
		d.AddSText(0, 14, "but it will cost 50 gold", ColorWhitegold, AlignCenter)
		d.AddSText(0, 16, "just to take a look. ", ColorWhitegold, AlignCenter)
		d.AddSelectableSText(0, 18, "What have you got?", ColorWhite, AlignCenter)
		d.AddSelectableSText(0, 20, "Say goodbye", ColorWhite, AlignCenter)

		return
	}

	d.AddSelectableSText(0, 12, "Talk to Wirt", ColorBlue, AlignCenter)
	d.AddSelectableSText(0, 18, "Say goodbye", ColorWhite, AlignCenter)
}

// startTavern builds Ogden's menu (stores.cpp:1255-1271).
func (s *Session) startTavern() {
	d := &s.Dialog
	d.IsTextFullSize = false
	d.HasScrollbar = false
	d.AddSText(0, 1, "Welcome to the", ColorWhitegold, AlignCenter)
	d.AddSText(0, 3, "Rising Sun", ColorWhitegold, AlignCenter)
	d.AddSText(0, 9, "Would you like to:", ColorWhitegold, AlignCenter)
	d.AddSelectableSText(0, 12, "Talk to Ogden", ColorBlue, AlignCenter)
	d.AddSelectableSText(0, 18, "Leave the tavern", ColorWhite, AlignCenter)
	d.AddSLine(5)
	s.ItemCount = 20
}

// startBarmaid builds Gillian's menu (stores.cpp:1273-1285).
func (s *Session) startBarmaid() {
	d := &s.Dialog
	d.IsTextFullSize = false
	d.HasScrollbar = false
	d.AddSText(0, 2, "Gillian", ColorWhitegold, AlignCenter)
	d.AddSText(0, 9, "Would you like to:", ColorWhitegold, AlignCenter)
	d.AddSelectableSText(0, 12, "Talk to Gillian", ColorBlue, AlignCenter)
	d.AddSelectableSText(0, 14, "Access Storage", ColorWhite, AlignCenter)
	d.AddSelectableSText(0, 18, "Say goodbye", ColorWhite, AlignCenter)
	d.AddSLine(5)
	s.ItemCount = 20
}

// startDrunk builds Farnham's menu (stores.cpp:1273-1285's Drunk sibling,
// stores.cpp StartDrunk).
func (s *Session) startDrunk() {
	d := &s.Dialog
	d.IsTextFullSize = false
	d.HasScrollbar = false
	d.AddSText(0, 2, "Farnham the Drunk", ColorWhitegold, AlignCenter)
	d.AddSText(0, 9, "Would you like to:", ColorWhitegold, AlignCenter)
	d.AddSelectableSText(0, 12, "Talk to Farnham", ColorBlue, AlignCenter)
	d.AddSelectableSText(0, 18, "Say Goodbye", ColorWhite, AlignCenter)
	d.AddSLine(5)
	s.ItemCount = 20
}

// startTalk builds the towner talk list: the active quest dialogs, then the
// "Gossip" entry, then the back button (stores.cpp:1204-1253).
func (s *Session) startTalk(in MenuInputs) {
	d := &s.Dialog
	d.IsTextFullSize = false
	d.HasScrollbar = false
	d.AddSText(0, 2, "Talk to "+s.Towner, ColorWhitegold, AlignCenter)
	d.AddSLine(5)

	sn := 15 - len(in.TalkQuestLines)
	la := 2

	if len(in.TalkQuestLines) > 6 {
		sn = 14 - len(in.TalkQuestLines)/2
		la = 1
	}

	sn2 := sn - 2
	for _, line := range in.TalkQuestLines {
		d.AddSelectableSText(0, sn, line, ColorWhite, AlignCenter)
		sn += la
	}

	d.AddSelectableSText(0, sn2, "Gossip", ColorBlue, AlignCenter)
	d.AddOptionsBackButton()
}

// startNoMoney ports StoreNoMoney (stores.cpp:909-917): rebuild the page that
// failed and replace its body with the gold warning.
func (s *Session) startNoMoney(in GameView) error {
	if err := s.Start(s.OldActive, in); err != nil {
		return err
	}

	d := &s.Dialog
	d.HasScrollbar = false
	d.IsTextFullSize = true
	d.RenderGold = true
	d.ClearSText(5, 23)
	d.AddSelectableSText(0, 14, "You do not have enough gold", ColorWhite, AlignCenter)

	return nil
}

// startNoRoom ports StoreNoRoom (stores.cpp:919-925).
func (s *Session) startNoRoom(in GameView) error {
	if err := s.Start(s.OldActive, in); err != nil {
		return err
	}

	d := &s.Dialog
	d.HasScrollbar = false
	d.ClearSText(5, 23)
	d.AddSelectableSText(0, 14, "You do not have enough room in inventory", ColorWhite, AlignCenter)

	return nil
}

// startConfirm builds the yes/no confirmation page for the item under
// negotiation; OldActive is the page that asked (stores.cpp:927-969).
func (s *Session) startConfirm(in GameView) error {
	if err := s.Start(s.OldActive, in); err != nil {
		return err
	}

	d := &s.Dialog
	item := s.TempItem
	itemColor := TextColor(item)

	d.HasScrollbar = false
	d.ClearSText(5, 23)
	d.AddSText(20, 8, itemName(item), itemColor, AlignDefault)
	d.AddSTextVal(8, item.Ivalue)
	d.AddSText(0, 15, confirmPrompt(s.OldActive), ColorWhite, AlignCenter)
	d.AddSelectableSText(0, 18, "Yes", ColorWhite, AlignCenter)
	d.AddSelectableSText(0, 20, "No", ColorWhite, AlignCenter)

	return nil
}

// startIdentifyShow ports StartStorytellerIdentifyShow (stores.cpp:1190-1202):
// Cain shows the card of the item he just identified.
func (s *Session) startIdentifyShow(item entities.Item, in GameView) error {
	if err := s.Start(s.OldActive, in); err != nil {
		return err
	}

	d := &s.Dialog
	d.HasScrollbar = false
	d.ClearSText(5, 23)

	itemColor := TextColor(item)
	d.AddSText(0, 7, "This item is:", ColorWhite, AlignCenter)
	d.AddSText(20, 11, itemName(item), itemColor, AlignDefault)
	d.AddSelectableSText(0, 18, "Done", ColorWhite, AlignCenter)

	return nil
}

// confirmPrompt maps the page that asked for confirmation to its prompt
// (stores.cpp:940-964).
func confirmPrompt(old TalkID) string {
	switch old {
	case TalkIDBoyBuy:
		return "Do we have a deal?"
	case TalkIDStorytellerIdentify:
		return "Are you sure you want to identify this item?"
	case TalkIDHealerBuy, TalkIDSmithPremiumBuy, TalkIDWitchBuy, TalkIDSmithBuy:
		return "Are you sure you want to buy this item?"
	case TalkIDWitchRecharge:
		return "Are you sure you want to recharge this item?"
	case TalkIDSmithSell, TalkIDWitchSell:
		return "Are you sure you want to sell this item?"
	case TalkIDSmithRepair:
		return "Are you sure you want to repair this item?"
	default:
	}

	return ""
}

// itemName is Item::getName's panel-facing part: the identified name when the
// item is identified, else the unidentified one (items.h getName).
func itemName(item entities.Item) string {
	pick := item.IName[:]
	if item.Identified {
		pick = item.Name[:]
	}

	return cString(pick)
}

// cString trims a NUL-terminated name buffer to a Go string.
func cString(buf []byte) string {
	for i, b := range buf {
		if b == 0 {
			return string(buf[:i])
		}
	}

	return string(buf)
}
