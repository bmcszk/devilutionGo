---
# devilGo-7otq
title: Use latest stable Go (1.27.3) in go.mod + GHA; run go fix modernization
status: draft
type: task
priority: high
tags:
    - infra
created_at: 2026-10-20T13:02:57Z
updated_at: 2026-10-20T13:02:57Z
parent: devilGo-1n0a
---

## Goal
Use latest stable Go everywhere (code + GitHub Actions) and modernize with `go fix`.

## Acceptance criteria
- [ ] go.mod declares `go 1.27.3` (latest stable, released 2026-10-19, verified go.dev release history)
- [ ] .github/workflows/ci.yml: `go-version: '1.27.3'` exact + `golangci-lint-action` pinned `version: v2.15.0` (latest, verified 2026-10-20; fixes CI lint exit 3)
- [ ] `go fix ./...` (modernize) applied — only changes that keep gates green; any skipped fix documented
- [ ] All gates green: go build/vet/golangci-lint 0 findings/go test ./...
- [ ] CI green on pushed HEAD
- [ ] Local dev toolchain upgraded to 1.27.3 (pacman; environment note in PoW)

## Skills for coder
- Read /home/blaze/.agents/skills/go-dev/SKILL.md fully BEFORE editing.
- Read /home/blaze/.agents/skills/coding-rules/SKILL.md fully BEFORE editing (no narration comments; minimal diff).
- ponytail: smallest change — version bumps + `go fix` output only, no refactor rides.

## Rules for coder
- GIT JAIL: no git checkout/switch/reset/commit/push. Leaf work: edit files, run gates, report.
- Do NOT touch anything beyond go.mod, ci.yml, and files `go fix` rewrites.
