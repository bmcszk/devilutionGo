---
# devilGo-yxv5
title: Migrate configs and registries to YAML and CSV
status: todo
type: task
priority: normal
created_at: 2026-11-03T12:49:56Z
updated_at: 2026-11-03T12:50:42Z
parent: devilGo-o8e7
---

# Migrate configs and registries: INI view to YAML flat, tables to CSV tabular

## Goal

ADR-014 lane 3 + ADR-013 law: `devilutionGo-migrate --config` produces (a)
flat YAML views of registries and options for tooling, INI staying the
canonical settings format per ADR-004, and (b) CSV tabular outputs for the
gamedata tables the txtdata loaders decode, regenerated idempotently and
checked against the existing golden command.

## Research

- internal/engine/ini_import: INI reading; ADR-004 boundary law.
- internal/devilutionGo-txtdata note: the txtdata TSV loaders in
  internal/diablo/gamedata/* decode monster/missile/object/item/level/quest/
  towner tables; internal/game/stringmap.txtdata.py-style already has a TSV
  loader precedent for the engine side.
- cmd/gamedatagoldens: golden comparison command to extend.

## Scope

- cmd/devilutionGo-migrate config lane: YAML registry writer + CSV table writer.
- Golden tests: CSV regeneration is byte-stable across runs; YAML keys stable.

## Acceptance criteria

- [ ] Given one retail install, regeneration produces byte-identical CSV tables and stable YAML keys on a second run.
- [ ] cmd/gamedatagoldens accepts the migrated tables without edits to decoders.
- [ ] go test ./... green, gofmt clean.

## Libraries

- None new (CSV/YAML writers in stdlib or the parsing stacks used today).

## Skills for coder

- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/ponytail/SKILL.md
- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/go-tests/SKILL.md
## Dependencies & blockers

- Parent accepted; idempotence test must run in CI on synthetic fixtures.

## Open questions

- YAML split: one file per registry family unless tooling prefers one file.

## Triage

TRIAGE brainer evidence 2026-11-03: laws mapped from the accepted
ADR-013/014/015/018 set onto the surveyed tree; code seams verified live
(file list in Research); workers follow existing given/when/then test law.
