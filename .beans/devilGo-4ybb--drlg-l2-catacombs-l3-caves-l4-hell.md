---
# devilGo-4ybb
title: DRLG L2 Catacombs + L3 Caves + L4 Hell
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - drlg
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-16T14:16:44Z
parent: devilGo-97mn
blocked_by:
    - devilGo-76tu
---

# Remaining generators

## Goal
L2/L3/L4 parity.

## Research
drlg_l2.cpp:1780-1786 room sizes, :2830/:2839; drlg_l3.cpp:2185; drlg_l4.cpp:1228; fixtures levels/l{2,3,4}data + diablo/hellfire sets; tests drlg_l2..l4_test.cpp.

## Scope
internal/world/drlg_l{2,3,4}.go.

## Acceptance criteria
- [ ] All L2/L3/L4 fixtures byte-equal (diablo + hellfire)
- [ ] River/lava constraints (L3/L4) verified on 20 random seeds vs C++ dumps

## Skills for coder
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — fixture matrix

### game-dev skills (2026-10-16 audit)
- /home/blaze/.agents/skills/game-developer/SKILL.md — procedural generation architecture
- /home/blaze/.agents/skills/game-development/game-design/SKILL.md — level design principles

## Dependencies & blockers
dg-task-017 (shared engine).

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK


## IMPL NOTES (triage 2026-10-16)
1 root: Phase-2 generator task (dg-task-018): L2/L3/L4 parity; draft; wire blockedBy [76tu] ✓ = body's dg-task-017 shared engine.
2 local: ALL four cites verified exact: drlg_l2.cpp:1780-1786 (RoomMax=10/RoomMin=4) ✓, :2830 CreateL2Dungeon ✓, :2839 LoadPreL2Dungeon ✓, drlg_l3.cpp:2185 CreateL3Dungeon ✓, drlg_l4.cpp:1228 CreateL4Dungeon ✓; fixtures levels/l{2,3,4}data ✓ on disk; drlg_l2..l4_test.cpp ✓ on disk. ACs (byte-equal fixtures diablo+hellfire; 20-seed river/lava vs C++ dumps) are TDD-first and testable.
3 online: n/a.
4 solution: none — ready.
5 understanding: world layer; reuses 76tu's engine; the 20-seed C++-dump AC is the strongest form of parity for non-fixture paths.
6 repro: `sed -n '1780,1786p;2830,2843p' devilutionX/Source/levels/drlg_l2.cpp; sed -n '2185,2187p' devilutionX/Source/levels/drlg_l3.cpp; sed -n '1228,1230p' devilutionX/Source/levels/drlg_l4.cpp; ls devilutionX/test/fixtures/levels`
7 fix: none.
8 validation: n/a.
9 related: 76tu (dep), foug (dependent), vti7/m2x7 (transitive).
VERDICT: TRIAGED-OK


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.
