---
# devilGo-xcdz
title: Golden-test harness scaffolding
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-0
    - testing
created_at: 2026-10-15T16:35:16Z
updated_at: 2026-10-15T16:45:26Z
parent: devilGo-qna9
blocked_by:
    - devilGo-ohsu
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

## Skills for coder
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — harness structure
/home/blaze/.agents/skills/coding-rules/SKILL.md — one obvious way to load fixtures

## Dependencies & blockers
dg-task-004.

## Open questions
- StateHash field list v1: players/monsters/items/quests/seeds — finalize when entities land (dg-epic-04).


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.
