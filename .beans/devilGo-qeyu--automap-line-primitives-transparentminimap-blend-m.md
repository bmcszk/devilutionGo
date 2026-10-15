---
# devilGo-qeyu
title: Automap line primitives + transparent/minimap blend modes
status: draft
type: task
priority: normal
tags:
    - plan:task
    - graphics-extras
created_at: 2026-10-15T17:58:19Z
updated_at: 2026-10-15T21:13:42Z
parent: devilGo-qbiy
blocked_by:
    - devilGo-ru6l
---

# Automap primitives

## Goal
Port the exact pixel-plotted line/stamp primitives that make the automap look identical, plus its two blend modes; dg-task-013 covers the layout/text/arrow, this covers the raster primitives and blending.

## Research
- /tmp/dg-graphics.md G6 §H (scan loop, colours, stamp families), G8 §A (blend call sites).
- Source/engine/render/automap_render.cpp:26 DrawMapLine<DirX,DirY>, :43 DrawMapLineSteep<DirX,DirY>, :60/:78/:96/:101/:106/:111/:116/:121/:126/:131 wrappers, :139 DrawMapFreeLine (Bresenham), :166 SetMapPixel (minimap clip at :168 + transparent half-blend at :171)
- Source/automap.cpp:167 DrawDiamond, :200/:250 DrawWallConnections (NE/SE variants), :275 DrawDirt (explicit SetMapPixel list), :302-330 other stamps, :38-57 MapColors enum, :980 DrawAutomapTile (per-exploration colours, noConnect rules, lava/acid/water colour choice), :1776-1781 cells/scale computation, :1847-1861 double loop
- Source/automap.h:66 AmLineLength values, AmWidthOffset/AmHeightOffset tables (:41-64)
- Source/engine/render/primitive_render.cpp:18/:37/:89/:107/:140/:155/:159/:187/:215 half-transparent line/rect/pixel paths
- Source/levels/gendung.cpp:424 (.dun transparency layer → dTransVal), Source/automap.cpp:1528 automap.amp
- test/automap_test.cpp

## Scope
internal/render/automapstamps.go (all line families + dirt/stairs/diamond/grate/arch/cave-cross/fence stamps), blend-mode routing through dg-task-202's LUT, minimap rect clipping, scale/zoom handling. Extend dg-task-013's renderer to use them.

## Acceptance criteria
- [ ] For each AmLineLength × each of the 8 line functions, a 24×24 index-buffer stamp equals a C++ dump
- [ ] `DrawAutomapTile` golden for every AutomapTile::Types value with representative flag combos (incl. the noConnect special cases for L2/L3 doors and arches)
- [ ] AutomapType::Transparent compares equal to C++ `lut[color][dst]` results, not to an alpha blend
- [ ] Minimap clipping test (pixels outside MinimapRect are not written)
- [ ] automap_test.cpp ported cases pass

## Skills for coder
/home/blaze/.agents/skills/ebiten/rendering/SKILL.md
/home/blaze/.agents/skills/go-dev/SKILL.md — small integer loops, no float
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md

## Dependencies & blockers
blocked_by dg-task-010; needs dg-task-202 for the blend LUT. Blocks dg-epic-11 automap QoL items.

## Open questions
- Do we ever draw the automap through the sprite atlas instead of pixel stamps? No — stamps are required for parity; document the decision.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.


## Ebiten graphics-migration gates (from /tmp/dg-graphics.md G10/G11 — binding)
- [ ] Line primitives + transparent/minimap blend modes match G6 automap spec (stamps + line families).
