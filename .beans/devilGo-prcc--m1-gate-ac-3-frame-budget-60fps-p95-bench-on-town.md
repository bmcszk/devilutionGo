---
# devilGo-prcc
title: 'M1 gate: AC-3 frame budget — 60fps p95 bench on town + cathedral render'
status: draft
type: task
priority: normal
created_at: 2026-11-19T22:32:39Z
updated_at: 2026-11-19T22:33:24Z
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
- [ ] Bench exists, runs the full frame pipeline at the parity scene,
      and reports per-stage + total p95
- [ ] The 16.67ms p95 budget asserted (budget measured first, then
      pinned; the observed p95 recorded in the bean)
- [ ] Full gates green at HEAD through pre-commit + pre-push including
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
