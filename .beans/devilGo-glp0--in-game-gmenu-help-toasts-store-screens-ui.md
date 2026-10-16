---
# devilGo-glp0
title: In-game gmenu + help + toasts + store screens UI
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - ui
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-16T20:38:06Z
parent: devilGo-xih1
blocked_by:
    - devilGo-vf8q
    - devilGo-0qnd
---

# In-game chrome

## Goal
Pause menu (save/quit/options/tick slider), help overlay, diablo_msg toasts, store browse UI bound to stores core.

## Research
gmenu.cpp:185,251; gamemenu.cpp; help.cpp:143-191; diablo_msg.cpp:45-130; stores stext 20-line list flows (stores.cpp).

## Scope
internal/ui/{gmenu,help,toast,storeui}.go.

## Acceptance criteria
- [ ] game_menu_test port green
- [ ] Toast queue: init/add/draw/duration semantics match
- [ ] Store UI renders stock from stores core; buy/sell flows end-to-end with inventory

## Libraries
- stdlib only (no third-party deps at this stage)

## Skills for coder
/home/blaze/.agents/skills/ebiten/rendering/SKILL.md — list rendering

## Dependencies & blockers
devilGo-vf8q, devilGo-0qnd.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:stores-core dependency (devilGo-0qnd) not structurally wired though AC requires end-to-end store flows; stale dg-task refs.


## IMPL NOTES (triage 2026-10-16)
1 root: UI epic task — pause menu/help/toasts/store browse UI; parented xih1, blocked_by vf8q only.
2 local: gmenu.cpp:185 (gmenu_draw_pause) and :251 (gmenu_draw) verified; help.cpp:143 (DrawHelpSlider) / :191 (DrawHelp) verified; diablo_msg.cpp:45 (InitDiabloMsg) / :125-130 (3500ms default) verified; gamemenu tick slider 20..50 verified (gamemenu.cpp:177,271-273); AddSText/AddSLine stores.cpp:306-337 verified; test/game_menu_test.cpp exists.
3 online: n/a.
4 solution: AC bullet "Store UI renders stock from stores core; buy/sell flows end-to-end" requires the stores core (devilGo-0qnd, exists, draft) but it is NOT in blocked_by — only stale prose "devilGo-vf8q, devilGo-0qnd". Wire 0qnd or split store-UI AC out.
5 understanding: UI layer binding to gameplay-layer stores core — correct split; toast/gmenu/help are UI-only.
6 repro: `beans show devilGo-glp0`; `sed -n '125,130p' .../Source/diablo_msg.cpp`; `grep -n "AddSText(" .../Source/stores.cpp | head -2`; `beans show devilGo-0qnd`.
7 fix: `beans update devilGo-glp0 --blocked-by devilGo-0qnd` (keeps vf8q) + `--body-replace-old "devilGo-vf8q, devilGo-0qnd." --body-replace-new "devilGo-vf8q (widgets), devilGo-0qnd (stores core)."`
8 validation: `beans show devilGo-glp0` lists both blockers; ledger shows 0qnd blocking glp0.
9 related: vf8q, 0qnd, dcob, h6wn, xih1.
VERDICT: NEEDS-WORK:stores-core dependency (devilGo-0qnd) not structurally wired though AC requires end-to-end store flows; stale dg-task refs.


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).

ID MAP (fix2): all legacy dg-task-NNN references in this bean map to devilGo ids via docs/research/dg-id-map.json (canonical mapping in milestone devilGo-i56j). Operative deps = frontmatter (source of truth); prose refs are documentation.
