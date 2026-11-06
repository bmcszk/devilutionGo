---
# devilGo-2h1y
title: One devkit refresher command (ADR-021)
status: completed
type: task
priority: normal
created_at: 2026-11-06T16:07:08Z
updated_at: 2026-11-06T16:08:16Z
parent: devilGo-8csd
---

# One devkit refresher command (ADR-021)

## Goal

One devkit refresher command for the golden fixtures: `tools/goldens`, with
`dun`, `dpiece`, `town`, `town-inputs` and `gamedata` subcommands. It replaces
`tools/probe/dunconv` and `tools/gamedatagoldens`. ADR-021 records the decision
and marks the rejected-fold paragraph in `docs/adrs/adr-019.md:88-95` as
superseded in part.

## Context

`tools/probe/dunconv` already dispatched on `args[0]` across four modes with a
usage error for anything else, so the dispatch layer ADR-019 counted as new
already existed. The second program, `tools/gamedatagoldens`, ran the same job
for the gamedata family: read an offline input, write a committed golden through
the `testsupport` encoder the tests decode with. One `main` instead of two, one
devkit table row instead of two.

## Scope

- New `tools/goldens/` flat `package main`: subcommand table, one dispatch path,
  one usage line, `main_test.go` pinning the table and the usage line.
- `dun`, `dpiece`, `town`, `town-inputs` moved without behaviour change;
  `gamedata` is the former run body with `probeCommit` and the parity schema
  comment intact.
- Delete `tools/probe/dunconv/` and `tools/gamedatagoldens/`.
- Retarget `tools/probe/README.md`, `internal/testdata/drlg/README.md`,
  `docs/research/dg-conversion-inventory.md`, `justfile`.
- Out of scope: `cmd/` (two-binary law), `internal/` game code,
  `internal/testdata` fixture bytes, `scripts/simpurity`, `lint.sh`, the C++
  probe files, ADR bodies other than the one supersede sentence.

## Acceptance criteria

- [x] `go run ./tools/goldens` prints one usage line naming all five subcommands
- [x] `tools/probe/dunconv/` and `tools/gamedatagoldens/` are gone; `cmd/` still holds exactly `devilutionGo/` and `devilutionGo-migrate/`
- [x] `go run ./tools/goldens gamedata` runs from a clean `HEAD` export with no import errors and writes all seven tables, and the real tree's fixtures stay byte-identical
- [x] zero live references to `dunconv` or `gamedatagoldens` outside `.beans`, the ADR bodies that record history, and ADR-021's own cites
- [x] `just refresh-gamedata-goldens` calls the new path
- [x] gates: `go build ./...`, `gofmt -l .`, `./scripts/lint.sh`, `go vet ./internal/... ./tools/...`, `go test ./... -count=1`, `python3 scripts/beans-lint.py`, `just sqlc-check` all green

## Libraries

- none (stdlib only; reuses `internal/diablo/testsupport` encoders)

## Skills for coder

- `/home/blaze/.agents/skills/coding-rules/SKILL.md` — read fully BEFORE acting
- `/home/blaze/.agents/skills/go-dev/SKILL.md` — read fully BEFORE acting
- `/home/blaze/.agents/skills/go-tests/SKILL.md` — read fully BEFORE acting
- `/home/blaze/.agents/skills/ponytail/SKILL.md` — read fully BEFORE acting
- `/home/blaze/.agents/skills/beans/SKILL.md` — read fully BEFORE acting

## Triage

Delegated planner (pi-as-brainer lane contract for `dg-lane-goldens` s1) after
census of the tree at `8876df5`: `tools/probe/dunconv/main.go:39-46` is already a
four-mode dispatcher, `tools/gamedatagoldens/main.go` is a second `package main`
with no importers, and the ADR-019 rejection paragraph rests on the claim that
the two programs "share no code today". Re-evaluated: the dispatch layer is not
new, the fold deletes one `main` and one devkit table row, and both programs
already write through `testsupport`. Risk: stale recipe or doc paths, which the
acceptance criteria check by running the recipe and grepping the tree.


## Progress log

- 2026-11-06 lane `dg-lane-goldens` s1, branch `dg-goldens`, base `8876df5`.
- 2161327: ADR-021 + `tools/goldens` fold, dunconv/gamedatagoldens deleted, docs
  and justfile retargeted.
- f4dc01a: one-shot `scripts/convert_lighttables.py` deleted.

## Summary of Changes

`tools/goldens` is the one devkit refresher command. ADR-021 records the
decision; ADR-019 keeps its text and gains one sentence marking
`docs/adrs/adr-019.md:88-95` superseded in part.

- `tools/goldens/main.go`: package doc, subcommand table
  (`dun`, `dpiece`, `town`, `town-inputs`, `gamedata`), one dispatch path, one
  usage line listing all five. The four dungeon forms are the `dunconv` dispatch
  moved without a behaviour change; the bare `<artifact-dir> <corpus-dir>` form
  became `dun <artifact-dir> <corpus-dir>`.
- `tools/goldens/main_test.go`: table test for name plus argument-count
  resolution, rejection test for every wrong shape, and the one-line usage
  assertion. RED first: with the `gamedata` row removed the suite fails
  (`lookupRefresher(["gamedata"]) ok = false, want true` and
  `usage "usage: goldens <dun|dpiece|town|town-inputs>" does not name
  subcommand "gamedata"`), then green.
