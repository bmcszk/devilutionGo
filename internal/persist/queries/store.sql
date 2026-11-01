-- Queries of the live save store. sqlc generates internal/persist/sqlc from
-- this file; nothing in the store may hand-write these statements.

-- name: UpsertHero :exec
INSERT INTO heroes (slot, name, class_id, created_at, updated_at, pack, spawn, hellfire, multiplayer)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(slot) DO UPDATE SET
  name = excluded.name, class_id = excluded.class_id, updated_at = excluded.updated_at,
  pack = excluded.pack, spawn = excluded.spawn, hellfire = excluded.hellfire,
  multiplayer = excluded.multiplayer;

-- name: GetHero :one
SELECT sqlc.embed(heroes) FROM heroes WHERE slot = ?;

-- name: ListHeroes :many
SELECT sqlc.embed(heroes) FROM heroes ORDER BY slot;

-- name: DeleteHero :exec
DELETE FROM heroes WHERE slot = ?;

-- name: DeleteGame :exec
DELETE FROM games WHERE hero_slot = ?;

-- name: DeleteStash :exec
DELETE FROM stash WHERE hero_slot = ?;

-- name: UpsertGame :exec
INSERT INTO games (hero_slot, kind, data, updated_at) VALUES (?, ?, ?, ?)
ON CONFLICT(hero_slot, kind) DO UPDATE SET data = excluded.data, updated_at = excluded.updated_at;

-- name: GetGame :one
SELECT data FROM games WHERE hero_slot = ? AND kind = ?;

-- name: UpsertStash :exec
INSERT INTO stash (hero_slot, page_count, blob) VALUES (?, ?, ?)
ON CONFLICT(hero_slot) DO UPDATE SET page_count = excluded.page_count, blob = excluded.blob;

-- name: GetStash :one
SELECT page_count, blob FROM stash WHERE hero_slot = ?;

-- name: UpsertLevel :exec
INSERT INTO levels (hero_slot, level_id, data, updated_at) VALUES (?, ?, ?, ?)
ON CONFLICT(hero_slot, level_id) DO UPDATE SET data = excluded.data, updated_at = excluded.updated_at;

-- name: GetLevel :one
SELECT data FROM levels WHERE hero_slot = ? AND level_id = ?;

-- name: DeleteLevels :exec
DELETE FROM levels WHERE hero_slot = ?;
