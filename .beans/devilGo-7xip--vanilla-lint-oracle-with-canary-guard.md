---
# devilGo-7xip
title: Vanilla lint oracle with canary guard
status: completed
type: task
priority: normal
created_at: 2026-11-02T13:01:01Z
updated_at: 2026-11-02T13:19:27Z
parent: devilGo-1n0a
---

## Goal

Make the lint gate fail-closed and CI-faithful: vanilla golangci-lint at the
pin as the only code oracle (no plugin build), whenthen as a standalone
`go vet -vettool` step, and a seeded-violation canary that refuses 0-issues
verdicts from a blind oracle.

## Background

The custom-plugin golangci build (whenthen linked in) silently dropped revive
`unhandled-error` findings in multi-file packages that the vanilla binary at
the SAME pin reports (same embedded revive v1.17.1; the linked plugin perturbs
the goanalysis driver). Hooks ran the blind build and said 0 issues while CI
flagged `font.go:156`. The fix (commit 84e6773) was pushed to the PR branch
after PR #12 had already squash-merged, so it never landed on master — re-land
under this bean.

## Libraries

Stdtooling only: bash + pinned go toolchain binaries (golangci-lint v2.15.0,
whenthen v0.1.2 as vettool). No new Go libraries.

## Triage

brainer pass owed on the re-land PR (compact read-only brainer, pre-merge).

## Skills for coder

~/.agents/skills/coding-rules/SKILL.md
~/.agents/skills/go-dev/SKILL.md
~/.agents/skills/ponytail/SKILL.md

## Acceptance criteria

- [x] `scripts/lint.sh` runs vanilla golangci-lint@v2.15.0 with no plugin
      linking; `.golangci.yml` carries no whenthen registration;
      `.custom-gcl.yml` removed from the tree.
- [x] whenthen enforced via `go vet -vettool` inside `scripts/lint.sh`; a
      probe test without given/when/then phases fails the gate.
- [x] Canary guard: a seeded `hash.Write` violation fails lint; a blind
      build reports refusal instead of 0 issues.
- [x] Negative proof: restored pre-fix `font.go` (bare `hash.Write`) is
      caught by the new gate locally (was a silent pass under the plugin
      build).
- [x] Gates: go build / go vet / go test (37 pkgs) / gofmt clean; CI green
      on the re-land PR (run 41451597880 at 53fe3ec).

## Progress log

2026-11-02 re-land of 84e6773 (pushed after PR #12 squash-merge) onto branch dg-vanilla-lint-oracle-with-canary-guard off master 4ccee3b; PR #13 opened. Ci was red once: beans-lint required ponytail in Skills for coder (fixed). Brainer verdict FIX: F2 header-oracle-faithful + F3 set -euo + F4 canary diagnostics + F5 root guard applied in commit on this branch; F1 CI AC deliberately unticked until green; F6 closeout (ship + PoW) lands after CI green.

## Proof of Work

- A: `bash scripts/lint.sh ./... --timeout=5m` from repo root and from `internal/assets` (cwd guard) → `0 issues.` both.
- B: `hash.Write(glyph.Pixels)` reverted to bare call → `font.go:158:3: unhandled-error` FLAGGED; restore → green again.
- E: second run reuses pins (`cmp pins.want pins.txt`), no reinstall line.
- Canary + whenthen vet + repo lint all exercised on every run above.

2026-11-02 CI green at 53fe3ec (run 41451597880): shippable; F1 tick fulfilled for real.
