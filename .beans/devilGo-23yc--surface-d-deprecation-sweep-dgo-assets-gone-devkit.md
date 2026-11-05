---
# devilGo-23yc
title: 'Surface D: deprecation sweep - dgo-assets gone + devkit README'
status: todo
type: task
priority: high
tags:
    - plan:task
created_at: 2026-11-05T21:40:07Z
updated_at: 2026-11-05T21:41:44Z
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
- [ ] cmd/dgo-assets deleted; `go build ./...` green
- [ ] `grep -rn dgo-assets cmd/ docs/ tools/` -> no live references (only beans/archive + ADR-019 mention)
- [ ] One devkit README section covers the golden refreshers with concrete invocation examples
- [ ] Full gates green: build, gofmt, lint.sh, vet, go test ./... -count=1 (44 pkgs, incl. TestDPieceParity 256/256 + L1-L4 tile gates), beans-lint, just sqlc-check
- [ ] Retire the merged dg-pass3 + dg-conv worktrees/branches (cleanup law) in the same session

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
