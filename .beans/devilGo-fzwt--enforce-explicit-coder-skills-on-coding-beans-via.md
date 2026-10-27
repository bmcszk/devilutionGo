---
# devilGo-fzwt
title: Enforce explicit coder skills on coding beans via beans-lint
status: completed
type: task
priority: normal
created_at: 2026-10-21T11:05:01Z
updated_at: 2026-10-21T11:12:01Z
parent: devilGo-qo09
---

# Enforce explicit coder skills on coding beans via beans-lint

Recent RPC dispatches (devilGo-repair4..8) shipped without explicit `--skill coding-rules/ponytail/go-dev` flags — nothing failed because nothing checked. Structural fix at the contract source: the bean authoring gate.

## Scope
Extend `scripts/beans-lint.py` (mirrors existing `check_*` helpers + SKILLS_ROOTS machinery):
- New rule: every bean of type task/bug/feature in status todo or in-progress MUST contain a `## Skills for coder` section listing at least `~/.agents/skills/coding-rules/SKILL.md` and `~/.agents/skills/ponytail/SKILL.md`.
- Go-touching beans (body mentions `internal/` or `cmd/` with `.go` context, or `go test`/`gofmt` gates) must additionally list `~/.agents/skills/go-dev/SKILL.md`.
- Violation = exit 1. Update the module docstring. If beans-lint is not invoked by CI, wire it into ci.yml in the same pass (check `.github/workflows/ci.yml`).

## Acceptance criteria
- [ ] beans-lint exits 1 on a synthetic coding bean without the section, 0 with it
- [ ] Existing 99 beans still lint green after the rule (fix any legacy offenders in same pass)
- [ ] beans-lint runs in CI (or pre-commit) so the check fires without a human

## Skills for coder
- `~/.agents/skills/coding-rules/SKILL.md` — read fully BEFORE acting
- `~/.agents/skills/ponytail/SKILL.md` — read fully BEFORE acting

## Triage
Delegated read-only planner (pi-as-brainer session devilGo-plan9, /tmp/dg-plan9-report.md): repair4..8 dispatches omitted --skill flags with no failing check; beans-lint already owns SKILLS_ROOTS validation — rule belongs at bean authoring (contract source), not dispatch wrapper.

## Libraries
- stdlib only (re/glob/os) — extends existing check_* helpers

## Proof of Work
- check_coder_skills in scripts/beans-lint.py: task/bug/feature in todo/in-progress require ## Skills for coder with coding-rules + ponytail (+go-dev for Go-touching bodies); archive exempt.
- Synthetic tests: bad bean exit 1, good exit 0. 48 legacy todo/in-progress beans fixed via beans CLI.
- just beans-lint: OK (103 beans) after parent appended ## Triage + ## Libraries to the 3 new wave beans.
- ci.yml gained a beans-lint step (just beans-lint) so the rule fires in CI too.
Residual: none.
