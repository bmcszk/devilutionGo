-- Version 3: the tick-log column family (ADR-018 law 4). The log is not a
-- second save format: it lives in the same per-hero store the rest of the save
-- lives in (ADR-003), keyed by tick instead of by slot, because one store
-- already is one hero's game.
--
-- Two row sets carry a tick: the commands the kernel was given (the replay
-- input) and the events it emitted (the state changes the diverging runs are
-- compared on). The checksum row per tick is what a desync hunt bisects on.
CREATE TABLE IF NOT EXISTS tick_commands (
  tick INTEGER NOT NULL,
  seq INTEGER NOT NULL,
  kind TEXT NOT NULL,
  a INTEGER NOT NULL,
  b INTEGER NOT NULL,
  c INTEGER NOT NULL,
  down INTEGER NOT NULL,
  PRIMARY KEY (tick, seq)
);

CREATE TABLE IF NOT EXISTS tick_events (
  tick INTEGER NOT NULL,
  seq INTEGER NOT NULL,
  kind TEXT NOT NULL,
  a INTEGER NOT NULL,
  b INTEGER NOT NULL,
  c INTEGER NOT NULL,
  stream TEXT NOT NULL,
  roll INTEGER NOT NULL,
  PRIMARY KEY (tick, seq)
);

-- One cumulative trace checksum per tick, as the kernel reported it during the
-- recorded run. Replay recomputes it; a mismatch is a desync at that tick.
CREATE TABLE IF NOT EXISTS tick_checksums (
  tick INTEGER PRIMARY KEY NOT NULL,
  hash INTEGER NOT NULL
);

-- Snapshot rotation: every SnapshotEvery ticks the log stores the checksum (and
-- an optional opaque state image) so a load can start replaying from here
-- instead of from tick 0.
CREATE TABLE IF NOT EXISTS tick_snapshots (
  tick INTEGER PRIMARY KEY NOT NULL,
  hash INTEGER NOT NULL,
  state BLOB
);
