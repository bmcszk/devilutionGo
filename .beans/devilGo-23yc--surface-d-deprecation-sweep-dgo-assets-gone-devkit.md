---
# devilGo-23yc
title: 'Surface D: deprecation sweep - dgo-assets gone + devkit README'
status: completed
type: task
priority: high
tags:
    - plan:task
created_at: 2026-11-05T21:40:07Z
updated_at: 2026-11-05T22:21:45Z
parent: devilGo-8csd
blocked_by:
    - devilGo-j3ai
    - devilGo-bz7q
---

# Surface D: deprecation sweep — dgo-assets gone + one devkit README

## Goal
Per ADR-019: the `dgo-assets` command name goes away the same way `devilutionGo-import` did (adr-003.md:4 pattern; nothing calls it, no deprecation window). One devkit README section documents the golden refresher tooling so "what is dunconv?" questions end here.

## Research (pre-validated)
- Parent review APPROVE (/home/blaze/github/bmcszk/devilutionGo/.hermes/lane/dg-conv/dg-conv.review.md, task-dgconv-review-brainer).
- docs/research/dg-conversion-inventory.md: classification — reference instrumentation (tools/probe dvg_probe.cpp + dump-only patches, SHA-pinned) and golden refreshers (dunconv, gamedatagoldens) are NOT conversion programs; they need one documented home, not more programs.

## Scope
- Delete cmd/dgo-assets (folded surfaces B/C already carry the behavior).
- tools/probe/README.md: update the post-dgo-assets reality; add ONE "Golden refresher tooling (devkit)" section documenting dunconv + gamedatagoldens (and affix_probe if related) with their inputs/outputs — test-infra classification per ADR-019.
- Sweep live references to dgo-assets in cmd/, docs/, tools/ (beans/ADR history exempt).
- Update the ADR-013/ADR-018 READMEs if they name dgo-assets as the tool surface.

## Acceptance criteria
- [x] cmd/dgo-assets deleted (3 files); `go build ./...` green
- [x] `grep -rn dgo-assets cmd/ docs/ tools/`: 0 in cmd/ and tools/; docs hits limited to ADR-019 and the snapshot-flagged research ledger (see IMPL NOTES)
- [x] One devkit section in tools/probe/README.md covers dunconv + gamedatagoldens + affix_probe with invocation examples and their real gates
- [x] Full gates green: build, gofmt, lint.sh, vet, go test ./... -count=1 (43 ok / 0 FAIL after the deletion), beans-lint, just sqlc-check
- [ ] (NOT DONE, parent-owned) Retire the dg-pass3 + dg-conv worktrees/branches: outside this lane's scope guard, which forbids deleting worktrees outside this repo's own lane dirs

## Libraries

No new library bindings. Everything stays within the pinned catalog (ADR-009); internal/assets converters already committed.

## Skills for coder
- /home/blaze/.agents/skills/ponytail/SKILL.md
- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/beans/SKILL.md

## Dependencies & blockers
blockedBy: Surface B, Surface C. Parent: epic devilGo-8csd.

## Triage

Pre-validated by the ADR-019 review (APPROVE): brainer session `task-dgconv-review-brainer` (/home/blaze/github/bmcszk/devilutionGo/.hermes/lane/dg-conv/dg-conv.review.md), evidence in `docs/research/dg-conversion-inventory.md` (file:line cites) and docs/adrs/adr-019.md (merged @ 1d3bbbc). No open questions.

## IMPL NOTES (dg-surface s4, commit 43651a4)

### Deletion and identity retarget

- `cmd/dgo-assets` deleted (3 files: `main.go`, `internal/cli/cli.go`,
  `internal/cli/cli_test.go`). Behavior parity was already proven on Surface C:
  every one of its 7 test cases has a migrate counterpart, plus migrate extras
  (`TestAssetsVerifyAcceptsTheRootAsAPositionalArgument`,
  `TestAssetsWithoutASubcommandRunsTheExtractionPass`,
  `TestAssetsHelpNamesEveryNestedSubcommand`). Nothing else imported it: `go
  build ./...` and `go test ./...` are green after the deletion.
