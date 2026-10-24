---
# devilGo-ocl1
title: 'Fix coding-rules violations: phase markers + script docstrings (All x All re-sweep)'
status: todo
type: task
priority: high
tags:
    - audit-gap
created_at: 2026-10-24T20:31:38Z
updated_at: 2026-10-24T20:31:38Z
parent: devilGo-97mn
---

# Fix coding-rules violations: phase markers + script docstrings

## Goal

All x All coding-rules re-sweep after user findings: given/when/then
markers wrong-order or missing across test files (light/tables_test.go,
assets/cl2.go named; "there may be others too") and TLDR script
docstrings (scripts/fixture_fixtures.py). Reopened-bean successor of
devilGo-klu5 whose sweep PoW missed these.

## Research

Parent mechanical scan at HEAD 0baead9: 7 files with when-before-given,
1 with then-before-when, 8 test files with no markers at all,
fixture_fixtures.py docstring 1434 chars. Full findings:
/tmp/devilgo-sweep/research-rules-violations.md (worker re-verifies —
scan is a lead, not a verdict).

## Scope

- All 52 internal/**/*_test.go: per-file verdict; reorder/add markers per
  coding-rules ## Comments line 49 (markers required when phases not
  obvious; order given→when→then; no marker-narration essays).
- internal/assets/cl2.go + flagged .go: comments-only narration cleanup.
- scripts/*.py: docstrings to contract-only.
- Comments/test-structure only: no config, no nolint, no assertion or
  behavior changes, citations preserved.

## Acceptance criteria

- [ ] Per-file verdict line for all 52 test files + 4 scripts in report
- [ ] No when-before-given / then-before-when remaining (parent re-scan)
- [ ] No unmarked 3-phase test where phases are not obvious (parent spot-check)
- [ ] fixture_fixtures.py module docstring ≤ ~8 contract lines
- [ ] Gates: build rc0, vet rc0, gofmt clean (internal/ cmd/), lint 0
      issues, go test -count=1 ./... ≥9705 green, beans-lint OK

## Skills for coder

- ~/.agents/skills/coding-rules
- ~/.agents/skills/ponytail
- ~/.agents/skills/go-dev
- ~/.agents/skills/golang-testing
- ~/.hermes/skills/go-dev-notes

## Triage

User findings 2026-10-24 with named files; parent scan quantified the
spread (research file). Successor of reopened devilGo-klu5; audit trail:
user message 2026-10-24 + research-rules-violations.md.

## Libraries

stdlib only; no new deps (test/comment restructuring, no behavior).
