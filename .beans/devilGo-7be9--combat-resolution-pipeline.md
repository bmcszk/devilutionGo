---
# devilGo-7be9
title: Combat resolution pipeline
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - combat
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-16T14:26:52Z
parent: devilGo-nxxo
blocked_by:
    - devilGo-r462
---

# Hit & damage

## Goal
Melee/ranged/magic hit rolls + damage with exact roll order and 1/64 fixed point.

## Research
player.cpp:524-660 (see epic); player.h:575-666; RandomIntBetween random.hpp:403.

## Scope
internal/gameplay/combat.go: PlrHitMonst, PlrHitPlr(-PvP), MonstHitPlayer port, ApplyPlrDamage/DamageMonster, durability hooks.

## Acceptance criteria
- [ ] Roll-tape parity harness green for 3 classes × 5 monsters × 1000 seeds
- [ ] Damage in 1/64 units asserted (no float anywhere)
- [ ] Clamp 5..95 and petrified hit=0 rules tested

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — explicit int math
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — tape comparison

## Dependencies & blockers
dg-task-016.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:remap-deps (S1/S2 only; cites verified exact)


## IMPL NOTES (triage 2026-10-16)
1 root: melee/ranged/magic hit+damage port with exact roll order; = dg-task-021.
2 local: beans show ✓; PlrHitMonst player.cpp:528 (cited range :524-660 ✓ covers); clamp(hper,5,95) :548 ✓ and MonstHitPlayer :715 ✓ (cited :712 = GenerateRnd(100) roll); Petrified early-return :543 ✓; player.h:575-580 to-hit ✓; RandomIntBetween Source/engine/random.hpp:403 ✓ (exact line); test harness = own tape (dg-sub-0005 lineage).
3 online: n/a
4 solution: remap deps (dg-task-016→devilGo-r462); S2 pkg root; optionally pin tape fixture location (test/parity pattern per AGENTS.md).
5 understanding: game-layer combat formulas over entity structs (r462); consumes RNG (m2x7); feeds missile engine (mwlm) and saves.
6 repro: beans show devilGo-7be9; sed -n '524,560p;710,716p' devilutionX/Source/player.cpp; sed -n '403p' devilutionX/Source/engine/random.hpp
7 fix: beans update devilGo-7be9 --blocked-by devilGo-r462; body-append id map.
8 validation: beans show lists blocked_by devilGo-r462.
9 related: devilGo-r462, devilGo-mwlm, devilGo-m2x7
VERDICT: NEEDS-WORK:remap-deps (S1/S2 only; cites verified exact)


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).