- `internal/assets/manifest/manifest.go`: `Tool` const now records the folded
  identity, `"devilutionGo-migrate"` (was the dead writer name `"dgo-assets"`;
  no production code read it, only the migrate fixtures that build an
  inventory). The package doc's family-enumeration pointer now names
  `cmd/devilutionGo-migrate`, the only command importing `internal/diablo`.
- `internal/assets/manifest/manifest_test.go`: two fixtures that hardcoded the
  literal `"dgo-assets"` now use `manifest.Tool`, so the writer name has one
  home.
- `cmd/devilutionGo-migrate/assets_inventory_list_test.go`: the comment on
  `TestAssetsWithoutASubcommandRunsTheExtractionPass` no longer names the
  deleted command; the behavior it pins is unchanged.

### Devkit section

`tools/probe/README.md` gains one "Golden refresher tooling (devkit)" section:
a refresher/reads/writes/gate table, then `dunconv` (both artifact forms,
`internal/testdata/drlg` and `internal/testdata/drlg/dpiece` as corpus dirs) and
`gamedatagoldens` (committed TSVs in, `parity.Fixture` JSON out, no MPQ, no
network), plus a pointer to `affix_probe`'s own README. Classification is
ADR-019 test instrumentation, cited by section name, not restated. Fixed in the
same edit: the `dvg_probe dpiece` row told readers to convert into
`internal/testdata/drlg`, but `runDpiece` writes `<corpus-dir>/<base>.csv`, so
the corpus dir is `internal/testdata/drlg/dpiece` (matches
`internal/testdata/drlg/README.md:118`).

### Live-reference sweep (honest census)

`grep -rn dgo-assets cmd/ docs/ tools/` after the sweep: **0 in `cmd/`, 0 in
`tools/`**. Remaining `docs/` hits, none of them instructing anyone to run the
command:

- `docs/adrs/adr-019.md` (9 hits): the ADR that deletes it. History, exempt.
- `docs/research/dg-conversion-inventory.md` (10 hits): the pre-ADR evidence
  page ADR-019 cites. Kept verbatim; a snapshot status note at the top now says
  the `dgo-assets` rows are history and its jobs live in
  `devilutionGo-migrate assets`.
- `docs/requirements.md` (1 hit) and `docs/research/dg-graphics-part2.md`
  (1 hit): verbatim requirements study and research volume, each quoting an
  imported bean's scope line. Editing them would falsify the study record.

So `docs/` holds 21 hits total, all history; the census is by file so a reviewer
can re-run it: `grep -rc dgo-assets docs/`.

One further hit sits outside that grep path and outside my scope:
`internal/testdata/gfx/producer/make_fixtures.py:288` defaults `SCRATCH` to
`/tmp/dgo-assets-fixtures`. Goldens/testdata are immutable this lane, and it is
a scratch-directory label, not a tool reference.

### Residual

- Worktree/branch retirement (`dg-pass3`, `dg-conv`) is parent-owned by this
  lane's scope guard. For the record: neither branch is SHA-merged
  (`git rev-list --left-right --count master...dg-pass3` = 2/11,
  `master...dg-conv` = 2/2); their work landed through squash PRs #35 and #36.

## Proof of Work

Run in `/home/blaze/github/bmcszk/dg-surface` at `dg-surface` / 43651a4.

```text
$ go build ./...
(exit 0, no output)
$ gofmt -l .
(empty)
$ ./scripts/lint.sh
lint: sim purity gate
simpurity: internal/game internal/diablo clean (226 sources)
0 issues.
$ go vet ./internal/...
(exit 0, no output)
$ go test ./... -count=1 | grep -c '^ok'
43
$ go test ./... -count=1 | grep -vc '^ok\|^?'
0
$ python3 scripts/beans-lint.py
beans-lint: OK (170 beans)
$ just sqlc-check
no raw SQL outside sqlc-generated code
$ grep -rn dgo-assets cmd/
(empty)
$ grep -rn dgo-assets tools/
(empty)
```

Package count dropped 44 -> 43 because `cmd/dgo-assets` was its own package and
its coverage already lives in `cmd/devilutionGo-migrate` (7 of 7 cases). No test
was skipped, weakened or deleted.
