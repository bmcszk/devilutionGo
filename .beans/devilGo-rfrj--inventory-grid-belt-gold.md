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
updated_at: 2026-10-16T14:06:45Z
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

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:remap-deps (S1/S2 only)


## IMPL NOTES (triage 2026-10-16)
1 root: 40-cell grid packing, 8-slot belt, gold stacks, cursor hold; = dg-task-024.
2 local: beans show ✓; player.h:38-39 EXACT lines = `InventoryGridCells = 40` / `MaxBeltItems = 8` ✓; inv.cpp exists ✓; test/inv_test.cpp + test/inventory_ui_test.cpp exist ✓ (both found by find).
3 online: n/a
4 solution: remap deps (dg-task-016→devilGo-r462); S2 pkg root (AGENTS.md names internal/game/inventory explicitly); AC could enumerate which inv_test cases, but "inv_test cases pass" is testable as-is.
5 understanding: game-layer inventory logic consumed by itemgen/stores/persistence; UI part goes to panels epic (devilGo-mlv2).
6 repro: beans show devilGo-rfrj; sed -n '38,39p' devilutionX/Source/player.h; ls devilutionX/test/inv_test.cpp devilutionX/test/inventory_ui_test.cpp
7 fix: beans update devilGo-rfrj --blocked-by devilGo-r462; body-append id map.
8 validation: beans show lists blocked_by devilGo-r462.
9 related: devilGo-r462, devilGo-onlm, devilGo-0qnd, devilGo-mlv2
VERDICT: NEEDS-WORK:remap-deps (S1/S2 only)
