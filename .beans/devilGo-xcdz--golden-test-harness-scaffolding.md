---
# devilGo-xcdz
title: Golden-test harness scaffolding
status: in-progress
type: task
priority: normal
tags:
    - plan:task
    - phase-0
    - testing
created_at: 2026-10-15T16:35:16Z
updated_at: 2026-10-20T00:40:35Z
parent: devilGo-qna9
blocked_by:
    - devilGo-ohsu
    - devilGo-1spo
---

# Parity harness foundations

## Goal
Test helpers every later epic uses: fixture loader (copies of devilutionX test/fixtures), state-hash function, seed-dungeon comparison, demo-replay runner interface.

## Research
- Fixtures: test/fixtures/{diablo,hellfire,levels,txtdata,timedemo}; timedemo oracle test/timedemo_test.cpp:28-110 uses pfile_compare_hero_demo vs demo_0_reference_spawn_0.sv.
- .dun layout: 2×u16 LE dims then arrays (verified hexdump of 1-2588.dun).

## Scope
internal/testsupport: LoadDunFixture(seed), StateHash(GameState) (FNV over canonical fields), CompareGrids, RunDemoTape interface; Makefile target `parity` placeholder.

## Acceptance criteria
- [ ] LoadDunFixture parses all test/fixtures/diablo/*.dun (dims plausible 40x40..112x112 semantics documented)
- [ ] StateHash stable across runs and process restarts (no map iteration)
- [ ] Harness doc: how to regenerate fixtures from C++ build


## Libraries
- github.com/google/go-cmp — deep-diff golden comparisons; testify optional

## Skills for coder
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — harness structure
/home/blaze/.agents/skills/coding-rules/SKILL.md — one obvious way to load fixtures

## Dependencies & blockers
devilGo-ohsu.

## Open questions
OPEN-QUESTION RESOLUTION 2026-10-16: DECIDED — StateHash v1 field list fixed: players, monsters, items, quests, seeds (the five systems the parity harness pins). Additions require a milestone decision, not a silent field add.
- StateHash field list v1: players/monsters/items/quests/seeds — finalize when entities land (dg-epic-04).


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK


## IMPL NOTES (triage 2026-10-16)
1 root: testsupport primitives (dun fixture loader, StateHash FNV, CompareGrids, demo-tape interface); status `draft`, parent qna9, blocked_by ohsu (frontmatter ✓; body says devilGo-ohsu stale).
2 local: fixtures verified — test/fixtures/{diablo,hellfire,levels,txtdata,timedemo} all exist; 1-2588.dun 54404 bytes, hexdump first bytes `2800 2800` = 40×40 ✓ (2×u16 LE dims claim + "40x40..112x112" plausible); timedemo oracle verified — timedemo_test.cpp:94 `pfile_compare_hero_demo(demoNumber, true)` and fixtures/timedemo/WarriorLevel1to2/{demo_0.dmo, demo_0_reference_spawn_0.sv, spawn_0.sv} exist. Stale bits: "Makefile target parity" — repo uses just and justfile already has test-parity (:117).
3 online: n/a (no external claims).
4 solution: ready after promotion; cleanups: devilGo-ohsu → devilGo-ohsu, Makefile → justfile parity target.
5 understanding: test-layer foundation consumed by 1n0a (testing epic extends internal/testsupport) and 7aul (pixel comparator); distinct scope from both — no duplication.
6 repro: `beans show devilGo-xcdz --body-only`; `xxd …/test/fixtures/diablo/1-2588.dun | head -1`; `sed -n '94p' …/test/timedemo_test.cpp`; `ls …/test/fixtures/timedemo/WarriorLevel1to2/`; `sed -n '13,16p' .beans/devilGo-xcdz*.md`.
7 fix: `beans update devilGo-xcdz --body-replace-old "devilGo-ohsu." --body-replace-new "devilGo-ohsu."` and `--body-replace-old "Makefile target \`parity\` placeholder" --body-replace-new "\`just test-parity\` target (justfile:117)"`.
8 validation: `beans show devilGo-xcdz --body-only | grep dg-task` → empty; grep Makefile → empty.
9 related: ohsu (blocker), 1n0a (consumer, epic), 7aul (sibling comparator, NOT a duplicate).
VERDICT: TRIAGED-OK

## Progress log
- 2026-10-20 claimed by orchestrator after 1spo close (93ff0f5). Blockers vsx3/ohsu/1spo all completed.

## Progress log
- 2026-10-20 slice 1 landed: internal/testsupport dun.go+grids.go, LoadDunFixture + CompareGrids, 3 tests (parse-all goldens, CompareGrids, l1-main-2588 byte round-trip). Worker proc_b49b847c9727 exit 0, gates green, 20 pkgs ok.
