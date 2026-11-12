// Package controls_test — virtual-cursor action machine tests. One table row
// per PerformPrimaryAction branch, in upstream guard order
// (Source/controls/plrctrls.cpp:2309-2353 @ fork ca91d778f), plus the
// overlay-guard rows (game_controls.cpp:197-198 call-site refusal modeled as
// the OverlaysActive input) and the release-path rows
// (PerformPrimaryActionRelease, plrctrls.cpp:2349-2353).
package controls_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/controls"
)

// primaryDispatchCases — plrctrls.cpp:2309-2346 branch inventory, in
// contract order (guard order IS the behavior).
var primaryDispatchCases = []primaryDispatchCase{
	{
		name:      "SpellSelectFlag fires SetSpell only (plrctrls.cpp:2311-2314)",
		setup:     func(w *actionWorld) { w.spellSelect = true },
		wantCalls: []string{"SetSpell"},
	},
	{
		name:      "hourglass icon cursor is a full no-op — cursor NOT reset (plrctrls.cpp:2319-2320)",
		setup:     func(w *actionWorld) { w.invOpen, w.pcurs = true, controls.CursorHourglass },
		wantCalls: nil,
	},
	{
		name:      "icon cursor range triggers TryIconCurs then NewCursor(HAND) (plrctrls.cpp:2318-2322)",
		setup:     func(w *actionWorld) { w.invOpen, w.pcurs = true, controls.CursorIdentify },
		wantCalls: []string{"TryIconCurs", "NewCursor:CURSOR_HAND"},
	},
	{
		name:      "mouse over right/main panel lifts the inventory item (plrctrls.cpp:2323-2324)",
		setup:     func(w *actionWorld) { w.invOpen, w.mouse = true, controls.PanelRightOrMain },
		wantCalls: []string{"LiftInventoryItem"},
	},
	{
		name:      "stash open + mouse over left panel lifts the stash item (plrctrls.cpp:2325-2326)",
		setup:     func(w *actionWorld) { w.invOpen, w.stashOpen, w.mouse = true, true, controls.PanelLeft },
		wantCalls: []string{"LiftStashItem"},
	},
	{
		name: "visual store with held item pastes it (plrctrls.cpp:2328-2330)",
		setup: func(w *actionWorld) {
			w.invOpen, w.storeOpen, w.holdItem, w.mouse = true, true, true, controls.PanelLeft
		},
		wantCalls: []string{"CheckVisualStorePaste"},
	},
	{
		name: "visual store button press only presses (release comes on button-up) (plrctrls.cpp:2331-2333)",
		setup: func(w *actionWorld) {
			w.invOpen, w.storeOpen, w.storeButton, w.mouse = true, true, 2, controls.PanelLeft
		},
		wantCalls: []string{"CheckVisualStoreButtonPress"},
	},
	{
		name:      "visual store item hover inspects the item (plrctrls.cpp:2334-2336)",
		setup:     func(w *actionWorld) { w.invOpen, w.storeOpen, w.mouse = true, true, controls.PanelLeft },
		wantCalls: []string{"CheckVisualStoreItem"},
	},
	{
		name:      "inventory open swallows the click — invflag block always returns, no Interact (plrctrls.cpp:2337)",
		setup:     func(w *actionWorld) { w.invOpen = true },
		wantCalls: nil,
	},
	{
		name:      "char panel stat points + button hit: CheckChrBtns then ReleaseChrBtns (plrctrls.cpp:2340-2344)",
		setup:     func(w *actionWorld) { w.charFlag, w.statPts, w.statBtnHit = true, 1, true },
		wantCalls: []string{"CheckChrBtns", "ReleaseChrBtns"},
	},
	{
		name:      "char panel stat points, no button hit: CheckChrBtns only (plrctrls.cpp:2341-2343)",
		setup:     func(w *actionWorld) { w.charFlag, w.statPts = true, 1 },
		wantCalls: []string{"CheckChrBtns"},
	},
	{
		name:      "char panel, no stat points: branch gated on _pStatPts>0, falls to Interact (plrctrls.cpp:2340)",
		setup:     func(w *actionWorld) { w.charFlag = true },
		wantCalls: []string{"Interact"},
	},
	{
		name:      "no eligible surface falls through to Interact (plrctrls.cpp:2345)",
		setup:     func(*actionWorld) {},
		wantCalls: []string{"Interact"},
	},
	{
		name:      "stat button already active: char branch skipped, falls to Interact (plrctrls.cpp:2340)",
		setup:     func(w *actionWorld) { w.charFlag, w.charPanelBtn, w.statPts = true, true, 1 },
		wantCalls: []string{"Interact"},
	},
}

