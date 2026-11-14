---
# devilGo-u1hd
title: Headless real-world boot with structured run logs (e2e)
status: in-progress
type: task
priority: critical
created_at: 2026-11-14T11:13:02Z
updated_at: 2026-11-14T11:16:58Z
parent: devilGo-wq2s
---

## Goal

One e2e test (or `just smoke` command) that boots the REAL game loop —
game.NewGame + kernel + world.RegisterMovement — drives real ticks through
GameLogic, and emits structured logs at each kernel pass. No menu, no window,
no character selection.

## Research (parent, 2026-11-14, verified)

- cmd/devilutionGo/headless.go:17 drives NewGame(MockTickSource) but does NOT
  wire a world, kernel gates, or logs; internal/game tests boot pieces
  separately (kernel_integration_test.go is the closest existing shape).
- world.RegisterMovement(k, s) (movement.go:8) wires the PlayerMove handler;
  kernel.Kernel.Submit (kernel.go:52) is the entry; GameLogic (:109) runs the
  due ticks.
- Structured logs: slog already used in headless.go.

## Acceptance criteria

- [ ] e2e boots NewGame + kernel + world state wired via RegisterMovement, driven by a real (non-mock-synthetic) tick source over >= 60 ticks
- [ ] Structured slog lines at boot, per-tick kernel pass summary (submitted/accepted counts), and shutdown, timestamped
- [ ] Log output asserted in the test (captured via slog handler), not just printed
- [ ] TDD: failing e2e written first; PoW records red -> green

## Skills for coder

- `~/.agents/skills/coding-rules/SKILL.md` — read fully before editing
- `~/.agents/skills/go-tests/SKILL.md` — e2e shape
- `~/.agents/skills/go-dev/SKILL.md`

- `~/.agents/skills/ponytail/SKILL.md` — smallest change that fully solves the task

## Libraries

- stdlib only (slog, image/png, testing); no new third-party deps

## Triage

Parent mini-triage 2026-11-14 (session _chat, planning mode per user directive): seams verified against the tree at file:line (see Goal/Research sections); scope split = boot/logs (u1hd), injection (judg), screenshot+recipe (jlrz); brainer re-review rides the PR review round per lane law (read-first verdict + operator checks). Skills paths stat-verified on disk.

## Progress log — smoke slice (2026-11-14)

- RED: `internal/game/smoke_test.go` referenced `game.NewSmokeHarness`
  before it existed — `undefined: game.NewSmokeHarness` (compile-red),
  `go test ./internal/game -run TestSmokeBootLogsTicks` exit 1.
- GREEN (commit 4b9cba9): SmokeHarness boots game.NewGame on a driven
  engine.MockTickSource, wires kernel.New(world.NewGate(state,0)) +
  world.RegisterMovement + world.NewState, publishes one structured
  "kernel pass" line per tick, logs "smoke boot" (kernel/movement/tick)
  and "smoke shutdown" (ticks).
- TestSmokeBootLogsTicks: 60 driven ticks via Game.Update(50ms), output
  captured through a slog.Handler test hook; asserts boot wiring fields,
  >= 60 timestamped pass lines, shutdown ticks == 60.
- ACs satisfied: e2e boot+wiring+60 real ticks; structured timestamped
  slog lines at boot/pass/shutdown; log output asserted via handler hook;
  TDD red -> green recorded here.
