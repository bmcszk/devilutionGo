---
# devilGo-v4lb
title: Port A* pathfinder + crawl
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-0
    - engine
created_at: 2026-10-15T16:35:16Z
updated_at: 2026-10-17T11:06:08Z
parent: devilGo-qna9
blocked_by:
    - devilGo-m2x7
---

# Pathfinding port

## Goal
A* with devilutionX tie-breaking so walk paths match monster/player movement.

## Research
- Source/engine/path.cpp:33-34 costs (axis 100, diagonal 101), :37 MaxPathNodes 1024, :184 FindPath signature with canStep/posOk closures, :178 GetPathDirection PathDirections table {5,1,6,2,0,3,8,4,7}, :285 FindClosestValidPosition via Crawl, frontier comparator tie-breaks (lower heuristic, diagonal first, coordinate order) at :207-236.
- Source/crawl.hpp spiral walk; tests test/path_test.cpp, test/crawl_test.cpp, benchmarks path_benchmark.cpp, crawl_benchmark.cpp.

## Scope
internal/engine/path.go + crawl.go; FindPath(canStep, posOk func, start, dest, path []int8) int; direction encoding identical to int8 PathDirections.

## Acceptance criteria
- [ ] All ported test/path_test.cpp cases pass
- [ ] Path direction bytes match C++ for 100 randomized grids (fixture)
- [ ] Benchmark: 10k FindPath calls on 40x40 map < 50ms (parity-ish with path_benchmark)


## Libraries
- github.com/SolarLune/paths — A* + crawl; port C++ path logic 1:1 on top, lib for grid primitives only

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — algorithm port fidelity, closures for canStep/posOk
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — bench + table tests

### game-dev skills (2026-10-16 audit)
- /home/blaze/.agents/skills/game-developer/SKILL.md — ECS/gameplay architecture patterns for AI movement
- /home/blaze/.agents/skills/game-development/2d-games/SKILL.md — 2D tile pathing/collision principles

## Dependencies & blockers
devilGo-m2x7 (RNG not strictly needed but engine pkg cohesive).

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK


## IMPL NOTES (triage 2026-10-16)
1 root: A* + crawl with devilutionX tie-breaking; status `draft`, parent qna9, blocked_by m2x7 (frontmatter ✓; body says devilGo-m2x7 stale, with honest note "RNG not strictly needed but engine pkg cohesive").
2 local: all cited constants verified — path.cpp:33-34 PathAxisAlignedStepCost=100/PathDiagonalStepCost=101 (:32-33), MaxPathNodes 1024 at :37, FindPath(canStep,posOk,…) at :184, PathDirections {5,1,6,2,0,3,8,4,7} at :180 (bean says :178, ±2 line drift), FindClosestValidPosition-via-Crawl at :285, frontier comparator tie-breaks (lower h, diagonal-first, coordinate order) within :207-236 ✓; crawl at Source/crawl.hpp (correct path, NOT engine/) + crawl.cpp; test/path_test.cpp (273 ln), crawl_test.cpp (62 ln), path_benchmark.cpp, crawl_benchmark.cpp all exist. ACs testable (ported tests, 100-grid fixture, 10k<50ms bench). Layer note: internal/engine/path.go vs AGENTS.md table placing A* under internal/game/grid — inherits lfvi layout decision (C++ has it in Source/engine, so engine placement is fidelity-faithful either way).
3 online: n/a (no external claims).
4 solution: ready after promotion; cleanup: devilGo-m2x7 → devilGo-m2x7.
5 understanding: engine-layer algorithm core shared by player walk + monster AI; direction int8 encoding must match PathDirections byte-for-byte for animation parity.
6 repro: `beans show devilGo-v4lb --body-only`; `sed -n '30,38p;178,190p' …/Source/engine/path.cpp`; `ls …/Source/crawl.hpp …/test/path_test.cpp …/test/crawl_benchmark.cpp`; `sed -n '13,16p' .beans/devilGo-v4lb*.md`.
7 fix: `beans update devilGo-v4lb --body-replace-old "devilGo-m2x7 (RNG not strictly needed but engine pkg cohesive)." --body-replace-new "devilGo-m2x7 (RNG not strictly needed but engine pkg cohesive)."`
8 validation: `beans show devilGo-v4lb --body-only | grep dg-task` → empty.
9 related: m2x7 (blocker), lfvi (tree), qna9 (parent); no duplicate pathfinding bean in ledger (scan confirmed only v4lb).
VERDICT: TRIAGED-OK
