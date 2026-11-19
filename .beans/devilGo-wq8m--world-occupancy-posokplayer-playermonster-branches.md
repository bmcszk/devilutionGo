---
# devilGo-wq8m
title: 'World occupancy: PosOkPlayer player/monster branches (D1 resolution)'
status: draft
type: task
priority: normal
created_at: 2026-11-19T22:32:40Z
updated_at: 2026-11-19T22:33:50Z
parent: devilGo-97mn
---

# World occupancy: PosOkPlayer player/monster branches (D1 resolution)

## Goal
When the world slice gains player/monster occupancy registries, port the
posOkPlayer branches that stop at tiles-only today: upstream
`PosOkPlayer(const Player &player, Point position)` (Source/player.cpp:3089
at pin e1ad0c2b6) checks occupancy, not just walkability; the M1 wave
records this as deviation D1 (one-player scripted walk cannot hit the
gap).

## Research
- Anchor verified at pin e1ad0c2b6 (clean): player.cpp:3089.
- Go side: internal/game/world posOkPlayer stops at tile rules
  (walkable-only, no registries).

## Acceptance criteria
- [ ] Occupancy registries exist (player, monster) in the world slice
- [ ] PosOkPlayer branches ported and pinned against the upstream
      decision table
- [ ] D1 deviation retired in the milestone body's disposition section

## Triage
Filed 2026-11-19 as the D1 successor during the wave planning pass
(planner brainer pass + parent verification vs the live tree). Needs its
own triage research when the world slice opens entity registries.

## Libraries
stdlib only.

## Skills for coder
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/tdd/SKILL.md
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md
