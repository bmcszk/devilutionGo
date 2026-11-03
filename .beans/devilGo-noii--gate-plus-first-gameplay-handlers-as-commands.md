---
# devilGo-noii
title: Gate plus first gameplay handlers as commands
status: draft
type: task
priority: normal
created_at: 2026-11-03T12:49:56Z
updated_at: 2026-11-03T14:23:09Z
parent: devilGo-xc37
---

# Gate plus first gameplay handlers as commands

## Goal

ADR-018 laws 1, 3 and 6 for the first systems: input and UI paths become
command producers and stop mutating state directly. The gate drops impossible
commands (dead target, out of range, cooldown) before any dispatch.
Gameplay handlers register with the kernel for slices they own; wider change
goes out as commands, and events announce what actually changed. First slice
kept narrow and honest: the debug router plus one gameplay store path.

## Research

- Input surface today: internal/game/debug.go with DebugRouterFromFlag wired
  in cmd/devilutionGo/main.go; there is no click router in the game layer yet,
  so the command mapping for clicks is produced in this child against the
  demo record surface (internal/engine/demo).
- internal/gameplay: inventory/stores/itemgen state; handlers subscribe for
  the slice they own.
- [ADR-018](../docs/adrs/adr-018.md) role table; [ADR-015](../docs/adrs/adr-015.md)
  ordered pass.

## Scope

- Command mapping for debug inputs + first click commands; gate
  implementation; handler registration for one store path; event
  vocabulary for that path.
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

- Which gameplay store goes first; proposal: gold/inventory, smallest blast radius.

## Triage

TRIAGE brainer evidence 2026-11-03: brainer REQUEST-CHANGES on the architecture
PR produced findings; this text was rewritten against the recorded law and the
verified tree (real file seams, no invented files), lint-clean.
