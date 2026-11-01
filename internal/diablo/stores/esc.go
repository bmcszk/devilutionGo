package stores

// Esc ports StoreESC (stores.cpp:2481-2553): the escape key walks back one
// page, and the top-level menus close the store.
//
// The qtextflag branch at the top of StoreESC closes the gossip text box, which
// this data port does not own; the caller handles it before calling Esc.
func (s *Session) Esc(in GameView) error {
	switch s.Active {
	case TalkIDSmith, TalkIDWitch, TalkIDBoy, TalkIDBoyBuy, TalkIDHealer,
		TalkIDStoryteller, TalkIDTavern, TalkIDDrunk, TalkIDBarmaid:
		s.Active = TalkIDNone

		return nil
	case TalkIDGossip:
		return s.escTo(s.OldActive, s.OldTextLine, in)
	case TalkIDSmithBuy:
		return s.escTo(TalkIDSmith, 12, in)
	case TalkIDSmithPremiumBuy:
		return s.escTo(TalkIDSmith, 14, in)
	case TalkIDSmithSell:
		return s.escTo(TalkIDSmith, 16, in)
	case TalkIDSmithRepair:
		return s.escTo(TalkIDSmith, 18, in)
	case TalkIDWitchBuy:
		return s.escTo(TalkIDWitch, 14, in)
	case TalkIDWitchSell:
		return s.escTo(TalkIDWitch, 16, in)
	case TalkIDWitchRecharge:
		return s.escTo(TalkIDWitch, s.witchRechargeBackLine(in), in)
	case TalkIDHealerBuy:
		return s.escTo(TalkIDHealer, 14, in)
	case TalkIDStorytellerIdentify:
		return s.escTo(TalkIDStoryteller, 14, in)
	case TalkIDStorytellerIdentifyShow:
		return s.Start(TalkIDStorytellerIdentify, in)
	case TalkIDNoMoney, TalkIDNoRoom, TalkIDConfirm:
		if err := s.Start(s.OldActive, in); err != nil {
			return err
		}

		s.Dialog.CurrentTextLine = s.OldTextLine
		s.Dialog.ScrollPos = s.OldScrollPos

		return nil
	case TalkIDNone:
	}

	return nil
}
