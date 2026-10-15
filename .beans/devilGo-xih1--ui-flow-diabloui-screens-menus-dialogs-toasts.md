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
updated_at: 2026-10-15T16:37:34Z
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
