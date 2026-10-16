---
# devilGo-9o2h
title: Vision system — VisionRays, DoVision, TransList, ProcessVisionList (L5)
status: draft
type: task
priority: normal
tags:
    - phase-lighting
created_at: 2026-10-15T22:41:05Z
updated_at: 2026-10-16T14:01:03Z
parent: devilGo-97mn
blocked_by:
    - devilGo-66o7
    - devilGo-b782
    - devilGo-kx0l
    - devilGo-bodd
    - devilGo-m18k
---

# Vision system — LOS rays + dFlags + TransList

## Goal
Port the complete vision/visibility system (L5): VisionRays[23][15] Bresenham table, DoVision ray-walk with 4 callbacks (markVisible/markTransparent/passesLight/inBounds), TransList transparent-tile reveal, ProcessVisionList two-pass per-player update with friendlyMode automap bits.

## Research
- Source/vision.hpp:20-30 DoVision signature with 4 function_ref callbacks
- Source/vision.cpp:11-110 VisionRays[23][15] array verbatim (23 rays one quadrant, trailing {0,0} trimmed)
- Source/vision.cpp:98-120 DoVision: for each ray, step points; on !passesLight break; call markVisible/markTransparent; mirror across 8 octants
- Source/lighting.cpp:84-94 DoVisionFlags: dFlags|=Explored|Lit|Visible; automap SetAutomapView if doAutomap
- Source/lighting.cpp:431-468 ProcessVisionList: PASS A un-vision changed/inactive; PASS B apply per player; friendlyMode→MAP_EXP_OTHERS else MAP_EXP_NONE; MyPlayer visible=true; TransList reset each call
- TransList indexed by dTransVal (from dTransVal array)

## Scope
Package: `internal/light/vision.go`
- VisionRays const [[23][15]DisplacementI8] embedded verbatim
- `func DoVision(pos Point, radius uint8, mv, mt, pl, ib func(Point))` with ray walk, mirror, early break on BlockLight
- `func ProcessVisionList(players []*Player)` two-pass logic, TransList reset, VisionActive/VisionList mutation
- `TransList` as [256]bool (dTransVal range 0..255)

## Acceptance criteria
- [ ] `TestVisionRays` all 23 rays x 15 steps match C++ exact coordinates (golden array)
- [ ] `TestDoVisionStopAtWall` ray breaks at BlockLight tile, no reveal beyond
- [ ] `TestTransListSet` transparent tile (dTransVal!=0) sets TransList[trans]=true
- [ ] `TestProcessVisionListParity` 4-player scenario (MyPlayer+friendly+hostile+off-level) → dFlags bits + TransList match C++ fixture
- [ ] `TestVisionUnlight` changed/inactive vision cleared via DoUnVision radius+2 rectangle

## Skills for coder
- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/ebiten/rendering/SKILL.md

## Dependencies & blockers
- devilGo-gk5e (dungeon bounds, dTransVal array)

## Open questions
- None - VisionRays table and ray algorithm fully transcribed from vision.cpp

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:wrong-deps — wire (5 render/light beans) contradicts body (gk5e) and both mis-target; correct blocker is vti7; 2 cite drifts; content itself verified exact.
