---
# devilGo-xc37
title: Deterministic command/event simulation kernel (ADR-015, ADR-018)
status: todo
type: feature
priority: normal
created_at: 2026-11-03T12:49:56Z
updated_at: 2026-11-03T12:50:42Z
parent: devilGo-i56j
---

# Deterministic command/event simulation kernel

## Goal

A kernel inside `internal/game` that runs the game tick loop per the accepted
architecture decisions: commands in through one gate, one ordered pass per
tick, events out for every state change, and the tick log as the single
artifact behind replay, netcode and desync hunting. Law sources:
[ADR-015](../docs/adrs/adr-015.md), [ADR-018](../docs/adrs/adr-018.md);
layout law [ADR-008](../docs/adrs/adr-008.md) (kernel lives in the game
layer, no framework dir). Upstream tick semantics stay 1:1: gnTickDelay 50ms,
20 tps, progress ramp, startup burst (internal/game/game.go).

## Research

- [ADR-015](../docs/adrs/adr-015.md) / [ADR-018](../docs/adrs/adr-018.md) laws 1-6;
  [ADR-017](../docs/adrs/adr-017.md) headless; [ADR-016](../docs/adrs/adr-016.md) loopback seam.
- Existing seams in the tree: internal/game/game.go (`Game.Update`, `OnTick`,
  `DebugRouter`), internal/engine/tick.go (`TickSource`, `MockTickSource`),
  internal/engine/demo (format v3, playback consumer, writer: demomode.cpp parity),
  internal/game/autotest.go (autotest driver), internal/game/click.go (click routing),
  cmd/devilutionGo/main.go (demo Replay/Movie/Enumerate modes + PRNG seeding),
  internal/persist (SaveGameData/ItemPack/PlayerPack codecs, sqlite store,
  demo_compare, memmap_diff), internal/diablo/missiles (sink hook injection seam).
- Bus choice recorded in [docs/go-libraries.md](../docs/go-libraries.md):
  watermill carries presentation fan-out only, never sim-mutating handlers.

## Scope

Children, in build order:

1. Kernel core: command/event types + ordered dispatcher wired into ticks.
2. Gate + first gameplay handlers: input and UI flow as commands.
3. Tick log, snapshot every 1000 ticks, per-tick checksum verification.
4. Watermill GoChannel fan-out for presentation only.
5. Sim-purity lint gate + headless run alignment.

## Acceptance criteria

- [ ] Kernel settles every tick by one ordered pass, and the pass order is proven by tests.
- [ ] Same seed + same command file reaches the same final state, twice in a row and across runs.
- [ ] No output-specific i/o, no wall clock, no rendering packages imported by sim code.
- [ ] Whole go test ./... green and gofmt clean at every child merge.

## Libraries

- watermill (already cataloged; presentation fan-out only; no other new deps)
- No additional ECS for the kernel (donburi stays where it is today).

## Skills for coder

- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/ponytail/SKILL.md
- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/go-tests/SKILL.md

## Dependencies & blockers

- ADRs land through the architecture PR; nothing else blocks the core.

## Open questions

- Tick-log durability: append-only file + snapshot (demo-style) vs SQLite rows.
  Proposal: append-only file + snapshot, mirroring the demo format.

## Triage

TRIAGE brainer evidence 2026-11-03: deep research confrontation of
"everything is an event" (AoE 1500-archers lockstep, Factorio FFF 30/188,
OpenTTD desync docs, gafferongames lockstep series, Game Programming Patterns
event-queue + observer chapters, Bevy observer-ordering doc, Watermill
GoChannel docs + DeepWiki, Unreal GAS). Verdict went into the accepted
architecture PR as the kernel record; this feature is its execution.
