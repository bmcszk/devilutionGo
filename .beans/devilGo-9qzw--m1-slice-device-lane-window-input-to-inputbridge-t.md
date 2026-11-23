---
# devilGo-9qzw
title: 'M1 slice: device lane — window input to inputbridge to live kernel'
status: completed
type: task
priority: normal
created_at: 2026-11-20T12:19:10Z
updated_at: 2026-11-23T11:44:38Z
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
- [x] runWindowed assigns the real kernel; key/mouse events become
      bridge decisions; PlayerMoved events appear in the tick log during
      a real window session (evidence captured)
- [x] ESC/q quit preserved rc 0; windowless env still fails with the
      named error
- [x] Full gates green; RED first (a failing boot-wiring test or the
      honest falsification if the wiring test cannot fail unchanged —
      record which and why)

## Triage
Planner wave (wave-s6-plan.md s6e), filed 2026-11-20; blocked by the
bridge walk rows (s6b) and the retail town layout (s6c) — the demo walk
needs both.

## Libraries
ebiten (existing dep), stdlib.

## Skills for coder
- ~/.agents/skills/ponytail/SKILL.md — smallest landable diff
- ~/.agents/skills/ebiten/input/SKILL.md
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.agents/skills/tdd/SKILL.md
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md



## Progress log

- 2026-11-20T16:30Z: claimed — s6e dispatched under contract harness.s6e.md (worktree ftbu-s6, feat/m1-interactive-demo; base 59287ba carries loaders, walk rows, retail town, widget kit), status flip.


## Proof of Work

- Worker: task-dg-m1-s6e-coder-coder (111 tool calls) landed 242c806; review FAIL (F1 vacuous test + unproducible RED records, F2 second pass driver) → fix task-dg-m1-s6efix-coder-coder (116 tool calls) landed fc09e35 — pushed; parent verified ls-remote == HEAD.
- F1 resolved: device_lane_test.go rebuilt (5 real black-box tests, no build tag — gates run them); honest RED + mutation flips recorded (kernel unwired → boot assertion fails; handlers unwired → hero tile pinned; each output verbatim in the lane log `## s6efix`).
- F2 resolved: the window layer no longer calls Pass/Tick — the Game is the single pass driver (device_lane.go:32, window_ebiten.go:96); decisions buffer, the settled pass submits with the kernel's tick number.
- F3-F5 disclosures landed (direct (50,50) bootstrap raise named; mouse lane deferred; sim-vs-scene split + cadence recorded for later waves).
- Re-review: brainer PASS 8/8 (.hermes/lane/ftbu-m1/harness.s6e.re-review.md) — N1 stale-line-citation note recorded (not phantom; disclosed two-push amend).
- Parent gates at fc09e35: full suite 54 pkgs ok, cmd suite ok, trail f1ae95a0… unchanged.
- Keyboard arrows walk the hero in a real window session; PlayerMoved events land in the tick log through the kernel's real pass.
