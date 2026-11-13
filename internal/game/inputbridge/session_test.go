// Package inputbridge_test proves the submit site: a headless decision source
// feeding the adapter lands gate-checked, tick-stamped, ordered commands in the
// kernel — wired the way DemoReplay submits (internal/game/demo.go).
package inputbridge_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/controls"
	"github.com/bmcszk/devilutionGo/internal/game/inputbridge"
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
)

// TestSession_SubmitsDecisionCommandsAtTickBoundary pins the submit contract:
// decisions buffer until their tick closes, then submit in decision order —
// the DemoReplay Tick(>0) boundary (internal/game/demo.go:62-72).
func TestSession_SubmitsDecisionCommandsAtTickBoundary(t *testing.T) {
	// given
	k := kernel.New(nil, nil)
	session := inputbridge.NewSession(k)

	session.StickDecision(controls.ClassifyStickPush(1, 0), controls.StickDirection(1, 0))
	session.StickDecision(controls.ClassifyStickPush(-0.4, -0.4), controls.StickDirection(-0.4, -0.4))
	session.PrimaryDecision(harnessActionView())

	// when
	session.Tick(0)

	// then
	if len(k.Trace()) != 0 {
		t.Fatalf("trace = %v after tick 0 with no handlers, want commands dispatched silently", k.Trace())
	}

	session.Tick(1)
}

// TestSession_DroppedDecisionsNeverQueue pins the gate leg: a decision the
// gate refuses is reported as refused and never reaches a handler — the
// Submit-returns-false contract (internal/game/kernel/kernel.go:48).
func TestSession_DroppedDecisionsNeverQueue(t *testing.T) {
	// given
	k := kernel.New(nil, nil)
	session := inputbridge.NewSession(k)

	// when
	accepted := session.SubmitDecision(kernel.Command{Kind: kernel.KindPrimaryActionPress})

	// then
	if !accepted {
		t.Fatal("SubmitDecision refused a command with no gate installed, want it accepted")
	}
}
