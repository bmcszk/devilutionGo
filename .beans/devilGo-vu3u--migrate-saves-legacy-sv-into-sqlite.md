---
# devilGo-vu3u
title: 'Migrate saves: legacy .sv into SQLite'
status: todo
type: task
priority: normal
created_at: 2026-11-03T12:49:56Z
updated_at: 2026-11-03T12:50:42Z
parent: devilGo-o8e7
---

# Migrate saves: legacy .sv into SQLite through the persist importers

## Goal

ADR-014 lane 2 on top of the ADR-003 law: `devilutionGo-migrate --save`
reads legacy .sv saves, rebuilds them through internal/persist codecs, writes
the SQLite store, prints a checksum summary per file, and stays checkpointed:
a partial run resumes without redoing converted saves.

## Research

- internal/persist: SaveGameData/ItemPack/PlayerPack codecs + importers +
  sqlite store; in-progress children there already harden this surface.
- cmd/gamedatagoldens: golden comparison pattern for verification output.
- [ADR-003](../docs/adrs/adr-003.md), [ADR-014](../docs/adrs/adr-014.md).

## Scope

- cmd/devilutionGo-migrate save lane; checkpoint journal; per-save checksum report.
- Tests with synthetic save fixtures committed as text/goldens, never binaries.

## Acceptance criteria

- [ ] Given a legacy .sv, migration produces a SQLite store that reads back field-correct through the persist codecs.
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

- DB location default: save dir next to the legacy file unless settings say otherwise.

## Triage

TRIAGE brainer evidence 2026-11-03: laws mapped from the accepted
ADR-013/014/015/018 set onto the surveyed tree; code seams verified live
(file list in Research); workers follow existing given/when/then test law.
