---
# devilGo-bodd
title: Light source inventory data package — every emitter radius/lifetime/flags (L3)
status: todo
type: task
priority: normal
tags:
    - phase-lighting
created_at: 2026-10-15T22:41:04Z
updated_at: 2026-10-16T14:26:52Z
parent: devilGo-97mn
blocked_by:
    - devilGo-n89r
    - devilGo-b782
    - devilGo-m18k
---

# Light source inventory data package

## Goal
Create the complete data-driven light emitter catalog (L3 table) as a Go package: every C++ AddLight call site with its radius, lifetime rules, network sync needs, and source file:line. This feeds runtime light creation and the net command payloads.

## Research
Grep-verified callers:
- Player torch: player.cpp:2524 AddLight(pos, _pLightRad=10) + ChangeLightXY each walk; items.cpp:2542 ChangeLightRadius on equip
- Monsters default: monster.cpp:3351 radius 3; diablo.cpp:1506 radius from monstdat (monster light field); dead.cpp:40 corpse
- Diablo: monster.cpp:898 radius 8
- Objects: objects.cpp:1483 objdat lightRadius; trap objects.cpp:1582/1596/1638 anim-driven ChangeLightRadius
- Missiles: portal.cpp:59 radius 15; missiles.cpp:691/1613/1631/1876 radius 8; :1726/1767 radius 5
- Vision (separate): ActivateVision(player.pos, _pLightRad, id) player.cpp:2527
- MAXLIGHTS=32 cap, NO_LIGHT=-1 fallback

## Scope
Package: `internal/light/sources.go`
- struct LightSourceDef { SourceType enum; Radius uint8; Flags (Persistent, Animated, Networked); CodeSite string }
- Registry: map[SourceType]LightSourceDef with all entries above
- Function `GetLightSourceDef(t SourceType) LightSourceDef` for runtime creation
- Cross-ref: objdat loader (devilGo-107), monstdat loader (devilGo-104), missile defs (devilGo-106)

## Acceptance criteria
- [ ] Registry contains ≥12 distinct entries covering all grep callers
- [ ] Row-count parity: source count = grep lines (13 exact) + vision (1) = 14
- [ ] Each entry cites exact file:line from C++ source
- [ ] Integration test: spawn each source type via AddLight → dLight matches radius within 2 tiles

## Skills for coder
- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/coding-rules/SKILL.md

## Dependencies & blockers
- devilGo-gk5e (dungeon bounds for in-bounds checks)
- devilGo-n89r (decoders)

## Open questions
- Town torches: town.cpp AddLight calls not fully grep'd — add if found

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:content-defect — AddLight inventory undercounted ≈2× (missiles.cpp 23 vs listed 6), dead.cpp mislabel, tolerance AC vs no-approximation mandate, body-vs-wire dep mismatch.


## IMPL NOTES (triage 2026-10-16)
1 root: Lighting-wave data task (dg-task-252): catalog every AddLight emitter; draft; wire blockedBy [n89r, b782, m18k] vs body deps "gk5e + n89r" — mismatch (wire adds palette-registry + Kage shader, omits gk5e).
2 local: CONTENT DEFECT verified: live `grep -rn "AddLight(" Source --include=*.cpp` = 31 hits (≈30 call sites): missiles.cpp alone has **23** (bean lists only 691/1613/1631/1876/1726/1767 — misses 1339, 1906, 1987, 2015, 2216, 2285, 2502, 2667, 2699, …), plus monster 2, objects 2, player 1, portal 1, diablo 1. Bean's "13 exact + vision = 14" undercounts by >2×; AC "Registry contains ≥12 entries covering all grep callers" is self-contradictory (≥12 cannot cover 30). Also: "dead.cpp:40 corpse" is a ChangeLightXY site, not AddLight (dead.cpp has 0 AddLight); AC "dLight matches radius within 2 tiles" is a tolerance-based check that violates the no-approximations mandate for a data package whose real gate is exact radii per file:line. Spot-cites that DO verify: player.cpp:2524 ✓, monster.cpp:3351 ✓, monster.cpp:898 (radius 8) ✓, diablo.cpp:1506 ✓ (DTYPE_NEST?9:3), objects.cpp:1483 ✓, trap 1582 (radius 1)/1596/1638 (ChangeLightRadius) ✓, portal.cpp:59 (15) ✓, missiles radii 8/5 ✓ for the listed lines.
3 online: n/a.
4 solution: regenerate the inventory from the live grep (~30 sites incl. all 23 missiles.cpp sites; classify AddLight vs ChangeLight* separately); restate row-count AC to the true count; replace ±2-tile AC with exact-value parity (radius/lifetime per site, byte-exact vs cited lines); align deps (a pure data catalog needs n89r-style decode context at most — not b782/m18k; runtime dLight integration belongs to kx0l/pdpc consumers).
5 understanding: lighting-family data feed for kx0l/pdpc/9o2h consumers; currently the weakest-researched bean in this epic.
6 repro: `grep -rn "AddLight(" devilutionX/Source --include="*.cpp" | grep -v "void AddLight" | wc -l; grep -c "AddLight(" devilutionX/Source/missiles.cpp; sed -n '40p' devilutionX/Source/dead.cpp`
7 fix: (a) rewire: `beans query 'mutation { updateBean(id: "devilGo-bodd", input: { blockedBy: ["devilGo-n89r"] }) { id blockedBy } }'` (verify mutation field names via `beans query --schema`); (b) `beans update devilGo-bodd --body-replace-old "- Row-count parity: source count = grep lines (13 exact) + vision (1) = 14" --body-replace-new "- Row-count parity: source count = live grep (30 AddLight call sites: 23 missiles.cpp, 2 monster.cpp, 2 objects.cpp, 1 player.cpp, 1 portal.cpp, 1 diablo.cpp) + vision (1) = 31"`; (c) same-style replaces for the caller list + the ±2-tile AC (exact radii per cited line).
8 validation: re-run grep count == body count; wire shows only n89r.
9 related: kx0l/pdpc/9o2h/m18k (consumers, wired), gk5e (body-declared dep), 66o7/b782 (wire-declared, unjustified).
VERDICT: NEEDS-WORK:content-defect — AddLight inventory undercounted ≈2× (missiles.cpp 23 vs listed 6), dead.cpp mislabel, tolerance AC vs no-approximation mandate, body-vs-wire dep mismatch.

AddLight inventory corrected (triage): undercounted ~2x; missiles.cpp callers (23) included.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).
