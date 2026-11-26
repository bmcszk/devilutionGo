---
# devilGo-2yge
title: status.sh must report the pinned lint oracle (scripts/lint.sh), not raw golangci-lint
status: completed
type: task
priority: normal
created_at: 2026-11-26T19:59:17Z
updated_at: 2026-11-26T20:21:48Z
parent: devilGo-p1t8
---

# status.sh must report the pinned lint oracle (scripts/lint.sh), not raw golangci-lint

## Goal
scripts/status.sh's lint line runs the SYSTEM golangci-lint binary directly (status.sh:9: `lint=$(golangci-lint run ./... ...)`), bypassing the pinned fail-closed oracle (scripts/lint.sh: vanilla golangci v2.15.0 + whenthen vettool + canary + sim-purity + executable-surface). Status output can therefore disagree with the gates: a system-binary version drift or plugin difference reports clean while the real gate fails (the exact oracle-law violation lint.sh's header documents, and the drift class the gate-parity wave killed for hooks/CI).

## Research (triage source: gate-parity triage, out-of-scope list)
Recorded finding: "scripts/status.sh's raw golangci-lint call (oracle-law violation, separate bean)". The fix shape follows lint.sh's own law: the only way to run lint is scripts/lint.sh.

## Scope
- scripts/status.sh:9: replace the raw call with `scripts/lint.sh ./... --timeout=5m`, reporting clean/FAIL in the same compact line style as the build/tests lines (lint.sh's output does not carry the `golangci-lint: ` prefix the sed expects, so the extraction goes).
- No other lines change. status.sh already runs full build + test, so the added oracle time is in line with the command's existing cost.

## Acceptance criteria
- [x] status.sh's lint line reports the pinned oracle's verdict: seeded violation (one badly-formatted file) -> status shows lint FAIL; restored -> clean (both tails verbatim in the PoW)
- [x] `grep -n golangci scripts/status.sh` shows no direct golangci-lint invocation (only the scripts/lint.sh path); gates green at head (build/vet/lint/test tails pasted)

## Libraries
stdlib bash + the pinned lint oracle via scripts/lint.sh.

## Skills for coder
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.agents/skills/ponytail/SKILL.md — smallest landable diff (also beans-lint-mandated)
- ~/.agents/skills/shell-traps/SKILL.md — the sed-extraction silent-empty trap is exactly this class
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md — oracle law

## Triage
Scoping recorded by the gate-parity triage (task-gp-triage-brainer, out-of-scope list: "status.sh's raw golangci-lint call (oracle-law violation, separate bean)"); parent verified the violation live (status.sh:9) today. PROMOTE-TODO.



## Proof of Work

- One commit 2de019e `chore(status): report the pinned lint oracle, not raw golangci-lint`; 1 file, 1 line: the raw `golangci-lint run` call replaced by `scripts/lint.sh ./... --timeout=5m` reporting clean/FAIL in the sibling compact style.
- Seeded proof (both tails verbatim in the wave log): badly-formatted .go planted -> `lint: FAIL`; restored -> `lint: clean`. `grep -n golangci scripts/status.sh` -> zero matches (rc=1), so no direct binary invocation remains.
- Review (task-so-review-brainer): **APPROVE** - only lint path is the pinned oracle, FAIL path prints, no other lines touched, scope jail held.
- Parent gates at 2de019e: gofmt empty, build + vet rc0, scripts suite ok.
- Residual: none. (status.sh's lint line now takes ~10s - the oracle install is cached after the first run.)
