// Harness for the inputbridge tests: an all-fields-scripted ActionView
// mirroring controls' own actionWorld discipline, plus the shared when/then
// assert helpers.
package inputbridge_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/controls"
	"github.com/bmcszk/devilutionGo/internal/diablo/inputbridge"
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
)

// actionView is the harness ActionView: every field is a test-scripted value.
type actionView struct {
	spellSelect    bool
	invOpen        bool
	pcurs          controls.CursorID
	charFlag       bool
	charPanelBtn   bool
	statPts        int
	stashOpen      bool
	storeOpen      bool
	holdItemEmpty  bool
	storeButton    int
	mouse          controls.PanelRegion
	overlaysActive bool
}

// stickCommandCase is one classified push, keyed by the upstream lines it pins.
type stickCommandCase struct {
	name string
	x, y float64
	want kernel.Command
}

func newActionView() actionView {
	return actionView{pcurs: controls.CursorHand, storeButton: -1, holdItemEmpty: true}
}

func (v actionView) SpellSelectFlag() bool              { return v.spellSelect }
func (v actionView) InventoryOpen() bool                { return v.invOpen }
func (v actionView) Cursor() controls.CursorID          { return v.pcurs }
func (v actionView) CharFlag() bool                     { return v.charFlag }
func (v actionView) CharPanelButtonActive() bool        { return v.charPanelBtn }
func (v actionView) StatPoints() int                    { return v.statPts }
func (v actionView) StashOpen() bool                    { return v.stashOpen }
func (v actionView) VisualStoreOpen() bool              { return v.storeOpen }
func (v actionView) HoldItemEmpty() bool                { return v.holdItemEmpty }
func (v actionView) StoreButton() int                   { return v.storeButton }
func (v actionView) MouseInPanel() controls.PanelRegion { return v.mouse }
func (v actionView) OverlaysActive() bool               { return v.overlaysActive }

// stickDecisionCommand is the when for one stick decision: the decision cores
// classify (x,y) and the bridge carries their outputs verbatim. A dropped
// decision fails the test: every case here wants one command.
func stickDecisionCommand(t *testing.T, x, y float64) kernel.Command {
	t.Helper()

	push := controls.ClassifyStickPush(x, y)
	dir := controls.StickDirection(x, y)

	cmd, ok := inputbridge.StickCommand(push, dir)
	if !ok {
		t.Fatal("StickCommand produced no command, want one per decision")
	}

	return cmd
}

// primaryDecisionCommand is the when for one press decision: it fails the test
// when the bridge drops the decision — every press names a row (upstream's
// guard order returns an action or a deliberate no-op, never silence).
func primaryDecisionCommand(t *testing.T, v actionView) kernel.Command {
	t.Helper()

	cmd, ok := inputbridge.PrimaryActionCommand(v)
	if !ok {
		t.Fatal("PrimaryActionCommand produced no command, want one per decision")
	}

	return cmd
}

// releaseDecisionCommand is the when for a release whose guard held: it fails
// the test when the bridge drops the release.
func releaseDecisionCommand(t *testing.T, v actionView) kernel.Command {
	t.Helper()

	cmd, ok := inputbridge.ReleaseCommand(v)
	if !ok {
		t.Fatal("ReleaseCommand produced no command, want the release")
	}

	return cmd
}

// harnessActionView builds the default plain-World view: no panels, no
// overlays — the bare Interact fallthrough.
func harnessActionView() actionView { return newActionView() }
