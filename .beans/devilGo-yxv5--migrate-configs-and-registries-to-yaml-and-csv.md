---
# devilGo-yxv5
title: 'Migrate settings: diablo.ini checks plus registry YAML views'
status: todo
type: task
priority: normal
created_at: 2026-11-03T12:49:56Z
updated_at: 2026-11-03T14:28:46Z
parent: devilGo-o8e7
---

# Migrate settings: diablo.ini checks plus registry YAML views

## Goal

ADR-014 settings subcommand (ADR-004 law): `devilutionGo-migrate settings`
reads diablo.ini, checks validity, carries the file over byte-identical into
the engine settings path, and emits flat YAML registry views for engine
tooling only. INI stays canonical; YAML is a view, never a second source of
truth. CSV table exports belong to the assets subcommand, not this lane.

## Research

- internal/engine/ini_import.go already reads diablo.ini and
  internal/engine/roundtrip_test.go proves byte-identical round trips.
- [ADR-004](../docs/adrs/adr-004.md): settings live in INI, not SQLite.
- [ADR-013](../docs/adrs/adr-013.md): YAML for flat registry views.

## Scope

- settings subcommand: INI validity checks, byte-identical carryover,
  flat YAML registry view writer.
- Tests: round trip on synthetic INI fixtures committed as text/goldens.

## Acceptance criteria

- [ ] Given a diablo.ini, migration carries it over byte-identical and reports violations it cannot fix.
- [ ] YAML views regenerate byte-identical on a second run.
- [ ] go test ./... green, gofmt clean.

## Libraries

- None new (INI reader exists; YAML writing via the existing stdlib stack).

## Skills for coder

- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/ponytail/SKILL.md
- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/go-tests/SKILL.md
## Dependencies & blockers

- Parent accepted; runs after the engine INI round-trip plumbing it reuses.

## Open questions

- YAML split: one file per registry family unless tooling prefers one file.

## Triage

TRIAGE brainer evidence 2026-11-03: brainer REQUEST-CHANGES on the architecture
PR produced findings; this text was rewritten against the recorded law and the
verified tree (real file seams, no invented files), lint-clean.
