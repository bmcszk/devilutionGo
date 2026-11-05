---
# devilGo-8csd
title: One conversion surface fold (ADR-019)
status: todo
type: epic
priority: high
tags:
    - plan:task
created_at: 2026-11-05T21:40:03Z
updated_at: 2026-11-05T22:21:12Z
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
- [ ] `devilutionGo-migrate` covers assets+saves+settings with subcommand tree intact (ADR-014 three top-level commands)
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
