package persist_test

import (
	"context"
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game/kernel"
	"github.com/bmcszk/devilutionGo/internal/persist"
)

// fnvOffset and fnvPrime are the FNV-1a 64 constants the world hash folds with.
const (
	fnvOffset uint64 = 14695981039346656037
	fnvPrime  uint64 = 1099511628211
)

// stateImageSize is how many bytes a synthetic snapshot image holds.
const stateImageSize = 16

// syntheticWorld is the tick-log test state: a kernel whose handler mutates this
// world from commands only, so replay determinism is a property of the log and
// not of the fixture.
type syntheticWorld struct {
	moves int32
	hash  uint64
}

// encode is the snapshot state image: the world as the replay needs it back.
func (w *syntheticWorld) encode() []byte {
	buf := make([]byte, stateImageSize)
	binary.LittleEndian.PutUint32(buf[:4], uint32(w.moves))
	binary.LittleEndian.PutUint64(buf[8:], w.hash)

	return buf
}

func TestTickLogAppendLoad(t *testing.T) {
	// given a table of tick records to append
	ctx := context.Background()

	cases := []struct {
		name    string
		records []persist.TickRecord
	}{
		{
			name: "one tick with one command and one event",
			records: []persist.TickRecord{{
				Tick:     0,
				Commands: []kernel.Command{{Tick: 0, Kind: kernel.KindKey, A: 3, Down: true}},
				Events:   []kernel.Event{{Tick: 0, Kind: kernel.KindPlayerMoved, A: 3, Stream: "combat", Roll: 9}},
				Checksum: 0x0102,
			}},
		},
		{
			name: "a tick that passed with no command",
			records: []persist.TickRecord{
				{Tick: 0, Checksum: 1},
				{Tick: 1, Checksum: 2},
			},
		},
		{
			name: "ordered events within one tick",
			records: []persist.TickRecord{{
				Tick: 5,
				Commands: []kernel.Command{
					{Tick: 5, Kind: kernel.KindPlayerMove, A: 1},
					{Tick: 5, Kind: kernel.KindPlayerAttack, A: 2},
				},
				Events: []kernel.Event{
					{Tick: 5, Kind: kernel.KindPlayerMoved, A: 1},
					{Tick: 5, Kind: kernel.KindPlayerStruck, A: 2},
					{Tick: 5, Kind: kernel.KindMonsterDied, A: 2, B: 1},
				},
				Checksum: 0xffffffffffffffff,
			}},
		},
	}

	// when each record set is appended to a fresh store and loaded back
	for _, tc := range cases {
		// then every tick round-trips with its payload in order
		t.Run(tc.name, func(t *testing.T) {
			appendLoadCase(ctx, t, tc.records)
		})
	}
}

func TestTickLogAppendTwiceRejectsDuplicateTick(t *testing.T) {
	// given a store that already recorded tick 4
	ctx := context.Background()
	store := openStore(ctx, t)

	if err := store.AppendTick(ctx, persist.TickRecord{Tick: 4, Checksum: 7}); err != nil {
		t.Fatalf("AppendTick: %v", err)
	}

	// when the same tick is appended again
	err := store.AppendTick(ctx, persist.TickRecord{Tick: 4, Checksum: 8})
	// then the duplicate is refused instead of rewriting the hash trail
	requireAppendRejected(t, err)
}

func TestTickLogReplayTenThousandTicks(t *testing.T) {
	// given a 10,000-tick synthetic run recorded into the store
	ctx := context.Background()
	store := openStore(ctx, t)

	wantHash := recordRun(ctx, t, store, 10000)

	// when the log is replayed into a fresh kernel
	k, world := newSyntheticRig()

	result, err := persist.Replay(ctx, store, k, nil)
	// then replay reproduces the recorded checksum trail, the final state hash
	// and the world the recorded run ended in
	if err != nil {
		t.Fatalf("Replay: %v", err)
	}

	if result.Mismatch != nil {
		t.Fatalf("replay diverged: %s", result.Mismatch)
	}

	if len(result.Trail) != 10000 {
		t.Fatalf("trail length = %d, want 10000", len(result.Trail))
	}

	if result.Final != persist.FoldChecksums(result.Trail) {
		t.Fatalf("final = %016x, want the fold of the trail", result.Final)
	}

	if world.hash != wantHash {
		t.Fatalf("replayed world hash = %016x, want %016x", world.hash, wantHash)
	}
}

