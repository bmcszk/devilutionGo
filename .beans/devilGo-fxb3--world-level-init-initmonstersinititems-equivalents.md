---
# devilGo-fxb3
title: 'World level init: InitMonsters/InitItems equivalents (D2 resolution)'
status: draft
type: task
priority: normal
created_at: 2026-11-19T22:32:40Z
updated_at: 2026-11-19T22:33:50Z
parent: devilGo-97mn
blocked_by:
    - devilGo-wq8m
    - devilGo-6851
    - devilGo-t8dw
---

# World level init: InitMonsters/InitItems equivalents (D2 resolution)

## Goal
When real monster/item registries land in the world slice, port the
LoadGameLevel init phases the world loadImage boundary leaves out today:
upstream `LoadGameLevelDungeon` (Source/diablo.cpp:3142 at pin
e1ad0c2b6) runs InitGolems :3147, InitObjects :3148, InitMonsters :3153,
InitItems :3154, InitMissiles :3160, InitCorpses :3161; the M1 world
loadImage carries pieces + SOL + floor items only and the harness script
places monsters/items. This is deviation D2.

## Research
- Anchors verified at pin e1ad0c2b6 (clean) — diablo.cpp:3142-3161.
- Dependency order: this port needs the create-pack and save/load verb
  slices first (entity identity) and D1's occupancy for placement rules.

## Acceptance criteria
- [ ] Monster/item/object registries exist in the world slice
- [ ] The Init* phase equivalents ported with upstream-parity pins
- [ ] D2 deviation retired in the milestone body's disposition section

## Triage
Filed 2026-11-19 as the D2 successor during the wave planning pass
(planner brainer pass + parent verification vs the live tree). Needs its
own triage research when world registries open.

## Libraries
stdlib only.

## Skills for coder
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/tdd/SKILL.md
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md
