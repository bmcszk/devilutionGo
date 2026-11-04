package persist

import (
	"context"
	"fmt"

	"github.com/bmcszk/devilutionGo/internal/game/kernel"
	"github.com/bmcszk/devilutionGo/internal/persist/sqlc"
)

// TickRecord is one recorded tick: the commands the kernel was given (the
// replay input), the events it emitted (the state changes a desync hunt
// compares), and the checksum of those events — one per tick, not cumulative, so
// a replay can start at a snapshot and still check its own ticks.
type TickRecord struct {
	// Tick is the pass tick; it comes from the kernel, never from a clock.
	Tick uint32
	// Commands are the commands dispatched in that pass, in submission order.
	Commands []kernel.Command
	// Events are the events the pass emitted, in trace order.
	Events []kernel.Event
	// Checksum is kernel.TraceHash over this tick's events.
	Checksum uint64
}

// TickChecksum is one tick's event checksum, the hash trail the divergence
// detector bisects.
type TickChecksum struct {
	// Tick is the tick the checksum was taken at.
	Tick uint32
	// Hash is the checksum of that tick's events.
	Hash uint64
}

// Snapshot is one snapshot-rotation row: the checksum at a cadence tick plus the
// opaque state image the game's snapshotter produced there.
type Snapshot struct {
	// Tick is the cadence tick the snapshot was taken at.
	Tick uint32
	// Checksum is the trace checksum at that tick.
	Checksum uint64
	// State is the opaque state image; nil when the game records none.
	State []byte
}

// AppendTick writes one tick of the log: its commands, its events and its
// checksum, atomically.
func (s *SQLiteStore) AppendTick(ctx context.Context, rec TickRecord) error {
	return s.AppendTicks(ctx, []TickRecord{rec}, nil)
}

// AppendTicks writes a batch of ticks and the snapshots they produced in one
// transaction, so a live game can record many ticks per fsync while a tick batch
// stays all-or-nothing. A tick already in the log is not rewritten: the checksum
// is the tick's primary key, so a duplicate insert fails the whole batch instead
// of silently corrupting the hash trail.
func (s *SQLiteStore) AppendTicks(ctx context.Context, ticks []TickRecord, snapshots []Snapshot) error {
	if len(ticks) == 0 && len(snapshots) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("persist: begin tick log batch: %w", err)
	}

	if err := insertTickBatch(ctx, s.queries.WithTx(tx), ticks, snapshots); err != nil {
		_ = tx.Rollback()

		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("persist: commit tick log batch: %w", err)
	}

	return nil
}

// insertTickBatch writes every tick, then every snapshot, inside one tx.
func insertTickBatch(ctx context.Context, q *sqlc.Queries, ticks []TickRecord, snapshots []Snapshot) error {
	for _, rec := range ticks {
		if err := insertTickRows(ctx, q, rec); err != nil {
			return err
		}
	}

	for _, snapshot := range snapshots {
		if err := q.InsertTickSnapshot(ctx, snapshotToParams(snapshot)); err != nil {
			return fmt.Errorf("persist: append snapshot tick %d: %w", snapshot.Tick, err)
		}
	}

	return nil
}

// insertTickRows writes one tick's commands, events and checksum.
func insertTickRows(ctx context.Context, q *sqlc.Queries, rec TickRecord) error {
	if err := insertTickCommands(ctx, q, rec); err != nil {
		return err
	}

	if err := insertTickEvents(ctx, q, rec); err != nil {
		return err
	}

	// The checksum is the tick's primary key, so a replayed append collides.
	err := q.InsertTickChecksum(ctx, sqlc.InsertTickChecksumParams{
		Tick: int64(rec.Tick),
		Hash: int64(rec.Checksum),
	})
	if err != nil {
		return fmt.Errorf("persist: append tick %d checksum: %w", rec.Tick, err)
	}

	return nil
}

// insertTickCommands writes the pass's commands in submission order.
func insertTickCommands(ctx context.Context, q *sqlc.Queries, rec TickRecord) error {
	for seq, cmd := range rec.Commands {
		if err := q.InsertTickCommand(ctx, commandToParams(rec.Tick, seq, cmd)); err != nil {
			return fmt.Errorf("persist: append tick %d command %d: %w", rec.Tick, seq, err)
		}
	}

	return nil
}

// insertTickEvents writes the pass's events in trace order.
func insertTickEvents(ctx context.Context, q *sqlc.Queries, rec TickRecord) error {
	for seq, ev := range rec.Events {
		if err := q.InsertTickEvent(ctx, eventToParams(rec.Tick, seq, ev)); err != nil {
			return fmt.Errorf("persist: append tick %d event %d: %w", rec.Tick, seq, err)
		}
	}

	return nil
}

// AppendSnapshot writes one snapshot-rotation row.
func (s *SQLiteStore) AppendSnapshot(ctx context.Context, snapshot Snapshot) error {
	if err := s.queries.InsertTickSnapshot(ctx, snapshotToParams(snapshot)); err != nil {
		return fmt.Errorf("persist: append snapshot tick %d: %w", snapshot.Tick, err)
	}

	return nil
}