// TestTickLogReplayFromSnapshot checks the load path ADR-018 law 4 names:
// snapshot plus replay after it reproduces the same ticks as a full replay.
func TestTickLogReplayFromSnapshot(t *testing.T) {
	// given a recorded run with a snapshot cadence
	ctx := context.Background()
	store := openStore(ctx, t)
	ticks := 1000

	recordRun(ctx, t, store, ticks)

	// when the last snapshot seeds a replay
	snapshots, err := store.LoadSnapshots(ctx)
	if err != nil {
		t.Fatalf("LoadSnapshots: %v", err)
	}

	full, err := persist.Replay(ctx, store, mustRig(), nil)
	if err != nil {
		t.Fatalf("Replay full: %v", err)
	}

	fromSnapshot, err := persist.Replay(
		ctx,
		store,
		mustRigFromState(snapshots[len(snapshots)-1].State),
		&snapshots[len(snapshots)-1],
	)
	if err != nil {
		t.Fatalf("Replay from snapshot: %v", err)
	}

	// then replay starts at the snapshot instead of tick 0 and reproduces the
	// trail of the ticks after it
	requireSnapshotCadence(t, snapshots, ticks/250)
	requireSnapshotTrailMatchesTail(t, full, fromSnapshot)
}

func TestTickLogReplayNamesMismatchedTick(t *testing.T) {
	// given a log that recorded a wrong checksum at tick 2, as a nondeterministic
	// run would
	ctx := context.Background()
	store := openStore(ctx, t)

	recordCorruptedRun(ctx, t, store, 5, 2)

	// when the log is replayed
	result, err := persist.Replay(ctx, store, mustRig(), nil)
	// then the first disagreeing tick is named
	if err != nil {
		t.Fatalf("Replay: %v", err)
	}

	if result.Mismatch == nil || result.Mismatch.Tick != 2 {
		t.Fatalf("mismatch = %v, want tick 2", result.Mismatch)
	}
}

func TestCompareLogsDivergence(t *testing.T) {
	// given a table of run lengths and the tick the right run diverges at
	ctx := context.Background()

	cases := []struct {
		name      string
		leftTicks int
		// divergeAt is the tick the right run changes its command at; negative
		// means the right run replays the left one exactly.
		divergeAt  int
		rightTicks int
		wantOK     bool
		wantTick   uint32
	}{
		{name: "identical logs", leftTicks: 8, divergeAt: -1, rightTicks: 8, wantOK: true},
		{name: "divergence at the first tick", leftTicks: 8, divergeAt: 0, rightTicks: 8, wantTick: 0},
		{name: "divergence mid run", leftTicks: 8, divergeAt: 5, rightTicks: 8, wantTick: 5},
		{
			name:       "truncated log diverges at the first tick it is missing",
			leftTicks:  8,
			divergeAt:  -1,
			rightTicks: 6,
			wantTick:   6,
		},
	}

	// when each pair of runs is recorded and compared
	for _, tc := range cases {
		// then the reported divergence matches the case
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantOK {
				compareLogsAgreement(ctx, t, tc.leftTicks, tc.divergeAt, tc.rightTicks)

				return
			}

			diff := compareLogsDivergence(ctx, t, tc.leftTicks, tc.divergeAt, tc.rightTicks)
			requireDivergenceAt(t, diff, tc.wantTick)
		})
	}
}

// newSyntheticRig returns a kernel wired to a fresh world. The handler folds each
// move into the world state and emits one event carrying the new state, so the
// trace and the world hash move together.
func newSyntheticRig() (*kernel.Kernel, *syntheticWorld) {
	return newSeededRig(nil)
}

// newSeededRig returns a kernel whose world starts at the snapshot image, which
// is how a load resumes: snapshot plus replay of the ticks after it.
func newSeededRig(state []byte) (*kernel.Kernel, *syntheticWorld) {
	world := &syntheticWorld{hash: fnvOffset}
	if len(state) == stateImageSize {
		world.moves = int32(binary.LittleEndian.Uint32(state[:4]))
		world.hash = binary.LittleEndian.Uint64(state[8:])
	}

	k := kernel.New(nil, nil)
	k.Register(kernel.KindPlayerMove, func(c kernel.Command, _ *kernel.Streams) []kernel.Event {
		world.moves += c.A
		world.hash = (world.hash ^ uint64(uint32(world.moves))) * fnvPrime

		return []kernel.Event{{Kind: kernel.KindPlayerMoved, A: world.moves, B: c.B}}
	})

	return k, world
}

// recordRun drives the rig for ticks ticks, records every pass into the writer,
// and returns the final world hash.
func recordRun(ctx context.Context, t *testing.T, store *persist.SQLiteStore, ticks int) uint64 {
	t.Helper()

	k, world := newSyntheticRig()
	writer := store.NewTickLogWriter(persist.TickLogOptions{
		SnapshotEvery: 250, Batch: 250,
		Snapshot: world.encode,
	})

	for tick := range ticks {
		cmd := kernel.Command{Tick: uint32(tick), Kind: kernel.KindPlayerMove, A: int32(tick%7 + 1), B: 2}
		if !k.Submit(cmd) {
			t.Fatalf("tick %d: gate rejected the command", tick)
		}

		events := k.Pass(uint32(tick))

		rec := persist.TickRecord{
			Tick: uint32(tick), Commands: []kernel.Command{cmd},
			Events: events, Checksum: kernel.TraceHash(events),
		}
		if err := writer.Record(ctx, rec); err != nil {
			t.Fatalf("Record tick %d: %v", tick, err)
		}
	}

	if err := writer.Flush(ctx); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	return world.hash
}

