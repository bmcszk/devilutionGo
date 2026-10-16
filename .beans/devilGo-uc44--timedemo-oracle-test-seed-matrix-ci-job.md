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
updated_at: 2026-10-16T14:06:45Z
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


## IMPL NOTES (triage 2026-10-16)
1 root: CI oracle wiring — run demo fixtures, compare hero saves via pfile_compare parity, seed matrix <5min; status `draft`, parent 1n0a, blocked_by w9or ✓ (demo engine first).
2 local: `test/timedemo_test.cpp` — MPQ-skip guard ~:47-52 (HaveMainData → GTEST_SKIP, cited :47-49, regionally correct), `pfile_compare_hero_demo` at :94 with `ASSERT_EQ(result.status, Same)` ~:96-100 ✓; implementation `Source/pfile.cpp:665` ✓; fixture matrix on disk ✓. Body deps cite stale `dg-task-044, dg-epic-08` (044 = demo engine → w9or, correctly in frontmatter; epic-08 = persistence → devilGo-6xu2, MISSING from frontmatter despite the oracle requiring the .sv compare path). AC list has no explicit TDD-first item.
3 online: n/a (no external claims).
4 solution: add devilGo-6xu2 (or its .sv-compare child) to frontmatter deps; refresh stale ids; make TDD-first explicit (e.g. "- [ ] TDD: oracle red against a deliberately-corrupted spawn_0.sv copy before wiring CI").
5 understanding: the end-state oracle of the whole program — everything upstream must make this green; consumes w9or playback + persistence compare.
6 repro: `sed -n '47,52p;92,100p' devilutionX/test/timedemo_test.cpp`; `beans show devilGo-uc44 --json` (blockedByIds: [w9or] only).
7 fix: `beans update devilGo-uc44 --blocked-by devilGo-6xu2`; body replace "dg-task-044, dg-epic-08." → "devilGo-w9or, devilGo-6xu2."; append explicit TDD-first AC as above.
8 validation: `beans query --json '{ bean(id:"devilGo-uc44") { blockedByIds } }'` includes 6xu2; body grep TDD.
9 related: devilGo-w9or (dep), devilGo-6xu2 (persistence), devilGo-i2cr (sibling CI gate).
VERDICT: NEEDS-WORK:frontmatter-missing-persistence-dep;stale-dg-task-044/dg-epic-08;no-explicit-TDD-first-AC
