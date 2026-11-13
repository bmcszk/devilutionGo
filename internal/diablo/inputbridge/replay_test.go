// Package inputbridge_test proves the replay leg (AC 3): a recorded decision
// session replayed through the tick log reproduces the recorded checksum
// trail — the bridge is event-sourced, not a callback bus with extra steps.
// The replay runs on persist.Replay's shipped loop, so the stored checksum
// trail is actually compared, not just rewritten (review F4); the rig mirrors
// the persist tick-log rig (internal/persist/ticklog_test.go recordRun).
package inputbridge_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/bmcszk/devilutionGo/internal/diablo/controls"
	"github.com/bmcszk/devilutionGo/internal/diablo/inputbridge"
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
	"github.com/bmcszk/devilutionGo/internal/game/world"
	"github.com/bmcszk/devilutionGo/internal/persist"
)

// TestReplay_RecordedDecisionSessionReproducesTheTraceHash is the load-bearing
// proof (AC 3): the recorded session replays through the tick log to the same
// folded trail hash — same commands, same pass order, same events, and the
// stored per-tick checksums actually verified against the replay (ADR-018
// law 4).
func TestReplay_RecordedDecisionSessionReproducesTheTraceHash(t *testing.T) {
	// given: a recorded session and the folded trail hash its live run produced
	final, path := recordDecisionSession(t.Context(), t, 12)

	// when: a fresh kernel replays the recorded commands through the log
	result := replayDecisionLog(t.Context(), t, path)

	// then
	if result.Mismatch != nil {
		t.Fatalf("replay diverged from the stored checksum trail: %v", result.Mismatch)
	}
	if final == 0 {
		t.Fatal("recorded run produced a zero folded trail hash, want the decisions' events")
	}
	if result.Final != final {
		t.Fatalf("replay hashed %016x, want the recorded %016x — the bridge must be event-sourced",
			result.Final, final)
	}
	t.Logf("PROOF recorded=%016x replayed=%016x equal=%v", final, result.Final, result.Final == final)
}

// TestReplay_AlteredCommandDiverges pins the desync value of the proof: one
// altered payload word in the log splits the replay from the recorded trail —
// a silent callback bus would hide this.
func TestReplay_AlteredCommandDiverges(t *testing.T) {
	// given
	final, path := recordDecisionSession(t.Context(), t, 12)

	// when: the stored StickPush rows gain a payload delta
	tamperCommandRows(t, path,
		"UPDATE tick_commands SET a = a + 1 WHERE kind = ?",
		string(kernel.KindStickPush),
	)

	// then: persist.Replay names the first tick whose checksum disagrees
	result := replayDecisionLog(t.Context(), t, path)
	if result.Mismatch == nil {
		t.Fatal("altered log replayed without a mismatch, want the stored trail to diverge")
	}
	if result.Final == final {
		t.Fatal("altered log folded to the recorded hash, want divergence")
	}
}

// replayDecisionLog opens the recorded store afresh and runs persist.Replay
// over it: the shipped per-tick submit+pass loop, which compares every
// replayed tick's checksum against the stored trail and names the first
// disagreement (internal/persist/ticklog.go Replay).
func replayDecisionLog(ctx context.Context, t *testing.T, path string) persist.ReplayResult {
	t.Helper()

	result, err := persist.Replay(ctx, replayLog(ctx, t, path), bridgeRig(7), nil)
	if err != nil {
		t.Fatalf("persist.Replay: %v", err)
	}

	return result
}

// replayLog reopens the store at path (the replay leg must read the persisted
// rows, not the writer's memory).
func replayLog(ctx context.Context, t *testing.T, path string) *persist.SQLiteStore {
	t.Helper()

	store, err := persist.OpenSQLiteStore(ctx, path)
	if err != nil {
		t.Fatalf("inputbridge: open replay store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	return store
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
// decision tape, records every pass into the writer, and returns the folded
// trail hash the live run produced (FoldChecksums over the per-tick checksums
// — the value persist.Replay compares).
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

	trail := make([]persist.TickChecksum, 0, ticks)

	for tick := range ticks {
		playDecisionTape(session, tick)
		session.Tick(uint32(tick))

		rec := persist.TickRecord{
			Tick:     uint32(tick),
			Commands: k.TraceCommandsAt(uint32(tick)),
			Events:   k.TraceEventsAt(uint32(tick)),
			Checksum: k.TraceChecksumAt(uint32(tick)),
		}
		trail = append(trail, persist.TickChecksum{Tick: rec.Tick, Hash: rec.Checksum})
		if err := writer.Record(ctx, rec); err != nil {
			t.Fatalf("Record tick %d: %v", tick, err)
		}
	}

	if err := writer.Flush(ctx); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	return persist.FoldChecksums(trail), path
}

// tamperCommandRows opens the recorded store's SQLite file directly and
// applies one statement — the desync probe (the same hand-written-rows pattern
// as persist's openSchemaDB, internal/persist/ticklog_test.go).
func tamperCommandRows(t *testing.T, path, stmt string, args ...any) {
	t.Helper()

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open store for tamper: %v", err)
	}
	defer func() { _ = db.Close() }()

	ctx := t.Context()
	if _, err := db.ExecContext(ctx, stmt, args...); err != nil {
		t.Fatalf("tamper %q: %v", stmt, err)
	}
}
