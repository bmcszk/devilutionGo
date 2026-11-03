---
# devilGo-v7ft
title: 'Migrate assets: MPQ to assets dir plus YAML asset lists'
status: in-progress
type: task
priority: high
created_at: 2026-11-03T12:49:56Z
updated_at: 2026-11-03T14:55:41Z
parent: devilGo-o8e7
---

# Migrate assets: MPQ to assets dir plus CSV tables and YAML asset lists

## Goal

ADR-014 assets subcommand on top of the ADR-013 law: `devilutionGo-migrate
assets` walks the configured MPQ set, extracts through the proven decoders at
full coverage, writes the assets dir layout, and emits YAML asset lists plus
CSV tables. The YAML is the manifest record: family, archive path, format,
frame layout, hashes; the graphics manifest work converges onto this same
YAML form (no TSV fork).

## Research

- internal/assets: exhaustive extraction proven on retail DIABDAT + hellfire
  + patch MPQs (100% of records, zero failures; PCX decodes).
- The graphics manifest feature names the manifest craft this tool converges
  with; ADR-013 fixed YAML as the list format.
- docs/recipes holds the extraction conventions workers follow today.

## Scope

- assets subcommand; YAML manifest writer; CSV table writer; hash summary.
- Tests on synthetic fixtures only (no MPQ in the repository); retail
  verification documented as an environment amendment, never a silent skip.

## Acceptance criteria

- [ ] Given a retail MPQ in the configured dir, migration writes the assets dir, a valid YAML manifest and the CSV tables with per-file hashes.
- [ ] Re-running is idempotent (same output hash summary).
- [ ] go test ./... green on synthetic fixtures; no binaries committed.

## Libraries

- None new.

## Skills for coder

- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/ponytail/SKILL.md
- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/go-tests/SKILL.md
## Dependencies & blockers

- Parent accepted; graphics manifest work converges onto the YAML target built here.

## Open questions

- Manifest filename + location: assets/manifest.yaml unless the graphics
  manifest feature already fixed another.

## Triage

TRIAGE brainer evidence 2026-11-03: brainer REQUEST-CHANGES on the architecture
PR produced findings; this text was rewritten against the recorded law and the
verified tree (real file seams, no invented files), lint-clean.

Claim evidence 2026-11-03: lane dispatched. Worker session task-assets-lane-coder,
branch dg-migrate-assets (worktree /home/blaze/github/bmcszk/dg-lane-migrate),
off master 1cb9ffd. Deliverable: the migration assets subcommand per this
bean; PR targets master.
