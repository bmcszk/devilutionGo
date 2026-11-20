---
# devilGo-9qzw
title: 'M1 slice: device lane — window input to inputbridge to live kernel'
status: draft
type: task
priority: normal
created_at: 2026-11-20T12:19:10Z
updated_at: 2026-11-20T12:19:16Z
parent: devilGo-k736
blocked_by:
    - devilGo-9ukg
    - devilGo-t8dw
---

# M1 slice: device lane — window input → inputbridge → live kernel

## Goal
The windowed boot wires NO kernel and reads only ESC/q (cmd/devilutionGo/
window_ebiten.go:38-39, window_scene.go:48-57). This slice: the ebiten
Update loop reads real keys/mouse and produces decisions into
inputbridge.Session (the one submit contract, internal/diablo/
inputbridge/session.go:15-69) at the tick boundary; runWindowed wires
the real kernel (world.NewSessionGate + handler register + streams).
Keyboard arrows walk the hero; the window layer stays the only place
raw codes appear (v9xy note in-tree).

## Research
- Plan: wave-s6-plan.md (facts cited file:line at master ce342c7).
- The sim cadence already runs from real time (frameClock.step 50ms,
  window.go:10-27); only the command lane is missing.
- Presentation reads events; it NEVER writes sim state (ADR-018 law 2).

## Scope
- cmd/devilutionGo (window loop, boot wiring), internal/game (kernel
  plumbing on the Game struct as needed); NO world/sim changes (s6b
  landed the handlers this consumes).

## Acceptance criteria
- [ ] runWindowed assigns the real kernel; key/mouse events become
      bridge decisions; PlayerMoved events appear in the tick log during
      a real window session (evidence captured)
- [ ] ESC/q quit preserved rc 0; windowless env still fails with the
      named error
- [ ] Full gates green; RED first (a failing boot-wiring test or the
      honest falsification if the wiring test cannot fail unchanged —
      record which and why)

## Triage
Planner wave (wave-s6-plan.md s6e), filed 2026-11-20; blocked by the
bridge walk rows (s6b) and the retail town layout (s6c) — the demo walk
needs both.

## Libraries
ebiten (existing dep), stdlib.

## Skills for coder
- ~/.agents/skills/ebiten/input/SKILL.md
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.agents/skills/tdd/SKILL.md
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md
