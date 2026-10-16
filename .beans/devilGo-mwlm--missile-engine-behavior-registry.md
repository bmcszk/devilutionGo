---
# devilGo-mwlm
title: Missile engine + behavior registry
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - missiles
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-16T14:01:04Z
parent: devilGo-nxxo
blocked_by:
    - devilGo-7be9
---

# Missiles

## Goal
All missile behaviors table-driven; motion/collision exact.

## Research
misdat.h:166-196; missiles.cpp:479,636,728-1242; missile_sprites.tsv; tests missiles_test.cpp.

## Scope
Registry map[MissileID]{Add,Process}; fixed-point motion; CheckMissileCol; trap/golem/unique spawns via AddMissile chains.

## Acceptance criteria
- [ ] misdat.tsv row count == registry size (no missing)
- [ ] missiles_test port passes
- [ ] 16.16 traveled arithmetic property-tested vs C++ probe

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — registry pattern, no reflection

## Dependencies & blockers
dg-task-021.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:AC1-row-count==registry-size-is-false (verified against disk; would fail a correct port)
