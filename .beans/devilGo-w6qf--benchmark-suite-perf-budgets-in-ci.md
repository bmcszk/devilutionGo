---
# devilGo-w6qf
title: Benchmark suite + perf budgets in CI
status: draft
type: task
priority: normal
tags:
    - plan:task
    - parity
    - bench
created_at: 2026-10-15T16:35:18Z
updated_at: 2026-10-15T16:45:26Z
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

## Dependencies & blockers
Depends on epics 02/03 progressively.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.
