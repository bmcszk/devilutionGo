---
# devilGo-ej38
title: 'coding-rules full pass round 2: batch 3/4'
status: todo
type: task
priority: normal
created_at: 2026-10-21T16:01:14Z
updated_at: 2026-10-21T16:01:40Z
parent: devilGo-kefy
---

# Coding-rules full pass round 2 — batch 3 of 4: internal/diablo rest + internal/game (26 files)

Parent epic: devilGo-kefy-epic (round-2 full pass, coding-rules 2026-10-21 revision).

## Scope
Files 53-78 of /tmp/dg-sweep2-batch3.txt (internal/diablo rest + internal/game). One pi RPC dispatch per file; exact per-file verdict marker required.

## Acceptance criteria
- [ ] 26 verdicts == 26 files in batch list, unique, no missing
- [ ] Contracts include phase-marker preservation in _test.go, tightest visibility, call-order layout, duplication check
- [ ] Every FIXED file package-gated green
- [ ] Commit per file, gates green between commits

## Libraries
- none new

## Skills for coder
- /home/blaze/.agents/skills/coding-rules/SKILL.md — read fully BEFORE acting on every file
- /home/blaze/.agents/skills/ponytail/SKILL.md — read fully BEFORE acting (outranked)
- /home/blaze/.agents/skills/go-dev/SKILL.md — read fully BEFORE acting
- /home/blaze/.agents/skills/golang-testing/SKILL.md — read fully BEFORE acting (test files)
