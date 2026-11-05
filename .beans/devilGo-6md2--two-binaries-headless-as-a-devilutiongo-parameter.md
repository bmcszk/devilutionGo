---
# devilGo-6md2
title: 'Two binaries: headless as a devilutionGo parameter'
status: completed
type: task
priority: normal
created_at: 2026-11-05T22:27:16Z
updated_at: 2026-11-05T23:21:47Z
parent: devilGo-8csd
---

# Two binaries: headless as a devilutionGo parameter

## Goal
One game binary. `devilutionGo` runs windowed and headless, chosen by a
command-line parameter; `cmd/devilutionGo-headless` disappears. Implements
decision 1 of docs/adrs/adr-020.md.

## Research (pre-validated)
- ADR-020 classifies `cmd/devilutionGo-headless` as the same game with another
  entry point (docs/adrs/adr-020.md:20-33).
- Both commands build the same game: `game.NewGame` plus a flag router
  (cmd/devilutionGo/main.go:14-16); headless adds `engine.MockTickSource`, a
  50ms tick loop and `--ticks` (cmd/devilutionGo-headless/main.go:35-48,:55-57).
- ADR-017 already routes run modes inside one program
  (docs/adrs/adr-017.md:10-12), so a mode parameter is the house pattern.
- Names to retarget: `justfile:9-10` (`build-headless`, `bin/devilutionGo-headless`),
  `docs/recipes/justfile-patterns.md:23` (`BIN_HEADLESS`), `DEVELOPMENT.md:11,:43`,
  `AGENTS.md:65,:102`. No CI workflow builds it (grep over `.github/`, `scripts/`,
  `justfile` finds only the three justfile lines).
- Test surface that moves with the code: `cmd/devilutionGo-headless/main_test.go`
  (tick loop, exit codes, app-fatal boundary).

## Scope
- Add the headless parameter to `cmd/devilutionGo` (flag or subcommand; pick one
  and keep `--ticks` semantics `cmd/devilutionGo-headless/main.go:57`).
- Move the tick loop and its tests into `cmd/devilutionGo`.
- Delete `cmd/devilutionGo-headless/`.
- Retarget `justfile` recipes and the docs names listed above.
- Keep exit codes: 0 success, 1 fatal, 64 usage
  (cmd/devilutionGo-headless/main.go:20-27).

## Acceptance criteria
- [x] `go build ./...` produces exactly one game command under `cmd/` (plus `devilutionGo-migrate`): `ls cmd/` = devilutionGo, devilutionGo-migrate
- [x] headless mode runs from `devilutionGo` with the same tick count and exit codes; `--ticks` still means 50ms ticks (`cmd/devilutionGo/headless.go:12`)
- [x] `cmd/devilutionGo-headless` deleted; no live reference outside ADR-019/020 and history files (gate grep = 0)
- [x] the headless tests pass from `cmd/devilutionGo` unchanged in intent (move table in Summary of Changes)
- [x] `just build` (or the retargeted recipe) builds the one binary; docs updated
- [x] gofmt, vet, golangci-lint, `go test ./... -count=1`, beans-lint all clean

## Triage
Triage evidence: no separate brainer session ran for this bean; this lane's
own disk-first read is the source of record and it landed as ADR-020, which
names the classification and the real call sites (docs/adrs/adr-020.md). Reading `cmd/devilutionGo/main.go`,
`cmd/devilutionGo-headless/main.go` and `justfile` confirms no shared code path
is lost: the two `main` packages differ only in tick source and flags. No C++
parity question is involved, so this needs no upstream probe.

## Libraries
- none (no new dependency; stdlib `flag` or the existing subcommand router)

## Skills for coder
- `/home/blaze/.agents/skills/coding-rules/SKILL.md` — read fully BEFORE acting
- `/home/blaze/.agents/skills/go-dev/SKILL.md` — read fully BEFORE acting
- `/home/blaze/.agents/skills/ponytail/SKILL.md` — read fully BEFORE acting
- `/home/blaze/.agents/skills/beans/SKILL.md` — read fully BEFORE acting

## Summary of Changes

