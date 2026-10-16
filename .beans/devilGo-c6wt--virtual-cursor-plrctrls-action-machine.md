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
updated_at: 2026-10-16T14:06:44Z
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


## IMPL NOTES (triage 2026-10-16)
1 root: Input epic task — gamepad action machine; blocked_by sf6k (matches tree dg-task-032←031).
2 local: hooks verified: plrctrls_every_frame :2305, plrctrls_after_check_curs_move :2264, plrctrls_after_game_logic :2312 (Movement(*MyPlayer)); file 2614 lines; PadMenuNavigatorActive/PadHotspellMenuActive guard verified :1861 — AC "no action while overlays active" directly testable.
3 online: n/a.
4 solution: AC mixes levels: "action resolution identical to mouse path for a scripted 20-step session" requires gameplay entities (combat/interact/towners) + render screen→tile, none of which are structural deps (only sf6w blocked; gameplay epic nxxo not wired). Split AC: (a) headless action-machine unit tests now; (b) 20-step session parity gated on nxxo + render cursor seam. Prose dep "dg-task-031" stale → sf6k (structured already correct).
5 understanding: input layer integrating with gameplay hooks — epic k736 body already says "integrates with gameplay in epic-05 hooks"; the bean should express that gate explicitly.
6 repro: `beans show devilGo-c6wt`; `sed -n '2305,2314p' .../Source/controls/plrctrls.cpp`; `beans list --json | grep -c devilGo-nxxo`.
7 fix: `beans update devilGo-c6wt --body-replace-old "dg-task-031." --body-replace-new "devilGo-sf6k (structured). End-to-end 20-step session AC additionally gated on gameplay epic devilGo-nxxo + render cursor seam."` + split AC bullets.
8 validation: `beans show devilGo-c6wt` shows split AC + nxxo gate.
9 related: sf6k, t4fz, nxxo children, k736.
VERDICT: NEEDS-WORK:end-to-end AC (20-step session parity) not achievable at this bean's dependency level — gameplay/render deps not wired; AC needs unit-vs-e2e split.
