---
# devilGo-foug
title: Themes + monster/item population + towners
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - world
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-16T20:37:08Z
parent: devilGo-97mn
blocked_by:
    - devilGo-4ybb
    - devilGo-1e1b
    - devilGo-778h
    - devilGo-fpkz
    - devilGo-r462
    - devilGo-v791
    - devilGo-1spo
---

# Populate the world

## Goal
Theme rooms; monster type/spawn lists; ground items; towners with animOrder.

## Research
themes.cpp:349-449; monster.cpp:566-603,195,282-292; items.cpp:1475,3195,3422; towners.cpp:760,811; townerdat TSV animOrder column (100+ entries per NPC).

## Scope
internal/world/themes.go + internal/entity population fns.

## Acceptance criteria
- [ ] Theme placement equal on 10 seeds/level-type vs C++ dump
- [ ] Spawn lists (type indices, positions, unique placement flags) equal on 3 seeds
- [ ] Towner animation sequences tick exactly per TSV order

## Skills for coder
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — dump-comparison tests

## Dependencies & blockers (fix2: all edges verified wired — r462, v791, 778h, 1e1b)
dg-task-016, dg-task-018.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:wire-deps (r462 missing) + 1 wrong cite — ACs and remaining research verified.


## IMPL NOTES (triage 2026-10-16)
1 root: Phase-2 task (dg-task-020): theme rooms, spawn lists, ground items, towner animOrder; draft; wire blockedBy [4ybb] only; body declares dg-task-016 (r462) + dg-task-018 (4ybb) — r462 NOT wired and NOT transitive (4ybb chain = 76tu/vti7/m2x7/lfvi, no r462).
2 local: Cites verified: monster.cpp:195 InitMonster ✓, :282-292 CanPlaceMonster ✓, :566-603 AddMonsterType call region ✓; items.cpp:1475 ItemRndDur ✓, :3195 SpawnUnique ✓, :3422 SpawnItem ✓ — ALL exact; towners.cpp:760 InitTowners ✓, :811 ProcessTowners ✓; townerdat animOrder column real (col 11, 100+ entries/NPC verified live). ONE cite error: "themes.cpp:349-449" for theme placement — :349 is PlaceThemeMonsts; InitThemes is at :826 (Theme_* :377-785, CreateThemeRooms :913). ACs (10-seed theme parity, 3-seed spawn lists, towner tick order) testable dump-comparisons.
3 online: n/a.
4 solution: wire r462 (+v791/778h via their own fixes — population consumes both); fix the themes.cpp cite.
5 understanding: the integration leaf of epic-04 — pulls 778h tables + v791 placement + r462 entities + 4ybb levels into a populated world; correctly ordered last in the epic.
6 repro: `sed -n '195,196p;282,284p' devilutionX/Source/monster.cpp; sed -n '1475p;3195p;3422p' devilutionX/Source/items.cpp; grep -n "void InitThemes\|void PlaceThemeMonsts" devilutionX/Source/levels/themes.cpp; awk -F'\t' 'NR==1{for(i=1;i<=NF;i++) if($i=="animOrder") print i}' devilutionX/assets/txtdata/towners/towners.tsv`
7 fix: `beans update devilGo-foug --blocked-by devilGo-r462`; `beans update devilGo-foug --body-replace-old "themes.cpp:349-449;" --body-replace-new "themes.cpp:826-880 (InitThemes; PlaceThemeMonsts :349, Theme_* :377-785, CreateThemeRooms :913);"`
8 validation: foug blockedByIds includes r462; body cites :826.
9 related: 4ybb/r462 (deps), 778h/v791 (upstream, wired via their fixes), h6u0 (towner content), 97mn epic ACs 2-4.
VERDICT: NEEDS-WORK:wire-deps (r462 missing) + 1 wrong cite — ACs and remaining research verified.


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).

ID MAP (fix2): all legacy dg-task-NNN references in this bean map to devilGo ids via docs/research/dg-id-map.json (canonical mapping in milestone devilGo-i56j). Operative deps = frontmatter (source of truth); prose refs are documentation.
