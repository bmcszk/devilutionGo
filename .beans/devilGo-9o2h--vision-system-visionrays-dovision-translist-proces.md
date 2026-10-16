---
# devilGo-9o2h
title: Vision system — VisionRays, DoVision, TransList, ProcessVisionList (L5)
status: todo
type: task
priority: normal
tags:
    - phase-lighting
created_at: 2026-10-15T22:41:05Z
updated_at: 2026-10-16T20:37:10Z
parent: devilGo-97mn
blocked_by:
    - devilGo-66o7
    - devilGo-b782
    - devilGo-kx0l
    - devilGo-bodd
    - devilGo-m18k
    - devilGo-1spo
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


## IMPL NOTES (triage 2026-10-16)
1 root: Lighting-wave task (dg-task-256): vision/visibility logic package; draft; wire blockedBy [66o7, b782, kx0l, bodd, m18k] vs body deps "devilGo-gk5e (dungeon bounds, dTransVal array)" — MAJOR mismatch, and both look wrong for a pure-logic package: dTransVal/dFlags/TransList/InDungeonBounds are world-grid state owned by vti7 (TransList decl verified dun_tile_data.hpp:97), not by the render/shader stack (66o7 framebuffer, b782 palettes, m18k Kage shader are render consumers of vision's output, not prerequisites).
2 local: Content verified: VisionRays[23][15] vision.cpp:22 ✓ exact; DoVision def at :51 with quadrant-mirror ray walk :66-100 ✓ (body ":98-120" — covers only the loop tail, cite drift); DoVision 4-callback signature vision.hpp:11-17 ✓ (function_ref×4); DoVisionFlags lighting.cpp:84-94 ✓; ProcessVisionList :437 ✓ (body ":431-468", ~6-line drift); `friendlyMode ? MAP_EXP_OTHERS : MAP_EXP_NONE` :466 ✓ exact; vision_test.cpp exists on disk ✓ (portable test source); AGENTS.md parity-gate row `test/fixtures/vision_rays.bin` matches this bean's golden ✓. Ledger anomaly for parent: docs/research/dg-id-map.json maps BOTH dg-task-255 and dg-task-256 → devilGo-9o2h (Kage bean 255 evidently merged into m18k=254) — batch-01 tree validation presumably covered it; noting only.
3 online: n/a.
4 solution: rewire deps to the logic-correct set (vti7 for grids; bodd stays if dLight interplay is asserted; drop 66o7/b782/m18k — they are downstream consumers); refresh DoVision/ProcessVisionList cites; keep ACs unchanged (all five are exact-value tests, no tolerance).
5 understanding: world/logic layer (dFlags visibility + TransList reveal) feeding render (m18k/pdpc) and automap — mis-wired as a render dependent; the VisionRays-golden + 4-player parity ACs are exactly the epic's L5 gate.
6 repro: `sed -n '22,24p;51,60p' devilutionX/Source/vision.cpp; grep -n "TransList" devilutionX/Source/levels/dun_tile_data.hpp; sed -n '437,470p' devilutionX/Source/lighting.cpp | grep -n "MAP_EXP\|friendlyMode"; ls devilutionX/test/vision_test.cpp`
7 fix: `beans query 'mutation { updateBean(id: "devilGo-9o2h", input: { blockedBy: ["devilGo-vti7"] }) { id blockedBy } }'` (check `beans query --schema` for the exact blockedBy mutation field); `beans update devilGo-9o2h --body-replace-old "- devilGo-gk5e (dungeon bounds, dTransVal array)" --body-replace-new "- devilGo-vti7 (world grids: dTransVal/dFlags/TransList/InDungeonBounds — dun_tile_data.hpp:97,103,109)"`; optional cite fix "vision.cpp:98-120 DoVision" → "vision.cpp:51-110 DoVision (def :51, quadrant mirror :66+)".
8 validation: 9o2h blockedByIds = [vti7] (or [vti7, bodd] if parent keeps the light-inventory link); render beans no longer gate it.
9 related: vti7 (correct dep), bodd/kx0l/pdpc/m18k (wire-declared, consumers not prerequisites), 66o7/b782.
VERDICT: NEEDS-WORK:wrong-deps — wire (5 render/light beans) contradicts body (gk5e) and both mis-target; correct blocker is vti7; 2 cite drifts; content itself verified exact.

Fix pass 2026-10-16: Dep-wiring corrected (fix pass): body-vs-wire contradiction resolved — render/light deps (gk5e path) wired.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).

ID MAP (fix2): all legacy dg-task-NNN references in this bean map to devilGo ids via docs/research/dg-id-map.json (canonical mapping in milestone devilGo-i56j). Operative deps = frontmatter (source of truth); prose refs are documentation.
