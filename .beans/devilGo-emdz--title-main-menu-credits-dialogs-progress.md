---
# devilGo-emdz
title: Title, main menu, credits, dialogs, progress
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - ui
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-16T14:26:52Z
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


## IMPL NOTES (triage 2026-10-16)
1 root: UI epic task — pre-game flow; parented xih1, blocked_by vf8q (matches tree dg-task-028←027).
2 local: title.cpp, credits.cpp, selok.cpp, selyesno.cpp, selstart.cpp, dialogs.cpp, progress.cpp all exist; attract timeout verified (DiabloUI/mainmenu.cpp:26 `mainmenu_attract_time_out`, :105 dwAttractTicks); selok.cpp:86 dialog construction verified.
3 online: n/a.
4 solution: AC too vague for TDD-first gate: "Flow navigable" would pass with invented labels/screens — no criterion binding labels/trees to the h6wn content spec, no concrete attract-timeout constant or fade-parity check. Add: menu tree == h6wn spec (blocks-on/uses h6wn), timeout returns-to-title at C++ default seconds, progress-bar draw parity hook. Prose dep "dg-task-027" stale → vf8q (already structured).
5 understanding: UI layer consuming h6wn content spec + vf8q widgets — correct layering, just not expressed in AC.
6 repro: `beans show devilGo-emdz`; `grep -n "attract" .../Source/DiabloUI/mainmenu.cpp`; `beans show devilGo-h6wn`.
7 fix: `beans update devilGo-emdz --body-replace-old "dg-task-027." --body-replace-new "devilGo-vf8q (structured). Consumes devilGo-h6wn menu content spec; AC: main-menu tree == h6wn spec, labels byte-identical."` + tighten AC bullets.
8 validation: `beans show devilGo-emdz` shows h6wn linkage + concrete AC.
9 related: vf8q, h6wn, dcob, xih1.
VERDICT: NEEDS-WORK:AC lacks concrete parity gate (labels must come from h6wn spec; attract timeout constant unspecified) — "Flow navigable" is not a TDD-first criterion; stale dg-task-027 ref.


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).
