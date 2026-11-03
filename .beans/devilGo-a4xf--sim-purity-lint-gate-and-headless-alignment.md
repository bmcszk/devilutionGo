---
# devilGo-a4xf
title: Sim-purity lint gate and headless alignment
status: draft
type: task
priority: normal
created_at: 2026-11-03T12:49:56Z
updated_at: 2026-11-03T12:50:42Z
parent: devilGo-xc37
---

# Sim-purity lint gate and headless alignment

## Goal

ADR-017 and the sim-purity discipline as a CI gate plus headless truth:
sim code (internal/game kernel, internal/diablo sim paths) must not import
rendering, ebiten, wall-clock time or file io; cmd/devilutionGo-headless
runs the kernel with a demo file end to end without a window.

## Research

- internal/game + internal/diablo import graph today (clean; keep it so).
- cmd/devilutionGo-headless exists with its own test run.
- scripts/lint.sh is the gate family home; coders run it pre-push.

## Scope

- scripts: new lint rule (import + time.Now/file-io ban list for sim paths).
- headless: kernel demo replay path proven in tests; no window, no GL.

## Acceptance criteria

- [ ] scripts/lint.sh fails when a sim package imports rendering or ebiten.
- [ ] headless binary replays a demo file with no display; go test ./... green.

## Libraries

- None new.

## Skills for coder

- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/ponytail/SKILL.md
- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/go-tests/SKILL.md
## Dependencies & blockers

- Kernel core at minimum; best after tick log so headless replays are real.

## Open questions

- None.

## Triage

TRIAGE brainer evidence 2026-11-03: kernel/migration laws come straight from the
accepted architecture records ADR-013..018 (deep research confrontation:
Game Programming Patterns event-queue chapter, Bevy observer-ordering doc,
Watermill GoChannel reference, AoE/Factorio/OpenTTD lockstep and desync
write-ups, devilutionX demo format parity). Code seams verified in the tree
surveys; workers get exact files in Research.
