---
# devilGo-7be9
title: Combat resolution pipeline
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - combat
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-16T14:01:03Z
parent: devilGo-nxxo
blocked_by:
    - devilGo-r462
---

# Hit & damage

## Goal
Melee/ranged/magic hit rolls + damage with exact roll order and 1/64 fixed point.

## Research
player.cpp:524-660 (see epic); player.h:575-666; RandomIntBetween random.hpp:403.

## Scope
internal/gameplay/combat.go: PlrHitMonst, PlrHitPlr(-PvP), MonstHitPlayer port, ApplyPlrDamage/DamageMonster, durability hooks.

## Acceptance criteria
- [ ] Roll-tape parity harness green for 3 classes × 5 monsters × 1000 seeds
- [ ] Damage in 1/64 units asserted (no float anywhere)
- [ ] Clamp 5..95 and petrified hit=0 rules tested

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — explicit int math
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — tape comparison

## Dependencies & blockers
dg-task-016.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:remap-deps (S1/S2 only; cites verified exact)
