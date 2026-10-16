---
# devilGo-ohsu
title: Tick accumulator + GameLogic skeleton + headless mode
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-0
    - loop
created_at: 2026-10-15T16:35:16Z
updated_at: 2026-10-16T14:16:44Z
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


## IMPL NOTES (triage 2026-10-16)
1 root: Update()/Draw() split w/ accumulator mirroring nthread; GameLogic ordered no-op phases; headless runner. Status `draft`, parent qna9, blocked_by lfvi (frontmatter ✓; body says dg-task-001 stale).
2 local: all citations verified — diablo.cpp:146 gnTickDelay=50 ✓ (:148 region), GameLogic() body at :1524-1564 (claimed 1523-1568, ✓ order: ProcessInput→Players→Monsters|Towners→Objects→Missiles→Items→Light→Vision→sound→triggers→quests→redraw→pfile), game_loop startup burst `nTickRate * 3` at :3492 ✓, nthread.cpp:214-243 has_500ms_passed >10-tick reset ✓ (incl. multiplayer no-reset caveat inside cited range), UpdateProgressToNextGameTick at :246 with fraction = ticksAdvanced*128/gnTickDelay clamped 0..128 ✓ (baseValueFraction=128 at animationinfo.h:114). Skills exist (ebiten/game-loop, go-fluent-testing). ACs testable (mock-clock 20-tick series, 500ms→clamp-10, headless N ticks exit 0, AllocsPerRun on Draw). Note: cmd/dgoheadless naming inherits lfvi decision.
3 online: ebiten.org performance-tips URL — standard doc, low risk; verify at implementation time.
4 solution: ready after promotion; cleanup: dg-task-001 → devilGo-lfvi. Consider one-line note that the >10 reset is single-player semantics (MP no-reset rule lands with netcode, bean 4jo7).
5 understanding: the loop is the port's heartbeat: 20TPS fixed sim + 60fps interpolated draw via ProgressToNextGameTick; everything deterministic hangs off this cadence.
6 repro: `beans show devilGo-ohsu --body-only`; `sed -n '1523,1568p;3488,3496p' …/Source/diablo.cpp`; `sed -n '214,258p' …/Source/nthread.cpp`; `grep -n baseValueFraction …/Source/engine/animationinfo.h`; `sed -n '13,16p' .beans/devilGo-ohsu*.md`.
7 fix: `beans update devilGo-ohsu --body-replace-old "dg-task-001." --body-replace-new "devilGo-lfvi."`
8 validation: `beans show devilGo-ohsu --body-only | grep dg-task` → empty.
9 related: lfvi (blocker), xcdz (blocked by this), 4jo7 (netcode turn engine — related cadence port, NOT a duplicate), qna9 (parent).
VERDICT: TRIAGED-OK

---

## Verdict roll-up
devilGo-m2x7: TRIAGED-OK
devilGo-lfvi: DECISION-NEEDED (package layout + binary naming: AGENTS.md three-layer vs Appendix E 14-package — cascades ledger-wide)
devilGo-xcdz: TRIAGED-OK
devilGo-v4lb: TRIAGED-OK
devilGo-ohsu: TRIAGED-OK

BATCH-02-COMPLETE


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.
