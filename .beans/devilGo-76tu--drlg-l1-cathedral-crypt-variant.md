---
# devilGo-76tu
title: DRLG L1 Cathedral + Crypt variant
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - drlg
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-16T14:01:03Z
parent: devilGo-97mn
blocked_by:
    - devilGo-vti7
---

# L1 generator

## Goal
CreateL5Dungeon parity incl. crypt branch.

## Research
drlg_l1.cpp:1298-1312 entry; internals :1200-1294 (Fix*, Substitution, lamps, FillFloor, SelectChamber); crypt :1214,1221; Pass3/DRLG_LPass3 gendung.h:88; fixtures test/fixtures/diablo + levels/l1data.

## Scope
internal/world/drlg_l1.go + crypt.go sharing engine.

## Acceptance criteria
- [ ] All diablo L1 fixtures byte-equal
- [ ] Hellfire crypt fixtures byte-equal
- [ ] Quest chamber selection matches SelectChamber logic for all availability combos

## Skills for coder
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — fixture matrix
/home/blaze/.agents/skills/go-dev/SKILL.md — RNG-call-order discipline

## Dependencies & blockers
dg-task-015, dg-task-002.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK
