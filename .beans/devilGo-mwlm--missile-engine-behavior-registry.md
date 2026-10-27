---
# devilGo-mwlm
title: Missile engine + behavior registry
status: in-progress
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - missiles
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-21T11:07:45Z
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
- [ ] misdat.tsv row count == registry size (no missing) — fix2: count asserted from on-disk TSV at runtime, never hardcoded.
- [ ] missiles_test port passes
- [ ] 16.16 traveled arithmetic property-tested vs C++ probe

## Libraries
- stdlib only (no third-party deps at this stage)

## Skills for coder
- \`~/.agents/skills/coding-rules/SKILL.md\` — read fully BEFORE acting
- \`~/.agents/skills/ponytail/SKILL.md\` — read fully BEFORE acting
/home/blaze/.agents/skills/go-dev/SKILL.md — registry pattern, no reflection

## Dependencies & blockers
devilGo-7be9.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:AC1-row-count==registry-size-is-false (verified against disk; would fail a correct port)


## IMPL NOTES (triage 2026-10-16)
1 root: missile runtime port (motion/collision/behaviors); = devilGo-mwlm; predates the data split (vhgd).
2 local: beans show ✓; misdat.h:166-170 MissileData AddFn/ProcessFn fn-pointers ✓; CheckMissileCol missiles.cpp:479 ✓; MoveMissileAndCheckMissileCol :636 ✓; test/missiles_test.cpp exists ✓.
3 online: n/a
4 solution: AC1 "misdat.tsv row count == registry size (no missing)" is FALSE as written — disk: 68 base/108 hf rows vs 51/74 distinct addFn and 37/46 distinct processFn (measured with cut|sort -u), and MissileID enum = 111 ids vs 68 tabled base rows. Replace with vhgd's name-set criterion. Also state registry ownership: vhgd owns table+registry shape, this bean owns runtime implementations (vhgd already documents the split; mirror it here). Remap deps (devilGo-7be9→devilGo-7be9) + add blocked_by devilGo-vhgd; S2.
5 understanding: game-layer runtime consuming diablo-layer misdat registry (vhgd); spawns feed quests/cow content (lsz6, h6u0).
6 repro: beans show devilGo-mwlm; tail -n +2 devilutionX/assets/txtdata/missiles/misdat.tsv | cut -f2 | sort -u | wc -l (=51); cut -f3 (=37); same on mods/hf (=74/46)
7 fix: beans update devilGo-mwlm --blocked-by devilGo-7be9, then --blocked-by devilGo-vhgd; body-replace-old "misdat.tsv row count == registry size (no missing)" --body-replace-new "every misdat.tsv addFn/processFn name resolves to a registered implementation (51/37 distinct base, 74/46 hf per devilGo-vhgd); MissileID-enum (111) vs tabled-rows (68/108) gap explicitly enumerated".
8 validation: beans show devilGo-mwlm shows corrected AC + both edges.
9 related: devilGo-vhgd, devilGo-7be9, devilGo-lsz6, devilGo-h6u0
VERDICT: NEEDS-WORK:AC1-row-count==registry-size-is-false (verified against disk; would fail a correct port)

AC1 row-count claim FALSE vs disk (triage-verified). AC rewritten: row count must be asserted from the on-disk TSV at test runtime, not hardcoded.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).

## Progress log

- 2026-10-18 claimed: slice 1 dispatched via pi (_glm_flash_first, contract /tmp/dg-mwlm1-contract.md). Attempt 1 timeout-killed at 900 s mid-healthy-run (wrapper budget, not pi); attempt 2 completed EXIT:0.
- 2026-10-18 slice 1 verified by parent: 7/7 targeted tests PASS (RegistryCoversMisdat counts corpus at runtime; Motion16Dot16; MidPosition; PathDistance), gofmt/vet/lint clean, 19 pkgs ok. Committed a2b723b (pending push).
- Remaining: slice 2 = CheckMissileCol + missiles_test.cpp rotation port + property test; slice 3 = behavior logic for stubbed ids; migrate global Rows slice into MissileList struct.
- 2026-10-18 slice 2 verified by parent: CheckMissileCol port (319 lines, zero logic changes by completers), rotation.go 1:1 helper, tape-driven rotation tests, 13 collision branch tests (wall-stop world-grid, hit-callback spies, flag table w/ missiles.cpp:544 semantics), Rows global migrated into MissileList payload. 22/22 package tests, 19 pkgs ok, vet/lint/gofmt clean. RED evidence reviewed (inverted-expectation probe, reverted).
- Remaining: slice 3 = behavior logic for stubbed registry ids + DeleteMissiles GC pass.
- 2026-10-18 slice 3a-beta verified by parent: 8 process fns (UpdateMissileVelocity/PutMissile/UpdateMissilePos/MoveMissilePos/MoveMissile/MoveMissileAndCheckMissileCol/ProcessArrow/ProcessGenericProjectile) + 5 ids wired real (Firebolt MagmaBall BloodStar Acid Arrow). 42/42 package tests, 19 pkgs ok, gates clean. Session-persistent run (devilgo-qna9) — resume directives worked; sim-then-write pattern observed.
- Remaining: 3a-ii (ElementalArrow/MissileExplosion/Fireball/ChargedBolt/HolyBolt), 3b specials, 3c spawn chains, DeleteMissiles GC.
- 2026-10-18 slice 3a-ii verified by parent: ProcessElementalArrow/Fireball/MissileExplosion/ChargedBolt/HolyBolt + 5 ids wired real. 56/56 package tests, 19 pkgs ok, gates clean. Session devilgo-qna9 carried context across fix cycles (lint -> test fix -> lint -> test fix).
- Remaining: 3b specials (NovaBall/Guardian/Teleport/Lightning/Apocalypse...), 3c spawn chains, DeleteMissiles GC.

## Zero-dependency test constraint\n\nReplace live-corpus/upstream-TSV reads with committed Go-owned fixtures. Standard Go tests must not use DG_DEVILUTIONX, locate the devilutionX tree, or skip when upstream files are absent. Missing or corrupt committed fixtures must fail the test. C++ probes are offline fixture producers only.
