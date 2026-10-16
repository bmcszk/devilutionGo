---
# devilGo-qna9
title: 'Phase 0: Repo scaffold, engine skeleton, CI gates'
status: draft
type: epic
priority: normal
tags:
    - plan:epic
    - phase-0
created_at: 2026-10-15T16:37:32Z
updated_at: 2026-10-16T14:06:45Z
parent: devilGo-i56j
---

# Phase 0 foundations

## Goal
Stand up github.com/bmcszk/devilutionGo (module exists, empty) as a buildable Ebitengine project with a fixed-20TPS tick accumulator skeleton, headless build mode, lint/vet/test CI, and the golden-test harness scaffolding that every later epic depends on.

## Research
- devilutionX loop facts: gnTickDelay=50ms (Source/diablo.cpp:146), nthread_has_500ms_passed catch-up rule >10 ticks (Source/nthread.cpp:214-243), GameLogic fixed order (Source/diablo.cpp:1523-1568), tick-rate slider 20-50 (Source/gamemenu.cpp:274-277). Full study: docs/research/dg-requirements.md §1, Appendix E/F/M.
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

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK


## IMPL NOTES (triage 2026-10-16)
1 root: First-work epic: module tree, 20TPS tick accumulator, RNG port, A*, headless tag, CI. Root epic (no blockers) — matches wiring (blockedBy: []).
2 local: Cites verified live: Source/diablo.cpp:146 `gnTickDelay = 50` ✓; nthread.cpp:214-243 catch-up `> gnTickDelay * 10` ✓; random.cpp:16-20 LCG 0x015A4E35 ✓; path.cpp:32-37 costs 100/101, MaxPathNodes 1024 ✓; gamemenu.cpp:274-277 tick slider ✓; GameLogic diablo.cpp:1523 ✓. All 5 skills paths exist on disk ✓. go.mod/justfile/.golangci.yml present in devilutionGo ✓.
3 online: ebitengine performance-tips URL cited — previously verified in plan wave; not re-fetched (no decision hinges on it).
4 solution: none blocking. Minor: cmd naming `cmd/dgo`/`cmd/dgoheadless` vs AGENTS.md layout `cmd/devilutionGo(-headless)`; "testdata/" vs "test/fixtures" — align at implementation. dg-epic label cleanup optional.
5 understanding: engine layer foundation everything else compiles against; correct as first unblocked epic (children lfvi/m2x7 already `todo`).
6 repro: `sed -n '146p' Source/diablo.cpp`; `sed -n '214,243p' Source/nthread.cpp | grep -n "10"`; `ls /home/blaze/.agents/skills/ebiten/game-loop/SKILL.md`
7 fix: none required. Optional: `beans update devilGo-qna9 --body-replace-old "cmd/dgo + cmd/dgoheadless" --body-replace-new "cmd/devilutionGo + cmd/devilutionGo-headless (per AGENTS.md layout)"`
8 validation: re-read body; ACs unchanged and testable (mock-clock 20TPS, catch-up clamp, 0..128 fraction — all unit-testable).
9 related: children lfvi, m2x7, v4lb, ohsu, xcdz (batch-02); blocks 5wsb (wired).
VERDICT: TRIAGED-OK
