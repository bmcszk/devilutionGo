---
# devilGo-0tsf
title: Options + ini import + SimOptions freeze
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-3
    - platform
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-15T16:45:27Z
parent: devilGo-yjuv
---

# Options

## Goal
Typed options tree, persistence, migration, sim snapshot.

## Research
options.h/cpp structure; audio callbacks sound.cpp:183-185; ini_test.cpp.

## Scope
internal/platform/options.go, ini_import.go.

## Acceptance criteria
- [ ] ini_test port green
- [ ] TickRate option maps 20..50 like gamemenu slider; audio rate {22050,44100,48000}

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — config struct + validation

## Dependencies & blockers
None.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.
