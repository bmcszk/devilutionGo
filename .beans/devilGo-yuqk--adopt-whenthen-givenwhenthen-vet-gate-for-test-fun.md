---
# devilGo-yuqk
title: Adopt whenthen given/when/then vet gate for test functions
status: draft
type: task
priority: normal
created_at: 2026-10-26T18:03:13Z
updated_at: 2026-10-26T18:03:40Z
parent: devilGo-1n0a
---

## Goal

Make `go vet -vettool=whenthen ./...` green repo-wide, then decide whether the
gate belongs in CI. Until then the check stays opt-in so it never blocks pushes
or the branch's lint landing.

## Current state (measured 2026-10-26 on epic/qna9-phase-0)

- `just lint-tests` (= `go vet -vettool=whenthen ./...`) exits 1 with
  **79 findings across 29 test files** (tallies from
  `grep -cE 'missing .*phase\(s\)'` / `grep -oE '^[^:]+_test\.go' | sort -u`).
  Findings are all of one class: `Test<Name>: missing when, then phase(s)` —
  the tool wants an explicit given/when/then phase structure in each test
  function.
- Heaviest files: `internal/diablo/missiles/displacement_test.go` (8),
  `internal/diablo/missiles/collision_gates_test.go` (7),
  `internal/diablo/animation/animationinfo_test.go` (7),
  `internal/diablo/missiles/process_test.go` (6),
  `internal/light/tables_test.go` (5).
- The whenthen step reached `just lint` by accident (worker over-scope, commit
  `6229efa`); it has never been green on this branch — e.g.
  `internal/diablo/entities/size_test.go` (untouched since `23edc2f`) fails.
  It therefore blocked every push. Reverted to opt-in in commit
  `fix(tooling): keep the whenthen DSL check opt-in`.

## Scope of the migration

Rewrite the 79 flagged test functions into the explicit given/when/then DSL
(fluent `parts` struct + `.and()` chaining per the go-fluent-testing skill).

Hard constraint: **no assertion changes, no test weakening, no coverage loss** —
the rewrite is structural only. Ported-1:1 tests keep their golden/tape inputs.

## Acceptance criteria

- [ ] `just lint-tests` exits 0 with zero findings.
- [ ] `go test -count=1 ./...` unchanged: same tests, same assertions, all ok.
- [ ] No `//nolint`, no excluded files, no per-directory opt-outs.
- [ ] Diff reviewed for "assertion unchanged" proof (before/after per file).
- [ ] Decide + record whether `lint-tests` joins `just ci` (then update the
      epic's CI notes) or stays opt-in.
- [ ] a Proof-of-Work section with the RED `just lint-tests` run, the GREEN run and
      `go test ./...` tail before the bean closes.

## Notes

- Tool install: `(cd ../whenthen && go install ./cmd/whenthen)`; `just
  lint-tests` resolves it via `command -v whenthen` or `$(go env GOPATH)/bin`.
- Mechanical chunking: 29 files → one subtask per file or small file group
  (~10 min each), subagents can work in parallel on disjoint files.
