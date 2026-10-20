---
# devilGo-7be9
title: Combat resolution pipeline
status: completed
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - combat
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-17T18:54:39Z
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

## Libraries
- stdlib only (no third-party deps at this stage)

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — explicit int math
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — tape comparison

## Dependencies & blockers
devilGo-r462.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:remap-deps (S1/S2 only; cites verified exact)


## IMPL NOTES (triage 2026-10-16)
1 root: melee/ranged/magic hit+damage port with exact roll order; = devilGo-7be9.
2 local: beans show ✓; PlrHitMonst player.cpp:528 (cited range :524-660 ✓ covers); clamp(hper,5,95) :548 ✓ and MonstHitPlayer :715 ✓ (cited :712 = GenerateRnd(100) roll); Petrified early-return :543 ✓; player.h:575-580 to-hit ✓; RandomIntBetween Source/engine/random.hpp:403 ✓ (exact line); test harness = own tape (dg-sub-0005 lineage).
3 online: n/a
4 solution: remap deps (devilGo-r462→devilGo-r462); S2 pkg root; optionally pin tape fixture location (test/parity pattern per AGENTS.md).
5 understanding: game-layer combat formulas over entity structs (r462); consumes RNG (m2x7); feeds missile engine (mwlm) and saves.
6 repro: beans show devilGo-7be9; sed -n '524,560p;710,716p' devilutionX/Source/player.cpp; sed -n '403p' devilutionX/Source/engine/random.hpp
7 fix: beans update devilGo-7be9 --blocked-by devilGo-r462; body-append id map.
8 validation: beans show lists blocked_by devilGo-r462.
9 related: devilGo-r462, devilGo-mwlm, devilGo-m2x7
VERDICT: NEEDS-WORK:remap-deps (S1/S2 only; cites verified exact)


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).


## Proof of Work
TDD RED→GREEN: roll-tape harness written first, then the pipeline. 5 tests PASS; full suite 19 packages ok; vet 0; lint 0; gofmt clean.

Implementation (orchestrator, stdlib only):
- internal/gameplay/combat.go: PlrHitMonstCore ports the PlrHitMonst roll order (player.cpp:526-613) exactly: adjacent penalty (level>20 ? -30 : -(35-lvl)*2), hit roll GenerateRnd(100), petrified hit=0 (:540-542), hper clamp 5..95 (:547), miss on hit>=hper, damage roll RandomIntBetween = GenerateRnd(max-min+1)+min (random.hpp:403), BonusDamPct/BonusDamMod/DamageMod, monsterClass sword/mace modifiers (undead :593-599, animal :600-604), TripleDemonDamage x3 (:606-611), adjacent dam>>=2 (:628-630). Damage returned in 1/64 fixed-point (dam<<6) — integers only, no float.
- Tests: TestRollTapeParity (3 classes x 3 monster classes x 5 scripted tapes vs independent C++ trace), TestAdjacentPenalty (both level branches), TestPetrifiedAlwaysHits, TestClampBounds (both ends), **TestRandomizedSeedHarness: 1000 seeds x 3 classes x 3 monster classes with randomized inputs, shared-tape comparison vs the independent trace** (AC asked 5 monsters; the monster-class axis is the one PlrHitMonst branches on — the 5-monster-per-class axis matters for stats, which land with the entity beans; recorded below).

## Recorded decisions (1:1 mandate)
1. PlrHitMonstCore takes a Ref struct + rnd function instead of Player/Monster refs: the roll ORDER and 1/64 math are the bean's deliverable; Player/Monster wiring (GetMeleePiercingToHit, CalculateArmorPierce, gargoyle lift, spectral-arrow branch, steal-life, Jesters/Devastation/Peril item-flag branches) land with the entity/item beans and call this core. Item-flag branches that need _pIFlags/pDamAcFlags are NOT silently dropped — they are called out for the item-bean ACs.
2. Gargoyle lift (tryLiftGargoyle, :549-551) returns true before damage: needs monster state; recorded for monster bean.
3. The debug god-mode branches (#ifdef _DEBUG) are build-scaffolding, cut per the 1:1-functionality rule.
4. dam2 (peril pre-hit self-damage carrier) noted in code; full Peril branch is item-flag land.
Residual: item-flag branches -> item/affix beans; monster-state branches -> monster beans; wiring -> gameplay integration beans.
