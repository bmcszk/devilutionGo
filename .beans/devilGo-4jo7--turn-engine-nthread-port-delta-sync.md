---
# devilGo-4jo7
title: Turn engine (nthread port) + delta sync
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-3
    - net
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-16T14:01:03Z
parent: devilGo-zo8j
blocked_by:
    - devilGo-omu6
---

# Lockstep

## Goal
Turn cadence, in-transit batching, delta export/import.

## Research
nthread.cpp:80-176; base.h PlayerState; msg.cpp:218-275 deltas, :698; multi.cpp:665.

## Scope
internal/net/turn.go, delta.go integrated with engine tick.

## Acceptance criteria
- [ ] Epic's 10k-tick zero-divergence and converge tests green
- [ ] Delta payload byte-size within C++ norms for same scenarios

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — channel-based turn pump, deterministic apply order

## Dependencies & blockers
dg-task-038, dg-task-039.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:missing-blocked-by-mfds + wrong line cite (698→2774)
