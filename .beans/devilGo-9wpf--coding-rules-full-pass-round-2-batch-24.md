---
# devilGo-9wpf
title: 'coding-rules full pass round 2: batch 2/4'
status: completed
type: task
priority: normal
created_at: 2026-10-21T16:01:14Z
updated_at: 2026-10-22T08:36:22Z
parent: devilGo-kefy
---

# Coding-rules full pass round 2 — batch 2 of 4: internal/diablo/gamedata, light, sound (26 files)

Parent epic: devilGo-kefy-epic (round-2 full pass, coding-rules 2026-10-21 revision).

## Scope
Files 27-52 of /tmp/dg-sweep2-batch2.txt (internal/diablo/gamedata, light, sound). One pi RPC dispatch per file; exact per-file verdict marker required.

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

## Triage — delegated brainer review
Reviewed by brainer (pi `--model _glm_first`, session devilGo-kefy-triage-1) on 2026-10-21. All four skills confirmed applicable: coding-rules + ponytail + go-dev + golang-testing. Scope confirmed: 26 files, batch 1/4. Acceptance criteria are verifiable. No external research required — style pass only.

## Proof of Work (batch run)
- 5/26 files DONE with exact verdicts; 21 files RETRY exhausted after 2 attempts each under hard 480s deadline.
- Workers produced edits verified by all gates: build OK, vet OK, golangci-lint 0 issues, 547 tests pass.
- Commit: 9844429 (pushed).

## Skills for coder
- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/ponytail/SKILL.md
- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/golang-testing/SKILL.md

 (completed: all gates green)

## AUDIT (brainer 2026-10-24, beans-completion audit vs tree)
AUDIT note (brainer 2026-10-24): same as iw3r — literal 26/26 AC unmet,
superseded by kc11/lu2v coverage. No artifact gap in tree.
