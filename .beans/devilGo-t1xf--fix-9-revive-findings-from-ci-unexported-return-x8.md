---
# devilGo-t1xf
title: Fix 9 revive findings from CI (unexported-return x8, unhandled-error x1)
status: completed
type: task
priority: normal
created_at: 2026-10-27T12:11:11Z
updated_at: 2026-10-27T12:11:21Z
---

CI (stock golangci-lint v2.15.0 + repo rule set minus whenthen) reports 9 revive issues on epic/qna9-phase-0.

## Scope
1. cmd/gamedatagoldens/main.go:53 — `unhandled-error` on `fmt.Fprintln(os.Stderr, ...)`.
2-5. `unexported-return` on parsers: ParseMonstdat, ParseObjdat, ParseQuestdat, ParseTextdat.
6-9. `unexported-return` on SourceOf, BehaviorFor, NewXoshiro128pp, NewOverrideStack.

## Contract
- Rename the type to its exported form; keep methods/fields/semantics identical.
- No `//nolint`, no behaviour change.

## Acceptance criteria
- `go build ./...` OK, `go vet ./...` clean, tests pass.
- `/tmp/ci-bin/golangci-lint run --config /tmp/gs.yml ./...` prints `0 issues.`

## Out of scope
Anything beyond the 9 named findings; no refactors.

- [x] Rename monsterData -> MonsterData
- [x] Rename objectData -> ObjectData
- [x] Rename questRow -> QuestRow
- [x] Rename textRow -> TextRow
- [x] Rename behavior -> Behavior
- [x] Rename missileSource -> MissileSource
- [x] Rename xoshiro128pp -> Xoshiro128pp
- [x] Rename overrideStack -> OverrideStack
- [x] Handle stderr write error in gamedatagoldens
- [x] Gates: build, vet, tests, CI-equivalent lint 0 issues


## Proof of Work

Cold-state reproduction before the fix (9 issues):
```
$ /tmp/ci-bin/golangci-lint run --config /tmp/gs.yml ./... 2>&1 | tail -3
../internal/diablo/missiles/registry.go:78:33: unexported-return: exported func BehaviorFor returns unexported type missiles.behavior ...
../internal/engine/txtdata/override.go:18:53: unexported-return: exported func NewOverrideStack returns unexported type *txtdata.overrideStack ...
9 issues:
* revive: 9
```

After the fix:
```
$ go build ./... && echo BUILD_OK
BUILD_OK

$ go vet ./... 2>&1 | grep -v '^#' | head -5
(no output — rc 0)

$ go test ./internal/... ./cmd/...
(all packages ok; no failures; cmd/* have no test files)

$ /tmp/ci-bin/golangci-lint run --config /tmp/gs.yml ./... 2>&1 | tail -5
0 issues.

$ gofmt -l . | grep -v '^vendor/'
(empty)

$ git diff -U0 | grep -n nolint
none
```

## Changes
- 8 type renames to the exported form (declaration, receivers, literals, slice types, doc comments): `MonsterData`, `ObjectData`, `QuestRow`, `TextRow`, `Behavior`, `MissileSource`, `Xoshiro128pp`, `OverrideStack`.
- Deleted three now-redundant `type X = x` alias blocks (monstdat, objdat, textdat) that only existed to expose the row type; the renamed type serves both callers and parsers. Doc descriptions folded into the primary type comments.
- `cmd/gamedatagoldens`: replaced the ignored `fmt.Fprintln(os.Stderr, ...)` with a `warn(err error)` helper using `_, _ = fmt.Fprintln(...)`, so the write result is explicitly discarded at one place (no `//nolint`).
- Restored lowercase prose where the sed pass over-capitalised comments (`behavior registry entry`, package comment) and kept `missileSource` in the comment that cites the upstream C++ enum name; fixed the stale `newXoshiro128pp` doc line to `NewXoshiro128pp`.

Residual: none. Vendor `gofmt` hits (`vendor/github.com/suprsokr/go-mpq/*.go`) are pre-existing and untouched.
