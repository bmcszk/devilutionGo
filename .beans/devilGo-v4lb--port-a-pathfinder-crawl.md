---
# devilGo-v4lb
title: Port A* pathfinder + crawl
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-0
    - engine
created_at: 2026-10-15T16:35:16Z
updated_at: 2026-10-15T16:45:27Z
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

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — algorithm port fidelity, closures for canStep/posOk
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — bench + table tests

## Dependencies & blockers
dg-task-002 (RNG not strictly needed but engine pkg cohesive).

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.
