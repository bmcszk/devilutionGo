---
# devilGo-h2zj
title: Monster pack placement — sim wiring (PlaceMonster verb, dMonster occupancy, unique placement)
status: draft
type: task
created_at: 2026-11-28T14:32:05Z
updated_at: 2026-11-28T14:32:05Z
parent: devilGo-97mn
---

# Monster pack placement — sim wiring (PlaceMonster verb, dMonster occupancy, unique placement)

## Goal
The SIM-WIRING half of PlaceGroup parity, deferred from the goal2 triage (the data half lands in v791 as a pure function): a Monster entity, InitMonster, dMonster occupancy checks, the PlaceMonster kernel verb + gate + handler, unique-monster placement (PlaceUniqueMonst/PlaceUniqueMonsters - needs quest/level state).

## Research
Upstream anchors pinned e1ad0c2b6 (monster.cpp PlaceGroup/PlaceMonster/PlaceUniqueMonst, trigs/trigs.cpp quest placement). The data half (constants, placement algorithm with the vanilla yp += deltaX bug verbatim, IsMonsterAvailable gating) lands first in devilGo-v791 (goal2 wave) - this bean consumes it.

## Acceptance criteria
- [ ] PlaceMonster is a kernel command settled in the ordered pass; dMonster occupancy prevents overlap; packs spawn from the v791 pure function during dungeon gen; unique placement gated on quest state; all sim-pure (no rendering/wall-clock).

## Libraries
stdlib only.

## Skills for coder
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/go-tests/SKILL.md
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.agents/skills/ponytail/SKILL.md
- ~/.agents/skills/tdd/SKILL.md
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md

## Triage
Deferred from devilGo-v791 by the goal2 combined triage (task-g2-triage-brainer, 2026-11-27); blocked by a Monster-entity bean (parent resolves the id at claim time).
