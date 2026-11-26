---
# devilGo-n5an
title: 'framebudget: CI-class budget headroom nearly consumed by runner wobble (town row spiked past budget)'
status: draft
type: bug
priority: normal
created_at: 2026-11-26T00:36:31Z
updated_at: 2026-11-26T00:36:31Z
parent: devilGo-p1t8
---

# framebudget: the CI-class budget headroom is nearly consumed by runner wobble (town row spiked 27.195ms vs 26.672ms budget)

## Goal
The CI-class AC-3 budget (ciBudget = 16.67ms x 1.6 = 26.672ms) was calibrated against a 21.138ms p95 baseline (run 44656997769; comment in framebudget_bench_test.go:34-36 says "revisit on CI hardware changes"). On PR #86's first CI run the town row measured 27.195ms p95 (world=22.777ms) — 2% past the budget; the documented re-run on the identical commit passed the same gate. The calibration baseline is ~28% stale at p95 spikes: the gate now sits so close to the wobble envelope that a shared-runner spike flips it without any code regression.

## Research
- Failure: run 45728266581 (PR #86 first ci job) — town row p95 27.195ms over 128 frames, budget 26.672ms, world=22.777ms/panel=0.008/text=0.051/present=5.169.
- Non-regression evidence: the identical tree (f6155fe vs 1fa553e differ only in .beans) passes the STRICTER local mid-hardware budget (16.67ms) in the untagged full suite; the hl4 diff's untagged hot-path additions are nil-guarded accessors and a gate-rule registration (nanosecond-class); documented single re-run passed in 2m45s.
- The gate's own calibration note anticipates this class: "revisit on CI hardware changes".

## Acceptance criteria
- [ ] Recalibrate or harden: either re-derive the CI-class factor from a fresh multi-run p95 distribution (and update the comment block with the new baseline + date), or make the row noise-hardened (e.g. best-of-2 child processes, or p90), with the calibration history preserved in the comment
- [ ] The chosen fix lands via its own triaged flow; the wobble observation (this bean) stays as the evidence record

## Libraries
stdlib.

## Skills for coder
- ~/.agents/skills/go-tests/SKILL.md
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/ponytail/SKILL.md — smallest landable diff
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md

## Triage
Filed 2026-11-25 from PR #86's CI failure + documented re-run (parent session; evidence above). Draft; deep triage owed (decide recalibrate-vs-harden against the gate's ranking intent: "factor 1.6 ranks a real regression (anything past 26.7ms)").
