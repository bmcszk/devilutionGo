---
# devilGo-dcob
title: Hero select/create/delete + class stats
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
    - devilGo-emdz
---

# Hero screens

## Goal
selhero over persistence (list/create/delete), class stat preview.

## Research
DiabloUI/hero/selhero.cpp; _uiheroinfo diabloui.h:60-72; callbacks :81,93; pfile_ui_* API pfile.h:120-125; MAX_CHARACTERS 99 pfile.h:24.

## Scope
internal/ui/selhero.go wired to internal/persist hero store interface (stub until epic-08).

## Acceptance criteria
- [ ] Create warrior/rogue/sorcerer(+monk HF) with defaults from playerdat
- [ ] Delete confirm dialog; 99-slot cap enforced

## Skills for coder
/home/blaze/.agents/skills/ebiten/input/SKILL.md — text input for name (16 chars, _uiheroinfo.name)

## Dependencies & blockers
dg-task-028.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:AC omits Bard/Barbarian class entries (selhero.cpp:437-463) + no class-stats row-exact parity criterion; stale dg-task-028 prose ref.
