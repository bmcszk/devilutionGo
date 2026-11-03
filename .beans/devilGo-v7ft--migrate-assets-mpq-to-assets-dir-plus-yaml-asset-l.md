---
# devilGo-v7ft
title: 'Migrate assets: MPQ to assets dir plus YAML asset lists'
status: todo
type: task
priority: high
created_at: 2026-11-03T12:49:56Z
updated_at: 2026-11-03T12:50:42Z
parent: devilGo-o8e7
---

# Migrate assets: MPQ into assets dir plus YAML asset lists

## Goal

ADR-014 lane 1 + ADR-013 YAML law: `devilutionGo-migrate --assets` walks the
configured MPQ set, extracts through the proven decoders at full coverage,
writes the assets dir layout, and emits YAML asset lists (manifest: family,
archive path, format, frame layout, hashes). Schema follows the graphics
manifest feature so renderer tooling and migration do not fork.

## Research

- internal/assets: exhaustive extraction proven on retail DIABDAT + hellfire
  + patch MPQs (100% of records, zero failures; PCX decodes).
- The graphics manifest feature names the manifest target this tool must
  produce (see parent feature links at dispatch time).
- docs/recipes holds the extraction conventions workers follow today.

## Scope

- cmd/devilutionGo-migrate assets lane; YAML manifest writer; hash summary print.
- Tests on synthetic fixtures only (no MPQ in repo); retail verification
  documented as an environment amendment, never a silent skip.

## Acceptance criteria

- [ ] Given a retail MPQ in the configured dir, migration writes the assets dir and a valid YAML manifest with per-file hashes.
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

- Parent accepted; graphics manifest feature defines the YAML schema target.

## Open questions

- Manifest filename + location: assets/manifest.yaml unless the graphics
  manifest feature already fixed one.

## Triage

TRIAGE brainer evidence 2026-11-03: kernel/migration laws come straight from the
accepted architecture records ADR-013..018 (deep research confrontation:
Game Programming Patterns event-queue chapter, Bevy observer-ordering doc,
Watermill GoChannel reference, AoE/Factorio/OpenTTD lockstep and desync
write-ups, devilutionX demo format parity). Code seams verified in the tree
surveys; workers get exact files in Research.
