---
# devilGo-rfrj
title: Inventory grid + belt + gold
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - inv
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-15T16:45:26Z
parent: devilGo-nxxo
blocked_by:
    - devilGo-r462
---

# Inventory

## Goal
40-cell grid packing, 8-slot belt, gold stacks, cursor hold item.

## Research
player.h:38-39,218-227; inv.cpp placement rules; tests inv_test.cpp, inventory_ui_test.cpp.

## Scope
internal/gameplay/inventory.go: place/remove/auto-arrange, belt transfer, 2x3 item shapes, gold max.

## Acceptance criteria
- [ ] inv_test cases pass
- [ ] Belt hotkeys consume potions/scrolls correctly

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — small array logic

## Dependencies & blockers
dg-task-016.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.
