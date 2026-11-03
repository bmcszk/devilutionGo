---
# devilGo-osb7
title: 'Kernel core: commands, events, ordered dispatcher in internal/game'
status: todo
type: task
priority: high
created_at: 2026-11-03T12:49:56Z
updated_at: 2026-11-03T12:50:42Z
parent: devilGo-xc37
---

# Kernel core: command/event types and ordered dispatcher inside internal/game

## Goal

The kernel itself, per [ADR-015](../docs/adrs/adr-015.md) +
[ADR-018](../docs/adrs/adr-018.md): typed commands, typed events, a
registered-handlers dispatcher that runs exactly one ordered pass per tick,
and events emitted only for actual state changes. `Game.Update` computes the
tick budget as today (gnTickDelay 50ms, 20 tps, startup burst) and hands each
tick to the kernel; the `OnTick` callback becomes the presentation carrier.
The demo adapter (internal/engine/demo playback consumer) translates recorded
input records into commands, so existing Replay/Movie modes feed the kernel.

## Research

- internal/game/game.go: tick constants + Game struct + Update flow.
- internal/engine/tick.go: TickSource + MockTickSource (deterministic clock seam).
- internal/engine/demo/format.go, playback.go, writer.go: record tags and consumer.
- internal/game/autotest.go: existing autotest driver, natural harness for
  replay goldens. cmd/devilutionGo/main.go: demo modes + PRNG reseeding.
- Layout law [ADR-008](../docs/adrs/adr-008.md): kernel code lives in
  internal/game; no new top-level dir.

## Scope

- internal/game: kernel types + dispatcher + integration into Game.Update.
- demo-to-command adapter beside the existing playback consumer.
- Tests: order proof (two handlers pure in registration order), replay golden
  (same seed + same demo file twice, identical event trace hash).

## Acceptance criteria

- [ ] Given two handlers on one command, when the kernel runs a pass, the events come out in registration order every time.
- [ ] Given the same seed and demo file, replay twice and the event trace hash matches.
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

- Event typing: interface + methods vs struct union; prefer struct union to
  keep binary parity-friendly.

## Triage

TRIAGE brainer evidence 2026-11-03: kernel/migration laws come straight from the
accepted architecture records ADR-013..018 (deep research confrontation:
Game Programming Patterns event-queue chapter, Bevy observer-ordering doc,
Watermill GoChannel reference, AoE/Factorio/OpenTTD lockstep and desync
write-ups, devilutionX demo format parity). Code seams verified in the tree
surveys; workers get exact files in Research.
