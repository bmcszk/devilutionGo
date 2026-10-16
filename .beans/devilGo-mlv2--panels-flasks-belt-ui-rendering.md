---
# devilGo-mlv2
title: Panels, flasks, belt UI rendering
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - ui
created_at: 2026-10-15T16:35:16Z
updated_at: 2026-10-16T14:01:04Z
parent: devilGo-qbiy
blocked_by:
    - devilGo-n89r
---

# Panel chrome

## Goal
Control panel, life/mana flasks, belt slots, buttons from ui_art with exact rects.

## Research
Source/control/control_panel.cpp:351 DrawPanelBox, :435-445 button rects; control_flasks.cpp; panels/mainpanel.cpp; tests panel_state_test.cpp, inventory_ui_test.cpp.

## Scope
internal/render/panels: mainpanel, flasks (fill fraction), belt (8 slots), buttons (press state).

## Acceptance criteria
- [ ] Flask fill matches player HP/MP fraction per C++ formula
- [ ] Panel button hover/press visuals state-machine ported (test)

## Skills for coder
/home/blaze/.agents/skills/ebiten/rendering/SKILL.md — static art blits

## Dependencies & blockers
dg-task-007.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.


## Ebiten graphics-migration gates (from /tmp/dg-graphics.md G10/G11 — binding)
- [ ] UI layer rendered from UI atlas page; draw order after world buffer; per-frame cost <100 draw calls.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:minor (wrong mainpanel.cpp path; stale body dep; flask AC wording vs integer fillPer) — otherwise ready.
