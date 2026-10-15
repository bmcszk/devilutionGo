---
# devilGo-j9be
title: Lighting + transparency at draw time
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - render
created_at: 2026-10-15T16:35:16Z
updated_at: 2026-10-15T21:13:42Z
parent: devilGo-qbiy
blocked_by:
    - devilGo-ru6l
---

# Light & blend

## Goal
Light-table TRN application, blended transparency, color-cycling palettes.

## Research
ClxDrawLight/Blended scrollrt.cpp:249-258; LightTables 16×256 engine/light_tables; cycling diablo.cpp:3509-3529 (caves/hell/hive/crypt tables); benchmark light_render_benchmark.cpp.

## Scope
Light index→ColorScale mapping precomputed; blend via alpha composite; cycling timer updating palette variant images at tick boundary.

## Acceptance criteria
- [ ] 16 light levels visually hashed equal to C++ for a wall+floor scene
- [ ] Blended arch transparency correct vs reference
- [ ] Cycling advances only on ticks (test)

## Skills for coder
/home/blaze/.agents/skills/ebiten/rendering/SKILL.md — ColorScale/blend options

## Dependencies & blockers
dg-task-010.

## Open questions
- ColorScale vs 16 pre-shaded image sets: memory bench decides (H6-adjacent).


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.


## Ebiten graphics-migration gates (from /tmp/dg-graphics.md G10/G11 — binding)
- [ ] Light tables applied via index math identical to G5 formula (float32 arithmetic mirror verified byte-exact vs C++ dump `test/fixtures/light_render_benchmark/dLight.dmp`).
- [ ] Transparency blending uses the blend LUT (G4 256×256) with EXACT operand order per G8; pixel-diff vs C++ frame clean.
