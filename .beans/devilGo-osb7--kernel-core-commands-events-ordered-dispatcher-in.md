---
# devilGo-osb7
title: 'Kernel core: commands, events, ordered dispatcher in internal/game'
status: todo
type: task
priority: high
created_at: 2026-11-03T12:49:56Z
updated_at: 2026-11-03T14:23:09Z
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

- [ ] Given two handlers on one command, when the kernel runs a pass, events come out in registration order every time.
- [ ] Given the same seed and demo-record file, replay twice and the event trace hash matches.
- [ ] Whole go test ./... green, gofmt clean.

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
