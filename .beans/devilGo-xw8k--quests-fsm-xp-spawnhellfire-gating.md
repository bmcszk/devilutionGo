---
# devilGo-xw8k
title: Quests FSM + XP + spawn/hellfire gating
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - quests
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-16T20:38:07Z
parent: devilGo-nxxo
blocked_by:
    - devilGo-r462
    - devilGo-pu3r
---

# Quests & progression

## Goal
Quest lifecycle, kill hooks, multiplayer sync vars, XP table, gating flags.

## Research
questdat.hpp:102-129; quests.cpp:60-110,227,397,610; Experience.tsv; gating monster.cpp:3159, player.cpp:3479,3688-3701.

## Scope
internal/gameplay/quests.go, xp.go; SimFlags{Spawn,Hellfire,Multiplayer} helper API.

## Acceptance criteria
- [ ] quests_test port passes
- [ ] Level-15 seed quest pools equal C++ on 10 seeds
- [ ] XP thresholds + stat/point award flow

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — small FSM clarity

## Dependencies & blockers
dg-task-016, dg-task-008.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:missing-deps-devilGo-lsz6+devilGo-pu3r (+remap)


## IMPL NOTES (triage 2026-10-16)
1 root: quest runtime FSM, kill hooks, XP table, gating flags; = dg-task-026; runtime half of lsz6.
2 local: beans show ✓; quests.cpp cites consistent with lsz6 verification (:60,:88-92,:227 region); quests_test.cpp exists ✓; Experience.tsv present in assets/txtdata ✓; deps dg-task-016→r462 live ✓, dg-task-008→zvwd live ✓.
3 online: n/a
4 solution: missing deps on its own data suppliers — devilGo-lsz6 (questdat tables; lsz6 documents "FSM runtime stays dg-task-026" but xw8k doesn't reference back) and devilGo-pu3r (Experience.tsv rows; AC "XP thresholds" would otherwise re-transcribe pu3r's table). Remap; S2; mirror the split note.
5 understanding: game-layer runtime over diablo-layer quest data; SimFlags{Spawn,Hellfire,Multiplayer} helper matches epic scope.
6 repro: beans show devilGo-xw8k; ls devilutionX/assets/txtdata/Experience.tsv devilutionX/test/quests_test.cpp
7 fix: beans update devilGo-xw8k --blocked-by devilGo-r462, then --blocked-by devilGo-lsz6, then --blocked-by devilGo-pu3r; body-append id map + note "questdat/XP data owned by devilGo-lsz6 / devilGo-pu3r".
8 validation: beans show devilGo-xw8k lists three real edges.
9 related: devilGo-lsz6, devilGo-pu3r, devilGo-r462, devilGo-zvwd, devilGo-orkj
VERDICT: NEEDS-WORK:missing-deps-devilGo-lsz6+devilGo-pu3r (+remap)


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).

ID MAP (fix2): all legacy dg-task-NNN references in this bean map to devilGo ids via docs/research/dg-id-map.json (canonical mapping in milestone devilGo-i56j). Operative deps = frontmatter (source of truth); prose refs are documentation.
