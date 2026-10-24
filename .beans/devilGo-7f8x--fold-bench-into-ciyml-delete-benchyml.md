---
# devilGo-7f8x
title: Fold bench into ci.yml, delete bench.yml
status: completed
type: task
priority: normal
created_at: 2026-10-21T11:04:57Z
updated_at: 2026-10-21T11:12:01Z
parent: devilGo-qo09
---

# Fold bench into ci.yml, delete bench.yml

Branch `epic/qna9-phase-0` gets merged and deleted; bench.yml (push: [epic/qna9-phase-0] + PR) loses its push trigger post-merge. Bench is non-gating and cheap — fold it into ci.yml's trigger surface.

## Scope
- Add a second `bench` job to `.github/workflows/ci.yml` (checkout, setup-go, budget gate `go test ./internal/engine/ -count=1`, non-gating `go test -run XXX -bench . -benchtime 100x ./internal/...`, artifact upload renamed `bench-output-${{ github.run_id }}`).
- Delete `.github/workflows/bench.yml`.
- Trigger unchanged: push master + pull_request.

## Acceptance criteria
- [ ] ci.yml carries both `ci` and `bench` jobs; bench.yml gone
- [ ] actionlint / YAML valid; workflow triggers on master push + PR
- [ ] CI green on push

## Skills for coder
- \`~/.agents/skills/go-dev/SKILL.md\` — read fully BEFORE acting
- `~/.agents/skills/coding-rules/SKILL.md` — read fully BEFORE acting
- `~/.agents/skills/ponytail/SKILL.md` — read fully BEFORE acting

## Triage
Delegated read-only planner (pi-as-brainer session devilGo-plan9, /tmp/dg-plan9-report.md): verified bench.yml trigger = push [epic/qna9-phase-0] + PR; branch merged+deleted post-review, so push trigger dies; folding into ci.yml keeps one trigger surface.

## Libraries
- none (workflow YAML only)

## Proof of Work
- ci.yml now carries ci + bench jobs (bench steps carried verbatim, artifact renamed bench-output-${{ github.run_id }}); bench.yml deleted; triggers push master + pull_request.
- YAML parse check (yaml.safe_load) exit 0. Parent gates: gofmt 0, build OK, vet OK, test 547/28, golangci-lint 0.
Residual: live GHA run verdict pending (checked after push).
