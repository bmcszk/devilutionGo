---
# devilGo-w9or
title: Demo/timedemo record-replay engine
status: draft
type: task
priority: normal
tags:
    - plan:task
    - parity
    - demo
created_at: 2026-10-15T16:35:18Z
updated_at: 2026-10-16T14:01:06Z
parent: devilGo-1n0a
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
dg-epic-01 loop; full value after epic-08.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:frontmatter-empty-deps-vs-body-deps;no-explicit-TDD-first-AC;stale-dg-epic-refs
