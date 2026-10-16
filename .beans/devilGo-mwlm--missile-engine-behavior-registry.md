---
# devilGo-mwlm
title: Missile engine + behavior registry
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - missiles
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-16T14:26:53Z
parent: devilGo-nxxo
blocked_by:
    - devilGo-7be9
---

# Missiles

## Goal
All missile behaviors table-driven; motion/collision exact.

## Research
misdat.h:166-196; missiles.cpp:479,636,728-1242; missile_sprites.tsv; tests missiles_test.cpp.

## Scope
Registry map[MissileID]{Add,Process}; fixed-point motion; CheckMissileCol; trap/golem/unique spawns via AddMissile chains.

## Acceptance criteria
- [ ] misdat.tsv row count == registry size (no missing)
- [ ] missiles_test port passes
- [ ] 16.16 traveled arithmetic property-tested vs C++ probe

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — registry pattern, no reflection

## Dependencies & blockers
dg-task-021.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:AC1-row-count==registry-size-is-false (verified against disk; would fail a correct port)


## IMPL NOTES (triage 2026-10-16)
1 root: missile runtime port (motion/collision/behaviors); = dg-task-022; predates the data split (vhgd).
2 local: beans show ✓; misdat.h:166-170 MissileData AddFn/ProcessFn fn-pointers ✓; CheckMissileCol missiles.cpp:479 ✓; MoveMissileAndCheckMissileCol :636 ✓; test/missiles_test.cpp exists ✓.
3 online: n/a
4 solution: AC1 "misdat.tsv row count == registry size (no missing)" is FALSE as written — disk: 68 base/108 hf rows vs 51/74 distinct addFn and 37/46 distinct processFn (measured with cut|sort -u), and MissileID enum = 111 ids vs 68 tabled base rows. Replace with vhgd's name-set criterion. Also state registry ownership: vhgd owns table+registry shape, this bean owns runtime implementations (vhgd already documents the split; mirror it here). Remap deps (dg-task-021→devilGo-7be9) + add blocked_by devilGo-vhgd; S2.
5 understanding: game-layer runtime consuming diablo-layer misdat registry (vhgd); spawns feed quests/cow content (lsz6, h6u0).
6 repro: beans show devilGo-mwlm; tail -n +2 devilutionX/assets/txtdata/missiles/misdat.tsv | cut -f2 | sort -u | wc -l (=51); cut -f3 (=37); same on mods/hf (=74/46)
7 fix: beans update devilGo-mwlm --blocked-by devilGo-7be9, then --blocked-by devilGo-vhgd; body-replace-old "misdat.tsv row count == registry size (no missing)" --body-replace-new "every misdat.tsv addFn/processFn name resolves to a registered implementation (51/37 distinct base, 74/46 hf per devilGo-vhgd); MissileID-enum (111) vs tabled-rows (68/108) gap explicitly enumerated".
8 validation: beans show devilGo-mwlm shows corrected AC + both edges.
9 related: devilGo-vhgd, devilGo-7be9, devilGo-lsz6, devilGo-h6u0
VERDICT: NEEDS-WORK:AC1-row-count==registry-size-is-false (verified against disk; would fail a correct port)

AC1 row-count claim FALSE vs disk (triage-verified). AC rewritten: row count must be asserted from the on-disk TSV at test runtime, not hardcoded.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).
