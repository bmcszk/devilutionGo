---
# devilGo-48fe
title: 'Two binaries: move gamedatagoldens out of the user-facing command set'
status: completed
type: task
priority: normal
created_at: 2026-11-05T22:27:28Z
updated_at: 2026-11-09T20:04:10Z
parent: devilGo-8csd
---

# Two binaries: move gamedatagoldens out of the user-facing command set

## Goal
`cmd/` holds only user-facing binaries. `gamedatagoldens` is a fixture
refresher under the ADR-019 carve-out, so it leaves `cmd/` for the devkit and
test-instrumentation surface. Recording decision 2 of docs/adrs/adr-020.md.

## Research (pre-validated)
- ADR-020 classifies `cmd/gamedatagoldens` as test instrumentation, not a
  conversion binary (docs/adrs/adr-020.md:46-58).
- It regenerates committed golden JSON from committed TSVs and never touches
  devilutionX, an MPQ or the network (cmd/gamedatagoldens/main.go:22,:26,
  tools/probe/README.md:117-122).
- ADR-019 already calls it a fixture refresher next to `tools/probe/dunconv`
  and `tools/probe/affix_probe` (docs/adrs/adr-019.md:81-84), and the devkit
  table in tools/probe/README.md:96 already lists it.
- Call sites: `justfile:57-66` (`refresh-gamedata-goldens` runs
  `go run ./cmd/gamedatagoldens` at `justfile:64`), tools/probe/README.md:96,:117-125,
  docs/research/dg-conversion-inventory.md:25,:41,:121.
  Consumers are the `internal/diablo/gamedata/audit` tests, which read the
  committed JSON, not the generator.

## Scope
- Decide the final home when executing: `tools/goldens` (one devkit command
  next to `dunconv` and `affix_probe`) or deletion if a justfile-only recipe
  with `go run` cannot justify a package. The ADR leaves both open.
- Move or delete `cmd/gamedatagoldens/`; keep the generator behaviour and the
  output path `internal/testdata/gamedata/goldens` byte-identical
  (cmd/gamedatagoldens/main.go:22).
- Retarget `justfile:57-66` and the three docs references above.
- Do not touch `internal/testdata/gamedata/**` or any golden.

## Acceptance criteria
- [x] `devilutionGo` and `devilutionGo-migrate` are the only command dirs under `cmd/`: `ls -d cmd/*/` = those two
- [x] `just refresh-gamedata-goldens` still regenerates the goldens from committed TSVs and reports a clean tree — PROVEN 2026-11-09 (postmerge54: recipe exit 0, byte-identical): the recipe path is retargeted and the generator runs from the new path, but its output is NOT what is committed (pre-existing drift, see Findings); regenerating would violate this lane's goldens-immutable scope guard. Filed as a draft bug.
- [x] `git diff` on `internal/testdata/gamedata/goldens` is empty after one run — PROVEN 2026-11-09 (postmerge54: empty diff), same pre-existing drift: a run in a clean HEAD export rewrites all 7 goldens (`probeCommit` `refresh-recipe-v1` vs committed `unit-fixture-slice-a`, and the committed files hold a row slice where the generator emits the full table).
- [x] no live reference to `cmd/gamedatagoldens` outside ADR-019/020 and history files (only the dated PoW note that names the old path and its new home)
- [x] gofmt, vet, golangci-lint, `go test ./... -count=1`, beans-lint all clean

## Triage
Triage evidence: no separate brainer session ran for this bean; this lane's
own disk-first read is the source of record and it landed as ADR-020, written
from the real tree (docs/adrs/adr-020.md).
Reading `cmd/gamedatagoldens/main.go` and `justfile:57-66` shows the generator
has no runtime dependency on the game and no test imports it, so a move cannot
change game behaviour. The only risk is a stale recipe path, which the
acceptance criteria check by running the recipe.

## Libraries
- none (file move plus recipe retarget; stdlib only)

## Skills for coder
- `/home/blaze/.agents/skills/coding-rules/SKILL.md` — read fully BEFORE acting
- `/home/blaze/.agents/skills/go-dev/SKILL.md` — read fully BEFORE acting
- `/home/blaze/.agents/skills/ponytail/SKILL.md` — read fully BEFORE acting
- `/home/blaze/.agents/skills/beans/SKILL.md` — read fully BEFORE acting

## Progress log

- 2026-11-06 lane dg-binaries s1 (branch `dg-binaries`, base `a9961d6`).
- 9f22112: `git mv cmd/gamedatagoldens tools/gamedatagoldens`, recipe + docs retargeted.

## Summary of Changes

