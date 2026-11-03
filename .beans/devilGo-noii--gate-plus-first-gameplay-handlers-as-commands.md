---
# devilGo-noii
title: Gate plus first gameplay handlers as commands
status: draft
type: task
priority: normal
created_at: 2026-11-03T12:49:56Z
updated_at: 2026-11-03T12:50:42Z
parent: devilGo-xc37
---

# Gate at the edge plus first gameplay handlers

## Goal

ADR-018 laws 1, 3 and 6 over the first systems: input and UI clicks become
commands and stop mutating state directly. The gate drops impossible commands
(dead target, out of range, cooldown) before any dispatch. Gameplay systems
start registering handlers whose slices they own; wider change requested by
command, events announce what changed. First slice scoped narrow and honest:
click routing and one gameplay store path.

## Research

- internal/game/click.go: existing click router; keep its fixed cursor metrics.
- internal/game/DebugRouter: becomes another command producer.
- internal/gameplay: inventory/stores/itemgen state lives here; handlers
  subscribe for the slice they own.
- [ADR-018](../docs/adrs/adr-018.md) role table; [ADR-015](../docs/adrs/adr-015.md)
  ordered pass.

## Scope

- Command mapping for click + debug inputs; gate implementation; handler
  registration for one store path; event vocabulary for that path.
- Tests: given/when/then for each gate rule; no side effects after a
  rejected command; handler order proof.

## Acceptance criteria

- [ ] Given a dead target click, when it enters the gate, the command is dropped and no handler runs.
- [ ] Given valid commands, state changes only through handlers registered with the kernel.
- [ ] Whole go test ./... green, gofmt clean.

## Libraries

- None new.

## Skills for coder

- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/ponytail/SKILL.md
- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/go-tests/SKILL.md
## Dependencies & blockers

- Kernel core child settled first.

## Open questions

- Which store goes first; proposal: gold/inventory, smallest blast radius.

## Triage

TRIAGE brainer evidence 2026-11-03: kernel/migration laws come straight from the
accepted architecture records ADR-013..018 (deep research confrontation:
Game Programming Patterns event-queue chapter, Bevy observer-ordering doc,
Watermill GoChannel reference, AoE/Factorio/OpenTTD lockstep and desync
write-ups, devilutionX demo format parity). Code seams verified in the tree
surveys; workers get exact files in Research.
