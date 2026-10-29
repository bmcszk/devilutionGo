---
# devilGo-qo09
title: 'Fix-wave: CI consolidation, dispatch-skill enforcement, readable fixtures'
status: completed
type: feature
priority: normal
created_at: 2026-10-21T11:04:39Z
updated_at: 2026-10-29T16:04:49Z
parent: devilGo-p1t8
---

# Fix-wave: CI consolidation, dispatch-skill enforcement, readable fixtures

Reviewer round-1 comments (verbatim, 2026-10-21) turned into a fix wave.

## Research
- Plan by read-only agent (session devilGo-plan9, /tmp/dg-plan9-report.md), live state verified:
  - bench.yml triggers on push: [epic/qna9-phase-0] + pull_request; branch gets merged+deleted → post-merge bench has no trigger. ci.yml = single `ci` job (vet → golangci-lint → go test).
  - Recent RPC dispatches (devilGo-repair4..8) omitted explicit --skill coding-rules/ponytail/go-dev flags; nothing failed because nothing checked. beans-lint.py already has SKILLS_ROOTS path-validation machinery to extend.
  - Lighttable fixtures are xxd .hex dumps (6 files) decoded by decodeHexDump in internal/light/tables_test.go — machine-readable, not human-readable.

## Acceptance criteria
- [ ] bench runs from ci.yml (survives branch merge+delete); bench.yml deleted
- [ ] beans-lint fails any coding bean (task/bug/feature, todo+) lacking `## Skills for coder` with coding-rules + ponytail (+ go-dev when Go paths touched)
- [ ] lighttable fixtures are human-readable TSV with table/dimension/source headers; Go decoder reads TSV; byte-for-byte identical arrays vs .hex (PoW proves parity)
- [ ] All three tasks closed with PoW; CI green after push

## Skills for coder
- `~/.agents/skills/coding-rules/SKILL.md` — binding style rules for every file touched
- `~/.agents/skills/ponytail/SKILL.md` — smallest change that fully solves

## Progress log
- 2026-10-21: created from reviewer comments; triage by plan9 agent complete.

## Proof of Work\n- All 3 children completed with PoW (devilGo-7f8x bench-fold, devilGo-fzwt beans-lint skills gate, devilGo-d12h TSV fixtures).\n- Parent gates on HEAD bb073ba: gofmt clean; go build OK; go vet OK; go test 547 passed / 28 packages; golangci-lint 0 issues; just beans-lint OK (103 beans); just data-audit ok; fixture parity 6/6 byte-identical.\n- CI: run 39591099202 — ci PASS (incl. new beans-lint step) + bench PASS (bench now inside ci.yml).\n- PR #1 open, body lists all qo09 beans, ready marker @ bb073ba, zero commits after.\nResidual: none.
