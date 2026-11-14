---
# devilGo-s7x0
title: 'Warrior standing in town: real plrgfx idle frame drawn over the world'
status: in-progress
type: task
priority: critical
created_at: 2026-11-14T12:33:20Z
updated_at: 2026-11-14T12:36:42Z
parent: devilGo-hw9y
---

## Goal

The warrior, standing, in the town scene: load the real plrgfx warrior
STAND (idle) frames from the archives, pick frame 0, and draw it over the
world at the hero's tile through the Presenter pipeline — the "real
character standing in the village" the user asked for.

## Research (parent, verified 2026-11-14)

- plrgfx layout at the pin: plrgfx/common.cel + per-class files; the
  warrior STAND set for town armor is one of the plrgfx archives
  (beapst/CL2-family — verify exact record names against the pin's
  plr.cpp/plr gfx tables and the archive listfile; cite in PoW).
- Decode path exists: assets DecodeCL2/DecodeCEL produce sprite lists;
  Presenter draws Surface content through the palette+shader pipeline —
  the hero frame must go through the same palette mapping as the world
  (no independent RGB side-channel).
- Direction/animation discipline: STAND has 8 directions x N frames; this
  bean pins direction = the caller's facing (default S) and frame = 0
  only — animation cycling is a later bean.
- The hero's tile comes from the smoke lane's world state (world.State
  Hero()) — draw at WorldToScreen(heroTile) with the standard sprite
  anchor (feet at tile center-bottom, upstream's offset rule).

## Acceptance criteria

- [ ] Warrior STAND frame 0 (one direction) decodes from the real archives (skip-if-absent guard)
- [ ] Drawn into the composed town scene at the hero tile with upstream's sprite anchor; headless test asserts hero pixels present over the world at the expected screen rect
- [ ] Palette-consistent: hero pixels go through the same Presenter palette mapping (no raw RGB bypass)
- [ ] PNG artifact shows the figure visibly standing in town (parent visual check)
- [ ] TDD red -> green; gates green

## Libraries

- existing internal deps (assets/render) + stdlib; no new third-party

## Skills for coder

- `~/.agents/skills/coding-rules/SKILL.md` — read fully before editing
- /home/blaze/.agents/skills/ebiten/rendering/SKILL.md
- `~/.agents/skills/go-tests/SKILL.md`

- `~/.agents/skills/ponytail/SKILL.md`

## Triage

Parent mini-triage 2026-11-14 (session _chat, user priority directive): seams verified in-tree at file:line (see Research); scope split = town render inputs (6asa), window boot (v9xy), hero sprite (s7x0), live capture (1awt); grouped PR, brainer re-review rides the PR review round per lane law. Skills paths stat-verified.
