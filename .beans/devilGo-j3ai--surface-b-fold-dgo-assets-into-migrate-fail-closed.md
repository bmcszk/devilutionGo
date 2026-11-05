---
# devilGo-j3ai
title: 'Surface B: fold dgo-assets into migrate + fail-closed assets verify'
status: completed
type: task
priority: high
tags:
    - plan:task
created_at: 2026-11-05T21:40:07Z
updated_at: 2026-11-05T22:00:27Z
parent: devilGo-8csd
blocked_by:
    - devilGo-srb3
---

# Surface B: fold dgo-assets into devilutionGo-migrate + fail-closed `assets verify`

## Goal
Per ADR-019 (Decision part 2): the graphics manifest surface becomes a NESTED subcommand of the one program. `cmd/dgo-assets/internal/cli` moves into `cmd/devilutionGo-migrate`; users get `devilutionGo-migrate assets verify`. Top-level subcommands stay exactly three (assets, saves, settings) per ADR-014.

## Research (pre-validated)
- Parent review APPROVE (/home/blaze/github/bmcszk/devilutionGo/.hermes/lane/dg-conv/dg-conv.review.md, task-dgconv-review-brainer).
- docs/research/dg-conversion-inventory.md: dgo-assets subcommands check/list/convert (cli.go:64-68); check defaults to `assets/manifest.yaml` (cli.go:104) and fails open (cli.go:118-131).
- ADR-019 rationale for nesting over a `--verify` flag: assets refuses to run without archives (assets.go:129-140); verify needs no archives; two jobs must not share a flag set.

## Scope
- Move internal/cli (commands, flags, fixtures tests) into cmd/devilutionGo-migrate (subcommand registration pattern matching the existing assets/saves/settings wiring).
- `assets verify`: reads `assets/inventory.yaml` (Surface A rename), resolves every required path under a root, prints `missing <path> (requested by <loader>)` per miss, exits 1 on any miss, exit 0 only on full coverage. Requires no MPQ archives.
- `check` semantics fold into verify (check's coverage output is verify's zero-miss path + a summary line).
- The old `dgo-assets` command STAYS in place, deprecated-but-working, until Surface D deletes it — so this task and Surface C land independently reviewable. ADR-019's "no deprecation window" applies to the NAME going away in D.
- list: keep as a nested `assets list` for the manifest record + inventory (evidence cites in inventory.md "Gates per program").

## Acceptance criteria
- [x] `devilutionGo-migrate assets verify <root>` fail-closed: prints per-miss lines, exit 1 on any miss (integration test with a synthetic missing entry)
- [x] `migrate assets verify` passes on a fixture where everything resolves; requires no MPQ (test asserts archive independence)
- [x] dgo-assets tests ported green under cmd/devilutionGo-migrate (no test weakened or deleted; one redirect documented in the commit body and pinned by TestAssetsWithoutASubcommandRunsTheExtractionPass)
- [x] Top-level `migrate --help` still shows exactly assets, saves, settings
- [x] Full gates green

## Libraries

No new library bindings. Everything stays within the pinned catalog (ADR-009); internal/assets converters already committed.

## Skills for coder
- /home/blaze/.agents/skills/ponytail/SKILL.md
- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/go-tests/SKILL.md
- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/beans/SKILL.md

## Dependencies & blockers
blockedBy: Surface A (inventory rename). Blocks Surface D. Parent: epic devilGo-8csd.

## Triage

Pre-validated by the ADR-019 review (APPROVE): brainer session `task-dgconv-review-brainer` (/home/blaze/github/bmcszk/devilutionGo/.hermes/lane/dg-conv/dg-conv.review.md), evidence in `docs/research/dg-conversion-inventory.md` (file:line cites) and docs/adrs/adr-019.md (merged @ 1d3bbbc). No open questions.

## Proof of Work

All commands run at repo root on branch dg-surface @ 35de077 (post-change):

    go build ./...                     -> BUILD_OK
    gofmt -l .                         -> (empty output)
    ./scripts/lint.sh                  -> 0 issues. (simpurity: 226 sources clean)
    go vet ./internal/...              -> VET_OK
    go test ./... -count=1             -> 44 packages ok, zero FAIL, exit 0
    python3 scripts/beans-lint.py      -> beans-lint: OK (170 beans)
    just sqlc-check                    -> no raw SQL outside sqlc-generated code

RED before the implementation (`go test ./cmd/devilutionGo-migrate/ -count=1`):

    --- FAIL: TestAssetsVerifyAcceptsTheRootAsAPositionalArgument (0.00s)
        assets_inventory_test.go:115: exit code = 1, want 0 (stderr "devilutionGo-migrate assets verify: open inventory assets/inventory.yaml: ...")
    --- FAIL: TestUsageListsExactlyTheThreeTopLevelSubcommands (0.00s)
        main_test.go:192: top-level subcommands = [subcommands: Run "devilutionGo-migrate], want [assets saves settings]
    FAIL github.com/bmcszk/devilutionGo/cmd/devilutionGo-migrate

The salvaged suite from the interrupted session was re-verified green after salvage, not claimed as RED.

Real fail-closed run (synthetic inventory, one entry absent, no archive configured):

    $ go run ./cmd/devilutionGo-migrate assets verify --manifest /tmp/sb-smoke/inventory.yaml /tmp/sb-smoke/root
    missing fonts/smallfont.trn (requested by Source/engine/assets.cpp:410)
    devilutionGo-migrate assets verify: 1 of 2 entries do not resolve under /tmp/sb-smoke/root
    RC=1

Residual: no committed `assets/inventory.yaml` exists yet (Surface A residual), so a bare `assets verify` fails closed on a missing inventory file instead of listing misses; the per-miss path is pinned by fixtures. `manifest.Tool` still reads `dgo-assets` until Surface D. `cmd/dgo-assets` stays live until Surface D.

## Summary of Changes

- `cmd/dgo-assets/internal/cli` folded into `cmd/devilutionGo-migrate` as nested `assets verify|list|convert`; the top level keeps exactly assets/saves/settings (ADR-014).
- `assets verify` is the fail-closed successor of `check`: reads `assets/inventory.yaml`, validates it, prints `missing <path> (requested by <loader>)` per miss, exits 1 on any miss, exits 0 with a coverage line on full coverage, and reads no MPQ archive. Root comes from `--root` or the positional `<root>` the bean names.
- `assets list` and `assets convert` keep their dgo-assets semantics; convert is the folded path Surface C says wins until D.
- Tests: 5 ported tests plus 3 new ones (positional root, exactly-three top-level subcommands, bare `assets` runs extraction). Port of the old `dgo-assets` no-argument usage case is redirected and pinned instead of dropped.

IMPL NOTES: 6b17abb = CLI fold + nested verify/list/convert + ports; 35de077 = review breadcrumb fix in the epic and surface bodies. Prior-session salvage: the untracked assets_*.go files carried the same contract, finished here with lint/structure fixes and the two missing ACs.

## Deviation

The lane brief named `/home/blaze/github/bmcszk/devilutionGo/.hermes/lane/_pre-2026-11-04/dg-conv.review.md` as the breadcrumb target. That directory holds no dg-conv review (`ls ... | grep dg-conv` empty); the review lives at `/home/blaze/github/bmcszk/devilutionGo/.hermes/lane/dg-conv/dg-conv.review.md` (verified: `VERDICT: APPROVE`), so the breadcrumbs point there.
