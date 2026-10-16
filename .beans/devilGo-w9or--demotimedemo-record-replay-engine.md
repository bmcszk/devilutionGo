---
# devilGo-w9or
title: Demo/timedemo record-replay engine
status: todo
type: task
priority: normal
tags:
    - plan:task
    - parity
    - demo
created_at: 2026-10-15T16:35:18Z
updated_at: 2026-10-16T20:38:07Z
parent: devilGo-1n0a
blocked_by:
    - devilGo-ohsu
---

# Demo

## Goal
Record cmd stream + input events; deterministic playback; --timedemo uncapped.

## Research
demomode.cpp:39 Version 3, :43-58 event structs, :595-599 record bytes (ProgressToNextGameTick|0x80), :626/:674; dmo fixtures; help text diablo.cpp:1075-1079.

## Scope
internal/demo: reader/writer, playback driving engine ticks, cmd injection points.

## Acceptance criteria
- [ ] Record→replay of a 60s scripted session reproduces identical StateHash at 10 checkpoints
- [ ] Version byte 3; format doc in package

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — deterministic serialization

## Dependencies & blockers
- devilGo-xcdz (fix2: missing edge wired)
dg-epic-01 loop; full value after epic-08.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:frontmatter-empty-deps-vs-body-deps;no-explicit-TDD-first-AC;stale-dg-epic-refs


## IMPL NOTES (triage 2026-10-16)
1 root: Port demomode record/replay (Version 3 format) driving engine ticks; status `draft`, parent 1n0a, frontmatter `blocked_by` EMPTY while body declares deps.
2 local: `Source/engine/demomode.cpp` — Version=3 at ~line 47 (bean cites :39; drifted, value correct), event structs 51+ ✓, `ProgressToNextGameTick | 0b10000000` at :595 ✓, InitRecording ~:626 ✓; `Source/diablo.cpp:1075-1079` timedemo/demo/record flags ✓ and `--timedemo` at :1033/:1127; fixtures `test/fixtures/timedemo/WarriorLevel1to2/{demo_0.dmo, spawn_0.sv, demo_0_reference_spawn_0.sv}` ✓. Body deps cite stale `dg-epic-01 loop` / `epic-08`; real deps are the tick loop (devilGo-ohsu) and StateHash/RunDemoTape harness (devilGo-xcdz) — neither in frontmatter.
3 online: n/a (no external claims).
4 solution: add frontmatter deps (ohsu, xcdz), refresh stale epic refs; AC list lacks an explicit TDD-first item (record-format golden byte-dump vs C++ demo_0.dmo would be the natural red-first test — fixtures already on disk, so make it explicit).
5 understanding: the deterministic-playback substrate uc44's oracle and every StateHash parity test ride on; engine/game layer.
6 repro: `sed -n '47p;595,596p' devilutionX/Source/engine/demomode.cpp`; `ls devilutionX/test/fixtures/timedemo/WarriorLevel1to2/`; `beans show devilGo-w9or --json` (blockedByIds empty).
7 fix: `beans update devilGo-w9or --blocked-by devilGo-ohsu --blocked-by devilGo-xcdz` (CLI equivalent: set both); body replace "dg-epic-01 loop; full value after epic-08" → "devilGo-ohsu loop; devilGo-xcdz StateHash; full value after devilGo-6xu2 persistence"; append explicit TDD-first AC: "- [ ] TDD: parsed demo_0.dmo golden (events+timing bytes) asserted before record path exists".
8 validation: `beans query --json '{ bean(id:"devilGo-w9or") { blockedByIds } }'` shows both; body grep TDD present.
9 related: devilGo-uc44 (child consumer), devilGo-7aul (timedemo-cited fixtures), devilGo-6xu2 (hero compare).
VERDICT: NEEDS-WORK:frontmatter-empty-deps-vs-body-deps;no-explicit-TDD-first-AC;stale-dg-epic-refs


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).

ID MAP (fix2): all legacy dg-task-NNN references in this bean map to devilGo ids via docs/research/dg-id-map.json (canonical mapping in milestone devilGo-i56j). Operative deps = frontmatter (source of truth); prose refs are documentation.
