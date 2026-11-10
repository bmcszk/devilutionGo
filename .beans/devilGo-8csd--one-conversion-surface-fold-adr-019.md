---
# devilGo-8csd
title: One conversion surface fold (ADR-019)
status: completed
type: epic
priority: high
tags:
    - plan:task
created_at: 2026-11-05T21:40:03Z
updated_at: 2026-11-10T06:13:20Z
---

# One conversion surface fold (ADR-019)

## Goal
End with ONE conversion program: `devilutionGo-migrate` handles MPQ -> assets (with CLX conversion), saves, settings. `dgo-assets` dies; test instrumentation stays carved out. Implements docs/adrs/adr-019.md (merged via "ADR-019: one conversion surface (inventory + ADR draft)").

## Research (pre-validated triage)
Parent review APPROVE: /home/blaze/github/bmcszk/devilutionGo/.hermes/lane/dg-conv/dg-conv.review.md (brainer session task-dgconv-review-brainer). Facts: docs/research/dg-conversion-inventory.md — split pipelines; two files named manifest.yaml (cmd/devilutionGo-migrate/manifest.go:85-92 per-file vs internal/assets/manifest/manifest.go:91-96 per-loader); `dgo-assets check` fail-open on wrong manifest (cmd/dgo-assets/internal/cli/cli.go:118-131).

## Children (execution order)
- Surface A: manifest model fail-closed + inventory rename (FIRST, blocks B and C)
- Surface B: fold dgo-assets into migrate + nested `assets verify` (blocked by A)
- Surface C: inline CLX conversion in migrate assets (blocked by A)
- Surface D: deprecation sweep — dgo-assets deletion + one devkit README (blocked by B and C)

## Acceptance criteria
- [x] `devilutionGo-migrate` covers assets+saves+settings with subcommand tree intact (ADR-014 three top-level commands)
- [x] `grep -r dgo-assets` in cmd/, docs/, tools/ returns no live references: 0 in cmd/ and tools/, docs hits are ADR-019 + the snapshot-flagged research ledger
- [x] All parity + migrate + assets tests green throughout; ADR-005 embed parity preserved (final state: 43 pkgs ok, 0 FAIL)
- [x] beans-lint OK at every commit (170 beans, re-run after each Surface D bean edit)

## Triage

Pre-validated by the ADR-019 review (APPROVE): brainer session `task-dgconv-review-brainer` (/home/blaze/github/bmcszk/devilutionGo/.hermes/lane/dg-conv/dg-conv.review.md), evidence in `docs/research/dg-conversion-inventory.md` (file:line cites) and docs/adrs/adr-019.md (merged @ 1d3bbbc). No open questions.

## DECIDED / progress (dg-surface lane)

- Surface D landed in 43651a4: `cmd/dgo-assets` deleted, `manifest.Tool` records
  the folded identity, `tools/probe/README.md` carries the one devkit section
  (dunconv + gamedatagoldens + affix_probe, ADR-019 test instrumentation).
- Remaining epic AC: `saves` and `settings` are still `not implemented yet`
  stubs (`cmd/devilutionGo-migrate/main.go`), so the ADR-014 three-command tree
  is only one third built. Not provable this lane; follow-up work, not a
  Surface D defect.
