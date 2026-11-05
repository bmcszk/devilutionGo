---
# devilGo-srb3
title: 'Surface A: manifest model fail-closed + inventory rename'
status: todo
type: task
priority: high
tags:
    - plan:task
created_at: 2026-11-05T21:40:07Z
updated_at: 2026-11-05T21:41:44Z
parent: devilGo-8csd
---

# Surface A: manifest model fail-closed + inventory rename

## Goal
Per ADR-019 (part 3 of its Decision): the two same-named YAML files split honestly. The migration extraction record keeps the name `manifest.yaml` (per-file shape, `cmd/devilutionGo-migrate/manifest.go:85-92`). The graphics requirements list is written as `assets/inventory.yaml` (per-loader shape, `internal/assets/manifest/manifest.go:91-96`). And the model fails closed.

## Research (pre-validated)
- Parent review APPROVE (/home/blaze/github/dg-conv.review.md, brainer session task-dgconv-review-brainer).
- Findings in docs/research/dg-conversion-inventory.md: `internal/assets/manifest` Validate refuses only missing tool + wrong version (manifest.go:126-131) and does NOT refuse an empty entry list — that is how `dgo-assets check` silently passes ("0 entries in 0 families", cli.go:118-131) against the wrong file.
- Merging the shapes was rejected in ADR-019 with evidence (a record cannot express a requirement; per-loader rows carry loader/width_source extraction cannot produce).

## Scope
- internal/assets/manifest: `Validate` rejects empty/None entries list, missing `tool`, wrong `version` (each with a test). Keep per-loader entry shape as the model (renaming model types optional; no behavior change beyond fail-closed).
- Writers: graphics inventory output path/name -> `assets/inventory.yaml`; migration list stays `manifest.yaml`. Update everything that reads either name (cli.go default `assets/manifest.yaml` -> `assets/inventory.yaml`).
- Move/port the tests that covered the old single-name flow (they follow the writer).

## Acceptance criteria
- [ ] Validate fails closed: empty entries, missing tool, wrong version -> error (unit tests)
- [ ] Files named per ADR-019: extraction = `manifest.yaml`, requirements = `assets/inventory.yaml` (integration test asserts exact filenames)
- [ ] All existing migrate + dgo-assets tests green after the rename; no shape change beyond the entry-list fail-closed
- [ ] Full gates green (build, gofmt, lint.sh, vet, go test ./... -count=1, beans-lint, just sqlc-check)

## Libraries

No new library bindings. Everything stays within the pinned catalog (ADR-009); internal/assets converters already committed.

## Skills for coder
- /home/blaze/.agents/skills/ponytail/SKILL.md
- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/go-tests/SKILL.md
- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/beans/SKILL.md

## Dependencies & blockers
blockedBy: none (FIRST). Blocks Surface B and Surface C. Parent: epic devilGo-8csd (One conversion surface fold).

## Triage

Pre-validated by the ADR-019 review (APPROVE): brainer session `task-dgconv-review-brainer` (/home/blaze/github/dg-conv.review.md), evidence in `docs/research/dg-conversion-inventory.md` (file:line cites) and docs/adrs/adr-019.md (merged @ 1d3bbbc). No open questions.
