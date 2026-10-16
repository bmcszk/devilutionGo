---
# devilGo-xih1
title: 'UI flow: DiabloUI screens, menus, dialogs, toasts'
status: draft
type: epic
priority: normal
tags:
    - plan:epic
    - phase-2
created_at: 2026-10-15T16:37:33Z
updated_at: 2026-10-16T14:06:45Z
parent: devilGo-i56j
blocked_by:
    - devilGo-qbiy
---

# Game UI

## Goal
Title → main menu → hero select/create → (MP: connection/game select) → in-game gmenu/help/dialogs/toasts/progress.

## Research
DiabloUI screens diabloui.cpp:689 UiInitialize; mainmenu.cpp:34,69-70; hero/_uiheroinfo diabloui.h:60-93; multi/selconn+selgame; selok/selyesno/selstart; dialogs/progress/scrollbar/button/text_input; gmenu.cpp:185,251 + gmenu.h:101,106 sliders; gamemenu.cpp:274-277 tick slider; help.cpp:143,167,191; diablo_msg.cpp:45,125-130 (3500ms); interfac progress. Tests: game_menu_test, ui_test.hpp.

## Scope
internal/ui: screen stack, widgets (Button/MenuList/Slider/Scrollbar/TextInput/Dialog), all screens, fade transitions, focus/input routing kb+mouse(+gamepad later).

## Acceptance criteria
- [ ] Full flow title→new warrior→enter town works with mouse and keyboard-only
- [ ] Screen transitions fade like C++ (palette fade hooks)
- [ ] gmenu slider 20-50 tick-rate writes SimOptions live in single-player (restart in MP)
- [ ] Toast queue respects 3500ms default and stacking order
- [ ] game_menu_test port green; error dialogs (app_fatal equivalents) never crash

## Skills for coder
/home/blaze/.agents/skills/ebiten/input/SKILL.md — event routing
/home/blaze/.agents/skills/ebiten/rendering/SKILL.md — art-driven widgets
/home/blaze/.agents/skills/coding-rules/SKILL.md — widget primitives minimal

## Dependencies & blockers
blocked_by dg-epic-03. Blocks milestones M1/M2 (needs hero select + pause menu).

## Open questions
- ebitenui for settings screen only (H10): trial in this epic.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK


## IMPL NOTES (triage 2026-10-16)
1 root: Screen stack + widgets + all DiabloUI screens. Wired blockedBy qbiy ✓ (body "dg-epic-03"), blocks vinx ✓. 5 children.
2 local: Cites verified: DiabloUI/diabloui.cpp:689 UiInitialize ✓; diablo_msg.cpp:125-130 default duration 3500ms ✓; gamemenu.cpp tick-rate branches :162-168 (body 274-277 slider SetValue ✓ both present). Skills paths ✓ (ebiten/input, ebiten/rendering, coding-rules).
3 online: n/a.
4 solution: none — ready. ebitenui trial correctly parked as open question (H10).
5 understanding: game/ui layer over engine/render; gmenu 20-50 tick slider ↔ SimOptions live-write AC is testable.
6 repro: `grep -n "UiInitialize" Source/DiabloUI/diabloui.cpp`; `grep -n "3500" Source/diablo_msg.cpp`
7 fix: none.
8 validation: children (vf8q, emdz, dcob, glp0, h6wn) triaged in batch-08.
9 related: blocker qbiy; blocked vinx; sibling k736 (input routing into screens).
VERDICT: TRIAGED-OK
