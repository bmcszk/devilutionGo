---
# devilGo-uc44
title: Timedemo oracle test + seed matrix CI job
status: draft
type: task
priority: normal
tags:
    - plan:task
    - parity
    - ci
created_at: 2026-10-15T16:35:18Z
updated_at: 2026-10-16T14:01:05Z
parent: devilGo-1n0a
blocked_by:
    - devilGo-w9or
---

# The oracle

## Goal
Wire reference-save comparison + fixture matrix into CI.

## Research
timedemo_test.cpp:47-49,96-100; fixtures listing; pfile_compare_hero_demo.

## Scope
internal/demo/timedemo_test.go; justfile parity target; CI workflow.

## Acceptance criteria
- [ ] Green with spawn.mpq present; graceful skip otherwise
- [ ] Matrix run < 5min CI

## Skills for coder
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — CI-friendly skips and reports

## Dependencies & blockers
dg-task-044, dg-epic-08.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:frontmatter-missing-persistence-dep;stale-dg-task-044/dg-epic-08;no-explicit-TDD-first-AC
