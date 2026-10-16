---
# devilGo-emdz
title: Title, main menu, credits, dialogs, progress
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - ui
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-16T14:01:04Z
parent: devilGo-xih1
blocked_by:
    - devilGo-vf8q
---

# Front screens

## Goal
The pre-game flow complete.

## Research
title.cpp, mainmenu.cpp, credits.cpp+lines, support_lines, selok/selyesno/selstart, dialogs.cpp, progress.cpp.

## Scope
internal/ui screens + fade transitions + attract-mode timeout (mainmenu_attract_time_out).

## Acceptance criteria
- [ ] Flow navigable; attract timeout returns to title
- [ ] Progress bars on level load + saves

## Skills for coder
/home/blaze/.agents/skills/ebiten/rendering/SKILL.md — static art + text

## Dependencies & blockers
dg-task-027.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:AC lacks concrete parity gate (labels must come from h6wn spec; attract timeout constant unspecified) — "Flow navigable" is not a TDD-first criterion; stale dg-task-027 ref.