// recordDivergingRun records ticks ticks where the tick divergeAt carries a
// different command, so the checksum trail splits there.
func recordDivergingRun(ctx context.Context, t *testing.T, store *persist.SQLiteStore, ticks, divergeAt int) {
	t.Helper()

	k, _ := newSyntheticRig()
	writer := store.NewTickLogWriter(persist.TickLogOptions{SnapshotEvery: -1, Batch: 4})

	for tick := range ticks {
		payload := int32(tick%7 + 1)
		if tick == divergeAt {
			payload += 100
		}

		cmd := kernel.Command{Tick: uint32(tick), Kind: kernel.KindPlayerMove, A: payload, B: 2}
		k.Submit(cmd)

		events := k.Pass(uint32(tick))

		rec := persist.TickRecord{
			Tick: uint32(tick), Commands: []kernel.Command{cmd},
			Events: events, Checksum: kernel.TraceHash(events),
		}
		if err := writer.Record(ctx, rec); err != nil {
			t.Fatalf("Record tick %d: %v", tick, err)
		}
	}

	if err := writer.Flush(ctx); err != nil {
		t.Fatalf("Flush: %v", err)
	}
}

// recordCorruptedRun records ticks ticks but flips the checksum of corruptedTick.
func recordCorruptedRun(ctx context.Context, t *testing.T, store *persist.SQLiteStore, ticks, corruptedTick int) {
	t.Helper()

	k, _ := newSyntheticRig()

	for tick := range ticks {
		cmd := kernel.Command{Tick: uint32(tick), Kind: kernel.KindPlayerMove, A: 1}
		k.Submit(cmd)

		checksum := kernel.TraceHash(eventsForTick(k, uint32(tick)))
		if tick == corruptedTick {
			checksum ^= 0xdead
		}

		rec := persist.TickRecord{
			Tick: uint32(tick), Commands: []kernel.Command{cmd},
			Events: k.Pass(uint32(tick)), Checksum: checksum,
		}
		if err := store.AppendTick(ctx, rec); err != nil {
			t.Fatalf("AppendTick tick %d: %v", tick, err)
		}
	}
}

// eventsForTick runs the kernel's pass for tick and returns what it emitted.
func eventsForTick(k *kernel.Kernel, tick uint32) []kernel.Event {
	return k.Pass(tick)
}

// appendLoadCase appends records to a fresh store and loads the log back.
func appendLoadCase(ctx context.Context, t *testing.T, records []persist.TickRecord) {
	t.Helper()

	store := openStore(ctx, t)

	for _, rec := range records {
		if err := store.AppendTick(ctx, rec); err != nil {
			t.Fatalf("AppendTick: %v", err)
		}
	}

	got, err := store.LoadTickLog(ctx)
	if err != nil {
		t.Fatalf("LoadTickLog: %v", err)
	}

	requireTickRecords(t, got, records)
}

// compareLogsRun records a left and a right run and compares the two logs.
func compareLogsRun(
	ctx context.Context,
	t *testing.T,
	leftTicks, divergeAt, rightTicks int,
) (persist.TickLogDiff, bool) {
	t.Helper()

	// given two runs of the same synthetic world
	left, right := openStore(ctx, t), openStore(ctx, t)
	recordRun(ctx, t, left, leftTicks)
	recordDivergingRun(ctx, t, right, rightTicks, divergeAt)

	// when the two logs are compared
	diff, ok, err := persist.CompareLogs(ctx, left, right)
	if err != nil {
		t.Fatalf("CompareLogs: %v", err)
	}

	return diff, ok
}

// compareLogsAgreement records two matching runs and pins that CompareLogs agrees.
func compareLogsAgreement(ctx context.Context, t *testing.T, leftTicks, divergeAt, rightTicks int) {
	t.Helper()

	diff, ok := compareLogsRun(ctx, t, leftTicks, divergeAt, rightTicks)
	if !ok {
		t.Fatalf("ok = false, want true (diff %s)", diff)
	}
}

// compareLogsDivergence records two differing runs and returns the reported diff.
func compareLogsDivergence(
	ctx context.Context,
	t *testing.T,
	leftTicks, divergeAt, rightTicks int,
) persist.TickLogDiff {
	t.Helper()

	diff, ok := compareLogsRun(ctx, t, leftTicks, divergeAt, rightTicks)
	if ok {
		t.Fatalf("ok = true, want a divergence (diff %s)", diff)
	}

	return diff
}

