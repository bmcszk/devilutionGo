---
# devilGo-ym0r
title: 'persist: DeleteHero partial-failure regression test'
status: todo
type: task
priority: normal
created_at: 2026-11-02T14:25:47Z
updated_at: 2026-11-02T21:56:44Z
---
---

# persist: DeleteHero partial-failure regression test (PR14 brainer L1)

## Goal

Pin DeleteHero atomicity with a regression test: injecting a failing
delete (ctx cancel between deletes or a dependent-table constraint)
must leave hero + all dependent rows intact. Upstream anchor:
internal/persist/store_sqlite.go:152-173 (BeginTx wrapper, four
sequential deletes).

## Research

PR14 brainer finding 1 (Low): removing the BeginTx wrapper keeps every
existing test green — atomicity is pinned only by reading the code.

## Triage

- Triage evidence: PR14 brainer review finding 1 (Low) — DeleteHero
  atomicity pinned only by code reading; regression test demanded.
  Source: brainer session on dg-persist-deletehero-atomic-tx lane
  (merged PR #14, store_sqlite.go:152-173).

- 2026-11-02: wave closed by coder (a8b0354): ym0r atomicity regression test (RED on unwrapped BeginTx, restored) + oqho LoadLevel StateHash round-trip + xwtt AC closes (v0/v2 migration replay, .sv export hex pin, uiheroinfo mapping). Gates: persist pkg ok 9.9s, lint 0 issues, sqlc-check clean, pushed.

## Acceptance criteria

- [ ] Regression test injects a failing delete (ctx cancel between
      deletes or dependent-table constraint)
- [ ] Test asserts hero + all dependent rows survive (no partial state)
- [ ] Mutation-checked: unwrapping BeginTx turns the test RED, restored
      after
- [ ] Gates green: go test ./internal/persist/ -count=1, lint 0 issues

## Skills for coder

- `~/.agents/skills/coding-rules/SKILL.md` — read fully BEFORE acting
- `~/.agents/skills/ponytail/SKILL.md` — read fully BEFORE acting
- `~/.agents/skills/go-tests/SKILL.md` — regression-test structure
- `~/.agents/skills/go-dev/SKILL.md` — database access rules

## Libraries

- stdlib only; modernc.org/sqlite via existing store

## Dependencies & blockers

None — test-only change on the existing SQLiteStore.

## Open questions

None.
