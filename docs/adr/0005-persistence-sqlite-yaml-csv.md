# ADR-005: Persistence: SQLite saves, YAML configs, CSV tables

Date: 2026-11-03

**Status:** Proposed

**Decision.**
- SQLite (modernc.org/sqlite, pure Go, CGO_ENABLED=0 preserved) as the save
  store: one DB per character (event-sourced log + materialized snapshots),
  event log retained for replay/audit.
- Configs/registries (spells, monsters, items, affixes, levels): **YAML** for
  flat structures, **CSV** for tabular (bulk) structures — mirrors your
  migration directive exactly.
- Save migration (your item 2): original `.sv`/MPQ saves + configs + MPQ
  assets → modern layout (assets/ dir, sqlite saves, yaml/csv configs). One
  migration binary, idempotent, dry-run + checksums.

**Consequences.** + no cgo, cross-compile stays trivial; migration becomes a
repeatable tool not a one-off; − save schema versions must be handled via
event-log versioning.
