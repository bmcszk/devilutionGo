---
# devilGo-c6wt
title: Virtual cursor + plrctrls action machine
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - input
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-16T14:01:03Z
parent: devilGo-k736
blocked_by:
    - devilGo-sf6k
---

# plrctrls

## Goal
Gamepad-equivalent of mouse actions: move/interact/attack/spell/menu via virtual cursor.

## Research
plrctrls.cpp full module; hooks every_frame:2305, after_check_curs_move:2264, after_game_logic:2312.

## Scope
internal/input/plrctrls.go + cursor pick integration point (render epic provides screen→tile).

## Acceptance criteria
- [ ] Action resolution identical to mouse path for a scripted 20-step session
- [ ] No action while PadMenu/Hotspell overlays active

## Skills for coder
/home/blaze/.agents/skills/ebiten/input/SKILL.md — buttons + axes combos

## Dependencies & blockers
dg-task-031.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:end-to-end AC (20-step session parity) not achievable at this bean's dependency level — gameplay/render deps not wired; AC needs unit-vs-e2e split.