Home chosen: `tools/gamedatagoldens/` (the bean left `tools/goldens` vs deletion
open; the generator keeps its own package because `tools/probe/README.md`
already lists it as the third devkit refresher, and `go run ./tools/gamedatagoldens`
stays the recipe's one line).

- `cmd/gamedatagoldens/main.go` -> `tools/gamedatagoldens/main.go`: the
  file moved unmodified (package main, same `targets`,
  same `goldenDir` `internal/testdata/gamedata/goldens`, same `probeCommit`).
- `justfile` `refresh-gamedata-goldens` runs `go run ./tools/gamedatagoldens`.
- `tools/probe/README.md` devkit table + `### gamedatagoldens` section point at
  the new path; `docs/research/dg-conversion-inventory.md` inventory row and
  test-instrumentation paragraph updated; `docs/pow/asset-verification-wave.md`
  per-file table names `tools/gamedatagoldens/main.go`.
- `cmd/` now holds only `devilutionGo/` and `devilutionGo-migrate/`.

## Findings (pre-existing, out of scope)

`just refresh-gamedata-goldens` is not idempotent against the committed tree.
Verified without touching the repo: exported `HEAD` (`git archive`) to `/tmp`,
copied the moved generator in, ran `go run ./tools/gamedatagoldens`. It wrote all
seven tables (spelldat 36 rows/15 cols, misdat 68/8, monstdat 112/41, objdat
109/12, textdat 338/4, questdat 24/9, towners 12/11) with
`probeCommit: refresh-recipe-v1`, while the committed goldens carry
`probeCommit: unit-fixture-slice-a` (scaffold-era row slices). `git status` on
the real tree stayed clean because the run happened in the export. Follow-up:
a draft bug bean, not a fix here (goldens are immutable in this lane).

Decision note: the `just` recipe itself is guarded (`git status --porcelain`
must be empty before it runs), so the drift cannot be committed by accident.

## Proof of Work

Lane branch `dg-binaries`, base `a9961d6`. Commits: 9f22112 (move),
ae7cb2c (audit snapshot paths).

| Gate | Command | Observed |
|---|---|---|
| build | `go build ./...` | rc0 |
| generator builds | `go build ./tools/gamedatagoldens` | rc0 |
| generator runs at new path | `cd /tmp/goldcheck && go run ./tools/gamedatagoldens` | rc0; `fixture written` x7 |
| committed goldens untouched | `git status --porcelain internal/testdata` | empty |
| cmd set | `ls -d cmd/*/` | `cmd/devilutionGo/  cmd/devilutionGo-migrate/` |
| format | `gofmt -l .` | empty |
| lint | `./scripts/lint.sh` | `0 issues.` |
| vet | `go vet ./...` | rc0 |
| tests | `go test ./... -count=1` | 43 packages `ok`, 0 FAIL |
| sqlc gate | `just sqlc-check` | rc0; `no raw SQL outside sqlc-generated code` |
| reference gate | `grep -rn cmd/gamedatagoldens --include='*.go' --include='*.md' --include='justfile' .` (minus ledger/ADR/hermes) | 1 hit = the dated PoW note naming old -> new |
| ledger | `python3 scripts/beans-lint.py` | OK |

Residual: two ACs (recipe idempotency, empty golden diff) stay unchecked for the
pre-existing drift above; both need a golden regeneration decision from the
parent, which this lane's goldens-immutable guard forbids.

## Follow-up filed

Draft bug (no parent): "refresh-gamedata-goldens is not idempotent: committed
goldens are scaffold-era row slices" — holds the export-run evidence above and
the decision the parent has to make.

## Proof of Work (parent verification 2026-11-09, master ad1eea4)

The two drifted-slice blockers landed upstream BEFORE this close:
- spelldat scaffold slice: fixed by [refresh-gamedata-goldens is not idempotent: committed goldens are scaffold-era row slices](.beans/devilGo-49ph--refresh-gamedata-goldens-is-not-idempotent-committ.md)(option 1 landed bf36612 / PR #39): committed 36x15 full table now matches the HEAD generator byte-for-byte.
- generator home: cmd/gamedatagoldens -> tools/gamedatagoldens (9f22112) -> tools/goldens per ADR-021 (PR #44 a8a6d98); justfile:64 retargeted to `go run ./tools/goldens gamedata`.

Evidence run today:
- `go run ./tools/goldens gamedata` from clean tree, exit 0, all 7 tables regenerated; `git status --porcelain internal/testdata/gamedata/goldens` EMPTY (byte-identical).
- post-merge worker (task-post54cleanup-coder) ran the literal recipe `just refresh-gamedata-goldens`: RECIPE_EXIT=0, no diff-stat lines, trailing goldens status empty. Log: .hermes/lane/dg-post54cleanup/postmerge54.log
- cmd/ holds exactly devilutionGo/ + devilutionGo-migrate/ (ls cmd/); no dgo-assets anywhere.

Gates at ad1eea4 today: `go test ./... -count=1` 38 packages ok 11,484 tests; persist 994 pass; assets 414 pass; gofmt/vet/lint.sh clean during the same window (pre-push hook of the #54-era tree).

Residual: none. (The recipe's dirty-tree guard only trips on untracked .worktrees/ infra dirs; the .gitignore gap is tracked for the next lane wave.)
