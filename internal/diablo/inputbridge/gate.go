package inputbridge

import (
	"github.com/bmcszk/devilutionGo/internal/diablo/controls"
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
)

// OverlayGate is the overlay-refusal gate rule (devilGo-iq0d review F2): while
// an overlay owns the pointer, the press kinds the upstream call site refuses
// drop at the gate — the GATE owns the refusal, the machine's decision stays
// unchanged (ADR-018 law 6). The rule reuses the controls package's own
// OverlaysActive definition (the pinned game_controls.cpp:124-133 FINGERDOWN
// in-game-menu route and the :197-198 PadMenuNavigatorActive ||
// PadHotspellMenuActive route); it does not fork it.
//
// The button-up release never drops: upstream routes both release call sites
// BEFORE the refusals (game_controls.cpp:182-190 FINGERUP, :399-410 physical
// pad release). Every non-press kind is another rule's business and passes.
func OverlayGate(view controls.ActionView) kernel.Gate {
	return kernel.GateFunc(func(c kernel.Command) bool {
		return !view.OverlaysActive() || !overlayRefused(c)
	})
}

// overlayRefused mirrors the press-kind refusal set: the kinds whose commands
// only exist because the call site refused (game_controls.cpp:197-198), plus
// the KindStoreAct press verbs — the release verb is exempt (plrctrls.cpp
// PerformPrimaryActionRelease path, gate-free upstream).
func overlayRefused(c kernel.Command) bool {
	switch c.Kind {
	case kernel.KindPrimaryActionPress, kernel.KindSpellSelect, kernel.KindIconCursorAct:
		return true
	case kernel.KindStoreAct:
		return c.A != int32(StoreActButtonRelease)
	default:
		return false
	}
}
