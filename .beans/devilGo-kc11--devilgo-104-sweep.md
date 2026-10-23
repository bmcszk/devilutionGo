---
# devilGo-kc11
title: devilGo-104-sweep
status: in-progress
type: task
priority: critical
created_at: 2026-10-22T09:23:49Z
updated_at: 2026-10-23T13:03:09Z
---

## Scope
Apply coding-rules to all 104 Go files in devilutionGo repo via pi delegation.

## Skills for coder
- /home/blaze/.agents/skills/coding-rules/SKILL.md — read fully BEFORE acting on every file
- /home/blaze/.agents/skills/ponytail/SKILL.md — read fully BEFORE acting on every file (outranked by coding-rules on conflict)
- /home/blaze/.agents/skills/go-dev/SKILL.md — read fully BEFORE acting

## Acceptance Criteria
- Each of 104 .go files processed via individual pi dispatch
- Rules applied: narration comments removed, why-only kept, test phase markers present
- All gates green per file (build, vet, lint, tests)
- No hallucinated changes - each dispatch verifies with tool output
- PR pushed with all changes


## Progress log
- 2026-10-23 claimed: superseded the 4 round-2 batch children (iw3r/9wpf/ej38/ar2y all completed); remaining scope = 15 uncommitted test-marker files (given/when/then phase markers, matches commit 219ad07 pattern) then gates + push.
