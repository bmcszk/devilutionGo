---
# devilGo-vu3u
title: 'Migrate saves: legacy .sv into SQLite'
status: todo
type: task
priority: normal
created_at: 2026-11-03T12:49:56Z
updated_at: 2026-11-03T14:23:09Z
parent: devilGo-o8e7
---

# Migrate saves: legacy .sv and .sp into SQLite

## Goal

ADR-014 saves subcommand on top of the ADR-003 law: `devilutionGo-migrate
saves` reads legacy .sv and .sp saves, rebuilds them through internal/persist
codecs, writes the per-hero .db files plus global.db, prints a checksum
summary per file, and stays checkpointed: a partial run resumes without
redoing converted saves.

## Research

- internal/persist: SaveGameData/ItemPack/PlayerPack codecs + importers +
  sqlite store; in-progress children there already harden this surface.
- cmd/gamedatagoldens: golden fixture pattern for verification output
  (zero-dependency policy: committed fixtures only).
- [ADR-003](../docs/adrs/adr-003.md), [ADR-014](../docs/adrs/adr-014.md).

## Scope

- saves subcommand; checkpoint journal; per-save checksum report;
  synthetic save fixtures committed as text/goldens, never binaries.

## Acceptance criteria

- [ ] Given a legacy .sv or .sp, migration produces the store files that read back field-correct through the persist codecs.
- [ ] Interrupted run resumes; already-converted saves are not rewritten.
- [ ] go test ./... green, gofmt clean; no save binaries committed.

## Libraries

- modernc.org/sqlite (already present).

## Skills for coder

- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/ponytail/SKILL.md
- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/go-tests/SKILL.md
## Dependencies & blockers

- sqlite store work in progress inside the persistence feature; this lane
  consumes its importers unchanged.

## Open questions

- Destination path: next to the legacy file unless settings say otherwise.

## Triage

TRIAGE brainer evidence 2026-11-03: brainer REQUEST-CHANGES on the architecture
PR produced findings; this text was rewritten against the recorded law and the
verified tree (real file seams, no invented files), lint-clean.
