package persist

import (
	"context"
	"fmt"

	"github.com/bmcszk/devilutionGo/internal/game/kernel"
)

// DefaultSnapshotEvery is the snapshot cadence ADR-018 law 4 names: the kernel
// snapshots every 1,000 ticks, and the cadence is tunable per game.
const DefaultSnapshotEvery = 1000

// TickLogWriter observes a running kernel and writes its ticks into the store's
// tick-log column family (ADR-018 law 4). It is a pure observer: it reads the
// commands a pass was given, the events it emitted and its checksum, and never
// touches gameplay state, so recording can be switched off without changing the
// simulation.
type TickLogWriter struct {
	store *SQLiteStore
	opts  TickLogOptions
}

// TickLogOptions configures the writer.
type TickLogOptions struct {
	// SnapshotEvery is the snapshot cadence in ticks; 0 means
	// DefaultSnapshotEvery, a negative value disables snapshots.
	SnapshotEvery int
	// Snapshot produces the opaque state image stored with a snapshot. Nil
	// records the checksum only, which is what replay validates.
	Snapshot func() []byte
}

// ReplayResult is what a replay reproduced: the per-tick checksum trail, the
// final checksum, and the tick where a stored checksum disagreed.
type ReplayResult struct {
	// Trail is the checksum trail the replay produced, one entry per tick.
	Trail []TickChecksum
	// Final is the last checksum of the trail.
	Final uint64
	// Mismatch names the first tick whose stored checksum differs from the
	// replayed one, nil when the replay agreed everywhere.
	Mismatch *Divergence
}

// Divergence is the first tick two runs disagree on: both checksums and the
// event evidence each side holds for that tick.
type Divergence struct {
	// Tick is the first tick whose checksums differ.
	Tick uint32
	// Left and Right are the checksums the two runs recorded at Tick.
	Left, Right uint64
	// LeftEvents and RightEvents are the events each log recorded at Tick, so
	// the divergence can be read without replaying.
	LeftEvents, RightEvents []kernel.Event
}

// TickLogDiff is the divergence between two recorded logs, with the tick counts
// so a truncated log is not mistaken for an equal one.
type TickLogDiff struct {
	// Divergence is the first tick the two logs disagree on.
	Divergence
	// LeftTicks and RightTicks are how many ticks each log recorded.
	LeftTicks, RightTicks int
}

// NewTickLogWriter returns a writer appending into store.
func (s *SQLiteStore) NewTickLogWriter(opts TickLogOptions) *TickLogWriter {
	if opts.SnapshotEvery == 0 {
		opts.SnapshotEvery = DefaultSnapshotEvery
	}

	return &TickLogWriter{store: s, opts: opts}
}

// Record appends one tick. Call it after the kernel pass for that tick, with the
// pass's commands and the events the pass returned; the checksum is the kernel's
// own cumulative trace checksum, never recomputed here.
func (w *TickLogWriter) Record(ctx context.Context, rec TickRecord) error {
	if err := w.store.AppendTick(ctx, rec); err != nil {
		return err
	}

	if !w.snapshotDue(rec.Tick) {
		return nil
	}

	return w.store.AppendSnapshot(ctx, w.snapshot(rec))
}

// snapshotDue reports whether tick lands on the snapshot cadence.
func (w *TickLogWriter) snapshotDue(tick uint32) bool {
	if w.opts.SnapshotEvery < 0 {
		return false
	}

	return tick%uint32(w.opts.SnapshotEvery) == 0
}

// snapshot builds the snapshot row for a cadence tick.
func (w *TickLogWriter) snapshot(rec TickRecord) Snapshot {
	var state []byte
	if w.opts.Snapshot != nil {
		state = w.opts.Snapshot()
	}

	return Snapshot{Tick: rec.Tick, Checksum: rec.Checksum, State: state}
}

// Replay drives k with the recorded commands and compares the checksums it
// produces with the ones the log stored.
//
// The caller supplies the kernel because the handlers are game state: replay
// determinism means the same handlers, the same RNG seed and the same commands
// reproduce the recorded hashes (ADR-015).
//
// A nil snapshot replays from the beginning; otherwise replay starts after the
// snapshot tick, whose stored checksum seeds the comparison.
func Replay(ctx context.Context, store *SQLiteStore, k *kernel.Kernel, snapshot *Snapshot) (ReplayResult, error) {
	log, err := store.LoadTickLog(ctx)
	if err != nil {
		return ReplayResult{}, err
	}

	result, err := replayLog(log, k, snapshot)
	if err != nil {
		return ReplayResult{}, err
	}

	if ok, div := CompareTrails(logChecksums(log), result.Trail); !ok {
		result.Mismatch = &div
	}

	return result, nil
}