// requireDivergenceAt pins the first diverging tick and both checksums.
func requireDivergenceAt(t *testing.T, diff persist.TickLogDiff, wantTick uint32) {
	t.Helper()

	if diff.Tick != wantTick {
		t.Fatalf("diverging tick = %d, want %d (%s)", diff.Tick, wantTick, diff)
	}

	if diff.Left == diff.Right && diff.LeftTicks == diff.RightTicks {
		t.Fatalf("checked equal checksums reported as divergent: %s", diff)
	}

	requireDivergenceText(t, diff.Divergence)
}

// requireAppendRejected pins that a duplicate tick append is refused.
func requireAppendRejected(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Fatal("second AppendTick for tick 4 succeeded, want an error")
	}
}

// requireSnapshotCadence checks the recorded snapshot count.
func requireSnapshotCadence(t *testing.T, snapshots []persist.Snapshot, want int) {
	t.Helper()

	if len(snapshots) != want {
		t.Fatalf("snapshots = %d, want %d", len(snapshots), want)
	}
}

// requireSnapshotTrailMatchesTail checks the snapshot replay reproduces exactly
// the ticks the full replay ran after the snapshot.
func requireSnapshotTrailMatchesTail(t *testing.T, full, fromSnapshot persist.ReplayResult) {
	t.Helper()

	if fromSnapshot.Mismatch != nil {
		t.Fatalf("snapshot replay diverged: %s", fromSnapshot.Mismatch)
	}

	if len(fromSnapshot.Trail) >= len(full.Trail) {
		t.Fatalf("snapshot replay read %d ticks, want fewer than the full %d",
			len(fromSnapshot.Trail), len(full.Trail))
	}

	tail := full.Trail[len(full.Trail)-len(fromSnapshot.Trail):]
	for i := range tail {
		if fromSnapshot.Trail[i] != tail[i] {
			t.Fatalf("snapshot trail[%d] = %+v, want %+v", i, fromSnapshot.Trail[i], tail[i])
		}
	}
}

// requireDivergenceText pins the report format a desync log quotes.
func requireDivergenceText(t *testing.T, diff persist.Divergence) {
	t.Helper()

	want := fmt.Sprintf("divergence at tick %d: left=%016x right=%016x", diff.Tick, diff.Left, diff.Right)
	if got := diff.String(); got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
}

// requireTickRecords compares a loaded log with the appended records.
func requireTickRecords(t *testing.T, got, want []persist.TickRecord) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("loaded %d ticks, want %d", len(got), len(want))
	}

	for i := range want {
		requireTickRecord(t, i, got[i], want[i])
	}
}

// requireTickRecord pins one tick's header and payload.
func requireTickRecord(t *testing.T, i int, got, want persist.TickRecord) {
	t.Helper()

	if got.Tick != want.Tick || got.Checksum != want.Checksum {
		t.Fatalf("tick %d = (%d, %016x), want (%d, %016x)", i, got.Tick, got.Checksum, want.Tick, want.Checksum)
	}

	requireTickPayload(t, i, got, want)
}

// requireTickPayload pins one tick's command and event slices.
func requireTickPayload(t *testing.T, i int, got, want persist.TickRecord) {
	t.Helper()

	if len(got.Commands) != len(want.Commands) || len(got.Events) != len(want.Events) {
		t.Fatalf("tick %d payload sizes = (%d commands, %d events), want (%d, %d)", i,
			len(got.Commands), len(got.Events), len(want.Commands), len(want.Events))
	}

	requireTickCommands(t, i, got.Commands, want.Commands)
	requireTickEvents(t, i, got.Events, want.Events)
}

// requireTickCommands pins one tick's commands against the recorded ones.
func requireTickCommands(t *testing.T, i int, got, want []kernel.Command) {
	t.Helper()

	for j := range want {
		if got[j] != want[j] {
			t.Fatalf("tick %d command %d = %+v, want %+v", i, j, got[j], want[j])
		}
	}
}

// requireTickEvents pins one tick's events against the recorded ones.
func requireTickEvents(t *testing.T, i int, got, want []kernel.Event) {
	t.Helper()

	for j := range want {
		if got[j] != want[j] {
			t.Fatalf("tick %d event %d = %+v, want %+v", i, j, got[j], want[j])
		}
	}
}

// mustRig returns a fresh kernel for a replay the test does not inspect.
func mustRig() *kernel.Kernel {
	k, _ := newSyntheticRig()

	return k
}

// mustRigFromState returns a kernel whose world was restored from a snapshot.
func mustRigFromState(state []byte) *kernel.Kernel {
	k, _ := newSeededRig(state)

	return k
}
