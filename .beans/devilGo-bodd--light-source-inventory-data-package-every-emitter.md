---
# devilGo-bodd
title: Light source inventory data package — every emitter radius/lifetime/flags (L3)
status: completed
type: task
priority: normal
tags:
    - phase-lighting
created_at: 2026-10-15T22:41:04Z
updated_at: 2026-11-07T18:31:05Z
parent: devilGo-97mn
blocked_by:
    - devilGo-n89r
    - devilGo-b782
    - devilGo-m18k
---

# Light source inventory data package

## Goal
Create the complete data-driven light emitter catalog (L3 table) as a Go package: every C++ AddLight (inventory ~2x undercount corrected per triage: include missiles.cpp callers, 23 sites) call site with its radius, lifetime rules, network sync needs, and source file:line. This feeds runtime light creation and the net command payloads.

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

## Libraries
- stdlib only (no third-party deps at this stage)

## Skills for coder
- \`~/.agents/skills/ponytail/SKILL.md\` — read fully BEFORE acting
- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/coding-rules/SKILL.md

## Dependencies & blockers
- devilGo-gk5e (dungeon bounds for in-bounds checks)
- devilGo-n89r (decoders)

## Open questions
OPEN-QUESTION RESOLUTION 2026-10-16: DECIDED: town torches — full grep of town.cpp at claim time is the first AC step (inventory = AddLight call sites in town.cpp, monsters.cpp, missiles.cpp, dead.cpp).
- Town torches: town.cpp AddLight calls not fully grep'd — add if found

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:content-defect — AddLight inventory undercounted ≈2× (missiles.cpp 23 vs listed 6), dead.cpp mislabel, tolerance AC vs no-approximation mandate, body-vs-wire dep mismatch.


## IMPL NOTES (triage 2026-10-16)
1 root: Lighting-wave data task (devilGo-bodd): catalog every AddLight emitter; draft; wire blockedBy [n89r, b782, m18k] vs body deps "gk5e + n89r" — mismatch (wire adds palette-registry + Kage shader, omits gk5e).
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


## Proof of Work

Upstream verified by reading, not assumed: /tmp/dvlx-pinned @
ca91d778fba92a1f680afb1fb2cdf09337ff1b42, Source/lighting.h:22-42,
Source/lighting.cpp:240-398, Source/engine/lighting_defs.hpp:7-9.

### Delivered

internal/light/light_list.go (new)
- MaxLights=32 (lighting_defs.hpp:7), NoLight=-1 (:9).
- Light/LightPosition verbatim (lighting.h:22-36).
- LightList{Lights, Active, Count} == Lights/ActiveLights/ActiveLightCount
  (lighting.h:40-42, lighting.cpp:34-36).
- NewLightList == InitLighting (lighting.cpp:240-252) incl. the iota identity
  permutation at :249.
- AddLight (:254-274) -> NoLight at the cap; AddUnLight (:276-288);
  ChangeLightRadius (:290-306); ChangeLightXY (:308-324); ChangeLightOffset
  (:326-345) with the unchanged-offset early return; ChangeLight (:347-364);
  ProcessLightList (:366-398).
- DoUnLight/DoLighting/TileHasAny(Solid) are injected as LightGrid rather
  than reached for: dLight lives in internal/diablo/levels, so wiring is a
  later bean. DoUnVision is a typed VisionUnlighter seam only — vision is NOT
  ported (separate bean).
- CloseLights resets the pool; documented as a Go-side addition because
  upstream re-runs InitLighting on level change (diablo.cpp:3495) without
  clearing Lights.

internal/light/registry.go (new)
- SourceType + LightSourceDef{Type, Radius, Flags, Cite} + Has.
- FlagAnimated/FlagPersistent/FlagNetworked.
- 20 rows covering all 30 AddLight call sites measured by grep, each citing
  file:line: player.cpp:2524, monster.cpp:3351, :898, missiles.cpp:1339,
  diablo.cpp:1506, objects.cpp:1483 via :4208/:4212/:4241, objects.cpp:1582/
  :1596/:1638, missiles.cpp:691/1613/1631/1726/1767/1876/1906/1987/2015/
  2216/2285/2502/2667/2699/2725/2770/3055/3078/3390/3470/3490/3599/3623,
  portal.cpp:59.
- GetLightSourceDef, LightSourceTypes, (*LightList).AddSourceLight.

### Name mapping (no invented API)

Mission names ChangeLightFade/ChangeLightOff/MakeLightList do not exist
upstream. Mapped to the real symbols by reading lighting.cpp: ChangeLightFade
-> ChangeLight, ChangeLightOff -> AddUnLight, MakeLightList ->
ProcessLightList; the mapping is recorded in the light_list.go header comment.

### Tests

internal/light/light_list_test.go (19 tests) + registry_test.go (9 tests).
Covered: first-slot id, stored fields, slot exhaustion -> NoLight, freed-slot
reuse, un-light clears once then the pool is quiet, NO_LIGHT guards on all
five mutators, radius round-trip, XY keeps radius, ChangeLight sets both,
offset unchanged -> no redraw, two-pass clear-then-draw order, swapped-in
light redrawn in the same pass, solid-tile skip, CloseLights reset, vision
seam, registry size >= 12 and == 20, per-row radius, per-row cites, three
flag groups, unknown type -> zero row, AddSourceLight radius.

RED proof the tests bite (each mutation reverted):
- registry SrcMissileBolt 5 -> 6: FAIL TestRegistryRadii
- drop the IsSolid skip: FAIL TestProcessLightListShortCircuitsSolidTile
- drop the i-- re-scan: FAIL TestProcessLightListDrawsSwappedInLight

### Gates (verbatim in s1.log)

build OK (rc=0); gofmt -s -l . empty (0 lines); scripts/lint.sh ./...
0 issues (simpurity: internal/game internal/diablo clean, 237 sources);
go vet ./... clean; go test ./internal/... 41 ok / 0 FAIL; beans-lint OK
(181 beans); just sqlc-check clean.

### Scope honored

No vision (separate bean), no renderer, no level/renderer wiring, no
monstdat/objdat loader change — those fields are cited, not loaded. No
Ebitengine import in internal/light.