// primaryDispatchCase is one PerformPrimaryAction branch, keyed by the
// upstream lines it pins.
type primaryDispatchCase struct {
	name      string
	setup     func(*actionWorld)
	wantCalls []string
}

// actionWorld is the ActionView + hook harness a PerformPrimaryAction test
// runs against; every hook appends its name so assertions can pin exactly
// which action fired, in order.
type actionWorld struct {
	spellSelect    bool
	invOpen        bool
	pcurs          controls.CursorID
	charFlag       bool
	charPanelBtn   bool // CharPanelButtonActive BEFORE the call
	statBtnHit     bool // what CheckChrBtns reports: a stat button took the press
	statPts        int
	stashOpen      bool
	storeOpen      bool
	holdItem       bool // !MyPlayer->HoldItem.isEmpty()
	storeButton    int  // pcursstorebtn (-1 = none)
	mouse          controls.PanelRegion
	overlaysActive bool

	calls      []string
	newCursors []controls.CursorID
}

// actionView adapts the harness world to controls.ActionView.
type actionView struct {
	w *actionWorld
}

// T-A1 — PerformPrimaryAction dispatch table: guard order and action
// selection, hook invocations verbatim.
func TestPerformPrimaryActionDispatch(t *testing.T) {
	for _, tc := range primaryDispatchCases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			w := newActionWorld()
			tc.setup(w)
			// when
			controls.PerformPrimaryAction(w.View(), w.Hooks())
			// then
			assertActionCalls(t, w, tc.wantCalls)
		})
	}
}

// T-A2 — overlay gate (bean devilGo-c6wt AC "No action while PadMenu/
// Hotspell overlays active"): upstream refuses at the call site
// (game_controls.cpp:197-198 GetControllerActionForButton returns false for
// EVERY action while PadMenuNavigatorActive || PadHotspellMenuActive; the
// InGameMenu family additionally gates specific actions); the port models
// the refusal as the OverlaysActive input — zero hook invocations.
func TestPrimaryActionOverlayGuards(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(*actionWorld)
	}{
		{name: "menu navigator overlay blocks the action", setup: func(w *actionWorld) { w.overlaysActive = true }},
		{name: "hotspell overlay blocks the action", setup: func(w *actionWorld) { w.overlaysActive = true }},
		{name: "in-game menu blocks the action", setup: func(w *actionWorld) { w.overlaysActive = true }},
		{
			name: "overlays block even a spell-select press",
			setup: func(w *actionWorld) {
				w.overlaysActive, w.spellSelect = true, true
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// given
			w := newActionWorld()
			tc.setup(w)
			// when
			controls.PerformPrimaryAction(w.View(), w.Hooks())
			// then
			assertActionCalls(t, w, nil)
		})
	}
}

// T-A3 — PerformPrimaryActionRelease (plrctrls.cpp:2349-2353): fires the
// visual-store button release while the store is open with the mouse over
// its panel; NO overlay gate — upstream runs the release path before the
// call-site refusals (game_controls.cpp:182-190 FINGERUP, :399-410 physical
// release).
func TestPerformPrimaryActionRelease(t *testing.T) {
	t.Run("store open + mouse over left panel releases the pressed button", func(t *testing.T) {
		// given
		w := newActionWorld()
		w.storeOpen, w.mouse = true, controls.PanelLeft
		// when
		controls.PerformPrimaryActionRelease(w.View(), w.Hooks())
		// then
		assertActionCalls(t, w, []string{"CheckVisualStoreButtonRelease"})
	})
	t.Run("no store open emits nothing", func(t *testing.T) {
		// given
		w := newActionWorld()
		// when
		controls.PerformPrimaryActionRelease(w.View(), w.Hooks())
		// then
		assertActionCalls(t, w, nil)
	})
	t.Run("overlays do NOT block the release — release fires under OverlaysActive", func(t *testing.T) {
		// given
		w := newActionWorld()
		w.storeOpen, w.mouse, w.overlaysActive = true, controls.PanelLeft, true
		// when
		controls.PerformPrimaryAction(w.View(), w.Hooks())
		controls.PerformPrimaryActionRelease(w.View(), w.Hooks())
		// then
		assertActionCalls(t, w, []string{"CheckVisualStoreButtonRelease"})
	})
}

