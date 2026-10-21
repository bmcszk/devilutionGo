---
# devilGo-qo09
title: 'Fix-wave: CI consolidation, dispatch-skill enforcement, readable fixtures'
status: in-progress
type: epic
priority: normal
created_at: 2026-10-21T11:04:39Z
updated_at: 2026-10-21T11:05:42Z
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
