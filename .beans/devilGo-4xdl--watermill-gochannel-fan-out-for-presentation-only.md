---
# devilGo-4xdl
title: Watermill GoChannel fan-out for presentation only
status: draft
type: task
priority: normal
created_at: 2026-11-03T12:49:56Z
updated_at: 2026-11-03T12:50:42Z
parent: devilGo-xc37
---

# Watermill GoChannel fan-out for presentation only

## Goal

ADR-018 laws 2 and 5 as code: after each ordered pass the kernel publishes
its events on the in-process bus for presentation consumers (renderer, sound,
UI flourishes). Sim-mutating handlers never subscribe to the bus; the kernel
calls them itself. This adds watermill to go.mod exactly as the catalog row
already records it: in-process dispatch, backend swappable, presentation only.

## Research

- docs/go-libraries.md catalog entry for the bus (GoChannel backend note).
- Watermill reference research from the architecture wave: default publish is
  non-blocking; cross-subscriber order is not guaranteed; therefore the kernel
  owns dispatch and the bus only fans out display traffic.
- Presentation consumers today: cmd/devilutionGo UI assembly, future sound
  catalog work.

## Scope

- go.mod addition per catalog; kernel publishes events post-pass; thin
  subscriber helper for presentation packages.
- Tests: zero-subscriber kernel unaffected; with N subscribers the ordered
  pass result is bit-identical; a content-change attempt from a subscriber
  is rejected by test guard.

## Acceptance criteria

- [ ] Given 0 or 3 subscribers, the kernel pass produces identical state and event trace.
- [ ] go.mod carries watermill matching the catalog row; go.sum clean.

## Libraries

- ThreeDotsLabs/watermill (catalog; the only new dependency in this feature)

## Skills for coder

- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/ponytail/SKILL.md
- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/go-tests/SKILL.md
## Dependencies & blockers

- Kernel core + tick log children first (bus consumer has nothing to read
  until events exist).

## Open questions

- None; sync fan-out inside the pass stays forbidden by law 3.

## Triage

TRIAGE brainer evidence 2026-11-03: kernel/migration laws come straight from the
accepted architecture records ADR-013..018 (deep research confrontation:
Game Programming Patterns event-queue chapter, Bevy observer-ordering doc,
Watermill GoChannel reference, AoE/Factorio/OpenTTD lockstep and desync
write-ups, devilutionX demo format parity). Code seams verified in the tree
surveys; workers get exact files in Research.
