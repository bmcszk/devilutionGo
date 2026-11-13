// Package inputbridge_test proves the overlay-refusal gate rule: while an
// overlay owns the pointer, the overlay-route press kinds drop at the kernel
// gate — the GATE owns the refusal (ADR-018 law 6), mirroring upstream's
// call-site refusals (game_controls.cpp:124-133, :197-198).
package inputbridge_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/controls"
	"github.com/bmcszk/devilutionGo/internal/diablo/inputbridge"
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
)

// overlayRouteCases pins the refusal row set while overlays are active: the
// press kinds the call site refuses drop; the button-up release survives —
// upstream routes both release call sites BEFORE the refusals
// (game_controls.cpp:182-190 FINGERUP, :399-410 physical pad release) — and
// every non-press kind is another rule's business, not this one's.
var overlayRouteCases = []overlayGateCase{
	{
		name: "the bare world press drops (KindPrimaryActionPress, game_controls.cpp:197-198)",
		cmd:  kernel.Command{Kind: kernel.KindPrimaryActionPress},
		want: false,
	},
	{
		name: "the spell-select press drops (KindSpellSelect, game_controls.cpp:127-130)",
		cmd:  kernel.Command{Kind: kernel.KindSpellSelect},
		want: false,
	},
	{
		name: "the icon-cursor act drops (KindIconCursorAct, game_controls.cpp:197-198)",
		cmd:  kernel.Command{Kind: kernel.KindIconCursorAct, A: int32(controls.CursorIdentify)},
		want: false,
	},
	{
		name: "the store paste press drops (KindStoreAct paste verb, game_controls.cpp:197-198)",
		cmd:  kernel.Command{Kind: kernel.KindStoreAct, A: int32(inputbridge.StoreActPaste)},
		want: false,
	},
	{
		name: "store button release survives: the release path is gate-free " +
			"upstream (game_controls.cpp:182-190, :399-410)",
		cmd:  kernel.Command{Kind: kernel.KindStoreAct, A: int32(inputbridge.StoreActButtonRelease)},
		want: true,
	},
	{
		name: "the stick push passes — the overlay refusal is the press family's rule",
		cmd:  kernel.Command{Kind: kernel.KindStickPush},
		want: true,
	},
	{
		name: "the player move passes — the gameplay gate owns it",
		cmd:  kernel.Command{Kind: kernel.KindPlayerMove},
		want: true,
	},
}

// overlayGateCase is one command offered to the gate while overlays are
// active, keyed by the upstream refusal that owns it.
type overlayGateCase struct {
	name string
	cmd  kernel.Command
	want bool
}

// TestOverlayGate_DropsOverlayRouteCommandsWhileActive pins the rule row set:
// overlays active — the press kinds drop, the release and every other kind
// pass.
func TestOverlayGate_DropsOverlayRouteCommandsWhileActive(t *testing.T) {
	for _, tc := range overlayRouteCases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			v := newActionView()
			v.overlaysActive = true
			gate := inputbridge.OverlayGate(v)

			// when
			got := gate.Allow(tc.cmd)

			// then
			if got != tc.want {
				t.Fatalf("OverlayGate.Allow(%s) = %v under active overlays, want %v", tc.cmd.Kind, got, tc.want)
			}
		})
	}
}

// TestOverlayGate_AllowsEverythingWhenOverlaysInactive pins the rule's off
// state: no overlays, no drops — the press kinds reach the machine untouched.
func TestOverlayGate_AllowsEverythingWhenOverlaysInactive(t *testing.T) {
	for _, tc := range overlayRouteCases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			gate := inputbridge.OverlayGate(newActionView())

			// when
			got := gate.Allow(tc.cmd)

			// then
			if !got {
				t.Fatalf("OverlayGate.Allow(%s) = false with no overlays active, want it through", tc.cmd.Kind)
			}
		})
	}
}

// TestOverlayGate_KernelSubmitDropsUnderOverlays pins the composition: the
// rule wired into the kernel gate surface drops the refused press at Submit,
// so it never reaches a handler (ADR-018 law 6).
func TestOverlayGate_KernelSubmitDropsUnderOverlays(t *testing.T) {
	// given
	v := newActionView()
	v.overlaysActive = true
	k := kernel.New(inputbridge.OverlayGate(v), nil)

	// when
	press := k.Submit(kernel.Command{Kind: kernel.KindPrimaryActionPress})
	release := k.Submit(kernel.Command{Kind: kernel.KindStoreAct, A: int32(inputbridge.StoreActButtonRelease)})

	// then
	if press {
		t.Fatal("kernel Submit accepted KindPrimaryActionPress under active overlays, want the gate to refuse it")
	}
	if !release {
		t.Fatal("kernel Submit refused the store button release under active overlays, " +
			"want it through — the release path is gate-free upstream")
	}
}
