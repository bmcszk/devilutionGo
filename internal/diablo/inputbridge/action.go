package inputbridge

import (
	"github.com/bmcszk/devilutionGo/internal/diablo/controls"
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
)

// actionCommandMachine is the ActionHooks recorder: it watches which hook the
// decision machine fires and carries the command that hook's row means. The
// refusal rows fire no hook, so it falls back to the view: the hourglass press
// (plrctrls.cpp:2319-2320) and the inventory-open fall-through
// (plrctrls.cpp:2337) name their own handler-free kinds and still count as
// emitted.
type actionCommandMachine struct {
	view controls.ActionView
	cmd  kernel.Command
	// fired records whether any hook fired; the refusal rows set it through
	// their own branch below.
	fired bool
}

// PrimaryActionCommand translates one PerformPrimaryAction dispatch decision
// into the command its row means: one decision, one command, the same guard
// order as the machine (plrctrls.cpp:2309-2346). The bridge re-decides with
// the machine itself — it does not re-derive any branch — so the row order
// stays the upstream order by construction.
//
// Refusal rows keep their commands: the overlay refusal still means a press
// (the gate refuses it, game_controls.cpp:127-130), the hourglass press stays
// a KindHourglassNoop (plrctrls.cpp:2319-2320) and the inventory-open
// fall-through stays a KindInventorySwallow (plrctrls.cpp:2337). Handler-free
// kinds keep the refusal in the tick log without letting any handler fire.
func PrimaryActionCommand(view controls.ActionView) (kernel.Command, bool) {
	// The call site refuses before the machine: GetControllerActionForButton
	// returns false while the menu navigator or hotspell overlay is up
	// (game_controls.cpp:197-198), so the adapter mirrors that order — the
	// command is emitted and inputbridge.OverlayGate — the kernel gate rule
	// over the same OverlaysActive source — drops it at Submit (ADR-018 law 6;
	// pinned by gate_test.go, devilGo-iq0d review F2).
	machine := actionCommandMachine{view: view}
	if view.OverlaysActive() {
		machine.emit(kernel.Command{Kind: kernel.KindPrimaryActionPress})
	}

	controls.PerformPrimaryAction(machine.view, machine.hooks())

	return machine.pressCommand()
}

// ReleaseCommand translates one PerformPrimaryActionRelease decision into its
// StoreAct button-release command (plrctrls.cpp:2349-2353). The release path
// is gate-free upstream — both release call sites run before the call-site
// refusals (game_controls.cpp:182-190, :399-410) — so this command must never
// grow a gate rule; the gate test pins that.
func ReleaseCommand(view controls.ActionView) (kernel.Command, bool) {
	machine := actionCommandMachine{view: view}
	controls.PerformPrimaryActionRelease(machine.view, machine.hooks())

	return machine.releaseCommand()
}

// pressCommand reports the press decision's command: the fired row's command,
// or the handler-free refusal row the machine's guard order left silent —
// hourglass (cursor not reset, nothing fired, plrctrls.cpp:2319-2320) and the
// inventory-open swallow (the invflag block always returns,
// plrctrls.cpp:2337). Every press names a row, so ok is always true.
func (m *actionCommandMachine) pressCommand() (kernel.Command, bool) {
	if m.fired {
		return m.cmd, true
	}

	if m.view.Cursor() == controls.CursorHourglass {
		return kernel.Command{Kind: kernel.KindHourglassNoop}, true
	}

	return kernel.Command{Kind: kernel.KindInventorySwallow}, true
}

// releaseCommand reports the release decision's command: the fired row's
// command, or nothing — upstream's release guard failing fires nothing
// (plrctrls.cpp:2350), and the release path has no refusal rows.
func (m *actionCommandMachine) releaseCommand() (kernel.Command, bool) {
	return m.cmd, m.fired
}

// hooks returns the recorder hooks: every hook stores the command its
// upstream row means and marks the decision fired.
func (m *actionCommandMachine) hooks() controls.ActionHooks {
	return controls.ActionHooks{
		SetSpell: func() {
			m.emit(kernel.Command{Kind: kernel.KindSpellSelect})
		},
		TryIconCurs: func() {
			// plrctrls.cpp:2318-2322: the cursor acted, then reset to hand.
			m.emit(kernel.Command{Kind: kernel.KindIconCursorAct, A: int32(m.view.Cursor())})
		},
		// NewCursor is the paired second half of the icon-cursor row (the reset
		// to CURSOR_HAND, plrctrls.cpp:2322): presentation state, not a command —
		// the row's IconCursorAct command already covers it.
		NewCursor: func(controls.CursorID) {},
		LiftInventoryItem: func() {
			m.emit(kernel.Command{Kind: kernel.KindInventoryLift, A: int32(m.view.MouseInPanel())})
		},
		LiftStashItem: func() {
			m.emit(kernel.Command{Kind: kernel.KindStashLift})
		},
		CheckVisualStorePaste: func() {
			m.emit(kernel.Command{Kind: kernel.KindStoreAct, A: int32(StoreActPaste)})
		},
		CheckVisualStoreButtonPress: func() {
			m.emit(kernel.Command{
				Kind: kernel.KindStoreAct,
				A:    int32(StoreActButtonPress),
				B:    int32(m.view.StoreButton()),
			})
		},
		CheckVisualStoreItem: func() {
			m.emit(kernel.Command{Kind: kernel.KindStoreAct, A: int32(StoreActItem)})
		},
		CheckChrBtns: func() bool {
			m.emit(kernel.Command{Kind: kernel.KindStatAllocate})

			// upstream flips CharPanelButtonActive when a button took the press
			// (plrctrls.cpp:2340-2343), pairing the ReleaseChrBtns second half;
			// the row's one StatAllocate command already covers the pair.
			return true
		},
		ReleaseChrBtns: func() {},
		Interact: func() {
			m.emit(kernel.Command{Kind: kernel.KindPrimaryActionPress})
		},
		CheckVisualStoreButtonRelease: func() {
			// plrctrls.cpp:2350-2352: the button-up half of the visual-store
			// press (PerformPrimaryActionRelease).
			m.emit(kernel.Command{Kind: kernel.KindStoreAct, A: int32(StoreActButtonRelease)})
		},
	}
}

// emit records one row's command.
func (m *actionCommandMachine) emit(cmd kernel.Command) {
	m.cmd = cmd
	m.fired = true
}
