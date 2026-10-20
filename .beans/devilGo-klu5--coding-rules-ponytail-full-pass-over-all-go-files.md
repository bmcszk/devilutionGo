---
# devilGo-klu5
title: coding-rules + ponytail full pass over all Go files (per-file verdicts)
status: draft
type: task
priority: high
tags:
    - quality
created_at: 2026-10-20T13:03:05Z
updated_at: 2026-10-20T13:03:05Z
parent: devilGo-1n0a
---

## Goal
Full coding-rules + ponytail pass over EVERY .go file in the repo (post 3-layer restructure: engine/, game/, diablo/, cmd/), per user directive 2026-10-20. Prior sweeps touched 11/96 files; this epic-round must produce per-file verdicts.

## Acceptance criteria
- [ ] One verdict line per .go file (CLEAN or file:line finding) — count verified against `find . -name '*.go'` inventory
- [ ] All coding-rules violations fixed: narration comments gone, intent/mechanism separation, no vague names, duplication collapsed, file-structure order, black-box given/when/then tests
- [ ] All ponytail violations fixed: dead abstraction/wrapper/option removed, shortest working diff, comment only why, no features nobody asked for (deletion beats addition)
- [ ] `//nolint:revive // 1:1 port` functions stay structurally whole (1:1 fidelity mandate); comments inside them may still be cleaned
- [ ] Golden/parity test assertion semantics untouched
- [ ] Gates: go build ./... rc0; gofmt -l empty; go vet clean; golangci-lint 0 findings; go test ./... all ok; per-file net LOC delta table in PoW

## Skills for coder
- Read /home/blaze/.agents/skills/coding-rules/SKILL.md fully BEFORE editing — binding checklist.
- Read /home/blaze/.agents/skills/ponytail/SKILL.md fully BEFORE editing — smallest complete change, deletion beats addition.
- Read /home/blaze/.agents/skills/go-dev/SKILL.md fully BEFORE editing.

## Rules for coder
- GIT JAIL: no git checkout/switch/reset/commit/push; read-only git allowed.
- No new deps, no nolint/suppressions, no test weakening.
- One domain per wave; report per-file LOC delta.
