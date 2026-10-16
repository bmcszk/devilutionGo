---
# devilGo-vf8q
title: Widget kit + screen stack
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - ui
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-16T20:38:07Z
parent: devilGo-xih1
blocked_by:
    - devilGo-v1hy
---

# Widgets

## Goal
Six primitives + stack + focus model.

## Research
DiabloUI/button.cpp, scrollbar.cpp, text_input.hpp, ui_item.h; gmenu item/sliders.

## Scope
internal/ui/widget.go, screen.go; art loading via assets; input routing.

## Acceptance criteria
- [ ] Keyboard nav (arrows+enter+esc) on every primitive
- [ ] Slider emits int steps like gmenu_slider_get semantics


## Libraries
- github.com/ebitenui/ebitenui — widget kit (optional; eval vs hand-rolled at claim, record decision)

## Skills for coder
/home/blaze/.agents/skills/ebiten/input/SKILL.md — key/edge detection

### game-dev skills (2026-10-16 audit)
- /home/blaze/.agents/skills/game-developer/SKILL.md — UI state machines
- /home/blaze/.agents/skills/game-development/game-design/SKILL.md — menu/UX flow

## Dependencies & blockers
dg-task-014 (text).

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK


## IMPL NOTES (triage 2026-10-16)
1 root: UI epic foundation task — six widget primitives + screen stack + focus; blocked_by v1hy (bitmap text renderer, exists in qbiy) matching tree dg-task-027←dg-task-014.
2 local: DiabloUI/button.cpp, scrollbar.cpp, text_input.hpp, ui_item.h (UiItemBase:30) all exist; gmenu_slider_get gmenu.h:101 + gmenu_slider_steps :106 verified.
3 online: n/a.
4 solution: none — AC unit-testable (keyboard nav per primitive; slider int-step semantics vs gmenu_slider_get). Nit: prose "dg-task-014 (text)" stale → v1hy (structured already correct); Goal says "six primitives" — enumerate them in AC to make "every primitive" checkable.
5 understanding: UI layer core; correct dependency (text renderer before widget kit).
6 repro: `beans show devilGo-vf8q`; `grep -n "gmenu_slider_get" .../Source/gmenu.h`; `ls .../Source/DiabloUI/{button,scrollbar}.cpp`.
7 fix: none required (optional prose id remap).
8 validation: n/a.
9 related: v1hy, emdz, dcob, glp0, xih1.
VERDICT: TRIAGED-OK


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.
