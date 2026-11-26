---
# devilGo-n5an
title: 'framebudget: CI-class budget headroom nearly consumed by runner wobble (town row spiked past budget)'
status: completed
type: bug
priority: normal
created_at: 2026-11-26T00:36:31Z
updated_at: 2026-11-26T19:14:09Z
parent: devilGo-p1t8
---

# framebudget: the CI-class budget headroom is nearly consumed by runner wobble (town row spiked 27.195ms vs 26.672ms budget)

## Goal
Harden the CI-class frame-budget gate against single-sample runner stalls WITHOUT changing the budget value or factor: each CI row is measured in 2 independent child processes and the parent fails only when BOTH fail. The ranking intent (anything past ~26.7ms is a regression signal) is preserved; a whole-minute co-tenant stall can still produce a false red, but no fix shape can produce a false green for a real regression — that asymmetry is why hardening wins.

## Research
- Original evidence (2026-11-25): PR #86 first ci run measured town row p95 27.195ms vs ciBudget 26.672ms; the identical tree passed the same gate on the documented re-run. Calibration baseline 21.138ms p95 (run 44656997769).
- Fresh telemetry (2026-11-26, five consecutive master pushes, town-scene total p95 over 128 frames): 13.124 / 13.141 / 14.014 / 15.412 / 15.985ms. The spike day is not the class's centre — recalibrating the baseline from these runs would be a date-stamped magic number AND re-create the knife edge (14.3x1.6=22.9 fails the spike harder; ~1.9 clears it by 0.005ms).
- PATH CORRECTION (triage-verified): the gate lives in `internal/integration/`, NOT internal/engine/ — framebudget_bench_test.go (budgets, frameBudgetClass, p95 sites), framebudget_row_test.go (measureRowInOwnProcess, one-row-per-child runner), framebudget_stage_test.go, framebudget_town_test.go.
- Triage decision (task-n5an-triage-brainer, parent-accepted): HARDEN via min-of-2; rejected recalibration (knife-edge + throws away the runner-class-ratio meaning of the 1.6 factor) and p90/percentile changes (rejected above). No new constants, no dependencies, no workflow change (the gate runs in ci's go test; a second child ~2x the row's own time, inside the measured window).

## DECIDED (parent, 2026-11-26)
- This bean IS the fix unit (the triage's recommendation over the evidence-record/new-bean split — avoids a phantom bean). The wobble observation moves into the PoW as evidence.
- ciBudget value and the 1.6 factor: UNCHANGED (ponytail never-cut: the calibration ratio is the gate's meaning).
- Scope: internal/integration/framebudget_bench_test.go (budgetClass gains passes int; CI passes=2, local passes=1; ciBudget comment rewritten with the full derivation + history) + framebudget_row_test.go (run `passes` children, fail only if ALL fail, quote every failing pass; local path byte-identical). NO change to town/stage tests, p95/percentile, budgets, workflow.
- TDD shape (binding): the discriminating RED is a fail-then-pass child sequence under the NEW passes=2 rule — pre-fix code (fail-on-first) FAILS it, post-fix PASSES; plus the mutation flip back to fail-on-first going red. A pin the pre-fix code also passes certifies nothing.

## Acceptance criteria
- [x] `internal/integration/framebudget_bench_test.go`: ciBudget comment shows the full derivation and history in the file itself - the 16.67ms budget x 1.6 = 26.672ms, calibration run 44656997769 (21.138ms cathedral, 2026-11-20), spike run 45728266581 (27.195ms town, identical tree passed on re-run), the five same-day quiet runs (13.124 / 13.141 / 14.014 / 15.412 / 15.985ms, 2026-11-26) - and states that the factor is a runner-class ratio, not re-derived from the quiet runs; the `* 1.6` value is unchanged in the diff.
- [x] `internal/integration/framebudget_row_test.go`: with `CI=true` the town and cathedral rows are each measured in 2 child processes and the parent FAILS only when both passes fail, with every failing pass's p95 + stage table in the message; with `CI` unset exactly 1 child process runs and the strict 16.67ms budget is asserted (local behaviour byte-identical). Proven by a unit test whose fail-then-pass case FAILS on the pre-fix parent (RED output in PoW) and a mutation flip back to fail-on-first going red.
- [x] The first CI run after the change names its class in the pass log (`budget %.3fms, CI runner`), and a real regression is still caught: the PoW records ONE CI run of the changed gate on an untouched tree (green, both passes' p95s quoted) - operator-run, since the brainer and this triage cannot execute tests.

## Libraries
stdlib (os/exec already in use).

## Skills for coder
- ~/.agents/skills/go-tests/SKILL.md — red first; benchmarks reference at references/benchmarks.md
- ~/.agents/skills/coding-rules/SKILL.md — no bool-flag params (passes int, not retries bool); never weaken a test to green
- ~/.agents/skills/go-dev/SKILL.md — repo patterns; no nolint; %w wrapping
- ~/.agents/skills/ponytail/SKILL.md — smallest landable diff; its never-cut list includes the calibration ratio
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md — binding pitfall: a town-row p95 marginally past ciBudget is drift until proven otherwise; never edit ciBudget silently to green

## Triage
Filed 2026-11-25 from PR #86's CI failure + documented re-run. Deep triage DONE 2026-11-26 (task-n5an-triage-brainer; parent accepted the decision + the path correction). Verdict PROMOTE-TODO as the fix unit; deferred follow-up named in the triage (committed CI telemetry trail -> future draft bean). Full verdict: .hermes/lane/gate-parity/n5an-triage.md



## Proof of Work

- TDD: the discriminating walk verbatim in .hermes/lane/framebudget/wave-result.md - RED (fail-then-pass + all-missed-quoting + clean-exit-no-table all fail on the pre-fix fail-on-first rule; the local single-pass case passing is CORRECT: local behaviour unchanged), GREEN after the fix, mutation flip back to fail-on-first re-reds (pin not regression-blind), restored.
- Real-path sanity: CI=true -> town row measured in 2 child processes, BOTH passes' p95s logged (5.649ms / 5.154ms total), row held; CI unset -> exactly 1 child, strict 16.67ms class, cathedral row 4.963ms.
- Review (task-fb-review-brainer, gate-semantics strict round): **APPROVE** - budget arithmetic untouched, rule exactly as decided, TDD honest, scope jail held (2 files). Full verdict: .hermes/lane/framebudget/review-result.md
- Parent gate re-run at 95d6c24: gofmt empty; builds + vet both tags rc0; lint 0 issues; untagged suite 54 pkgs ok; windowed pair ok; beans-lint OK (238).
- AC3 evidence + honest deviation: the PR's own CI run 45881359741 ran the changed gate on the untouched tree - `ok github.com/bmcszk/devilutionGo/internal/integration 6.978s` inside `go test ./...`, run green (https://github.com/bmcszk/devilutionGo/actions/runs/45881359741). The pass log does NOT carry the per-pass p95 lines: pass-mode t.Logf output is hidden without -v, so the AC's "names its class in the pass log" clause is unobservable on a green default run. The quoted-both-passes evidence exists in the wave log's local CI=true run (2 passes, both quoted). Recorded as a parent-accepted deviation of the AC's letter; the intent (first CI run green on the untouched tree) is met. If the class line is wanted in CI pass logs, that is a -v/print follow-up, filed nowhere yet - noted here.
- Residual: 5 same-day CI runs remain a small sample (the fix is robust to it by design - min-of-2 re-derives nothing); the deferred CI-telemetry-trail follow-up bean is named in the triage and filed nowhere yet.
