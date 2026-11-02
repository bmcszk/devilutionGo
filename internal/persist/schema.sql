-- Version 1 of the live save store (docs/research/dg-requirements.md,
-- "Data structures (SQLite schema proposal)"). This is the single source of
-- the DDL: the migrate() chain embeds this file and sqlc is generated against
-- it, so the schema the store applies and the schema sqlc type-checks against
-- cannot drift. Columns the store always writes are NOT NULL, which is also
-- what lets sqlc emit plain Go types instead of sql.Null* wrappers.

CREATE TABLE IF NOT EXISTS meta (
  key TEXT PRIMARY KEY NOT NULL,
  value BLOB NOT NULL
);

CREATE TABLE IF NOT EXISTS heroes (
  id INTEGER PRIMARY KEY,
  slot INTEGER NOT NULL UNIQUE,
  name TEXT NOT NULL DEFAULT '',
  class_id INTEGER NOT NULL DEFAULT 0,
  created_at INTEGER NOT NULL DEFAULT 0,
  updated_at INTEGER NOT NULL DEFAULT 0,
  pack BLOB NOT NULL,
  spawn INTEGER NOT NULL DEFAULT 0,
  hellfire INTEGER NOT NULL DEFAULT 0,
  multiplayer INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS games (
  hero_slot INTEGER NOT NULL,
  kind TEXT NOT NULL,
  data BLOB NOT NULL,
  updated_at INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (hero_slot, kind)
);

CREATE TABLE IF NOT EXISTS stash (
  hero_slot INTEGER PRIMARY KEY NOT NULL,
  page_count INTEGER NOT NULL DEFAULT 0,
  -- nullable: the row exists from the first save, the page image only once a
  -- page is written.
  blob BLOB
);
