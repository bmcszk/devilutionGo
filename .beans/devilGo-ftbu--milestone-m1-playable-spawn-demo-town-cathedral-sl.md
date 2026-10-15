---
# devilGo-ftbu
title: 'Milestone M1: Playable Spawn demo (town + cathedral slice)'
status: draft
type: epic
priority: normal
tags:
    - plan:epic
    - milestone
    - phase-1
created_at: 2026-10-15T16:37:33Z
updated_at: 2026-10-15T16:37:34Z
parent: devilGo-i56j
blocked_by:
    - devilGo-5wsb
    - devilGo-qbiy
---

# M1 Playable Spawn slice

## Goal
A human can: launch, pick warrior, walk Tristram, enter cathedral (L1), fight zombies/skeletons with melee, pick up items, return to town, save/load via SQLite, all with sound.

## Research
M1 matrix /tmp/dg-requirements.md Appendix F.2; spawn gating (quests.cpp:96-100, monster.cpp:3159); town + L1 tasks; stores subset.

## Scope
Integration: epics 02/03 + world L1/town + entity basics + combat melee + inventory minimal + UI hero flow + persistence minimal + input kb/mouse + audio.

## Acceptance criteria
- [ ] End-to-end scripted session (boot→create→town→dungeon→kill 10→loot→save→load) runs headless with zero errors and stable StateHash on replay
- [ ] Seed parity for L1 fixtures (from dg-task-017) included in gate
- [ ] 60fps rendering of town+dungeon on mid hardware (no frame >16ms p95)
- [ ] Manual smoke checklist documented (screenshots attached to bean)

## Skills for coder
/home/blaze/.agents/skills/ebiten/game-loop/SKILL.md, /home/blaze/.agents/skills/ebiten/rendering/SKILL.md, /home/blaze/.agents/skills/ebiten/input/SKILL.md, /home/blaze/.agents/skills/ebiten/audio/SKILL.md — integration surfaces
/home/blaze/.agents/skills/go-dev/SKILL.md — wiring discipline

## Dependencies & blockers
blocked_by dg-epic-02, dg-epic-03 (+ tasks 017/019/021 partial from epics 04/05 — coordinate; the milestone may pull specific tasks early rather than wait for full epics).

## Open questions
None.
