---
# devilGo-foug
title: Themes + monster/item population + towners
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - world
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-16T14:01:04Z
parent: devilGo-97mn
blocked_by:
    - devilGo-4ybb
---

# Populate the world

## Goal
Theme rooms; monster type/spawn lists; ground items; towners with animOrder.

## Research
themes.cpp:349-449; monster.cpp:566-603,195,282-292; items.cpp:1475,3195,3422; towners.cpp:760,811; townerdat TSV animOrder column (100+ entries per NPC).

## Scope
internal/world/themes.go + internal/entity population fns.

## Acceptance criteria
- [ ] Theme placement equal on 10 seeds/level-type vs C++ dump
- [ ] Spawn lists (type indices, positions, unique placement flags) equal on 3 seeds
- [ ] Towner animation sequences tick exactly per TSV order

## Skills for coder
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — dump-comparison tests

## Dependencies & blockers
dg-task-016, dg-task-018.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:wire-deps (r462 missing) + 1 wrong cite — ACs and remaining research verified.
