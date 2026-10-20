---
# devilGo-7otq
title: Use latest stable Go (1.27.3) in go.mod + GHA; run go fix modernization
status: todo
type: task
priority: high
tags:
    - infra
created_at: 2026-10-20T13:02:57Z
updated_at: 2026-10-20T13:16:38Z
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

## Triage (2026-10-20, research-verified)
- Live state: go.mod `go 1.27.1` (line 3); ci.yml `go-version: "1.27"` (line 14) + golangci-lint-action@v6 UNPINNED version. Verified online: latest stable go = 1.27.3 (2026-10-19, go.dev release history); latest golangci-lint = v2.15.0 (2026-10-16, official changelog).
- CI lint failure observed on 8dcb739 run #39413471115: `golangci-lint exit code 3` = action default linter build vs go 1.27 packages mismatch; pin v2.15.0 fixes.
- `go fix ./...`: go1.27 modernize analyzers; gate every rewrite (build/vet/lint/test); revert any hunks that break golden parity tests, list as residual.
- Local toolchain: go1.27.1 installed (arch); pacman upgrade to 1.27.3 noted in AC.
- Wave order rationale: modernization FIRST, before coding-rules/ponytail sweep, so `go fix` rewrites are not churned twice.
- Ready: ACs complete, evidence cited. → todo.

## Libraries
- stdlib only (no third-party deps at this stage)

## Triage correction (2026-10-20)
Prior inline triage withdrawn per user directive — triage is pi-brainer lane. Reset to draft for brainer triage pass.

## Amendment (brainer triage 2026-10-20) — fidelity exclusion
- go fix hunks landing inside functions whose comments cite upstream file:line (17 hits: engine/random.go, diablo/txtdata/mana.go, diablo/gamedata/quests/quests.go, diablo/save/diablosha.go, diablo/animation/animationinfo.go) are REVERTED to preserve structural parity; each reverted hunk listed in PoW as "skipped: 1:1 port". Non-ported hunks (incl. _test.go) applied. Proof: go fix -diff before/after in PoW.
## Amendment 2 — action pin
- golangci/golangci-lint-action@v9 (v9.3.0) with version: v2.15.0 — @v6 predates golangci-lint v2 and defaulted to v1.64.8 (the CI exit-3 root cause, confirmed from run log).
## Amendment 3 — toolchain
- PoW shows build executed with go1.27.3 via GOTOOLCHAIN=auto from go.mod; pacman only packages 2:1.27.2-1 today — upgrade when Arch ships it, record actual local version.
## Amendment 4 — scope closures
- bench.yml go-version: stable left floating deliberately (resolves 1.27.3+).
- "CI green on pushed HEAD" verified-by-orchestrator, not leaf-verifiable (GIT JAIL).