// CompareTrails returns the first tick whose checksums differ. It reports true
// when the two trails agree on every shared tick and have the same length; a
// trail that is a strict prefix of the other diverges at the first missing tick,
// with the missing side's checksum zero.
func CompareTrails(left, right []TickChecksum) (bool, Divergence) {
	for i := range max(len(left), len(right)) {
		l, r := at(left, i), at(right, i)
		if l.Tick != r.Tick || l.Hash != r.Hash {
			// Name a tick that exists: when one trail ends early, the divergence
			// is at the last tick the longer trail recorded.
			return false, Divergence{Tick: max(l.Tick, r.Tick), Left: l.Hash, Right: r.Hash}
		}
	}

	return true, Divergence{}
}

// CompareLogs finds the first tick two recorded logs disagree on, with the
// events of that tick from both sides. It reads only the checksums and that one
// tick's events: desync triage names one tick, not a full diff.
func CompareLogs(ctx context.Context, left, right *SQLiteStore) (TickLogDiff, bool, error) {
	leftLog, err := left.LoadTickLog(ctx)
	if err != nil {
		return TickLogDiff{}, false, err
	}

	rightLog, err := right.LoadTickLog(ctx)
	if err != nil {
		return TickLogDiff{}, false, err
	}

	ok, div := CompareTrails(logChecksums(leftLog), logChecksums(rightLog))
	if !ok {
		div.LeftEvents, div.RightEvents = eventsAt(leftLog, div.Tick), eventsAt(rightLog, div.Tick)
	}

	return TickLogDiff{Divergence: div, LeftTicks: len(leftLog), RightTicks: len(rightLog)}, ok, nil
}

// String names the divergence the way a desync report does: tick and both
// checksums.
func (d Divergence) String() string {
	return fmt.Sprintf("divergence at tick %d: left=%016x right=%016x", d.Tick, d.Left, d.Right)
}

// replayLog runs the ticks of a log, seeded by an optional snapshot.
func replayLog(log []TickRecord, k *kernel.Kernel, snapshot *Snapshot) (ReplayResult, error) {
	from := uint32(0)
	result := ReplayResult{Trail: make([]TickChecksum, 0, len(log))}

	if snapshot != nil {
		from = snapshot.Tick + 1
		result.Trail = append(result.Trail, TickChecksum{Tick: snapshot.Tick, Hash: snapshot.Checksum})
	}

	if err := replayTicks(log, k, from, &result); err != nil {
		return ReplayResult{}, err
	}

	if len(result.Trail) > 0 {
		result.Final = result.Trail[len(result.Trail)-1].Hash
	}

	return result, nil
}

// replayTicks replays every recorded tick from `from` on, appending its hash.
func replayTicks(log []TickRecord, k *kernel.Kernel, from uint32, result *ReplayResult) error {
	for _, rec := range log {
		if rec.Tick < from {
			continue
		}

		hash, err := replayTick(k, rec)
		if err != nil {
			return err
		}

		result.Trail = append(result.Trail, TickChecksum{Tick: rec.Tick, Hash: hash})
	}

	return nil
}

// replayTick submits one recorded tick's commands in order and runs the pass.
func replayTick(k *kernel.Kernel, rec TickRecord) (uint64, error) {
	for _, cmd := range rec.Commands {
		if !k.Submit(cmd) {
			return 0, fmt.Errorf("persist: replay tick %d: command %q rejected by the gate: %w",
				rec.Tick, cmd.Kind, ErrTickLogReplay)
		}
	}

	k.Pass(rec.Tick)

	return k.TraceHash(), nil
}

// logChecksums projects a log's stored per-tick checksums.
func logChecksums(log []TickRecord) []TickChecksum {
	trail := make([]TickChecksum, 0, len(log))
	for _, rec := range log {
		trail = append(trail, TickChecksum{Tick: rec.Tick, Hash: rec.Checksum})
	}

	return trail
}

// at returns trail[i], or a zero entry when the trail is shorter, so a missing
// tick compares as a divergence instead of an index panic.
func at(trail []TickChecksum, i int) TickChecksum {
	if i < len(trail) {
		return trail[i]
	}

	return TickChecksum{}
}

// eventsAt returns the events one log recorded at tick.
func eventsAt(log []TickRecord, tick uint32) []kernel.Event {
	for _, rec := range log {
		if rec.Tick == tick {
			return rec.Events
		}
	}

	return nil
}
