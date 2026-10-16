---
# devilGo-ohsu
title: Tick accumulator + GameLogic skeleton + headless mode
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-0
    - loop
created_at: 2026-10-15T16:35:16Z
updated_at: 2026-10-16T14:01:05Z
parent: devilGo-qna9
blocked_by:
    - devilGo-lfvi
---

# The 20 TPS heart

## Goal
Update()/Draw() split with accumulator mirroring nthread semantics; GameLogic() ordered no-op phases; headless runner.

## Research
- Source/diablo.cpp:1523-1568 GameLogic order; :3490 game_loop startup burst (nTickRate*3); Source/nthread.cpp:214-243 has_500ms_passed + >10 reset; :246 UpdateProgressToNextGameTick.
- Ebiten guidance /home/blaze/.agents/skills/ebiten/game-loop/SKILL.md; perf https://ebitengine.org/en/documents/performancetips.html

## Scope
internal/engine/tick.go (TickSource interface, real + mock), internal/game/game.go (Update accumulates, runs GameLogic stubs in order, clamps 10), cmd/dgoheadless.

## Acceptance criteria
- [ ] Mock-clock test: dt=16.67ms x60 → exactly 20 GameLogic calls, ProgressToNextGameTick values match precomputed series
- [ ] dt=500ms once → clamped to 10 ticks, then next Update realigns last_tick (no runaway)
- [ ] Headless main runs N ticks from flag, exits 0, prints tick count
- [ ] No allocations in Draw() path (assert with testing.AllocsPerRun on stub draw)

## Skills for coder
/home/blaze/.agents/skills/ebiten/game-loop/SKILL.md — Update/Draw contract and SetTPS decision (we use 60TPS + accumulator, documented why)
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — mock clock tests

## Dependencies & blockers
dg-task-001.

## Open questions
- Startup triple-tick burst: replicate exactly (yes, behind startup flag).


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK
