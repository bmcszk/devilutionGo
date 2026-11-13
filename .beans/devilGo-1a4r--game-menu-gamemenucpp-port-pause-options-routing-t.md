---
# devilGo-1a4r
title: 'Game menu: gamemenu.cpp port (pause, options routing, tick-rate slider)'
status: todo
type: task
priority: normal
created_at: 2026-10-29T15:44:55Z
updated_at: 2026-11-13T00:33:34Z
parent: devilGo-xih1
---

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
- ~/.agents/skills/ponytail/SKILL.md

- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/ebiten/input/SKILL.md

## Dependencies & blockers

blocked_by devilGo-vf8q (widget kit + screen stack) — menus render through it.
Can start the menu-TREE data + routing layer before vf8q lands (pure logic + tests),
UI wiring after.

## Libraries

Binds ebiten v2 for rendering (owner flag in Scope); everything else stdlib.
## Triage

- Brainer beans-audit of the ledger ran 2026-10-28 (devilGo-5wsb-beansaudit session; PR #3 verdict log). Per-bean brainer pass owed before implementation dispatch; wave placement, blockers and skills list set during the 2026-10-29 reorg (d083780).

## Architecture guardrails (tech-debt triage 2026-11-12, binding for this bean's lane)

Source: the ok51 tech-debt triage (task-techdeptTriage-brainer) — these are the
lessons from the mistakes already made, applied BEFORE this bean starts. The
contract for this bean restates whichever apply; violating one is a review
REQUEST-CHANGES, not a debate.

1. Event sourcing (ADR-018): state changes this bean introduces go through
   kernel Commands with Gate rules and emitted Events — never direct mutations
   or a private callback bus. New Kind names live in the kernel package
   (law 5), defined per-domain WITH this wiring (no wholesale definition wave,
   no side-channel kinds). Presentation reads events only; it never writes sim
   state.
2. Layering (ADR-008): this bean's package imports only from its own layer and
   below (engine < game < diablo). Check the import direction BEFORE writing
   code; a generic mechanism never lands under internal/diablo, and diablo
   semantics (pcurs ordinals, PLR_MODE, panel flags) never leak into
   internal/game or internal/engine.