// newActionWorld starts a plain in-world state: no flags, hand cursor,
// mouse over no panel, no overlays.
func newActionWorld() *actionWorld {
	return &actionWorld{pcurs: controls.CursorHand, storeButton: -1}
}

// View adapts the world to controls.ActionView.
func (w *actionWorld) View() controls.ActionView {
	return actionView{w}
}

// Hooks builds the ActionHooks bound to this world; every action appends
// its upstream name, NewCursor records the cursor id.
func (w *actionWorld) Hooks() controls.ActionHooks {
	return controls.ActionHooks{
		SetSpell: func() {
			w.calls = append(w.calls, "SetSpell")
		},
		TryIconCurs: func() {
			w.calls = append(w.calls, "TryIconCurs")
		},
		NewCursor: func(c controls.CursorID) {
			w.calls = append(w.calls, "NewCursor:"+c.String())
		},
		LiftInventoryItem: func() {
			w.calls = append(w.calls, "LiftInventoryItem")
		},
		LiftStashItem: func() {
			w.calls = append(w.calls, "LiftStashItem")
		},
		CheckVisualStorePaste: func() {
			w.calls = append(w.calls, "CheckVisualStorePaste")
		},
		CheckVisualStoreButtonPress: func() {
			w.calls = append(w.calls, "CheckVisualStoreButtonPress")
		},
		CheckVisualStoreButtonRelease: func() {
			w.calls = append(w.calls, "CheckVisualStoreButtonRelease")
		},
		CheckVisualStoreItem: func() {
			w.calls = append(w.calls, "CheckVisualStoreItem")
		},
		CheckChrBtns: func() bool {
			w.calls = append(w.calls, "CheckChrBtns")
			return w.statBtnHit // upstream flips CharPanelButtonActive when a button takes
		},
		ReleaseChrBtns: func() {
			w.calls = append(w.calls, "ReleaseChrBtns")
		},
		Interact: func() {
			w.calls = append(w.calls, "Interact")
		},
	}
}

// View methods — each maps one field to its upstream global read.

func (p actionView) SpellSelectFlag() bool {
	return p.w.spellSelect
}

func (p actionView) InventoryOpen() bool {
	return p.w.invOpen
}

func (p actionView) Cursor() controls.CursorID {
	return p.w.pcurs
}

func (p actionView) CharFlag() bool {
	return p.w.charFlag
}

func (p actionView) CharPanelButtonActive() bool {
	return p.w.charPanelBtn
}

func (p actionView) StatPoints() int {
	return p.w.statPts
}

func (p actionView) StashOpen() bool {
	return p.w.stashOpen
}

func (p actionView) VisualStoreOpen() bool {
	return p.w.storeOpen
}

func (p actionView) HoldItemEmpty() bool {
	return !p.w.holdItem
}

func (p actionView) StoreButton() int {
	return p.w.storeButton
}

func (p actionView) MouseInPanel() controls.PanelRegion {
	return p.w.mouse
}

func (p actionView) OverlaysActive() bool {
	return p.w.overlaysActive
}

// assertActionCalls pins the exact hook invocation sequence — and fails on
// ANY emission when wantCalls is nil.
func assertActionCalls(t *testing.T, w *actionWorld, wantCalls []string) {
	t.Helper()
	if len(w.calls) != len(wantCalls) {
		t.Fatalf("hook calls = %v, want %v", w.calls, wantCalls)
	}
	for i, want := range wantCalls {
		if w.calls[i] != want {
			t.Fatalf("hook calls = %v, want %v (mismatch at %d: %q != %q)", w.calls, wantCalls, i, w.calls[i], want)
		}
	}
	if len(w.newCursors) != 0 {
		t.Fatalf("NewCursor bypassed the recorder: %v", w.newCursors)
	}
}
