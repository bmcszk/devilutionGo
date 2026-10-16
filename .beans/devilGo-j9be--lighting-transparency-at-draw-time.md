---
# devilGo-j9be
title: Lighting + transparency at draw time
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - render
created_at: 2026-10-15T16:35:16Z
updated_at: 2026-10-16T14:26:53Z
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
Light index→ColorScale (SUPERSEDED by G4 palette decision — do not port; TRN path only) mapping precomputed; blend via alpha composite; cycling timer updating palette variant images at tick boundary.

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


## Ebiten graphics-migration gates (from docs/research/dg-graphics.md G10/G11 — binding)
- [ ] Light tables applied via index math identical to G5 formula (float32 arithmetic mirror verified byte-exact vs C++ dump `test/fixtures/light_render_benchmark/dLight.dmp`).
- [ ] Transparency blending uses the blend LUT (G4 256×256) with EXACT operand order per G8; pixel-diff vs C++ frame clean.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:scope-conflicts-with-G4-decision (ColorScale line + stale open question) + stale body dep dg-task-010.


## IMPL NOTES (triage 2026-10-16)
1 root: Draw-time light/blend/cycling application bean. AC testable (16-level hash, blend vs reference, tick-bound cycling) but scope text contradicts the recorded graphics decision.
2 local: `beans show` + probes: `ClxDrawLight` confirmed at `Source/engine/render/scrollrt.cpp:252-258` (bean's 249-258 ✓); cycling `Source/diablo.cpp:3509` `diablo_color_cyc_logic()` caves/hell/nest/crypt ✓; `test/light_render_benchmark.cpp` exists ✓; `dLight.dmp` fixture EXISTS in C++ tree ✓. BUT scope line "Light index→ColorScale mapping precomputed" contradicts recorded G4 decision — docs/research/dg-graphics.md:18-19: tables via "index textures + one palette-lookup Kage fragment shader", NOT ColorScale; 66o7's body explicitly says "Replace the Volume I §7 ColorScale plan". Open question "ColorScale vs 16 pre-shaded image sets" is already DECIDED by G4 → stale. Body dep `dg-task-010` dead; frontmatter says `devilGo-ru6l` ✓.
3 online: n/a (no external claims).
4 solution: rewrite scope per G4 (draw-time application = light-table index → palette-row selection + blend-LUT composition over the index buffer, per 66o7 architecture); delete the stale ColorScale open question; sync body dep to ru6l; keep the G5/G8 gates (they are correct and self-consistent — the gates already contradict the scope line, which proves the scope line is the leftover).
5 understanding: Diablo-layer draw-path glue between ru6l (world draw) and m18k (shader); 66o7 owns the architecture, this bean owns when/where light+blend+cycling apply per sprite/tile.
6 repro: `beans show devilGo-j9be --body-only`; `sed -n '249,258p' <devilutionX-repo>/Source/engine/render/scrollrt.cpp`; `sed -n '3509,3529p' .../Source/diablo.cpp`; `grep -n "ColorScale" docs/research/dg-graphics.md | head -3`; `grep -A2 '^blocked_by' .beans/devilGo-j9be--*.md`.
7 fix: `beans update devilGo-j9be --body-replace-old "Light index→ColorScale mapping precomputed; blend via alpha composite; cycling timer updating palette variant images at tick boundary." --body-replace-new "Per G4/66o7: light index selects the palette row on the index framebuffer (no ColorScale — G4 decision); transparency via the 256×256 blend LUT with exact operand order (G8); cycling timer advances palette variants at tick boundary only."`; `beans update devilGo-j9be --body-replace-old "- ColorScale vs 16 pre-shaded image sets: memory bench decides (H6-adjacent)." --body-replace-new "- None (ColorScale question closed by G4 decision: index framebuffer + palette-lookup shader)."`; `beans update devilGo-j9be --body-replace-old "dg-task-010." --body-replace-new "devilGo-ru6l (world draw order)."`.
8 validation: `beans show devilGo-j9be --body-only | grep -ci colorscale` → 0; body dep line shows devilGo-ru6l.
9 related: ru6l (dep), m18k (blocked by this), 66o7 (architecture owner), kx0l (table source), pdpc (lightmap selection).
VERDICT: NEEDS-WORK:scope-conflicts-with-G4-decision (ColorScale line + stale open question) + stale body dep dg-task-010.

Decision-alignment (triage): ColorScale line + stale open question superseded by G4 palette-decision and 2026-10-16 resolution model; body cites IMPL NOTES.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).
