---
# devilGo-xw8k
title: Quests FSM + XP + spawn/hellfire gating
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - quests
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-16T14:01:06Z
parent: devilGo-nxxo
blocked_by:
    - devilGo-r462
---

# Quests & progression

## Goal
Quest lifecycle, kill hooks, multiplayer sync vars, XP table, gating flags.

## Research
questdat.hpp:102-129; quests.cpp:60-110,227,397,610; Experience.tsv; gating monster.cpp:3159, player.cpp:3479,3688-3701.

## Scope
internal/gameplay/quests.go, xp.go; SimFlags{Spawn,Hellfire,Multiplayer} helper API.

## Acceptance criteria
- [ ] quests_test port passes
- [ ] Level-15 seed quest pools equal C++ on 10 seeds
- [ ] XP thresholds + stat/point award flow

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — small FSM clarity

## Dependencies & blockers
dg-task-016, dg-task-008.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:missing-deps-devilGo-lsz6+devilGo-pu3r (+remap)
