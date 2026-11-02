-- Version 2: per-level save data. Upstream writes one archive entry per
-- permanent dungeon level (loadsave.cpp:2771-2900, level entries
-- "perml%02d"), so the SQLite store keeps them as rows keyed by hero slot and
-- level id instead of as opaque archive entries.

CREATE TABLE IF NOT EXISTS levels (
  hero_slot INTEGER NOT NULL,
  level_id INTEGER NOT NULL,
  data BLOB NOT NULL,
  updated_at INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (hero_slot, level_id)
);