- Worktree/branch retirement for `dg-pass3` and `dg-conv` is parent-owned per
  the lane scope guard; neither branch is SHA-merged (squash PRs #35, #36).

## ADR-020 execution (lane dg-binaries s1, 2026-11-06)

- Fold landed 6372b4b / ae7cb2c: `devilutionGo` is the one game binary with
  `--headless` (+ `-ticks N`); `cmd/devilutionGo-headless` is gone and its six
  test cases moved into `cmd/devilutionGo/main_test.go`. Its bean (headless as a
  devilutionGo parameter) is completed.
- Move landed 9f22112: `gamedatagoldens` -> `tools/gamedatagoldens`, recipe and
  devkit docs retargeted, no golden regenerated. Its bean stays in-progress on
  two ACs that the pre-existing golden drift makes unprovable inside a
  goldens-immutable lane (draft bug filed).
- `cmd/` now holds `devilutionGo/` and `devilutionGo-migrate/` only. The epic AC
  on the ADR-014 three-command migrate tree is unchanged and still unprovable
  (saves/settings are stubs).

## Progress log

- 2026-11-09 parent sweep (desk task-dgmain): last open child [Two binaries: move gamedatagoldens out of the user-facing command set](.beans/devilGo-48fe--two-binaries-move-gamedatagoldens-out-of-the-user.md) closed (postmerge54 recipe exit 0, empty diff, worker log at .hermes/lane/dg-post54cleanup/postmerge54.log). All fold surfaces A-D + binaries + devkit chain verified terminal on master ad1eea4.
- Remaining AC (migrate covering assets+saves+settings with the ADR-014 three-command tree) is now in-flight through the ready children of [One migration tool devilutionGo-migrate (ADR-014)](.beans/devilGo-o8e7--one-migration-tool-devilutiongo-migrate-adr-014.md): [Migrate saves: legacy .sv into SQLite](.beans/devilGo-vu3u--migrate-saves-legacy-sv-into-sqlite.md) + [Migrate settings: diablo.ini checks plus registry YAML views](.beans/devilGo-yxv5--migrate-configs-and-registries-to-yaml-and-csv.md) (+ the draft [Migration recipes and CI fixture path](.beans/devilGo-w0nb--migration-recipes-and-ci-fixture-path.md)) - the fold epic stays open until those land.
- Claimed 2026-11-09 parent (this session's goal = finishing the ADR-019 fold chain).

## Summary

Every fold surface shipped through user merges: ADR-019 inventory + ADR draft (#35 assembly route on the ledger line through #36/#37), the migration binary assets/saves/settings suite, ADR-020 two-binary law (devilutionGo + devilutionGo-migrate only), ADR-021 devkit refresher fold, dgo-assets gone, gamedatagoldens command moved out of cmd/ entirely (tools/goldens), headless as a devilutionGo parameter.

## Proof of Work (parent verification, 2026-11-11, master 67ac9c5)

- ALL 7 children terminal (query result, parent-run: 23yc, 2h1y, 6md2, bz7q, j3ai, srb3, 48fe - completed; r7qf-slot in the ledger tree is the scrapped 0xBF duplicate).
- ADR-014 final-state gate (worker evidence run, verbatim in .hermes/lane/dg-migrate-settings/postmerge56.log, continuation section): `go run ./cmd/devilutionGo-migrate --help` lists EXACTLY the three real subcommands (assets with nested verify/list/convert, saves, settings) - NO "(not implemented yet)" note anywhere; per-subcommand -h help printed for saves + settings.
- Gates at 67ac9c5 (synced main checkout, parent-approved worker run): go build OK; gofmt EMPTY; lint.sh 0 issues (sim purity 240 sources); go test ./... -count=1 rc 0 (46 pkgs, no FAIL); just sqlc-check clean; beans-lint OK (184 beans).
- Each subcommand's own evidence: assets lane PR #24 + fold PRs #35/#36/#37; saves PR #55 (review APPROVE static, fix round c3b9f53-era, merged 257b2cc); settings PR #56 (review REQUEST-CHANGES -> 7 findings fixed, merged 67ac9c5).
- No dgo-assets live refs; cmd/ contains only devilutionGo/ + devilutionGo-migrate/ (verified by the earlier waves and re-proved in the saved gate tails).

## Deviations recorded (fidelity gate)

- .sp files are NOT carried (help text says so explicitly): recorded deviation, not a defect - upstream has no .sp save families (the saves lane verified this against the pinned clone).
- ADR-020 wave's then-unprovable ACs (the goldens recipe checks) closed in the same week: the spelldat drift root was fixed upstream (devilGo-49ph option 1 landed bf36612 / PR #39), after which recipe and generator re-proofs ran clean (postmerge54 worker log).

Residual: none for the epic scope as written (the ADR-019 GOAL: ONE conversion surface). The persist-hardening follow-ups (spawn stash row model / ExportStash asymmetry draft) are tracked as their own beans under the persistence feature.
