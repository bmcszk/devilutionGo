package persist

import (
	"context"
	"encoding/binary"
	"fmt"
	"hash/fnv"

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
//
// Ticks buffer until the flush batch is full, so a live game pays one
// transaction per batch instead of one fsync per tick; Flush writes the tail.
type TickLogWriter struct {
	store   *SQLiteStore
	opts    TickLogOptions
	pending []pendingTick
}

// pendingTick is one buffered tick plus the snapshot it triggered, captured when
// the tick was recorded rather than when the batch flushes.
type pendingTick struct {
	rec      TickRecord
	snapshot *Snapshot
}

// TickLogOptions configures the writer.
type TickLogOptions struct {
	// SnapshotEvery is the snapshot cadence in ticks; 0 means
	// DefaultSnapshotEvery, a negative value disables snapshots.
	SnapshotEvery int
	// Batch is how many ticks one transaction writes; 0 or 1 writes every tick
	// on its own.
	Batch int
	// Snapshot produces the opaque state image stored with a snapshot. Nil
	// records the checksum only, which is what replay validates.
	Snapshot func() []byte
}

// ReplayResult is what a replay reproduced: the per-tick checksum trail, the
// final state hash, and the tick where a stored checksum disagreed.
type ReplayResult struct {
	// Trail is the checksum trail the replay produced, one entry per tick.
	Trail []TickChecksum
	// Final is FoldChecksums over the trail: the run's final state hash.
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

// Record buffers one tick. Call it after the kernel pass for that tick, with the
// pass's commands and the events the pass returned; the checksum is
// kernel.TraceHash over those events, never a clock and never a re-simulation.
func (w *TickLogWriter) Record(ctx context.Context, rec TickRecord) error {
	pending := pendingTick{rec: rec}
	if w.snapshotDue(rec.Tick) {
		snapshot := w.snapshot(rec)
		pending.snapshot = &snapshot
	}

	w.pending = append(w.pending, pending)

	if len(w.pending) < w.batch() {
		return nil
	}

	return w.Flush(ctx)
}

// Flush writes the buffered ticks and their snapshots in one transaction.
func (w *TickLogWriter) Flush(ctx context.Context) error {
	if len(w.pending) == 0 {
		return nil
	}

	ticks := make([]TickRecord, 0, len(w.pending))
	snapshots := make([]Snapshot, 0, len(w.pending))

	for _, pending := range w.pending {
		ticks = append(ticks, pending.rec)
		if pending.snapshot != nil {
			snapshots = append(snapshots, *pending.snapshot)
		}
	}

	w.pending = nil

	return w.store.AppendTicks(ctx, ticks, snapshots)
}

// batch is the flush batch size, at least 1.
func (w *TickLogWriter) batch() int {
	return max(w.opts.Batch, 1)
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

// Replay drives k with the recorded commands and compares the per-tick
// checksums it produces with the ones the log stored.
//
// The caller supplies the kernel because the handlers are game state: replay
// determinism means the same handlers, the same RNG seed and the same commands
// reproduce the recorded hashes (ADR-015), and a snapshot replay additionally
// needs the state the snapshot image holds.
//
// A nil snapshot replays from the beginning; otherwise replay starts at the tick
// after the snapshot and validates only the ticks it ran.
func Replay(ctx context.Context, store *SQLiteStore, k *kernel.Kernel, snapshot *Snapshot) (ReplayResult, error) {
	log, err := store.LoadTickLog(ctx)
	if err != nil {
		return ReplayResult{}, err
	}

	result, err := replayLog(log, k, replayFrom(snapshot))
	if err != nil {
		return ReplayResult{}, err
	}

	// Compare only the ticks this replay covered: a snapshot replay validates its
	// own range, not the ticks it never ran.
	stored := checksumsFrom(logChecksums(log), replayFrom(snapshot))

	if ok, div := CompareTrails(stored, result.Trail); !ok {
		result.Mismatch = &div
	}

	return result, nil
}

// FoldChecksums folds a checksum trail into one run hash: FNV-1a 64 over each
// tick and its event checksum, in tick order. Replay reproduces the trail, so it
// reproduces this final state hash too.
func FoldChecksums(trail []TickChecksum) uint64 {
	h := fnv.New64a()

	var tick [4]byte
	var hash [8]byte

	for _, entry := range trail {
		binary.LittleEndian.PutUint32(tick[:], entry.Tick)
		binary.LittleEndian.PutUint64(hash[:], entry.Hash)
		_, _ = h.Write(tick[:])
		_, _ = h.Write(hash[:])
	}

	return h.Sum64()
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

// replayFrom is the first tick a replay runs: 0 from scratch, otherwise the
// tick after the snapshot.
func replayFrom(snapshot *Snapshot) uint32 {
	if snapshot == nil {
		return 0
	}

	return snapshot.Tick + 1
}

// String names the divergence the way a desync report does: tick and both
// checksums.
func (d Divergence) String() string {
	return fmt.Sprintf("divergence at tick %d: left=%016x right=%016x", d.Tick, d.Left, d.Right)
}

// replayLog runs the ticks of a log from `from` on.
func replayLog(log []TickRecord, k *kernel.Kernel, from uint32) (ReplayResult, error) {
	result := ReplayResult{Trail: make([]TickChecksum, 0, len(log))}

	if err := replayTicks(log, k, from, &result); err != nil {
		return ReplayResult{}, err
	}

	result.Final = FoldChecksums(result.Trail)

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

// replayTick submits one recorded tick's commands in order, runs the pass and
// returns the checksum of the events that pass emitted.
func replayTick(k *kernel.Kernel, rec TickRecord) (uint64, error) {
	for _, cmd := range rec.Commands {
		if !k.Submit(cmd) {
			return 0, fmt.Errorf("persist: replay tick %d: command %q rejected by the gate: %w",
				rec.Tick, cmd.Kind, ErrTickLogReplay)
		}
	}

	return kernel.TraceHash(k.Pass(rec.Tick)), nil
}

// logChecksums projects a log's stored per-tick checksums.
func logChecksums(log []TickRecord) []TickChecksum {
	trail := make([]TickChecksum, 0, len(log))
	for _, rec := range log {
		trail = append(trail, TickChecksum{Tick: rec.Tick, Hash: rec.Checksum})
	}

	return trail
}

// checksumsFrom drops the trail entries before tick.
func checksumsFrom(trail []TickChecksum, tick uint32) []TickChecksum {
	from := 0
	for from < len(trail) && trail[from].Tick < tick {
		from++
	}

	return trail[from:]
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
