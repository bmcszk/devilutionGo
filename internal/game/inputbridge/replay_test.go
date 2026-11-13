// Package inputbridge_test proves the replay leg (AC 3): a recorded decision
// session replayed through the tick log reproduces the same trace hash — the
// bridge is event-sourced, not a callback bus with extra steps. Mirrors the
// persist tick-log rig (internal/persist/ticklog_test.go recordRun/replay).
package inputbridge_test

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/controls"
	"github.com/bmcszk/devilutionGo/internal/game/inputbridge"
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
	"github.com/bmcszk/devilutionGo/internal/game/world"
	"github.com/bmcszk/devilutionGo/internal/persist"
)

// TestReplay_RecordedDecisionSessionReproducesTheTraceHash is the load-bearing
// proof (AC 3): the recorded session replays through the tick log to the same
// trace hash — same commands, same pass order, same events, same checksum
// trail (ADR-018 law 4).
func TestReplay_RecordedDecisionSessionReproducesTheTraceHash(t *testing.T) {
	// given: a recorded session and the hash its live run produced
	hash, path := recordDecisionSession(t.Context(), t, 12)

	// when: a fresh kernel replays the recorded commands through the log
	replayHash, err := replayDecisionLog(t.Context(), t, path)
	if err != nil {
		t.Fatalf("replayDecisionLog: %v", err)
	}

	// then
	if hash == 0 {
		t.Fatal("recorded run produced a zero trace hash, want the decisions' events")
	}
	if replayHash != hash {
		t.Fatalf("replay hashed %016x, want the recorded %016x — the bridge must be event-sourced",
			replayHash, hash)
	}
	t.Logf("PROOF recorded=%016x replayed=%016x equal=%v", hash, replayHash, replayHash == hash)
}

// TestReplay_AlteredCommandDiverges pins the desync value of the proof: one
// altered payload word in the log splits the replay hash from the recorded
// one — a silent callback bus would hide this.
func TestReplay_AlteredCommandDiverges(t *testing.T) {
	// given
	hash, path := recordDecisionSession(t.Context(), t, 12)

	// when: the log's first stick command gains a payload delta
	replayHash, err := replayDecisionLog(t.Context(), t, path, func(cmd *kernel.Command) {
		if cmd.Kind == kernel.KindStickPush {
			cmd.A++
		}
	})
	if err != nil {
		t.Fatalf("replayDecisionLog: %v", err)
	}

	// then
	if replayHash == hash {
		t.Fatal("altered log replayed to the recorded hash, want divergence")
	}
}

// replayDecisionLog loads the tick log from the recorded store, feeds its
// commands to a fresh rig kernel in tick order, and returns the replay's
// trace hash. Each mutator, when present, adjusts one stored command first —
// the desync probe.
func replayDecisionLog(
	ctx context.Context,
	t *testing.T,
	path string,
	mutators ...func(*kernel.Command),
) (uint64, error) {
	t.Helper()

	store, err := persist.OpenSQLiteStore(ctx, path)
	if err != nil {
		return 0, fmt.Errorf("inputbridge: open replay store: %w", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	log, err := store.LoadTickLog(ctx)
	if err != nil {
		return 0, fmt.Errorf("inputbridge: load replay log: %w", err)
	}

	k := bridgeRig(7)

	for _, rec := range log {
		replayRecordedTick(t, k, rec, mutators)
	}

	return k.TraceHash(), nil
}

// replayRecordedTick is the when for one recorded tick: apply the desync
// mutators to the stored commands, submit them through the gate in order and
// run the pass. A gate-refused command is a replay failure, not a silent drop
// — the log must replay through the same gate it was recorded through.
func replayRecordedTick(t *testing.T, k *kernel.Kernel, rec persist.TickRecord, mutators []func(*kernel.Command)) {
	t.Helper()

	for i := range rec.Commands {
		for _, mutate := range mutators {
			mutate(&rec.Commands[i])
		}
	}

	for _, cmd := range rec.Commands {
		if !k.Submit(cmd) {
			t.Fatalf("replay tick %d: command %q rejected by the gate", rec.Tick, cmd.Kind)
		}
	}

	k.Pass(rec.Tick)
}

// playDecisionTape is the scripted session: a firm east push every other tick
// and a bare world press every third — the recorded decision stream.
func playDecisionTape(session *inputbridge.Session, tick int) {
	if tick%2 == 0 {
		session.StickDecision(controls.ClassifyStickPush(1, 0), controls.StickDirection(1, 0))
	}
	if tick%3 == 0 {
		session.PrimaryDecision(harnessActionView())
	}
}

// bridgeRig is the replay rig: the world movement handler over the grid slice
// plus the StickPush handler, whose event announces the resolved facing.
func bridgeRig(seed uint32) *kernel.Kernel {
	k := kernel.New(nil, kernel.NewStreams(seed, "ai"))
	world.RegisterMovement(k, world.NewState())

	k.Register(kernel.KindStickPush, func(c kernel.Command, _ *kernel.Streams) []kernel.Event {
		return []kernel.Event{{Kind: kernel.KindPointerMoved, A: c.A, B: c.B, C: c.C}}
	})

	return k
}

// recordDecisionSession drives the session for ticks ticks with a scripted
// decision tape, records every pass into the writer, and returns the trace
// hash the run produced.
func recordDecisionSession(ctx context.Context, t *testing.T, ticks int) (uint64, string) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "bridge.db")
	store, err := persist.OpenSQLiteStore(ctx, path)
	if err != nil {
		t.Fatalf("OpenSQLiteStore: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	k := bridgeRig(7)
	session := inputbridge.NewSession(k)
	writer := store.NewTickLogWriter(persist.TickLogOptions{SnapshotEvery: -1, Batch: 8})

	for tick := range ticks {
		playDecisionTape(session, tick)
		session.Tick(uint32(tick))

		rec := persist.TickRecord{
			Tick:     uint32(tick),
			Commands: k.TraceCommandsAt(uint32(tick)),
			Events:   k.TraceEventsAt(uint32(tick)),
			Checksum: k.TraceChecksumAt(uint32(tick)),
		}
		if err := writer.Record(ctx, rec); err != nil {
			t.Fatalf("Record tick %d: %v", tick, err)
		}
	}

	if err := writer.Flush(ctx); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	return k.TraceHash(), path
}
