---
# devilGo-ocl1
title: 'Fix coding-rules violations: phase markers + script docstrings (All x All re-sweep)'
status: in-progress
type: task
priority: high
tags:
    - audit-gap
created_at: 2026-10-24T20:31:38Z
updated_at: 2026-10-26T23:55:41Z
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

## Proof of Work — process_ii_test.go whenthen (session note)

Fixed the 4 whenthen findings in `internal/diablo/missiles/process_ii_test.go`:
- `TestProcessFireballImpactRing`, `TestProcessElementalArrowFlight`,
  `TestProcessChargedBoltHit`: added `// then` before the existing assert-helper call.
- `TestProcessMissileExplosionRamp`: added `// when:` (seed + first ramp tick) and
  `// then:` (final tick) markers; helper calls and assertions untouched.

```
$ golangci-lint run ./internal/diablo/missiles/ --max-issues-per-linter 0 --max-same-issues 0 2>&1 | grep -c process_ii_test.go
0
$ go test -count=1 ./internal/diablo/missiles/
ok  github.com/bmcszk/devilutionGo/internal/diablo/missiles  0.024s
$ gofmt -l internal/diablo/missiles/process_ii_test.go
(empty)
```

Residual: package lint is still RED on 17 whenthen findings in OTHER files
(collision_gates_test.go x7, collision_test.go x1, process_test.go x9, incl. two
"when and then on the same line" at process_test.go:414 and :542). Not touched —
same rule, separate files.

Residual: package lint red on 17 whenthen findings in collision_gates_test.go,
collision_test.go, process_test.go (separate files, same rule).

## Progress: internal/diablo/entities/size_test.go

TestStructSizesPinned restructured to top-level given/when/then (table -> measure loop -> assert loop). whenthen linter now clean; subtest names + fatal message unchanged.

```
$ golangci-lint run ./internal/diablo/entities/ --max-issues-per-linter 0 --max-same-issues 0
0 issues.
$ go test -count=1 ./internal/diablo/entities/
ok  github.com/bmcszk/devilutionGo/internal/diablo/entities  0.002s
$ gofmt -l internal/diablo/entities/size_test.go
(empty)
```
