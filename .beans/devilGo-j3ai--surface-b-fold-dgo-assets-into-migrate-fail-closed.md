---
# devilGo-j3ai
title: 'Surface B: fold dgo-assets into migrate + fail-closed assets verify'
status: todo
type: task
priority: high
tags:
    - plan:task
created_at: 2026-11-05T21:40:07Z
updated_at: 2026-11-05T21:41:44Z
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
- [ ] `devilutionGo-migrate assets verify <root>` fail-closed: prints per-miss lines, exit 1 on any miss (integration test with a synthetic missing entry)
- [ ] `migrate assets verify` passes on a fixture where everything resolves; requires no MPQ (test asserts archive independence)
- [ ] dgo-assets tests ported green under cmd/devilutionGo-migrate (no test weakened or deleted; a redirect documented in the PR if any)
- [ ] Top-level `migrate --help` still shows exactly assets, saves, settings
- [ ] Full gates green

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
