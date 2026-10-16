---
# devilGo-ru6l
title: Isometric world renderer (floor/walls/order)
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - render
created_at: 2026-10-15T16:35:16Z
updated_at: 2026-10-16T14:01:05Z
parent: devilGo-qbiy
blocked_by:
    - devilGo-n89r
---

# World draw

## Goal
Faithful scrollrt port for static world + entity order.

## Research
scrollrt.cpp:949-1036 (DrawFloor/DrawTileContent/wall pre-pass), :790-980 (DrawDungeon order), :1229 DrawGame; displacement worldToScreen Source/engine/displacement.hpp:151-164.

## Scope
internal/render/world.go with tile iterators mirroring row walk; DrawDungeon implementing the exact order incl. negative dPlayer for moving players; dirt-tile OOB fallback (scrollrt.cpp DrawDirtTile).

## Acceptance criteria
- [ ] Order-instrumented test equals expected sequence for a scripted 3×3 scene
- [ ] Golden hash at progress 0/64/128 for one seed
- [ ] Zero Draw allocs

## Skills for coder
/home/blaze/.agents/skills/ebiten/rendering/SKILL.md — batching identical-options draws
/home/blaze/.agents/skills/go-dev/SKILL.md — index math without bounds pain

## Dependencies & blockers
dg-task-007 (sprites).

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.


## Ebiten graphics-migration gates (from /tmp/dg-graphics.md G10/G11 — binding)
- [ ] Draw order matches `scrollrt.cpp:781 DrawDungeon` 12-layer order exactly; entity y-sort verified by parity frame.
- [ ] Render loop lives in Draw() only; no state mutation; per-frame cost ≤7k triangles (G10 budget).


## Resolution gate (grilling decision #5)
- [ ] 640×480 world logical surface rendered correctly, AND 2x window (1280×960) integer-scaled without artifacts — parity golden frames at BOTH sizes.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:stale-body-deps (dg-task-007) + research paths missing `engine/render/` prefix; logic/AC otherwise verified sound (order test, golden hash 0/64/128, zero-alloc — all testable; TDD gates present).
