// Package inputbridge is the controls-to-kernel bridge (devilGo-iq0d): it
// translates the fork stick and primary-action decision cores' outputs into
// typed kernel commands, so player input reaches the simulation only through
// the gate and the ordered pass (ADR-015, ADR-018). The device lane (real
// axis and key reading) reads these commands later; a headless decision source
// feeds the adapters directly.
//
// The adapters are pure decision translation: every accepted decision yields
// exactly one command whose payload words carry the decision cores' own
// ordinals (integer-only per ADR-015). Refusals are owned by the gate, not by
// the bridge: the bridge emits what the decision machine wanted, the gate
// drops what cannot change anything (ADR-018 law 6).
package inputbridge

import (
	"github.com/bmcszk/devilutionGo/internal/diablo/controls"
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
)

// StoreAct verbs are the KindStoreAct A word: the one visual-store press
// family and its button-up release, one value per upstream branch.
const (
	// StoreActPaste: a held item pastes into the hovered store slot
	// (plrctrls.cpp:2328-2330 CheckVisualStorePaste).
	StoreActPaste StoreAct = iota
	// StoreActButtonPress: the hovered store button presses; the release
	// comes on button-up (plrctrls.cpp:2331-2333 CheckVisualStoreButtonPress).
	StoreActButtonPress
	// StoreActItem: the hovered store item is inspected
	// (plrctrls.cpp:2334-2336 CheckVisualStoreItem).
	StoreActItem
	// StoreActButtonRelease: the pressed store button releases
	// (plrctrls.cpp:2350-2352 CheckVisualStoreButtonRelease).
	StoreActButtonRelease
)

// StoreAct names one visual-store action inside a KindStoreAct command.
type StoreAct int32

// StickCommand translates one classified left-stick decision into the
// StickPush command the kernel handler will apply. The payload carries the
// decision cores' own outputs: A the StickPush severity, B/C the resolved
// 8-way sector (the ordinals Movement reads FaceDir with). A dead stick with
// no facing is the nothing decision: it yields no command, so a quiet gamepad
// floods neither the queue nor the log. The command carries no tick — the
// submit site stamps it (ADR-018 law 1).
func StickCommand(push controls.StickPush, dir controls.AxisDirection) (kernel.Command, bool) {
	if push == controls.PushNone && dir.X == controls.AxisXNone && dir.Y == controls.AxisYNone {
		return kernel.Command{}, false
	}

	return kernel.Command{
		Kind: kernel.KindStickPush,
		A:    int32(push),
		B:    int32(dir.X),
		C:    int32(dir.Y),
	}, true
}
