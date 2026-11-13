// Package inputbridge_test proves the primary-action bridge: one command per
// PerformPrimaryAction dispatch row, the refusals owned by the gate, the
// release path gate-free — upstream rows cited per test.
package inputbridge_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/controls"
	"github.com/bmcszk/devilutionGo/internal/game/inputbridge"
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
)

var primaryCommandCases = []struct {
	name  string
	setup func(*actionView)
	want  kernel.Command
}{
	{
		name:  "spell-select press assigns the spell (plrctrls.cpp:2311-2314 SetSpell)",
		setup: func(v *actionView) { v.spellSelect = true },
		want: kernel.Command{
			Kind: kernel.KindSpellSelect,
		},
	},
	{
		name:  "icon cursor in the item range acts on the target (plrctrls.cpp:2318-2322 TryIconCurs+NewCursor)",
		setup: func(v *actionView) { v.invOpen, v.pcurs = true, controls.CursorIdentify },
		want: kernel.Command{
			Kind: kernel.KindIconCursorAct,
			A:    int32(controls.CursorIdentify),
		},
	},
	{
		name:  "right/main panel lift lifts the inventory item (plrctrls.cpp:2323-2324 LiftInventoryItem)",
		setup: func(v *actionView) { v.invOpen, v.mouse = true, controls.PanelRightOrMain },
		want: kernel.Command{
			Kind: kernel.KindInventoryLift,
			A:    int32(controls.PanelRightOrMain),
		},
	},
	{
		name:  "stash open + left panel lifts the stash item (plrctrls.cpp:2325-2326 LiftStashItem)",
		setup: func(v *actionView) { v.invOpen, v.stashOpen, v.mouse = true, true, controls.PanelLeft },
		want: kernel.Command{
			Kind: kernel.KindStashLift,
		},
	},
	{
		name: "visual store with held item pastes it (plrctrls.cpp:2328-2330 CheckVisualStorePaste)",
		setup: func(v *actionView) {
			v.invOpen, v.storeOpen, v.holdItemEmpty, v.mouse = true, true, false, controls.PanelLeft
		},
		want: kernel.Command{
			Kind: kernel.KindStoreAct,
			A:    int32(inputbridge.StoreActPaste),
		},
	},
	{
		name: "visual store button press only presses (plrctrls.cpp:2331-2333 ButtonPress)",
		setup: func(v *actionView) {
			v.invOpen, v.storeOpen, v.storeButton, v.mouse = true, true, 2, controls.PanelLeft
		},
		want: kernel.Command{
			Kind: kernel.KindStoreAct,
			A:    int32(inputbridge.StoreActButtonPress),
			B:    2,
		},
	},
	{
		name: "visual store hover inspects the item (plrctrls.cpp:2334-2336 CheckVisualStoreItem)",
		setup: func(v *actionView) {
			v.invOpen, v.storeOpen, v.mouse = true, true, controls.PanelLeft
		},
		want: kernel.Command{
			Kind: kernel.KindStoreAct,
			A:    int32(inputbridge.StoreActItem),
		},
	},
	{
		name:  "char panel stat allocation (plrctrls.cpp:2340-2344 CheckChrBtns/ReleaseChrBtns)",
		setup: func(v *actionView) { v.charFlag, v.statPts = true, 1 },
		want: kernel.Command{
			Kind: kernel.KindStatAllocate,
		},
	},
	{
		name:  "bare world press falls through to Interact (plrctrls.cpp:2345)",
		setup: func(*actionView) {},
		want: kernel.Command{
			Kind: kernel.KindPrimaryActionPress,
		},
	},
}

// TestPrimaryActionCommand_TestCase pins the row->kind mapping: every
// PerformPrimaryAction branch that fires a hook maps to its command kind with
// its payload (one command per decision, no callback re-run).
func TestPrimaryActionCommand_TestCase(t *testing.T) {
	for _, tc := range primaryCommandCases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			v := newActionView()
			tc.setup(&v)

			// when
			got := primaryDecisionCommand(t, v)
			// then
			if got != tc.want {
				t.Fatalf("PrimaryActionCommand = %+v, want %+v", got, tc.want)
			}
		})
	}
}

// TestPrimaryActionCommand_RefusalsOwnedByTheGate pins the refusal rows: the
// gate owns them (ADR-018 law 6), so the decision yields the command the
// machine wants and the gate drops it — the refusal must NOT change the
// command's kind or payload.
func TestPrimaryActionCommand_RefusalsOwnedByTheGate(t *testing.T) {
	refusals := []struct {
		name  string
		setup func(*actionView)
		want  kernel.Command
	}{
		{
			name:  "overlay refusal: the gate refuses the press (game_controls.cpp:127-130)",
			setup: func(v *actionView) { v.overlaysActive = true },
			want: kernel.Command{
				Kind: kernel.KindPrimaryActionPress,
			},
		},
		{
			name:  "hourglass cursor: handler-free noop, cursor NOT reset (plrctrls.cpp:2319-2320)",
			setup: func(v *actionView) { v.invOpen, v.pcurs = true, controls.CursorHourglass },
			want: kernel.Command{
				Kind: kernel.KindHourglassNoop,
			},
		},
		{
			name:  "inventory-open swallow: the invflag block always returns (plrctrls.cpp:2337)",
			setup: func(v *actionView) { v.invOpen = true },
			want: kernel.Command{
				Kind: kernel.KindInventorySwallow,
			},
		},
	}

	for _, tc := range refusals {
		t.Run(tc.name, func(t *testing.T) {
			// given
			v := newActionView()
			tc.setup(&v)

			// when
			got := primaryDecisionCommand(t, v)
			// then
			if got != tc.want {
				t.Fatalf("PrimaryActionCommand = %+v, want %+v", got, tc.want)
			}
		})
	}
}

// TestReleaseCommand_GateFree pins the release path: upstream runs the release
// call sites BEFORE the overlay refusals (game_controls.cpp:182-190 FINGERUP,
// :399-410 physical pad release), so the command is emitted regardless of
// overlays; the kernel wires no gate rule for it.
func TestReleaseCommand_GateFree(t *testing.T) {
	t.Run("store open + mouse over left panel releases the pressed button (plrctrls.cpp:2350-2352)", func(t *testing.T) {
		// given
		v := newActionView()
		v.storeOpen, v.mouse = true, controls.PanelLeft

		// when
		got := releaseDecisionCommand(t, v)
		// then
		want := kernel.Command{Kind: kernel.KindStoreAct, A: int32(inputbridge.StoreActButtonRelease)}
		if got != want {
			t.Fatalf("ReleaseCommand = %+v, want %+v", got, want)
		}
	})
	t.Run("no store open emits nothing (plrctrls.cpp:2350: the guard fails)", func(t *testing.T) {
		// given
		v := newActionView()

		// when
		_, ok := inputbridge.ReleaseCommand(v)
		// then
		if ok {
			t.Fatal("ReleaseCommand emitted a command, want none when the release guard fails")
		}
	})
	t.Run("overlays do NOT block the release (game_controls.cpp:182-190)", func(t *testing.T) {
		// given
		v := newActionView()
		v.storeOpen, v.mouse, v.overlaysActive = true, controls.PanelLeft, true

		// when
		got := releaseDecisionCommand(t, v)
		// then
		if got.A != int32(inputbridge.StoreActButtonRelease) {
			t.Fatalf("ReleaseCommand A = %d, want the button-release verb", got.A)
		}
	})
}
