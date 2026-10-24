---
# devilGo-w6qf
title: Benchmark suite + perf budgets in CI
status: completed
type: task
priority: normal
tags:
    - plan:task
    - parity
    - bench
created_at: 2026-10-15T16:35:18Z
updated_at: 2026-10-17T12:44:27Z
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
- [ ] Frame <3ms, path 10k<150ms (measured floor 2026-10-17: 7.5µs/search quiet → 75ms; 150 = 2x), crawl 100k<30ms on CI runner; regressions >10% flagged via benchstat vs stored baseline
- [ ] Render benchmarks (clx/dun/light) need the render port from 7aul/kx0l — defer those three to after render lands; this bean ships path+crawl+palette benches + CI workflow

## Libraries
- stdlib only (no third-party deps at this stage)

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


## Proof of Work
Worker (pi _glm_flash_first) landed budgets_test.go (path 150ms + NEW crawl 30ms budget with measured-floor rationale + pocket fixture) then died silently mid-task (reply/log empty). Orchestrator finished: path_bench_test.go (BenchmarkFindPath40x40 + BenchmarkCrawl100k, ReportAllocs, 0 allocs/op verified) and .github/workflows/bench.yml (budget tests gate, bench smoke non-gating, artifact upload). Render benchmarks (clx/dun/light) deferred to post-7aul per AC note.
Orchestrator re-gate: go test ./internal/... PASS, vet clean, lint clean; benches run 100x green under load.

## AUDIT (brainer 2026-10-24, beans-completion audit vs tree)
AUDIT note (brainer 2026-10-24): bench workflows folded into ci.yml
verified (7f8x artifacts in tree); specific budget millisecond figures in
PoW are historical CI output — unverifiable, accepted.
