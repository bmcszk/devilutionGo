---
# devilGo-ocl1
title: 'Fix coding-rules violations: phase markers + script docstrings (All x All re-sweep)'
status: in-progress
type: task
priority: high
tags:
    - audit-gap
created_at: 2026-10-24T20:31:38Z
updated_at: 2026-10-24T20:42:36Z
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
fixture_fixtures.py docstring 1434 chars. FILE-STRUCTURE scan (same HEAD,
66 non-test .go files): 12 declarations sit after functions, violating
coding-rules ## File structure order (imports/const/var/type before
funcs; typed-enum-after-own-type is idiom-exempt) — audio.go:50,109,
leveldat.go:185, misdat.go:142, misdat/registry.go:141, monstdat.go:151,256,
objdat.go:114,153, missiles/process.go:869,892, missiles/registry.go:56.
Full findings: /tmp/devilgo-sweep/research-rules-violations.md +
/tmp/devilgo-sweep/layout-scan-go.md (worker re-verifies — scans are
leads, not verdicts).

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
- [ ] File-structure: no const/var declaration after a function in any
      non-test .go file (parent re-scan, idiom-exempt typed enums)
- [ ] fixture_fixtures.py module docstring ≤ ~8 contract lines
- [ ] Gates: build rc0, vet rc0, gofmt clean (internal/ cmd/), lint 0
      issues, go test -count=1 ./... ≥9705 green, beans-lint OK

## Skills for coder

- ~/.agents/skills/coding-rules/SKILL.md
- ~/.agents/skills/ponytail/SKILL.md
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/golang-testing/SKILL.md
- ~/.hermes/skills/go-dev-notes/SKILL.md

## Triage

Delegated brainer-adjacent research: parent mechanical scan (research file
/tmp/devilgo-sweep/research-rules-violations.md, HEAD 0baead9) after
brainer-audit trail on devilGo-klu5; user findings 2026-10-24 name the
concrete files. Successor of reopened devilGo-klu5.

## Libraries

stdlib only; no new deps (test/comment restructuring, no behavior).


Claimed 2026-10-24: dispatched to pi coder worker devilGo-ocl1-coder (omni/_worker, standard mode, contract /tmp/devilgo-sweep/contract-rul3.md, log ocl1-coder.log, proc_960e21f80d2a).