// LoadTickLog reads the whole recorded log in tick order: each tick's commands,
// its events and its checksum. The checksum rows are the spine, so a tick is in
// the log exactly when its checksum is; a failed tick append leaves none.
func (s *SQLiteStore) LoadTickLog(ctx context.Context) ([]TickRecord, error) {
	trail, err := s.LoadChecksums(ctx)
	if err != nil {
		return nil, err
	}

	log := make([]TickRecord, len(trail))
	index := make(map[int64]int, len(trail))

	for i, entry := range trail {
		log[i] = TickRecord{Tick: entry.Tick, Checksum: entry.Hash}
		index[int64(entry.Tick)] = i
	}

	if err := s.readTickCommands(ctx, log, index); err != nil {
		return nil, err
	}

	if err := s.readTickEvents(ctx, log, index); err != nil {
		return nil, err
	}

	return log, nil
}

// readTickCommands fills each tick's commands in submission order.
func (s *SQLiteStore) readTickCommands(ctx context.Context, log []TickRecord, index map[int64]int) error {
	rows, err := s.queries.ListTickCommands(ctx)
	if err != nil {
		return fmt.Errorf("persist: load tick commands: %w", err)
	}

	for _, row := range rows {
		cmd := kernel.Command{
			Tick: uint32(row.Tick), Kind: kernel.Kind(row.Kind),
			A: int32(row.A), B: int32(row.B), C: int32(row.C), Down: row.Down != 0,
		}
		i := index[row.Tick]
		log[i].Commands = append(log[i].Commands, cmd)
	}

	return nil
}

// readTickEvents fills each tick's events in trace order.
func (s *SQLiteStore) readTickEvents(ctx context.Context, log []TickRecord, index map[int64]int) error {
	rows, err := s.queries.ListTickEvents(ctx)
	if err != nil {
		return fmt.Errorf("persist: load tick events: %w", err)
	}

	for _, row := range rows {
		ev := kernel.Event{
			Tick: uint32(row.Tick), Kind: kernel.Kind(row.Kind),
			A: int32(row.A), B: int32(row.B), C: int32(row.C), Stream: row.Stream, Roll: uint32(row.Roll),
		}
		i := index[row.Tick]
		log[i].Events = append(log[i].Events, ev)
	}

	return nil
}

// LoadChecksums reads the per-tick checksum trail in tick order.
func (s *SQLiteStore) LoadChecksums(ctx context.Context) ([]TickChecksum, error) {
	rows, err := s.queries.ListTickChecksums(ctx)
	if err != nil {
		return nil, fmt.Errorf("persist: load checksums: %w", err)
	}

	trail := make([]TickChecksum, len(rows))
	for i, row := range rows {
		// uint64 checksums ride in SQLite INTEGER columns as their
		// two's-complement int64 image.
		trail[i] = TickChecksum{Tick: uint32(row.Tick), Hash: uint64(row.Hash)}
	}

	return trail, nil
}

// LoadSnapshots reads the snapshot-rotation rows in tick order.
func (s *SQLiteStore) LoadSnapshots(ctx context.Context) ([]Snapshot, error) {
	rows, err := s.queries.ListTickSnapshots(ctx)
	if err != nil {
		return nil, fmt.Errorf("persist: load snapshots: %w", err)
	}

	snapshots := make([]Snapshot, 0, len(rows))
	for _, row := range rows {
		snapshots = append(snapshots, Snapshot{
			Tick: uint32(row.Tick), Checksum: uint64(row.Hash), State: row.State,
		})
	}

	return snapshots, nil
}

// commandToParams maps one recorded command onto its tick_commands row.
func commandToParams(tick uint32, seq int, cmd kernel.Command) sqlc.InsertTickCommandParams {
	// SQLite has no bool column; the pressed flag rides as 0/1.
	down := int64(0)
	if cmd.Down {
		down = 1
	}

	return sqlc.InsertTickCommandParams{
		Tick: int64(tick), Seq: int64(seq), Kind: string(cmd.Kind),
		A: int64(cmd.A), B: int64(cmd.B), C: int64(cmd.C), Down: down,
	}
}

// eventToParams maps one emitted event onto its tick_events row.
func eventToParams(tick uint32, seq int, ev kernel.Event) sqlc.InsertTickEventParams {
	return sqlc.InsertTickEventParams{
		Tick: int64(tick), Seq: int64(seq), Kind: string(ev.Kind),
		A: int64(ev.A), B: int64(ev.B), C: int64(ev.C), Stream: ev.Stream, Roll: int64(ev.Roll),
	}
}

// snapshotToParams maps one snapshot row onto its tick_snapshots row.
func snapshotToParams(snapshot Snapshot) sqlc.InsertTickSnapshotParams {
	return sqlc.InsertTickSnapshotParams{
		Tick: int64(snapshot.Tick), Hash: int64(snapshot.Checksum), State: snapshot.State,
	}
}
