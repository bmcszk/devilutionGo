---
# devilGo-onlm
title: 'Item generation: base/affix/unique + durability'
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - items
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-15T16:45:26Z
parent: devilGo-nxxo
blocked_by:
    - devilGo-r462
---

# Itemgen

## Goal
SpawnItem/SpawnUnique/affix rolls parity.

## Research
items.cpp:1048,1138,1218-1224,1475,3195,3422; itemdat.h:628-655; T_MASK treasure decode; tests items_test.cpp.

## Scope
internal/gameplay/itemgen.go + affix application + price calc + durability/break flow (player.cpp:505-522).

## Acceptance criteria
- [ ] 10000-seed histogram parity (epic criterion) for 3 classes
- [ ] Unique availability + min-level gates tested
- [ ] Durability break removes equipment + CalcPlrInv call verified

## Skills for coder
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — distribution tests

## Dependencies & blockers
dg-task-016, dg-task-008.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.
