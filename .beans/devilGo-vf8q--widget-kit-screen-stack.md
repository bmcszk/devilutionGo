---
# devilGo-vf8q
title: Widget kit + screen stack
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - ui
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-15T16:45:27Z
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

## Skills for coder
/home/blaze/.agents/skills/ebiten/input/SKILL.md — key/edge detection

## Dependencies & blockers
dg-task-014 (text).

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.
