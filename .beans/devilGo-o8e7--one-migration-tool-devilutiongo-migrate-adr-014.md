---
# devilGo-o8e7
title: One migration tool devilutionGo-migrate (ADR-014)
status: todo
type: feature
priority: normal
created_at: 2026-11-03T12:49:56Z
updated_at: 2026-11-03T14:55:41Z
parent: devilGo-i56j
---

# One migration tool devilutionGo-migrate

## Goal

One CLI, `cmd/devilutionGo-migrate`, with the record's three subcommands
(ADR-014): `assets` (MPQ contents into the assets dir plus CSV tables and
YAML lists from ADR-013), `saves` (legacy .sv and .sp into the per-hero .db
files plus global.db, replacing the stand-alone importer command), and
`settings` (diablo.ini checks and carryover, ADR-004). Thin CLI over already
round-trip-proven producers; checkpointed and rerunnable end to end.

## Research

- internal/assets decodes 100% of records from retail MPQs (asset
  verification PoW); the assets subcommand is a CLI over these decoders.
- internal/persist offers SaveGameData/ItemPack/PlayerPack codecs, the sqlite
  store and importers; cmd/gamedatagoldens is the house pattern for golden
  fixture output (zero-dependency policy: committed TSVs in internal/testdata
  only, never an MPQ).
- internal/engine/ini_import.go reads diablo.ini (roundtrip_test.go proves
  byte-identical carryover); [ADR-013](../docs/adrs/adr-013.md) sets the YAML
  list law and [ADR-004](../docs/adrs/adr-004.md) keeps the INI canonical.

## Scope

Children:

1. Assets subcommand: MPQ to assets dir, CSV tables and YAML asset lists.
2. Saves subcommand: legacy .sv and .sp into SQLite through persist importers.
3. Settings subcommand: diablo.ini checks and carryover plus registry YAML views.
4. Recipes page + CI fixture path.

## Acceptance criteria

- [ ] One binary covers all three subcommands with checkpointed, rerunnable steps.
- [ ] No MPQ or save binaries enter the repository; outputs land in dirs from the environment (never test dependencies).
- [ ] Each subcommand prints a final checksum summary and exits nonzero on any mismatch.

## Libraries

- modernc.org/sqlite (store driver per the accepted stack)
- No new dependencies.

## Skills for coder

- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/ponytail/SKILL.md
- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/go-tests/SKILL.md
## Dependencies & blockers

- The saves subcommand leans on the persistence feature in progress; it reuses
  its importers unchanged.

## Open questions

- None blocking.

## Triage

TRIAGE brainer evidence 2026-11-03: brainer REQUEST-CHANGES on the architecture
PR produced findings; this text was rewritten against the recorded law and the
verified tree (real file seams, no invented files), lint-clean.
