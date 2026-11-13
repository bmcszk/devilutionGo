// Package inputbridge_test proves the submit site: a headless decision source
// feeding the adapter lands gate-checked, tick-stamped, ordered commands in the
// kernel — wired the way DemoReplay submits (internal/game/demo.go).
package inputbridge_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/controls"
	"github.com/bmcszk/devilutionGo/internal/diablo/inputbridge"
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
)

// TestSession_SubmitsDecisionCommandsAtTickBoundary pins the submit contract:
// decisions buffer until their tick closes, then submit at that tick in
// decision order — nothing dispatches before the boundary and nothing leaks
// across one (the DemoReplay Tick(>0) discipline, internal/game/demo.go:62-72;
// review F5).
func TestSession_SubmitsDecisionCommandsAtTickBoundary(t *testing.T) {
	// given
	k := kernel.New(nil, nil)
	session := inputbridge.NewSession(k)

	session.StickDecision(controls.ClassifyStickPush(1, 0), controls.StickDirection(1, 0))
	session.PrimaryDecision(harnessActionView())

	// when the tick has not closed yet
	before := k.TraceCommandsAt(0)

	// then nothing dispatched
	if len(before) != 0 {
		t.Fatalf("%d commands dispatched before the tick boundary, want the decisions buffered", len(before))
	}

	// when tick 0 closes
	session.Tick(0)

	// then both decisions dispatched at tick 0 in decision order
	recorded := k.TraceCommandsAt(0)
	if len(recorded) != 2 {
		t.Fatalf("tick 0 dispatched %d commands, want the 2 buffered decisions", len(recorded))
	}
	if recorded[0].Kind != kernel.KindStickPush || recorded[1].Kind != kernel.KindPrimaryActionPress {
		t.Fatalf("tick 0 dispatched %s then %s, want the stick decision before the primary decision",
			recorded[0].Kind, recorded[1].Kind)
	}

	// given the next tick's decision buffers
	session.StickDecision(controls.ClassifyStickPush(-0.4, -0.4), controls.StickDirection(-0.4, -0.4))

	// when tick 1 closes
	session.Tick(1)

	// then only the new decision dispatched, at tick 1
	next := k.TraceCommandsAt(1)
	if len(next) != 1 || next[0].Kind != kernel.KindStickPush {
		t.Fatalf("tick 1 dispatched %+v, want only the decision buffered for tick 1", next)
	}
}

// TestSession_CommandsCarryTheirPassTick pins the stamp discipline: every
// buffered command leaves Tick stamped with the pass tick it belongs to
// (mirroring demo.go's buffer; ADR-018 law 1, review F3). Tick 0 is excluded
// from the tape: zero is indistinguishable from unstamped.
func TestSession_CommandsCarryTheirPassTick(t *testing.T) {
	// given
	k := kernel.New(nil, nil)
	session := inputbridge.NewSession(k)

	session.StickDecision(controls.ClassifyStickPush(1, 0), controls.StickDirection(1, 0))

	// when tick 1 closes
	session.Tick(1)

	// then the dispatched command carries the pass tick
	recorded := k.TraceCommandsAt(1)
	if len(recorded) != 1 {
		t.Fatalf("tick 1 dispatched %d commands, want the one decision", len(recorded))
	}
	if recorded[0].Tick != 1 {
		t.Fatalf("dispatched %s carries Tick %d, want the pass tick 1", recorded[0].Kind, recorded[0].Tick)
	}
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