- `tools/goldens/{dpiece,town,inputs}.go`: moved from `tools/probe/dunconv`,
  signatures take the positional path slice, log lines say `goldens ...`.
- `tools/goldens/gamedata.go`: the former `tools/gamedatagoldens/main.go` run
  body, `probeCommit` and the parity schema comment intact, `main`/`warn`
  dropped because the dispatcher reports errors.
- `tools/probe/dunconv/` and `tools/gamedatagoldens/` deleted.
- Retargeted: `tools/probe/README.md` (devkit table now one row, `### goldens`
  and `### gamedata` steps), `internal/testdata/drlg/README.md` step 4/5 and the
  town-inputs command, `docs/research/dg-conversion-inventory.md` two inventory
  rows plus two consumer rows, `justfile` `refresh-gamedata-goldens`.
- Rider: `scripts/convert_lighttables.py` deleted. Its six `.tsv` goldens are
  committed and no `.hex` input is left, so the one-shot migration has no
  remaining input. The only reference was a past-tense Proof of Work line in the
  completed "Convert lighttable fixtures .hex to human-readable TSV" bean, which
  records history rather than telling anyone to run it; nothing else in the repo
  names the file.

## Reference sweep

| Token | At base `8876df5` | At HEAD |
|---|---|---|
| `dunconv` | 8 files / 30 occurrences | 3 files / 7 occurrences |
| `gamedatagoldens` | 7 files / 22 occurrences | 4 files / 13 occurrences |

All remaining hits are records of the old state plus ADR-021's own cites:
`docs/adrs/adr-019.md` 5, `docs/adrs/adr-020.md` 7, `docs/adrs/adr-021.md` 4,
`docs/pow/asset-verification-wave.md` 3. Zero hits in `tools/`, `internal/`,
`justfile`, `scripts/` or any live doc; `git grep` outside `.beans`,
`docs/adrs` and `docs/pow` is empty. `.beans` holds 56 occurrences across 12
bean bodies, exempt per the lane contract.

## Decision notes

- ADR-020 is not edited: the two-binary law stands as written and
  `tools/goldens` is test instrumentation under the ADR-019 carve-out, not a
  third user-facing binary. `cmd/` is untouched.
- The devkit table in `tools/probe/README.md` is now one row for one command
  instead of two rows for two programs; the gate column keeps both consumers
  (`internal/diablo/levels` and `internal/diablo/gamedata/audit`).
- `scripts/simpurity` and `lint.sh` stay as they are: the purity gate guards
  wall-clock and file io leaking into game code, a different class from the
  fixture refreshers.

## Findings (pre-existing, out of scope)

`tools/goldens gamedata` is not idempotent against the committed goldens, the
same drift the "Refresh gamedata goldens is not idempotent against committed
output" bean already tracks: the generator writes full tables with
`probeCommit: refresh-recipe-v1`, while the committed JSON carries
`unit-fixture-slice-a` row slices. Verified in a `git archive HEAD` export under
`/tmp`, so the real tree's fixtures stayed byte-identical (`git status` clean
apart from the staged commits).

`docs/fixtures.md:61,:101` still describe the lighttable fixtures as `xxd` hex
dumps; that staleness predates this lane (the `.tsv` migration landed earlier)
and the file does not name either refresher.

## Proof of Work

Branch `dg-goldens`, base `8876df5`. Commits: 2161327 (fold + ADR-021),
f4dc01a (one-shot script deleted).

| Gate | Command | Observed |
|---|---|---|
| build | `go build ./...` | `build rc0` |
| format | `gofmt -l .` | 0 files |
| lint | `./scripts/lint.sh` | `simpurity: internal/game internal/diablo clean (229 sources)` + `0 issues.` |
| vet | `go vet ./internal/... ./tools/...` | `vet rc0` |
| tests | `go test ./... -count=1` | 44 `ok` lines, 0 `FAIL` (`tools/goldens` is the 44th) |
| beans ledger | `python3 scripts/beans-lint.py` | `beans-lint: OK (177 beans)` |
| sqlc | `just sqlc-check` | `sqlc generate` + `no raw SQL outside sqlc-generated code`, rc0 |
| refresh path resolves | `git archive HEAD \| tar -x -C /tmp/dg-goldens-gate && go run ./tools/goldens gamedata` | 7 `fixture written` lines (spelldat 36 rows/15 cols, misdat 68/8, monstdat 112/41, objdat 109/12, textdat 338/4, questdat 24/9, towners 12/11), rc0 |
| usage line | `go run ./tools/goldens` | `goldens failed error="usage: goldens <dun|dpiece|town|town-inputs|gamedata>"`, exit 1 |
| dungeon wiring | `go run ./tools/goldens dun /tmp/dg-empty /tmp/dg-corpus` | `no .dun artifacts in /tmp/dg-empty`, exit 1 (fails closed, writes nothing) |
| goldens untouched | `git status --porcelain internal/testdata` | empty |
| cmd set | `ls -d cmd/*/` | `cmd/devilutionGo/  cmd/devilutionGo-migrate/` |
