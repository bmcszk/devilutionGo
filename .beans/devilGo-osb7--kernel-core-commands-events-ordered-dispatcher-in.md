---
# devilGo-osb7
title: 'Kernel core: commands, events, ordered dispatcher in internal/game'
status: completed
type: task
priority: high
created_at: 2026-11-03T12:49:56Z
updated_at: 2026-11-03T15:30:59Z
parent: devilGo-xc37
---

# Kernel core: commands, events, ordered dispatcher in internal/game

## Goal

The kernel itself, per [ADR-015](../docs/adrs/adr-015.md) +
[ADR-018](../docs/adrs/adr-018.md): typed commands, typed events, a
registered-handlers dispatcher that runs exactly one ordered pass per tick,
and events emitted only for actual state changes. `Game.Update` computes the
tick budget as today (gnTickDelay 50ms, 20 tps, startup burst) and hands each
tick to the kernel; `OnTick` stays the presentation callback carrier.
The demo records (internal/engine/demo) translate into commands through an
adapter beside the playback consumer, so recorded files feed the kernel.

## Research

- internal/game/game.go: tick constants + Game struct + Update flow.
- internal/engine/tick.go: TickSource + MockTickSource (deterministic clock seam).
- internal/engine/demo: format.go (record tags; demomode.cpp parity),
  playback.go consumer, format_test.go + playback_test.go round trips.
- internal/engine/random.go: seeded generator; bookkeeping for kernel RNG
  streams. cmd/devilutionGo/main.go wiring (Game + DebugRouterFromFlag).
- internal/game/debug.go/debug_flag.go: DebugRouterFromFlag is the one input
  surface today; it becomes a command producer in the next child.
- Layout law [ADR-008](../docs/adrs/adr-008.md): kernel code lives in
  internal/game; no new top-level dir.

## Scope

- internal/game: kernel types + dispatcher + integration into Game.Update.
- demo-to-command adapter beside internal/engine/demo playback.
- Tests: order proof (handlers fire in registration order), replay golden
  (same seed + same demo file twice, identical event trace hash), all in
  internal/game plus the demo adapter package.

## Acceptance criteria

- [x] Given two handlers on one command, when the kernel runs a pass, events come out in registration order every time.
- [x] Given the same seed and demo-record file, replay twice and the event trace hash matches.
- [x] Whole go test ./... green, gofmt clean.

## Libraries

- None new. Watermill comes later on purpose (child 4).

## Skills for coder

- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/ponytail/SKILL.md
- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/go-tests/SKILL.md
## Dependencies & blockers

- Parent kernel feature in todo.

## Open questions

- Event typing: interface + methods vs struct union; prefer struct union for
  parity-friendly sizes.

## Triage

TRIAGE brainer evidence 2026-11-03: brainer REQUEST-CHANGES on the architecture
PR produced findings; this text was rewritten against the recorded law and the
verified tree (real file seams, no invented files), lint-clean.

Claim evidence 2026-11-03: lane dispatched. Worker session task-kernel-core-coder,
branch dg-kernel-core (worktree /home/blaze/github/bmcszk/dg-lane-kernel),
off master 1cb9ffd. Deliverable: the kernel core per this bean; PR targets
master.

## Progress

- `internal/game/kernel` landed: `Command`/`Event` (integer-only payload),
  `Kernel` with `Register`/`Submit`/`Pass` (exactly one pass per tick, handler
  order = registration order), `Trace`/`TraceHash` (+ free `kernel.TraceHash(events)`
  for the per-tick hook), `Gate` seam with one rule (`DropUnnamed`), and `Streams`
  RNG bookkeeping over `engine.NewXoshiro128pp` with sorted stream names.
- `internal/game/game.go`: `Kernel` + `TickCount` fields; `GameLogic` calls
  `runKernelPass()` (one pass per settled tick). Tick budget untouched
  (gnTickDelay 50ms, 20 tps, startup burst) — proven by
  `TestGame_Update_KernelPassDoesNotChangeTickBudget`.
- `internal/game/demo.go`: `DemoReplay` implements `demo.Consumer`, maps
  MouseMotion/MouseButton/MouseWheel/Key/Quit to kernel commands, buffers them
  until the next Tick record, submits + runs one pass, `Replay` returns the
  trace hash.

## Proof of Work

Branch `dg-kernel-core`, commits `10dac47..df33bad` (4 commits, pushed).

Files added:
- `internal/game/kernel/command.go` — package doc (laws 1/2/3/6), `Kind`
  constants, `Command`, `Event`, `Gate`, `Handler`.
- `internal/game/kernel/streams.go` — `Streams` + `NewStreams` over
  `engine.NewXoshiro128pp`, `Names`, `Draw` (sorted names, no map-order walk).
- `internal/game/kernel/gate.go` — `GateFunc`, `DropUnnamed` (first gate rule).
- `internal/game/kernel/kernel.go` — `Kernel` `New`/`Register`/`Submit`/`Pending`/
  `Pass`/`Trace`/`TraceHash` + `kernel.TraceHash(events)` per-tick hook.
- `internal/game/kernel/kernel_test.go`, `internal/game/kernel_integration_test.go`.

Files changed:
- `internal/game/game.go` — `Kernel *kernel.Kernel` + `TickCount uint32`;
  `GameLogic` calls `runKernelPass()` (one pass per settled tick). Tick budget
  code untouched.
- `internal/game/demo.go` — `DemoReplay` (`demo.Consumer`) command mapping,
  per-tick buffering, `Replay(io.Reader) (uint64, error)`.

Acceptance evidence:
- `TestPass_TwoHandlersOnOneCommandEmitInRegistrationOrder` — PASS (order
  `first,second` on both passes; event kinds in the same order).
- `TestPass_RunsOncePerTick` — PASS (second `Pass(0)` returns nil, runs no handler).
- `TestSubmit_GateDropsCommandAndEmitsNothing` — PASS (gate rejects, 0 handler
  runs, 0 events, empty trace).
- `TestDemoReplay_SameSeedAndStreamHashIdentically` — PASS (same seed + same
  recorded stream → identical trace hash and event count twice).
- `TestDemoReplay_DifferentSeedChangesTheHash`, `TestGame_Update_HandsEachSettledTickToTheKernel`,
  `TestGame_Update_KernelPassDoesNotChangeTickBudget` — PASS.
- Gates: `gofmt -l .` empty; `go vet ./internal/...` rc0;
  `./scripts/lint.sh ./...` → `0 issues.`; `go test ./... -count=1` → 38 ok, 0 FAIL.

Deviations / notes:
- No wall-clock, goroutine, or map-iteration order in the kernel: chains are
  slices, streams are sorted. The machine lint gate for that class is bean
  `devilGo-a4xf` ("Sim-purity lint gate"), not built here — this repo has no
  such linter yet and adding one is outside this contract.
- Law 4 storage (tick-log writer) intentionally not implemented: `Trace()` +
  `TraceHash` are the seam it consumes.
- `TestPass_HandlerDrawingFromStreamIsDeterministic` covers RNG stream
  determinism; streams are only reachable through handlers, so `Streams.Draw`
  on an unknown name draws nothing.

Review wave 2026-11-03: brainer verdict APPROVE with three non-blocking
findings after claims 1-10 verified with file:line citations
(/home/blaze/github/dg-kernel-core.prreview2.log). Finding 1 (Pending
surface) and 3 (Command.Tick doc note) fixed on this branch in the coder
session; finding 2 (acceptance tick marks) closed here directly from the
verified evidence.
