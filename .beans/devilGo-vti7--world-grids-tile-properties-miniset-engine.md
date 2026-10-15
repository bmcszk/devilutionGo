---
# devilGo-vti7
title: World grids + tile properties + miniset engine
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - world
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-15T16:45:27Z
parent: devilGo-97mn
blocked_by:
    - devilGo-m2x7
---

# Grids foundation

## Goal
The 12 world arrays + SOL loading + Miniset match/place + transparency bookkeeping.

## Research
dun_tile_data.hpp:26-120; gendung.h Miniset :35-70; tile_properties.cpp:11-41 + CanStep rules tile_properties.hpp:28-37; SOL loading LoadLevelSOLData.

## Scope
internal/world/grid.go, tileprops.go, miniset.go, trans.go (DRLG_InitTrans/MRectTrans/RectTrans/CopyTrans).

## Acceptance criteria
- [ ] tile_properties_test.cpp port passes
- [ ] Miniset matches/place unit tests incl. Protected flag and crypt respectProtected=false quirk
- [ ] TransList semantics (TransVal) identical on scripted ops

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — fixed arrays, no slices in hot paths

## Dependencies & blockers
dg-task-002.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.
