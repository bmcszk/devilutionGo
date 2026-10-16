---
# devilGo-qeyu
title: Automap line primitives + transparent/minimap blend modes
status: todo
type: task
priority: normal
tags:
    - plan:task
    - graphics-extras
created_at: 2026-10-15T17:58:19Z
updated_at: 2026-10-16T20:39:19Z
parent: devilGo-qbiy
blocked_by:
    - devilGo-ru6l
    - devilGo-66o7
    - devilGo-1spo
---

# Automap primitives

## Goal
Port the exact pixel-plotted line/stamp primitives that make the automap look identical, plus its two blend modes; devilGo-zdj7 covers the layout/text/arrow, this covers the raster primitives and blending.

## Research
- docs/research/dg-graphics.md G6 §H (scan loop, colours, stamp families), G8 §A (blend call sites).
- Source/engine/render/automap_render.cpp:26 DrawMapLine<DirX,DirY>, :43 DrawMapLineSteep<DirX,DirY>, :60/:78/:96/:101/:106/:111/:116/:121/:126/:131 wrappers, :139 DrawMapFreeLine (Bresenham), :166 SetMapPixel (minimap clip at :168 + transparent half-blend at :171)
- Source/automap.cpp:167 DrawDiamond, :200/:250 DrawWallConnections (NE/SE variants), :275 DrawDirt (explicit SetMapPixel list), :302-330 other stamps, :38-57 MapColors enum, :980 DrawAutomapTile (per-exploration colours, noConnect rules, lava/acid/water colour choice), :1776-1781 cells/scale computation, :1847-1861 double loop
- Source/automap.h:66 AmLineLength values, AmWidthOffset/AmHeightOffset tables (:41-64)
- Source/engine/render/primitive_render.cpp:18/:37/:89/:107/:140/:155/:159/:187/:215 half-transparent line/rect/pixel paths
- Source/levels/gendung.cpp:424 (.dun transparency layer → dTransVal), Source/automap.cpp:1528 automap.amp
- test/automap_test.cpp

## Scope
internal/render/automapstamps.go (all line families + dirt/stairs/diamond/grate/arch/cave-cross/fence stamps), blend-mode routing through devilGo-66o7's LUT, minimap rect clipping, scale/zoom handling. Extend devilGo-zdj7's renderer to use them.

## Acceptance criteria
- [ ] For each AmLineLength × each of the 8 line functions, a 24×24 index-buffer stamp equals a C++ dump
- [ ] `DrawAutomapTile` golden for every AutomapTile::Types value with representative flag combos (incl. the noConnect special cases for L2/L3 doors and arches)
- [ ] AutomapType::Transparent compares equal to C++ `lut[color][dst]` results, not to an alpha blend
- [ ] Minimap clipping test (pixels outside MinimapRect are not written)
- [ ] automap_test.cpp ported cases pass

## Libraries
- stdlib only (no third-party deps at this stage)

## Skills for coder
/home/blaze/.agents/skills/ebiten/rendering/SKILL.md
/home/blaze/.agents/skills/go-dev/SKILL.md — small integer loops, no float
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md

## Dependencies & blockers
- devilGo-zdj7 (fix2: extension edge wired)
blocked_by devilGo-ru6l; needs devilGo-66o7 for the blend LUT. Blocks dg-epic-11 automap QoL items.

## Open questions
OPEN-QUESTION RESOLUTION 2026-10-16: DECIDED: automap = pixel stamps, never atlas (parity mandate). Documented here.
- Do we ever draw the automap through the sprite atlas instead of pixel stamps? No — stamps are required for parity; document the decision.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.


## Ebiten graphics-migration gates (from docs/research/dg-graphics.md G10/G11 — binding)
- [ ] Line primitives + transparent/minimap blend modes match G6 automap spec (stamps + line families).

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:wire-deps (66o7 blend-LUT + zdj7 extension edges missing; 2 cite drifts)


## IMPL NOTES (triage 2026-10-16)
1 root: Graphics-wave task (devilGo-qeyu): automap raster primitives + 2 blend modes; complementary split vs zdj7 (layout) declared in body ("devilGo-zdj7 covers layout/text/arrow"). Wire: blockedBy [ru6l] only.
2 local: Cites verified: automap_render.cpp:26 DrawMapLine*/43 Steep/139 DrawMapFreeLine ✓ (spot `sed -n '26p;43p;139p'`); SetMapPixel :166 with minimap clip :168 + Transparent half-blend :171 ✓; automap.cpp:167 DrawDiamond ✓, :980/:1341/:1427/:1752/:1776-1781/:1847-1861 ✓, :1528 automap.amp ✓, :275 DrawDirt ✓; primitive_render.cpp:18/:37 paletteTransparencyLookup paths ✓; gendung.cpp:424 dTransVal ✓. DRIFT: DrawWallConnections at :615 (not :200/:250, and one fn not NE/SE pair); AmLineLength at automap.h:74 (not :66); AmWidth/HeightOffset :41-64 ✓. Skills ✓ (3/3 exist).
3 online: n/a (no external claims).
4 solution: Wire missing 2 edges the ACs require: 66o7 (blend LUT — "AutomapType::Transparent compares equal to C++ lut[color][dst]" cannot pass without it) and zdj7 ("Extend devilGo-zdj7's renderer"). Optional cite refresh (:615, :74).
5 understanding: engine/render layer; stamps-not-atlas decision correctly recorded in Open questions; consumes 66o7 LUT, extends zdj7.
6 repro: `sed -n '166,171p' Source/engine/render/automap_render.cpp; grep -n "void DrawWallConnections" Source/automap.cpp; beans query --json '{ bean(id:"devilGo-qeyu") { blockedByIds } }'`
7 fix: `beans update devilGo-qeyu --blocked-by devilGo-66o7 --blocked-by devilGo-zdj7` (verify additivity, not replace: re-query edges after); optionally `--body-replace-old ":200/:250 DrawWallConnections (NE/SE variants)" --body-replace-new ":615 DrawWallConnections"` and `--body-replace-old "automap.h:66 AmLineLength" --body-replace-new "automap.h:74 AmLineLength"`.
8 validation: `beans query --json '{ bean(id:"devilGo-qeyu") { blockedByIds } }'` shows ru6l+66o7+zdj7; body cites match live lines.
9 related: zdj7 (layout twin), 66o7 (LUT), 7aul (automap-on golden frame), ru6l (blocker).
VERDICT: NEEDS-WORK:wire-deps (66o7 blend-LUT + zdj7 extension edges missing; 2 cite drifts)


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).

ID MAP (fix2): all legacy dg-task-NNN references in this bean map to devilGo ids via docs/research/dg-id-map.json (canonical mapping in milestone devilGo-i56j). Operative deps = frontmatter (source of truth); prose refs are documentation.


