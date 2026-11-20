---
# devilGo-prcc
title: 'M1 gate: AC-3 frame budget — 60fps p95 bench on town + cathedral render'
status: completed
type: task
priority: normal
created_at: 2026-11-19T22:32:39Z
updated_at: 2026-11-20T02:05:18Z
parent: devilGo-1n0a
blocked_by:
    - devilGo-6851
    - devilGo-t8dw
---

---
# devilGo-prcc
title: 'M1 gate: AC-3 frame budget — 60fps p95 bench on town + cathedral render'
status: completed
type: task
priority: normal
created_at: 2026-11-19T22:32:39Z
updated_at: 2026-11-20T01:35:01Z
parent: devilGo-1n0a
blocked_by:
    - devilGo-6851
    - devilGo-t8dw
---

---
# devilGo-prcc
title: 'M1 gate: AC-3 frame budget — 60fps p95 bench on town + cathedral render'
status: completed
type: task
priority: normal
created_at: 2026-11-19T22:32:39Z
updated_at: 2026-11-20T01:33:52Z
parent: devilGo-1n0a
blocked_by:
    - devilGo-6851
    - devilGo-t8dw
---

# M1 gate: AC-3 frame budget — 60fps p95 bench on town + cathedral render

## Goal
A frame-timing benchmark over the render pipeline (world draw + palette
presenter + sprite/panel stages) proves the AC-3 budget: p95 frame time
under 16.67ms (60fps) for the town + cathedral slice on this box's
mid-hardware baseline, with the measured budget recorded in this bean.

## Research
- Engine budgets precedent: repo bench tests already carry budget pins
  (the CI bench job pattern, green on #77 run 44435571799).
- The render path: ebiten screen Image + index framebuffer palette
  lookup pass (the 66o7 lane) + world draw; the bench freezes the scene
  at the parity seed.

## Scope
- internal/integration (bench) — rendering packages only through their
  public entry; no render-instrument rewrites.

## Acceptance criteria
- [x] Bench exists, runs the full frame pipeline at the parity scene,
      and reports per-stage + total p95
- [x] The 16.67ms p95 budget asserted (budget measured first, then
      pinned; the observed p95 recorded in the bean)
- [x] Full gates green at HEAD through pre-commit + pre-push including
      the bench (short-run mode locally)

## Triage
Planner wave: delegated brainer plan pass + parent verification, 2026-11-19.
Plan: .hermes/lane/ftbu-m1/wave-s5-plan.md. Blocked-by: the verb slices.

## Libraries
stdlib testing/bench + the existing ebiten render path; no new deps.

## Skills for coder
- ~/.agents/skills/ebiten/game-loop/SKILL.md — frame pipeline
- ~/.agents/skills/ebiten/rendering/SKILL.md — draw path surfaces
- ~/.agents/skills/go-tests/SKILL.md — bench + budget pin design
- ~/.agents/skills/coding-rules/SKILL.md — style + gate honesty
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md — lane laws



## Progress log

- 2026-11-20T01:45Z: promoted (triage: planner plan + parent dep verification; base now carries verb slices + AC-2 matrix) and claimed — AC-3 dispatched under contract harness.ac3.md (worktree ftbu-s5, feat/m1-create-pack), status flip.


## Proof of Work

- Contract: .hermes/lane/ftbu-m1/harness.ac3.md (require_edit harness.s5a.log, sentinel M1-REAL-L1-AC3-DONE). Worker: task-dg-m1-ac3-coder-coder (262 tool calls), branch feat/m1-create-pack @ bf0833f — pushed; parent verified ls-remote == HEAD, sentinel present.
- Scene scope (recorded per review finding 1): the bench measures the CATHEDRAL parity scene (Cathedral L1, seed 2588, hero at entry 77,46, radius-10 dLight cone, committed CEL frames); town is covered by pipeline identity, not a measured town scene. Follow-up filed: the town-scene bench row bean (under Testing & parity, blocked by this bean).
- Measured p95 over 128 frames (budget 16.670ms, parent re-ran): world 4.416-4.476ms, panel 0.002ms, text 0.029-0.030ms, present 0.945-1.063ms, TOTAL 5.273-5.408ms — 3.1x headroom; the p95 assertion is a real failing gate, not a log line.
- RED/falsification narration verified in the lane log (contract step 1); trails byte-identical across worker runs; no suppressions, scope integration-only (2 new files, +1049).
- Cross-review: brainer APPROVE 8/8 (.hermes/lane/ftbu-m1/harness.ac3.review.md); findings 2 (advisory headroom pin, later wave), 3 (GetDunFrame third copy near threshold — render-lane backlog), 4 (disclosed CEL remap + hero-sprite limits) recorded here as the named residual.
- Residual: town-scene row (the town-scene bench row bean); advisory pin + render dedup backlog named above.



## Proof of Work (addendum — CI gate fix)

- CI on the wave PR red at bf0833f-class head (run 44656997769): TestM1FrameBudgetP95 p95 21.138ms (world 16.960ms) on the 2-core shared runner vs 16.670ms — the review's finding 2 realized as a hardware-class mismatch (AC scopes 16.67ms to mid hardware; CI runner is a different class).
- Fix commit c129cb8 (worker task-dg-m1-s5fix-coder-coder, contract harness.s5fix.md): budget selected by runner class via the CI env — strict 16.670ms stays the local/declared-baseline gate; CI carries the recorded factor 1.6 (26.672ms) with the failing run cited in the test comment. BOTH assertions real: no skip, no build tag, no gate-file edit; parent found no suppression markers in the diff; parent re-ran the local strict gate (PASS total 5.317ms) + full suite (54 ok) + lint.
- CI green at final PR head c129cb8 (run 44657848445: ci 1m51s, bench 37s); PR marked ready for review after mergeable=CLEAN + checks green (ready-flip law).