`devilutionGo` is the one game binary and takes its run mode from the
command line: `--headless` runs the mock-tick loop (50 ms steps, `-ticks N`),
the default is the windowed path. `cmd/devilutionGo-headless` is deleted; its
logic lives in `cmd/devilutionGo/main.go` (dispatch + flags + exit codes) and
`cmd/devilutionGo/headless.go` (the tick loop), and its whole test file moved to
`cmd/devilutionGo/main_test.go`.

Test move, per case (no case deleted, no assertion relaxed):

| Old (cmd/devilutionGo-headless) | New (cmd/devilutionGo) | Fate |
|---|---|---|
| `TestExitOnError_Fatal_ExitsOneWithStderrMessage` | same name | unchanged; same `exitOnError` boundary, child exit 1 + `Error\ntest\n` |
| `TestHeadlessFatalHelper` | same name | unchanged |
| `TestRun_FlagError_ExitsSixtyFour` | same name | unchanged (`run(["-nope"])` -> 64) |
| `TestRun_ZeroTicks_ExitsZero` | same name | intent kept, arg made explicit: `run(["--headless"])` instead of `run(nil)`, because `nil` now selects the windowed mode |
| `TestReplay_DemoFile_HashesIdentically` | same name | unchanged |
| `TestHeadlessBinary_HasNoDisplayDependency` | `TestHeadlessMode_HasNoDisplayDependency` | same assertion (`go list -deps .` has no ebiten and no `/render`) |
| - | `TestRun_HelpShowsHeadlessSurface` | new: `--help` exits 0 and names `-headless` and `-ticks` |

`flag.ContinueOnError` returns `flag.ErrHelp` for `--help`; `run` maps that to
exit 0 and only a real parse error to 64 (diablo.cpp:1093).

Retargeted: `justfile` (`build-gui`/`build-headless` -> one `build`),
`docs/recipes/justfile-patterns.md` (`BIN_HEADLESS` gone, `run-headless` runs the
one binary with `--headless --ticks 100`), `DEVELOPMENT.md`, `README.md`,
`AGENTS.md`, `docs/research/dg-conversion-inventory.md`,
`cmd/devilutionGo-migrate/cli_test.go` (comment) and
`docs/pow/asset-verification-wave.md` (per-file table).

## Proof of Work

Lane branch `dg-binaries`, base `a9961d6`. Commits: 6372b4b (fold),
ae7cb2c (audit snapshot paths).

| Gate | Command | Observed |
|---|---|---|
| build | `go build ./...` | rc0 |
| one command | `ls -d cmd/*/` | `cmd/devilutionGo/  cmd/devilutionGo-migrate/` |
| format | `gofmt -l .` | empty |
| lint | `./scripts/lint.sh` | `0 issues.` (`simpurity: internal/game internal/diablo clean (226 sources)`) |
| vet | `go vet ./...` | rc0 |
| tests | `go test ./... -count=1` | 43 packages `ok`, 0 FAIL |
| moved cases | `go test ./cmd/devilutionGo/ -run 'TestRun|TestExitOnError|TestReplay|TestHeadlessMode' -v` | 6 PASS (names above) |
| help surface | `go run ./cmd/devilutionGo --help` | rc0; prints `-headless  run without a window, then exit` and `-ticks int  with -headless, run this many 50ms game ticks, then exit 0` |
| headless run | `go run ./cmd/devilutionGo --headless --ticks 3` | rc0; `INFO headless run complete ticks=3` |
| windowed run | `go run ./cmd/devilutionGo` | rc0; `INFO devilutionGo started debug=false` |
| recipe | `just build` | `bin/devilutionGo` 4.3 MB |
| reference gate | `grep -rn devilutionGo-headless --include='*.go' --include='*.md' --include='justfile' . \| grep -v -e .beans/ -e docs/adrs/ -e .hermes/ \| wc -l` | 0 |
| ledger | `python3 scripts/beans-lint.py` | OK |

Residual: `TestHeadlessMode_HasNoDisplayDependency` asserts over the whole
command's dependency graph, which is display-free today because the windowed
mode is still a stub; the verbatim assertion is kept deliberately, and the
test's doc comment records that the ebiten bring-up has to re-scope it to the
headless run path.
