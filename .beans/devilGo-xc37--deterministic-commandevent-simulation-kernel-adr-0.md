---
# devilGo-xc37
title: Deterministic command/event simulation kernel (ADR-015, ADR-018)
status: todo
type: feature
priority: normal
created_at: 2026-11-03T12:49:56Z
updated_at: 2026-11-03T14:23:09Z
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
layer). Upstream tick semantics stay 1:1: gnTickDelay 50ms, 20 tps,
progress ramp, startup burst (internal/game/game.go).

## Research

- [ADR-015](../docs/adrs/adr-015.md) / [ADR-018](../docs/adrs/adr-018.md) laws
  1-6; [ADR-017](../docs/adrs/adr-017.md) headless; [ADR-016](../docs/adrs/adr-016.md)
  loopback seam.
- Verified seams in this tree: internal/game/game.go (Game.Update, OnTick,
  DebugRouter), internal/game/debug.go + debug_flag.go (DebugRouterFromFlag),
  internal/engine/tick.go (TickSource, MockTickSource), internal/engine/random.go
  (seeded generator + roundtrip), internal/engine/demo (format.go: demomode.cpp
  parity records; playback.go consumer; format_test.go + playback_test.go prove
  the decode round trip), internal/persist (SaveGameData/ItemPack/PlayerPack
  codecs, sqlite store, demo_compare, memmap_diff), cmd/devilutionGo/main.go
  (Game + DebugRouterFromFlag wiring; GUI entry is a stub today),
  cmd/devilutionGo-headless/main.go (headless tick loop over MockTickSource
  with appfat exit codes).
- Bus choice recorded in [docs/go-libraries.md](../docs/go-libraries.md):
  watermill carries presentation fan-out only, never sim-mutating handlers.

## Scope

Children, in build order:

1. Kernel core: command/event types + ordered dispatcher wired into ticks.
2. Gate + first gameplay handlers: input and UI flow as commands.
3. Tick log, snapshot every 1000 ticks, per-tick checksum verification
   (container: SQLite store per ADR-018 law 4).
4. Watermill GoChannel fan-out for presentation only.
5. Sim-purity lint gate + headless run alignment.

## Acceptance criteria

- [ ] Kernel settles every tick by one ordered pass, and the pass order is proven by tests.
- [ ] Same seed + same command file reaches the same final state, twice in a row and across runs.
- [ ] No rendering packages or wall clock imported by sim code.
- [ ] Whole go test ./... green and gofmt clean at every child merge.

## Libraries

- watermill (cataloged; presentation fan-out only; no other new deps).
- No additional ECS for the kernel.

## Skills for coder

- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/ponytail/SKILL.md
- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/go-tests/SKILL.md
## Dependencies & blockers

- The architecture records land with this PR; no other blocker.

## Open questions

- None; the container law settled the tick-log storage question.

## Triage

TRIAGE brainer evidence 2026-11-03: brainer REQUEST-CHANGES on the architecture
PR produced findings; this text was rewritten against the recorded law and the
verified tree (real file seams, no invented files), lint-clean.
