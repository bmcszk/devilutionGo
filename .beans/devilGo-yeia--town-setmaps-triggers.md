---
# devilGo-yeia
title: Town + setmaps + triggers
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - world
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-16T14:01:06Z
parent: devilGo-97mn
blocked_by:
    - devilGo-76tu
---

# Static maps

## Goal
Town.dun load + warp gating; all SL_* setmaps + their objects/triggers/palettes.

## Research
town.cpp:210-234 IsWarpOpen; setmaps.cpp:112-165 (sklkng/bonecha/foulwatr/vile + arena maps) with per-map palettes/transparency; trigs.cpp Init*Triggers; drlg_quests.

## Scope
internal/world/{town,setmaps,trigs}.go.

## Acceptance criteria
- [ ] Town loads from MPQ, warps respect quest flags
- [ ] All setmaps byte-load (dPiece grids) equal fixtures; triggers fire at documented tiles

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — data-driven map loaders

## Dependencies & blockers
dg-task-017.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:ac-stale-vs-decision — "loads from MPQ" contradicts recorded extract-at-build decision; "Town.dun" file does not exist (4 sector .dun files).
