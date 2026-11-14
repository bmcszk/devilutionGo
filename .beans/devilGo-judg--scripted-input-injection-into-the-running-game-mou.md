---
# devilGo-judg
title: Scripted input injection into the running game (mouse -> kernel commands -> world state)
status: in-progress
type: task
priority: critical
created_at: 2026-11-14T11:13:02Z
updated_at: 2026-11-14T11:16:58Z
parent: devilGo-wq2s
---

## Goal

Scripted input events go STRAIGHT into the running game from outside the
game package: a test harness (later a debug flag) feeds mouse events like a
driver would, they become kernel commands, the gate passes them, the hero
moves. This is the "element tester" surface the user asked for.

## Research (parent, 2026-11-14, verified)

- Two proven paths feed input as commands: DemoReplay.MouseButton
  (demo.go:65, kind KindMouseButton) and ClickRouter.Command (click.go:35)
  which resolves pointer position -> tile via the host-supplied TileAt func
  and raises walk/attack. Both are PUBLIC entry points a driver can call.
- The kernel gate path is real: RegisterMovement (movement.go:8) +
  Kernel.Submit (kernel.go:52); StepTo (state.go:30) mutates the hero tile.
- ebiten key/mouse events are NOT needed for this bean — the injection API
  must be driver-agnostic (plain Go function calls), so the future ebiten
  window becomes just another driver.

## Acceptance criteria

- [ ] A small exported injection surface (driver API) accepts MouseButton/Move events with screen coords and feeds them into the running Game loop
- [ ] e2e: inject a click at a walkable tile while the loop runs -> hero tile changes to it within the driven ticks; log line names the command kind + gate verdict
- [ ] A blocked case is pinned (unwalkable tile -> gate refusal logged, hero unchanged)
- [ ] TDD: red -> green recorded in PoW

## Skills for coder

- `~/.agents/skills/coding-rules/SKILL.md` — read fully before editing
- `~/.agents/skills/go-tests/SKILL.md`
- `~/.agents/skills/go-dev/SKILL.md`

- `~/.agents/skills/ponytail/SKILL.md` — smallest change that fully solves the task

## Libraries

- stdlib only (slog, image/png, testing); no new third-party deps

## Triage

Parent mini-triage 2026-11-14 (session _chat, planning mode per user directive): seams verified against the tree at file:line (see Goal/Research sections); scope split = boot/logs (u1hd), injection (judg), screenshot+recipe (jlrz); brainer re-review rides the PR review round per lane law (read-first verdict + operator checks). Skills paths stat-verified on disk.

## Progress log — smoke slice (2026-11-14)

- RED: `h.RunTicks` / `h.Click` undefined (vet failure), then the blocked
  case failed on the missing verdict line before the verdict path landed.
- GREEN (commit 2aadae5): SmokeHarness.Click feeds a
  demo.MouseButtonRecord through ClickRouter (fixed 64px-per-tile camera,
  TileAt = x/64, y/64) into kernel.Submit; RunTicks drives the loop then
  logs a "click verdict" line naming kind + applied/refused. The
  walkability refusal happens in the movement handler (kernel law 2: no
  change, no event), so the harness publisher tracks the PlayerMoved
  event and surfaces applied vs refused — the gate-pass-then-no-op can
  no longer stay silent.
- TestSmokeInjectedClickMovesHero: hero tile changes to the clicked
  adjacent tile; log names PlayerMove + accepted/applied.
  TestSmokeInjectedClickOnUnwalkableRefused: SOLData[0]=TileSolid on the
  target tile -> hero unchanged, verdict line names PlayerMove + refused.
- ACs satisfied: exported driver-agnostic injection surface (plain Go
  calls, no ebiten); walkable-click e2e with kind+verdict log; blocked
  case pinned; TDD red -> green recorded.
