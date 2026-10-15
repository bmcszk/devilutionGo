---
# devilGo-glp0
title: In-game gmenu + help + toasts + store screens UI
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - ui
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-15T16:45:26Z
parent: devilGo-xih1
blocked_by:
    - devilGo-vf8q
---

# In-game chrome

## Goal
Pause menu (save/quit/options/tick slider), help overlay, diablo_msg toasts, store browse UI bound to stores core.

## Research
gmenu.cpp:185,251; gamemenu.cpp; help.cpp:143-191; diablo_msg.cpp:45-130; stores stext 20-line list flows (stores.cpp).

## Scope
internal/ui/{gmenu,help,toast,storeui}.go.

## Acceptance criteria
- [ ] game_menu_test port green
- [ ] Toast queue: init/add/draw/duration semantics match
- [ ] Store UI renders stock from stores core; buy/sell flows end-to-end with inventory

## Skills for coder
/home/blaze/.agents/skills/ebiten/rendering/SKILL.md — list rendering

## Dependencies & blockers
dg-task-027, dg-task-025.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.
