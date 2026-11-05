---
# devilGo-6md2
title: 'Two binaries: headless as a devilutionGo parameter'
status: todo
type: task
priority: normal
created_at: 2026-11-05T22:27:16Z
updated_at: 2026-11-05T22:27:37Z
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
- [ ] `go build ./...` produces exactly one game command under `cmd/` (plus `devilutionGo-migrate`)
- [ ] headless mode runs from `devilutionGo` with the same tick count and exit codes; `--ticks` still means 50ms ticks
- [ ] `cmd/devilutionGo-headless` deleted; no live reference outside ADR-019/020 and history files
- [ ] the headless tests pass from `cmd/devilutionGo` unchanged in intent
- [ ] `just build` (or the retargeted recipe) builds the one binary; docs updated
- [ ] gofmt, vet, golangci-lint, `go test ./... -count=1`, beans-lint all clean

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
