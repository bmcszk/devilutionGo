---
# devilGo-4ybb
title: DRLG L2 Catacombs + L3 Caves + L4 Hell
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - drlg
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-15T16:45:26Z
parent: devilGo-97mn
blocked_by:
    - devilGo-76tu
---

# Remaining generators

## Goal
L2/L3/L4 parity.

## Research
drlg_l2.cpp:1780-1786 room sizes, :2830/:2839; drlg_l3.cpp:2185; drlg_l4.cpp:1228; fixtures levels/l{2,3,4}data + diablo/hellfire sets; tests drlg_l2..l4_test.cpp.

## Scope
internal/world/drlg_l{2,3,4}.go.

## Acceptance criteria
- [ ] All L2/L3/L4 fixtures byte-equal (diablo + hellfire)
- [ ] River/lava constraints (L3/L4) verified on 20 random seeds vs C++ dumps

## Skills for coder
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — fixture matrix

## Dependencies & blockers
dg-task-017 (shared engine).

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.
