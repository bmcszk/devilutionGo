---
# devilGo-qna9
title: 'Phase 0: Repo scaffold, engine skeleton, CI gates'
status: draft
type: epic
tags:
    - plan:epic
    - phase-0
created_at: 2026-10-15T16:37:32Z
updated_at: 2026-10-15T16:37:32Z
parent: devilGo-i56j
---

# Phase 0 foundations

## Goal
Stand up github.com/bmcszk/devilutionGo (module exists, empty) as a buildable Ebitengine project with a fixed-20TPS tick accumulator skeleton, headless build mode, lint/vet/test CI, and the golden-test harness scaffolding that every later epic depends on.

## Research
- devilutionX loop facts: gnTickDelay=50ms (Source/diablo.cpp:146), nthread_has_500ms_passed catch-up rule >10 ticks (Source/nthread.cpp:214-243), GameLogic fixed order (Source/diablo.cpp:1523-1568), tick-rate slider 20-50 (Source/gamemenu.cpp:274-277). Full study: /tmp/dg-requirements.md §1, Appendix E/F/M.
- Ebitengine Update/Draw/TPS + performance rules: https://ebitengine.org/en/documents/performancetips.html and /home/blaze/.agents/skills/ebiten/game-loop/SKILL.md
- RNG to port day one: Source/engine/random.cpp:16-96, random.hpp:55-105 (LCG 0x015A4E35, GenerateRnd high-bits trick, INT_MIN case).
- A* pathfinder: Source/engine/path.cpp:184-283 (costs 100/101, MaxPathNodes 1024).

## Scope
- cmd/dgo + cmd/dgoheadless entrypoints; internal/engine (tick accumulator, random, path, direction, animation fixed-point); internal/game ebiten.Game impl; justfile targets mirroring make vet test lint; .golangci already present; headless build tag; testdata/ skeleton.
- NOT in scope: any gameplay, assets, rendering beyond a test rect.

## Acceptance criteria
- [ ] `go build ./...` green on linux/amd64 and windows/amd64 without CGO
- [ ] `go vet ./...` and `golangci-lint run` clean (repo .golangci.yml)
- [ ] `go test ./...` passes with ported random vectors (from test/random_test.cpp) and path cases (test/path_test.cpp)
- [ ] Ebiten window 640x480 fixed Layout opens; Update accumulator runs exactly 20 ticks/sec at default gnTickDelay=50ms (asserted via mock clock test, ±0 ticks over 10s)
- [ ] Headless build (tag `headless`) compiles same logic package with null renderer and runs 1000 ticks in <2s
- [ ] Tick accumulator clamps catch-up at 10 ticks like nthread.cpp:233-243 (unit test with synthetic dt)
- [ ] ProgressToNextGameTick computed as 0..128 fraction (unit test at dt=0,25,50ms)

## Skills for coder
/home/blaze/.agents/skills/ebiten/game-loop/SKILL.md — Update/Draw/TPS accumulator pattern this skeleton implements
/home/blaze/.agents/skills/go-setup/SKILL.md — module layout decisions for the internal/ tree
/home/blaze/.agents/skills/go-dev/SKILL.md — error handling and package-boundary rules
/home/blaze/.agents/skills/coding-rules/SKILL.md — top-level functions read as intent; minimal diffs
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — table-test structure for the accumulator/RNG vectors

## Dependencies & blockers
None (root epic). Blocks: dg-epic-02, dg-epic-03, dg-epic-12.

## Open questions
- Mock-clock injection point for Update tests (decide: time source interface in engine.TickSource).
