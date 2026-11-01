package stores

// escHandlers maps a sub-page to its escape action (stores.cpp:2481-2553).
var escHandlers = map[TalkID]escHandler{
	TalkIDGossip:                  (*Session).escToOldPage,
	TalkIDSmithBuy:                escTo(TalkIDSmith, 12),
	TalkIDSmithPremiumBuy:         escTo(TalkIDSmith, 14),
	TalkIDSmithSell:               escTo(TalkIDSmith, 16),
	TalkIDSmithRepair:             escTo(TalkIDSmith, 18),
	TalkIDWitchBuy:                escTo(TalkIDWitch, 14),
	TalkIDWitchSell:               escTo(TalkIDWitch, 16),
	TalkIDWitchRecharge:           (*Session).escWitchRecharge,
	TalkIDHealerBuy:               escTo(TalkIDHealer, 14),
	TalkIDStorytellerIdentify:     escTo(TalkIDStoryteller, 14),
	TalkIDStorytellerIdentifyShow: (*Session).escShowIdentify,
	TalkIDNoMoney:                 (*Session).escRestore,
	TalkIDNoRoom:                  (*Session).escRestore,
	TalkIDConfirm:                 (*Session).escRestore,
}

// escHandler runs one page's escape action.
type escHandler func(*Session, GameView) error

// Esc ports StoreESC (stores.cpp:2481-2553): the escape key walks back one
// page, and the top-level menus close the store.
//
// The qtextflag branch at the top of StoreESC closes the gossip text box, which
// this data port does not own; the caller handles it before calling Esc.
func (s *Session) Esc(in GameView) error {
	if isTopLevelStore(s.Active) {
		s.Active = TalkIDNone

		return nil
	}

	handler, ok := escHandlers[s.Active]
	if !ok {
		return nil
	}

	return handler(s, in)
}

// isTopLevelStore reports whether page is a store's own menu, which escape
// closes outright (stores.cpp:2483-2493).
func isTopLevelStore(page TalkID) bool {
	switch page {
	case TalkIDSmith, TalkIDWitch, TalkIDBoy, TalkIDBoyBuy, TalkIDHealer,
		TalkIDStoryteller, TalkIDTavern, TalkIDDrunk, TalkIDBarmaid:
		return true
	default:
		return false
	}
}

// escTo builds a handler that walks back to parent at parentLine.
func escTo(parent TalkID, parentLine int) escHandler {
	return func(s *Session, in GameView) error {
		return s.escTo(parent, parentLine, in)
	}
}

// escToOldPage returns from a gossip text to the page that opened it
// (stores.cpp:2495-2497).
func (s *Session) escToOldPage(in GameView) error {
	return s.escTo(s.OldActive, s.OldTextLine, in)
}

// escWitchRecharge returns to the witch's menu on the recharge line, which
// depends on the store UI (stores.cpp:2533-2536).
func (s *Session) escWitchRecharge(in GameView) error {
	return s.escTo(TalkIDWitch, witchRechargeBackLine(in), in)
}

// escShowIdentify returns from the identify result page to the list
// (stores.cpp:2545-2546).
func (s *Session) escShowIdentify(in GameView) error {
	return s.Start(TalkIDStorytellerIdentify, in)
}

// escRestore returns from a failure or confirmation page to the page that
// opened it, restoring its selected line and scroll (stores.cpp:2547-2552).
func (s *Session) escRestore(in GameView) error {
	if err := s.Start(s.OldActive, in); err != nil {
		return err
	}

	s.Dialog.CurrentTextLine = s.OldTextLine
	s.Dialog.ScrollPos = s.OldScrollPos

	return nil
}
