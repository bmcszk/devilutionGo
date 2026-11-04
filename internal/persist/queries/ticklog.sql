-- Queries of the tick-log column family (schema_v3.sql). sqlc generates
-- internal/persist/sqlc from this file; the store keeps no tick-log SQL of its
-- own, the same law queries/store.sql carries.

-- name: InsertTickCommand :exec
INSERT INTO tick_commands (tick, seq, kind, a, b, c, down) VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: InsertTickEvent :exec
INSERT INTO tick_events (tick, seq, kind, a, b, c, stream, roll) VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: InsertTickChecksum :exec
INSERT INTO tick_checksums (tick, hash) VALUES (?, ?);

-- name: InsertTickSnapshot :exec
INSERT INTO tick_snapshots (tick, hash, state) VALUES (?, ?, ?);

-- The replay input, in submission order within a tick.
-- name: ListTickCommands :many
SELECT tick, kind, a, b, c, down FROM tick_commands ORDER BY tick, seq;

-- The emitted state changes, in trace order within a tick.
-- name: ListTickEvents :many
SELECT tick, kind, a, b, c, stream, roll FROM tick_events ORDER BY tick, seq;

-- The hash trail a divergence hunt bisects, one row per recorded tick.
-- name: ListTickChecksums :many
SELECT tick, hash FROM tick_checksums ORDER BY tick;

-- The snapshot rotation rows a load resumes from.
-- name: ListTickSnapshots :many
SELECT tick, hash, state FROM tick_snapshots ORDER BY tick;
