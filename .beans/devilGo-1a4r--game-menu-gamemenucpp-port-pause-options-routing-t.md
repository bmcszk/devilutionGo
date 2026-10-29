---
# devilGo-1a4r
title: 'Game menu: gamemenu.cpp port (pause, options routing, tick-rate slider)'
status: todo
type: task
priority: normal
created_at: 2026-10-29T15:44:55Z
updated_at: 2026-10-29T16:14:42Z
parent: devilGo-xih1
---

## Goal

Port `gamemenu.cpp` (391 lines): in-game pause menu — Esc menu, options routing
(tick-rate slider 20-50 → gnTickDelay, per diablo.cpp:146), new-game/quit flows,
`gamemenu_previous`/`gamemenu_enter` state machine.

## Research

Source/gamemenu.cpp:1-391 (menu tables GameMenuNewGame/GameMenuSave/GameMenuOptions...);
Source/gamemenu.cpp:277 tick-rate; interfac.cpp screen transitions; DiabloUI/SettingsMenu
for the options submenus; menus.cpp routing.

## Scope

internal/ui/gamemenu.go (screen-stack entry per xih1 widget kit) + routing hooks in
internal/game. Options persistence lands later via yjuv SimOptions — this bean only wires
the menu tree + actions that exist (save/load once oqho/xwtt land: gate those two entries
on their presence, stub with "not available" like spawn did).

## Acceptance criteria

- [ ] Menu tree structure matches GameMenuTables (entry-for-entry, incl. spawn/hellfire gating)
- [ ] Esc → pause → resume round-trip with correct screen stack behavior
- [ ] Tick-rate slider adjusts gnTickDelay equivalent 20-50 (engine tick.go honors it live)
- [ ] Menu actions call the right engine hooks (new game starts run, quit exits cleanly)
- [ ] Screen transitions match C++ (no flicker, same redraw regions)

## Skills for coder

- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/ebiten/input/SKILL.md

## Dependencies & blockers

blocked_by devilGo-vf8q (widget kit + screen stack) — menus render through it.
Can start the menu-TREE data + routing layer before vf8q lands (pure logic + tests),
UI wiring after.
