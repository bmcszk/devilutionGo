---
# devilGo-w6qf
title: Benchmark suite + perf budgets in CI
status: todo
type: task
priority: normal
tags:
    - plan:task
    - parity
    - bench
created_at: 2026-10-15T16:35:18Z
updated_at: 2026-10-16T20:38:07Z
parent: devilGo-1n0a
---

# Bench

## Goal
Ports of the six C++ benchmarks with budgets.

## Research
clx_render/dun_render/light_render/crawl/path/palette_blending benchmarks; ebiten perf tips.

## Scope
internal/**/*_bench.go + benchstat tracking in CI.

## Acceptance criteria
- [ ] Frame <3ms, path 10k<50ms, crawl 100k<30ms on CI runner; regressions >10% flagged

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — benchmark correctness (b.ReportAllocs, ResetTimer)

## Dependencies & blockers (fix2: edges now wired — see frontmatter; earlier "wired" stamp was premature)
Depends on epics 02/03 progressively.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:no-explicit-TDD-first-AC;absolute-budgets-on-unpinned-CI-runner;stale-dg-epic-refs


## IMPL NOTES (triage 2026-10-16)
1 root: Port the six devilutionX benchmarks with absolute budgets + benchstat regression flagging; status `draft`, parent 1n0a, frontmatter `blocked_by` EMPTY.
2 local: All six benchmark sources exist: `test/{clx_render,crawl,dun_render,light_render,palette_blending,path}_benchmark.cpp` + `docs/benchmarking.md` (ls verified). Skills path `/home/blaze/.agents/skills/go-dev/SKILL.md` exists. Body dep line says "Depends on epics 02/03" — stale pre-migration ids (should be devilGo-5wsb / devilGo-qbiy).
3 online: n/a (no external claims).
4 solution: add explicit TDD-first AC (port each C++ benchmark's op/fixture list first, assert parity of workload before budgets), pin the CI runner class for absolute budgets ("<3ms on CI runner" is machine-dependent — a blocking gate needs a named runner or a statistical procedure), refresh dg-epic-02/03 refs.
5 understanding: engine/game-layer perf gate; consumed by epic qbiy's "<3ms frame" AC — this bean is its enforcement mechanism.
6 repro: `ls <devilutionX-repo>/test/*_benchmark.cpp`; `beans show devilGo-w6qf --json | grep -A3 blocked` (empty).
7 fix: `beans update devilGo-w6qf --body-append "- [ ] TDD: ported workload counts (frames/path calls/crawl steps) asserted equal to the C++ benchmark inputs before budgets are enforced"` + replace body "epics 02/03" with "devilGo-5wsb / devilGo-qbiy".
8 validation: re-run `beans show devilGo-w6qf --body-only | grep TDD` present; dep line has devilGo ids.
9 related: devilGo-qbiy (renderer epic, <3ms AC), devilGo-uc44 (other CI gate).
VERDICT: NEEDS-WORK:no-explicit-TDD-first-AC;absolute-budgets-on-unpinned-CI-runner;stale-dg-epic-refs

Fix pass 2026-10-16: Dep-wiring applied (fix pass): declared edges wired in tracker.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).
