---
# devilGo-zdj7
title: Automap renderer
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - render
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-15T16:45:25Z
parent: devilGo-qbiy
blocked_by:
    - devilGo-ru6l
---

# Automap

## Goal
Automap tiles, player arrow, text labels, modes.

## Research
Source/automap.cpp:980 DrawAutomapTile, :1341 Plr, :1427 Text, :1752-1874 DrawAutomap; modes automap.h:83,91; automap_render.cpp:168-171; test automap_test.cpp.

## Scope
internal/render/automap.go reusing world SOL data; toggle key; minimap/transparent variants.

## Acceptance criteria
- [ ] automap_test cases pass
- [ ] Player arrow rotates with _pdir

## Skills for coder
/home/blaze/.agents/skills/ebiten/rendering/SKILL.md — line/stamp drawing

## Dependencies & blockers
dg-task-010.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.
